<script setup>
import {computed, ref} from 'vue'

const props = defineProps({
  adminState: { type: Object, required: true },
  bossCycleEnabled: { type: Boolean, required: true },
  bossForm: { type: Object, required: true },
  bossTemplates: { type: Array, required: true },
  equipmentOptions: { type: Array, required: true },
  hasBoss: { type: Boolean, required: true },
  hasEquipmentTemplates: { type: Boolean, required: true },
  lootRows: { type: Array, required: true },
  saving: { type: Boolean, required: true },
  selectedBossTemplate: { type: Object, default: null },
  selectedBossTemplateId: { type: String, required: true },
  addLootRow: { type: Function, required: true },
  deactivateBoss: { type: Function, required: true },
  deleteBossTemplate: { type: Function, required: true },
  disableBossCycle: { type: Function, required: true },
  editBossTemplate: { type: Function, required: true },
  enableBossCycle: { type: Function, required: true },
  findEquipmentTemplate: { type: Function, required: true },
  formatItemStats: { type: Function, required: true },
  removeLootRow: { type: Function, required: true },
  saveBossTemplate: { type: Function, required: true },
  saveLoot: { type: Function, required: true },
  selectBossTemplate: { type: Function, required: true },
})

const partTypeLabels = { soft: '软组织', heavy: '重甲', weak: '弱点' }
const partTypeColors = { soft: '#4ade80', heavy: '#fbbf24', weak: '#f472b6' }

const selectedCell = ref(null)

const gridParts = computed(() => {
  const grid = Array.from({length: 5}, () => Array(5).fill(null))
  const layout = Array.isArray(props.bossForm.layout) ? props.bossForm.layout : []
  for (const part of layout) {
    if (part.x >= 0 && part.x < 5 && part.y >= 0 && part.y < 5) {
      grid[part.y][part.x] = part
    }
  }
  return grid
})

const bossPartTotalHp = computed(() => sumBossPartMaxHp(props.bossForm.layout))

function sumBossPartMaxHp(layout) {
  if (!Array.isArray(layout) || layout.length === 0) {
    return 0
  }
  return layout.reduce((total, part) => total + Math.max(1, Number(part?.maxHp ?? 0)), 0)
}

function selectCell(x, y) {
  const existing = findPart(x, y)
  if (existing) {
    selectedCell.value = { ...existing }
    return
  }
  selectedCell.value = { x, y, type: 'soft', displayName: '', imagePath: '', maxHp: 1000, currentHp: 1000, armor: 0, alive: true }
}

function findPart(x, y) {
  if (!Array.isArray(props.bossForm.layout)) return null
  return props.bossForm.layout.find((p) => p.x === x && p.y === y)
}

function saveCell() {
  if (!selectedCell.value) return
  const cell = normalizeBossPartCell(selectedCell.value)
  if (!Array.isArray(props.bossForm.layout)) {
    props.bossForm.layout = []
  }
  const existing = props.bossForm.layout.findIndex((p) => p.x === cell.x && p.y === cell.y)
  if (existing >= 0) {
    props.bossForm.layout[existing] = { ...cell }
  } else {
    props.bossForm.layout.push({ ...cell })
  }
  selectedCell.value = null
}

function normalizeBossPartCell(cell) {
  const maxHp = Math.max(1, Number(cell.maxHp || 1))
  return {
    ...cell,
    type: cell.type || 'soft',
    displayName: String(cell.displayName || '').trim(),
    imagePath: String(cell.imagePath || '').trim(),
    maxHp,
    currentHp: maxHp,
    armor: Math.max(0, Number(cell.armor || 0)),
    alive: true,
  }
}

function deleteCell() {
  if (!selectedCell.value) return
  const cell = selectedCell.value
  if (!Array.isArray(props.bossForm.layout)) return
  const existing = props.bossForm.layout.findIndex((p) => p.x === cell.x && p.y === cell.y)
  if (existing >= 0) {
    props.bossForm.layout.splice(existing, 1)
  }
  selectedCell.value = null
}

function cancelCell() {
  selectedCell.value = null
}
</script>

<template>
  <div class="admin-section">
    <div class="admin-grid">
      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">循环状态</p>
          <strong>{{ bossCycleEnabled ? '循环已开启' : '循环未开启' }}</strong>
        </div>

        <p class="social-card__copy">当前 Boss：{{ adminState.boss?.name || '暂无活动 Boss' }}</p>
        <div class="admin-cycle-pills">
          <span class="boss-stage__pill">
            {{ bossCycleEnabled ? '击败后会立即补下一只' : '击败后不会自动补位' }}
          </span>
          <span class="boss-stage__pill">
            {{ adminState.boss?.templateId ? `来源模板 ${adminState.boss.templateId}` : '当前没有绑定模板' }}
          </span>
        </div>

        <div v-if="hasBoss" class="admin-boss-summary">
          <p>实例 ID：{{ adminState.boss.id }}</p>
          <p>状态：{{ adminState.boss.status }} · 血量 {{ adminState.boss.currentHp }}/{{ adminState.boss.maxHp }}</p>
        </div>
        <p v-else class="feedback" style="margin-top: 0.75rem;">
          开启循环后，如果当前没有 Boss，会立刻从 Boss 池里随机刷出一只。
        </p>

        <div class="admin-inline-actions" style="margin-top: 1rem;">
          <button class="nickname-form__submit" type="button" :disabled="saving || bossCycleEnabled" @click="enableBossCycle">
            开启循环
          </button>
          <button class="nickname-form__ghost" type="button" :disabled="saving || !bossCycleEnabled" @click="disableBossCycle">
            停止循环
          </button>
          <button v-if="hasBoss" class="nickname-form__ghost" type="button" :disabled="saving" @click="deactivateBoss">
            {{ bossCycleEnabled ? '跳过当前 Boss' : '关闭当前 Boss' }}
          </button>
        </div>

        <div v-if="hasBoss && adminState.loot.length > 0" style="margin-top: 1rem;">
          <p class="vote-stage__eyebrow">当前实例掉落快照</p>
          <ul class="inventory-list">
            <li v-for="item in adminState.loot" :key="item.itemId" class="inventory-item">
              <div>
                <strong>{{ item.itemName || item.itemId }}</strong>
                <p>{{ item.itemId }} · {{ item.slot }} · 掉落几率 {{ item.dropRatePercent }}%</p>
                <p>{{ formatItemStats(item) }}</p>
              </div>
            </li>
          </ul>
        </div>

      </section>

      <section class="social-card">
        <div class="social-card__head">
          <p class="vote-stage__eyebrow">Boss 池模板</p>
          <strong>{{ bossTemplates.length }} 只</strong>
        </div>

        <form class="admin-form" @submit.prevent="saveBossTemplate">
          <input v-model="bossForm.id" class="nickname-form__input" type="text" placeholder="模板 ID，如 dragon" />
          <input v-model="bossForm.name" class="nickname-form__input" type="text" placeholder="Boss 显示名称" />
          <input
            class="nickname-form__input"
            type="number"
            min="0"
            :value="bossPartTotalHp"
            readonly
            placeholder="总血量"
            aria-label="Boss 总血量"
          />
          <p class="feedback">总血量由部位总血量决定。</p>
          <fieldset class="admin-fieldset">
            <legend class="admin-fieldset__legend">部位布局（点击网格格子编辑部位）</legend>
            <div class="boss-editor-grid">
              <div v-for="yi in 5" :key="'row-'+yi" class="boss-editor-grid__row">
                <div
                  v-for="xi in 5"
                  :key="'cell-'+yi+'-'+xi"
                  class="boss-editor-cell"
                  :class="{
                    'boss-editor-cell--filled': gridParts[yi-1][xi-1],
                    'boss-editor-cell--selected': selectedCell?.x === xi-1 && selectedCell?.y === yi-1,
                  }"
                  :style="gridParts[yi-1][xi-1] ? { '--cell-color': partTypeColors[gridParts[yi-1][xi-1].type] || '#64748b' } : {}"
                  @click="selectCell(xi-1, yi-1)"
                >
                  <template v-if="gridParts[yi-1][xi-1]">
                    <span class="boss-editor-cell__type">
                      {{ gridParts[yi-1][xi-1].displayName || partTypeLabels[gridParts[yi-1][xi-1].type] || gridParts[yi-1][xi-1].type }}
                    </span>
                    <span class="boss-editor-cell__hp">{{ gridParts[yi-1][xi-1].maxHp }}</span>
                  </template>
                  <span v-else class="boss-editor-cell__empty">+</span>
                </div>
              </div>
            </div>
            <div v-if="selectedCell" class="boss-editor-inspector">
              <div class="boss-editor-inspector__head">
                <strong>部位 ({{ selectedCell.x }}, {{ selectedCell.y }})</strong>
                <button class="nickname-form__ghost" type="button" @click="cancelCell">取消</button>
              </div>
              <div class="boss-editor-inspector__form">
                <label class="boss-editor-inspector__field">
                  <span>类型</span>
                  <select v-model="selectedCell.type" class="nickname-form__input">
                    <option value="soft">软组织</option>
                    <option value="heavy">重甲</option>
                    <option value="weak">弱点</option>
                  </select>
                </label>
                <label class="boss-editor-inspector__field">
                  <span>部位名称</span>
                  <input v-model="selectedCell.displayName" class="nickname-form__input" type="text" placeholder="例如：眼核" />
                </label>
                <label class="boss-editor-inspector__field">
                  <span>小图路径</span>
                  <input v-model="selectedCell.imagePath" class="nickname-form__input" type="text" placeholder="/images/boss/eye.png" />
                </label>
                <label class="boss-editor-inspector__field">
                  <span>血量</span>
                  <input v-model="selectedCell.maxHp" class="nickname-form__input" type="number" min="1" />
                </label>
                <label class="boss-editor-inspector__field">
                  <span>护甲</span>
                  <input v-model="selectedCell.armor" class="nickname-form__input" type="number" min="0" />
                </label>
              </div>
              <div class="boss-editor-inspector__actions">
                <button class="nickname-form__submit" type="button" @click="saveCell">保存部位</button>
                <button class="nickname-form__ghost boss-editor-inspector__delete" type="button" @click="deleteCell">删除</button>
              </div>
            </div>
          </fieldset>
          <button class="nickname-form__submit" type="submit" :disabled="saving">保存 Boss 模板</button>
        </form>

        <ul class="inventory-list">
          <li v-for="entry in bossTemplates" :key="entry.id" class="inventory-item inventory-item--stacked">
            <div>
              <strong>{{ entry.name }}</strong>
              <p>{{ entry.id }} · 血量 {{ entry.maxHp }} · 部位 {{ entry.layout?.length || 0 }} 个 · 装备 {{ entry.loot.length }} 件</p>
            </div>
            <div class="admin-inline-actions admin-inline-actions--stacked">
              <button class="inventory-item__action" type="button" @click="selectBossTemplate(entry.id)">编辑掉落</button>
              <button class="inventory-item__action" type="button" @click="editBossTemplate(entry)">编辑模板</button>
              <button class="nickname-form__ghost" type="button" @click="deleteBossTemplate(entry.id)">删除</button>
            </div>
          </li>
        </ul>
      </section>
    </div>

    <section class="social-card admin-section-card">
      <div class="social-card__head">
        <p class="vote-stage__eyebrow">模板掉落池</p>
        <strong>{{ selectedBossTemplate?.name || selectedBossTemplateId || '未选择模板' }}</strong>
      </div>

      <p class="feedback" style="margin-bottom: 0.75rem;">
        掉落池保存到模板上。Boss 刷出来时会复制一份到当前实例，所以你后面再改模板，不会改到场上的那只。
      </p>

      <p v-if="!hasEquipmentTemplates" class="feedback" style="margin-bottom: 0.75rem;">
        当前还没有装备模板，先去“装备”页创建装备，再回来配置掉落池。
      </p>

      <div class="admin-form admin-form--tight">
        <div v-for="(entry, index) in lootRows" :key="`${selectedBossTemplateId}-${index}-${entry.itemId}`" class="admin-inline-row">
          <div class="admin-loot-select">
            <select v-model="entry.itemId" class="nickname-form__input" :disabled="!hasEquipmentTemplates && !entry.itemId">
              <option value="">选择已有装备</option>
              <option v-if="entry.itemId && !findEquipmentTemplate(entry.itemId)" :value="entry.itemId">
                {{ entry.itemId }}（已删除的装备）
              </option>
              <option v-for="item in equipmentOptions" :key="item.itemId" :value="item.itemId">
                {{ item.name }} · {{ item.itemId }} · {{ item.slot }}
              </option>
            </select>
            <p v-if="findEquipmentTemplate(entry.itemId)" class="admin-loot-select__meta">
              {{ formatItemStats(findEquipmentTemplate(entry.itemId)) }}
            </p>
          </div>
          <input v-model="entry.dropRatePercent" class="nickname-form__input" type="number" min="0" max="100" step="0.01" placeholder="掉落几率 %" />
          <button class="nickname-form__ghost" type="button" @click="removeLootRow(index)">删</button>
        </div>
        <div class="admin-inline-actions">
          <button class="nickname-form__ghost" type="button" @click="addLootRow">加一行</button>
          <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveLoot">保存模板掉落池</button>
        </div>
      </div>
    </section>

  </div>
</template>

<style scoped>
.boss-editor-grid {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  background: var(--surface-2, #1e293b);
  border-radius: 8px;
  margin-bottom: 8px;
}
.boss-editor-grid__row {
  display: flex;
  gap: 4px;
}
.boss-editor-cell {
  width: 100%;
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border-radius: 6px;
  background: var(--surface-1, #0f172a);
  border: 2px solid transparent;
  cursor: pointer;
  font-size: 0.7rem;
  line-height: 1.2;
  transition: border-color 0.15s, background 0.15s;
  min-height: 56px;
}
.boss-editor-cell:hover {
  border-color: var(--accent, #38bdf8);
}
.boss-editor-cell--filled {
  background: color-mix(in srgb, var(--cell-color, #64748b) 25%, var(--surface-1, #0f172a));
  border-color: var(--cell-color, #64748b);
}
.boss-editor-cell--selected {
  box-shadow: 0 0 0 2px var(--accent, #38bdf8);
}
.boss-editor-cell__type {
  font-weight: 600;
  color: var(--cell-color, #94a3b8);
  font-size: 0.65rem;
}
.boss-editor-cell__hp {
  color: var(--text-2, #94a3b8);
  font-size: 0.6rem;
}
.boss-editor-cell__empty {
  color: var(--text-3, #64748b);
  font-size: 1.2rem;
  font-weight: 300;
}
.boss-editor-inspector {
  background: var(--surface-2, #1e293b);
  border-radius: 8px;
  padding: 12px;
  margin-top: 8px;
}
.boss-editor-inspector__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.boss-editor-inspector__head strong {
  font-size: 0.85rem;
}
.boss-editor-inspector__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 10px;
}
.boss-editor-inspector__field {
  display: flex;
  align-items: center;
  gap: 8px;
}
.boss-editor-inspector__field span {
  min-width: 48px;
  font-size: 0.8rem;
  color: var(--text-2, #94a3b8);
}
.boss-editor-inspector__field .nickname-form__input {
  flex: 1;
}
.boss-editor-inspector__actions {
  display: flex;
  gap: 8px;
}
.boss-editor-inspector__delete {
  color: #f87171 !important;
}
</style>
