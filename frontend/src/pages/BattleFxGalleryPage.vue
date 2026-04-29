<script setup>
import PixelEffectCanvas from '../components/PixelEffectCanvas.vue'

const smallEffects = [
  { key: 'storm_combo',      name: '风暴连击',   skill: '风暴追击',    trigger: 'storm_combo',      layer: 'L3',       tag: '斩击',     tagClass: 'tag--soft',    note: '绿色像素斩击沿左下→右上快速展开，末端渐隐。短促清晰可叠加。' },
  { key: 'auto_strike',       name: '自动重击',   skill: '碎甲重击',    trigger: 'auto_strike',      layer: 'L3 + L1',  tag: '打击',     tagClass: 'tag--heavy',   note: 'T形锤从左侧边缘立起，绕支点旋转90°砸向中心，命中产生碎片。' },
  { key: 'bleed',             name: '流血',        skill: '致命出血',    trigger: 'bleed',            layer: 'L1',       tag: '残留',     tagClass: 'tag--weak',    note: '暗红像素向周围微扩散后定格为血迹群，不漂浮不下落。' },
  { key: 'collapse_trigger',  name: '护甲崩塌',   skill: '护甲崩塌',    trigger: 'collapse_trigger', layer: 'L2+L1+L3', tag: '崩塌',     tagClass: 'tag--best',    note: '方案9像素盾牌爆裂四散：盾牌→每个像素从中心向外爆开→全部消散后重置。' },
  { key: 'doom_mark',         name: '末日标记',   skill: '末日审判',    trigger: 'doom_mark',        layer: 'L2',       tag: '标记',     tagClass: 'tag--heavy',   note: '暗红断裂像素环逐段出现后轻微收缩定格，不闭合不成光滑圆。末日审判天赋触发。' },
  { key: 'silver_storm',      name: '银色风暴',   skill: '白银风暴',    trigger: 'silver_storm',     layer: 'L3',       tag: '连斩',     tagClass: 'tag--soft',    note: '从顶部到底部快速刷出更多、更粗的随机银色刀光，落底后短暂残留并伴随轻微闪白。' },
]

const ultimateSkills = [
  {
    key: 'final_cut',
    name: '死亡狂喜',
    skill: '死亡狂喜',
    trigger: 'death_ecstasy_ult',
    tag: '终结',
    tagClass: 'tag--heavy',
    note: '终极技能演示区按真实战斗页 5x5 Boss 网格 1:1 复刻。下层是同尺寸格子，上层叠加终结斩特效；内部仍使用低分辨率像素画布，再按 5x5 战斗区等比放大。',
  },
  {
    key: 'judgment_day',
    name: '审判日',
    skill: '审判之日',
    trigger: 'judgment_day',
    size: 256,
    tag: '裁决',
    tagClass: 'tag--best',
    note: '像素十字裁决：横 + 竖两道黄金斩击从中心同时向四端展开，覆盖 5x5 中心行与中心列，中心交叉点最亮。2.2s 后十字整体破碎为金色粒子四散消失。',
  },
  {
    key: 'silver_storm',
    name: '白银风暴',
    skill: '白银风暴',
    trigger: 'silver_storm',
    tag: '连斩',
    tagClass: 'tag--soft',
    note: '同样使用真实战斗页 5x5 终极技能画布。覆盖层直接铺满整块网格，用来观察白银风暴在实战尺寸下的刀光密度、宽度和留场效果。',
  },
]

const ultimateRows = [
  [
    { type: 'heavy', name: '左肩甲', hp: 82 },
    { type: 'soft', name: '锁骨', hp: 76 },
    { type: 'heavy', name: '头甲', hp: 88 },
    { type: 'soft', name: '右肩', hp: 71 },
    { type: 'heavy', name: '臂盾', hp: 67 },
  ],
  [
    { type: 'soft', name: '左臂', hp: 74 },
    { type: 'weak', name: '左肺', hp: 45 },
    { type: 'soft', name: '咽喉', hp: 63 },
    { type: 'weak', name: '右肺', hp: 41 },
    { type: 'soft', name: '右臂', hp: 72 },
  ],
  [
    { type: 'heavy', name: '左肋甲', hp: 69 },
    { type: 'soft', name: '胸腔', hp: 58 },
    { type: 'weak', name: '胸甲核心', hp: 33, center: true },
    { type: 'soft', name: '心室', hp: 52 },
    { type: 'heavy', name: '右肋甲', hp: 64 },
  ],
  [
    { type: 'soft', name: '左腹', hp: 57 },
    { type: 'heavy', name: '盆骨甲', hp: 61 },
    { type: 'soft', name: '腹腔', hp: 49 },
    { type: 'heavy', name: '髋甲', hp: 65 },
    { type: 'soft', name: '右腹', hp: 55 },
  ],
  [
    { type: 'heavy', name: '左腿甲', hp: 73 },
    { type: 'soft', name: '左膝', hp: 62 },
    { type: 'heavy', name: '脊柱甲', hp: 79 },
    { type: 'soft', name: '右膝', hp: 59 },
    { type: 'heavy', name: '右腿甲', hp: 75 },
  ],
]
</script>

<template>
  <main class="page-shell battle-fx-gallery-page">
    <section class="bfxg">
      <header class="bfxg__header">
        <p class="vote-stage__eyebrow">内部演示页</p>
        <h1>像素特效方案图鉴墙</h1>
        <p class="bfxg__copy">
          低分辨率 Canvas 像素粒子动画方案评审。全部手写像素粒子，不接 OSS，不连后端，不触发真实战斗。
        </p>
      </header>

      <!-- 单特效图鉴墙 -->
      <section class="bfxg__section">
        <div class="bfxg__section-head">
          <h2>小技能特效</h2>
          <p>上半区只放单格触发的小技能预览，保持低分辨率像素粒子循环播放。</p>
        </div>
        <div class="bfxg__gallery">
          <article v-for="fx in smallEffects" :key="fx.key" class="bfxg__card">
            <div class="bfxg__card-preview">
              <div class="bfxg__canvas-wrap">
                <PixelEffectCanvas :effect="fx.key" :size="90" :loop="true" />
              </div>
            </div>
            <div class="bfxg__card-body">
              <div class="bfxg__card-head">
                <strong>{{ fx.name }}</strong>
                <code>{{ fx.trigger }}</code>
              </div>
              <div class="bfxg__skill-line">
                对应技能：<span>{{ fx.skill }}</span>
              </div>
              <div class="bfxg__card-meta">
                <span class="bfxg__layer">{{ fx.layer }}</span>
                <span class="bfxg__tag" :class="fx.tagClass">{{ fx.tag }}</span>
              </div>
              <p class="bfxg__card-note">{{ fx.note }}</p>
            </div>
          </article>
        </div>
      </section>

      <!-- 并发组合预览 -->
      <section class="bfxg__section">
        <div class="bfxg__section-head">
          <h2>5x5 终极技能演示</h2>
          <p>下半区固定复刻真实战斗页的 5x5 Boss 网格尺寸。终极技能画布为 5x5，并以同尺寸覆盖层叠加在格子上。</p>
        </div>
        <div class="bfxg__ultimate-grid-list">
          <article v-for="ultimate in ultimateSkills" :key="ultimate.key" class="bfxg__ultimate-card">
            <div class="bfxg__ultimate-preview">
              <div class="boss-part-grid bfxg__ultimate-grid">
                <div v-for="(row, yi) in ultimateRows" :key="yi" class="boss-part-grid__row">
                  <button
                    v-for="(zone, xi) in row"
                    :key="`${yi}-${xi}`"
                    class="boss-part-cell boss-zone-button"
                    :class="{
                      'boss-part-cell--alive': true,
                      'boss-part-cell--soft': zone.type === 'soft',
                      'boss-part-cell--heavy': zone.type === 'heavy',
                      'boss-part-cell--weak': zone.type === 'weak',
                      'boss-part-cell--center': !!zone.center,
                    }"
                    :style="{ '--part-color': zone.type === 'weak' ? '#ef4444' : zone.type === 'heavy' ? '#9ca3af' : '#4ade80' }"
                    type="button"
                    disabled
                  >
                    <div class="boss-part-cell__type">{{ zone.type === 'weak' ? '弱点' : zone.type === 'heavy' ? '重甲' : '软组织' }}</div>
                    <strong class="boss-zone-button__label">{{ zone.name }}</strong>
                    <div class="boss-part-cell__bar">
                      <span class="boss-part-cell__fill" :style="{ width: `${zone.hp}%` }"></span>
                    </div>
                    <div class="boss-zone-button__meta">
                      <span>血量 : {{ zone.hp }}%</span><br>
                      <span>护甲 : {{ zone.type === 'heavy' ? 320 : 0 }}</span>
                    </div>
                  </button>
                </div>
                <div class="bfxg__ultimate-overlay" aria-hidden="true">
                  <PixelEffectCanvas :effect="ultimate.key" :size="ultimate.size || 90" :loop="true" />
                </div>
              </div>
            </div>
            <div class="bfxg__card-body">
              <div class="bfxg__card-head">
                <strong>{{ ultimate.name }}</strong>
                <code>{{ ultimate.trigger }}</code>
              </div>
              <div class="bfxg__skill-line">
                对应技能：<span>{{ ultimate.skill }}</span>
              </div>
              <div class="bfxg__card-meta">
                <span class="bfxg__layer">L3 + 5x5 终极画布</span>
                <span class="bfxg__tag" :class="ultimate.tagClass">{{ ultimate.tag }}</span>
              </div>
              <p class="bfxg__card-note">{{ ultimate.note }}</p>
            </div>
          </article>
        </div>
      </section>
    </section>
  </main>
</template>
