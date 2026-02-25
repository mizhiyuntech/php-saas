export default defineNuxtRouteMiddleware(async (to) => {
  const authStore = useAuthStore()
  const settingsStore = useSettingsStore()

  if (import.meta.client) {
    authStore.loadFromStorage()
  }

  const publicPaths = ['/install', '/login', '/home', '/purchase']
  const isPublicPage = publicPaths.some(p => to.path === p || to.path.startsWith(p + '/'))

  if (to.path === '/install') {
    return
  }

  if (!settingsStore.installed && to.path !== '/install') {
    try {
      const config = useRuntimeConfig()
      const baseURL = config.public.apiBase as string || ''
      const res = await fetch(`${baseURL}/api/install/status`)
      const data = await res.json()
      if (data.code === 0) {
        settingsStore.installed = data.data.installed
      }
    } catch {
      // ignore
    }

    if (!settingsStore.installed) {
      return navigateTo('/install')
    }
  }

  if (to.path === '/') {
    if (authStore.isLoggedIn) {
      return navigateTo('/dashboard')
    }
    return navigateTo('/home')
  }

  if (!isPublicPage && !authStore.isLoggedIn) {
    return navigateTo('/login')
  }

  if (to.path === '/login' && authStore.isLoggedIn) {
    return navigateTo('/dashboard')
  }
})
