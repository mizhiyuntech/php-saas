<script setup lang="ts">
const { get, post, put, del } = useApi()
const toast = useToast()

useHead({ title: '套餐管理' })

const loading = ref(false)
const packages = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const programs = ref<any[]>([])
const filterProgram = ref('all')

const showForm = ref(false)
const editMode = ref(false)
const formLoading = ref(false)
const formData = reactive({
  id: 0,
  program_id: null as number | null,
  name: '',
  description: '',
  duration: 30,
  price: 0,
  sort_order: 0,
  status: 1
})

const statusOptions = [
  { label: '上架', value: 1 },
  { label: '下架', value: 0 }
]

async function fetchPrograms() {
  const res = await get('/api/programs/all')
  if (res.code === 0) {
    programs.value = (res.data || []).map((p: any) => ({ label: p.name, value: p.id }))
  }
}

async function fetchPackages() {
  loading.value = true
  try {
    const res = await get('/api/packages', { page: page.value, program_id: filterProgram.value === 'all' ? '' : filterProgram.value })
    if (res.code === 0) {
      packages.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editMode.value = false
  Object.assign(formData, { id: 0, program_id: null, name: '', description: '', duration: 30, price: 0, sort_order: 0, status: 1 })
  showForm.value = true
}

function openEdit(row: any) {
  editMode.value = true
  Object.assign(formData, row)
  showForm.value = true
}

async function handleSubmit() {
  if (!formData.program_id) {
    toast.add({ title: '请选择程序', color: 'error' })
    return
  }
  if (!formData.name) {
    toast.add({ title: '请输入套餐名称', color: 'error' })
    return
  }
  if (formData.price <= 0) {
    toast.add({ title: '请输入有效价格', color: 'error' })
    return
  }

  formLoading.value = true
  try {
    const body = {
      program_id: formData.program_id,
      name: formData.name,
      description: formData.description,
      duration: formData.duration,
      price: formData.price,
      sort_order: formData.sort_order,
      status: formData.status
    }

    const res = editMode.value
      ? await put(`/api/packages/${formData.id}`, body)
      : await post('/api/packages', body)

    if (res.code === 0) {
      toast.add({ title: editMode.value ? '修改成功' : '创建成功', color: 'success' })
      showForm.value = false
      fetchPackages()
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(row: any) {
  const res = await del(`/api/packages/${row.id}`)
  if (res.code === 0) {
    toast.add({ title: '删除成功', color: 'success' })
    fetchPackages()
  }
}

const columns = [
  { accessorKey: 'id', header: 'ID' },
  { accessorKey: 'program', header: '所属程序' },
  { accessorKey: 'name', header: '套餐名称' },
  { accessorKey: 'duration', header: '有效期' },
  { accessorKey: 'price', header: '价格' },
  { accessorKey: 'sort_order', header: '排序' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'actions', header: '操作' }
]

onMounted(() => {
  fetchPrograms()
  fetchPackages()
})
</script>

<template>
  <div class="space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between flex-wrap gap-2">
          <span class="font-medium text-gray-900 dark:text-white">套餐管理</span>
          <div class="flex items-center gap-2">
            <USelect v-model="filterProgram" :items="[{ label: '全部程序', value: 'all' }, ...programs]" value-key="value" class="w-32" />
            <UButton variant="outline" color="neutral" @click="fetchPackages">筛选</UButton>
            <UButton @click="openAdd">添加套餐</UButton>
          </div>
        </div>
      </template>

      <UTable :data="packages" :columns="columns" :loading="loading">
        <template #program-cell="{ row }">
          <span>{{ row.original.program?.name || '-' }}</span>
        </template>
        <template #duration-cell="{ row }">
          <span>{{ row.original.duration > 0 ? row.original.duration + '天' : '永久' }}</span>
        </template>
        <template #price-cell="{ row }">
          <span class="font-medium text-orange-600">{{ Number(row.original.price).toFixed(2) }} 元</span>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 1 ? 'success' : 'neutral'" variant="subtle">
            {{ row.original.status === 1 ? '上架' : '下架' }}
          </UBadge>
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
          <UPagination v-model="page" :total="total" :items-per-page="20" @update:model-value="fetchPackages" />
        </div>
      </template>
    </UCard>

    <UModal v-model:open="showForm">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ editMode ? '编辑套餐' : '添加套餐' }}</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showForm = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="所属程序" required>
              <USelect v-model="formData.program_id" :items="programs" value-key="value" placeholder="请选择程序" />
            </UFormField>
            <UFormField label="套餐名称" required>
              <UInput v-model="formData.name" placeholder="如：月度套餐、年度套餐" />
            </UFormField>
            <UFormField label="套餐描述">
              <UInput v-model="formData.description" placeholder="套餐说明" />
            </UFormField>
            <div class="grid grid-cols-2 gap-4">
              <UFormField label="有效期（天）" hint="0为永久">
                <UInput v-model.number="formData.duration" type="number" :min="0" />
              </UFormField>
              <UFormField label="价格（元）" required>
                <UInput v-model.number="formData.price" type="number" :min="0" step="0.01" />
              </UFormField>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <UFormField label="排序" hint="越小越靠前">
                <UInput v-model.number="formData.sort_order" type="number" :min="0" />
              </UFormField>
              <UFormField label="状态">
                <USelect v-model="formData.status" :items="statusOptions" value-key="value" />
              </UFormField>
            </div>
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
