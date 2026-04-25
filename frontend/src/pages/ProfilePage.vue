<script setup>
import {usePublicPageState} from './publicPageState'

const {
  AUTO_CLICK_RATE_LABEL,
  inventory,
  loadout,
  loadoutSlots,
  combatStats,
  nickname,
  nicknameDraft,
  passwordDraft,
  errorMessage,
  actioningItemId,
  activeHudTab,
  autoClickEnabled,
  autoClickTargetLabel,
  canStartAutoClick,
  autoClickStatus,
  gems,
  profileLoading,
  profileNotice,
  isLoggedIn,
  myClicks,
  myRank,
  myBossDamage,
  normalDamage,
  criticalDamage,
  equippedItems,
  formatRarityLabel,
  formatItemStats,
  formatItemStatLines,
  equipmentNameParts,
  equipmentNameClass,
  toggleItemEquip,
  formatNumber,
  formatStatWithDelta,
  formatPercentWithDelta,
  submitMessage,
  selectHudTab,
  toggleAutoClick,
  submitNickname,
  resetNickname,
} = usePublicPageState()
</script>

<template>
<section class="stage-layout stage-layout--single">
      <aside class="player-hud player-hud--page">
        <section class="player-hud__shell">
          <div class="player-hud__head">
            <div>
              <p class="vote-stage__eyebrow">角色资料</p>
              <strong>{{ isLoggedIn ? nickname : '未登录角色' }}</strong>
            </div>
            <span class="player-hud__pill">{{ isLoggedIn ? '已上墙' : '访客' }}</span>
          </div>

          <div class="player-hud__tabs">
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'account' }"
                type="button"
                @click="selectHudTab('account')"
            >
              账号
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'auto' }"
                type="button"
                @click="selectHudTab('auto')"
            >
              挂机
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'inventory' }"
                type="button"
                @click="selectHudTab('inventory')"
            >
              背包
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'stats' }"
                type="button"
                @click="selectHudTab('stats')"
            >
              属性
            </button>
            <button
                class="player-hud__tab"
                :class="{ 'player-hud__tab--active': activeHudTab === 'loadout' }"
                type="button"
                @click="selectHudTab('loadout')"
            >
              装备栏
            </button>
          </div>

          <p v-if="profileLoading" class="feedback">资料刷新中...</p>
          <div class="player-hud__content">
            <section v-if="activeHudTab === 'account'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">账号</p>
                <strong>{{ isLoggedIn ? nickname : '未登录' }}</strong>
              </div>

              <p class="player-hud__note">{{ profileNotice || (isLoggedIn ? `当前角色：${nickname}` : '登录后可查看背包、属性、装备栏和官方挂机状态。') }}</p>

              <form class="nickname-form player-hud__form" @submit.prevent="submitNickname">
                <input
                    v-model="nicknameDraft"
                    class="nickname-form__input"
                    type="text"
                    maxlength="20"
                    placeholder="比如：阿明"
                />
                <input
                    v-model="passwordDraft"
                    class="nickname-form__input"
                    type="password"
                    placeholder="输入密码"
                />
                <button class="nickname-form__submit" type="submit">
                  {{ isLoggedIn ? '切换账号' : '登录 / 首次认领' }}
                </button>
              </form>

              <button
                  v-if="isLoggedIn"
                  class="nickname-form__ghost player-hud__reset"
                  type="button"
                  @click="resetNickname"
              >
                退出登录
              </button>
            </section>

            <section v-else-if="activeHudTab === 'auto'" class="player-hud__panel player-hud__auto">
              <div class="player-hud__section-head">
                <div>
                  <p class="vote-stage__eyebrow">挂机</p>
                  <strong>官方挂机托管</strong>
                </div>
                <span class="player-hud__pill" :class="{ 'player-hud__pill--active': autoClickEnabled }">
                  {{ autoClickEnabled ? '运行中' : '未开启' }}
                </span>
              </div>

              <p class="player-hud__note">{{ autoClickStatus }}</p>

              <div class="player-hud__auto-meta">
                <span class="player-hud__auto-chip">目标：{{ autoClickTargetLabel }}</span>
                <span class="player-hud__auto-chip">{{ AUTO_CLICK_RATE_LABEL }}</span>
              </div>

              <button
                  class="nickname-form__submit player-hud__auto-button"
                  type="button"
                  :disabled="!autoClickEnabled && !canStartAutoClick"
                  @click="toggleAutoClick"
              >
                {{ autoClickEnabled ? '关闭挂机' : '开启挂机' }}
              </button>
            </section>

            <section v-else-if="activeHudTab === 'inventory'" class="player-hud__panel">
              <div class="player-hud__section-head">
                <p class="vote-stage__eyebrow">背包</p>
                <strong>{{ inventory.length }} 件</strong>
              </div>

              <div v-if="inventory.length === 0" class="leaderboard-list leaderboard-list--empty">
                <p>先去打 Boss 或等后台发装备，背包就会慢慢满起来。</p>
              </div>

              <ul v-else class="inventory-list">
                <li v-for="item in inventory" :key="item.itemId" class="inventory-item inventory-item--panel">
                  <div class="inventory-item__top">
                    <div class="inventory-item__main">
                      <strong>
                        <span v-if="equipmentNameParts(item).prefix">{{ equipmentNameParts(item).prefix }}</span>
                        <span :class="equipmentNameClass(item)">{{ equipmentNameParts(item).text }}</span>
                      </strong>
                      <div class="inventory-item__meta">
                        <span class="inventory-item__chip">{{ formatRarityLabel(item.rarity) }}</span>
                        <span class="inventory-item__chip">类型:{{ item.slot || '未分类' }}</span>
                        <span class="inventory-item__chip">库存:{{ item.quantity }}</span>
                      </div>
                    </div>
                  </div>

                  <ul class="inventory-item__stats inventory-item__stats--stacked">
                    <li v-for="line in formatItemStatLines(item)" :key="line">
                      {{ line }}
                    </li>
                  </ul>

                  <div class="inventory-item__footer">
                    <span
                        class="inventory-item__state"
                        :class="{ 'inventory-item__state--active': item.equipped }"
                    >
                      {{ item.equipped ? '已穿戴' : '待命中' }}
                    </span>

                    <div class="inventory-item__actions">
                      <button
                          class="inventory-item__action"
                          type="button"
                          :disabled="!isLoggedIn || actioningItemId === item.itemId"
                          @click="toggleItemEquip(item.itemId, item.equipped)"
                      >
                        {{ item.equipped ? '卸下' : '穿戴' }}
                      </button>
                    </div>
                  </div>
                </li>
              </ul>
            </section>

            <section v-else-if="activeHudTab === 'stats'" class="player-hud__panel">
              <div class="player-hud__section-head">
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

            <section v-else-if="activeHudTab === 'loadout'" class="player-hud__panel">
              <div class="player-hud__section-head">
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

          </div>
        </section>
      </aside>


    </section>
</template>
