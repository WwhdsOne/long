function trimNickname(value) {
  return String(value ?? '').trim()
}

function buildNicknameQuery(nickname) {
  const normalized = trimNickname(nickname)
  if (!normalized) {
    return ''
  }

  return `?nickname=${encodeURIComponent(normalized)}`
}

function defaultWebSocketUrl() {
  const location = globalThis.location
  const protocol = location?.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = location?.host || 'localhost'
  return `${protocol}//${host}/api/ws`
}

function defaultEventSourceUrl(nickname) {
  return `/api/events${buildNicknameQuery(nickname)}`
}

export function createRealtimeTransport(options = {}) {
  const createWebSocket = options.createWebSocket || ((url) => new WebSocket(url))
  const createEventSource = options.createEventSource || ((url) => new EventSource(url))
  const buildWebSocketUrl = options.buildWebSocketUrl || defaultWebSocketUrl
  const buildEventSourceUrl = options.buildEventSourceUrl || defaultEventSourceUrl
  const onSnapshot = options.onSnapshot || (() => {})
  const onPublicDelta = options.onPublicDelta || (() => {})
  const onUserDelta = options.onUserDelta || (() => {})
  const onClickAck = options.onClickAck || (() => {})
  const onTransportState = options.onTransportState || (() => {})
  const onTransportError = options.onTransportError || (() => {})

  let ws = null
  let wsOpen = false
  let sse = null
  let currentNickname = ''
  let closed = false
  let reconnectTimer = 0
  let state = {
    connected: false,
    degraded: false,
    mode: 'idle',
  }

  function updateState(nextState) {
    state = { ...state, ...nextState }
    onTransportState({ ...state })
  }

  function closeWebSocket() {
    if (!ws) {
      wsOpen = false
      return
    }

    const socket = ws
    ws = null
    wsOpen = false
    socket.close?.()
  }

  function closeEventSource() {
    if (!sse) {
      return
    }

    const source = sse
    sse = null
    source.close?.()
  }

  function clearReconnectTimer() {
    if (!reconnectTimer) {
      return
    }

    globalThis.clearTimeout?.(reconnectTimer)
    reconnectTimer = 0
  }

  function scheduleWebSocketRetry() {
    if (closed || reconnectTimer) {
      return
    }

    reconnectTimer = globalThis.setTimeout(() => {
      reconnectTimer = 0
      if (closed) {
        return
      }
      connectWebSocket()
    }, 3000)
  }

  function handleSocketMessage(raw) {
    let message
    try {
      message = JSON.parse(raw)
    } catch {
      onTransportError('实时消息解析失败，请刷新页面后重试。')
      return
    }

    switch (message?.type) {
      case 'snapshot':
        clearReconnectTimer()
        onSnapshot(message.public ?? {}, message.user ?? null)
        updateState({
          connected: true,
          degraded: false,
          mode: 'ws',
        })
        return
      case 'public_delta':
        clearReconnectTimer()
        onPublicDelta(message.payload ?? {})
        updateState({
          connected: true,
          degraded: false,
          mode: 'ws',
        })
        return
      case 'user_delta':
        clearReconnectTimer()
        onUserDelta(message.payload ?? {})
        updateState({
          connected: true,
          degraded: false,
          mode: 'ws',
        })
        return
      case 'click_ack':
        clearReconnectTimer()
        onClickAck(message.payload ?? {})
        updateState({
          connected: true,
          degraded: false,
          mode: 'ws',
        })
        return
      case 'error':
        onTransportError(message.message || '实时消息处理失败，请稍后重试。')
        return
      case 'pong':
        return
      default:
        onTransportError('收到不支持的实时消息，请刷新页面后重试。')
    }
  }

  function startEventSourceFallback(message) {
    if (closed) {
      return
    }

    closeWebSocket()
    if (sse) {
      return
    }

    updateState({
      connected: false,
      degraded: true,
      mode: 'sse',
    })
    onTransportError(message || '实时主链路暂时不可用，已切回兼容模式。')
    scheduleWebSocketRetry()

    let source
    try {
      source = createEventSource(buildEventSourceUrl(currentNickname))
    } catch {
      onTransportError('实时连接初始化失败，请稍后刷新页面。')
      return
    }

    sse = source
    source.onopen = () => {
      if (sse !== source) {
        return
      }
      updateState({
        connected: true,
        degraded: true,
        mode: 'sse',
      })
    }
    source.onerror = () => {
      if (sse !== source) {
        return
      }
      updateState({
        connected: false,
        degraded: true,
        mode: 'sse',
      })
      onTransportError('实时连接暂时不可用，页面会自动重连。')
    }

    const handleNamedEvent = (applier) => (event) => {
      try {
        const payload = JSON.parse(event.data)
        applier(payload)
        updateState({
          connected: true,
          degraded: true,
          mode: 'sse',
        })
      } catch {
        onTransportError('实时消息解析失败，请刷新页面后重试。')
      }
    }

    source.addEventListener('public_state', handleNamedEvent(onPublicDelta))
    source.addEventListener('user_state', handleNamedEvent(onUserDelta))
  }

  function connectWebSocket() {
    clearReconnectTimer()
    closeEventSource()
    updateState({
      connected: false,
      degraded: false,
      mode: 'ws',
    })

    let socket
    try {
      socket = createWebSocket(buildWebSocketUrl())
    } catch {
      startEventSourceFallback('实时主链路暂时不可用，已切回兼容模式。')
      return
    }

    ws = socket
    socket.onopen = () => {
      if (ws !== socket) {
        return
      }
      wsOpen = true
      socket.send(JSON.stringify({
        type: 'hello',
        nickname: currentNickname,
      }))
    }
    socket.onmessage = (event) => {
      if (ws !== socket) {
        return
      }
      handleSocketMessage(event.data)
    }
    socket.onerror = () => {
      if (ws !== socket) {
        return
      }
      startEventSourceFallback('实时主链路暂时不可用，已切回兼容模式。')
    }
    socket.onclose = () => {
      if (ws !== socket) {
        return
      }
      startEventSourceFallback('实时主链路暂时不可用，已切回兼容模式。')
    }
  }

  return {
    connect({ nickname = '' } = {}) {
      closed = false
      currentNickname = trimNickname(nickname)
      connectWebSocket()
    },
    reconnect({ nickname = '' } = {}) {
      currentNickname = trimNickname(nickname)
      closeEventSource()
      closeWebSocket()
      connectWebSocket()
    },
    sendClick(slug, ticket, behavior = {}) {
      if (!ws || !wsOpen) {
        return false
      }

      try {
        ws.send(JSON.stringify({
          type: 'click',
          slug,
          ticket,
          ...behavior,
        }))
        return true
      } catch {
        startEventSourceFallback('实时主链路暂时不可用，已切回兼容模式。')
        return false
      }
    },
    requestSync() {
      if (!ws || !wsOpen) {
        return false
      }

      try {
        ws.send(JSON.stringify({
          type: 'sync_request',
        }))
        return true
      } catch {
        startEventSourceFallback('实时主链路暂时不可用，已切回兼容模式。')
        return false
      }
    },
    close() {
      closed = true
      clearReconnectTimer()
      closeEventSource()
      closeWebSocket()
      updateState({
        connected: false,
        degraded: false,
        mode: 'idle',
      })
    },
  }
}
