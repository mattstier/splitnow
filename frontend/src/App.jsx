import { useState, useEffect, useRef } from 'react'

const ROOM = 'test'

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function App() {
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const wsRef = useRef(null)

  // runs once, opens a websocket
  useEffect(() => {
    // get all messages of the given room before opening the socket
    fetch('/messages?room=' + ROOM)
      .then(res => res.json())
      .then(history => {
      setMessages(history.map(m => ({ text: m.content, sender: m.sender, created_at: m.created_at, mine: false })))
    })

    const ws = new WebSocket(`ws://localhost:5173/ws?room=${ROOM}`)
    // here we define the message format with a flag 'mine' to handle own messages
    ws.onmessage = (e) => {
      const d = JSON.parse(e.data)
      setMessages((prev) => [...prev, { text: d.content, sender: d.sender, created_at: d.created_at, mine: false }])
    }
    wsRef.current = ws
    return () => ws.close()
  }, [])

  const send = () => {
    if (input.trim()) {
      // sends rawtext to backend
      wsRef.current.send(input)
      // flags message as own
      setMessages(prev => [...prev, { text: input, sender: 'You', created_at: new Date().toISOString(), mine: true }])
      // clearing text box
      setInput('')
    }
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4">
        <h1 className="text-lg font-semibold">Room: {ROOM}</h1>
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
