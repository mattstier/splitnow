import api from '../Api.js'

export const authService = {
  login: (email, password) => api.post('/login', { email, password }).then(r => r.data),
  register: (email, username, password) => api.post('/users', { email, username, password }),
}
