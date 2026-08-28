import { create } from 'zustand'
import { authService } from '../services/AuthService.js'
import { decodeJWT } from '../utils/JWT.js'

export const useAuthStore = create((set, get) => ({
  token: localStorage.getItem('token') || '',
  loginError: '',

  get username() {
    return decodeJWT(get().token)?.username || ''
  },

  register: (email, username, password) => {
    authService
      .register(email, username, password)
      .then(() => authService.login(email, password))
      .then((data) => {
        localStorage.setItem('token', data.token)
        set({ token: data.token, loginError: '' })
      })
      .catch((err) => {
        if (err.response?.status === 409)
          set({ loginError: 'username or email already taken' })
        else set({ loginError: 'registration failed' })
      })
  },

  login: (email, password) => {
    authService
      .login(email, password)
      .then((data) => {
        localStorage.setItem('token', data.token)
        set({ token: data.token, loginError: '' })
      })
      .catch(() => set({ loginError: 'invalid credentials' }))
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: '' })
  }
}))
