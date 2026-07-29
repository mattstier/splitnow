import { useState, useEffect, useRef } from 'react'

const ROOM = 'test'

function App() {
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const wsRef = useRef(null)

  useEffect(() => {
    const ws = new WebSocket(`ws://localhost:5173/ws?room=${ROOM}`)
    ws.onmessage = (e) => setMessages((prev) => [...prev, {text: e.data, mine: false}])
    wsRef.current = ws
    return () => ws.close()
  }, [])

  const send = () => {
    if (input.trim()) {
      wsRef.current.send(input)
      setMessages(prev => [...prev, { text: input, mine: true }])
      setInput('')
    }
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4">
        <h1 className="text-lg font-semibold">Room: {ROOM}</h1>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {messages.map((m, i) => (
          <div key={i} className={`flex ${m.mine ? 'justify-end' : 'justify-start'}`}>
            <div className={`rounded-lg px-4 py-2 max-w-[80%] ${m.mine ? 'bg-blue-600' : 'bg-zinc-800'}`}>
              {m.text}
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
