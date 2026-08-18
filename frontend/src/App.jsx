import { useState, useEffect, useRef } from 'react'

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('token') || '')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const [mode, setMode] = useState('login')
  const [loginError, setLoginError] = useState('')
  const [room, setRoom] = useState(null)
  const [rooms, setRooms] = useState([])
  const [myRooms, setMyRooms] = useState([])
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [newRoomName, setNewRoomName] = useState('')
  const wsRef = useRef(null)

  // own username, decoded from the JWT, used to flag own messages in history
  const myUsername = (() => {
    try {
      return JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))).username
    } catch {
      return ''
    }
  })()

  // makes a fetch with the token attached, and logs out on 401
  const authedFetch = (url, opts = {}) =>
    fetch(url, {
      ...opts,
      headers: { ...opts.headers, Authorization: 'Bearer ' + token }
    }).then(res => {
      if (res.status === 401) setToken('')
      return res
    })

  const finishAuth = (data) => {
    localStorage.setItem('token', data.token)
    setToken(data.token)
  }

  const login = () => {
    fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    })
      .then(res => {
        if (!res.ok) { setLoginError('invalid credentials'); return null }
        return res.json()
      })
      .then(data => { if (data) finishAuth(data) })
  }

  const register = () => {
    fetch('/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, username, password })
    })
      .then(res => {
        if (res.status === 409) { setLoginError('username or email already taken'); return null }
        if (!res.ok) { setLoginError('registration failed'); return null }
        // /users returns no token, so log in right after
        return fetch('/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        })
      })
      .then(res => res && res.json())
      .then(data => data && finishAuth(data))
  }

  const submit = mode === 'login' ? login : register

  // opens a room and clears the previous room's messages so they never leak in
  const openRoom = (r) => {
    setMessages([])
    setRoom(r)
  }

  // registers a new room and joins it
  const createRoom = () => {
    if (!newRoomName.trim()) return
    authedFetch('/rooms', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newRoomName })
    })
      .then(res => { if (!res.ok) return; return res.json() })
      .then(r => {
        if (!r) return
        setRooms(prev => [...prev, r])
        // the creator is auto-added as a member on the backend
        setMyRooms(prev => [...prev, r])
        openRoom({ id: r.id, name: r.name })
      })
  }

  // joins a public room and opens it
  const joinRoom = (id) => {
    authedFetch(`/rooms/${id}/join`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
      .then(res => {
        if (!res.ok && res.status !== 409) return // 409 = already a member
        const joined = rooms.find(r => r.id === id)
        if (joined) {
          setMyRooms(prev => [...prev, joined])
          openRoom({ id: joined.id, name: joined.name })
        }
      })
  }

  // leaves the current room and returns to the picker
  const leaveRoom = () => {
    if (!room) return
    authedFetch(`/rooms/${room.id}/leave`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
      .then(res => {
        if (!res.ok) return
        setMyRooms(prev => prev.filter(r => r.id !== room.id))
        setMessages([])
        setRoom(null)
      })
  }

  // load the list of rooms to join
  useEffect(() => {
    if (!token) return
    authedFetch('/rooms')
      .then(res => res.json())
      .then(setRooms)
    authedFetch('/rooms/mine')
      .then(res => res.json())
      .then(setMyRooms)
  }, [token])

  // runs when joining a room, opens a websocket
  useEffect(() => {
    if (!room) return

    // get all messages of the given room before opening the socket
    authedFetch('/messages?room=' + room.id)
      .then(res => {
        if (!res.ok) throw new Error('failed to load messages')
        return res.json()
      })
      .then(history => {
        setMessages(history.map(m => ({ id: m.id, deleted: m.deleted, text: m.content, sender: m.sender, created_at: m.created_at, mine: m.sender === myUsername })))
      })
      .catch(() => setMessages([]))

    const ws = new WebSocket('ws://localhost:5173/ws?token=' + token)
    ws.onopen = () => ws.send(JSON.stringify({ type: 'subscribe', room: room.id }))
    // here we define the message format with a flag 'mine' to handle own messages
    ws.onmessage = (e) => {
      const d = JSON.parse(e.data)
      // server-confirmed deletion: replace the message in place, keeping order
      if (d.deleted) {
        setMessages(prev => prev.map(m => m.id === d.id ? { ...m, deleted: true, text: '' } : m))
        return
      }
      if (d.content === undefined) return
      setMessages((prev) => [...prev, { id: d.id, deleted: false, text: d.content, sender: d.sender, created_at: d.created_at, mine: false }])
    }
    wsRef.current = ws
    return () => ws.close()
  }, [room?.id])

  const send = () => {
    if (input.trim()) {
      // sends rawtext to backend
      wsRef.current.send(JSON.stringify({ type: 'send', room: room.id, content: input }))
      // flags message as own
      setMessages(prev => [...prev, { text: input, sender: 'You', created_at: new Date().toISOString(), mine: true }])
      // clearing text box
      setInput('')
    }
  }

  // deletes a message; the confirmed update arrives via the ws broadcast
  const deleteMessage = (id) => {
    wsRef.current.send(JSON.stringify({ type: 'delete', room: room.id, message: id }))
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
            onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setLoginError('') }}
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
            onChange={(e) => setNewRoomName(e.target.value)}
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
    <div className="flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4 flex items-center gap-3">
        <button
          onClick={() => { setMessages([]); setRoom(null) }}
          className="bg-zinc-800 hover:bg-zinc-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
        >
          ← Back
        </button>
        <h1 className="text-lg font-semibold">Room: {room.name}</h1>
        <button
          onClick={leaveRoom}
          className="ml-auto bg-zinc-800 hover:bg-red-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
        >
          Leave
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((m, i) => (
          <div key={i} className={`flex ${m.mine ? 'justify-end' : 'justify-start'}`}>
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
          onChange={(e) => setInput(e.target.value)}
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
    </div>
  )
}

export default App
