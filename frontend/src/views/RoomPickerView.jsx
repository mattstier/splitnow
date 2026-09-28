import { useRoomStore } from '../stores/RoomStore.js'
import { useChatStore } from '../stores/ChatStore.js'
import { useEffect } from 'react'
import InputField from '../components/InputField.jsx'
import PrimaryButton from '../components/PrimaryButton.jsx'
import SecondaryButton from '../components/SecondaryButton.jsx'
import Banner from '../components/Banner.jsx'

// room picker view
export default function RoomPickerView() {
  const rooms = useRoomStore((s) => s.rooms)
  const myRooms = useRoomStore((s) => s.myRooms)
  const newRoomName = useRoomStore((s) => s.newRoomName)
  const lastError = useRoomStore((s) => s.lastError)
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
  }

  // opens a room and clears the previous room's messages so they never leak in
  const openRoom = (r) => {
    useChatStore.getState().clearChat()
    useRoomStore.getState().openRoom(r)
  }

  useEffect(() => {
    useRoomStore.getState().fetchRooms()
    useRoomStore.getState().fetchMyRooms()
  }, [])

  return (
    <div className="flex flex-col h-screen bg-zinc-900 text-white">
      <div className="border-b border-zinc-700 p-4">
        <h1 className="text-lg font-semibold">Rooms</h1>
      </div>

      <div className="border-b border-zinc-700 p-4 flex gap-2">
        <InputField
          className="flex-1"
          value={newRoomName}
          onChange={(e) => useRoomStore.getState().setNewRoomName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && createRoom()}
          placeholder="New room name..."
        />
        <PrimaryButton onClick={createRoom}>
          Create
        </PrimaryButton>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {lastError && <Banner tone="error">{lastError}</Banner>}

        <div>
          <h2 className="text-sm font-medium text-zinc-400 mb-2">My rooms</h2>
          {myRooms.length === 0 && !lastError && <p className="text-zinc-500">you have not joined any rooms yet</p>}
          {myRooms.map(r => (
            <SecondaryButton
              key={r.id}
              className="w-full text-left mb-2"
              onClick={() => openRoom({ id: r.id, name: r.name })}
            >
              {r.name}
            </SecondaryButton>
          ))}
        </div>

        <div>
          <h2 className="text-sm font-medium text-zinc-400 mb-2">All rooms</h2>
          {publicRooms.length === 0 && !lastError && <p className="text-zinc-500">no other rooms</p>}
          {publicRooms.map(r => (
            <div
              key={r.id}
              className="w-full flex items-center justify-between bg-zinc-800 rounded-lg px-4 py-2 mb-2"
            >
              <span className="font-medium">{r.name}</span>
              <PrimaryButton className="px-4 py-1 text-sm" onClick={() => joinRoom(r.id)}>
                Join
              </PrimaryButton>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
