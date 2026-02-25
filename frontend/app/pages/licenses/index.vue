<script setup lang="ts">
const { get, post, put, del } = useApi()
const toast = useToast()

useHead({ title: '授权管理' })

const loading = ref(false)
const licenses = ref<any[]>([])
const total = ref(0)
const page = ref(1)

const filterProgramId = ref('all')
const filterStatus = ref('all')
const filterKeyword = ref('')

const programs = ref<any[]>([])

const showCreate = ref(false)
const createLoading = ref(false)
const createForm = reactive({
  program_id: null as number | null,
  count: 1,
  duration: 0,
  remark: ''
})

const showEdit = ref(false)
const editLoading = ref(false)
const editForm = reactive({
  id: 0,
  status: 0,
  duration: 0,
  remark: ''
})

const statusLabels: Record<number, string> = {
  0: '未使用',
  1: '已激活',
  2: '已过期',
  3: '已禁用'
}

const statusColors: Record<number, string> = {
  0: 'info',
  1: 'success',
  2: 'warning',
  3: 'error'
}

const statusOptions = [
  { label: '未使用', value: 0 },
  { label: '已激活', value: 1 },
  { label: '已过期', value: 2 },
  { label: '已禁用', value: 3 }
]

async function fetchPrograms() {
  const res = await get('/api/programs/all')
  if (res.code === 0) {
    programs.value = (res.data || []).map((p: any) => ({ label: p.name, value: p.id }))
  }
}

async function fetchLicenses() {
  loading.value = true
  try {
    const res = await get('/api/licenses', {
      page: page.value,
      program_id: filterProgramId.value === 'all' ? '' : filterProgramId.value,
      status: filterStatus.value === 'all' ? '' : filterStatus.value,
      keyword: filterKeyword.value
    })
    if (res.code === 0) {
      licenses.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(createForm, { program_id: null, count: 1, duration: 0, remark: '' })
  showCreate.value = true
}

function openEdit(row: any) {
  Object.assign(editForm, { id: row.id, status: row.status, duration: row.duration, remark: row.remark })
  showEdit.value = true
}

async function handleCreate() {
  if (!createForm.program_id) {
    toast.add({ title: '请选择程序', color: 'error' })
    return
  }
  createLoading.value = true
  try {
    const res = await post('/api/licenses', createForm)
    if (res.code === 0) {
      toast.add({ title: `成功生成 ${res.data.count} 个授权码`, color: 'success' })
      showCreate.value = false
      fetchLicenses()
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    createLoading.value = false
  }
}

async function handleEdit() {
  editLoading.value = true
  try {
    const res = await put(`/api/licenses/${editForm.id}`, {
      status: editForm.status,
      duration: editForm.duration,
      remark: editForm.remark
    })
    if (res.code === 0) {
      toast.add({ title: '修改成功', color: 'success' })
      showEdit.value = false
      fetchLicenses()
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    editLoading.value = false
  }
}

async function handleDelete(row: any) {
  const res = await del(`/api/licenses/${row.id}`)
  if (res.code === 0) {
    toast.add({ title: '删除成功', color: 'success' })
    fetchLicenses()
  }
}

function copyKey(key: string) {
  navigator.clipboard.writeText(key)
  toast.add({ title: '已复制到剪贴板', color: 'success' })
}

const columns = [
  { accessorKey: 'id', header: 'ID' },
  { accessorKey: 'license_key', header: '授权码' },
  { accessorKey: 'program', header: '所属程序' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'duration', header: '有效期(天)' },
  { accessorKey: 'activated_at', header: '激活时间' },
  { accessorKey: 'expires_at', header: '过期时间' },
  { accessorKey: 'created_at', header: '创建时间' },
  { accessorKey: 'actions', header: '操作' }
]

onMounted(() => {
  fetchPrograms()
  fetchLicenses()
})
</script>

<template>
  <div class="space-y-4">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between flex-wrap gap-2">
          <span class="font-medium text-gray-900 dark:text-white">授权码管理</span>
          <div class="flex items-center gap-2 flex-wrap">
            <UInput v-model="filterKeyword" placeholder="搜索授权码/备注" class="w-40" />
            <USelect v-model="filterProgramId" :items="[{ label: '全部程序', value: 'all' }, ...programs]" value-key="value" class="w-32" />
            <USelect v-model="filterStatus" :items="[{ label: '全部状态', value: 'all' }, ...statusOptions]" value-key="value" class="w-28" />
            <UButton variant="outline" color="neutral" @click="fetchLicenses">搜索</UButton>
            <UButton @click="openCreate">生成授权码</UButton>
          </div>
        </div>
      </template>

      <UTable :data="licenses" :columns="columns" :loading="loading">
        <template #license_key-cell="{ row }">
          <div class="flex items-center gap-1">
            <code class="text-xs bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded">{{ row.original.license_key }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyKey(row.original.license_key)" />
          </div>
        </template>
        <template #program-cell="{ row }">
          <span>{{ row.original.program?.name || '-' }}</span>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="(statusColors[row.original.status] as any) || 'neutral'" variant="subtle">
            {{ statusLabels[row.original.status] || '未知' }}
          </UBadge>
        </template>
        <template #duration-cell="{ row }">
          <span>{{ row.original.duration > 0 ? row.original.duration + '天' : '永久' }}</span>
        </template>
        <template #activated_at-cell="{ row }">
          <span class="text-sm">{{ row.original.activated_at?.substring(0, 19).replace('T', ' ') || '-' }}</span>
        </template>
        <template #expires_at-cell="{ row }">
          <span class="text-sm">{{ row.original.expires_at?.substring(0, 19).replace('T', ' ') || '-' }}</span>
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
          <UPagination
            v-model="page"
            :total="total"
            :items-per-page="20"
            @update:model-value="fetchLicenses"
          />
        </div>
      </template>
    </UCard>

    <!-- Create Modal -->
    <UModal v-model:open="showCreate">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">生成授权码</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showCreate = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="选择程序" required>
              <USelect v-model="createForm.program_id" :items="programs" value-key="value" placeholder="请选择程序" />
            </UFormField>
            <div class="grid grid-cols-2 gap-4">
              <UFormField label="生成数量">
                <UInput v-model.number="createForm.count" type="number" :min="1" :max="100" />
              </UFormField>
              <UFormField label="有效期（天）" hint="0为永久">
                <UInput v-model.number="createForm.duration" type="number" :min="0" />
              </UFormField>
            </div>
            <UFormField label="备注">
              <UInput v-model="createForm.remark" placeholder="可选备注" />
            </UFormField>
          </div>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton variant="outline" color="neutral" @click="showCreate = false">取消</UButton>
              <UButton :loading="createLoading" @click="handleCreate">生成</UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>

    <!-- Edit Modal -->
    <UModal v-model:open="showEdit">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">编辑授权码</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showEdit = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="状态">
              <USelect v-model="editForm.status" :items="statusOptions" value-key="value" />
            </UFormField>
            <UFormField label="有效期（天）" hint="0为永久">
              <UInput v-model.number="editForm.duration" type="number" :min="0" />
            </UFormField>
            <UFormField label="备注">
              <UInput v-model="editForm.remark" placeholder="可选备注" />
            </UFormField>
          </div>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton variant="outline" color="neutral" @click="showEdit = false">取消</UButton>
              <UButton :loading="editLoading" @click="handleEdit">保存</UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
