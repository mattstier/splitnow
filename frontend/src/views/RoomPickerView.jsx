import { useRoomStore } from '../stores/RoomStore.js'
import { useChatStore } from '../stores/ChatStore.js'

// login/register view
export default function RoomPickerView() {
  const rooms = useRoomStore((s) => s.rooms)
  const myRooms = useRoomStore((s) => s.myRooms)
  const newRoomName = useRoomStore((s) => s.newRoomName)
  const myRoomIds = new Set(myRooms.map(r => r.id))
  const publicRooms = rooms.filter(r => !myRoomIds.has(r.id))

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

  // opens a room and clears the previous room's messages so they never leak in
  const openRoom = (r) => {
    useChatStore.getState().clearChat()
    useRoomStore.getState().openRoom(r)
  }

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
