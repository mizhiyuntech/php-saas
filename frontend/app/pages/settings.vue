<script setup lang="ts">
const { get, put, upload } = useApi()
const toast = useToast()
const settingsStore = useSettingsStore()

useHead({ title: '系统设置' })

const loading = ref(true)
const saving = ref(false)
const savingAuth = ref(false)

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

const unauthHTML = ref('')
const showPreview = ref(false)

async function fetchSettings() {
  loading.value = true
  try {
    const [settingsRes, authRes] = await Promise.all([
      get('/api/settings'),
      get('/api/settings/unauth-page')
    ])
    if (settingsRes.code === 0 && settingsRes.data) {
      Object.assign(form, settingsRes.data)
    }
    if (authRes.code === 0 && authRes.data) {
      unauthHTML.value = authRes.data.html || ''
    }
  } finally {
    loading.value = false
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

async function handleSaveAuthPage() {
  savingAuth.value = true
  try {
    const res = await put('/api/settings/unauth-page', { html: unauthHTML.value })
    if (res.code === 0) {
      toast.add({ title: '保存成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    savingAuth.value = false
  }
}

function resetAuthPage() {
  unauthHTML.value = ''
  toast.add({ title: '已恢复为系统默认页面，请保存', color: 'info' })
}

async function uploadIcon(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  const formData = new FormData()
  formData.append('file', input.files[0])
  try {
    const res = await upload('/api/settings/upload-icon', formData)
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
  const formData = new FormData()
  formData.append('file', input.files[0])
  try {
    const res = await upload('/api/settings/upload-favicon', formData)
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
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">未授权页面</span>
          <div class="flex items-center gap-2">
            <UButton variant="outline" color="neutral" size="xs" @click="showPreview = !showPreview">
              {{ showPreview ? '编辑' : '预览' }}
            </UButton>
            <UButton variant="outline" color="neutral" size="xs" @click="resetAuthPage">
              恢复默认
            </UButton>
          </div>
        </div>
      </template>

      <div class="space-y-3">
        <UAlert
          title="说明"
          description="此页面在注入授权SDK后，程序未通过授权校验时展示给用户。留空则使用系统默认页面。支持完整HTML，包含表单字段 _sys_action=activate、_sys_lk（授权码）、_sys_pid（程序ID）、_sys_url（服务器地址）"
          color="info"
          variant="subtle"
        />

        <template v-if="showPreview">
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <iframe
              v-if="unauthHTML"
              :srcdoc="unauthHTML"
              class="w-full h-96 bg-white"
              sandbox="allow-forms"
            />
            <div v-else class="h-96 flex items-center justify-center text-sm text-gray-400 bg-gray-50 dark:bg-gray-800">
              当前使用系统默认页面
            </div>
          </div>
        </template>
        <template v-else>
          <UTextarea
            v-model="unauthHTML"
            placeholder="留空使用系统默认页面，或输入自定义HTML..."
            :rows="14"
            class="font-mono text-xs"
          />
        </template>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton :loading="savingAuth" @click="handleSaveAuthPage">保存未授权页面</UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
