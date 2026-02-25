<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const config = useRuntimeConfig()
const baseURL = config.public.apiBase as string || ''
const settingsStore = useSettingsStore()
const toast = useToast()

useHead({ title: '购买授权' })

const step = ref(1)
const loading = ref(false)

const programs = ref<any[]>([])
const packages = ref<any[]>([])
const paymentMethods = ref<any[]>([])

const selectedProgram = ref<any>(null)
const selectedPackage = ref<any>(null)
const selectedPayment = ref('')
const buyerEmail = ref('')

const orderResult = ref<any>(null)
const paying = ref(false)

onMounted(async () => {
  await settingsStore.fetchSiteInfo()
  await fetchPrograms()
})

async function apiFetch(url: string) {
  const res = await fetch(`${baseURL}${url}`)
  return res.json()
}

async function apiPost(url: string, body: any) {
  const res = await fetch(`${baseURL}${url}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
  return res.json()
}

async function fetchPrograms() {
  const res = await apiFetch('/api/public/programs')
  if (res.code === 0) programs.value = res.data || []
}

async function selectProgram(program: any) {
  selectedProgram.value = program
  loading.value = true
  try {
    const [pkgRes, payRes] = await Promise.all([
      apiFetch(`/api/public/packages?program_id=${program.id}`),
      apiFetch('/api/public/payment-methods')
    ])
    if (pkgRes.code === 0) packages.value = pkgRes.data || []
    if (payRes.code === 0) paymentMethods.value = payRes.data || []
  } finally {
    loading.value = false
  }
  step.value = 2
}

function selectPkg(pkg: any) {
  selectedPackage.value = pkg
  step.value = 3
}

async function handlePay() {
  if (!selectedPayment.value) {
    toast.add({ title: '请选择支付方式', color: 'error' })
    return
  }

  paying.value = true
  try {
    const res = await apiPost('/api/public/order', {
      package_id: selectedPackage.value.id,
      payment_method: selectedPayment.value,
      buyer_email: buyerEmail.value
    })

    if (res.code === 0) {
      orderResult.value = res.data
      if (res.data.pay_url && res.data.pay_type === 'redirect') {
        window.location.href = res.data.pay_url
      } else if (res.data.pay_url) {
        navigateTo(res.data.pay_url.replace(/^https?:\/\/[^/]+/, ''))
      } else {
        step.value = 4
      }
    } else {
      toast.add({ title: res.message || '创建订单失败', color: 'error' })
    }
  } finally {
    paying.value = false
  }
}

function goBack() {
  if (step.value > 1) step.value--
}

function restart() {
  step.value = 1
  selectedProgram.value = null
  selectedPackage.value = null
  selectedPayment.value = ''
  orderResult.value = null
}

const paymentIcons: Record<string, string> = {
  wechat: 'i-lucide-smartphone',
  alipay: 'i-lucide-wallet',
  epay: 'i-lucide-credit-card'
}
</script>

<template>
  <div class="w-full max-w-xl mx-auto">
    <!-- Steps indicator -->
    <div class="flex items-center justify-center gap-2 mb-6">
      <template v-for="s in 4" :key="s">
        <div
          class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
          :class="step >= s ? 'bg-primary-500 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-500'"
        >{{ s }}</div>
        <div v-if="s < 4" class="w-10 h-0.5" :class="step > s ? 'bg-primary-500' : 'bg-gray-200 dark:bg-gray-700'" />
      </template>
    </div>

    <div class="text-center mb-4">
      <h1 class="text-xl font-bold text-gray-900 dark:text-white">{{ settingsStore.siteTitle }} - 购买授权</h1>
    </div>

    <!-- Step 1: Select Program -->
    <UCard v-if="step === 1">
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">选择程序</span>
      </template>
      <div v-if="programs.length === 0" class="text-center py-8 text-gray-400">
        暂无可购买的程序
      </div>
      <div v-else class="space-y-2">
        <button
          v-for="prog in programs"
          :key="prog.id"
          class="w-full flex items-center gap-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-400 dark:hover:border-primary-600 transition-colors text-left"
          @click="selectProgram(prog)"
        >
          <div class="w-10 h-10 rounded-lg bg-primary-50 dark:bg-primary-950 flex items-center justify-center shrink-0">
            <UIcon name="i-lucide-box" class="w-5 h-5 text-primary-500" />
          </div>
          <div class="min-w-0">
            <p class="font-medium text-gray-900 dark:text-white">{{ prog.name }}</p>
            <p v-if="prog.description" class="text-sm text-gray-500 truncate">{{ prog.description }}</p>
          </div>
          <UIcon name="i-lucide-chevron-right" class="w-4 h-4 text-gray-400 ml-auto shrink-0" />
        </button>
      </div>
    </UCard>

    <!-- Step 2: Select Package -->
    <UCard v-if="step === 2">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">选择套餐 - {{ selectedProgram?.name }}</span>
          <UButton variant="ghost" color="neutral" size="xs" @click="goBack">返回</UButton>
        </div>
      </template>
      <div v-if="loading" class="text-center py-8 text-gray-400">加载中...</div>
      <div v-else-if="packages.length === 0" class="text-center py-8 text-gray-400">暂无可用套餐</div>
      <div v-else class="space-y-2">
        <button
          v-for="pkg in packages"
          :key="pkg.id"
          class="w-full flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-400 dark:hover:border-primary-600 transition-colors text-left"
          @click="selectPkg(pkg)"
        >
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ pkg.name }}</p>
            <p v-if="pkg.description" class="text-xs text-gray-500 mt-0.5">{{ pkg.description }}</p>
            <p class="text-xs text-gray-400 mt-1">有效期：{{ pkg.duration > 0 ? pkg.duration + '天' : '永久' }}</p>
          </div>
          <div class="text-right shrink-0 ml-4">
            <p class="text-lg font-bold text-orange-600">{{ Number(pkg.price).toFixed(2) }}</p>
            <p class="text-xs text-gray-400">元</p>
          </div>
        </button>
      </div>
    </UCard>

    <!-- Step 3: Payment -->
    <UCard v-if="step === 3">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium text-gray-900 dark:text-white">确认支付</span>
          <UButton variant="ghost" color="neutral" size="xs" @click="goBack">返回</UButton>
        </div>
      </template>

      <div class="space-y-5">
        <!-- Order summary -->
        <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">程序</span>
            <span class="text-gray-900 dark:text-white">{{ selectedProgram?.name }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">套餐</span>
            <span class="text-gray-900 dark:text-white">{{ selectedPackage?.name }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">有效期</span>
            <span class="text-gray-900 dark:text-white">{{ selectedPackage?.duration > 0 ? selectedPackage.duration + '天' : '永久' }}</span>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-700 pt-2 flex justify-between">
            <span class="font-medium text-gray-700 dark:text-gray-300">应付金额</span>
            <span class="text-xl font-bold text-orange-600">{{ Number(selectedPackage?.price).toFixed(2) }} 元</span>
          </div>
        </div>

        <!-- Email -->
        <UFormField label="邮箱（选填）" hint="支付成功后接收授权码">
          <UInput v-model="buyerEmail" type="email" placeholder="your@email.com" />
        </UFormField>

        <!-- Payment methods -->
        <div>
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">选择支付方式</p>
          <div v-if="paymentMethods.length === 0" class="text-sm text-gray-400">暂无可用支付方式，请联系管理员</div>
          <div v-else class="grid grid-cols-2 gap-2">
            <button
              v-for="m in paymentMethods"
              :key="m.type"
              class="flex items-center gap-2 p-3 rounded-lg border transition-colors"
              :class="selectedPayment === m.type
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-950'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'"
              @click="selectedPayment = m.type"
            >
              <UIcon :name="paymentIcons[m.type] || 'i-lucide-credit-card'" class="w-5 h-5 text-gray-600 dark:text-gray-400" />
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ m.label }}</span>
            </button>
          </div>
        </div>

        <UButton block size="lg" :loading="paying" :disabled="!selectedPayment" @click="handlePay">
          立即支付 {{ Number(selectedPackage?.price).toFixed(2) }} 元
        </UButton>
      </div>
    </UCard>

    <!-- Step 4: Waiting / Manual check -->
    <UCard v-if="step === 4">
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">订单已创建</span>
      </template>
      <div class="text-center py-6 space-y-3">
        <UIcon name="i-lucide-clock" class="w-12 h-12 text-orange-500 mx-auto" />
        <p class="text-gray-600 dark:text-gray-400">订单号：<code class="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-sm">{{ orderResult?.order_no }}</code></p>
        <p class="text-sm text-gray-500">请完成支付，支付成功后可在结果页查看授权码</p>
        <NuxtLink :to="`/purchase/result?order_no=${orderResult?.order_no}`">
          <UButton variant="outline" class="mt-2">查看支付结果</UButton>
        </NuxtLink>
      </div>
    </UCard>

    <div class="text-center mt-4">
      <button v-if="step > 1 && step < 4" class="text-sm text-gray-400 hover:text-gray-600" @click="restart">重新选择</button>
    </div>
  </div>
</template>
