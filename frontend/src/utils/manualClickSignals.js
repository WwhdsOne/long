export async function sha256Hex(value) {
    const subtle = globalThis.crypto?.subtle ?? null
    if (!subtle) {
        throw new Error('当前环境不支持指纹摘要')
    }
    const digest = await subtle.digest('SHA-256', new TextEncoder().encode(String(value ?? '')))
    return Array.from(new Uint8Array(digest))
        .map((item) => item.toString(16).padStart(2, '0'))
        .join('')
}
