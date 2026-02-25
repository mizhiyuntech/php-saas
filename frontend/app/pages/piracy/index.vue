<script setup lang="ts">
const { get, post, put, del } = useApi()
const toast = useToast()

useHead({ title: '盗版管理' })

const loading = ref(false)
const records = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')

const showForm = ref(false)
const editMode = ref(false)
const formLoading = ref(false)
const formData = reactive({
  id: 0,
  domain: '',
  ip: '',
  message: '',
  status: 1
})

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 }
]

async function fetchRecords() {
  loading.value = true
  try {
    const res = await get('/api/piracy', { page: page.value, keyword: keyword.value })
    if (res.code === 0) {
      records.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editMode.value = false
  Object.assign(formData, { id: 0, domain: '', ip: '', message: '', status: 1 })
  showForm.value = true
}

function openEdit(row: any) {
  editMode.value = true
  Object.assign(formData, row)
  showForm.value = true
}

async function handleSubmit() {
  if (!formData.domain) {
    toast.add({ title: '请输入盗版域名', color: 'error' })
    return
  }
  if (!formData.message) {
    toast.add({ title: '请输入盗版提示信息', color: 'error' })
    return
  }

  formLoading.value = true
  try {
    const body = {
      domain: formData.domain,
      ip: formData.ip,
      message: formData.message,
      status: formData.status
    }

    let res
    if (editMode.value) {
      res = await put(`/api/piracy/${formData.id}`, body)
    } else {
      res = await post('/api/piracy', body)
    }

    if (res.code === 0) {
      toast.add({ title: editMode.value ? '修改成功' : '添加成功', color: 'success' })
      showForm.value = false
      fetchRecords()
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(row: any) {
  const res = await del(`/api/piracy/${row.id}`)
  if (res.code === 0) {
    toast.add({ title: '删除成功', color: 'success' })
    fetchRecords()
  }
}

const columns = [
  { accessorKey: 'id', header: 'ID' },
  { accessorKey: 'domain', header: '盗版域名' },
  { accessorKey: 'ip', header: '服务器IP' },
  { accessorKey: 'message', header: '盗版提示信息' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'created_at', header: '添加时间' },
  { accessorKey: 'actions', header: '操作' }
]

onMounted(fetchRecords)
</script>

<template>
  <div class="space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between flex-wrap gap-2">
          <span class="font-medium text-gray-900 dark:text-white">盗版管理</span>
          <div class="flex items-center gap-2">
            <UInput v-model="keyword" placeholder="搜索域名/IP/提示" class="w-44" @keyup.enter="fetchRecords" />
            <UButton variant="outline" color="neutral" @click="fetchRecords">搜索</UButton>
            <UButton @click="openAdd">添加盗版记录</UButton>
          </div>
        </div>
      </template>

      <UAlert
        title="说明"
        description="添加盗版域名或服务器IP后，当被注入授权SDK的程序在该域名/IP下运行时，将实时显示盗版提示页面，阻止程序正常使用。提示信息支持实时修改，无需重新注入。"
        color="info"
        variant="subtle"
        class="mb-4"
      />

      <UTable :data="records" :columns="columns" :loading="loading">
        <template #domain-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded">{{ row.original.domain }}</code>
        </template>
        <template #ip-cell="{ row }">
          <span class="text-sm">{{ row.original.ip || '-' }}</span>
        </template>
        <template #message-cell="{ row }">
          <span class="text-sm text-red-600 dark:text-red-400">{{ row.original.message }}</span>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 1 ? 'success' : 'neutral'" variant="subtle">
            {{ row.original.status === 1 ? '启用' : '禁用' }}
          </UBadge>
        </template>
        <template #created_at-cell="{ row }">
          <span class="text-sm">{{ row.original.created_at?.substring(0, 19).replace('T', ' ') }}</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UButton size="xs" variant="ghost" color="neutral" @click="openEdit(row.original)">编辑</UButton>
            <UButton size="xs" variant="ghost" color="error" @click="handleDelete(row.original)">删除</UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-500">共 {{ total }} 条</span>
          <UPagination v-model="page" :total="total" :items-per-page="20" @update:model-value="fetchRecords" />
        </div>
      </template>
    </UCard>

    <UModal v-model:open="showForm">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ editMode ? '编辑盗版记录' : '添加盗版记录' }}</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showForm = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="盗版域名" required hint="不含 http:// 前缀">
              <UInput v-model="formData.domain" placeholder="例如: pirate-site.com" />
            </UFormField>
            <UFormField label="服务器IP" hint="可选，填写更精确">
              <UInput v-model="formData.ip" placeholder="例如: 1.2.3.4" />
            </UFormField>
            <UFormField label="盗版提示信息" required>
              <UTextarea v-model="formData.message" placeholder="例如: 此程序为盗版，请停止使用" :rows="3" />
            </UFormField>
            <UFormField label="状态">
              <USelect v-model="formData.status" :items="statusOptions" value-key="value" />
            </UFormField>
          </div>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton variant="outline" color="neutral" @click="showForm = false">取消</UButton>
              <UButton :loading="formLoading" @click="handleSubmit">确定</UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
