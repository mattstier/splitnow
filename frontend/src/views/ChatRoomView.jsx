import { useChatStore } from '../stores/ChatStore.js'
import { useRoomStore } from '../stores/RoomStore.js'
import { useAuthStore } from '../stores/AuthStore.js'
import { useRef, useState, useEffect } from 'react'

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// chatroom view
export default function ChatRoomView() {

  const token = useAuthStore((s) => s.token)
  const room = useRoomStore((s) => s.room)
  const messages = useChatStore((s) => s.messages)
  const nextLink = useChatStore((s) => s.nextLink)
  const input = useChatStore((s) => s.input)
  const [showBottomBtn, setShowBottomBtn] = useState(false)
  const scrollRef = useRef(null)
  const loadingOlderRef = useRef(false)

  // leaves the current room and returns to the picker
  const leaveRoom = () => {
    if (!room) return
    useChatStore.getState().clearChat()
    useRoomStore.getState().leaveRoom()
  }

  const loadOlder = () => {
    if (!nextLink || loadingOlderRef.current) return
    loadingOlderRef.current = true
    useChatStore.getState().loadOlder()
    loadingOlderRef.current = false
  }

  // shows the jump-to-bottom button when the user scrolls away from the latest message
  const scrollToBottom = () => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const onScroll = () => {
      const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 60
      setShowBottomBtn(!nearBottom)
    }
    el.addEventListener('scroll', onScroll)
    return () => el.removeEventListener('scroll', onScroll)
  }, [room?.id])

  // load the list of rooms to join
  useEffect(() => {
    if (!token) return
    useRoomStore.getState().fetchRooms()
    useRoomStore.getState().fetchMyRooms()
  }, [token])

  // runs when joining a room: fetch messages + open websocket
  useEffect(() => {
    if (!room) return
    useChatStore.getState().fetchMessages(room.id)
    useChatStore.getState().connect(token, room.id)
    return () => useChatStore.getState().disconnect()
  }, [room?.id])

  const send = () => {
    if (input.trim()) {
      useChatStore.getState().send(room.id, input)
    }
  }

  const deleteMessage = (id) => {
    useChatStore.getState().deleteMessage(room.id, id)
  }

  return (
    <div className="relative flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4 flex items-center gap-3">
        <button
          onClick={() => {
            useChatStore.getState().clearChat()
            useRoomStore.getState().openRoom(null)
          }}
          className="bg-zinc-800 hover:bg-zinc-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
        >
          ← Back
        </button>
        <h1 className="text-lg font-semibold">Room: {room.name}</h1>
        {nextLink && (
          <button
            onClick={loadOlder}
            className="bg-zinc-800 hover:bg-zinc-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
          >
            Load History
          </button>
        )}
        <button
          onClick={leaveRoom}
          className="ml-auto bg-zinc-800 hover:bg-red-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
        >
          Leave
        </button>
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((m, i) => (

          // skip rendering offscreen rows so long history stays smooth
          <div key={i} className={`[content-visibility:auto] flex ${m.mine ? 'justify-end' : 'justify-start'}`}>
            <div className="max-w-[80%]">
              <div className={`text-xs text-zinc-500 mb-1 ${m.mine ? 'text-right' : 'text-left'}`}>
                {m.sender} · {formatTime(m.created_at)}
              </div>
              <div className={`rounded-lg px-4 py-2 ${m.mine ? 'bg-blue-600' : 'bg-zinc-800'}`}>
                {m.deleted ? <span className="italic text-zinc-200">**message deleted**</span> : m.text}
              </div>
            </div>
            {m.mine && m.id != null && !m.deleted && (
              <button
                onClick={() => deleteMessage(m.id)}
                className="self-start text-xs text-zinc-500 hover:text-red-400 cursor-pointer"
              >
                delete
              </button>
            )}
          </div>
        ))}
      </div>

      <div className="border-t border-zinc-700 p-4 flex gap-2">
        <input
          className="flex-1 bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
          value={input}
          onChange={(e) => useChatStore.getState().setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && send()}
          placeholder="Type a message..."
        />
        <button
          className="bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer"
          onClick={send}
        >
          Send
        </button>
      </div>

      {showBottomBtn && (
        <button
          onClick={scrollToBottom}
          className="absolute bottom-20 right-4 z-10 bg-blue-600 hover:bg-blue-700 rounded-full w-10 h-10 text-xl flex items-center justify-center cursor-pointer shadow-lg"
          title="Jump to latest"
        >
          ↓
        </button>
      )}
    </div>

  )
}
