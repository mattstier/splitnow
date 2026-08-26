import { create } from 'zustand'
import { authService } from '../services/Auth.js'

export const useAuthStore = create((set) => ({
  token: localStorage.getItem('token') || '',
  loginError: '',

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
