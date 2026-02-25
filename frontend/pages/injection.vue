<script setup lang="ts">
const { get, upload } = useApi()
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
          description="上传您的程序ZIP压缩包，选择对应的开发语言，系统将自动在压缩包内生成授权验证目录（_auth）和对应语言的授权验证代码文件。此操作可节省手动对接API的时间。"
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
              <p class="text-sm text-gray-500 mb-2">点击或拖拽上传ZIP文件</p>
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
        <span class="font-medium text-gray-900 dark:text-white">注入后的文件结构</span>
      </template>
      <div class="text-sm text-gray-600 dark:text-gray-400 space-y-1">
        <p>注入完成后，您的压缩包中将新增以下目录和文件：</p>
        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 mt-2 font-mono text-xs space-y-1">
          <p>your-project/</p>
          <p class="ml-4">_auth/</p>
          <p class="ml-8">check_license.* &nbsp;&nbsp;(授权验证代码)</p>
          <p class="ml-8">config.* &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(配置文件)</p>
          <p class="ml-8">README.md &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(使用说明)</p>
          <p class="ml-4">... (原有文件保持不变)</p>
        </div>
      </div>
    </UCard>
  </div>
</template>
