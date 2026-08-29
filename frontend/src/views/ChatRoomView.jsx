import { useChatStore } from '../stores/ChatStore.js'
import { useRoomStore } from '../stores/RoomStore.js'
import { useAuthStore } from '../stores/AuthStore.js'
import { useRef, useState, useEffect } from 'react'
import InputField from '../components/InputField.jsx'
import PrimaryButton from '../components/PrimaryButton.jsx'
import SecondaryButton from '../components/SecondaryButton.jsx'
import MessageBubble from '../components/MessageBubble.jsx'

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
  const lastError = useChatStore(s => s.lastError)

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
      {
        lastError && (
          <div className="flex items-ce messagenter gap-2 bg-red-900/60 border border-red-700 text-red-100 px-4 py-2 text-sm">
            <span>Failed to send message</span>
          </div>
        )
      }
      <div className="border-b border-zinc-700 p-4 flex items-center gap-3">
        <SecondaryButton
          className="px-3 py-1 text-sm"
          onClick={() => {
            useChatStore.getState().clearChat()
            useRoomStore.getState().openRoom(null)
          }}
        >
          ← Back
        </SecondaryButton>
        <h1 className="text-lg font-semibold">Room: {room.name}</h1>
        {nextLink && (
          <SecondaryButton className="px-3 py-1 text-sm" onClick={loadOlder}>
            Load History
          </SecondaryButton>
        )}
        <SecondaryButton
          className="px-3 py-1 text-sm ml-auto hover:bg-red-700"
          onClick={leaveRoom}
        >
          Leave
        </SecondaryButton>
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages
          .filter((m) => m.status !== 'failed')
          .map((m, i) => (

            // skip rendering offscreen rows so long history stays smooth
            <div key={i} className={`[content-visibility:auto] flex ${m.mine ? 'justify-end' : 'justify-start'}`}>
              <MessageBubble key={i} message={m} onDelete={deleteMessage} />
            </div>
          ))}
      </div>

      <div className="border-t border-zinc-700 p-4 flex gap-2">
        <InputField
          className="flex-1"
          value={input}
          onChange={(e) => useChatStore.getState().setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && send()}
          placeholder="Type a message..."
        />
        <PrimaryButton onClick={send}>
          Send
        </PrimaryButton>
      </div>

      {showBottomBtn && (
        <PrimaryButton
          onClick={scrollToBottom}
          className="absolute bottom-20 right-4 z-10 rounded-full w-10 h-10 text-xl flex items-center justify-center shadow-lg"
          title="Jump to latest"
        >
          ↓
        </PrimaryButton>
      )}
    </div>
  )
}
