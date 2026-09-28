import { create } from 'zustand'
import { authService } from '../services/AuthService.js'
import { decodeJWT } from '../utils/JWT.js'

export const useAuthStore = create((set, get) => ({
  token: localStorage.getItem('token') || '',
  lastError: '',
  registrationSuccessful: false,

  get username() {
    return decodeJWT(get().token)?.username || ''
  },

  register: async (email, username, password) => {
    return authService
      .register(email, username, password)
      .then(() => {
        set({ registrationSuccessful: true, lastError: '' })
      })
      .catch((err) => {
        set({
          registrationSuccessful: false,
          lastError: err.response?.status === 409 ? 'username or email already taken' : 'registration failed'
        })
      })
  },

  login: async (email, password) => {
    return authService
      .login(email, password)
      .then((data) => {
        localStorage.setItem('token', data.token)
        set({ token: data.token, lastError: '' })
      })
      .catch((err) => {
        set({
          registrationSuccessful: false,
          lastError: err.response?.status === 401 ? 'invalid credentials' : 'login failed'
        })
      })
  },

  registerAndLogin: async (email, username, password) => {
    await get().register(email, username, password)
    if (!get().registrationSuccessful) return
    await get().login(email, password)
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: '' })
  }
}))
