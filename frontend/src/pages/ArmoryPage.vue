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
  myClicks,
  myRank,
  myBossDamage,
  normalDamage,
  criticalDamage,
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
const bulkSalvageConfirmData = ref(null)
const bulkSalvageFeedback = ref('')
const bulkSalvaging = ref(false)
const salvageRuleModalOpen = ref(false)

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

function maxEnhanceLevel(rarity) {
  switch (String(rarity || '').trim()) {
    case '优秀':
      return 7
    case '稀有':
      return 10
    case '史诗':
      return 15
    case '传说':
      return 20
    case '至臻':
      return 25
    case '普通':
    default:
      return 5
  }
}

function isEnhanceMax(item) {
  if (!item) return false
  return Number(item.enhanceLevel || 0) >= maxEnhanceLevel(item.rarity)
}

async function confirmEnhance() {
  const item = enhanceConfirmItem.value
  if (!item) return
  if (isEnhanceMax(item)) {
    enhanceFeedback.value = '无法继续强化，强化已达上限'
    return
  }
  const result = await enhanceItem(item.instanceId || item.itemId)
  if (result?.ok === false) {
    enhanceFeedback.value = result.message || '强化失败，请稍后重试。'
    return
  }
  enhanceConfirmItem.value = null
  enhanceFeedback.value = ''
}

function cancelEnhance() {
  enhanceConfirmItem.value = null
  enhanceFeedback.value = ''
}


const enhanceDisplayName = computed(() => {
  const name = enhanceConfirmItem.value?.name || enhanceConfirmItem.value?.itemId || ''
  return String(name).replace(/\s*\+\d+$/, '')
})

const enhanceLevel = computed(() => {
  const n = Number(enhanceConfirmItem.value?.enhanceLevel ?? 0)
  return Number.isFinite(n) ? n : 0
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
          <div class="me-card__stats">
            <article>
              <span>普通伤害</span>
              <strong>{{ normalDamage }}</strong>
            </article>
            <article>
              <span>暴击伤害</span>
              <strong>{{ criticalDamage }}</strong>
            </article>
            <article>
              <span>暴击率</span>
              <strong>{{ formatNumber(combatStats.criticalChancePercent, 2) }}%</strong>
            </article>
            <article>
              <span>我的 Boss 伤害</span>
              <strong>{{ myBossDamage }}</strong>
            </article>
            <article>
              <span>我的点击</span>
              <strong>{{ isLoggedIn ? myClicks : '--' }}</strong>
            </article>
            <article>
              <span>我的排名</span>
              <strong>{{ isLoggedIn ? `#${myRank ?? '--'}` : '--' }}</strong>
            </article>
          </div>
        </section>

        <section :id="sectionID('loadout')" class="armory-block">
          <div class="armory-block__head">
            <p class="vote-stage__eyebrow">装备栏</p>
            <strong>{{ equippedItems.length }} / {{ loadoutSlots.length }}</strong>
          </div>
          <div class="loadout-grid">
            <article v-for="slot in loadoutSlots" :key="slot.value" class="loadout-slot">
              <div class="loadout-slot__main">
                <span>{{ slot.label }}</span>
                <strong v-if="loadout[slot.value]">
                  <span v-if="equipmentNameParts(loadout[slot.value]).prefix">{{ equipmentNameParts(loadout[slot.value]).prefix }}</span>
                  <span :class="equipmentNameClass(loadout[slot.value])">{{ equipmentNameParts(loadout[slot.value]).text }}</span>
                </strong>
                <strong v-else>未穿戴</strong>
              </div>
              <ul v-if="loadout[slot.value]" class="loadout-slot__attrs">
                <li>{{ formatRarityLabel(loadout[slot.value].rarity) }}</li>
                <li v-for="line in formatItemStatLines(loadout[slot.value])" :key="line">
                  {{ line }}
                </li>
              </ul>
              <p v-else class="loadout-slot__empty">暂无属性</p>
            </article>
          </div>
        </section>
      </aside>

      <section :id="sectionID('inventory')" class="armory-layout__right armory-block">
        <div class="armory-block__head">
          <p class="vote-stage__eyebrow">背包</p>
          <div class="armory-inventory-head__actions">
            <strong>{{ inventory.length }} 件</strong>
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
        <div v-if="props.focusSection === 'resources'" class="leaderboard-list leaderboard-list--empty">
          <p>资源、背包、属性、装备栏已拆成独立页签。</p>
        </div>
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
              <span v-else class="armory-backpack-cell__fallback">{{ equipmentNameParts(item).text.slice(0, 1) || '?' }}</span>
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
                <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
              </strong>
              <span>{{ formatRarityLabel(item.rarity) }} · {{ item.slot || '未分类' }}</span>
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
              {{ enhanceDisplayName }} +{{ enhanceLevel + 1 }}
            </strong>
          </div>
          <button class="nickname-form__ghost" type="button" @click="cancelEnhance">关闭</button>
        </div>
        <div class="leaderboard-list">
          <p>金币：{{ enhanceGoldCost(enhanceConfirmItem.enhanceLevel) }}</p>
          <p>强化石：{{ enhanceStoneCost(enhanceConfirmItem.enhanceLevel) }}</p>
          <p>强化上限：+{{ maxEnhanceLevel(enhanceConfirmItem.rarity) }}</p>
        </div>
        <p v-if="enhanceFeedback" class="feedback feedback--error">{{ enhanceFeedback }}</p>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelEnhance">取消</button>
          <button class="nickname-form__submit" type="button" @click="confirmEnhance">确认强化</button>
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
          <p>自动排除：穿戴中 {{ bulkSalvageConfirmData.excludedEquipped }} 件、已锁定 {{ bulkSalvageConfirmData.excludedLocked }} 件、至臻 {{ bulkSalvageConfirmData.excludedTopRarity }} 件。</p>
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
          <p>若已强化，将额外返还 60% 已消耗强化石（向下取整）。</p>
        </div>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelSalvage">取消</button>
          <button class="nickname-form__submit" type="button" @click="confirmSalvage">确认分解</button>
        </div>
      </article>
    </section>
  </section>
</template>
