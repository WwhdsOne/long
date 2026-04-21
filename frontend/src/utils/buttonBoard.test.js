import { describe, expect, it } from 'vitest'

import { collectButtonTags, filterAndSortButtons, formatDropRate } from './buttonBoard'

describe('buttonBoard', () => {
  const buttons = [
    { key: 'feel', label: '有感觉吗', sort: 20, tags: ['日常', '聊天'] },
    { key: 'boss', label: '集火 Boss', sort: 10, tags: ['战斗'] },
    { key: 'star', label: '星光应援', sort: 30, tags: ['活动', '聊天'] },
  ]

  it('会聚合并排序全部标签', () => {
    expect(collectButtonTags(buttons)).toEqual(['活动', '聊天', '日常', '战斗'])
  })

  it('会先按标签和关键字过滤，再把星光按钮排到最前面', () => {
    expect(
      filterAndSortButtons(buttons, {
        selectedTag: '聊天',
        query: '光',
        activeStarlightKeys: ['star'],
      }),
    ).toEqual([buttons[2]])

    expect(
      filterAndSortButtons(buttons, {
        selectedTag: '全部',
        query: '',
        activeStarlightKeys: ['feel', 'star'],
      }).map((button) => button.key),
    ).toEqual(['feel', 'star', 'boss'])
  })

  it('会把概率格式化成固定百分比文本', () => {
    expect(formatDropRate(25)).toBe('25%')
    expect(formatDropRate(33.3333)).toBe('33.33%')
    expect(formatDropRate(0)).toBe('0%')
  })
})
