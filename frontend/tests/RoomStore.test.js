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

  it('fetches subscribed rooms ', async () => {
    roomService.fetchMyRooms.mockResolvedValue({
      data: [
        { id: 1, name: 'Room1' },
        { id: 2, name: 'Room2' }
      ]
    })
    await useRoomStore.getState().fetchMyRooms()
    // assert no error state
    expect(useRoomStore.getState().lastError).toBeNull()
    expect(useRoomStore.getState().myRooms).toEqual(
      [
        { id: 1, name: 'Room1' },
        { id: 2, name: 'Room2' }
      ]
    )
  })

  it('fetches subscribed rooms with API error', async () => {
    roomService.fetchMyRooms.mockRejectedValue(new Error('Network Error'))
    await useRoomStore.getState().fetchMyRooms()
    // assert no error state
    expect(useRoomStore.getState().lastError).not.toBeNull()
    // assert that myRooms is unchanged (empty) on error
    expect(useRoomStore.getState().myRooms).toEqual([])
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

describe('joinRoom', () => {
  it('joins a room and adds it myRooms list', async () => {

    // seed the room list so the store can find the room to join
    useRoomStore.setState({ rooms: [{ id: 3, name: 'New Room' }] })

    roomService.joinRoom.mockResolvedValue({
      data: { id: 3, name: 'New Room' }
    })

    await useRoomStore.getState().joinRoom(3)
    const state = useRoomStore.getState()
    // assert no error state
    expect(state.lastError).toBeNull()
    // check that current room is set correctly
    expect(state.room).toEqual({ id: 3, name: 'New Room' })
    // check that it is under myRooms
    expect(state.myRooms).toContainEqual({ id: 3, name: 'New Room' })
  })

  it('tests that failing to join a room is handled', async () => {
    roomService.joinRoom.mockRejectedValue(new Error('Network Error'))
    await useRoomStore.getState().joinRoom(3)
    const state = useRoomStore.getState()
    // assert error state
    expect(state.lastError).not.toBeNull()
    // check that current room is not set
    expect(state.room).toBeNull()
    // check that room is not added to myRooms
    expect(state.myRooms).toEqual([])
  })

  it('that joining room is idempotent', async () => {
    // seed the room list so the store can find the room to join
    useRoomStore.setState({
      myRooms: [{ id: 3, name: 'New Room' }],
      rooms: [{ id: 3, name: 'New Room' }]
    })

    roomService.joinRoom.mockResolvedValue({
      data: {}
    })
    await useRoomStore.getState().joinRoom(3)
    const state = useRoomStore.getState()
    // assert no error state
    expect(state.lastError).toBeNull()
    // check that it contains the room exactly once in myRooms
    expect(state.myRooms).toEqual([{ id: 3, name: 'New Room' }])
    // check that current room is not set
    expect(state.room).toBeNull()
  })

  it('does not add room when join returns an error status', async () => {
    useRoomStore.setState({ rooms: [{ id: 3, name: 'New Room' }] })
    roomService.joinRoom.mockResolvedValue({ status: 500, data: {} })
    await useRoomStore.getState().joinRoom(3)
    const state = useRoomStore.getState()
    expect(state.lastError).toBeNull()
    expect(state.myRooms).toEqual([])
    expect(state.room).toBeNull()
  })
})

describe('leaveRoom', () => {
  it('leaves a room and removes it from myRooms list', async () => {
    // seed the room list so the store can find the room to leave
    useRoomStore.setState({
      room: { id: 3, name: 'New Room' },
      myRooms: [{ id: 3, name: 'New Room' }],
    })

    roomService.leaveRoom.mockResolvedValue({
      data: { id: 3, name: 'New Room' }
    })

    await useRoomStore.getState().leaveRoom(3)
    const state = useRoomStore.getState()
    // assert no error state
    expect(state.lastError).toBeNull()
    // check that current room is reset correctly
    expect(state.room).toBeNull()
    // check that it is removed from myRooms
    expect(state.myRooms).not.toContainEqual({ id: 3, name: 'New Room' })
  })

  it('leaves a room that is already unselected', async () => {
    // seed the room as null
    useRoomStore.setState({
      room: null,
    })

    roomService.leaveRoom.mockResolvedValue({
      data: { id: 3, name: 'New Room' }
    })

    await useRoomStore.getState().leaveRoom(3)
    const state = useRoomStore.getState()
    // assert no error state
    expect(state.lastError).toBeNull()
    // check that it is removed from myRooms
    expect(state.myRooms).not.toContainEqual({ id: 3, name: 'New Room' })
  })


  it('leaves room, fails with an API error', async () => {
    // seed the room list so the store can find the room to leave
    useRoomStore.setState({
      room: { id: 3, name: 'New Room' },
      myRooms: [{ id: 3, name: 'New Room' }],
    })

    roomService.leaveRoom.mockRejectedValue(new Error('Network Error'))

    await useRoomStore.getState().leaveRoom(3)
    const state = useRoomStore.getState()
    // assert error state
    expect(state.lastError).not.toBeNull()
    // check that current room is not changed
    expect(state.room).toEqual({ id: 3, name: 'New Room' })
    // check that current myRooms is not changed
    expect(state.myRooms).toContainEqual({ id: 3, name: 'New Room' })
  })
})
