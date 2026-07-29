import { useState, useEffect, useRef } from 'react'

const ROOM = 'test'

function App() {
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const wsRef = useRef(null)

  useEffect(() => {
    const ws = new WebSocket(`ws://localhost:5173/ws?room=${ROOM}`)
    ws.onmessage = (e) => setMessages((prev) => [...prev, e.data])
    wsRef.current = ws
    return () => ws.close()
  }, [])

  const send = () => {
    if (input.trim()) {
      wsRef.current.send(input)
      setInput('')
    }
  }

  return (
    <div style={{ padding: 20 }}>
      <h2>Room: {ROOM}</h2>
      <div style={{ marginBottom: 10 }}>
        {messages.map((m, i) => <div key={i}>{m}</div>)}
      </div>
      <input value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && send()} />
      <button onClick={send}>Send</button>
    </div>
  )
}

export default App
