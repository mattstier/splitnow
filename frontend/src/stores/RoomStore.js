import { create } from 'zustand'
import { roomService } from '../services/RoomService.js'

export const useRoomStore = create((set, get) => ({
  rooms: [],
  myRooms: [],
  room: null,
  lastError: null,
  newRoomName: '',

  fetchRooms: () => {
    return roomService.fetchRooms()
      .then(res => set({ rooms: res.data, lastError: null }))
      .catch(() => set({ rooms: [], lastError: 'Failed to fetch rooms' }))
  },

  fetchMyRooms: () => {
    return roomService.fetchMyRooms()
      .then(res => set({ myRooms: res.data, lastError: null }))
      .catch(() => set({ myRooms: [], lastError: 'Failed to fetch rooms' }))
  },

  createRoom: (name) => {
    return roomService.createRoom(name).then(res => {
      const r = res.data
      set(s => ({
        rooms: [...s.rooms, r],
        myRooms: [...s.myRooms, r],
        room: { id: r.id, name: r.name },
        lastError: null
      }))
    })
      .catch(() => set({ lastError: 'Failed to create room' }))
  },

  joinRoom: (id) => {
    const joined = get().rooms.find(r => r.id === id)
    return roomService.joinRoom(id).then(res => {
      if (res.status >= 400 && res.status !== 409) return
      if (joined) {
        // don't add the room if it's already in myRooms
        if (get().myRooms.some(r => r.id === joined.id)) return
        set(s => ({
          myRooms: [...s.myRooms, joined],
          room: { id: joined.id, name: joined.name },
          lastError: null
        }))
      }
    })
      .catch(() => set({ lastError: 'Failed to join room' }))
  },

  leaveRoom: () => {
    const room = get().room
    if (!room) return
    return roomService.leaveRoom(room.id).then(() => {
      set(s => ({
        myRooms: s.myRooms.filter(r => r.id !== room.id),
        room: null,
        lastError: null
      }))
    })
      .catch(() => set({ lastError: 'Failed to leave room' }))
  },

  openRoom: (r) => set({ room: r }),
  setNewRoomName: (name) => set({ newRoomName: name }),
}))
