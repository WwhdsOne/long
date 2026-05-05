import { describe, expect, it } from 'vitest'

import config, { devApiTarget } from '../vite.config.js'

describe('vite 开发代理', () => {
  it('会为 /api 显式开启 WebSocket 代理，并使用本地后端监听地址', () => {
    expect(devApiTarget).toBe('http://127.0.0.1:2333')
    expect(config.server?.proxy?.['/api']).toEqual({
      target: 'http://127.0.0.1:2333',
      changeOrigin: true,
      ws: true,
    })
  })
})
