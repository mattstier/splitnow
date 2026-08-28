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

  // map raw API messages into our format with a 'mine' flag
  _mapMessages: (rawMessages) => {
    const myUsername = useAuthStore.getState().username
    return rawMessages.map(m => ({
      id: m.id,
      deleted: m.deleted,
      text: m.content,
      sender: m.sender,
      created_at: m.created_at,
      mine: m.sender === myUsername
    }))
  },

  // fetch the first page of messages when entering a room
  fetchMessages: (roomID) => {
    chatService.fetchMessages(roomID)
      .then(res => {
        set({
          messages: get()._mapMessages(res.data.data).slice().reverse(), // API returns newest first, we want oldest first
          nextLink: res.data.links.next || null
        })
      })
      .catch(() => set({ messages: [] }))
  },

  // open a WebSocket connection to receive live messages
  connect: (token, roomID) => {
    const myUsername = useAuthStore.getState().username
    chatService.connect(token, roomID, (d) => {
      if (d.deleted) {
        set(s => ({
          messages: s.messages.map(m =>
            m.id === d.id ? { ...m, deleted: true, text: '' } : m
          )
        }))
        return
      }
      // skip messages without content (e.g.: pings)
      if (d.content === undefined) return
      set(s => ({
        messages: [...s.messages, {
          id: d.id,
          deleted: false,
          text: d.content,
          sender: d.sender,
          created_at: d.created_at,
          mine: d.sender === myUsername // mine flag to display own messages differently
        }]
      }))
    })
  },

  disconnect: () => chatService.disconnect(),

  // send a message over WebSocket and optimistically add it to the list
  send: (roomID, content) => {
    chatService.send(roomID, content)
    set(s => ({
      messages: [...s.messages, {
        text: content,
        sender: 'You',
        created_at: new Date().toISOString(),
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
