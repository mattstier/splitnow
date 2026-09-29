import api from '../Api.js'

let ws = null
let retryTimer = null

// WebSocket retry related constants
const wsRetrySetup = {
  maxAttemptCount: 5,
  baseIntervalMillis: 1000,
  maxIntervalMillis: 30000
}

// helper function computing the next retry interval
// using exponential backoff
export function retryInterval(attempt, setup) {
  return Math.min(
    setup.baseIntervalMillis * Math.pow(2, attempt),
    setup.maxIntervalMillis)
}

export const chatService = {
  fetchMessages: (roomID) => api.get(`/messages?room=${roomID}`),
  fetchPage: (url) => api.get(url),

  connect: (token, roomID, onMessage, onStatusChange) => {
    let retryAttempt = 0

    const open = () => {
      let webSocket = new WebSocket(`ws://${window.location.host}/ws?token=${token}`)
      webSocket.onopen = () => {
        retryAttempt = 0 // reset the backoff on successful connection
        onStatusChange(true)
        webSocket.send(JSON.stringify({ type: 'subscribe', room: roomID }))
      }
      webSocket.onmessage = (e) => onMessage(JSON.parse(e.data))
      webSocket.onclose = () => {
        if (webSocket !== ws) return
        onStatusChange(false)
        // give up after maxAttempts
        if (retryAttempt >= wsRetrySetup.maxAttemptCount) return
        // use the backoff delay
        retryTimer = setTimeout(open, retryInterval(retryAttempt, wsRetrySetup))
        retryAttempt++
      }
      webSocket.onerror = () => webSocket.close()
      ws = webSocket
    }
    open()
  },

  disconnect: () => {
    clearTimeout(retryTimer)
    retryTimer = null
    if (ws) { ws.close(); ws = null }
  },

  send: (roomID, content, replyTo = null, clientMsgID) => ws.send(JSON.stringify({ type: 'send', room: roomID, content, client_msg_id: clientMsgID, reply_to: replyTo })),
  deleteMessage: (roomID, messageID) => ws.send(JSON.stringify({ type: 'delete', room: roomID, message: messageID })),
}
