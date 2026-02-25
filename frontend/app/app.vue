<script setup lang="ts">
const settingsStore = useSettingsStore()
const authStore = useAuthStore()

onMounted(async () => {
  authStore.loadFromStorage()
  await settingsStore.checkInstallStatus()
  if (settingsStore.installed) {
    await settingsStore.fetchSiteInfo()
  }
})

useHead(() => ({
  titleTemplate: (title?: string) => {
    const siteTitle = settingsStore.siteTitle || '鱼跃授权'
    return title ? `${title} - ${siteTitle}` : siteTitle
  },
  link: settingsStore.siteFavicon
    ? [{ rel: 'icon', type: 'image/x-icon', href: settingsStore.siteFavicon }]
    : []
}))
</script>

<template>
  <UApp>
    <NuxtLoadingIndicator color="#3b82f6" :height="3" />
    <NuxtLayout>
      <NuxtPage :transition="{ name: 'page', mode: 'out-in' }" />
    </NuxtLayout>
  </UApp>
</template>

<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.15s ease;
}
.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>
