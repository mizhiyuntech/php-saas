import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '' as string,
    username: '' as string
  }),
  getters: {
    isLoggedIn: (state) => !!state.token
  },
  actions: {
    setAuth(token: string, username: string) {
      this.token = token
      this.username = username
      if (import.meta.client) {
        localStorage.setItem('yuyue_token', token)
        localStorage.setItem('yuyue_username', username)
      }
    },
    loadFromStorage() {
      if (import.meta.client) {
        this.token = localStorage.getItem('yuyue_token') || ''
        this.username = localStorage.getItem('yuyue_username') || ''
      }
    },
    logout() {
      this.token = ''
      this.username = ''
      if (import.meta.client) {
        localStorage.removeItem('yuyue_token')
        localStorage.removeItem('yuyue_username')
      }
    }
  }
})
