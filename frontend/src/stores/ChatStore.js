import { create } from 'zustand'
import { chatService } from '../services/ChatService.js'
import { useAuthStore } from './AuthStore.js'
import { defaultTimeout } from '../utils/Time.js'

export const useChatStore = create((set, get) => ({
  messages: [],
  // URL for the next page of older messages
  nextLink: null,
  input: '',
  lastError: null,
  sendTimeout: null,
  lastPendingMessage: null,
  // we assume that ws connection succeeded, until proven otherwise
  // to optimistically (not) render the connection error banner
  isConnected: true,

  setInput: (value) => set({ input: value }),

  clearChat: () => {
    // clear timeout timer
    clearTimeout(get().sendTimeout)
    set({ messages: [], nextLink: null, input: '', sendTimeout: null })
  },

  mergeMessage: (msg) => {
    set(s => {
      const pendingIdx = s.messages.findIndex(m => m.client_msg_id === msg.client_msg_id)
      const confirmedMsg = get()._toConfirmedMessage(msg)
      if (pendingIdx !== -1) {
        // reset timeout timer on actual merge
        clearTimeout(get().sendTimeout)
        set({ sendTimeout: null })

        const next = [...s.messages]
        next[pendingIdx] = confirmedMsg
        return { messages: next, lastError: null } // clear error
      }

      // if its not ours, append it
      return { messages: [...s.messages, confirmedMsg] }
    })
  },

  // map a single raw API message into our format with a 'mine' flag
  // and received status
  _toConfirmedMessage: (rawMessage) => {
    const myUsername = useAuthStore.getState().username
    return {
      ...rawMessage,
      text: rawMessage.content,
      status: "received",
      mine: rawMessage.sender === myUsername
    }
  },

  // maps all newly fetched messages to our frontend message format
  // and reverses the API response to display correctly
  _mapFetchedMessages: (rawMessages) => {
    return rawMessages
      .map(m => (get()._toConfirmedMessage(m)))
      .slice()
      .reverse() // API returns newest first, we want oldest first
  },

  // fetch the first page of messages when entering a room
  fetchMessages: (roomID) => {
    chatService.fetchMessages(roomID)
      .then(res => {
        set({
          messages:
            get()._mapFetchedMessages(res.data.data),
          nextLink: res.data.links.next || null,
          lastError: null
        })
      })
      .catch(() => set({ messages: [], lastError: 'Failed to fetch messages' }))
  },

  onStatusChange: (connected) => set({ isConnected: connected }),

  // open a WebSocket connection to receive live messages
  connect: (token, roomID) => {
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
      get().mergeMessage(msg)
    }, get().onStatusChange)
  },

  disconnect: () => chatService.disconnect(),

  // guard for socket-dependent actions (refuse them when offline)
  whenConnected: (action) => {
    if (!get().isConnected) {
      set({ lastError: 'You are offline' })
      return
    }
    action()
  },

  // send a message over WebSocket and optimistically add it to the list
  send: (roomID, content) => get().whenConnected(() => {
    // idempotency key for the pending messages
    // which is a temporary (probabilistically) unique ID of the client used when
    // reconciling optimistic messages with the server
    const clientMsgID = crypto.randomUUID();

    // set pending message as current
    set({ lastPendingMessage: content })

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

    const timer = setTimeout(() => {
      // reset timer before use
      clearTimeout(timer)
      const last = get().lastPendingMessage
      set(s => ({
        lastError: "Failed to send message",
        // mark timed out messages as failed
        messages: s.messages.map(m => m.client_msg_id === clientMsgID ? { ...m, status: 'failed' } : m),
        input: last ?? s.input // paste back message if it timed out
      }))
    }, defaultTimeout)
    // keep the timer handle so the echo confirmation can cancel it
    set({ sendTimeout: timer })
  }),


  // send a delete request over WebSocket
  deleteMessage: (roomID, messageID) => get().whenConnected(() =>
    chatService.deleteMessage(roomID, messageID)
  ),

  // load the next page of older messages + prepend them
  loadOlder: () => {
    const { nextLink } = get()
    if (!nextLink) return
    chatService.fetchPage(nextLink).then(res => {
      const page = res.data
      set(s => ({
        // prepend older messages in front of existing ones
        messages: [...get()._mapFetchedMessages(page.data), ...s.messages],
        nextLink: page.links.next || null
      }))
    }).catch(() => set({ lastError: 'Failed to load messages' }))
  },
}))
