import {describe, expect, it} from 'vitest'

import {buildOSSObjectKey, buildPublicAssetURL, sanitizeUploadFileName} from './ossUpload'

describe('ossUpload', () => {
    it('会清洗文件名里的空格和特殊字符，并保留扩展名', () => {
        expect(sanitizeUploadFileName('小小英雄 头像(最终版).webp')).toBe('file.webp')
        expect(sanitizeUploadFileName('spark cat@2x.webp')).toBe('spark-cat-2x.webp')
    })

    it('会基于目录和时间戳拼出 OSS 对象 key', () => {
        expect(buildOSSObjectKey('uploads/heroes/', 'spark cat.webp', 123456)).toBe('uploads/heroes/123456-spark-cat.webp')
        expect(buildOSSObjectKey('uploads/heroes', 'boss.png', 42)).toBe('uploads/heroes/42-boss.png')
    })

    it('会把 publicBaseUrl 和对象 key 组装成最终公开地址', () => {
        expect(buildPublicAssetURL('https://cdn.example.com/', 'uploads/heroes/1-a.webp')).toBe('https://cdn.example.com/uploads/heroes/1-a.webp')
        expect(buildPublicAssetURL('https://cdn.example.com', 'uploads/heroes/1-a.webp')).toBe('https://cdn.example.com/uploads/heroes/1-a.webp')
    })
})
