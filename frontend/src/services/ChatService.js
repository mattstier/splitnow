import api from '../Api.js'

let ws = null

export const chatService = {
  fetchMessages: (roomID) => api.get(`/messages?room=${roomID}`),
  fetchPage: (url) => api.get(url),

  connect: (token, roomID, onMessage) => {
    ws = new WebSocket('ws://localhost:5173/ws?token=' + token)
    ws.onopen = () => ws.send(JSON.stringify({ type: 'subscribe', room: roomID }))
    ws.onmessage = (e) => onMessage(JSON.parse(e.data))
  },

  disconnect: () => { if (ws) { ws.close(); ws = null } },

  send: (roomID, content) => ws.send(JSON.stringify({ type: 'send', room: roomID, content })),
  deleteMessage: (roomID, messageID) => ws.send(JSON.stringify({ type: 'delete', room: roomID, message: messageID })),
}
