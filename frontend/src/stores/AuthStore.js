import { create } from 'zustand'
import { authService } from '../services/AuthService.js'
import { decodeJWT } from '../utils/JWT.js'

export const useAuthStore = create((set, get) => ({
  token: localStorage.getItem('token') || '',
  lastError: '',

  get username() {
    return decodeJWT(get().token)?.username || ''
  },

  register: async (email, username, password) => {
    return authService
      .register(email, username, password)
      .catch((err) => {
        if (err.response?.status === 409) {
          set({ lastError: 'username or email already taken' })
        } else {
          set({ lastError: 'registration failed' })
        }
      })
  },

  login: async (email, password) => {
    return authService
      .login(email, password)
      .then((data) => {
        localStorage.setItem('token', data.token)
        set({ token: data.token, lastError: '' })
      })
      .catch(() => set({ lastError: 'invalid credentials' }))
  },

  registerAndLogin: async (email, username, password) => {
    return get()
      .register(email, username, password)
      .then(() => get().login(email, password))
      .catch((err) => {
        set({ lastError: err.message || 'registration or login failed' })
      })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: '' })
  }
}))
