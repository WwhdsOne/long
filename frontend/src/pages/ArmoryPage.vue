<script setup>
import {nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'

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
  await toggleItemEquip(item.itemId, item.equipped)
}

function handleContextSalvage() {
  const item = contextMenu.value.item
  if (!item) return
  salvageConfirmItem.value = item
  closeContextMenu()
}

function handleContextEnhance() {
  const item = contextMenu.value.item
  if (!item) return
  enhanceConfirmItem.value = item
  enhanceFeedback.value = ''
  closeContextMenu()
}

async function confirmSalvage() {
  const item = salvageConfirmItem.value
  if (!item) return
  await salvageItem(item.itemId)
  salvageConfirmItem.value = null
}

function cancelSalvage() {
  salvageConfirmItem.value = null
}

function enhanceGoldCost(level) {
  const safeLevel = Math.max(0, Number(level || 0))
  return Math.ceil(100 * (1.5 ** safeLevel))
}

function enhanceStoneCost(level) {
  const safeLevel = Math.max(0, Number(level || 0))
  return Math.ceil(1.4 ** safeLevel)
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
  const result = await enhanceItem(item.itemId)
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
          <strong>{{ inventory.length }} 件</strong>
        </div>
        <p :id="sectionID('resources')" class="armory-backpack-resources">
          资源：金币 {{ gold }} · 强化石 {{ stones }}
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
              :disabled="!isLoggedIn || actioningItemId === item.itemId"
              @click="toggleItemEquip(item.itemId, item.equipped)"
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
              {{ enhanceConfirmItem.name || enhanceConfirmItem.itemId }} +{{ enhanceConfirmItem.enhanceLevel || 0 }}
              →
              +{{ (enhanceConfirmItem.enhanceLevel || 0) + 1 }}
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
          <p>分解后会返还部分强化石。</p>
        </div>
        <div class="announcement-modal__actions">
          <button class="nickname-form__ghost" type="button" @click="cancelSalvage">取消</button>
          <button class="nickname-form__submit" type="button" @click="confirmSalvage">确认分解</button>
        </div>
      </article>
    </section>
  </section>
</template>
