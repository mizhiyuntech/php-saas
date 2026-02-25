import { defineStore } from 'pinia'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    siteTitle: '鱼跃授权',
    siteDescription: '',
    siteKeywords: '',
    footerCopyright: '',
    policeRecord: '',
    siteIcon: '',
    siteFavicon: '',
    installed: false
  }),
  actions: {
    async fetchSiteInfo() {
      try {
        const { get } = useApi()
        const res = await get('/api/site-info')
        if (res.code === 0 && res.data) {
          this.siteTitle = res.data.site_title || '鱼跃授权'
          this.siteDescription = res.data.site_description || ''
          this.siteKeywords = res.data.site_keywords || ''
          this.footerCopyright = res.data.footer_copyright || ''
          this.policeRecord = res.data.police_record || ''
          this.siteIcon = res.data.site_icon || ''
          this.siteFavicon = res.data.site_favicon || ''
        }
      } catch (e) {
        // ignore
      }
    },
    async checkInstallStatus() {
      try {
        const { get } = useApi()
        const res = await get('/api/install/status')
        if (res.code === 0) {
          this.installed = res.data.installed
        }
      } catch (e) {
        // ignore
      }
    },
    updateTitle(title: string) {
      this.siteTitle = title
      if (import.meta.client) {
        document.title = title
      }
    }
  }
})
