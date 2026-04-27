<script setup>
import { computed, onMounted, ref } from 'vue'
import { usePublicPageState } from './publicPageState'
import { effectAssetUrl } from '../utils/effectAssets'

const { isLoggedIn, talentPoints: sharedTalentPoints } = usePublicPageState()

const loading = ref(false)
const learnLoading = ref(false)
const errorMsg = ref('')
const treeDefs = ref(null)
const talentState = ref(null)
const selectedTree = ref('normal')
const selectedSubTree = ref('')
const activePanel = ref('main')
const selectedMarker = ref({ panel: '', id: '' })

const treeConfig = {
  normal: { name: '均衡攻势', color: '#2bb873' },
  armor: { name: '碎盾攻坚', color: '#c48a33' },
  crit: { name: '致命洞察', color: '#ca3e59' },
}

const tierLabels = { 0: '基石', 1: '一阶', 2: '二阶', 3: '三阶', 4: '终极' }
const tierRadiusPercent = { 0: 14, 1: 28, 2: 42, 3: 56, 4: 70 }
const trees = ['normal', 'armor', 'crit']
const arcStartAngle = 135
const arcEndAngle = 45

const learnedSet = computed(() => new Set(talentState.value?.talents || []))

// 小节点识别（ID 包含 _t1a/_t1b/_t2a/_t2b/_t3a/_t3b）
function isFillerNode(id) {
  if (!id) return false
  return /_(t[1-3][ab])$/.test(id)
}

// 将天赋 ID 映射到像素图标文件名
function talentIconPath(id) {
  if (!id) return ''
  // 小节点：normal_filler_t1a → talent-normal-t1a.png
  if (isFillerNode(id)) {
    const parts = id.split('_')
    return effectAssetUrl(`talent-${parts[0]}-${parts[2]}.png`)
  }
  // 主节点：talent-{id}.png
  return effectAssetUrl(`talent-${id}.png`)
}

// 每层节点总数（主 + 小），与后端 tierNodeCount 一致
const tierNodeCount = { 0: 1, 1: 5, 2: 5, 3: 4, 4: 1 }

// 当前主系某层已学节点数
function learnedInTierCount(tree, tier) {
  return (talentState.value?.talents || []).filter((id) => {
    const def = findDef(id)
    return def && def.tree === tree && def.tier === tier
  }).length
}

function isTierFull(tree, tier) {
  return learnedInTierCount(tree, tier) >= (tierNodeCount[tier] || 0)
}

// 某层是否前置层未满（主系层锁）
function isLayerLocked(def) {
  if (!def || !talentState.value?.tree) return false
  if (def.tree !== talentState.value.tree) return false
  if (def.tier === 0) return false
  return !isTierFull(talentState.value.tree, def.tier - 1)
}

const activeTierBonuses = computed(() => {
  if (!talentState.value?.tree) return []
  const tree = talentState.value.tree
  const bonusLabels = treeDefs.value?.trees?.[tree]?.tierCompletionBonuses || {}
  const bonuses = []
  for (let tier = 0; tier <= 4; tier++) {
    if (isTierFull(tree, tier)) {
      bonuses.push({
        tier,
        label: bonusLabels[tier] || `第 ${tier + 1} 层奖励`,
      })
    }
  }
  return bonuses
})

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

function safeJSON(response) {
  return response.json().catch(() => null)
}

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
    const def = findDef(id)
    return def && def.tree === talentState.value.subTree
  }).length
}

function canLearn(def) {
  if (!def || !talentState.value?.tree) return false
  if (isLearned(def.id)) return false
  if (!isPrerequisiteMet(def)) return false
  if (isLayerLocked(def)) return false

  const cost = Number(def.cost || 0)
  if (cost <= 0 || availableTalentPoints.value < cost) return false

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
  if (isLayerLocked(def)) return 'layer-locked'

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
  if (state === 'layer-locked') return '需先点满下层'
  if (state === 'insufficient') return '点数不足'
  return '锁定'
}

function prerequisiteLabel(def) {
  if (def?.prerequisiteName) return def.prerequisiteName
  if (!def?.prerequisite) return '无'
  const pre = findDef(def.prerequisite)
  return pre?.name || def.prerequisite
}

function effectDescription(def) {
  if (def?.effectDescription) return def.effectDescription
  return '效果说明待配置'
}

function stateReason(def) {
  if (!def) return ''
  const state = nodeState(def)
  if (state === 'learned') return '该节点已学习。'
  if (state === 'available') return '满足条件，点击即可学习。'
  if (state === 'layer-locked') return `需先点满第 ${def.tier - 1 + 1} 层所有节点（含小节点）。`
  if (state === 'insufficient') return `当前天赋点不足，需要 ${def.cost} 点。`
  if (!talentState.value?.tree) return '请先选择主系。'
  if (!isPrerequisiteMet(def)) return `前置未满足：${prerequisiteLabel(def)}`
  return '当前条件不满足。'
}

function coordinatesForTierNodes(nodes, tier) {
  if (nodes.length === 0) return []
  const radius = tierRadiusPercent[tier] || 20
  const centerAngle = (arcStartAngle + arcEndAngle) / 2

  // 基石层和终极层强制居中，满足“第一层中间”和“最后一层中间”布局要求。
  if (tier === 0 || tier === 4) {
    const gap = 10
    const middle = (nodes.length - 1) / 2
    return nodes.map((_, index) => {
      const angle = centerAngle + (middle - index) * gap
      const radian = (angle * Math.PI) / 180
      const x = Number((Math.cos(radian) * radius).toFixed(2))
      const y = Number((Math.sin(radian) * radius).toFixed(2))
      return {
        leftPercent: 50 + x,
        topPercent: 90 - y,
      }
    })
  }

  const step = nodes.length === 1 ? 0 : (arcStartAngle - arcEndAngle) / (nodes.length - 1)

  return nodes.map((_, index) => {
    const angle = arcStartAngle - step * index
    const radian = (angle * Math.PI) / 180
    const x = Number((Math.cos(radian) * radius).toFixed(2))
    const y = Number((Math.sin(radian) * radius).toFixed(2))
    return {
      leftPercent: 50 + x,
      topPercent: 90 - y,
    }
  })
}

function toPolarPoint(radius, angle) {
  const radian = (angle * Math.PI) / 180
  return {
    x: Number((50 + Math.cos(radian) * radius).toFixed(3)),
    y: Number((90 - Math.sin(radian) * radius).toFixed(3)),
  }
}

function arcPolylinePath(radius, segments = 30) {
  const points = []
  const step = (arcStartAngle - arcEndAngle) / segments
  for (let i = 0; i <= segments; i += 1) {
    const angle = arcStartAngle - step * i
    points.push(toPolarPoint(radius, angle))
  }
  if (points.length === 0) return ''
  return points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`).join(' ')
}

const arcGridPaths = computed(() => [0, 1, 2, 3, 4].map((tier) => ({
  tier,
  path: arcPolylinePath(tierRadiusPercent[tier] || 20),
})))

function mapRingNodes(defs, panel) {
  const byTier = new Map()
  for (const def of defs) {
    const tier = Number(def.tier || 0)
    if (!byTier.has(tier)) byTier.set(tier, [])
    byTier.get(tier).push(def)
  }

  const result = []
  for (const tier of [0, 1, 2, 3, 4]) {
    const tierDefs = byTier.get(tier) || []
    const coords = coordinatesForTierNodes(tierDefs, tier)
    tierDefs.forEach((def, index) => {
      result.push({
        ...def,
        panel,
        leftPercent: coords[index]?.leftPercent || 50,
        topPercent: coords[index]?.topPercent || 90,
      })
    })
  }
  return result
}

const mainRingNodes = computed(() => mapRingNodes(currentTreeDefs.value, 'main'))
const subRingNodes = computed(() => mapRingNodes(subTreeDefs.value, 'sub'))
const activeTree = computed(() => (activePanel.value === 'sub' && selectedSubTree.value ? selectedSubTree.value : selectedTree.value))
const activeRingNodes = computed(() => (activePanel.value === 'sub' ? subRingNodes.value : mainRingNodes.value))

const selectedNode = computed(() => {
  const targetPanel = selectedMarker.value.panel
  const targetID = selectedMarker.value.id
  if (!targetPanel || !targetID) return null

  const list = targetPanel === 'sub' ? subRingNodes.value : mainRingNodes.value
  return list.find((item) => item.id === targetID) || null
})

function ringNodeStyle(item) {
  return {
    left: `${item.leftPercent}%`,
    top: `${item.topPercent}%`,
  }
}

function detailFloatStyle(panel) {
  if (!selectedNode.value || selectedNode.value.panel !== panel) return {}
  const left = Math.max(8, Math.min(66, selectedNode.value.leftPercent + 8))
  const top = Math.max(6, Math.min(68, selectedNode.value.topPercent - 20))
  return {
    left: `${left}%`,
    top: `${top}%`,
  }
}

const activePlateTitle = computed(() => {
  if (activePanel.value === 'sub' && selectedSubTree.value) {
    return `副系：${treeConfig[selectedSubTree.value]?.name || ''}`
  }
  return `主系：${treeConfig[selectedTree.value]?.name || ''}`
})

const activePlateHint = computed(() => {
  if (activePanel.value === 'sub' && selectedSubTree.value) {
    return '副系最多学习 2 个中层节点'
  }
  return '点击节点可学习，或查看浮层说明'
})

function selectNode(def) {
  selectedMarker.value = { panel: def.panel, id: def.id }
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
    selectedSubTree.value = talentState.value?.subTree || ''
    if (!talentState.value?.subTree && activePanel.value === 'sub') activePanel.value = 'main'
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
    activePanel.value = 'main'
    selectedMarker.value = { panel: '', id: '' }
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
    activePanel.value = subTree ? 'sub' : 'main'
    if (!subTree && selectedMarker.value.panel === 'sub') {
      selectedMarker.value = { panel: '', id: '' }
    }
    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '选择副系失败'
  } finally {
    loading.value = false
  }
}

async function learnTalent(def) {
  if (!isLoggedIn.value || learnLoading.value || !def || !canLearn(def)) return
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
      throw new Error(payload?.message || '学习天赋失败')
    }

    await loadState()
  } catch (error) {
    errorMsg.value = error.message || '学习天赋失败'
  } finally {
    learnLoading.value = false
  }
}

function clickNode(def) {
  selectNode(def)
  if (canLearn(def)) {
    void learnTalent(def)
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
      <div class="talent-head__actions">
        <div class="talent-points">
          <span>当前天赋点</span>
          <strong>{{ availableTalentPoints }}</strong>
        </div>
        <button class="nickname-form__ghost" :disabled="loading" @click="resetTalents">洗点返还</button>
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
              class="talent-select__btn"
              :class="{ 'talent-select__btn--active': selectedSubTree === tree }"
              :style="{ '--tree-color': treeConfig[tree].color }"
              :disabled="loading || tree === selectedTree"
              @click="selectSubTree(tree)"
            >
              {{ treeConfig[tree].name }}
            </button>
            <button
              class="talent-select__btn"
              :class="{ 'talent-select__btn--active': selectedSubTree === '' }"
              :disabled="loading"
              @click="selectSubTree('')"
            >
              无
            </button>
          </div>
        </div>
      </section>

      <article class="talent-plate">
        <header class="talent-plate__head">
          <strong>{{ activePlateTitle }}</strong>
          <span>{{ activePlateHint }}</span>
        </header>

        <div class="talent-half-ring" :style="{ '--active-color': treeConfig[activeTree]?.color }">
          <!-- 层满奖励 -->
          <div v-if="activePanel === 'main' && activeTierBonuses.length > 0" class="talent-tier-bonuses">
            <span
              v-for="b in activeTierBonuses"
              :key="`bonus-${b.tier}`"
              class="talent-tier-bonuses__badge"
            >
              第 {{ b.tier + 1 }} 层满：{{ b.label }}
            </span>
          </div>

          <svg class="talent-arc-grid" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
            <path
              v-for="entry in arcGridPaths"
              :key="`main-path-${entry.tier}`"
              :d="entry.path"
              class="talent-arc-grid__path"
            />
          </svg>

          <button
            v-for="item in activeRingNodes"
            :key="`${activePanel}-${item.id}`"
            class="talent-dot"
            :class="[
              `talent-dot--${nodeState(item)}`,
              { 'talent-dot--selected': selectedMarker.panel === activePanel && selectedMarker.id === item.id },
              { 'talent-dot--filler': isFillerNode(item.id) },
            ]"
            :style="ringNodeStyle(item)"
            @click="clickNode(item)"
          >
            <img
              v-if="talentIconPath(item.id)"
              :src="talentIconPath(item.id)"
              class="talent-dot__icon"
              :class="{ 'talent-dot__icon--filler': isFillerNode(item.id) }"
              alt=""
            />
            <span class="talent-dot__name">{{ item.name }}</span>
            <span class="talent-dot__meta">{{ tierLabels[item.tier] }} · {{ item.cost }}</span>
          </button>

          <div v-if="selectedNode && selectedNode.panel === activePanel" class="talent-float" :style="detailFloatStyle(activePanel)">
            <strong>{{ selectedNode.name }}</strong>
            <p>状态：{{ stateLabel(selectedNode) }}</p>
            <p>消耗：{{ selectedNode.cost }} 天赋点</p>
            <p>前置：{{ prerequisiteLabel(selectedNode) }}</p>
            <p>效果：{{ effectDescription(selectedNode) }}</p>
            <p class="talent-float__hint">{{ stateReason(selectedNode) }}</p>
          </div>

          <div class="talent-tier-legend">
            <span v-for="tier in [0, 1, 2, 3, 4]" :key="`legend-main-${tier}`">{{ tierLabels[tier] }}</span>
          </div>
        </div>
      </article>
    </template>
  </section>
</template>

<style scoped>
.talent-page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 0.75rem;
}

.talent-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.talent-head__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.talent-points {
  min-width: 140px;
  padding: 0.58rem 0.85rem;
  border: 1px solid #2f454f;
  border-radius: 0.75rem;
  background: #102029;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.talent-points span {
  color: #8ea6b6;
  font-size: 0.8rem;
}

.talent-points strong {
  color: #f9efcd;
  font-size: 1.2rem;
}

.talent-select {
  margin: 0.75rem 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.6rem;
}

.talent-select__group {
  border: 1px solid #2d3f48;
  border-radius: 0.75rem;
  background: #0f1c23;
  padding: 0.62rem;
}

.talent-select__label {
  display: block;
  margin-bottom: 0.4rem;
  color: #7f99aa;
  font-size: 0.76rem;
  letter-spacing: 0.05em;
}

.talent-select__buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.talent-select__btn {
  border: 1px solid #35515d;
  border-radius: 999px;
  background: #132933;
  color: #dbe9f1;
  cursor: pointer;
  padding: 0.32rem 0.76rem;
  font-size: 0.84rem;
  transition: border-color 0.2s ease;
}

.talent-select__btn--active {
  border-color: var(--tree-color, #2bb873);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--tree-color, #2bb873) 40%, transparent);
}

.talent-select__btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.talent-plate {
  border: 1px solid #2c404a;
  border-radius: 0.8rem;
  background: linear-gradient(180deg, #142630 0%, #101b22 100%);
  overflow: hidden;
  margin-bottom: 0.75rem;
}

.talent-plate__head {
  padding: 0.62rem 0.82rem;
  border-bottom: 1px solid #243640;
  color: #d5ebf7;
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
  font-size: 0.8rem;
}

.talent-half-ring {
  position: relative;
  min-height: 600px;
  padding: 0.72rem;
}

.talent-arc-grid {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.talent-arc-grid__path {
  fill: none;
  stroke: color-mix(in srgb, var(--active-color, #2bb873) 50%, #304752);
  stroke-width: 0.25;
  stroke-dasharray: 1.2 0.9;
  opacity: 0.5;
}

.talent-dot {
  position: absolute;
  width: 76px;
  height: 76px;
  transform: translate(-50%, -50%);
  border: 1px solid #36505b;
  border-radius: 50%;
  background: #0f1a22;
  padding: 0.28rem;
  text-align: center;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.14rem;
}

.talent-dot__icon {
  width: 40px;
  height: 40px;
  image-rendering: pixelated;
  image-rendering: crisp-edges;
  object-fit: contain;
}

.talent-dot__icon--filler {
  width: 30px;
  height: 30px;
}

.talent-dot__name {
  color: #e6f2f8;
  font-size: 0.56rem;
  line-height: 1.18;
  max-width: 64px;
}

.talent-dot__meta {
  color: #8faabb;
  font-size: 0.52rem;
}

.talent-dot--learned {
  border-color: #4ece8e;
  background: #1a3d2d;
}

.talent-dot--available {
  border-color: var(--active-color, #2bb873);
}

.talent-dot--insufficient {
  border-color: #b7813c;
  background: #2a1f16;
}

.talent-dot--locked {
  opacity: 0.52;
}

.talent-dot--layer-locked {
  opacity: 0.45;
  border-color: #4a3a5a;
  background: #1a1222;
}

.talent-dot--filler {
  width: 62px;
  height: 62px;
  border-radius: 50%;
  border-style: dashed;
  opacity: 0.82;
}

.talent-dot--filler .talent-dot__name {
  font-size: 0.5rem;
  max-width: 52px;
}

.talent-dot--filler .talent-dot__meta {
  font-size: 0.46rem;
}

.talent-tier-bonuses {
  position: absolute;
  left: 0.8rem;
  bottom: 0.9rem;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  z-index: 2;
}

.talent-tier-bonuses__badge {
  font-size: 0.62rem;
  color: #f9efcd;
  background: #1a3d2d;
  border: 1px solid #4ece8e;
  border-radius: 999px;
  padding: 0.12rem 0.52rem;
  white-space: nowrap;
}

.talent-dot--selected {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--active-color, #2bb873) 45%, transparent);
}

.talent-float {
  position: absolute;
  width: min(310px, 80vw);
  border: 1px solid #3a505d;
  border-radius: 0.8rem;
  background: #0c1319;
  padding: 0.64rem;
  z-index: 3;
}

.talent-float strong {
  color: #f4f9fc;
  display: block;
  margin-bottom: 0.35rem;
}

.talent-float p {
  margin: 0.2rem 0;
  color: #bdd1de;
  font-size: 0.74rem;
  line-height: 1.35;
}

.talent-float__hint {
  color: #f0cd92 !important;
}

.talent-tier-legend {
  position: absolute;
  right: 0.8rem;
  top: 0.9rem;
  display: flex;
  flex-direction: column;
  gap: 0.24rem;
  font-size: 0.66rem;
  color: #91abbb;
}

@media (max-width: 980px) {
  .talent-select {
    grid-template-columns: 1fr;
  }

  .talent-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .talent-head__actions {
    width: 100%;
    justify-content: space-between;
  }

  .talent-half-ring {
    min-height: 520px;
  }

  .talent-dot {
    width: 68px;
    height: 68px;
  }
}
</style>
