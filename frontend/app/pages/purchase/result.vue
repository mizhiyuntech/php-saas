<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const route = useRoute()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase as string || ''
const settingsStore = useSettingsStore()
const toast = useToast()

useHead({ title: '支付结果' })

const loading = ref(true)
const order = ref<any>(null)
const polling = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await settingsStore.fetchSiteInfo()
  await fetchOrder()
  if (order.value && order.value.payment_status === 0) {
    startPolling()
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function fetchOrder() {
  const orderNo = route.query.order_no as string
  if (!orderNo) {
    loading.value = false
    return
  }

  try {
    const res = await fetch(`${baseURL}/api/public/order/${orderNo}`)
    const data = await res.json()
    if (data.code === 0) {
      order.value = data.data
      if (data.data.payment_status === 1 && pollTimer) {
        clearInterval(pollTimer)
      }
    }
  } finally {
    loading.value = false
  }
}

function startPolling() {
  polling.value = true
  pollTimer = setInterval(async () => {
    await fetchOrder()
    if (order.value?.payment_status === 1) {
      polling.value = false
      if (pollTimer) clearInterval(pollTimer)
    }
  }, 3000)
}

function copyKey(key: string) {
  navigator.clipboard.writeText(key)
  toast.add({ title: '授权码已复制', color: 'success' })
}
</script>

<template>
  <div class="w-full max-w-md mx-auto">
    <div class="text-center mb-4">
      <h1 class="text-xl font-bold text-gray-900 dark:text-white">{{ settingsStore.siteTitle }}</h1>
    </div>

    <UCard v-if="loading">
      <div class="text-center py-8 text-gray-400">加载中...</div>
    </UCard>

    <UCard v-else-if="!order">
      <div class="text-center py-8">
        <UIcon name="i-lucide-alert-circle" class="w-12 h-12 text-gray-400 mx-auto mb-2" />
        <p class="text-gray-500">订单不存在</p>
        <NuxtLink to="/purchase" class="mt-3 inline-block">
          <UButton variant="outline">返回购买页</UButton>
        </NuxtLink>
      </div>
    </UCard>

    <!-- Payment Success -->
    <UCard v-else-if="order.payment_status === 1">
      <div class="text-center py-4 space-y-4">
        <div class="w-16 h-16 mx-auto rounded-full bg-green-50 dark:bg-green-950 flex items-center justify-center">
          <UIcon name="i-lucide-check" class="w-8 h-8 text-green-500" />
        </div>
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">支付成功</h2>

        <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2 text-left">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">订单号</span>
            <span class="text-gray-700 dark:text-gray-300">{{ order.order_no }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">程序</span>
            <span class="text-gray-700 dark:text-gray-300">{{ order.program_name }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">金额</span>
            <span class="text-gray-700 dark:text-gray-300">{{ Number(order.amount).toFixed(2) }} 元</span>
          </div>
        </div>

        <div v-if="order.license_key" class="p-4 bg-blue-50 dark:bg-blue-950 rounded-lg">
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">您的授权码</p>
          <div class="flex items-center justify-center gap-2">
            <code class="text-lg font-bold text-primary-600 dark:text-primary-400">{{ order.license_key }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyKey(order.license_key)" />
          </div>
          <p class="text-xs text-gray-400 mt-2">请妥善保存您的授权码</p>
        </div>

        <NuxtLink to="/purchase">
          <UButton variant="outline">继续购买</UButton>
        </NuxtLink>
      </div>
    </UCard>

    <!-- Waiting for payment -->
    <UCard v-else>
      <div class="text-center py-6 space-y-4">
        <div class="w-16 h-16 mx-auto rounded-full bg-orange-50 dark:bg-orange-950 flex items-center justify-center">
          <UIcon name="i-lucide-clock" class="w-8 h-8 text-orange-500" :class="polling ? 'animate-pulse' : ''" />
        </div>
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">等待支付</h2>
        <p class="text-sm text-gray-500">订单号：{{ order.order_no }}</p>
        <p class="text-sm text-gray-400">{{ polling ? '正在等待支付结果...' : '支付完成后刷新页面查看结果' }}</p>
        <UButton variant="outline" @click="fetchOrder">刷新状态</UButton>
      </div>
    </UCard>
  </div>
</template>
