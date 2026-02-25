<script setup lang="ts">
const { get, post, put, del } = useApi()
const toast = useToast()

useHead({ title: '程序管理' })

const loading = ref(false)
const programs = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const searchName = ref('')

const showForm = ref(false)
const editMode = ref(false)
const formLoading = ref(false)
const formData = reactive({
  id: 0,
  name: '',
  description: '',
  version: '',
  language: 'php',
  status: 1
})

const languageOptions = [
  { label: 'PHP', value: 'php' },
  { label: 'Python', value: 'python' },
  { label: 'Node.js', value: 'nodejs' },
  { label: 'Go', value: 'go' },
  { label: 'Java', value: 'java' },
  { label: 'C#', value: 'csharp' }
]

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 }
]

async function fetchPrograms() {
  loading.value = true
  try {
    const res = await get('/api/programs', { page: page.value, name: searchName.value })
    if (res.code === 0) {
      programs.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editMode.value = false
  Object.assign(formData, { id: 0, name: '', description: '', version: '', language: 'php', status: 1 })
  showForm.value = true
}

function openEdit(row: any) {
  editMode.value = true
  Object.assign(formData, row)
  showForm.value = true
}

async function handleSubmit() {
  if (!formData.name) {
    toast.add({ title: '请输入程序名称', color: 'error' })
    return
  }

  formLoading.value = true
  try {
    const body = {
      name: formData.name,
      description: formData.description,
      version: formData.version,
      language: formData.language,
      status: formData.status
    }

    let res
    if (editMode.value) {
      res = await put(`/api/programs/${formData.id}`, body)
    } else {
      res = await post('/api/programs', body)
    }

    if (res.code === 0) {
      toast.add({ title: editMode.value ? '修改成功' : '创建成功', color: 'success' })
      showForm.value = false
      fetchPrograms()
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(row: any) {
  const res = await del(`/api/programs/${row.id}`)
  if (res.code === 0) {
    toast.add({ title: '删除成功', color: 'success' })
    fetchPrograms()
  } else {
    toast.add({ title: res.message, color: 'error' })
  }
}

const columns = [
  { accessorKey: 'id', header: 'ID' },
  { accessorKey: 'name', header: '程序名称' },
  { accessorKey: 'language', header: '开发语言' },
  { accessorKey: 'version', header: '版本' },
  { accessorKey: 'secret_key', header: '密钥' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'created_at', header: '创建时间' },
  { accessorKey: 'actions', header: '操作' }
]

onMounted(fetchPrograms)
</script>

<template>
  <div class="space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">程序管理</span>
          <div class="flex items-center gap-2">
            <UInput v-model="searchName" placeholder="搜索程序名称" @keyup.enter="fetchPrograms" />
            <UButton variant="outline" color="neutral" @click="fetchPrograms">搜索</UButton>
            <UButton @click="openAdd">添加程序</UButton>
          </div>
        </div>
      </template>

      <UTable :data="programs" :columns="columns" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 1 ? 'success' : 'error'" variant="subtle">
            {{ row.original.status === 1 ? '启用' : '禁用' }}
          </UBadge>
        </template>
        <template #secret_key-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">{{ row.original.secret_key?.substring(0, 12) }}...</code>
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
        <div class="flex justify-end">
          <UPagination
            v-model="page"
            :total="total"
            :items-per-page="20"
            @update:model-value="fetchPrograms"
          />
        </div>
      </template>
    </UCard>

    <UModal v-model:open="showForm">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ editMode ? '编辑程序' : '添加程序' }}</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showForm = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="程序名称" required>
              <UInput v-model="formData.name" placeholder="请输入程序名称" />
            </UFormField>
            <UFormField label="描述">
              <UTextarea v-model="formData.description" placeholder="请输入程序描述" />
            </UFormField>
            <div class="grid grid-cols-2 gap-4">
              <UFormField label="版本号">
                <UInput v-model="formData.version" placeholder="如 1.0.0" />
              </UFormField>
              <UFormField label="开发语言">
                <USelect v-model="formData.language" :items="languageOptions" value-key="value" />
              </UFormField>
            </div>
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
