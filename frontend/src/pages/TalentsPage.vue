<script setup>
import { ref, computed, onMounted } from 'vue'
import { usePublicPageState } from './publicPageState'

const { nickname, isLoggedIn, combatStats } = usePublicPageState()

const loading = ref(false)
const errorMsg = ref('')
const treeDefs = ref(null)
const talentState = ref(null)
const selectedTree = ref('normal')
const selectedSubTree = ref('')
const learnLoading = ref(false)

const treeConfig = {
  normal: { name: '均衡攻势', color: '#4ade80' },
  armor: { name: '碎盾攻坚', color: '#fbbf24' },
  crit: { name: '致命洞察', color: '#f472b6' },
}

const tiers = [0, 1, 2, 3, 4]
const tierLabels = { 0: '基石', 1: '第一层', 2: '第二层', 3: '第三层', 4: '终极' }

const currentTreeDefs = computed(() => {
  if (!treeDefs.value) return []
  const tree = treeDefs.value.trees
  switch (selectedTree.value) {
    case 'normal': return tree.normal?.talents || []
    case 'armor': return tree.armor?.talents || []
    case 'crit': return tree.crit?.talents || []
    default: return []
  }
})

const subTreeDefs = computed(() => {
  if (!selectedSubTree.value || !treeDefs.value) return []
  const tree = treeDefs.value.trees
  switch (selectedSubTree.value) {
    case 'normal': return tree.normal?.talents || []
    case 'armor': return tree.armor?.talents || []
    case 'crit': return tree.crit?.talents || []
    default: return []
  }
})

const learnedSet = computed(() => new Set(talentState.value?.talents || []))

function isLearned(id) {
  return learnedSet.value.has(id)
}

function isPrerequisiteMet(def, learned) {
  if (!def.prerequisite) return true
  return learned.has(def.prerequisite)
}

function canLearn(def) {
  if (!talentState.value) return false
  if (!talentState.value.tree) return false
  if (isLearned(def.id)) return false
  if (!isPrerequisiteMet(def, learnedSet.value)) return false

  if (def.tree === talentState.value.tree) return true
  if (def.tree === talentState.value.subTree) {
    if (def.tier === 0 || def.tier === 4) return false
    const subCount = (talentState.value.talents || []).filter(id => {
      const d = treeDefs.value ? findDef(id) : null
      return d && d.tree === talentState.value.subTree
    }).length
    return subCount < 2
  }
  return false
}

function findDef(id) {
  if (!treeDefs.value) return null
  for (const key of ['normal', 'armor', 'crit']) {
    const found = treeDefs.value.trees[key]?.talents?.find(t => t.id === id)
    if (found) return found
  }
  return null
}

const trees = ['normal', 'armor', 'crit']

async function loadDefs() {
  try {
    const res = await fetch('/api/talents/defs')
    if (!res.ok) throw new Error('加载天赋定义失败')
    treeDefs.value = await res.json()
  } catch (e) {
    errorMsg.value = e.message
  }
}

async function loadState() {
  if (!isLoggedIn.value) return
  try {
    const res = await fetch('/api/talents/state', { credentials: 'include' })
    if (!res.ok) {
      if (res.status !== 401) errorMsg.value = '加载天赋状态失败'
      return
    }
    talentState.value = await res.json()
    if (talentState.value?.tree) {
      selectedTree.value = talentState.value.tree
    }
    if (talentState.value?.subTree) {
      selectedSubTree.value = talentState.value.subTree
    }
  } catch (e) {
    errorMsg.value = e.message
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
    if (!res.ok) throw new Error('选择天赋树失败')
    selectedTree.value = tree
    await loadState()
  } catch (e) {
    errorMsg.value = e.message
  } finally {
    loading.value = false
  }
}

async function selectSubTree(tree) {
  if (!isLoggedIn.value) return
  if (tree === selectedTree.value) {
    selectedSubTree.value = ''
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/talents/select', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ tree: selectedTree.value, subTree: tree }),
    })
    if (!res.ok) throw new Error('选择副系失败')
    selectedSubTree.value = tree
    await loadState()
  } catch (e) {
    errorMsg.value = e.message
  } finally {
    loading.value = false
  }
}

async function learnTalent(talentId) {
  if (!isLoggedIn.value || learnLoading.value) return
  learnLoading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/talents/learn', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ talentId }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || '学习天赋失败')
    }
    await loadState()
  } catch (e) {
    errorMsg.value = e.message
  } finally {
    learnLoading.value = false
  }
}

async function resetTalents() {
  if (!isLoggedIn.value || !confirm('确定重置所有已学天赋？（主系副系保留）')) return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/talents/reset', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) throw new Error('重置天赋失败')
    await loadState()
  } catch (e) {
    errorMsg.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDefs()
  loadState()
})
</script>

<template>
  <section class="talent-page">
    <div class="vote-stage__head">
      <div>
        <p class="vote-stage__eyebrow">天赋系统</p>
        <h2>选择天赋系</h2>
        <p class="vote-stage__hint">
          主系可学习全部层级，副系最多学习 2 个中间节点。
        </p>
      </div>
    </div>

    <p v-if="errorMsg" class="feedback feedback--error">{{ errorMsg }}</p>

    <div v-if="!isLoggedIn" class="feedback-panel">
      <p>请先登录后再配置天赋。</p>
    </div>

    <template v-else>
      <div class="talent-tree-select">
        <div class="talent-tree-select__group">
          <label class="talent-tree-select__label">主系</label>
          <div class="talent-tree-select__buttons">
            <button
              v-for="t in trees"
              :key="'main-' + t"
              class="talent-tree-btn"
              :class="{ 'talent-tree-btn--active': selectedTree === t }"
              :style="{ '--tree-color': treeConfig[t].color }"
              @click="selectTree(t)"
              :disabled="loading"
            >
              {{ treeConfig[t].name }}
            </button>
          </div>
        </div>

        <div class="talent-tree-select__group">
          <label class="talent-tree-select__label">副系</label>
          <div class="talent-tree-select__buttons">
            <button
              v-for="t in trees"
              :key="'sub-' + t"
              class="talent-tree-btn talent-tree-btn--sub"
              :class="{ 'talent-tree-btn--active': selectedSubTree === t }"
              :style="{ '--tree-color': treeConfig[t].color }"
              @click="selectSubTree(t)"
              :disabled="loading || t === selectedTree"
            >
              {{ treeConfig[t].name }}
            </button>
            <button
              class="talent-tree-btn talent-tree-btn--sub talent-tree-btn--none"
              :class="{ 'talent-tree-btn--active': !selectedSubTree }"
              @click="selectedSubTree = ''"
              :disabled="loading"
            >
              无
            </button>
          </div>
        </div>
      </div>

      <div class="talent-reset-bar">
        <button class="nickname-form__ghost" @click="resetTalents" :disabled="loading">
          重置天赋
        </button>
      </div>

      <div v-if="talentState?.tree" class="talent-tree">
        <div v-for="tier in tiers" :key="tier" class="talent-tier">
          <div class="talent-tier__label">{{ tierLabels[tier] }}</div>
          <div class="talent-tier__nodes">
            <div
              v-for="def in currentTreeDefs.filter(d => d.tier === tier)"
              :key="def.id"
              class="talent-node"
              :class="{
                'talent-node--learned': isLearned(def.id),
                'talent-node--available': canLearn(def),
                'talent-node--locked': !isLearned(def.id) && !canLearn(def),
              }"
              :style="{ '--node-color': treeConfig[selectedTree]?.color }"
              @click="canLearn(def) && learnTalent(def.id)"
            >
              <div class="talent-node__icon">
                <span v-if="isLearned(def.id)">✓</span>
                <span v-else-if="canLearn(def.id)">+</span>
                <span v-else>🔒</span>
              </div>
              <div class="talent-node__info">
                <strong class="talent-node__name">{{ def.name }}</strong>
                <span class="talent-node__effect">{{ def.effectType }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="selectedSubTree && talentState?.subTree" class="talent-tree talent-tree--sub">
        <h3 class="talent-tree__sub-title">副系 —— {{ treeConfig[selectedSubTree]?.name }}</h3>
        <div v-for="tier in tiers" :key="'sub-' + tier" class="talent-tier">
          <div class="talent-tier__label">{{ tierLabels[tier] }}</div>
          <div class="talent-tier__nodes">
            <div
              v-for="def in subTreeDefs.filter(d => d.tier === tier)"
              :key="def.id"
              class="talent-node"
              :class="{
                'talent-node--learned': isLearned(def.id),
                'talent-node--available': canLearn(def),
                'talent-node--locked': !isLearned(def.id) && !canLearn(def),
              }"
              :style="{ '--node-color': treeConfig[selectedSubTree]?.color }"
              @click="canLearn(def) && learnTalent(def.id)"
            >
              <div class="talent-node__icon">
                <span v-if="isLearned(def.id)">✓</span>
                <span v-else-if="canLearn(def.id)">+</span>
                <span v-else>🔒</span>
              </div>
              <div class="talent-node__info">
                <strong class="talent-node__name">{{ def.name }}</strong>
                <span class="talent-node__effect">{{ def.effectType }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.talent-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem;
}

.talent-tree-select {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1rem;
}

.talent-tree-select__group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.talent-tree-select__label {
  font-size: 0.85rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.talent-tree-select__buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.talent-tree-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #334155;
  border-radius: 0.5rem;
  background: #1e293b;
  color: #e2e8f0;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.15s;
}

.talent-tree-btn:hover:not(:disabled) {
  border-color: var(--tree-color, #4ade80);
}

.talent-tree-btn--active {
  border-color: var(--tree-color, #4ade80);
  background: color-mix(in srgb, var(--tree-color, #4ade80) 15%, #1e293b);
}

.talent-tree-btn--none {
  --tree-color: #64748b;
}

.talent-tree-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.talent-reset-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
}

.talent-tree {
  margin-bottom: 2rem;
}

.talent-tree__sub-title {
  color: #94a3b8;
  margin-bottom: 1rem;
  font-size: 1rem;
}

.talent-tier {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.75rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
  background: #1e293b;
}

.talent-tier__label {
  min-width: 4rem;
  font-size: 0.8rem;
  color: #64748b;
  padding-top: 0.5rem;
}

.talent-tier__nodes {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  flex: 1;
}

.talent-node {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid #334155;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.15s;
  flex: 1;
  min-width: 140px;
  background: #0f172a;
}

.talent-node--learned {
  border-color: var(--node-color, #4ade80);
  background: color-mix(in srgb, var(--node-color, #4ade80) 10%, #0f172a);
}

.talent-node--available {
  border-color: #334155;
}

.talent-node--available:hover {
  border-color: var(--node-color, #4ade80);
  background: color-mix(in srgb, var(--node-color, #4ade80) 5%, #0f172a);
}

.talent-node--locked {
  opacity: 0.5;
  cursor: not-allowed;
}

.talent-node__icon {
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 0.75rem;
  background: #1e293b;
  flex-shrink: 0;
}

.talent-node--learned .talent-node__icon {
  background: var(--node-color, #4ade80);
  color: #000;
}

.talent-node__info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.talent-node__name {
  font-size: 0.85rem;
  color: #e2e8f0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.talent-node__effect {
  font-size: 0.7rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
