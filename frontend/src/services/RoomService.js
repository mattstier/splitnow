import api from '../Api.js'

export const roomService = {
  fetchRooms: () => api.get('/rooms'),
  fetchMyRooms: () => api.get('/rooms/mine'),
  createRoom: (name) => api.post('/rooms', { name }),
  joinRoom: (id) => api.post(`/rooms/${id}/join`, null, { validateStatus: () => true }),
  leaveRoom: (id) => api.post(`/rooms/${id}/leave`),
}
