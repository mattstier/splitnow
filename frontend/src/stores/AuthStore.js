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
          lastError: err.response?.status === 409 ? 'username or email already taken' : 'an unknown error has occured'
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
        if (err.response?.status === 401) {
          set({ lastError: 'invalid credentials' })
        } else if (!get().registrationSuccessful) {
          set({ lastError: 'registration failed' })
        } else {
          set({ lastError: 'an unknown error has occured' })
        }
      })
  },

  registerAndLogin: async (email, username, password) => {
    return get()
      .register(email, username, password)
      .then(() => get().login(email, password))
      .catch((err) => {
        set({ lastError: err.message })
      })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: '' })
  }
}))
