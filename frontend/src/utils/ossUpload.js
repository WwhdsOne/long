export function sanitizeUploadFileName(name) {
    const rawName = String(name || '').trim()
    const lastDotIndex = rawName.lastIndexOf('.')
    const hasExtension = lastDotIndex > 0 && lastDotIndex < rawName.length - 1
    const rawBase = hasExtension ? rawName.slice(0, lastDotIndex) : rawName
    const rawExtension = hasExtension ? rawName.slice(lastDotIndex + 1) : ''

    const base = sanitizeSegment(rawBase) || 'file'
    const extension = sanitizeSegment(rawExtension).toLowerCase()

    return extension ? `${base}.${extension}` : base
}

export function buildOSSObjectKey(dir, fileName, timestamp = Date.now()) {
    const prefix = String(dir || '').replace(/\/+$/, '')
    const sanitizedName = sanitizeUploadFileName(fileName)
    return prefix ? `${prefix}/${timestamp}-${sanitizedName}` : `${timestamp}-${sanitizedName}`
}

export function buildPublicAssetURL(publicBaseUrl, objectKey) {
    return `${String(publicBaseUrl || '').replace(/\/$/, '')}/${String(objectKey || '').replace(/^\/+/, '')}`
}

export function delay(ms) {
    return new Promise((resolve) => {
        window.setTimeout(resolve, ms)
    })
}

export function probeImage(url) {
    return new Promise((resolve) => {
        const image = new Image()
        image.onload = () => resolve(true)
        image.onerror = () => resolve(false)
        image.src = `${url}${url.includes('?') ? '&' : '?'}t=${Date.now()}`
    })
}

export async function waitForPublicImage(url, attempts = 6, interval = 800) {
    for (let attempt = 0; attempt < attempts; attempt += 1) {
        if (await probeImage(url)) {
            return true
        }

        if (attempt < attempts - 1) {
            await delay(interval)
        }
    }

    return false
}

export async function uploadImageWithPolicy(file, policy, options = {}) {
    const fetchImpl = options.fetchImpl ?? fetch
    const verifyPublicImage = options.verifyPublicImage ?? waitForPublicImage
    const now = options.now ?? Date.now

    const objectKey = buildOSSObjectKey(policy?.dir, file.name, now())
    const finalURL = buildPublicAssetURL(policy?.publicBaseUrl, objectKey)
    const formData = new FormData()
    formData.append('key', objectKey)
    formData.append('policy', policy.policy)
    formData.append('OSSAccessKeyId', policy.accessKeyId)
    formData.append('Signature', policy.signature)
    formData.append('success_action_status', '200')
    formData.append('file', file)

    try {
        const uploadResponse = await fetchImpl(policy.host, {
            method: 'POST',
            body: formData,
        })
        if (!uploadResponse.ok) {
            throw new Error('上传到 OSS 失败，请检查桶权限和上传策略。')
        }
    } catch (error) {
        const reachable = await verifyPublicImage(finalURL)
        if (!reachable) {
            throw new Error('图片可能已经传到 OSS，但浏览器无法确认上传结果。请给 OSS 配置 CORS，或检查 public_base_url 是否可公开访问。')
        }
    }

    return finalURL
}

function sanitizeSegment(value) {
    return String(value || '')
        .replace(/[^a-zA-Z0-9_-]+/g, '-')
        .replace(/-+/g, '-')
        .replace(/^-+|-+$/g, '')
}
