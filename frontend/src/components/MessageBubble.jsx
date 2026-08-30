import { formatTime } from "../utils/Time"
// components for a row, containing the message content in a bubble format, as well as <thead>
// header with the sender, timestamp and delete button
export default function MessageBubble({ message, onDelete }) {
  const textAlign = message.mine ? 'text-right' : 'text-left'
  const bubbleClass = message.status === 'pending'
    ? 'bg-yellow-600'
    : message.mine && message.status === 'received' ? 'bg-blue-600'
      : 'bg-zinc-800'

  return (
    <div className="max-w-[80%]">
      <div className={`text-xs text-zinc-500 mb-1 ${textAlign}`}>
        {message.sender} · {formatTime(message.created_at)}
      </div>
      <div className={`rounded-lg px-4 py-2 ${bubbleClass}`}>
        {message.deleted ? <span className="italic text-zinc-200">**message deleted**</span> : message.text}
      </div>
      {message.mine && message.id != null && !message.deleted && (
        <button
          onClick={() => onDelete(message.id)}
          className="self-start text-xs text-zinc-500 hover:text-red-400 cursor-pointer"
        >
          delete
        </button>
      )}
    </div>
  )
}
