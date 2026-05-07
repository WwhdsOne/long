export function assetUrl(path) {
    const rawBase = import.meta.env.BASE_URL || '/'
    const base = rawBase.endsWith('/') ? rawBase : `${rawBase}/`
    const normalizedPath = String(path || '').replace(/^\/+/, '')
    return `${base}${normalizedPath}`
}

