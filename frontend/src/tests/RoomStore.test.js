import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useRoomStore } from '../stores/RoomStore.js'

// mock the service at the module level — vitest replaces it before import
vi.mock('../services/RoomService.js', () => ({
  roomService: {
    fetchRooms: vi.fn(),
    fetchMyRooms: vi.fn(),
    createRoom: vi.fn(),
    joinRoom: vi.fn(),
    leaveRoom: vi.fn(),
  }
}))

import { roomService } from '../services/RoomService.js'

beforeEach(() => {
  // reset store between tests
  useRoomStore.setState({ rooms: [], myRooms: [], room: null, newRoomName: '' })
  vi.clearAllMocks()
})

describe('RoomStore', () => {
  it('fetches all rooms', async () => {
    roomService.fetchRooms.mockResolvedValue({ data: [{ id: 1, name: 'General' }] })

    await useRoomStore.getState().fetchRooms()

    expect(useRoomStore.getState().rooms).toEqual([{ id: 1, name: 'General' }])
  })
})
