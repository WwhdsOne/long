import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

function resolveDevApiTarget() {
    const explicitTarget = process.env.LONG_DEV_API_TARGET?.trim()
    if (explicitTarget) {
        return explicitTarget
    }

    const host = process.env.LONG_LISTEN_HOST?.trim() || '127.0.0.1'
    const port = process.env.LONG_LISTEN_PORT?.trim() || '2333'
    return `http://${host}:${port}`
}

export const devApiTarget = resolveDevApiTarget()

// https://vite.dev/config/
export default defineConfig({
    plugins: [vue()],
    build: {
        outDir: '../backend/public',
        emptyOutDir: true,
    },
    server: {
        host: '0.0.0.0',
        port: 5173,
        proxy: {
            '/api': {
                target: devApiTarget,
                changeOrigin: true,
                ws: true,
            },
        },
    },
})
