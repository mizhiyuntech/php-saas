<script setup lang="ts">
const { get } = useApi()

useHead({ title: '订单查询' })

const loading = ref(false)
const orders = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const totalAmount = ref(0)

const filterOrderNo = ref('')
const filterMethod = ref('')
const filterStatus = ref('')
const filterStartDate = ref('')
const filterEndDate = ref('')

const methodLabels: Record<string, string> = {
  wechat: '微信支付',
  alipay: '支付宝',
  epay: '易支付'
}

const statusLabels: Record<number, string> = {
  0: '待支付',
  1: '已支付',
  2: '支付失败',
  3: '已退款'
}

const statusColors: Record<number, string> = {
  0: 'warning',
  1: 'success',
  2: 'error',
  3: 'info'
}

const methodOptions = [
  { label: '全部方式', value: '' },
  { label: '微信支付', value: 'wechat' },
  { label: '支付宝', value: 'alipay' },
  { label: '易支付', value: 'epay' }
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待支付', value: '0' },
  { label: '已支付', value: '1' },
  { label: '支付失败', value: '2' },
  { label: '已退款', value: '3' }
]

async function fetchOrders() {
  loading.value = true
  try {
    const res = await get('/api/finance/orders', {
      page: page.value,
      order_no: filterOrderNo.value,
      payment_method: filterMethod.value,
      payment_status: filterStatus.value,
      start_date: filterStartDate.value,
      end_date: filterEndDate.value
    })
    if (res.code === 0) {
      orders.value = res.data.list || []
      total.value = res.data.total
      totalAmount.value = res.data.total_amount
    }
  } finally {
    loading.value = false
  }
}

const columns = [
  { accessorKey: 'id', header: 'ID' },
  { accessorKey: 'order_no', header: '订单号' },
  { accessorKey: 'program', header: '程序' },
  { accessorKey: 'amount', header: '金额' },
  { accessorKey: 'payment_method', header: '支付方式' },
  { accessorKey: 'payment_status', header: '状态' },
  { accessorKey: 'trade_no', header: '交易号' },
  { accessorKey: 'paid_at', header: '支付时间' },
  { accessorKey: 'created_at', header: '创建时间' }
]

onMounted(fetchOrders)
</script>

<template>
  <div class="space-y-4">
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <UCard>
        <div class="text-center">
          <p class="text-xs text-gray-500">累计收入</p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">{{ totalAmount.toFixed(2) }}</p>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between flex-wrap gap-2">
          <span class="font-medium text-gray-900 dark:text-white">订单查询</span>
          <div class="flex items-center gap-2 flex-wrap">
            <UInput v-model="filterOrderNo" placeholder="订单号" class="w-36" />
            <USelect v-model="filterMethod" :items="methodOptions" value-key="value" class="w-28" />
            <USelect v-model="filterStatus" :items="statusOptions" value-key="value" class="w-28" />
            <UInput v-model="filterStartDate" type="date" class="w-36" />
            <UInput v-model="filterEndDate" type="date" class="w-36" />
            <UButton variant="outline" color="neutral" @click="fetchOrders">查询</UButton>
          </div>
        </div>
      </template>

      <UTable :data="orders" :columns="columns" :loading="loading">
        <template #program-cell="{ row }">
          <span>{{ row.original.program?.name || '-' }}</span>
        </template>
        <template #amount-cell="{ row }">
          <span class="font-medium">{{ Number(row.original.amount).toFixed(2) }}</span>
        </template>
        <template #payment_method-cell="{ row }">
          <span>{{ methodLabels[row.original.payment_method] || row.original.payment_method }}</span>
        </template>
        <template #payment_status-cell="{ row }">
          <UBadge :color="(statusColors[row.original.payment_status] as any) || 'neutral'" variant="subtle">
            {{ statusLabels[row.original.payment_status] || '未知' }}
          </UBadge>
        </template>
        <template #paid_at-cell="{ row }">
          <span class="text-sm">{{ row.original.paid_at?.substring(0, 19).replace('T', ' ') || '-' }}</span>
        </template>
        <template #created_at-cell="{ row }">
          <span class="text-sm">{{ row.original.created_at?.substring(0, 19).replace('T', ' ') }}</span>
        </template>
      </UTable>

      <template #footer>
        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-500">共 {{ total }} 条</span>
          <UPagination
            v-model="page"
            :total="total"
            :items-per-page="20"
            @update:model-value="fetchOrders"
          />
        </div>
      </template>
    </UCard>
  </div>
</template>
