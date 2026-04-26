<script setup>
import { computed, ref, watch } from 'vue'

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
  saveBossCycleQueue: { type: Function, required: true },
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
const queueDialogOpen = ref(false)
const queueDraft = ref([])
const templateDialogOpen = ref(false)
const templatePage = ref(1)
const pageSize = 5

watch(
  () => props.adminState?.bossCycleQueue,
  (value) => {
    queueDraft.value = Array.isArray(value) ? [...value] : []
  },
  { immediate: true },
)

watch(
  () => props.bossTemplates.length,
  () => {
    const maxPage = Math.max(1, totalTemplatePages.value)
    if (templatePage.value > maxPage) {
      templatePage.value = maxPage
    }
  },
)

const bossPartTotalHp = computed(() => sumBossPartMaxHp(props.bossForm.layout))
const gridParts = computed(() => {
  const grid = Array.from({ length: 5 }, () => Array(5).fill(null))
  const layout = Array.isArray(props.bossForm.layout) ? props.bossForm.layout : []
  for (const part of layout) {
    if (part.x >= 0 && part.x < 5 && part.y >= 0 && part.y < 5) {
      grid[part.y][part.x] = part
    }
  }
  return grid
})
const totalTemplatePages = computed(() => Math.max(1, Math.ceil(props.bossTemplates.length / pageSize)))
const pagedBossTemplates = computed(() => {
  const start = (templatePage.value - 1) * pageSize
  return props.bossTemplates.slice(start, start + pageSize)
})
const queueTemplateEntries = computed(() => {
  const map = new Map(props.bossTemplates.map((item) => [item.id, item]))
  return queueDraft.value
    .map((templateId) => ({ templateId, template: map.get(templateId) || null }))
    .filter((item) => item.template)
})
const queueTemplateSet = computed(() => new Set(queueDraft.value))

function sumBossPartMaxHp(layout) {
  if (!Array.isArray(layout) || layout.length === 0) {
    return 0
  }
  return layout.reduce((total, part) => total + Math.max(1, Number(part?.maxHp ?? 0)), 0)
}

function findPart(x, y) {
  if (!Array.isArray(props.bossForm.layout)) return null
  return props.bossForm.layout.find((p) => p.x === x && p.y === y)
}

function selectCell(x, y) {
  const existing = findPart(x, y)
  if (existing) {
    selectedCell.value = { ...existing }
    return
  }
  selectedCell.value = { x, y, type: 'soft', displayName: '', imagePath: '', maxHp: 1000, currentHp: 1000, armor: 0, alive: true }
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

function deleteCell() {
  if (!selectedCell.value || !Array.isArray(props.bossForm.layout)) return
  const cell = selectedCell.value
  const existing = props.bossForm.layout.findIndex((p) => p.x === cell.x && p.y === cell.y)
  if (existing >= 0) {
    props.bossForm.layout.splice(existing, 1)
  }
  selectedCell.value = null
}

function resetLootRows() {
  props.lootRows.splice(0, props.lootRows.length, { itemId: '', dropRatePercent: '' })
}

function openCreateTemplateDialog() {
  props.bossForm.id = ''
  props.bossForm.name = ''
  props.bossForm.maxHp = ''
  props.bossForm.goldOnKill = 0
  props.bossForm.stoneOnKill = 0
  props.bossForm.layout = []
  resetLootRows()
  selectedCell.value = null
  templateDialogOpen.value = true
}

function openEditTemplateDialog(entry) {
  props.editBossTemplate(entry)
  props.selectBossTemplate(entry.id)
  selectedCell.value = null
  templateDialogOpen.value = true
}

function closeTemplateDialog() {
  templateDialogOpen.value = false
  selectedCell.value = null
}

async function saveTemplateDesign() {
  const lootSnapshot = props.lootRows.map((entry) => ({
    itemId: entry?.itemId || '',
    dropRatePercent: entry?.dropRatePercent ?? '',
  }))
  await props.saveBossTemplate()
  await props.saveLoot(lootSnapshot)
  templateDialogOpen.value = false
}

function openQueueDialog() {
  queueDraft.value = Array.isArray(props.adminState?.bossCycleQueue) ? [...props.adminState.bossCycleQueue] : []
  queueDialogOpen.value = true
}

function closeQueueDialog() {
  queueDialogOpen.value = false
}

function addTemplateToQueue(templateId) {
  if (!templateId || queueTemplateSet.value.has(templateId)) return
  queueDraft.value = [...queueDraft.value, templateId]
}

function removeQueueTemplate(index) {
  if (index < 0 || index >= queueDraft.value.length) return
  const next = [...queueDraft.value]
  next.splice(index, 1)
  queueDraft.value = next
}

function moveQueueTemplate(index, direction) {
  const nextIndex = index + direction
  if (index < 0 || nextIndex < 0 || index >= queueDraft.value.length || nextIndex >= queueDraft.value.length) return
  const next = [...queueDraft.value]
  ;[next[index], next[nextIndex]] = [next[nextIndex], next[index]]
  queueDraft.value = next
}

async function submitBossCycleQueue() {
  await props.saveBossCycleQueue(queueDraft.value)
  queueDialogOpen.value = false
}

function goPrevPage() {
  if (templatePage.value > 1) templatePage.value -= 1
}

function goNextPage() {
  if (templatePage.value < totalTemplatePages.value) templatePage.value += 1
}
</script>

<template>
  <div class="admin-section">
    <section class="social-card admin-section-card boss-cycle-row">
      <div class="boss-cycle-row__left">
        <p class="vote-stage__eyebrow">Boss 循环设置</p>
        <strong>{{ bossCycleEnabled ? '循环已开启' : '循环未开启' }}</strong>
        <p class="social-card__copy">当前 Boss：{{ adminState.boss?.name || '暂无活动 Boss' }}</p>
      </div>
      <div class="boss-cycle-row__right">
        <button class="nickname-form__ghost" type="button" :disabled="saving" @click="openQueueDialog">
          设置循环队列
        </button>
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
    </section>

    <section class="social-card admin-section-card">
      <div class="social-card__head">
        <div>
          <p class="vote-stage__eyebrow">Boss 模板</p>
          <strong>{{ bossTemplates.length }} 个模板</strong>
        </div>
        <button class="nickname-form__submit" type="button" :disabled="saving" @click="openCreateTemplateDialog">
          新增模板 Boss
        </button>
      </div>

      <div class="boss-template-grid">
        <article
          v-for="entry in pagedBossTemplates"
          :key="entry.id"
          class="boss-template-card"
          @click="openEditTemplateDialog(entry)"
        >
          <div class="boss-template-card__head">
            <strong>{{ entry.name }}</strong>
            <span>{{ entry.id }}</span>
          </div>
          <p>血量 {{ entry.maxHp }} · 部位 {{ entry.layout?.length || 0 }}</p>
          <p>掉落 {{ entry.loot.length }} · 金币 {{ entry.goldOnKill || 0 }} · 强化石 {{ entry.stoneOnKill || 0 }}</p>
          <button class="inventory-item__action" type="button" @click.stop="openEditTemplateDialog(entry)">编辑模板</button>
        </article>
      </div>

      <div class="boss-template-pagination">
        <button class="nickname-form__ghost" type="button" :disabled="templatePage <= 1" @click="goPrevPage">上一页</button>
        <span>第 {{ templatePage }} / {{ totalTemplatePages }} 页（每行 5 个）</span>
        <button class="nickname-form__ghost" type="button" :disabled="templatePage >= totalTemplatePages" @click="goNextPage">下一页</button>
      </div>
    </section>

    <div v-if="templateDialogOpen" class="dialog-mask">
      <section class="dialog-body">
        <div class="dialog-head">
          <strong>{{ bossForm.id ? `编辑模板：${bossForm.name || bossForm.id}` : '新增模板 Boss' }}</strong>
          <button class="nickname-form__ghost" type="button" @click="closeTemplateDialog">关闭</button>
        </div>

        <div class="dialog-grid">
          <div class="dialog-panel">
            <p class="vote-stage__eyebrow">模板基础信息与部位</p>
            <div class="admin-form">
              <input v-model="bossForm.id" class="nickname-form__input" type="text" placeholder="模板 ID，如 dragon" />
              <input v-model="bossForm.name" class="nickname-form__input" type="text" placeholder="Boss 显示名称" />
              <input v-model.number="bossForm.goldOnKill" class="nickname-form__input" type="number" min="0" placeholder="击杀金币基准" />
              <input v-model.number="bossForm.stoneOnKill" class="nickname-form__input" type="number" min="0" placeholder="击杀强化石基准" />
              <input class="nickname-form__input" type="number" min="0" :value="bossPartTotalHp" readonly aria-label="Boss 总血量" />
            </div>

            <fieldset class="admin-fieldset">
              <legend class="admin-fieldset__legend">部位布局（点击网格格子编辑）</legend>
              <div class="boss-editor-grid">
                <div v-for="yi in 5" :key="'row-' + yi" class="boss-editor-grid__row">
                  <div
                    v-for="xi in 5"
                    :key="'cell-' + yi + '-' + xi"
                    class="boss-editor-cell"
                    :class="{
                      'boss-editor-cell--filled': gridParts[yi - 1][xi - 1],
                      'boss-editor-cell--selected': selectedCell?.x === xi - 1 && selectedCell?.y === yi - 1,
                    }"
                    :style="gridParts[yi - 1][xi - 1] ? { '--cell-color': partTypeColors[gridParts[yi - 1][xi - 1].type] || '#64748b' } : {}"
                    @click="selectCell(xi - 1, yi - 1)"
                  >
                    <template v-if="gridParts[yi - 1][xi - 1]">
                      <span class="boss-editor-cell__type">
                        {{ gridParts[yi - 1][xi - 1].displayName || partTypeLabels[gridParts[yi - 1][xi - 1].type] || gridParts[yi - 1][xi - 1].type }}
                      </span>
                      <span class="boss-editor-cell__hp">{{ gridParts[yi - 1][xi - 1].maxHp }}</span>
                    </template>
                    <span v-else class="boss-editor-cell__empty">+</span>
                  </div>
                </div>
              </div>

              <div v-if="selectedCell" class="boss-editor-inspector">
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
                    <span>名称</span>
                    <input v-model="selectedCell.displayName" class="nickname-form__input" type="text" placeholder="例如：眼核" />
                  </label>
                  <label class="boss-editor-inspector__field">
                    <span>图片</span>
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
                  <button class="nickname-form__ghost" type="button" @click="deleteCell">删除</button>
                </div>
              </div>
            </fieldset>
          </div>

          <div class="dialog-panel">
            <p class="vote-stage__eyebrow">模板掉落设置</p>
            <p v-if="!hasEquipmentTemplates" class="feedback">先在装备页创建模板后再配置掉落。</p>
            <div class="admin-form admin-form--tight">
              <div v-for="(entry, index) in lootRows" :key="`${bossForm.id}-${index}-${entry.itemId}`" class="admin-inline-row">
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
              </div>
            </div>
          </div>
        </div>

        <div class="dialog-footer">
          <button class="nickname-form__ghost" type="button" @click="closeTemplateDialog">取消</button>
          <button class="nickname-form__submit" type="button" :disabled="saving" @click="saveTemplateDesign">保存模板（含掉落与部位）</button>
        </div>
      </section>
    </div>

    <div v-if="queueDialogOpen" class="dialog-mask">
      <section class="dialog-body dialog-body--queue">
        <div class="dialog-head">
          <strong>设置循环队列</strong>
          <button class="nickname-form__ghost" type="button" @click="closeQueueDialog">关闭</button>
        </div>
        <div class="queue-grid">
          <div class="queue-panel">
            <p class="vote-stage__eyebrow">当前队列</p>
            <ul class="queue-list">
              <li v-for="(entry, index) in queueTemplateEntries" :key="entry.templateId" class="queue-item">
                <div class="queue-item__meta">
                  <strong>{{ entry.template.name }}</strong>
                  <span>{{ entry.template.id }}</span>
                </div>
                <div class="queue-item__actions">
                  <button class="nickname-form__ghost" type="button" :disabled="index === 0" @click="moveQueueTemplate(index, -1)">上移</button>
                  <button class="nickname-form__ghost" type="button" :disabled="index === queueTemplateEntries.length - 1" @click="moveQueueTemplate(index, 1)">下移</button>
                  <button class="nickname-form__ghost" type="button" @click="removeQueueTemplate(index)">移除</button>
                </div>
              </li>
            </ul>
          </div>
          <div class="queue-panel">
            <p class="vote-stage__eyebrow">可选模板</p>
            <ul class="queue-template-list">
              <li v-for="entry in bossTemplates" :key="entry.id" class="queue-item">
                <div class="queue-item__meta">
                  <strong>{{ entry.name }}</strong>
                  <span>{{ entry.id }}</span>
                </div>
                <button class="nickname-form__ghost" type="button" :disabled="queueTemplateSet.has(entry.id)" @click="addTemplateToQueue(entry.id)">
                  {{ queueTemplateSet.has(entry.id) ? '已加入' : '加入队列' }}
                </button>
              </li>
            </ul>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="nickname-form__submit" type="button" :disabled="saving" @click="submitBossCycleQueue">保存循环队列</button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.boss-cycle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.boss-cycle-row__right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.boss-template-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 0.5rem;
}
.boss-template-card {
  border: 1px solid rgba(203, 213, 225, 0.8);
  border-radius: 10px;
  padding: 0.6rem;
  background: #ffffff;
  cursor: pointer;
  color: #1e293b;
}
.boss-template-card:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}
.boss-template-card__head {
  display: flex;
  flex-direction: column;
  margin-bottom: 0.25rem;
}
.boss-template-card__head span {
  color: #64748b;
  font-size: 0.75rem;
}
.boss-template-card p {
  margin: 0.15rem 0;
  font-size: 0.78rem;
}
.boss-template-pagination {
  margin-top: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

/* 弹窗遮罩 - 改亮色 */
.dialog-mask {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgba(100, 116, 139, 0.25);
  padding: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
/* 弹窗主体 - 白色亮色 */
.dialog-body {
  width: min(1200px, 100%);
  max-height: 90vh;
  overflow: auto;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  padding: 1rem;
  color: #1e293b;
}
.dialog-body--queue {
  width: min(980px, 100%);
}
.dialog-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
  color: #0f172a;
}
.dialog-grid {
  display: grid;
  grid-template-columns: 1.1fr 1fr;
  gap: 0.75rem;
}
/* 侧边面板 亮色 */
.dialog-panel {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.75rem;
  background: #f8fafc;
}

/* ============================================== */
/* 【重点】以下全部保留你原版深色部位布局，完全没改 */
/* ============================================== */
.boss-editor-grid {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  background: var(--surface-2, #1e293b);
  border-radius: 8px;
  margin-top: 0.5rem;
}
.boss-editor-grid__row {
  display: flex;
  gap: 4px;
}
.boss-editor-cell {
  width: 100%;
  aspect-ratio: 1;
  min-height: 52px;
  border-radius: 6px;
  border: 2px solid transparent;
  background: var(--surface-1, #0f172a);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 0.68rem;
  cursor: pointer;
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
}
.boss-editor-cell__hp {
  color: var(--text-2, #94a3b8);
  font-size: 0.62rem;
}
.boss-editor-cell__empty {
  color: var(--text-3, #64748b);
  font-size: 1.2rem;
  font-weight: 300;
}
/* 部位右侧属性面板 单独微调浅色，不影响格子 */
.boss-editor-inspector {
  margin-top: 0.6rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.6rem;
  background: #ffffff;
}
.boss-editor-inspector__form {
  display: grid;
  gap: 0.45rem;
}
.boss-editor-inspector__field {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.boss-editor-inspector__field span {
  min-width: 46px;
  color: #334155;
  font-size: 0.8rem;
}
.boss-editor-inspector__actions {
  margin-top: 0.5rem;
  display: flex;
  gap: 0.45rem;
}

/* 队列区域 亮色 */
.queue-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.queue-panel {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.75rem;
  background: #f8fafc;
}
.queue-list,
.queue-template-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.45rem;
}
.queue-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.45rem;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.queue-item__meta span {
  display: block;
  color: #64748b;
  font-size: 0.75rem;
}
.queue-item__actions {
  display: flex;
  gap: 0.35rem;
}

.dialog-footer {
  margin-top: 0.75rem;
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

@media (max-width: 1200px) {
  .boss-template-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (max-width: 900px) {
  .boss-cycle-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .dialog-grid,
  .queue-grid {
    grid-template-columns: 1fr;
  }
  .boss-template-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
