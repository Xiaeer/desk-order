import { getToken } from './auth'

const RECONNECT_DELAY = 2000
const REFRESH_DEBOUNCE = 150

function buildMerchantWsUrl(token) {
  const apiOrigin = (import.meta.env.VITE_API_ORIGIN || '').trim().replace(/\/$/, '')
  const baseUrl = new URL(apiOrigin || window.location.origin)
  const protocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'

  baseUrl.protocol = protocol
  baseUrl.pathname = '/api/v1/merchant/ws'
  baseUrl.search = ''
  baseUrl.searchParams.set('token', token)

  return baseUrl.toString()
}

export function setupMerchantOrderRealtime(onOrderEvent) {
  const token = getToken().trim()
  if (!token || typeof window === 'undefined' || typeof document === 'undefined') {
    return () => {}
  }

  let disposed = false
  let reconnectTimer = null
  let refreshTimer = null
  let socket = null

  function runRefresh(message) {
    window.clearTimeout(refreshTimer)
    refreshTimer = window.setTimeout(() => {
      Promise.resolve(onOrderEvent?.(message)).catch(() => {})
    }, REFRESH_DEBOUNCE)
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  function cleanupSocket() {
    if (!socket) {
      return
    }
    socket.onopen = null
    socket.onmessage = null
    socket.onerror = null
    socket.onclose = null
    socket.close()
    socket = null
  }

  function scheduleReconnect() {
    if (disposed || reconnectTimer || document.visibilityState !== 'visible') {
      return
    }
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, RECONNECT_DELAY)
  }

  function connect() {
    if (disposed || document.visibilityState === 'hidden' || socket) {
      return
    }

    socket = new window.WebSocket(buildMerchantWsUrl(token))
    socket.onmessage = event => {
      try {
        const message = JSON.parse(event.data || '{}')
        if (message.type === 'new_order' || message.type === 'order_update') {
          runRefresh(message)
        }
      } catch {
        // ignore malformed payloads and keep the connection alive
      }
    }
    socket.onerror = () => {
      socket?.close()
    }
    socket.onclose = () => {
      socket = null
      scheduleReconnect()
    }
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      runRefresh()
      connect()
      return
    }
    clearReconnectTimer()
    cleanupSocket()
  }

  document.addEventListener('visibilitychange', handleVisibilityChange)
  connect()

  return () => {
    disposed = true
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    clearReconnectTimer()
    window.clearTimeout(refreshTimer)
    cleanupSocket()
  }
}