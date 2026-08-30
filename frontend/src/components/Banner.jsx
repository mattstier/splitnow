// banner for transient status/errors shown above the chat
const tones = {
  error: 'bg-red-900/60 border-red-700 text-red-100',
  offline: 'bg-yellow-900/60 border-yellow-700 text-yellow-100',
}

export default function Banner({ tone = 'offline', children }) {
  return (
    <div className={`flex items-center gap-2 border px-4 py-2 text-sm ${tones[tone]}`}>
      {children}
    </div>
  )
}
