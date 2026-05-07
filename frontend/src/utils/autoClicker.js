export const AUTO_CLICK_INTERVAL_MS = 1000 / 3

export function createAutoClickLoop({
                                        onTick,
                                        intervalMs = AUTO_CLICK_INTERVAL_MS,
                                        setTimeoutFn = globalThis.setTimeout,
                                        clearTimeoutFn = globalThis.clearTimeout,
                                    } = {}) {
    let timerId = null
    let running = false

    const scheduleNext = () => {
        if (!running) {
            return
        }

        timerId = setTimeoutFn(() => {
            if (!running) {
                return
            }

            onTick?.()
            scheduleNext()
        }, intervalMs)
    }

    return {
        start() {
            if (running) {
                return
            }

            running = true
            scheduleNext()
        },
        stop() {
            running = false
            if (timerId !== null) {
                clearTimeoutFn(timerId)
                timerId = null
            }
        },
        isRunning() {
            return running
        },
    }
}
