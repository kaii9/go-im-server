let ws = null
let reconnectTimer = null
let handlers = []

export function connect(token) {
  if (ws && ws.readyState === WebSocket.OPEN) return

  ws = new WebSocket(`ws://${location.host}/ws?token=${token}`)

  ws.onopen = () => {
    console.log('WS connected')
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      handlers.forEach((fn) => fn(msg))
    } catch (e) {
      console.error('WS parse error', e)
    }
  }

  ws.onclose = () => {
    console.log('WS disconnected, reconnecting in 5s...')
    reconnectTimer = setTimeout(() => {
      const token = localStorage.getItem('token')
      if (token) connect(token)
    }, 5000)
  }

  ws.onerror = () => {
    ws.close()
  }
}

export function disconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.close()
    ws = null
  }
}

export function send(data) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(data))
  }
}

export function onMessage(handler) {
  handlers.push(handler)
  return () => {
    handlers = handlers.filter((h) => h !== handler)
  }
}
