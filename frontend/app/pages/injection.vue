<script setup lang="ts">
const { get } = useApi()
const toast = useToast()

useHead({ title: '在线注入' })

const languages = ref<any[]>([])
const selectedLanguage = ref('')
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const fileName = ref('')

onMounted(async () => {
  const res = await get('/api/injection/languages')
  if (res.code === 0) {
    languages.value = (res.data || []).map((l: any) => ({ label: l.label, value: l.value }))
  }
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
  if (!selectedFile.value) {
    toast.add({ title: '请上传ZIP文件', color: 'error' })
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('language', selectedLanguage.value)

    const config = useRuntimeConfig()
    const authStore = useAuthStore()
    const baseURL = config.public.apiBase as string || ''

    const response = await fetch(`${baseURL}/api/injection/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      },
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
          description="上传程序ZIP包并选择开发语言，系统将自动注入授权验证SDK并修改入口文件。注入后的程序启动时会进行授权校验，未授权则显示激活页面。SDK文件夹已做伪装处理，变量及通信均已加密。"
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

        <UFormField label="上传程序包" required hint="仅支持ZIP格式，最大30MB">
          <div class="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-6 text-center">
            <template v-if="!fileName">
              <UIcon name="i-lucide-upload" class="w-8 h-8 text-gray-400 mx-auto mb-2" />
              <p class="text-sm text-gray-500 mb-2">点击选择ZIP文件</p>
              <label>
                <UButton variant="outline" color="neutral" as="span" class="cursor-pointer">
                  选择文件
                </UButton>
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

        <UButton
          block
          size="lg"
          :loading="uploading"
          :disabled="!selectedLanguage || !selectedFile"
          @click="handleInject"
        >
          开始注入
        </UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">注入机制说明</span>
      </template>
      <div class="text-sm text-gray-600 dark:text-gray-400 space-y-3">
        <div>
          <p class="font-medium text-gray-700 dark:text-gray-300 mb-1">SDK伪装</p>
          <p>授权验证文件隐藏在 lib/.cache/ 目录下，文件名伪装为框架引导文件（bootstrap.*），不易被识别</p>
        </div>
        <div>
          <p class="font-medium text-gray-700 dark:text-gray-300 mb-1">自动引入</p>
          <p>系统会自动扫描程序入口文件（如 index.php / app.js / main.py 等），注入SDK引用代码，无需手动修改</p>
        </div>
        <div>
          <p class="font-medium text-gray-700 dark:text-gray-300 mb-1">安全校验</p>
          <p>采用 XOR 加密存储配置、HMAC-SHA256 签名通信、时间戳防重放，授权信息加密存储于 .lic 文件</p>
        </div>
        <div>
          <p class="font-medium text-gray-700 dark:text-gray-300 mb-1">未授权拦截</p>
          <p>程序启动时自动校验授权，未通过则显示授权激活页面，需输入授权码、程序ID和服务器地址。可在「系统设置」中自定义此页面</p>
        </div>
      </div>
    </UCard>
  </div>
</template>
