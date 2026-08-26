import { create } from 'zustand'
import { authService } from '../services/AuthService.js'

export const useAuthStore = create((set, get) => ({
  token: localStorage.getItem('token') || '',
  loginError: '',

  get username() {
    try {
      return JSON.parse(atob(get().token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))).username
    } catch {
      return ''
    }
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
