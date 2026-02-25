<script setup lang="ts">
const { get, put, upload } = useApi()
const toast = useToast()
const settingsStore = useSettingsStore()

useHead({ title: '系统设置' })

const loading = ref(true)
const saving = ref(false)
const savingEncrypt = ref(false)

const encryptForm = reactive({
  mode: 'none',
  key: ''
})

const encryptModes = [
  { label: '不加密', value: 'none' },
  { label: '全局加密（所有接口）', value: 'full' },
  { label: '部分加密（仅核心接口）', value: 'partial' }
]

const form = reactive({
  site_title: '',
  site_description: '',
  site_keywords: '',
  footer_copyright: '',
  police_record: '',
  site_url: '',
  site_icon: '',
  site_favicon: ''
})

async function fetchSettings() {
  loading.value = true
  try {
    const [settingsRes, encRes] = await Promise.all([
      get('/api/settings'),
      get('/api/settings/encrypt')
    ])
    if (settingsRes.code === 0 && settingsRes.data) Object.assign(form, settingsRes.data)
    if (encRes.code === 0 && encRes.data) {
      encryptForm.mode = encRes.data.mode || 'none'
      encryptForm.key = encRes.data.key || ''
    }
  } finally {
    loading.value = false
  }
}

async function handleSaveEncrypt() {
  if (encryptForm.mode !== 'none' && encryptForm.key.length < 16) {
    toast.add({ title: '加密密钥至少16位', color: 'error' })
    return
  }
  savingEncrypt.value = true
  try {
    const res = await put('/api/settings/encrypt', encryptForm)
    if (res.code === 0) toast.add({ title: '保存成功', color: 'success' })
    else toast.add({ title: res.message, color: 'error' })
  } finally {
    savingEncrypt.value = false
  }
}

async function generateKey() {
  const { post } = useApi()
  const res = await post('/api/settings/encrypt/generate-key')
  if (res.code === 0) {
    encryptForm.key = res.data.key
    toast.add({ title: '密钥已生成', color: 'success' })
  }
}

async function handleSave() {
  saving.value = true
  try {
    const res = await put('/api/settings', {
      site_title: form.site_title,
      site_description: form.site_description,
      site_keywords: form.site_keywords,
      footer_copyright: form.footer_copyright,
      police_record: form.police_record,
      site_url: form.site_url
    })
    if (res.code === 0) {
      toast.add({ title: '保存成功', color: 'success' })
      settingsStore.updateTitle(form.site_title)
      settingsStore.footerCopyright = form.footer_copyright
      settingsStore.policeRecord = form.police_record
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    saving.value = false
  }
}

async function uploadIcon(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  const fd = new FormData()
  fd.append('file', input.files[0])
  try {
    const res = await upload('/api/settings/upload-icon', fd)
    if (res.code === 0) {
      form.site_icon = res.data.url
      settingsStore.siteIcon = res.data.url
      toast.add({ title: '图标上传成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } catch {
    toast.add({ title: '上传失败', color: 'error' })
  }
}

async function uploadFavicon(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  const fd = new FormData()
  fd.append('file', input.files[0])
  try {
    const res = await upload('/api/settings/upload-favicon', fd)
    if (res.code === 0) {
      form.site_favicon = res.data.url
      settingsStore.siteFavicon = res.data.url
      toast.add({ title: '小图标上传成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } catch {
    toast.add({ title: '上传失败', color: 'error' })
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-4">
    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">基本设置</span>
      </template>
      <div class="space-y-4">
        <UFormField label="网站标题">
          <UInput v-model="form.site_title" placeholder="请输入网站标题" />
        </UFormField>
        <UFormField label="网站地址" hint="用于生成支付回调链接等">
          <UInput v-model="form.site_url" placeholder="如 https://your-domain.com" />
        </UFormField>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">SEO设置</span>
      </template>
      <div class="space-y-4">
        <UFormField label="网站描述">
          <UTextarea v-model="form.site_description" placeholder="请输入网站描述" :rows="2" />
        </UFormField>
        <UFormField label="网站关键词">
          <UInput v-model="form.site_keywords" placeholder="多个关键词用逗号分隔" />
        </UFormField>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">底部信息</span>
      </template>
      <div class="space-y-4">
        <UFormField label="版权信息">
          <UInput v-model="form.footer_copyright" placeholder="如 (c) 2026 鱼跃授权" />
        </UFormField>
        <UFormField label="公安备案号">
          <UInput v-model="form.police_record" placeholder="如 京公网安备 00000000号" />
        </UFormField>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">网站图标</span>
      </template>
      <div class="grid grid-cols-2 gap-6">
        <div class="space-y-2">
          <p class="text-sm text-gray-600 dark:text-gray-400">网站图标（大图标）</p>
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 text-center">
            <div v-if="form.site_icon" class="mb-2">
              <img :src="form.site_icon" alt="Site Icon" class="w-16 h-16 mx-auto object-contain" />
            </div>
            <div v-else class="w-16 h-16 mx-auto mb-2 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
              <UIcon name="i-lucide-image" class="w-6 h-6 text-gray-400" />
            </div>
            <label>
              <UButton variant="outline" color="neutral" size="sm" as="span" class="cursor-pointer">选择图片</UButton>
              <input type="file" accept=".png,.jpg,.jpeg,.ico,.webp" class="hidden" @change="uploadIcon" />
            </label>
          </div>
        </div>
        <div class="space-y-2">
          <p class="text-sm text-gray-600 dark:text-gray-400">网站小图标（Favicon）</p>
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 text-center">
            <div v-if="form.site_favicon" class="mb-2">
              <img :src="form.site_favicon" alt="Favicon" class="w-16 h-16 mx-auto object-contain" />
            </div>
            <div v-else class="w-16 h-16 mx-auto mb-2 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
              <UIcon name="i-lucide-image" class="w-6 h-6 text-gray-400" />
            </div>
            <label>
              <UButton variant="outline" color="neutral" size="sm" as="span" class="cursor-pointer">选择图片</UButton>
              <input type="file" accept=".png,.jpg,.jpeg,.ico,.webp" class="hidden" @change="uploadFavicon" />
            </label>
          </div>
        </div>
      </div>
    </UCard>

    <div class="flex justify-end">
      <UButton size="lg" :loading="saving" @click="handleSave">保存设置</UButton>
    </div>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">接口加密</span>
      </template>
      <div class="space-y-4">
        <UAlert
          title="说明"
          description="开启后，授权验证和盗版检测等接口的请求/响应数据将使用AES-256-GCM加密传输，防止中间人拦截和明文嗅探。注入的SDK会自动适配加密模式。"
          color="info"
          variant="subtle"
        />
        <UFormField label="加密模式">
          <USelect v-model="encryptForm.mode" :items="encryptModes" value-key="value" />
        </UFormField>
        <UFormField v-if="encryptForm.mode !== 'none'" label="加密密钥" hint="至少16位，建议32位">
          <div class="flex gap-2">
            <UInput v-model="encryptForm.key" class="flex-1 font-mono" placeholder="加密密钥" />
            <UButton variant="outline" color="neutral" @click="generateKey">生成</UButton>
          </div>
        </UFormField>
        <div v-if="encryptForm.mode !== 'none'" class="text-xs text-gray-500 space-y-1">
          <p v-if="encryptForm.mode === 'full'">全局加密：所有验证接口均加密</p>
          <p v-if="encryptForm.mode === 'partial'">部分加密：仅加密 /api/license/verify、/api/piracy/check、/api/auth/login</p>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <UButton :loading="savingEncrypt" @click="handleSaveEncrypt">保存加密配置</UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
