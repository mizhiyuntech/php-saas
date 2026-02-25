<script setup lang="ts">
const { get, put } = useApi()
const toast = useToast()

useHead({ title: '授权验证页面' })

const loading = ref(true)
const saving = ref(false)
const unauthHTML = ref('')
const showPreview = ref(false)

async function fetchPage() {
  loading.value = true
  try {
    const res = await get('/api/settings/unauth-page')
    if (res.code === 0 && res.data) {
      unauthHTML.value = res.data.html || ''
    }
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const res = await put('/api/settings/unauth-page', { html: unauthHTML.value })
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
  unauthHTML.value = ''
  toast.add({ title: '已恢复为系统默认页面，请保存', color: 'info' })
}

onMounted(fetchPage)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">授权验证页面</span>
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
          description="此页面在注入授权SDK后，程序未通过授权校验时展示给用户，用户需要在此页面输入授权码进行激活。留空则使用系统默认页面。表单必须包含字段：_sys_action=activate、_sys_lk（授权码）、_sys_pid（程序ID）、_sys_url（服务器地址）"
          color="info"
          variant="subtle"
        />

        <template v-if="showPreview">
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <iframe
              v-if="unauthHTML"
              :srcdoc="unauthHTML"
              class="w-full h-[500px] bg-white"
              sandbox="allow-forms"
            />
            <div v-else class="h-[500px] flex items-center justify-center text-sm text-gray-400 bg-gray-50 dark:bg-gray-800">
              当前使用系统默认页面
            </div>
          </div>
        </template>
        <template v-else>
          <UTextarea
            v-model="unauthHTML"
            placeholder="留空使用系统默认页面，或输入自定义HTML..."
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
