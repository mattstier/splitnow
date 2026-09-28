import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useChatStore } from '../src/stores/ChatStore.js'

// mock the service at the module level — vitest replaces it before import
vi.mock('../src/services/ChatService.js', () => ({
  chatService: {
    fetchMessages: vi.fn(),
    fetchPage: vi.fn(),
    connect: vi.fn(),
    disconnect: vi.fn(),
    send: vi.fn(),
    deleteMessage: vi.fn(),
  }
}))

// mock the auth store so the username is deterministic without touching localStorage
vi.mock('../src/stores/AuthStore.js', () => ({
  useAuthStore: {
    getState: () => ({ username: 'matestier' })
  }
}))

import { chatService } from '../src/services/ChatService.js'

beforeEach(() => {
  // reset store between tests
  useChatStore.setState({
    messages: [],
    nextLink: null,
    input: '',
    lastError: null,
    sendTimeout: null,
    lastPendingMessage: null,
    isConnected: true,
  })
  vi.clearAllMocks()
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('setInput', () => {
  it('updates the input field', () => {
    useChatStore.getState().setInput('hello')
    expect(useChatStore.getState().input).toBe('hello')
  })
})

describe('clearChat', () => {
  it('cancels the pending send timer and resets the chat state', () => {
    const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout')
    useChatStore.setState({ messages: [{ id: 1 }], nextLink: '/x', input: 'abc', sendTimeout: 123 })

    useChatStore.getState().clearChat()

    expect(clearTimeoutSpy).toHaveBeenCalledWith(123)
    const s = useChatStore.getState()
    expect(s.messages).toEqual([])
    expect(s.nextLink).toBeNull()
    expect(s.input).toBe('')
    expect(s.sendTimeout).toBeNull()
  })
})

describe('_toConfirmedMessage', () => {
  it('flags a message sent by the current user as mine', () => {
    const msg = useChatStore.getState()
      ._toConfirmedMessage({ id: 1, content: 'hi', sender: 'matestier' })

    expect(msg).toEqual({
      id: 1, content: 'hi', text: 'hi', sender: 'matestier',
      status: 'received', mine: true
    })
  })

  it('flags messages from other users as not mine', () => {
    const msg = useChatStore.getState()
      ._toConfirmedMessage({ id: 2, content: 'yo', sender: 'alice' })

    expect(msg.text).toBe('yo')
    expect(msg.status).toBe('received')
    expect(msg.mine).toBe(false)
  })
})

describe('mergeMessage', () => {
  it('appends a message that has no pending counterpart', () => {
    useChatStore.getState().mergeMessage({ id: 7, content: 'hello there', sender: 'alice' })

    const msg = useChatStore.getState().messages[0]
    expect(msg.text).toBe('hello there')
    expect(msg.status).toBe('received')
    expect(msg.mine).toBe(false)
  })

  it('replaces the pending message when the echo arrives', () => {
    useChatStore.setState({
      messages: [{ text: 'hi', sender: 'You', status: 'pending', client_msg_id: 'abc' }],
      sendTimeout: 42,
      lastError: 'Sending message timed out',
    })

    useChatStore.getState()
      .mergeMessage({ id: 9, content: 'hi', sender: 'matestier', client_msg_id: 'abc' })

    const s = useChatStore.getState()
    expect(s.messages).toHaveLength(1)
    expect(s.messages[0]).toMatchObject({ id: 9, text: 'hi', status: 'received', mine: true })
    expect(s.lastError).toBeNull()
    expect(s.sendTimeout).toBeNull()
  })
})

describe('fetchMessages', () => {
  it('maps and orders fetched messages oldest-first', async () => {
    chatService.fetchMessages.mockResolvedValue({
      data: {
        data: [
          { id: 2, content: 'newest', sender: 'matestier' },
          { id: 1, content: 'oldest', sender: 'alice' },
        ],
        links: { next: '/messages?page=2' }
      }
    })

    await useChatStore.getState().fetchMessages(1)

    const s = useChatStore.getState()
    expect(chatService.fetchMessages).toHaveBeenCalledWith(1)
    expect(s.messages).toEqual([
      { id: 1, content: 'oldest', text: 'oldest', sender: 'alice', status: 'received', mine: false },
      { id: 2, content: 'newest', text: 'newest', sender: 'matestier', status: 'received', mine: true },
    ])
    expect(s.nextLink).toBe('/messages?page=2')
    expect(s.lastError).toBeNull()
  })

  it('clears messages when the API fails', async () => {
    chatService.fetchMessages.mockRejectedValue(new Error('Network Error'))
    useChatStore.setState({ messages: [{ id: 1 }] })

    await useChatStore.getState().fetchMessages(1)

    // fetchMessages does not return its promise, so wait for the catch handler
    await vi.waitFor(() => {
      expect(useChatStore.getState().messages).toEqual([])
    })
  })
})

describe('loadOlder', () => {
  it('does nothing without a next link', async () => {
    await useChatStore.getState().loadOlder()

    expect(chatService.fetchPage).not.toHaveBeenCalled()
  })

  it('prepends older messages', async () => {
    useChatStore.setState({
      messages: [{ id: 2, text: 'existing' }],
      nextLink: '/messages?page=2'
    })
    chatService.fetchPage.mockResolvedValue({
      data: {
        data: [{ id: 1, content: 'older', sender: 'alice' }],
        links: { next: null }
      }
    })

    await useChatStore.getState().loadOlder()

    const s = useChatStore.getState()
    expect(chatService.fetchPage).toHaveBeenCalledWith('/messages?page=2')
    expect(s.messages.map(m => m.id)).toEqual([1, 2])
    expect(s.nextLink).toBeNull()
  })
})

describe('whenConnected', () => {
  it('sets an offline error and skips the action when disconnected', () => {
    useChatStore.setState({ isConnected: false })
    const action = vi.fn()

    useChatStore.getState().whenConnected(action)

    expect(action).not.toHaveBeenCalled()
    expect(useChatStore.getState().lastError).toBe('You are offline')
  })

  it('runs the action when connected', () => {
    const action = vi.fn()

    useChatStore.getState().whenConnected(action)

    expect(action).toHaveBeenCalled()
  })
})

describe('send', () => {
  it('does not send when offline', () => {
    useChatStore.setState({ isConnected: false })

    useChatStore.getState().send(1, 'hello')

    expect(chatService.send).not.toHaveBeenCalled()
    expect(useChatStore.getState().messages).toEqual([])
  })

  it('optimistically adds the message and clears the input', () => {
    useChatStore.getState().send(1, 'hello')

    const s = useChatStore.getState()
    expect(chatService.send).toHaveBeenCalledWith(1, 'hello', expect.any(String))
    expect(s.messages).toHaveLength(1)
    expect(s.messages[0]).toMatchObject({ text: 'hello', status: 'pending', mine: true, sender: 'You' })
    expect(s.input).toBe('')
    expect(s.lastPendingMessage).toBe('hello')
  })

  it('marks the message as failed when it times out', () => {
    vi.useFakeTimers()

    useChatStore.getState().send(1, 'hello')

    vi.advanceTimersByTime(4000)

    const s = useChatStore.getState()
    expect(s.lastError).toBe('Sending message timed out')
    expect(s.messages[0].status).toBe('failed')
    expect(s.input).toBe('hello')
  })
})

describe('deleteMessage', () => {
  it('sends a delete request when connected', () => {
    useChatStore.getState().deleteMessage(1, 5)

    expect(chatService.deleteMessage).toHaveBeenCalledWith(1, 5)
  })

  it('does not send when offline', () => {
    useChatStore.setState({ isConnected: false })

    useChatStore.getState().deleteMessage(1, 5)

    expect(chatService.deleteMessage).not.toHaveBeenCalled()
    expect(useChatStore.getState().lastError).toBe('You are offline')
  })
})

describe('connect', () => {
  let onMessage
  let onStatusChange

  beforeEach(() => {
    chatService.connect.mockImplementation((token, roomID, msgCb, statusCb) => {
      onMessage = msgCb
      onStatusChange = statusCb
    })
  })

  it('wires the status callback', () => {
    useChatStore.getState().connect('token', 1)

    onStatusChange(false)

    expect(useChatStore.getState().isConnected).toBe(false)
  })

  it('marks incoming deleted messages', () => {
    useChatStore.setState({ messages: [{ id: 5, text: 'remove me' }] })

    useChatStore.getState().connect('token', 1)
    onMessage({ deleted: true, id: 5 })

    const msg = useChatStore.getState().messages[0]
    expect(msg.deleted).toBe(true)
    expect(msg.text).toBe('')
  })

  it('ignores messages without content (e.g. pings)', () => {
    useChatStore.getState().connect('token', 1)
    onMessage({ type: 'ping' })

    expect(useChatStore.getState().messages).toEqual([])
  })

  it('merges regular incoming messages', () => {
    useChatStore.getState().connect('token', 1)
    onMessage({ id: 7, content: 'hi', sender: 'alice' })

    const msg = useChatStore.getState().messages[0]
    expect(msg).toMatchObject({ id: 7, text: 'hi', status: 'received', mine: false })
  })
})

describe('disconnect', () => {
  it('delegates to the chat service', () => {
    useChatStore.getState().disconnect()

    expect(chatService.disconnect).toHaveBeenCalled()
  })
})
