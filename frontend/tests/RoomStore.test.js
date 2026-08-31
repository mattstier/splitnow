import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useRoomStore } from '../src/stores/RoomStore.js'

// mock the service at the module level — vitest replaces it before import
vi.mock('../src/services/RoomService.js', () => ({
  roomService: {
    fetchRooms: vi.fn(),
    fetchMyRooms: vi.fn(),
    createRoom: vi.fn(),
    joinRoom: vi.fn(),
    leaveRoom: vi.fn(),
  }
}))

import { roomService } from '../src/services/RoomService.js'

beforeEach(() => {
  // reset store between tests
  useRoomStore.setState({ rooms: [], myRooms: [], room: null, newRoomName: '' })
  vi.clearAllMocks()
})

describe('fetchRooms', () => {

  it('fetch all rooms', async () => {
    roomService.fetchRooms.mockResolvedValue({
      data: [
        { id: 1, name: 'Room1' },
        { id: 2, name: 'Room2' },
      ]
    })
    await useRoomStore.getState().fetchRooms()
    expect(useRoomStore.getState().rooms).toEqual(
      [
        { id: 1, name: 'Room1' },
        { id: 2, name: 'Room2' }
      ]
    )
  })

  it('fetch empty rooms return', async () => {
    roomService.fetchRooms.mockResolvedValue({ data: [] })
    await useRoomStore.getState().fetchRooms()
    expect(useRoomStore.getState().rooms).toEqual([])
  })

  it('fetch empty rooms API error', async () => {
    roomService.fetchRooms.mockRejectedValue(new Error('Network Error'))
    const rooms = useRoomStore.getState().fetchRooms()
    // assert that it rejects/throws error
    await expect(rooms).rejects.toThrow()
    // check that it leaves rooms unchanged when API fails
    expect(useRoomStore.getState().rooms).toEqual([])
  })
})
