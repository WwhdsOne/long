import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const publicPageSource = readFileSync(path.resolve(currentDir, './PublicPage.vue'), 'utf8')
const styleSource = readFileSync(path.resolve(currentDir, '../style.css'), 'utf8')
const siteFooterLinkSection = styleSource.slice(
  styleSource.indexOf('.site-footer__link {'),
  styleSource.indexOf('.site-footer__link:hover {')
)

describe('PublicPage 备案页脚', () => {
  it('在公共站点底部展示备案号', () => {
    expect(publicPageSource).toContain('aria-label="网站备案信息"')
    expect(publicPageSource).toContain('京ICP备2025120689号-2')
    expect(publicPageSource).toContain('https://beian.miit.gov.cn/')
  })

  it('提供独立页脚样式', () => {
    expect(styleSource).toContain('.site-footer {')
    expect(styleSource).toContain('.site-footer__link {')
    expect(siteFooterLinkSection).toContain('color: var(--primary-deep);')
    expect(siteFooterLinkSection).toContain('font-weight: 800;')
    expect(siteFooterLinkSection).toContain('background: rgba(255, 255, 255, 0.78);')
  })
})
