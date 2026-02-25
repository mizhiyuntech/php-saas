<script setup lang="ts">
const { get } = useApi()
const toast = useToast()

useHead({ title: '在线注入' })

const languages = ref<any[]>([])
const selectedLanguage = ref('')
const selectedFramework = ref('')
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const fileName = ref('')

const injectLicense = ref(false)
const licenseKey = ref('')
const programId = ref('')
const serverUrl = ref('')

const phpFrameworks = [
  { label: '原生PHP（无框架）', value: 'native' },
  { label: 'Laravel', value: 'laravel' },
  { label: 'ThinkPHP', value: 'thinkphp' },
  { label: 'Yii', value: 'yii' },
  { label: 'Symfony', value: 'symfony' },
  { label: 'CodeIgniter', value: 'codeigniter' },
  { label: 'Slim', value: 'slim' },
  { label: 'Hyperf', value: 'hyperf' },
  { label: 'Webman', value: 'webman' },
  { label: 'Workerman', value: 'workerman' }
]

onMounted(async () => {
  const res = await get('/api/injection/languages')
  if (res.code === 0) {
    languages.value = (res.data || []).map((l: any) => ({ label: l.label, value: l.value }))
  }
})

watch(selectedLanguage, () => {
  selectedFramework.value = ''
})

function onFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    const file = input.files[0]
    if (file.size > 30 * 1024 * 1024) {
      toast.add({ title: '文件大小不能超过30MB', color: 'error' })
      return
    }
    if (!file.name.toLowerCase().endsWith('.zip')) {
      toast.add({ title: '仅支持ZIP格式文件', color: 'error' })
      return
    }
    selectedFile.value = file
    fileName.value = file.name
  }
}

async function handleInject() {
  if (!selectedLanguage.value) {
    toast.add({ title: '请先选择开发语言', color: 'error' })
    return
  }
  if (selectedLanguage.value === 'php' && !selectedFramework.value) {
    toast.add({ title: '请选择PHP框架', color: 'error' })
    return
  }
  if (!selectedFile.value) {
    toast.add({ title: '请上传ZIP文件', color: 'error' })
    return
  }
  if (injectLicense.value) {
    if (!licenseKey.value || !programId.value || !serverUrl.value) {
      toast.add({ title: '请填写完整的许可证信息', color: 'error' })
      return
    }
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('language', selectedLanguage.value)

    if (selectedLanguage.value === 'php' && selectedFramework.value) {
      formData.append('framework', selectedFramework.value)
    }

    if (injectLicense.value) {
      formData.append('inject_license', 'true')
      formData.append('license_key', licenseKey.value)
      formData.append('program_id', programId.value)
      formData.append('server_url', serverUrl.value)
    }

    const config = useRuntimeConfig()
    const authStore = useAuthStore()
    const baseURL = config.public.apiBase as string || ''

    const response = await fetch(`${baseURL}/api/injection/upload`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` },
      body: formData
    })

    if (response.ok) {
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      const disposition = response.headers.get('Content-Disposition')
      a.download = disposition?.match(/filename=(.+)/)?.[1] || 'authorized.zip'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
      toast.add({ title: '注入成功，文件已下载', color: 'success' })
    } else {
      const data = await response.json()
      toast.add({ title: data.message || '注入失败', color: 'error' })
    }
  } catch (e: any) {
    toast.add({ title: e.message || '注入失败', color: 'error' })
  } finally {
    uploading.value = false
  }
}

function clearFile() {
  selectedFile.value = null
  fileName.value = ''
}
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-4">
    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">在线注入授权</span>
      </template>

      <div class="space-y-6">
        <UAlert
          title="使用说明"
          description="上传程序ZIP包并选择开发语言，系统将自动注入授权验证SDK并修改入口文件。注入后的程序启动时会进行授权校验，未授权则显示激活页面。"
          color="info"
          variant="subtle"
        />

        <UFormField label="选择开发语言" required>
          <USelect
            v-model="selectedLanguage"
            :items="languages"
            value-key="value"
            placeholder="请选择程序的开发语言"
          />
        </UFormField>

        <!-- PHP framework selection -->
        <UFormField v-if="selectedLanguage === 'php'" label="选择PHP框架" required>
          <USelect
            v-model="selectedFramework"
            :items="phpFrameworks"
            value-key="value"
            placeholder="请选择使用的PHP框架"
          />
        </UFormField>

        <!-- Go source code warning -->
        <UAlert
          v-if="selectedLanguage === 'go'"
          title="Go语言注意事项"
          description="请上传Go项目的源代码压缩包，不要上传编译后的二进制文件。注入需要修改源码中的 main.go 文件，编译后的程序无法进行注入操作。上传后请重新编译项目以使授权验证生效。"
          color="warning"
          variant="subtle"
        />

        <UFormField label="上传程序包" required hint="仅支持ZIP格式，最大30MB">
          <div class="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-6 text-center">
            <template v-if="!fileName">
              <UIcon name="i-lucide-upload" class="w-8 h-8 text-gray-400 mx-auto mb-2" />
              <p class="text-sm text-gray-500 mb-2">点击选择ZIP文件</p>
              <label>
                <UButton variant="outline" color="neutral" as="span" class="cursor-pointer">选择文件</UButton>
                <input type="file" accept=".zip" class="hidden" @change="onFileSelect" />
              </label>
            </template>
            <template v-else>
              <div class="flex items-center justify-center gap-2">
                <UIcon name="i-lucide-file-archive" class="w-5 h-5 text-gray-500" />
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ fileName }}</span>
                <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="clearFile" />
              </div>
            </template>
          </div>
        </UFormField>

        <!-- License injection -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">注入许可证</p>
              <p class="text-xs text-gray-500 mt-0.5">预填授权信息，加密注入程序中，启动时自动验证</p>
            </div>
            <USwitch v-model="injectLicense" />
          </div>
          <template v-if="injectLicense">
            <UFormField label="授权码" required>
              <UInput v-model="licenseKey" placeholder="XXXX-XXXX-XXXX-XXXX" />
            </UFormField>
            <UFormField label="程序ID" required>
              <UInput v-model="programId" placeholder="在程序管理中创建后获取" />
            </UFormField>
            <UFormField label="授权服务器地址" required>
              <UInput v-model="serverUrl" placeholder="https://your-auth-server.com" />
            </UFormField>
          </template>
        </div>

        <UButton
          block
          size="lg"
          :loading="uploading"
          :disabled="!selectedLanguage || !selectedFile || (selectedLanguage === 'php' && !selectedFramework)"
          @click="handleInject"
        >
          开始注入
        </UButton>
      </div>
    </UCard>
  </div>
</template>
