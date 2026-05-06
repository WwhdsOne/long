<script setup>
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'

import {usePublicPageState} from './publicPageState'

const props = defineProps({
  focusSection: {
    type: String,
    default: 'resources',
  },
})

const {
  inventory,
  loadout,
  loadoutSlots,
  combatStats,
  nickname,
  actioningItemId,
  gold,
  stones,
  talentPoints,
  isLoggedIn,
  equippedItems,
  formatRarityLabel,
  formatNumber,
  formatItemStatLines,
  equipmentNameParts,
  equipmentNameClass,
  toggleItemEquip,
  salvageItem,
  toggleItemLock,
  salvageUnequippedItems,
  enhanceItem,
} = usePublicPageState()

const contextMenu = ref({
  open: false,
  x: 0,
  y: 0,
  item: null,
})
const salvageConfirmItem = ref(null)
const enhanceConfirmItem = ref(null)
const enhanceFeedback = ref('')
const enhanceSelectedLevels = ref(0)
const bulkSalvageConfirmData = ref(null)
const bulkSalvageFeedback = ref('')
const bulkSalvaging = ref(false)
const salvageRuleModalOpen = ref(false)

const enhanceAttackGrowth = 1.12
const enhancePercentGrowth = 1.08

function formatTrimmedNumber(value, digits = 2) {
  const normalized = Number(value ?? 0)
  if (!Number.isFinite(normalized)) return '0'
  return normalized.toFixed(digits).replace(/\.?0+$/, '')
}

function formatPercentValue(value) {
  return `${formatTrimmedNumber(value, 2)}%`
}

function formatRatioPercentValue(value) {
  return `${formatTrimmedNumber(Number(value ?? 0) * 100, 2)}%`
}

function formatMultiplierValue(value) {
  return `x${formatTrimmedNumber(value, 2)}`
}

function formatCritDamageBonus(value) {
  return `+${formatTrimmedNumber((Number(value ?? 0) - 1) * 100, 2)}%`
}

function formatArmorPenPercent(value) {
  return formatRatioPercentValue(value)
}

function formatPreviewPercent(value) {
  const normalized = Number(value ?? 0)
  const display = Math.abs(normalized) <= 1 ? normalized * 100 : normalized
  return `${formatTrimmedNumber(display, 2)}%`
}

function formatPreviewBonusPercent(value) {
  return `+${formatPreviewPercent(value)}`
}

const combatStatSummaryItems = computed(() => [
  {label: '攻击力', value: formatNumber(combatStats.value?.attackPower ?? 0)},
  {label: '暴击伤害', value: formatNumber(combatStats.value?.criticalDamage ?? 0)},
  {label: '暴击率', value: formatPercentValue(combatStats.value?.criticalChancePercent ?? 0)},
  {label: '护甲穿透', value: formatArmorPenPercent(combatStats.value?.armorPenPercent ?? 0)},
  {label: '暴击倍率', value: formatCritDamageBonus(combatStats.value?.critDamageMultiplier ?? 0)},
  {label: '全伤害加成', value: formatRatioPercentValue(combatStats.value?.allDamageAmplify ?? 0)},
  {label: '软组织增伤', value: formatRatioPercentValue(combatStats.value?.partTypeDamageSoft ?? 0)},
  {label: '重甲增伤', value: formatRatioPercentValue(combatStats.value?.partTypeDamageHeavy ?? 0)},
  {label: '弱点增伤', value: formatRatioPercentValue(combatStats.value?.partTypeDamageWeak ?? 0)},
  {label: '部位增伤', value: formatRatioPercentValue(combatStats.value?.perPartDamagePercent ?? 0)},
  {label: '低血量增伤倍率', value: formatMultiplierValue(combatStats.value?.lowHpMultiplier ?? 1)},
  {label: '低血量阈值', value: formatRatioPercentValue(combatStats.value?.lowHpThreshold ?? 0)},
])

const loadoutColumns = computed(() => [
  loadoutSlots.slice(0, 3),
  loadoutSlots.slice(3, 6),
])

function sectionID(section) {
  return `armory-${section}`
}

function scrollToSection(section) {
  const element = document.getElementById(sectionID(section))
  if (!element) return
  element.scrollIntoView({behavior: 'smooth', block: 'start'})
}

function closeContextMenu() {
  contextMenu.value.open = false
  contextMenu.value.item = null
}

function openItemContextMenu(event, item) {
  event.preventDefault()
  if (!isLoggedIn.value) return
  contextMenu.value = {
    open: true,
    x: event.clientX,
    y: event.clientY,
    item,
  }
}

async function handleContextToggleEquip() {
  const item = contextMenu.value.item
  if (!item) return
  closeContextMenu()
  await toggleItemEquip(item.instanceId || item.itemId, item.equipped)
}

function handleContextSalvage() {
  const item = contextMenu.value.item
  if (!item) return
  if (item.locked) {
    closeContextMenu()
    return
  }
  salvageConfirmItem.value = item
  closeContextMenu()
}

async function handleContextToggleLock() {
  const item = contextMenu.value.item
  if (!item) return
  closeContextMenu()
  await toggleItemLock(item.instanceId || item.itemId, item.locked)
}

function handleContextEnhance() {
  const item = contextMenu.value.item
  if (!item) return
  enhanceConfirmItem.value = item
  enhanceFeedback.value = ''
  enhanceSelectedLevels.value = Math.min(1, maxAffordableEnhanceLevelsByStone(item))
  closeContextMenu()
}

function salvageBaseReward(rarity) {
  switch (String(rarity || '').trim()) {
    case '至臻':
      return {gold: 10000, stones: 50}
    case '传说':
      return {gold: 2000, stones: 8}
    case '史诗':
      return {gold: 1000, stones: 3}
    case '稀有':
      return {gold: 500, stones: 1}
    case '优秀':
      return {gold: 300, stones: 1}
    case '神话':
      return {gold: 5000, stones: 20}
    case '普通':
    default:
      return {gold: 200, stones: 0}
  }
}

function salvagePreview(item) {
  return salvageBaseReward(item?.rarity)
}

function estimateBulkSalvage() {
  const candidates = inventory.value.filter((item) => !item.equipped && !item.locked && String(item.rarity || '').trim() !== '至臻')
  const byRarity = {}
  let gold = 0
  let stones = 0
  let hasEnhanced = false
  for (const item of candidates) {
    const rarity = String(item.rarity || '').trim() || '普通'
    const reward = salvageBaseReward(rarity)
    gold += reward.gold
    stones += reward.stones
    byRarity[rarity] = (byRarity[rarity] || 0) + 1
    if (Number(item.enhanceLevel || 0) > 0) {
      hasEnhanced = true
    }
  }
  return {
    total: candidates.length,
    byRarity,
    gold,
    stones,
    hasEnhanced,
    excludedEquipped: inventory.value.filter((item) => item.equipped).length,
    excludedLocked: inventory.value.filter((item) => item.locked).length,
    excludedTopRarity: inventory.value.filter((item) => String(item.rarity || '').trim() === '至臻').length,
  }
}

const canBulkSalvage = computed(() => estimateBulkSalvage().total > 0)

function openBulkSalvageConfirm() {
  bulkSalvageConfirmData.value = estimateBulkSalvage()
  bulkSalvageFeedback.value = ''
}

function openSalvageRuleModal() {
  salvageRuleModalOpen.value = true
}

function closeSalvageRuleModal() {
  salvageRuleModalOpen.value = false
}

function cancelBulkSalvage() {
  bulkSalvageConfirmData.value = null
  bulkSalvageFeedback.value = ''
}

async function confirmBulkSalvage() {
  if (!bulkSalvageConfirmData.value) return
  bulkSalvageFeedback.value = ''
  bulkSalvaging.value = true
  const result = await salvageUnequippedItems()
  bulkSalvaging.value = false
  if (!result) {
    bulkSalvageFeedback.value = '一键分解失败，请稍后重试。'
    return
  }
  cancelBulkSalvage()
}

async function confirmSalvage() {
  const item = salvageConfirmItem.value
  if (!item) return
  await salvageItem(item.instanceId || item.itemId)
  salvageConfirmItem.value = null
}

function cancelSalvage() {
  salvageConfirmItem.value = null
}

function enhanceGoldCost(level) {
  const safeLevel = Math.max(0, Number(level || 0))
  return Math.ceil(500 * (1.5 ** safeLevel))
}

function enhanceStoneCost(level) {
  const safeLevel = Math.max(0, Number(level || 0))
  return Math.ceil(3 * (1.5 ** safeLevel))
}

function enhanceBatchCost(level, levels) {
  const safeLevel = Math.max(0, Number(level || 0))
  const safeLevels = Math.max(0, Number(levels || 0))
  let goldCost = 0
  let stoneCost = 0
  for (let offset = 0; offset < safeLevels; offset += 1) {
    const nextLevel = safeLevel + offset
    goldCost += enhanceGoldCost(nextLevel)
    stoneCost += enhanceStoneCost(nextLevel)
  }
  return { goldCost, stoneCost }
}

function maxEnhanceLevel(rarity) {
  switch (String(rarity || '').trim()) {
    case '优秀':
      return 10
    case '稀有':
      return 15
    case '史诗':
      return 20
    case '传说':
      return 25
    case '至臻':
      return 35
    case '普通':
    default:
      return 5
  }
}

function isEnhanceMax(item) {
  if (!item) return false
  return Number(item.enhanceLevel || 0) >= maxEnhanceLevel(item.rarity)
}

function maxAffordableEnhanceLevelsByStone(item) {
  if (!item) return 0
  const currentLevel = Math.max(0, Number(item.enhanceLevel || 0))
  const maxLevel = maxEnhanceLevel(item.rarity)
  let remainingStones = Math.max(0, Number(stones.value || 0))
  let levels = 0
  let nextLevel = currentLevel

  while (nextLevel < maxLevel) {
    const cost = enhanceStoneCost(nextLevel)
    if (remainingStones < cost) break
    remainingStones -= cost
    levels += 1
    nextLevel += 1
  }

  return levels
}

function recoverBaseAttackPower(currentAttackPower, currentLevel) {
  const safeAttackPower = Math.max(0, Number(currentAttackPower || 0))
  const safeLevel = Math.max(0, Number(currentLevel || 0))
  if (safeLevel <= 0 || safeAttackPower <= 0) return safeAttackPower

  const approx = safeAttackPower / (enhanceAttackGrowth ** safeLevel)
  const minBase = Math.max(0, Math.floor(approx) - 8)
  const maxBase = Math.max(minBase, Math.ceil(approx) + 8)
  for (let baseAttackPower = minBase; baseAttackPower <= maxBase; baseAttackPower += 1) {
    if (Math.round(baseAttackPower * (enhanceAttackGrowth ** safeLevel)) === safeAttackPower) {
      return baseAttackPower
    }
  }
  return Math.max(0, Math.round(approx))
}

function previewScaledStat(currentValue, currentLevel, targetLevel) {
  const safeCurrentValue = Number(currentValue || 0)
  const safeCurrentLevel = Math.max(0, Number(currentLevel || 0))
  const safeTargetLevel = Math.max(safeCurrentLevel, Number(targetLevel || 0))
  if (!Number.isFinite(safeCurrentValue) || safeCurrentValue === 0) return 0
  if (safeTargetLevel === safeCurrentLevel) return safeCurrentValue

  const baseValue = safeCurrentValue / (enhancePercentGrowth ** safeCurrentLevel)
  return baseValue * (enhancePercentGrowth ** safeTargetLevel)
}

function buildEnhancePreviewItem(item, levels) {
  if (!item) return null

  const currentLevel = Math.max(0, Number(item.enhanceLevel || 0))
  const targetLevel = currentLevel + Math.max(0, Number(levels || 0))
  const preview = {
    attackPower: Math.max(0, Number(item.attackPower || 0)),
    armorPenPercent: Number(item.armorPenPercent || 0),
    critRate: Number(item.critRate || 0),
    critDamageMultiplier: Number(item.critDamageMultiplier || 0),
    partTypeDamageSoft: Number(item.partTypeDamageSoft || 0),
    partTypeDamageHeavy: Number(item.partTypeDamageHeavy || 0),
    partTypeDamageWeak: Number(item.partTypeDamageWeak || 0),
  }

  if (targetLevel === currentLevel) {
    return preview
  }

  if (preview.attackPower > 0) {
    const baseAttackPower = recoverBaseAttackPower(preview.attackPower, currentLevel)
    preview.attackPower = Math.round(baseAttackPower * (enhanceAttackGrowth ** targetLevel))
  }
  preview.armorPenPercent = previewScaledStat(preview.armorPenPercent, currentLevel, targetLevel)
  preview.critRate = previewScaledStat(preview.critRate, currentLevel, targetLevel)
  preview.critDamageMultiplier = previewScaledStat(preview.critDamageMultiplier, currentLevel, targetLevel)
  preview.partTypeDamageSoft = previewScaledStat(preview.partTypeDamageSoft, currentLevel, targetLevel)
  preview.partTypeDamageHeavy = previewScaledStat(preview.partTypeDamageHeavy, currentLevel, targetLevel)
  preview.partTypeDamageWeak = previewScaledStat(preview.partTypeDamageWeak, currentLevel, targetLevel)

  return preview
}

function pushEnhancePreviewRow(rows, label, currentValue, nextValue, formatter) {
  const safeCurrentValue = Number(currentValue || 0)
  const safeNextValue = Number(nextValue || 0)
  if (safeCurrentValue === 0 && safeNextValue === 0) return
  rows.push({
    label,
    currentText: formatter(safeCurrentValue),
    nextText: formatter(safeNextValue),
  })
}

const enhanceAffordableLevelsByStone = computed(() => maxAffordableEnhanceLevelsByStone(enhanceConfirmItem.value))

const enhanceTargetLevel = computed(() => enhanceLevel.value + Math.max(0, Number(enhanceSelectedLevels.value || 0)))

const enhanceBatchCostPreview = computed(() => enhanceBatchCost(enhanceLevel.value, enhanceSelectedLevels.value))

const enhancePreviewItem = computed(() => buildEnhancePreviewItem(enhanceConfirmItem.value, enhanceSelectedLevels.value))

const enhanceHasEnoughGold = computed(() => Number(gold.value || 0) >= enhanceBatchCostPreview.value.goldCost)

const enhancePreviewStatRows = computed(() => {
  const item = enhanceConfirmItem.value
  const preview = enhancePreviewItem.value
  if (!item || !preview) return []

  const rows = []
  pushEnhancePreviewRow(rows, '攻击力', item.attackPower, preview.attackPower, (value) => formatNumber(value))
  pushEnhancePreviewRow(rows, '护甲穿透', item.armorPenPercent, preview.armorPenPercent, formatArmorPenPercent)
  pushEnhancePreviewRow(rows, '暴击率', item.critRate, preview.critRate, formatRatioPercentValue)
  pushEnhancePreviewRow(rows, '暴击倍率', item.critDamageMultiplier, preview.critDamageMultiplier, formatPreviewBonusPercent)
  pushEnhancePreviewRow(rows, '软组织伤害', item.partTypeDamageSoft, preview.partTypeDamageSoft, formatPreviewPercent)
  pushEnhancePreviewRow(rows, '重甲伤害', item.partTypeDamageHeavy, preview.partTypeDamageHeavy, formatPreviewPercent)
  pushEnhancePreviewRow(rows, '弱点伤害', item.partTypeDamageWeak, preview.partTypeDamageWeak, formatPreviewPercent)
  return rows
})

const enhanceCanConfirm = computed(() => (
  enhanceSelectedLevels.value > 0
  && enhanceSelectedLevels.value <= enhanceAffordableLevelsByStone.value
  && enhanceHasEnoughGold.value
))

async function confirmEnhance() {
  const item = enhanceConfirmItem.value
  if (!item) return
  if (isEnhanceMax(item)) {
    enhanceFeedback.value = '无法继续强化，强化已达上限'
    return
  }
  if (enhanceSelectedLevels.value <= 0) {
    enhanceFeedback.value = '先选择要强化的等级。'
    return
  }
  const result = await enhanceItem(item.instanceId || item.itemId, enhanceSelectedLevels.value)
  if (result?.ok === false) {
    enhanceFeedback.value = result.message || '强化失败，请稍后重试。'
    return
  }
  enhanceConfirmItem.value = null
  enhanceFeedback.value = ''
  enhanceSelectedLevels.value = 0
}

function cancelEnhance() {
  enhanceConfirmItem.value = null
  enhanceFeedback.value = ''
  enhanceSelectedLevels.value = 0
}


const enhanceDisplayName = computed(() => {
  const name = enhanceConfirmItem.value?.name || enhanceConfirmItem.value?.itemId || ''
  return String(name).replace(/\s*\+\d+$/, '')
})

const enhanceLevel = computed(() => {
  const n = Number(enhanceConfirmItem.value?.enhanceLevel ?? 0)
  return Number.isFinite(n) ? n : 0
})

watch(enhanceAffordableLevelsByStone, (nextMax) => {
  if (enhanceSelectedLevels.value > nextMax) {
    enhanceSelectedLevels.value = nextMax
  }
})

onMounted(() => {
  nextTick(() => scrollToSection(props.focusSection))
  window.addEventListener('click', closeContextMenu)
  window.addEventListener('resize', closeContextMenu)
})

watch(
    () => props.focusSection,
    (nextSection) => {
      nextTick(() => scrollToSection(nextSection))
    },
)

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu)
  window.removeEventListener('resize', closeContextMenu)
})
</script>

<template>
  <section class="stage-layout stage-layout--single">
    <section class="armory-layout">
      <aside class="armory-layout__left">
        <section :id="sectionID('stats')" class="armory-block">
          <div class="armory-block__head">
            <p class="vote-stage__eyebrow">战斗属性</p>
            <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
          </div>
          <div class="armory-combat-summary">
            <article
                v-for="item in combatStatSummaryItems"
                :key="item.label"
                class="armory-combat-summary__item"
            >
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </article>
          </div>
        </section>

        <section :id="sectionID('loadout')" class="armory-block">
          <div class="armory-block__head">
            <p class="vote-stage__eyebrow">装备栏</p>
            <strong>{{ equippedItems.length }} / {{ loadoutSlots.length }}</strong>
          </div>
          <div class="loadout-grid loadout-grid--paired">
            <div v-for="(column, columnIndex) in loadoutColumns" :key="columnIndex" class="loadout-column">
              <article
                  v-for="slot in column"
                  :key="slot.value"
                  class="loadout-slot"
                  :class="{ 'loadout-slot--equipped': Boolean(loadout[slot.value]) }"
              >
                <div class="loadout-slot__visual">
                  <img
                      v-if="loadout[slot.value]?.imagePath"
                      class="loadout-slot__icon"
                      :src="loadout[slot.value].imagePath"
                      :alt="loadout[slot.value].imageAlt || loadout[slot.value].name || loadout[slot.value].itemId"
                  />
                  <span v-else-if="loadout[slot.value]" class="loadout-slot__fallback">
                    {{ equipmentNameParts(loadout[slot.value]).text.slice(0, 1) || '?' }}
                  </span>
                  <span v-else class="loadout-slot__placeholder">{{ slot.label }}</span>
                </div>
                <div class="loadout-slot__main">
                  <strong v-if="loadout[slot.value]" class="loadout-slot__name">
                    <span :class="['loadout-slot__rarity', equipmentNameClass(loadout[slot.value])]">{{ formatRarityLabel(loadout[slot.value].rarity) }}</span>
                    <span :class="equipmentNameClass(loadout[slot.value])"> · {{ equipmentNameParts(loadout[slot.value]).text }}</span>
                    <br>
                    <span class="loadout-slot__meta"> 强化 +{{ loadout[slot.value].enhanceLevel || 0 }}</span>
                  </strong>
                  <strong v-else class="loadout-slot__empty">未穿戴</strong>
                </div>
                <article v-if="loadout[slot.value]" class="armory-item-tooltip" aria-label="装备属性">
                  <p class="vote-stage__eyebrow">装备属性</p>
                  <strong>{{ loadout[slot.value].name || loadout[slot.value].itemId }}</strong>
                  <p>{{ formatRarityLabel(loadout[slot.value].rarity) }} · 强化
                    +{{ loadout[slot.value].enhanceLevel || 0 }}</p>
                  <ul v-if="formatItemStatLines(loadout[slot.value]).length > 0" class="armory-item-tooltip__stats">
                    <li v-for="line in formatItemStatLines(loadout[slot.value])" :key="line">{{ line }}</li>
                  </ul>
                  <p v-else>暂无词条</p>
                </article>
              </article>
            </div>
          </div>
        </section>
      </aside>

      <section :id="sectionID('inventory')" class="armory-layout__right armory-block">
        <div class="armory-block__head">
          <p class="vote-stage__eyebrow">背包</p>
          <div class="armory-inventory-head__actions">
            <strong>右键装备可以强化、拆解、装备、锁定 ｜</strong>
            <strong>当前有 {{ inventory.length }} 件装备</strong>
            <button
                class="nickname-form__ghost armory-inventory-head__bulk-button"
                type="button"
                @click="openSalvageRuleModal"
            >
              分解规则
            </button>
            <button
                class="nickname-form__ghost armory-inventory-head__bulk-button"
                type="button"
                :disabled="!isLoggedIn || !canBulkSalvage || bulkSalvaging"
                @click="openBulkSalvageConfirm"
            >
              一键分解未穿戴
            </button>
          </div>
        </div>
        <p :id="sectionID('resources')" class="armory-backpack-resources">
          资源：
          金币 <span class="num-gold">{{ gold }}</span>
          · 强化石 <span class="num-stone">{{ stones }}</span>
          · 天赋点 <span class="num-stone">{{ talentPoints }}</span>
        </p>
        <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
          <p>先去打 Boss，掉落会自动进背包。</p>
        </div>
        <div v-else class="armory-backpack-grid">
          <article
              v-for="item in inventory"
              :key="item.instanceId || `${item.itemId}-${item.name}`"
              class="armory-backpack-cell"
          >
            <button
                class="armory-backpack-cell__button"
                type="button"
                :disabled="!isLoggedIn || actioningItemId === (item.instanceId || item.itemId)"
                @click="toggleItemEquip(item.instanceId || item.itemId, item.equipped)"
                @contextmenu="openItemContextMenu($event, item)"
            >
              <img
                  v-if="item.imagePath"
                  class="armory-backpack-cell__icon"
                  :src="item.imagePath"
                  :alt="item.imageAlt || item.name || item.itemId"
              />
              <span v-else class="armory-backpack-cell__fallback">{{
                  equipmentNameParts(item).text.slice(0, 1) || '?'
                }}</span>
            </button>
            <article class="armory-item-tooltip" aria-label="装备属性">
              <p class="vote-stage__eyebrow">装备属性</p>
              <strong>{{ item.name || item.itemId }}</strong>
              <p>{{ formatRarityLabel(item.rarity) }} · 强化 +{{ item.enhanceLevel || 0 }}</p>
              <ul v-if="formatItemStatLines(item).length > 0" class="armory-item-tooltip__stats">
                <li v-for="line in formatItemStatLines(item)" :key="line">{{ line }}</li>
              </ul>
              <p v-else>暂无词条</p>
            </article>
            <div class="armory-backpack-cell__meta">
              <strong>
                <span :class="['armory-backpack-cell__rarity', equipmentNameClass(item)]">{{ formatRarityLabel(item.rarity) }}</span>
<!--                <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>-->
                <span style="font-size:15px">·</span>
                <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
              </strong>
              <span>{{ item.slot || '未分类' }}</span>
              <span>强化 +{{ item.enhanceLevel || 0 }} · {{ item.equipped ? '已装备' : '点击穿戴' }}</span>
              <span>{{ item.locked ? '已锁定（不参与一键分解）' : '未锁定' }}</span>
            </div>
          </article>
        </div>
      </section>

    </section>

    <section
        v-if="contextMenu.open && contextMenu.item"
        class="inventory-context-menu"
        :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
        @click.stop
    >
      <button class="inventory-context-menu__item" type="button" @click="handleContextToggleEquip">
        {{ contextMenu.item.equipped ? '卸下' : '穿戴' }}
      </button>
      <button class="inventory-context-menu__item" type="button" @click="handleContextEnhance">
        强化
      </button>
      <button class="inventory-context-menu__item" type="button" @click="handleContextToggleLock">
        {{ contextMenu.item.locked ? '解锁' : '锁定' }}
      </button>
      <button class="inventory-context-menu__item" type="button" @click="handleContextSalvage">
        分解
      </button>
    </section>

    <section v-if="enhanceConfirmItem" class="boss-drop-modal" aria-label="强化确认">
      <div class="boss-drop-modal__backdrop" @click="cancelEnhance"></div>
      <article class="boss-drop-modal__card">
        <div class="boss-drop-modal__head">
          <div>
            <p class="vote-stage__eyebrow">装备强化</p>
            <strong>
              {{ enhanceDisplayName }} +{{ enhanceLevel }}
              →
              {{ enhanceDisplayName }} +{{ enhanceTargetLevel }}
            </strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="cancelEnhance">关闭</button>
        </div>
        <div class="leaderboard-list armory-enhance-panel">
          <p>强化上限：+{{ maxEnhanceLevel(enhanceConfirmItem.rarity) }}</p>
          <p>可拖动强化：+0 ~ +{{ enhanceAffordableLevelsByStone }}</p>
          <p>当前资源：金币 {{ gold }} · 强化石 {{ stones }}</p>
          <div class="armory-enhance-panel__summary">
            <strong>已选强化 +{{ enhanceSelectedLevels }} 级</strong>
            <span>强化后等级 +{{ enhanceTargetLevel }}</span>
          </div>
          <label class="armory-enhance-panel__range" for="armory-enhance-range">
            <span>强化滑条</span>
            <input
              id="armory-enhance-range"
              v-model.number="enhanceSelectedLevels"
              type="range"
              min="0"
              :max="enhanceAffordableLevelsByStone"
              step="1"
              @input="enhanceFeedback = ''"
            />
          </label>
          <div class="armory-enhance-panel__scale">
            <span>+0</span>
            <span>+{{ enhanceAffordableLevelsByStone }}</span>
          </div>
          <p>累计金币：{{ enhanceBatchCostPreview.goldCost }}</p>
          <p>累计强化石：{{ enhanceBatchCostPreview.stoneCost }}</p>
          <ul v-if="enhancePreviewStatRows.length > 0" class="armory-enhance-preview">
            <li v-for="row in enhancePreviewStatRows" :key="row.label" class="armory-enhance-preview__item">
              {{ row.label }} {{ row.currentText }} -> {{ row.label }} {{ row.nextText }}
            </li>
          </ul>
          <p v-else>这件装备暂无可预览词条。</p>
          <p v-if="enhanceAffordableLevelsByStone === 0">当前强化石不足，无法继续强化。</p>
          <p v-else-if="!enhanceHasEnoughGold">金币不足，当前选择无法强化。</p>
        </div>
        <p v-if="enhanceFeedback" class="feedback feedback--error">{{ enhanceFeedback }}</p>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelEnhance">取消</button>
          <button class="nickname-form__submit" type="button" :disabled="!enhanceCanConfirm" @click="confirmEnhance">确认强化</button>
        </div>
      </article>
    </section>

    <section v-if="bulkSalvageConfirmData" class="boss-drop-modal" aria-label="一键分解确认">
      <div class="boss-drop-modal__backdrop" @click="cancelBulkSalvage"></div>
      <article class="boss-drop-modal__card">
        <div class="boss-drop-modal__head">
          <div>
            <p class="vote-stage__eyebrow">一键分解确认</p>
            <strong>即将分解 {{ bulkSalvageConfirmData.total }} 件未穿戴装备</strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="cancelBulkSalvage">关闭</button>
        </div>
        <div class="leaderboard-list">
          <p>预计金币：{{ bulkSalvageConfirmData.gold }}</p>
          <p>预计强化石：{{ bulkSalvageConfirmData.stones }}</p>
          <p v-if="bulkSalvageConfirmData.hasEnhanced">已强化装备会额外返还 60% 已消耗强化石（向下取整）。</p>
          <p>自动排除：穿戴中 {{ bulkSalvageConfirmData.excludedEquipped }} 件、已锁定
            {{ bulkSalvageConfirmData.excludedLocked }} 件、至臻 {{ bulkSalvageConfirmData.excludedTopRarity }} 件。</p>
          <p v-if="Object.keys(bulkSalvageConfirmData.byRarity).length > 0">
            分解明细：
            <span v-for="(count, rarity) in bulkSalvageConfirmData.byRarity" :key="rarity">
              {{ rarity }}×{{ count }}
            </span>
          </p>
        </div>
        <p v-if="bulkSalvageFeedback" class="feedback feedback--error">{{ bulkSalvageFeedback }}</p>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelBulkSalvage">取消</button>
          <button class="nickname-form__submit" type="button" :disabled="bulkSalvaging" @click="confirmBulkSalvage">
            {{ bulkSalvaging ? '分解中...' : '确认分解' }}
          </button>
        </div>
      </article>
    </section>

    <section v-if="salvageRuleModalOpen" class="boss-drop-modal" aria-label="分解规则">
      <div class="boss-drop-modal__backdrop" @click="closeSalvageRuleModal"></div>
      <article class="boss-drop-modal__card">
        <div class="boss-drop-modal__head">
          <div>
            <p class="vote-stage__eyebrow">装备分解规则</p>
            <strong>当前版本分解说明</strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="closeSalvageRuleModal">关闭</button>
        </div>
        <div class="leaderboard-list" style="line-height: 1.4; margin: 10px 0;">
          <p style="margin: 4px 0;">1. 一键分解仅处理未穿戴、未锁定、且非至臻装备。</p>
          <p style="margin: 4px 0;">2. 至臻装备默认不参与一键分解。</p>
          <p style="margin: 4px 0;">3. 已强化装备会额外返还 60% 已消耗强化石（向下取整）。</p>
          <p style="margin: 8px 0 6px;">4. 分解基础收益：</p>

          <table style="width: 100%; border-collapse: collapse; text-align: left; font-size: 14px;">
            <thead>
            <tr>
              <th style="border:1px solid #ccc; padding:6px; background:#f5f5f5;">装备品质</th>
              <th style="border:1px solid #ccc; padding:6px; background:#f5f5f5;">金币</th>
              <th style="border:1px solid #ccc; padding:6px; background:#f5f5f5;">强化石</th>
            </tr>
            </thead>
            <tbody>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">普通</td>
              <td style="border:1px solid #ccc; padding:6px;">200</td>
              <td style="border:1px solid #ccc; padding:6px;">0</td>
            </tr>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">优秀</td>
              <td style="border:1px solid #ccc; padding:6px;">300</td>
              <td style="border:1px solid #ccc; padding:6px;">1</td>
            </tr>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">稀有</td>
              <td style="border:1px solid #ccc; padding:6px;">500</td>
              <td style="border:1px solid #ccc; padding:6px;">1</td>
            </tr>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">史诗</td>
              <td style="border:1px solid #ccc; padding:6px;">1000</td>
              <td style="border:1px solid #ccc; padding:6px;">3</td>
            </tr>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">传说</td>
              <td style="border:1px solid #ccc; padding:6px;">2000</td>
              <td style="border:1px solid #ccc; padding:6px;">8</td>
            </tr>
            <tr>
              <td style="border:1px solid #ccc; padding:6px;">至臻</td>
              <td style="border:1px solid #ccc; padding:6px;">10000</td>
              <td style="border:1px solid #ccc; padding:6px;">50</td>
            </tr>
            </tbody>
          </table>
        </div>
        <div class="announcement-modal__actions">
          <button class="nickname-form__submit" type="button" @click="closeSalvageRuleModal">我知道了</button>
        </div>
      </article>
    </section>

    <section v-if="salvageConfirmItem" class="boss-drop-modal" aria-label="分解确认">
      <div class="boss-drop-modal__backdrop" @click="cancelSalvage"></div>
      <article class="boss-drop-modal__card">
        <div class="boss-drop-modal__head">
          <div>
            <p class="vote-stage__eyebrow">分解装备</p>
            <strong>确认分解这件装备？</strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="cancelSalvage">关闭</button>
        </div>
        <div class="leaderboard-list">
          <p>
            {{ salvageConfirmItem.name || salvageConfirmItem.itemId }}
          </p>
          <p>基础金币：{{ salvagePreview(salvageConfirmItem).gold }}</p>
          <p>基础强化石：{{ salvagePreview(salvageConfirmItem).stones }}</p>
          <p>若已强化，将额外返还 60% 已消耗强化石。</p>
        </div>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelSalvage">取消</button>
          <button class="nickname-form__submit" type="button" @click="confirmSalvage">确认分解</button>
        </div>
      </article>
    </section>
  </section>
</template>
