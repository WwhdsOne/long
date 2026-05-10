import {readFileSync} from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'

import {describe, expect, it} from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const pageSource = readFileSync(path.resolve(currentDir, './ArmoryPage.vue'), 'utf8')
const stateSource = readFileSync(path.resolve(currentDir, './publicPageState.js'), 'utf8')

describe('ArmoryPage 战斗属性与装备栏', () => {
    it('强化上限按当前稀有度规则展示', () => {
        expect(pageSource).toContain("case '优秀':\n      return 10")
        expect(pageSource).toContain("case '稀有':\n      return 15")
        expect(pageSource).toContain("case '史诗':\n      return 20")
        expect(pageSource).toContain("case '传说':\n      return 25")
        expect(pageSource).toContain("case '至臻':\n      return 35")
        expect(pageSource).toContain("default:\n      return 5")
    })

    it('战斗属性改为紧凑两列摘要，不再拆分基础属性和强化属性', () => {
        expect(pageSource).not.toContain('基础属性')
        expect(pageSource).not.toContain('强化属性')
        expect(pageSource).not.toContain('criticalCount')
        expect(pageSource).toContain('const combatStatSummaryItems = computed(() => [')
        expect(pageSource).toContain('class="armory-combat-summary"')
        expect(pageSource).toContain('class="armory-combat-summary__item"')
        expect(pageSource).toContain('formatArmorPenPercent')
        expect(pageSource).toContain('formatCritDamageBonus')
        expect(pageSource).toContain('function formatMagicProcRateValue(value) {')
        expect(pageSource).toContain("{label: '魔法触发率', value: formatMagicProcRateValue(combatStats.value?.magicProcRate ?? 0)}")
    })

    it('装备栏改为左右 3 + 3 的正方形槽位，并复用装备详情浮层', () => {
        expect(pageSource).toContain('const loadoutColumns = computed(() => [')
        expect(pageSource).toContain('class="loadout-grid loadout-grid--paired"')
        expect(pageSource).toContain('class="loadout-slot__visual"')
        expect(pageSource).toContain('class="loadout-slot__placeholder"')
        expect(pageSource).toContain('class="armory-item-tooltip"')
        expect(pageSource).toContain('item.imagePath')
        expect(pageSource).toContain('词条：{{ loadout[slot.value].affixCount || 0 }} / {{ loadout[slot.value].affixLimit || 0 }}')
        expect(pageSource).toContain('词条：{{ item.affixCount || 0 }} / {{ item.affixLimit || 0 }}')
        expect(pageSource).toContain('尚未铭刻')
    })

    it('强化弹窗支持滑条预览与批量提交', () => {
        expect(pageSource).toContain('type="range"')
        expect(pageSource).toContain('enhancePreviewStatRows')
        expect(pageSource).toContain('enhanceAffordableLevelsByStone')
        expect(pageSource).toContain('enhanceSelectedLevels')
        expect(pageSource).toContain('->')
        expect(stateSource).toContain('async function enhanceItem(instanceId, levels = 1)')
        expect(stateSource).toContain('JSON.stringify({nickname: nickname.value, levels})')
    })

    it('强化预期中的比例属性按真实百分比展示', () => {
        expect(pageSource).toContain('function formatCritDamageExtraBonus(value) {')
        expect(pageSource).toContain('return `+${formatTrimmedNumber(Number(value ?? 0) * 100, 2)}%`')
        expect(pageSource).toContain('const enhanceMagicProcRateStep = 0.001')
        expect(pageSource).toContain('const enhanceFlatPercentStep = 0.001')
        expect(pageSource).toContain('function previewFlatStepStat(currentValue, currentLevel, targetLevel) {')
        expect(pageSource).toContain('function previewMagicProcRateBonus(currentValue, currentLevel, targetLevel) {')
        expect(pageSource).toContain("pushEnhancePreviewRow(rows, '暴击倍率', item.critDamageMultiplier, preview.critDamageMultiplier, formatCritDamageExtraBonus)")
        expect(pageSource).toContain("pushEnhancePreviewRow(rows, '软组织伤害', item.partTypeDamageSoft, preview.partTypeDamageSoft, formatRatioPercentValue)")
        expect(pageSource).toContain("pushEnhancePreviewRow(rows, '重甲伤害', item.partTypeDamageHeavy, preview.partTypeDamageHeavy, formatRatioPercentValue)")
        expect(pageSource).toContain("pushEnhancePreviewRow(rows, '弱点伤害', item.partTypeDamageWeak, preview.partTypeDamageWeak, formatRatioPercentValue)")
        expect(pageSource).toContain('preview.armorPenPercent = previewFlatStepStat(preview.armorPenPercent, currentLevel, targetLevel)')
        expect(pageSource).toContain('preview.critRate = previewFlatStepStat(preview.critRate, currentLevel, targetLevel)')
        expect(pageSource).toContain('preview.magicProcRateBonus = previewMagicProcRateBonus(preview.magicProcRateBonus, currentLevel, targetLevel)')
        expect(pageSource).toContain("pushEnhancePreviewRow(rows, '魔法触发率', item.magicProcRateBonus, preview.magicProcRateBonus, formatMagicProcRateValue)")
    })

    it('一键分解预估与规则文案同步保留每种装备一件最高强化', () => {
        expect(pageSource).toContain('const groupedCandidates = new Map()')
        expect(pageSource).toContain('excludedKeptHighest += 1')
        expect(pageSource).toContain("keepCandidates.sort((left, right) => left.instanceId.localeCompare(right.instanceId))")
        expect(pageSource).toContain('即将分解 {{ bulkSalvageConfirmData.total }} 件装备')
        expect(pageSource).toContain('按规则保留 {{ bulkSalvageConfirmData.excludedKeptHighest }} 件')
        expect(pageSource).toContain('一键分解会按 itemId 分组，每种装备保留 1 件最高强化。')
        expect(pageSource).toContain('若最高强化并列，则随机保留其中 1 件。')
    })

    it('右键菜单和确认弹窗支持铭刻，并明确不可撤销与刻印石消耗', () => {
        expect(pageSource).toContain('handleContextInscribe')
        expect(pageSource).toContain('铭刻')
        expect(pageSource).toContain('确认铭刻')
        expect(pageSource).toContain('将消耗 刻印石 x1')
        expect(pageSource).toContain('铭刻会为该装备随机新增 1 条词条')
        expect(pageSource).toContain('铭刻后词条将与该装备永久绑定，无法撤销')
        expect(pageSource).toContain('当前词条：{{ inscribeConfirmItem.affixCount || 0 }} / {{ inscribeConfirmItem.affixLimit || 0 }}')
        expect(pageSource).toContain('确认铭刻')
        expect(stateSource).toContain('const inscriptionStones = ref(0)')
        expect(stateSource).toContain("fetch(`/api/equipment/${encodeURIComponent(instanceId)}/inscribe`")
        expect(stateSource).toContain("inscriptionStones.value = Number(payload.inscriptionStones ?? 0)")
    })

    it('铭刻词条展示会补齐效果类型说明而不是只显示数值', () => {
        expect(stateSource).toContain("const affixEffectLabels = {")
        expect(stateSource).toContain("magic_proc_rate_bonus: '魔法触发率'")
        expect(stateSource).toContain("armor_pen_percent: '额外护甲穿透'")
        expect(stateSource).toContain("function formatItemAffixLine(affix) {")
        expect(pageSource).toContain('{{ formatItemAffixLine(affix) }}')
        expect(pageSource).not.toContain('{{ affix.name }} {{ affix.valueText }}')
    })

    it('资源条把金币强化石刻印石和天赋点字样替换为图标', () => {
        expect(pageSource).toContain('class="armory-backpack-resources__item"')
        expect(pageSource).toContain('const resourceIcons = {')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E9%87%91%E5%B8%81.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%BC%BA%E5%8C%96%E7%9F%B3.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%88%BB%E5%8D%B0%E7%9F%B3.png')
        expect(pageSource).toContain('https://hai-world2.oss-cn-beijing.aliyuncs.com/resource/%E5%A4%A9%E8%B5%8B%E7%82%B9.png')
        expect(pageSource).toContain('class="armory-backpack-resources__icon"')
        expect(pageSource).not.toContain('金币 <span class="num-gold">{{ gold }}</span>')
        expect(pageSource).not.toContain('强化石 <span class="num-stone">{{ stones }}</span>')
        expect(pageSource).not.toContain('刻印石 <span class="num-stone">{{ inscriptionStones }}</span>')
        expect(pageSource).not.toContain('天赋点 <span class="num-stone">{{ talentPoints }}</span>')
    })
})
