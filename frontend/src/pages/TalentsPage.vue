<script setup>
import { computed, onMounted, ref, reactive } from 'vue'
import { usePublicPageState } from './publicPageState'
import { effectAssetUrl } from '../utils/effectAssets'
import { watch } from 'vue'


const { isLoggedIn, talentPoints: sharedTalentPoints } = usePublicPageState()

const loading = ref(false)
const learnLoading = ref(false)
const errorMsg = ref('')
const toastMsg = ref('')
let toastTimer = 0

function showToast(msg) {
  toastMsg.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 2500)
}

const treeDefs = ref(null)
const talentState = ref(null)
const talentEffectLines = ref({})
const talentEffectDescriptions = ref({})
const selectedTree = ref('normal')
const selectedMarker = ref({ panel: 'main', id: '' })

const confirmModal = reactive({
  show: false,
  title: '',
  message: '',
  resolve: null,
})

function showConfirm(title, message) {
  return new Promise((resolve) => {
    confirmModal.title = title
    confirmModal.message = message
    confirmModal.resolve = resolve
    confirmModal.show = true
  })
}

function confirmOK() {
  confirmModal.show = false
  if (confirmModal.resolve) confirmModal.resolve(true)
}

function confirmCancel() {
  confirmModal.show = false
  if (confirmModal.resolve) confirmModal.resolve(false)
}

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
const talentCostLevelExponent = 0.85
const talentCostMultiplier = 1.8
const defaultOverflowUpgradeCost = 1000
const overflowBonusLabels = {
  soft_damage: '软组织增伤',
  weak_damage: '弱点增伤',
  heavy_damage: '重甲增伤',
  crit_damage: '暴击伤害',
  attack_power: '攻击力',
  all_damage: '全伤害',
}

const learnedMap = reactive(talentState.value?.talents || {})

function nodeLevel(talentId) {
  return learnedMap[talentId] || 0
}

function isFillerNode(id) {
  if (!id) return false
  return /_(t[1-3][ab])$/.test(id)
}

function talentIconPath(id) {
  if (!id) return ''
  if (isFillerNode(id)) {
    const parts = id.split('_')
    return effectAssetUrl(`talent-${parts[0]}-${parts[2]}.png`)
  }
  return effectAssetUrl(`talent-${id}.png`)
}

function totalNodesInTier(tree, tier) {
  return (treeDefs.value?.trees?.[tree]?.talents || []).filter((def) => def.tier === tier).length
}

function learnedInTierCount(tree, tier) {
  let count = 0
  for (const def of treeDefs.value?.trees?.[tree]?.talents || []) {
    if (def.tier === tier && nodeLevel(def.id) > 0) count++
  }
  return count
}

function isTierFull(tree, tier) {
  const needed = totalNodesInTier(tree, tier)
  if (needed === 0) return false
  return learnedInTierCount(tree, tier) >= needed
}

function isLayerLocked(def) {
  if (!def || !talentState.value) return false
  if (def.tier === 0) return false
  return !isTierFull(def.tree, def.tier - 1)
}

const activeTierBonuses = computed(() => {
  if (!talentState.value) return []
  const tree = selectedTree.value
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

const currentTierCompletionBonuses = computed(() => {
  const bonusLabels = treeDefs.value?.trees?.[selectedTree.value]?.tierCompletionBonuses || {}
  return [0, 1, 2, 3, 4].map((tier) => ({
    tier,
    label: bonusLabels[tier] || `${tierLabels[tier] || `第 ${tier + 1} 层`} 点满后获得额外加成`,
  }))
})

const availableTalentPoints = computed(() => {
  if (typeof talentState.value?.talentPoints === 'number') {
    return Math.max(0, Number(talentState.value.talentPoints))
  }
  return Math.max(0, Number(sharedTalentPoints.value || 0))
})

const overflowLevel = computed(() => Math.max(0, Number(talentState.value?.overflowLevel || 0)))

const overflowUpgradeCost = computed(() => Math.max(0, Number(talentState.value?.overflowUpgradeCost || defaultOverflowUpgradeCost)))

const overflowTotalSpent = computed(() => overflowLevel.value * overflowUpgradeCost.value)

const overflowBonusEntries = computed(() => {
  const bonuses = talentState.value?.overflowBonuses || {}
  return Object.entries(overflowBonusLabels)
    .map(([key, label]) => ({
      key,
      label,
      count: Math.max(0, Number(bonuses[key] || 0)),
    }))
    .filter((item) => item.count > 0)
})

const currentTreeDefs = computed(() => {
  if (!treeDefs.value?.trees) return []
  return treeDefs.value.trees[selectedTree.value]?.talents || []
})

function safeJSON(response) {
  return response.json().catch(() => null)
}

function normalizeTalentResponse(data) {
  return {
    ...(data || {}),
    talents: data?.talents || {},
    overflowBonuses: data?.overflowBonuses || {},
    overflowLevel: Math.max(0, Number(data?.overflowLevel || 0)),
    overflowUpgradeCost: Math.max(0, Number(data?.overflowUpgradeCost || defaultOverflowUpgradeCost)),
  }
}

function applyTalentResponse(data) {
  talentState.value = normalizeTalentResponse(data)
  talentEffectLines.value = talentState.value?.effectLines || {}
  talentEffectDescriptions.value = talentState.value?.effectDescriptions || {}
  Object.keys(learnedMap).forEach(k => delete learnedMap[k])
  if (talentState.value?.talents) {
    Object.assign(learnedMap, talentState.value.talents)
  }
}

function formatOverflowBonus(item) {
  const percent = (item.count * 0.1).toFixed(1)
  return `${item.label}：${item.count} 次（+${percent}%）`
}

function isLearned(id) {
  return nodeLevel(id) > 0
}

function canLearn(def) {
  if (!def) return false
  if (nodeLevel(def.id) >= (def.maxLevel || 5)) return false
  if (isLearned(def.id)) return false
  if (isLayerLocked(def)) return false
  return availableTalentPoints.value >= talentCostForLevel(def, 1)
}

function nodeState(def) {
  const lv = nodeLevel(def.id)
  if (lv >= (def.maxLevel || 5)) return 'maxed'
  if (lv > 0) return 'upgradable'
  if (isLayerLocked(def)) return 'layer-locked'
  if (availableTalentPoints.value < talentCostForLevel(def, 1)) return 'insufficient'
  return 'available'
}

function stateLabel(def) {
  const state = nodeState(def)
  const map = {
    upgradable: '可升级',
    maxed: '已满级',
    available: '可学习',
    insufficient: '天赋点不足',
    'layer-locked': '上一层未点满',
  }
  return map[state] || '未知'
}

function effectDescription(def) {
  return talentEffectDescriptions.value?.[def?.id]
    || def?.effectDescription
    || def?.description
    || def?.effect
    || '暂无效果说明'
}

function effectLines(def, curLv) {
  return talentEffectLines.value?.[def?.id] || []
}

function stateReason(def) {
  const state = nodeState(def)
  if (state === 'upgradable') return '点击可升级'
  if (state === 'maxed') return '已提升至最高等级'
  if (state === 'available') return '点击即可学习'
  if (state === 'insufficient') return '当前天赋点不足'
  if (state === 'layer-locked') return '需要先点满上一层天赋'
  return ''
}

function talentCostForLevel(def, targetLevel) {
  if (!def?.cost || !targetLevel || targetLevel < 1) return 0
  // Tier 0 基石使用指数增长公式，与后端一致
  if (def.tier === 0) {
    return Math.round(def.cost * talentCostMultiplier * Math.pow(3.0, targetLevel - 1))
  }
  return Math.round(def.cost * Math.pow(targetLevel, talentCostLevelExponent) * talentCostMultiplier)
}

function talentCostToUpgrade(def, fromLevel, toLevel) {
  if (!def?.cost || !toLevel || toLevel <= fromLevel) return 0
  let total = 0
  for (let level = fromLevel + 1; level <= toLevel; level += 1) {
    total += talentCostForLevel(def, level)
  }
  return total
}

function upgradeCost(def) {
  const lv = nodeLevel(def.id)
  if (lv >= (def.maxLevel || 5)) return 0
  return talentCostToUpgrade(def, lv, lv + 1)
}

function toPolarPoint(radius, angle) {
  const rad = (angle * Math.PI) / 180
  return {
    x: 50 + radius * Math.cos(rad),
    y: 86 - radius * Math.sin(rad),
  }
}

function coordinatesForTierNodes(defs, tier) {
  const radius = tierRadiusPercent[tier] || 20
  const count = defs.length
  if (count <= 0) return []

  if (count === 1) {
    return [toNodePercent(radius, 90)]
  }

  const start = arcStartAngle
  const end = arcEndAngle
  const step = (start - end) / (count - 1)

  return defs.map((_, index) => {
    const angle = start - step * index
    return toNodePercent(radius, angle)
  })
}

function toNodePercent(radius, angle) {
  const point = toPolarPoint(radius, angle)
  return {
    leftPercent: point.x,
    topPercent: point.y,
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

const selectedNode = computed(() => {
  const targetID = selectedMarker.value.id
  if (!targetID) return null
  return mainRingNodes.value.find((item) => item.id === targetID) || null
})

function ringNodeStyle(item) {
  return {
    left: `${item.leftPercent}%`,
    top: `${item.topPercent}%`,
  }
}

function detailFloatStyle() {
  if (!selectedNode.value) return {}

  const node = selectedNode.value
  let left = node.leftPercent + 8
  let top = node.topPercent - 20

  if (node.leftPercent > 70) {
    left = node.leftPercent - 34
  }

  if (node.leftPercent < 30) {
    left = node.leftPercent + 10
  }

  left = Math.max(4, Math.min(68, left))
  top = Math.max(4, Math.min(72, top))

  return {
    left: `${left}%`,
    top: `${top}%`,
  }
}

const activePlateTitle = computed(() => treeConfig[selectedTree.value]?.name || '')

const activePlateHint = computed(() => '点击节点可学习，或查看浮层说明')

function selectNode(def) {
  selectedMarker.value = { panel: 'main', id: def.id }
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

    applyTalentResponse(await res.json())
    if (!selectedTree.value) selectedTree.value = 'normal'
  } catch (error) {
    errorMsg.value = error.message || '加载天赋状态失败'
  }
}

async function selectTree(tree) {
  selectedTree.value = tree
  selectedMarker.value = { panel: 'main', id: '' }
  await loadState()
}

async function handleNodeClick(item) {
  if (!isLoggedIn.value || learnLoading.value) return
  const currentLevel = nodeLevel(item.id)
  const maxLevel = item.maxLevel || 5

  if (currentLevel >= maxLevel) return // 已满级

  const targetLevel = currentLevel + 1
  const cost = talentCostToUpgrade(item, currentLevel, targetLevel)

  if (cost > availableTalentPoints.value) {
    showToast('天赋点不足')
    return
  }

  learnLoading.value = true
  errorMsg.value = ''
  try {
    const resp = await fetch('/api/talents/upgrade', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ talentId: item.id, targetLevel }),
    })
    if (!resp.ok) {
      const data = await safeJSON(resp)
      throw new Error(data?.message || '升级失败')
    }
    const data = await resp.json()
    applyTalentResponse({ ...(talentState.value || {}), ...data })
    talentEffectLines.value = data.effectLines || talentEffectLines.value
    talentEffectDescriptions.value = data.effectDescriptions || talentEffectDescriptions.value
  } catch (e) {
    showToast(e.message)
  } finally {
    learnLoading.value = false
  }
}

async function handleOverflowUpgrade() {
  if (!isLoggedIn.value || learnLoading.value) return
  if (overflowUpgradeCost.value > availableTalentPoints.value) {
    showToast('天赋点不足')
    return
  }

  learnLoading.value = true
  errorMsg.value = ''
  try {
    const resp = await fetch('/api/talents/upgrade', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ talentId: 'overflow_sink', targetLevel: 1 }),
    })
    if (!resp.ok) {
      const data = await safeJSON(resp)
      throw new Error(data?.message || '溢出强化失败')
    }
    const data = await resp.json()
    applyTalentResponse({ ...(talentState.value || {}), ...data })
  } catch (error) {
    showToast(error.message || '溢出强化失败')
  } finally {
    learnLoading.value = false
  }
}

function clearNode() {
  selectedMarker.value = { panel: 'main', id: '' }
}

async function resetTalents() {
  if (!isLoggedIn.value) return
  if (!(await showConfirm('确认洗点', '普通天赋与溢出强化都会被清空并返还，确定继续吗？'))) return

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

onMounted(async () => {
  await loadDefs()
  await loadState()
})


watch(isLoggedIn, (val) => {
  if (val) {
    loadState()
  }
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
      <section class="talent-select talent-select--main-only">
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
      </section>

      <section class="talent-guide">
        <div class="talent-guide__block">
          <strong>解锁规则</strong>
          <p>上一层所有节点到 Lv1，才能学习下一层节点。</p>
        </div>
        <div class="talent-guide__block">
          <strong>当前主系层满额外加成</strong>
          <ul class="talent-guide__list">
            <li v-for="bonus in currentTierCompletionBonuses" :key="`guide-bonus-${bonus.tier}`">
              {{ tierLabels[bonus.tier] }}：{{ bonus.label }}
            </li>
          </ul>
        </div>
        <div class="talent-guide__block">
          <strong>其他注意事项</strong>
          <ul class="talent-guide__list">
            <li>节点首次学习和后续升级都会消耗天赋点，等级越高消耗越多。</li>
            <li>层满加成在该层所有节点达到 Lv1 后立即生效，切换主系时会同步查看对应文案。</li>
            <li>洗点时普通天赋点与溢出强化消耗都会返还。</li>
          </ul>
        </div>
      </section>

      <section class="talent-overflow">
        <div class="talent-overflow__head">
          <div>
            <p class="vote-stage__eyebrow">独立节点</p>
            <h3>天赋点溢出强化</h3>
          </div>
          <button
            class="talent-overflow__button"
            :disabled="learnLoading || overflowUpgradeCost > availableTalentPoints"
            @click="handleOverflowUpgrade"
          >
            消耗 1000 点随机强化
          </button>
        </div>
        <div class="talent-overflow__stats">
          <div class="talent-overflow__stat">
            <span>当前可用天赋点</span>
            <strong>{{ availableTalentPoints }}</strong>
          </div>
          <div class="talent-overflow__stat">
            <span>溢出等级</span>
            <strong>{{ overflowLevel }}</strong>
          </div>
          <div class="talent-overflow__stat">
            <span>累计消耗</span>
            <strong>{{ overflowTotalSpent }}</strong>
          </div>
          <div class="talent-overflow__stat">
            <span>单次消耗</span>
            <strong>{{ overflowUpgradeCost }}</strong>
          </div>
        </div>
        <div class="talent-overflow__body">
          <div class="talent-overflow__panel">
            <strong>随机池</strong>
            <ul class="talent-guide__list">
              <li>软组织 / 弱点 / 重甲 / 暴击伤害 / 攻击力 / 全伤害</li>
              <li>每次命中固定获得 +0.1%</li>
              <li>洗点时与普通天赋一起返还</li>
            </ul>
          </div>
          <div class="talent-overflow__panel">
            <strong>已获得属性汇总</strong>
            <ul v-if="overflowBonusEntries.length > 0" class="talent-guide__list">
              <li v-for="item in overflowBonusEntries" :key="item.key">
                {{ formatOverflowBonus(item) }}
              </li>
            </ul>
            <p v-else class="talent-overflow__empty">尚未进行溢出强化。</p>
          </div>
        </div>
      </section>

      <article class="talent-plate">
        <header class="talent-plate__head">
          <strong>{{ activePlateTitle }}</strong>
          <span>{{ activePlateHint }}</span>
        </header>

        <div class="talent-half-ring" :style="{ '--active-color': treeConfig[selectedTree]?.color }">
          <Transition name="toast">
            <div v-if="toastMsg" class="talent-toast">{{ toastMsg }}</div>
          </Transition>
          <div v-if="activeTierBonuses.length > 0" class="talent-tier-bonuses">
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
              v-for="item in mainRingNodes"
              :key="`main-${item.id}`"
              class="talent-dot"
              :class="[
              `talent-dot--${nodeState(item)}`,
              `talent-dot--lv${nodeLevel(item.id)}`,
              { 'talent-dot--filler': isFillerNode(item.id) },
              { 'talent-dot--selected': selectedMarker?.id === item.id },
            ]"
              :style="ringNodeStyle(item)"
              @click="handleNodeClick(item)"
              @mouseenter="selectNode(item)"
              @mouseleave="clearNode"
          >
            <span class="talent-dot__level" v-if="nodeLevel(item.id) > 0">
              Lv{{ nodeLevel(item.id) }}
            </span>
            <img
                v-if="talentIconPath(item.id)"
                :src="talentIconPath(item.id)"
                class="talent-dot__icon"
                :class="{ 'talent-dot__icon--filler': isFillerNode(item.id) }"
                alt=""
            />
            <span class="talent-dot__name">{{ item.name }}</span>
            <span class="talent-dot__meta">{{ tierLabels[item.tier] || '' }} · {{ talentCostForLevel(item, 1) }}点</span>
          </button>

          <div v-if="selectedNode" class="talent-float" :style="detailFloatStyle()">
            <strong>{{ selectedNode.name }}</strong>
            <p>状态：{{ stateLabel(selectedNode) }}</p>
            <p v-if="nodeLevel(selectedNode.id) > 0">
              等级：Lv{{ nodeLevel(selectedNode.id) }} / Lv{{ selectedNode.maxLevel || 5 }}
            </p>
            <p v-if="nodeState(selectedNode) === 'upgradable' || nodeState(selectedNode) === 'available'">
              下一级消耗：{{ upgradeCost(selectedNode) }} 天赋点
            </p>
            <p>{{ effectDescription(selectedNode) }}</p>
            <div class="talent-float__effects">
              <div v-for="line in effectLines(selectedNode, nodeLevel(selectedNode.id))" :key="line.label" class="talent-float__effect-line">
                <span class="talent-float__effect-label">{{ line.label }}</span>
                <span class="talent-float__effect-value">{{ line.text }}</span>
              </div>
            </div>
            <p class="talent-float__hint">{{ stateReason(selectedNode) }}</p>
          </div>

          <div class="talent-tier-legend">
            <span v-for="tier in [0, 1, 2, 3, 4]" :key="`legend-main-${tier}`">{{ tierLabels[tier] }}</span>
          </div>
        </div>
      </article>
    </template>

    <Teleport to="body">
      <div v-if="confirmModal.show" class="confirm-overlay" @click.self="confirmCancel">
        <div class="confirm-dialog">
          <h3 class="confirm-dialog__title">{{ confirmModal.title }}</h3>
          <p class="confirm-dialog__message">{{ confirmModal.message }}</p>
          <div class="confirm-dialog__actions">
            <button class="confirm-dialog__btn confirm-dialog__btn--cancel" @click="confirmCancel">取消</button>
            <button class="confirm-dialog__btn confirm-dialog__btn--ok" @click="confirmOK">确认</button>
          </div>
        </div>
      </div>
    </Teleport>
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
  grid-template-columns: 1fr;
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

.talent-guide {
  margin-bottom: 0.75rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.talent-guide__block {
  border: 1px solid #2d3f48;
  border-radius: 0.8rem;
  background: linear-gradient(180deg, #101d25 0%, #0d171d 100%);
  padding: 0.8rem 0.9rem;
}

.talent-guide__block strong {
  display: block;
  margin-bottom: 0.45rem;
  color: #f4f9fc;
  font-size: 0.84rem;
}

.talent-guide__block p {
  margin: 0;
  color: #a9bfcb;
  font-size: 0.78rem;
  line-height: 1.6;
}

.talent-guide__list {
  margin: 0;
  padding-left: 1rem;
  color: #a9bfcb;
  font-size: 0.78rem;
  line-height: 1.6;
}

.talent-guide__list li + li {
  margin-top: 0.22rem;
}

.talent-overflow {
  margin-bottom: 0.75rem;
  border: 1px solid #335064;
  border-radius: 0.9rem;
  background: linear-gradient(135deg, #16232c 0%, #0d171d 100%);
  padding: 0.9rem;
}

.talent-overflow__head {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: flex-start;
  margin-bottom: 0.8rem;
}

.talent-overflow__head h3 {
  margin: 0.2rem 0 0;
  color: #f4f9fc;
}

.talent-overflow__button {
  border: 1px solid #e3b86a;
  border-radius: 999px;
  background: linear-gradient(135deg, #f0cd92, #d28d2f);
  color: #1f1305;
  font-size: 0.84rem;
  font-weight: 700;
  padding: 0.58rem 1rem;
  cursor: pointer;
}

.talent-overflow__button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.talent-overflow__stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.6rem;
  margin-bottom: 0.8rem;
}

.talent-overflow__stat {
  border: 1px solid #29404f;
  border-radius: 0.8rem;
  background: rgba(8, 17, 24, 0.72);
  padding: 0.75rem 0.8rem;
}

.talent-overflow__stat span {
  display: block;
  color: #8ea6b6;
  font-size: 0.76rem;
  margin-bottom: 0.28rem;
}

.talent-overflow__stat strong {
  color: #f9efcd;
  font-size: 1.02rem;
}

.talent-overflow__body {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.talent-overflow__panel {
  border: 1px solid #253a48;
  border-radius: 0.8rem;
  background: rgba(11, 19, 26, 0.78);
  padding: 0.8rem 0.9rem;
}

.talent-overflow__panel strong {
  display: block;
  margin-bottom: 0.5rem;
  color: #f4f9fc;
}

.talent-overflow__empty {
  margin: 0;
  color: #a9bfcb;
  font-size: 0.78rem;
  line-height: 1.6;
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
  z-index: 1;
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

.talent-dot--learned,
.talent-dot--upgradable {
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
  pointer-events: none;
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
  .talent-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .talent-guide {
    grid-template-columns: 1fr;
  }

  .talent-head__actions {
    width: 100%;
    justify-content: space-between;
  }

  .talent-overflow__head {
    flex-direction: column;
  }

  .talent-overflow__body {
    grid-template-columns: 1fr;
  }

  .talent-overflow__stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .talent-half-ring {
    min-height: 520px;
  }

  .talent-dot {
    width: 68px;
    height: 68px;
  }
}

/* Level brightness progression */
.talent-dot--lv1 { filter: brightness(1.0); }
.talent-dot--lv2 { filter: brightness(1.1); box-shadow: 0 0 8px var(--active-color); }
.talent-dot--lv3 { filter: brightness(1.2); box-shadow: 0 0 14px var(--active-color), 0 0 28px color-mix(in srgb, var(--active-color) 30%, transparent); }
.talent-dot--lv4 { filter: brightness(1.35); box-shadow: 0 0 22px var(--active-color), 0 0 44px color-mix(in srgb, var(--active-color) 40%, transparent); }
.talent-dot--lv5 { filter: brightness(1.5); box-shadow: 0 0 30px var(--active-color), 0 0 60px color-mix(in srgb, var(--active-color) 50%, transparent); animation: lv5-pulse 2s ease-in-out infinite; }

.talent-dot--maxed {
  border-color: #f0cd92 !important;
}

@keyframes lv5-pulse {
  0%, 100% { box-shadow: 0 0 28px var(--active-color), 0 0 56px color-mix(in srgb, var(--active-color) 45%, transparent); }
  50% { box-shadow: 0 0 34px var(--active-color), 0 0 68px color-mix(in srgb, var(--active-color) 60%, transparent); }
}

.talent-dot__level {
  position: absolute;
  top: -6px;
  right: -6px;
  font-size: 0.52rem;
  font-weight: 700;
  color: #0f1a22;
  background: var(--active-color, #2bb873);
  border-radius: 999px;
  padding: 0.08rem 0.24rem;
  line-height: 1;
  z-index: 2;
}

/* 确认弹窗 */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 9000;
  background: rgba(4, 8, 14, 0.72);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-dialog {
  width: min(340px, 88vw);
  background: #0f1c23;
  border: 1px solid #2d3f48;
  border-radius: 1rem;
  padding: 1.4rem 1.2rem 1rem;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.confirm-dialog__title {
  margin: 0 0 0.5rem;
  font-size: 0.96rem;
  font-weight: 700;
  color: #f4f9fc;
}

.confirm-dialog__message {
  margin: 0 0 1.2rem;
  font-size: 0.82rem;
  color: #8faabb;
  line-height: 1.5;
}

.confirm-dialog__actions {
  display: flex;
  gap: 0.6rem;
  justify-content: flex-end;
}

.confirm-dialog__btn {
  min-width: 80px;
  min-height: 40px;
  border-radius: 0.6rem;
  border: 0;
  font-size: 0.84rem;
  font-weight: 700;
  cursor: pointer;
  padding: 0.4rem 1rem;
  transition: opacity 0.15s;
}

.confirm-dialog__btn:active {
  opacity: 0.8;
}

.confirm-dialog__btn--cancel {
  color: #8faabb;
  background: #15242e;
  border: 1px solid #2d3f48;
}

.confirm-dialog__btn--ok {
  color: #fff7fa;
  background: linear-gradient(135deg, #e11d48, #be123c);
  box-shadow: 0 4px 14px rgba(225, 29, 72, 0.28);
}

/* Toast 悬浮提示 */
.talent-toast {
  position: absolute;
  top: 10%;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
  padding: 0.42rem 1rem;
  border-radius: 0.6rem;
  background: rgba(225, 29, 72, 0.92);
  color: #fff7fa;
  font-size: 0.82rem;
  font-weight: 700;
  pointer-events: none;
  white-space: nowrap;
}

.toast-enter-active { transition: opacity 0.2s, transform 0.2s; }
.toast-leave-active { transition: opacity 0.35s, transform 0.35s; }
.toast-enter-from { opacity: 0; transform: translateX(-50%) translateY(-6px); }
.toast-leave-to { opacity: 0; transform: translateX(-50%) translateY(-10px); }

/* 效果数值列表 */
.talent-float__effects {
  margin: 0.35rem 0 0;
  padding-top: 0.35rem;
  border-top: 1px solid #253540;
}

.talent-float__effect-line {
  display: flex;
  justify-content: space-between;
  gap: 0.4rem;
  margin-bottom: 0.12rem;
  font-size: 0.72rem;
}

.talent-float__effect-label {
  color: #7f99aa;
  flex-shrink: 0;
}

.talent-float__effect-value {
  color: #f0cd92;
  text-align: right;
  font-weight: 600;
  word-break: keep-all;
}
</style>
