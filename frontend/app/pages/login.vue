<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const { post } = useApi()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const toast = useToast()
const router = useRouter()

const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

async function handleLogin() {
  if (!form.username || !form.password) {
    toast.add({ title: '请输入用户名和密码', color: 'error' })
    return
  }

  loading.value = true
  try {
    const res = await post('/api/auth/login', form)
    if (res.code === 0) {
      authStore.setAuth(res.data.token, res.data.username)
      toast.add({ title: '登录成功', color: 'success' })
      await settingsStore.fetchSiteInfo()
      router.push('/dashboard')
    } else {
      toast.add({ title: res.message || '登录失败', color: 'error' })
    }
  } catch (e: any) {
    toast.add({ title: e.message || '登录失败', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-sm mx-auto">
    <UCard>
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ settingsStore.siteTitle }}</h1>
          <p class="mt-1 text-sm text-gray-500">管理员登录</p>
        </div>
      </template>

      <div class="space-y-4">
        <UFormField label="用户名">
          <UInput v-model="form.username" class="w-full" placeholder="请输入用户名" @keyup.enter="handleLogin" />
        </UFormField>
        <UFormField label="密码">
          <UInput v-model="form.password" class="w-full" type="password" placeholder="请输入密码" @keyup.enter="handleLogin" />
        </UFormField>
      </div>

      <template #footer>
        <UButton block :loading="loading" @click="handleLogin">
          登录
        </UButton>
      </template>
    </UCard>
  </div>
</template>
