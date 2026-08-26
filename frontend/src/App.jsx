import { useAuthStore } from './stores/AuthStore.js'
import { useRoomStore } from './stores/RoomStore.js'
import AuthView from './views/AuthView.jsx'
import RoomPickerView from './views/RoomPickerView.jsx'
import ChatRoomView from './views/ChatRoomView.jsx'

function App() {
  const token = useAuthStore((s) => s.token)
  const room = useRoomStore((s) => s.room)

  // login/register view
  if (!token) {
    return (
      <AuthView />
    )
  }

  // room picker view
  if (!room) {
    return (
      <RoomPickerView />
    )
  }

  // chat view
  return (
    <ChatRoomView />
  )
}

export default App
