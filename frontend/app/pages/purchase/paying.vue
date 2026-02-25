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
  if (!orderNo) {
    loading.value = false
    return
  }
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
      const orderNo = route.query.order_no as string
      router.replace(`/purchase/result?order_no=${orderNo}`)
    }
  }, 3000)
}

function goResult() {
  const orderNo = route.query.order_no as string
  router.push(`/purchase/result?order_no=${orderNo}`)
}

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
              <span class="text-2xl font-bold text-orange-600">{{ Number(payInfo.amount).toFixed(2) }}</span>
            </div>
          </div>

          <!-- WeChat Pay -->
          <div v-if="payInfo.payment_method === 'wechat'" class="text-center space-y-4">
            <div class="p-6 bg-green-50 dark:bg-green-950 rounded-lg">
              <UIcon name="i-lucide-smartphone" class="w-12 h-12 text-green-600 mx-auto mb-3" />
              <p class="text-sm font-medium text-green-800 dark:text-green-300 mb-2">微信扫码支付</p>
              <p class="text-xs text-green-600 dark:text-green-400">请使用微信扫描下方二维码完成支付</p>
              <div class="mt-4 w-48 h-48 mx-auto bg-white rounded-lg border-2 border-green-200 flex items-center justify-center">
                <div class="text-center">
                  <UIcon name="i-lucide-scan" class="w-8 h-8 text-gray-300 mx-auto mb-2" />
                  <p class="text-xs text-gray-400">请在后台配置微信支付</p>
                  <p class="text-xs text-gray-400">JSAPI/Native接口</p>
                </div>
              </div>
            </div>
            <UAlert
              title="支付步骤"
              description="1. 打开微信，点击右上角扫一扫  2. 扫描上方二维码  3. 确认支付金额并完成付款  4. 支付成功后页面将自动跳转"
              color="info"
              variant="subtle"
            />
          </div>

          <!-- Alipay -->
          <div v-if="payInfo.payment_method === 'alipay'" class="text-center space-y-4">
            <div class="p-6 bg-blue-50 dark:bg-blue-950 rounded-lg">
              <UIcon name="i-lucide-wallet" class="w-12 h-12 text-blue-600 mx-auto mb-3" />
              <p class="text-sm font-medium text-blue-800 dark:text-blue-300 mb-2">支付宝支付</p>
              <p class="text-xs text-blue-600 dark:text-blue-400">请使用支付宝扫描下方二维码或点击按钮跳转支付</p>
              <div class="mt-4 w-48 h-48 mx-auto bg-white rounded-lg border-2 border-blue-200 flex items-center justify-center">
                <div class="text-center">
                  <UIcon name="i-lucide-scan" class="w-8 h-8 text-gray-300 mx-auto mb-2" />
                  <p class="text-xs text-gray-400">请在后台配置支付宝</p>
                  <p class="text-xs text-gray-400">当面付/电脑网站支付接口</p>
                </div>
              </div>
            </div>
            <a v-if="payInfo.return_url" :href="payInfo.return_url">
              <UButton variant="outline" block>使用支付宝网页支付</UButton>
            </a>
            <UAlert
              title="支付步骤"
              description="1. 打开支付宝APP扫描二维码，或点击上方按钮跳转支付宝网页支付  2. 确认支付金额并完成付款  3. 支付成功后页面将自动跳转"
              color="info"
              variant="subtle"
            />
          </div>

          <!-- Polling status -->
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
            <UButton variant="outline" @click="goResult">我已支付</UButton>
          </div>
        </template>
      </UCard>
    </template>
  </div>
</template>
