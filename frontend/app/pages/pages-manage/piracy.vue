<script setup lang="ts">
const { get, put } = useApi()
const toast = useToast()

useHead({ title: '盗版提示页面' })

const loading = ref(true)
const saving = ref(false)
const piracyHTML = ref('')
const showPreview = ref(false)

async function fetchPage() {
  loading.value = true
  try {
    const res = await get('/api/settings/piracy-page')
    if (res.code === 0 && res.data) {
      piracyHTML.value = res.data.html || ''
    }
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const res = await put('/api/settings/piracy-page', { html: piracyHTML.value })
    if (res.code === 0) {
      toast.add({ title: '保存成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    saving.value = false
  }
}

function resetPage() {
  piracyHTML.value = ''
  toast.add({ title: '已恢复为系统默认页面，请保存', color: 'info' })
}

const previewHTML = computed(() => {
  if (!piracyHTML.value) return ''
  return piracyHTML.value.replace('{{MESSAGE}}', '这是一个盗版提示示例信息')
})

onMounted(fetchPage)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">盗版提示页面</span>
          <div class="flex items-center gap-2">
            <UButton variant="outline" color="neutral" size="xs" @click="showPreview = !showPreview">
              {{ showPreview ? '编辑' : '预览' }}
            </UButton>
            <UButton variant="outline" color="neutral" size="xs" @click="resetPage">恢复默认</UButton>
          </div>
        </div>
      </template>

      <div class="space-y-3">
        <UAlert
          title="说明"
          description="当检测到程序运行在盗版域名/IP上时，将展示此页面。页面中的 {{MESSAGE}} 占位符会被替换为盗版管理中填写的具体提示信息。留空则使用系统默认页面。盗版提示信息支持实时更新，修改后立即生效。"
          color="info"
          variant="subtle"
        />

        <template v-if="showPreview">
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <iframe
              v-if="piracyHTML || true"
              :srcdoc="previewHTML || '<div style=\'padding:40px;text-align:center;color:#999\'>当前使用系统默认页面</div>'"
              class="w-full h-[500px] bg-white"
              sandbox=""
            />
          </div>
        </template>
        <template v-else>
          <UTextarea
            v-model="piracyHTML"
            placeholder="留空使用系统默认页面，或输入自定义HTML...&#10;使用 {{MESSAGE}} 占位符显示盗版提示信息"
            :rows="18"
            class="font-mono text-xs"
          />
        </template>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton :loading="saving" @click="handleSave">保存页面</UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
