export default defineNuxtConfig({
  modules: ['@nuxt/ui', '@pinia/nuxt'],
  ssr: false,
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE || 'http://localhost:3132'
    }
  },
  compatibilityDate: '2025-01-01',
  app: {
    head: {
      charset: 'utf-8',
      viewport: 'width=device-width, initial-scale=1',
      title: '鱼跃授权'
    }
  },
  nitro: {
    output: {
      publicDir: '.output/public'
    }
  }
})
