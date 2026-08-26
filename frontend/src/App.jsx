import { useState, useEffect, useRef } from 'react'
import { useAuthStore } from './stores/AuthStore.js'
import { useRoomStore } from './stores/RoomStore.js'
import { useChatStore } from './stores/ChatStore.js'

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function App() {
  const token = useAuthStore((s) => s.token)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const [mode, setMode] = useState('login')
  const loginError = useAuthStore((s) => s.loginError)
  const room = useRoomStore((s) => s.room)
  const rooms = useRoomStore((s) => s.rooms)
  const myRooms = useRoomStore((s) => s.myRooms)
  const newRoomName = useRoomStore((s) => s.newRoomName)
  const messages = useChatStore((s) => s.messages)
  const nextLink = useChatStore((s) => s.nextLink)
  const input = useChatStore((s) => s.input)
  const [showBottomBtn, setShowBottomBtn] = useState(false)
  const scrollRef = useRef(null)
  const loadingOlderRef = useRef(false)

  const submit = mode === 'login'
    ? () => useAuthStore.getState().login(email, password)
    : () => useAuthStore.getState().register(email, username, password)

  // opens a room and clears the previous room's messages so they never leak in
  const openRoom = (r) => {
    useChatStore.getState().clearChat()
    setShowBottomBtn(false)
    useRoomStore.getState().openRoom(r)
  }

  // registers a new room and joins it
  const createRoom = () => {
    if (!newRoomName.trim()) return
    useRoomStore.getState().createRoom(newRoomName)
    useRoomStore.getState().setNewRoomName('')
  }

  // joins a public room and opens it
  const joinRoom = (id) => {
    useRoomStore.getState().joinRoom(id)
    const joined = rooms.find(r => r.id === id)
    if (joined) openRoom({ id: joined.id, name: joined.name })
  }

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

  // room picker view
  if (!token) {
    return (
      <div className="flex flex-col items-center justify-center h-screen bg-zinc-900 text-white">
        <div className="w-80 space-y-4">
          <h1 className="text-lg font-semibold text-center">{mode === 'login' ? 'Log in' : 'Register'}</h1>
          {mode === 'register' && (
            <input
              type="text"
              className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Username"
            />
          )}
          <input
            type="email"
            className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email"
          />
          <input
            type="password"
            className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && submit()}
            placeholder="Password"
          />
          <button
            className="w-full bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer"
            onClick={submit}
          >
            {mode === 'login' ? 'Log in' : 'Register'}
          </button>
          <button
            className="w-full text-zinc-400 hover:text-white text-sm cursor-pointer"
            onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); useAuthStore.setState({ loginError: '' }) }}
          >
            {mode === 'login' ? 'need an account? Register' : 'have an account? Log in'}
          </button>
          {loginError && <p className="text-red-500 text-sm text-center">{loginError}</p>}
        </div>
      </div>
    )
  }

  // room picker view
  if (!room) {
    const myRoomIds = new Set(myRooms.map(r => r.id))
    const publicRooms = rooms.filter(r => !myRoomIds.has(r.id))
    return (
      <div className="flex flex-col h-screen bg-zinc-900 text-white">
        <div className="border-b border-zinc-700 p-4">
          <h1 className="text-lg font-semibold">Rooms</h1>
        </div>

        <div className="border-b border-zinc-700 p-4 flex gap-2">
          <input
            className="flex-1 bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            value={newRoomName}
            onChange={(e) => useRoomStore.getState().setNewRoomName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && createRoom()}
            placeholder="New room name..."
          />
          <button
            className="bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer"
            onClick={createRoom}
          >
            Create
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          <div>
            <h2 className="text-sm font-medium text-zinc-400 mb-2">My rooms</h2>
            {myRooms.length === 0 && <p className="text-zinc-500">you have not joined any rooms yet</p>}
            {myRooms.map(r => (
              <button
                key={r.id}
                onClick={() => openRoom({ id: r.id, name: r.name })}
                className="w-full text-left bg-zinc-800 hover:bg-zinc-700 rounded-lg px-4 py-2 font-medium cursor-pointer mb-2"
              >
                {r.name}
              </button>
            ))}
          </div>

          <div>
            <h2 className="text-sm font-medium text-zinc-400 mb-2">All rooms</h2>
            {publicRooms.length === 0 && <p className="text-zinc-500">no other rooms</p>}
            {publicRooms.map(r => (
              <div
                key={r.id}
                className="w-full flex items-center justify-between bg-zinc-800 rounded-lg px-4 py-2 mb-2"
              >
                <span className="font-medium">{r.name}</span>
                <button
                  onClick={() => joinRoom(r.id)}
                  className="bg-blue-600 hover:bg-blue-700 rounded-lg px-4 py-1 text-sm font-medium cursor-pointer"
                >
                  Join
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  // chat view
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

export default App
