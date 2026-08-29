import { create } from 'zustand'
import { chatService } from '../services/ChatService.js'
import { useAuthStore } from './AuthStore.js'

export const useChatStore = create((set, get) => ({
  messages: [],

  // URL for the next page of older messages
  nextLink: null,

  input: '',

  setInput: (value) => set({ input: value }),

  clearChat: () => set({ messages: [], nextLink: null, input: '' }),

  mergeMessage: (msg, myUsername) => {
    set(s => {
      const pendingIdx = s.messages.findIndex(m => m.client_msg_id === msg.client_msg_id)
      const confirmedMsg = {
        id: msg.id,
        deleted: false,
        text: msg.content,
        sender: msg.sender,
        created_at: msg.created_at,
        status: 'received',
        mine: msg.sender === myUsername
      }
      if (pendingIdx !== -1) {
        const next = [...s.messages]
        next[pendingIdx] = confirmedMsg
        return { messages: next }
      }

      // if its not ours, append it
      return { messages: [...s.messages, confirmedMsg] }
    })
  },
  // map raw API messages into our format with a 'mine' flag
  _mapMessages: (rawMessages) => {
    const myUsername = useAuthStore.getState().username
    return rawMessages.map(m => ({
      id: m.id,
      deleted: m.deleted,
      text: m.content,
      sender: m.sender,
      created_at: m.created_at,
      status: "received",
      mine: m.sender === myUsername
    }))
  },

  // fetch the first page of messages when entering a room
  fetchMessages: (roomID) => {
    chatService.fetchMessages(roomID)
      .then(res => {
        set({
          messages:
            get()
              ._mapMessages(res.data.data)
              .slice()
              .reverse(), // API returns newest first, we want oldest first
          nextLink: res.data.links.next || null
        })
      })
      .catch(() => set({ messages: [] }))
  },

  // open a WebSocket connection to receive live messages
  connect: (token, roomID) => {
    const myUsername = useAuthStore.getState().username
    chatService.connect(token, roomID, (msg) => {
      if (msg.deleted) {
        set(s => ({
          messages: s.messages.map(m =>
            m.id === msg.id ? { ...m, deleted: true, text: '' } : m
          )
        }))
        return
      }
      // skip messages without content (e.g.: pings)
      if (msg.content === undefined) return
      get().mergeMessage(msg, myUsername)
    })
  },

  disconnect: () => chatService.disconnect(),

  // send a message over WebSocket and optimistically add it to the list
  send: (roomID, content) => {
    // idempotency key for the pending messages
    // which is a temporary (probabilistically) unique ID of the client used when
    // reconciling optimistic messages with the server
    const clientMsgID = crypto.randomUUID();

    chatService.send(roomID, content, clientMsgID)
    set(s => ({
      messages: [...s.messages, {
        text: content,
        sender: 'You',
        created_at: new Date().toISOString(),
        status: 'pending',
        client_msg_id: clientMsgID,
        mine: true
      }],
      input: ''
    }))
  },

  // send a delete request over WebSocket
  deleteMessage: (roomID, messageID) => {
    chatService.deleteMessage(roomID, messageID)
  },

  // load the next page of older messages + prepend them
  loadOlder: () => {
    const { nextLink } = get()
    if (!nextLink) return
    chatService.fetchPage(nextLink).then(res => {
      const page = res.data
      set(s => ({
        // prepend older messages in front of existing ones
        messages: [...get()._mapMessages(page.data).slice().reverse(), ...s.messages],
        nextLink: page.links.next || null
      }))
    })
  },
}))
