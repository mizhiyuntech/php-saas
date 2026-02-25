<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const { get, post } = useApi()
const toast = useToast()
const router = useRouter()
const settingsStore = useSettingsStore()

const loading = ref(false)
const step = ref(1)

const form = reactive({
  mysql_host: '127.0.0.1',
  mysql_port: 3306,
  mysql_user: 'root',
  mysql_password: '',
  mysql_database: 'yuyue_auth',
  redis_host: '127.0.0.1',
  redis_port: 6379,
  redis_password: '',
  admin_username: 'admin',
  admin_password: '',
  admin_password_confirm: '',
  site_title: '鱼跃授权'
})

onMounted(async () => {
  const res = await get('/api/install/status')
  if (res.code === 0 && res.data.installed) {
    router.replace('/login')
  }
})

function nextStep() {
  if (step.value === 1) {
    if (!form.mysql_host || !form.mysql_user || !form.mysql_database) {
      toast.add({ title: '请填写完整的数据库配置', color: 'error' })
      return
    }
  }
  if (step.value === 2) {
    if (!form.redis_host) {
      toast.add({ title: '请填写Redis配置', color: 'error' })
      return
    }
  }
  step.value++
}

function prevStep() {
  step.value--
}

async function handleInstall() {
  if (!form.admin_username || !form.admin_password) {
    toast.add({ title: '请填写管理员信息', color: 'error' })
    return
  }
  if (form.admin_password.length < 6) {
    toast.add({ title: '密码长度不能少于6位', color: 'error' })
    return
  }
  if (form.admin_password !== form.admin_password_confirm) {
    toast.add({ title: '两次密码输入不一致', color: 'error' })
    return
  }

  loading.value = true
  try {
    const res = await post('/api/install', {
      mysql_host: form.mysql_host,
      mysql_port: form.mysql_port,
      mysql_user: form.mysql_user,
      mysql_password: form.mysql_password,
      mysql_database: form.mysql_database,
      redis_host: form.redis_host,
      redis_port: form.redis_port,
      redis_password: form.redis_password,
      admin_username: form.admin_username,
      admin_password: form.admin_password,
      site_title: form.site_title
    })

    if (res.code === 0) {
      toast.add({ title: '安装成功', color: 'success' })
      settingsStore.installed = true
      settingsStore.updateTitle(form.site_title)
      setTimeout(() => router.push('/login'), 1000)
    } else {
      toast.add({ title: res.message || '安装失败', color: 'error' })
    }
  } catch (e: any) {
    toast.add({ title: e.message || '安装失败', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-lg">
    <UCard>
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">鱼跃授权</h1>
          <p class="mt-1 text-sm text-gray-500">系统安装向导</p>
        </div>
      </template>

      <!-- Steps indicator -->
      <div class="flex items-center justify-center gap-2 mb-6">
        <template v-for="s in 3" :key="s">
          <div
            class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
            :class="step >= s
              ? 'bg-primary-500 text-white'
              : 'bg-gray-200 dark:bg-gray-700 text-gray-500'"
          >
            {{ s }}
          </div>
          <div v-if="s < 3" class="w-12 h-0.5" :class="step > s ? 'bg-primary-500' : 'bg-gray-200 dark:bg-gray-700'" />
        </template>
      </div>

      <!-- Step 1: MySQL -->
      <div v-show="step === 1" class="space-y-4">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">数据库配置</h3>
        <div class="grid grid-cols-2 gap-4">
          <UFormField label="数据库地址">
            <UInput v-model="form.mysql_host" placeholder="127.0.0.1" />
          </UFormField>
          <UFormField label="端口">
            <UInput v-model.number="form.mysql_port" type="number" placeholder="3306" />
          </UFormField>
        </div>
        <UFormField label="用户名">
          <UInput v-model="form.mysql_user" placeholder="root" />
        </UFormField>
        <UFormField label="密码">
          <UInput v-model="form.mysql_password" type="password" placeholder="数据库密码" />
        </UFormField>
        <UFormField label="数据库名">
          <UInput v-model="form.mysql_database" placeholder="yuyue_auth" />
        </UFormField>
      </div>

      <!-- Step 2: Redis -->
      <div v-show="step === 2" class="space-y-4">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">Redis配置</h3>
        <div class="grid grid-cols-2 gap-4">
          <UFormField label="Redis地址">
            <UInput v-model="form.redis_host" placeholder="127.0.0.1" />
          </UFormField>
          <UFormField label="端口">
            <UInput v-model.number="form.redis_port" type="number" placeholder="6379" />
          </UFormField>
        </div>
        <UFormField label="密码（选填）">
          <UInput v-model="form.redis_password" type="password" placeholder="Redis密码" />
        </UFormField>
      </div>

      <!-- Step 3: Admin -->
      <div v-show="step === 3" class="space-y-4">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">管理员设置</h3>
        <UFormField label="网站标题">
          <UInput v-model="form.site_title" placeholder="鱼跃授权" />
        </UFormField>
        <UFormField label="管理员用户名">
          <UInput v-model="form.admin_username" placeholder="admin" />
        </UFormField>
        <UFormField label="管理员密码">
          <UInput v-model="form.admin_password" type="password" placeholder="至少6位" />
        </UFormField>
        <UFormField label="确认密码">
          <UInput v-model="form.admin_password_confirm" type="password" placeholder="再次输入密码" />
        </UFormField>
      </div>

      <template #footer>
        <div class="flex justify-between">
          <UButton
            v-if="step > 1"
            variant="outline"
            color="neutral"
            @click="prevStep"
          >
            上一步
          </UButton>
          <div v-else />
          <UButton
            v-if="step < 3"
            @click="nextStep"
          >
            下一步
          </UButton>
          <UButton
            v-else
            :loading="loading"
            @click="handleInstall"
          >
            开始安装
          </UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
