import { defineStore } from 'pinia'

interface Tab {
  id: string
  title: string
  path: string
  closable: boolean
}

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    tabs: [
      { id: 'dashboard', title: '仪表盘', path: '/', closable: false }
    ] as Tab[],
    activeTab: 'dashboard'
  }),
  actions: {
    addTab(tab: Tab) {
      const existing = this.tabs.find(t => t.id === tab.id)
      if (!existing) {
        this.tabs.push(tab)
      }
      this.activeTab = tab.id
    },
    removeTab(id: string) {
      const index = this.tabs.findIndex(t => t.id === id)
      if (index === -1) return

      const tab = this.tabs[index]
      if (!tab.closable) return

      this.tabs.splice(index, 1)

      if (this.activeTab === id) {
        const newIndex = Math.min(index, this.tabs.length - 1)
        this.activeTab = this.tabs[newIndex]?.id || 'dashboard'
        return this.tabs[newIndex]?.path || '/'
      }
      return null
    },
    setActive(id: string) {
      this.activeTab = id
    }
  }
})
