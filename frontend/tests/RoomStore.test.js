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
  useRoomStore.setState({ rooms: [], myRooms: [], room: null, newRoomName: '', lastError: null })
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
    // assert no error state
    expect(useRoomStore.getState().lastError).toBeNull()
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
    // assert no error state
    expect(useRoomStore.getState().lastError).toBeNull()
    expect(useRoomStore.getState().rooms).toEqual([])
  })

  it('fetch empty rooms API error', async () => {
    roomService.fetchRooms.mockRejectedValue(new Error('Network Error'))
    // on reject test fails
    await useRoomStore.getState().fetchRooms()
    // assert error state
    expect(useRoomStore.getState().lastError).not.toBeNull()
    // check that it leaves rooms unchanged when API fails
    expect(useRoomStore.getState().rooms).toEqual([])
  })
})

describe('createRoom', () => {
  it('creates a room and adds it to lists', async () => {
    roomService.createRoom.mockResolvedValue({
      data: { id: 3, name: 'New Room' }
    })

    await useRoomStore.getState().createRoom('New Room')

    const state = useRoomStore.getState()

    // assert no error state
    expect(useRoomStore.getState().lastError).toBeNull()
    // check that it is in the list of all rooms
    expect(state.rooms).toContainEqual({ id: 3, name: 'New Room' })
    // check that it is under myRooms
    expect(state.myRooms).toContainEqual({ id: 3, name: 'New Room' })
    // check that we automatically go to the room
    expect(state.room).toEqual({ id: 3, name: 'New Room' })
  })

  it('creates a room but return fails', async () => {
    roomService.createRoom.mockRejectedValue(
      new Error('Network Error')
    )

    await useRoomStore.getState().createRoom('New Room')

    const state = useRoomStore.getState()
    expect(state.rooms).toEqual([])
    expect(state.myRooms).toEqual([])
    expect(state.room).toBeNull()
    // assert error state
    expect(useRoomStore.getState().lastError).not.toBeNull()
  })

})
