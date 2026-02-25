<script setup lang="ts">
const { get } = useApi()

useHead({ title: '仪表盘' })

const stats = ref({
  program_count: 0,
  license_count: 0,
  active_license_count: 0,
  order_count: 0,
  total_income: 0
})

const orderStats = ref({
  today: 0,
  week: 0,
  month: 0,
  total: 0
})

const loading = ref(true)

onMounted(async () => {
  try {
    const [statsRes, orderRes] = await Promise.all([
      get('/api/dashboard/stats'),
      get('/api/finance/order-stats')
    ])
    if (statsRes.code === 0) stats.value = statsRes.data
    if (orderRes.code === 0) orderStats.value = orderRes.data
  } finally {
    loading.value = false
  }
})

const statCards = computed(() => [
  { label: '程序总数', value: stats.value.program_count, icon: 'i-lucide-box', color: 'text-blue-600' },
  { label: '授权码总数', value: stats.value.license_count, icon: 'i-lucide-key-round', color: 'text-green-600' },
  { label: '已激活', value: stats.value.active_license_count, icon: 'i-lucide-check-circle', color: 'text-orange-600' },
  { label: '订单总数', value: stats.value.order_count, icon: 'i-lucide-receipt', color: 'text-purple-600' }
])

const incomeCards = computed(() => [
  { label: '今日收入', value: orderStats.value.today },
  { label: '本周收入', value: orderStats.value.week },
  { label: '本月收入', value: orderStats.value.month },
  { label: '总收入', value: orderStats.value.total }
])
</script>

<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <UCard v-for="item in statCards" :key="item.label">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
            <UIcon :name="item.icon" :class="[item.color, 'w-5 h-5']" />
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ item.label }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ item.value }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">收入统计</span>
      </template>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="item in incomeCards" :key="item.label" class="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <p class="text-xs text-gray-500 mb-1">{{ item.label }}</p>
          <p class="text-lg font-bold text-gray-900 dark:text-white">{{ item.value.toFixed(2) }}</p>
        </div>
      </div>
    </UCard>
  </div>
</template>
