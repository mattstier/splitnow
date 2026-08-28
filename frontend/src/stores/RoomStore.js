import { create } from 'zustand'
import { roomService } from '../services/RoomService.js'

export const useRoomStore = create((set, get) => ({
  rooms: [],
  myRooms: [],
  room: null,
  newRoomName: '',

  fetchRooms: () => {
    roomService.fetchRooms().then(res => set({ rooms: res.data }))
  },

  fetchMyRooms: () => {
    roomService.fetchMyRooms().then(res => set({ myRooms: res.data }))
  },

  createRoom: (name) => {
    roomService.createRoom(name).then(res => {
      const r = res.data
      set(s => ({
        rooms: [...s.rooms, r],
        myRooms: [...s.myRooms, r],
        room: { id: r.id, name: r.name }
      }))
    })
  },

  joinRoom: (id) => {
    const joined = get().rooms.find(r => r.id === id)
    roomService.joinRoom(id).then(res => {
      if (res.status >= 400 && res.status !== 409) return
      if (joined) {
        set(s => ({
          myRooms: [...s.myRooms, joined],
          room: { id: joined.id, name: joined.name }
        }))
      }
    })
  },

  leaveRoom: () => {
    const room = get().room
    if (!room) return
    roomService.leaveRoom(room.id).then(() => {
      set(s => ({
        myRooms: s.myRooms.filter(r => r.id !== room.id),
        room: null
      }))
    })
  },

  openRoom: (r) => set({ room: r }),
  setNewRoomName: (name) => set({ newRoomName: name }),
}))
