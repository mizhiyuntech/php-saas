<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase as string || ''
const settingsStore = useSettingsStore()
const toast = useToast()

useHead({ title: '订单支付' })

const loading = ref(true)
const payInfo = ref<any>(null)
const polling = ref(true)
let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await settingsStore.fetchSiteInfo()
  await fetchPayInfo()
  startPolling()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function fetchPayInfo() {
  const orderNo = route.query.order_no as string
  if (!orderNo) { loading.value = false; return }
  try {
    const res = await fetch(`${baseURL}/api/public/pay-info/${orderNo}`)
    const data = await res.json()
    if (data.code === 0) {
      payInfo.value = data.data
      if (data.data.status === 'paid') {
        router.replace(`/purchase/result?order_no=${orderNo}`)
      }
    }
  } finally {
    loading.value = false
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    await fetchPayInfo()
    if (payInfo.value?.status === 'paid') {
      polling.value = false
      if (pollTimer) clearInterval(pollTimer)
      router.replace(`/purchase/result?order_no=${route.query.order_no}`)
    }
  }, 3000)
}

function goResult() {
  router.push(`/purchase/result?order_no=${route.query.order_no}`)
}

const qrImageUrl = computed(() => {
  if (!payInfo.value?.qr_code) return ''
  return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(payInfo.value.qr_code)}`
})

const methodLabels: Record<string, string> = {
  wechat: '微信支付',
  alipay: '支付宝',
  epay: '易支付'
}
</script>

<template>
  <div class="w-full max-w-md mx-auto">
    <div class="text-center mb-4">
      <h1 class="text-xl font-bold text-gray-900 dark:text-white">{{ settingsStore.siteTitle }}</h1>
    </div>

    <UCard v-if="loading">
      <div class="text-center py-12 text-gray-400">加载中...</div>
    </UCard>

    <UCard v-else-if="!payInfo">
      <div class="text-center py-8">
        <UIcon name="i-lucide-alert-circle" class="w-12 h-12 text-gray-400 mx-auto mb-2" />
        <p class="text-gray-500">订单不存在或已失效</p>
        <NuxtLink to="/purchase" class="mt-3 inline-block">
          <UButton variant="outline">返回购买页</UButton>
        </NuxtLink>
      </div>
    </UCard>

    <template v-else>
      <UCard>
        <template #header>
          <div class="text-center">
            <span class="font-medium text-gray-900 dark:text-white">{{ methodLabels[payInfo.payment_method] || '在线支付' }}</span>
          </div>
        </template>

        <div class="space-y-5">
          <!-- Order info -->
          <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500">订单号</span>
              <span class="font-mono text-gray-700 dark:text-gray-300">{{ payInfo.order_no }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">程序</span>
              <span class="text-gray-700 dark:text-gray-300">{{ payInfo.program_name }}</span>
            </div>
            <div class="flex justify-between border-t border-gray-200 dark:border-gray-700 pt-2">
              <span class="font-medium text-gray-700 dark:text-gray-300">应付金额</span>
              <span class="text-2xl font-bold text-orange-600">{{ Number(payInfo.amount).toFixed(2) }} 元</span>
            </div>
          </div>

          <!-- QR Code payment (WeChat / Alipay scan) -->
          <div class="text-center space-y-4">
            <div
              class="p-6 rounded-lg"
              :class="payInfo.payment_method === 'wechat' ? 'bg-green-50 dark:bg-green-950' : 'bg-blue-50 dark:bg-blue-950'"
            >
              <UIcon
                :name="payInfo.payment_method === 'wechat' ? 'i-lucide-smartphone' : 'i-lucide-wallet'"
                class="w-10 h-10 mx-auto mb-3"
                :class="payInfo.payment_method === 'wechat' ? 'text-green-600' : 'text-blue-600'"
              />
              <p class="text-sm font-medium mb-3" :class="payInfo.payment_method === 'wechat' ? 'text-green-800 dark:text-green-300' : 'text-blue-800 dark:text-blue-300'">
                {{ payInfo.payment_method === 'wechat' ? '请使用微信扫描二维码支付' : '请使用支付宝扫描二维码支付' }}
              </p>

              <!-- QR code from server -->
              <div v-if="qrImageUrl" class="w-52 h-52 mx-auto bg-white rounded-lg p-2">
                <img :src="qrImageUrl" alt="支付二维码" class="w-full h-full object-contain" />
              </div>

              <!-- No QR code available -->
              <div v-else class="w-52 h-52 mx-auto bg-white rounded-lg border-2 flex items-center justify-center" :class="payInfo.payment_method === 'wechat' ? 'border-green-200' : 'border-blue-200'">
                <div class="text-center px-4">
                  <UIcon name="i-lucide-alert-circle" class="w-8 h-8 text-gray-300 mx-auto mb-2" />
                  <p class="text-xs text-gray-400">{{ payInfo.payment_method === 'wechat' ? '请在后台正确配置微信支付V3密钥' : '请在后台正确配置支付宝应用密钥' }}</p>
                </div>
              </div>
            </div>

            <!-- Alipay web pay link -->
            <a v-if="payInfo.payment_method === 'alipay' && payInfo.return_url" :href="payInfo.return_url">
              <UButton variant="outline" block class="mt-2">使用支付宝网页支付</UButton>
            </a>
          </div>

          <!-- Polling indicator -->
          <div class="flex items-center justify-center gap-2 text-sm text-gray-500">
            <div v-if="polling" class="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
            <span>{{ polling ? '正在等待支付结果...' : '支付状态检测已停止' }}</span>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-between">
            <NuxtLink to="/purchase">
              <UButton variant="ghost" color="neutral">取消支付</UButton>
            </NuxtLink>
            <UButton variant="outline" @click="goResult">我已完成支付</UButton>
          </div>
        </template>
      </UCard>
    </template>
  </div>
</template>
