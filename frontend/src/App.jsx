import { useState, useEffect, useRef } from 'react'

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('token') || '')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loginError, setLoginError] = useState('')
  const [room, setRoom] = useState(null)
  const [rooms, setRooms] = useState([])
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [newRoomName, setNewRoomName] = useState('')
  const wsRef = useRef(null)

  // makes a fetch with the token attached, and logs out on 401
  const authedFetch = (url, opts = {}) =>
    fetch(url, {
      ...opts,
      headers: { ...opts.headers, Authorization: 'Bearer ' + token }
    }).then(res => {
      if (res.status === 401) setToken('')
      return res
    })

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
      .then(data => {
        if (!data) return
        localStorage.setItem('token', data.token)
        setToken(data.token)
      })
  }

  // registers a new room and joins it
  const createRoom = () => {
    if (!newRoomName.trim()) return
    authedFetch('/rooms', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newRoomName, created_by: 1 })
    })
      .then(res => { if (!res.ok) return; return res.json() })
      .then(r => {
        if (!r) return
        setRooms(prev => [...prev, r])
        setRoom({ id: r.id, name: r.name })
      })
  }

  // load the list of rooms to join
  useEffect(() => {
    if (!token) return
    authedFetch('/rooms')
      .then(res => res.json())
      .then(setRooms)
  }, [token])

  // runs when joining a room, opens a websocket
  useEffect(() => {
    if (!room) return

    // get all messages of the given room before opening the socket
    authedFetch('/messages?room=' + room.id)
      .then(res => res.json())
      .then(history => {
      setMessages(history.map(m => ({ text: m.content, sender: m.sender, created_at: m.created_at, mine: false })))
    })

    const ws = new WebSocket('ws://localhost:5173/ws?token=' + token)
    ws.onopen = () => ws.send(JSON.stringify({ type: 'subscribe', room: room.id }))
    // here we define the message format with a flag 'mine' to handle own messages
    ws.onmessage = (e) => {
      const d = JSON.parse(e.data)
      if (d.content === undefined) return
      setMessages((prev) => [...prev, { text: d.content, sender: d.sender, created_at: d.created_at, mine: false }])
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

  // room picker view
  if (!token) {
    return (
      <div className="flex flex-col items-center justify-center h-screen bg-zinc-900 text-white">
        <div className="w-80 space-y-4">
          <h1 className="text-lg font-semibold text-center">Log in</h1>
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
            onKeyDown={(e) => e.key === 'Enter' && login()}
            placeholder="Password"
          />
          <button
            className="w-full bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer"
            onClick={login}
          >
            Log in
          </button>
          {loginError && <p className="text-red-500 text-sm text-center">{loginError}</p>}
        </div>
      </div>
    )
  }

  // room picker view
  if (!room) {
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

        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {rooms.length === 0 && <p className="text-zinc-500">no rooms yet</p>}
          {rooms.map(r => (
            <button
              key={r.id}
              onClick={() => setRoom({ id: r.id, name: r.name })}
              className="w-full text-left bg-zinc-800 hover:bg-zinc-700 rounded-lg px-4 py-2 font-medium cursor-pointer"
            >
              {r.name}
            </button>
          ))}
        </div>
      </div>
    )
  }

  // chat view
  return (
    <div className="flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4 flex items-center gap-3">
        <button
          onClick={() => setRoom(null)}
          className="bg-zinc-800 hover:bg-zinc-700 rounded-lg px-3 py-1 text-sm cursor-pointer"
        >
          ← Back
        </button>
        <h1 className="text-lg font-semibold">Room: {room.name}</h1>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((m, i) => (
          <div key={i} className={`flex ${m.mine ? 'justify-end' : 'justify-start'}`}>
            <div className="max-w-[80%]">
              <div className={`text-xs text-zinc-500 mb-1 ${m.mine ? 'text-right' : 'text-left'}`}>
                {m.sender} · {formatTime(m.created_at)}
              </div>
              <div className={`rounded-lg px-4 py-2 ${m.mine ? 'bg-blue-600' : 'bg-zinc-800'}`}>
                {m.text}
              </div>
            </div>
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
