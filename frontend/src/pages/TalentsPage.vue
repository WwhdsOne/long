<script setup>
import { computed, onMounted, ref } from 'vue'
import { usePublicPageState } from './publicPageState'

const { isLoggedIn, talentPoints: sharedTalentPoints } = usePublicPageState()

const loading = ref(false)
const learnLoading = ref(false)
const errorMsg = ref('')
const treeDefs = ref(null)
const talentState = ref(null)
const selectedTree = ref('normal')
const selectedSubTree = ref('')
const selectedTalentID = ref('')

const treeConfig = {
  normal: { name: '均衡攻势', color: '#2ab06f' },
  armor: { name: '碎盾攻坚', color: '#c08a2e' },
  crit: { name: '致命洞察', color: '#c73b58' },
}

const tierLabels = { 0: '基石', 1: '一阶', 2: '二阶', 3: '三阶', 4: '终极' }
const tierRadius = { 0: 60, 1: 112, 2: 164, 3: 216, 4: 270 }

const learnedSet = computed(() => new Set(talentState.value?.talents || []))
const trees = ['normal', 'armor', 'crit']

const availableTalentPoints = computed(() => {
  if (typeof talentState.value?.talentPoints === 'number') {
    return Math.max(0, Number(talentState.value.talentPoints))
  }
  return Math.max(0, Number(sharedTalentPoints.value || 0))
})

const currentTreeDefs = computed(() => {
  if (!treeDefs.value?.trees) return []
  return treeDefs.value.trees[selectedTree.value]?.talents || []
})

const subTreeDefs = computed(() => {
  if (!selectedSubTree.value || !treeDefs.value?.trees) return []
  return treeDefs.value.trees[selectedSubTree.value]?.talents || []
})

const allCurrentNodes = computed(() => {
  const list = []
  list.push(...currentTreeDefs.value.map((item) => ({ ...item, panel: 'main' })))
  if (selectedSubTree.value) {
    list.push(...subTreeDefs.value.map((item) => ({ ...item, panel: 'sub' })))
  }
  return list
})

const selectedNode = computed(() => allCurrentNodes.value.find((item) => item.id === selectedTalentID.value) || null)

function findDef(id) {
  if (!treeDefs.value?.trees) return null
  for (const key of ['normal', 'armor', 'crit']) {
    const found = treeDefs.value.trees[key]?.talents?.find((item) => item.id === id)
    if (found) return found
  }
  return null
}

function isLearned(id) {
  return learnedSet.value.has(id)
}

function isPrerequisiteMet(def) {
  if (!def?.prerequisite) return true
  return learnedSet.value.has(def.prerequisite)
}

function subTreeLearnedCount() {
  if (!talentState.value?.subTree) return 0
  return (talentState.value.talents || []).filter((id) => {
    const d = findDef(id)
    return d && d.tree === talentState.value.subTree
  }).length
}

function canLearn(def) {
  if (!def || !talentState.value?.tree) return false
  if (isLearned(def.id)) return false
  if (!isPrerequisiteMet(def)) return false
  if (Number(def.cost || 0) <= 0) return false
  if (availableTalentPoints.value < Number(def.cost || 0)) return false

  if (def.tree === talentState.value.tree) return true
  if (def.tree === talentState.value.subTree) {
    if (def.tier === 0 || def.tier === 4) return false
    return subTreeLearnedCount() < 2
  }
  return false
}

function nodeState(def) {
  if (!def) return 'locked'
  if (isLearned(def.id)) return 'learned'
  if (!talentState.value?.tree) return 'locked'
  if (!isPrerequisiteMet(def)) return 'locked'
  if (def.tree !== talentState.value.tree && def.tree !== talentState.value.subTree) return 'locked'
  if (def.tree === talentState.value.subTree && (def.tier === 0 || def.tier === 4)) return 'locked'
  if (def.tree === talentState.value.subTree && subTreeLearnedCount() >= 2) return 'locked'
  if (availableTalentPoints.value < Number(def.cost || 0)) return 'insufficient'
  return 'available'
}

function stateLabel(def) {
  const state = nodeState(def)
  if (state === 'learned') return '已学'
  if (state === 'available') return '可学'
  if (state === 'insufficient') return '点数不足'
  return '锁定'
}

function stateReason(def) {
  if (!def) return ''
  const state = nodeState(def)
  if (state === 'learned') return '该节点已学习。'
  if (state === 'available') return '满足条件，点击即可学习。'
  if (state === 'insufficient') return '当前天赋点不足。'
  if (!talentState.value?.tree) return '请先选择主系。'
  if (!isPrerequisiteMet(def)) return `需要先学习前置节点：${def.prerequisite}`
  return '当前条件不满足。'
}

function nodeCoordinatesByTier(defs, tier) {
  const points = []
  const total = defs.length
  if (total <= 0) return points
  const start = 195
  const end = -15
  const step = total === 1 ? 0 : (start - end) / (total - 1)
  const radius = tierRadius[tier] || 120
  for (let i = 0; i < total; i += 1) {
    const angle = start - step * i
    points.push({
      radius,
      angle,
    })
  }
  return points
}

function mapRingNodes(defs) {
  const byTier = new Map()
  for (const def of defs) {
    const tier = Number(def.tier || 0)
    if (!byTier.has(tier)) byTier.set(tier, [])
    byTier.get(tier).push(def)
  }
  const result = []
  for (const tier of [0, 1, 2, 3, 4]) {
    const tierDefs = byTier.get(tier) || []
    const points = nodeCoordinatesByTier(tierDefs, tier)
    tierDefs.forEach((def, idx) => {
      result.push({
        ...def,
        angle: points[idx]?.angle || 0,
        radius: points[idx]?.radius || 120,
      })
    })
  }
  return result
}

const mainRingNodes = computed(() => mapRingNodes(currentTreeDefs.value))
const subRingNodes = computed(() => mapRingNodes(subTreeDefs.value))

function nodeStyle(item) {
  return {
    '--node-angle': `${item.angle}deg`,
    '--node-radius': `${item.radius}px`,
  }
}

async function loadDefs() {
  try {
    const res = await fetch('/api/talents/defs')
    if (!res.ok) throw new Error('加载天赋定义失败')
    treeDefs.value = await res.json()
  } catch (error) {
    errorMsg.value = error.message || '加载天赋定义失败'
  }
}

async function loadState() {
  if (!isLoggedIn.value) return
  try {
    const res = await fetch('/api/talents/state', { credentials: 'include' })
    if (!res.ok) {
      if (res.status !== 401) {
        const payload = await safeJSON(res)
        errorMsg.value = payload?.message || '加载天赋状态失败'
      }
      return
    }
    talentState.value = await res.json()
    if (talentState.value?.tree) selectedTree.value = talentState.value.tree
    if (talentState.value?.subTree) selectedSubTree.value = talentState.value.subTree
  } catch (error) {
    errorMsg.value = error.message || '加载天赋状态失败'
  }
}

async function selectTree(tree) {
  if (!isLoggedIn.value) return
  loading.value = true
  errorMsg.value = ''
  try {
    const body = { tree }
    if (selectedSubTree.value && selectedSubTree.value !== tree) {
      body.subTree = selectedSubTree.value
    }
    const res = await fetch('/api/talents/select', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    })
    if (!res.ok) {
      const payload = await safeJSON(res)
      throw new Error(payload?.message || '选择主系失败')
    }
    selectedTree.value = tree
    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '选择主系失败'
  } finally {
    loading.value = false
  }
}

async function selectSubTree(tree) {
  if (!isLoggedIn.value) return
  loading.value = true
  errorMsg.value = ''
  try {
    const subTree = tree === selectedTree.value ? '' : tree
    const res = await fetch('/api/talents/select', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ tree: selectedTree.value, subTree }),
    })
    if (!res.ok) {
      const payload = await safeJSON(res)
      throw new Error(payload?.message || '选择副系失败')
    }
    selectedSubTree.value = subTree
    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '选择副系失败'
  } finally {
    loading.value = false
  }
}

async function learnTalent(def) {
  if (!isLoggedIn.value || learnLoading.value || !def || !canLearn(def)) return
  selectedTalentID.value = def.id
  learnLoading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/talents/learn', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ talentId: def.id }),
    })
    if (!res.ok) {
      const payload = await safeJSON(res)
      throw new Error(payload?.message || '学习失败')
    }
    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '学习失败'
  } finally {
    learnLoading.value = false
  }
}

async function resetTalents() {
  if (!isLoggedIn.value) return
  if (!window.confirm('确定洗点并返还已消耗天赋点吗？')) return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/talents/reset', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) {
      const payload = await safeJSON(res)
      throw new Error(payload?.message || '洗点失败')
    }
    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '洗点失败'
  } finally {
    loading.value = false
  }
}

function handleNodeClick(def) {
  selectedTalentID.value = def.id
  if (canLearn(def)) {
    void learnTalent(def)
  }
}

function safeJSON(response) {
  return response.json().catch(() => null)
}

onMounted(() => {
  void loadDefs()
  void loadState()
})
</script>

<template>
  <section class="talent-page">
    <header class="talent-head">
      <div>
        <p class="vote-stage__eyebrow">天赋系统</p>
        <h2>半圆盘天赋树</h2>
      </div>
      <div class="talent-points">
        <span>当前天赋点</span>
        <strong>{{ availableTalentPoints }}</strong>
      </div>
    </header>

    <p v-if="errorMsg" class="feedback feedback--error">{{ errorMsg }}</p>

    <div v-if="!isLoggedIn" class="feedback-panel">
      <p>请先登录后再配置天赋。</p>
    </div>

    <template v-else>
      <section class="talent-select">
        <div class="talent-select__group">
          <span class="talent-select__label">主系</span>
          <div class="talent-select__buttons">
            <button
              v-for="tree in trees"
              :key="`main-${tree}`"
              class="talent-select__btn"
              :class="{ 'talent-select__btn--active': selectedTree === tree }"
              :style="{ '--tree-color': treeConfig[tree].color }"
              :disabled="loading"
              @click="selectTree(tree)"
            >
              {{ treeConfig[tree].name }}
            </button>
          </div>
        </div>

        <div class="talent-select__group">
          <span class="talent-select__label">副系</span>
          <div class="talent-select__buttons">
            <button
              v-for="tree in trees"
              :key="`sub-${tree}`"
              class="talent-select__btn talent-select__btn--sub"
              :class="{ 'talent-select__btn--active': selectedSubTree === tree }"
              :style="{ '--tree-color': treeConfig[tree].color }"
              :disabled="loading || tree === selectedTree"
              @click="selectSubTree(tree)"
            >
              {{ treeConfig[tree].name }}
            </button>
            <button
              class="talent-select__btn talent-select__btn--sub"
              :class="{ 'talent-select__btn--active': selectedSubTree === '' }"
              :disabled="loading"
              @click="selectSubTree('')"
            >
              无
            </button>
          </div>
        </div>
      </section>

      <section class="talent-layout">
        <article class="talent-plate">
          <header class="talent-plate__head">
            <strong>主系：{{ treeConfig[selectedTree]?.name }}</strong>
          </header>
          <div class="talent-half-ring" :style="{ '--active-color': treeConfig[selectedTree]?.color }">
            <div
              v-for="item in mainRingNodes"
              :key="`main-${item.id}`"
              class="talent-dot"
              :class="`talent-dot--${nodeState(item)}`"
              :style="nodeStyle(item)"
              @click="handleNodeClick(item)"
            >
              <span class="talent-dot__name">{{ item.name }}</span>
              <span class="talent-dot__cost">{{ item.cost }}</span>
            </div>
            <div class="talent-tier-legend">
              <span v-for="tier in [0, 1, 2, 3, 4]" :key="`main-tier-${tier}`">
                {{ tierLabels[tier] }}
              </span>
            </div>
          </div>
        </article>

        <article v-if="selectedSubTree" class="talent-plate talent-plate--sub">
          <header class="talent-plate__head">
            <strong>副系：{{ treeConfig[selectedSubTree]?.name }}</strong>
          </header>
          <div class="talent-half-ring" :style="{ '--active-color': treeConfig[selectedSubTree]?.color }">
            <div
              v-for="item in subRingNodes"
              :key="`sub-${item.id}`"
              class="talent-dot"
              :class="`talent-dot--${nodeState(item)}`"
              :style="nodeStyle(item)"
              @click="handleNodeClick(item)"
            >
              <span class="talent-dot__name">{{ item.name }}</span>
              <span class="talent-dot__cost">{{ item.cost }}</span>
            </div>
            <div class="talent-tier-legend">
              <span v-for="tier in [0, 1, 2, 3, 4]" :key="`sub-tier-${tier}`">
                {{ tierLabels[tier] }}
              </span>
            </div>
          </div>
        </article>

        <aside class="talent-detail">
          <header class="talent-detail__head">
            <strong>节点详情</strong>
            <button class="nickname-form__ghost" :disabled="loading" @click="resetTalents">洗点返还</button>
          </header>
          <div v-if="!selectedNode" class="talent-detail__empty">点击节点查看详情与消耗。</div>
          <div v-else class="talent-detail__body">
            <h3>{{ selectedNode.name }}</h3>
            <p>层级：{{ tierLabels[selectedNode.tier] }}</p>
            <p>消耗：{{ selectedNode.cost }} 天赋点</p>
            <p>前置：{{ selectedNode.prerequisite || '无' }}</p>
            <p>效果：{{ selectedNode.effectType }}</p>
            <p>状态：{{ stateLabel(selectedNode) }}</p>
            <p class="talent-detail__hint">{{ stateReason(selectedNode) }}</p>
          </div>
        </aside>
      </section>
    </template>
  </section>
</template>

<style scoped>
.talent-page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 1rem;
}

.talent-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.talent-points {
  padding: 0.75rem 1rem;
  border: 1px solid #31434f;
  border-radius: 0.75rem;
  background: #112029;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  min-width: 140px;
}

.talent-points span {
  color: #8ba8b9;
  font-size: 0.8rem;
}

.talent-points strong {
  color: #f8f2d6;
  font-size: 1.35rem;
}

.talent-select {
  margin: 1rem 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.talent-select__group {
  border: 1px solid #2c3e47;
  border-radius: 0.75rem;
  padding: 0.75rem;
  background: #0f1b22;
}

.talent-select__label {
  display: block;
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  color: #7c95a4;
  margin-bottom: 0.5rem;
}

.talent-select__buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.talent-select__btn {
  border: 1px solid #35505c;
  border-radius: 999px;
  background: #132833;
  color: #d7e7ef;
  padding: 0.35rem 0.8rem;
  cursor: pointer;
}

.talent-select__btn--active {
  border-color: var(--tree-color, #2ab06f);
  color: #fff;
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--tree-color, #2ab06f) 40%, transparent);
}

.talent-select__btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.talent-layout {
  display: grid;
  grid-template-columns: 1fr 1fr 300px;
  gap: 0.75rem;
}

.talent-plate {
  border: 1px solid #2c3d48;
  border-radius: 0.9rem;
  background: linear-gradient(180deg, #14252f 0%, #101b22 100%);
  overflow: hidden;
}

.talent-plate__head {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #24333c;
  color: #d9edf8;
}

.talent-half-ring {
  position: relative;
  height: 360px;
  padding: 1rem;
}

.talent-half-ring::before {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 1rem;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--active-color, #2ab06f);
  transform: translateX(-50%);
  box-shadow: 0 0 0 8px color-mix(in srgb, var(--active-color, #2ab06f) 20%, transparent);
}

.talent-dot {
  position: absolute;
  left: 50%;
  bottom: 1rem;
  transform:
    rotate(var(--node-angle))
    translateY(calc(var(--node-radius) * -1))
    rotate(calc(var(--node-angle) * -1));
  width: 92px;
  min-height: 56px;
  border-radius: 0.6rem;
  border: 1px solid #344854;
  background: #0f1820;
  cursor: pointer;
  padding: 0.35rem 0.4rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.12rem;
  text-align: center;
}

.talent-dot__name {
  font-size: 0.7rem;
  color: #dbe8ef;
  line-height: 1.2;
}

.talent-dot__cost {
  font-size: 0.66rem;
  color: #85a4b6;
}

.talent-dot--learned {
  border-color: #4fcf8c;
  background: #1b3a2d;
}

.talent-dot--available {
  border-color: var(--active-color, #2ab06f);
}

.talent-dot--insufficient {
  border-color: #b67c35;
  background: #2a2015;
}

.talent-dot--locked {
  opacity: 0.55;
  cursor: default;
}

.talent-tier-legend {
  position: absolute;
  right: 0.6rem;
  top: 0.8rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  color: #88a3b3;
  font-size: 0.7rem;
}

.talent-detail {
  border: 1px solid #2c3e49;
  border-radius: 0.9rem;
  background: #0f1920;
  padding: 0.9rem;
}

.talent-detail__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.85rem;
}

.talent-detail__empty {
  color: #88a4b4;
  font-size: 0.9rem;
}

.talent-detail__body h3 {
  margin: 0 0 0.55rem 0;
  color: #f2f7fa;
}

.talent-detail__body p {
  margin: 0.35rem 0;
  color: #b8d0dd;
  font-size: 0.85rem;
}

.talent-detail__hint {
  color: #f1cf8e;
}

@media (max-width: 1100px) {
  .talent-layout {
    grid-template-columns: 1fr;
  }

  .talent-select {
    grid-template-columns: 1fr;
  }
}
</style>
