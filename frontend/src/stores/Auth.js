import { create } from 'zustand'

export const useAuthStore = create((set) => ({
  token: localStorage.getItem('token') || '',

  finishAuth: (data) => {
    localStorage.setItem('token', data.token)
    set({ token: data.token })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: '' })
  },
}))
