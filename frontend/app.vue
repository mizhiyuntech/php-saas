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

useHead({
  titleTemplate: (title) => {
    const siteTitle = settingsStore.siteTitle || '鱼跃授权'
    return title ? `${title} - ${siteTitle}` : siteTitle
  }
})
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
