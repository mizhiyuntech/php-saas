<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const config = useRuntimeConfig()
const baseURL = config.public.apiBase as string || ''
const settingsStore = useSettingsStore()
const toast = useToast()

useHead({ title: '首页' })

const activeTab = ref('purchase')

const queryLicenseKey = ref('')
const queryOrderNo = ref('')
const verifyDomain = ref('')

const licenseResult = ref<any>(null)
const orderResult = ref<any>(null)
const domainResult = ref<any>(null)

const queryingLicense = ref(false)
const queryingOrder = ref(false)
const verifyingDomain = ref(false)

onMounted(async () => {
  await settingsStore.fetchSiteInfo()
})

async function apiFetch(url: string, body: any) {
  const res = await fetch(`${baseURL}${url}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
  return res.json()
}

async function handleQueryLicense() {
  if (!queryLicenseKey.value) {
    toast.add({ title: '请输入授权码', color: 'error' })
    return
  }
  queryingLicense.value = true
  licenseResult.value = null
  try {
    const res = await apiFetch('/api/public/query-license', { license_key: queryLicenseKey.value })
    if (res.code === 0) {
      licenseResult.value = res.data
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    queryingLicense.value = false
  }
}

async function handleQueryOrder() {
  if (!queryOrderNo.value) {
    toast.add({ title: '请输入订单号', color: 'error' })
    return
  }
  queryingOrder.value = true
  orderResult.value = null
  try {
    const res = await fetch(`${baseURL}/api/public/order/${queryOrderNo.value}`)
    const data = await res.json()
    if (data.code === 0) {
      orderResult.value = data.data
    } else {
      toast.add({ title: data.message, color: 'error' })
    }
  } finally {
    queryingOrder.value = false
  }
}

async function handleVerifyDomain() {
  if (!verifyDomain.value) {
    toast.add({ title: '请输入域名', color: 'error' })
    return
  }
  verifyingDomain.value = true
  domainResult.value = null
  try {
    const res = await apiFetch('/api/public/verify-domain', { domain: verifyDomain.value })
    if (res.code === 0) {
      domainResult.value = res.data
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    verifyingDomain.value = false
  }
}

const statusPayment: Record<number, string> = { 0: '待支付', 1: '已支付', 2: '支付失败' }

function copyText(t: string) {
  navigator.clipboard.writeText(t)
  toast.add({ title: '已复制', color: 'success' })
}

const tabs = [
  { value: 'purchase', label: '购买授权', icon: 'i-lucide-shopping-cart' },
  { value: 'license', label: '查询授权', icon: 'i-lucide-search' },
  { value: 'order', label: '查询订单', icon: 'i-lucide-receipt' },
  { value: 'domain', label: '域名验证', icon: 'i-lucide-globe' }
]
</script>

<template>
  <div class="w-full max-w-lg mx-auto space-y-4">
    <!-- Header -->
    <div class="text-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ settingsStore.siteTitle }}</h1>
      <p class="mt-1 text-sm text-gray-500">{{ settingsStore.siteDescription || '专业的程序授权管理系统' }}</p>
    </div>

    <!-- Tab selector -->
    <div class="flex gap-1 p-1 bg-gray-100 dark:bg-gray-800 rounded-lg">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        class="flex-1 flex items-center justify-center gap-1.5 py-2 px-2 rounded-md text-xs font-medium transition-colors"
        :class="activeTab === tab.value
          ? 'bg-white dark:bg-gray-900 text-gray-900 dark:text-white shadow-sm'
          : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
        @click="activeTab = tab.value"
      >
        <UIcon :name="tab.icon" class="w-3.5 h-3.5" />
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <!-- Purchase tab -->
    <UCard v-if="activeTab === 'purchase'">
      <div class="text-center py-6 space-y-4">
        <div class="w-14 h-14 mx-auto rounded-full bg-primary-50 dark:bg-primary-950 flex items-center justify-center">
          <UIcon name="i-lucide-shopping-cart" class="w-7 h-7 text-primary-500" />
        </div>
        <div>
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">购买程序授权</h2>
          <p class="text-sm text-gray-500 mt-1">选择程序和套餐，在线购买授权码</p>
        </div>
        <NuxtLink to="/purchase">
          <UButton size="lg">前往购买</UButton>
        </NuxtLink>
      </div>
    </UCard>

    <!-- License query tab -->
    <UCard v-if="activeTab === 'license'">
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">查询授权</span>
      </template>
      <div class="space-y-4">
        <div class="flex gap-2">
          <UInput v-model="queryLicenseKey" class="flex-1" placeholder="请输入授权码" @keyup.enter="handleQueryLicense" />
          <UButton :loading="queryingLicense" @click="handleQueryLicense">查询</UButton>
        </div>

        <div v-if="licenseResult" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500">授权码</span>
            <div class="flex items-center gap-1">
              <span class="font-mono text-gray-700 dark:text-gray-300">{{ licenseResult.license_key }}</span>
              <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(licenseResult.license_key)" />
            </div>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">所属程序</span>
            <span class="text-gray-700 dark:text-gray-300">{{ licenseResult.program_name }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">状态</span>
            <UBadge
              :color="licenseResult.status === 1 ? 'success' : licenseResult.status === 0 ? 'info' : 'error'"
              variant="subtle"
            >{{ licenseResult.status_text }}</UBadge>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">有效期</span>
            <span class="text-gray-700 dark:text-gray-300">{{ licenseResult.duration > 0 ? licenseResult.duration + '天' : '永久' }}</span>
          </div>
          <div v-if="licenseResult.activated_at" class="flex justify-between">
            <span class="text-gray-500">激活时间</span>
            <span class="text-gray-700 dark:text-gray-300">{{ licenseResult.activated_at?.substring(0, 19).replace('T', ' ') }}</span>
          </div>
          <div v-if="licenseResult.expires_at" class="flex justify-between">
            <span class="text-gray-500">过期时间</span>
            <span class="text-gray-700 dark:text-gray-300">{{ licenseResult.expires_at?.substring(0, 19).replace('T', ' ') }}</span>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Order query tab -->
    <UCard v-if="activeTab === 'order'">
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">查询订单</span>
      </template>
      <div class="space-y-4">
        <div class="flex gap-2">
          <UInput v-model="queryOrderNo" class="flex-1" placeholder="请输入订单号" @keyup.enter="handleQueryOrder" />
          <UButton :loading="queryingOrder" @click="handleQueryOrder">查询</UButton>
        </div>

        <div v-if="orderResult" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500">订单号</span>
            <span class="font-mono text-gray-700 dark:text-gray-300">{{ orderResult.order_no }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">程序</span>
            <span class="text-gray-700 dark:text-gray-300">{{ orderResult.program_name }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">金额</span>
            <span class="font-medium text-orange-600">{{ Number(orderResult.amount).toFixed(2) }} 元</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">状态</span>
            <UBadge :color="orderResult.payment_status === 1 ? 'success' : 'warning'" variant="subtle">
              {{ statusPayment[orderResult.payment_status] || '未知' }}
            </UBadge>
          </div>
          <div v-if="orderResult.license_key" class="border-t border-gray-200 dark:border-gray-700 pt-2 mt-2">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">授权码</span>
              <div class="flex items-center gap-1">
                <code class="font-bold text-primary-600 dark:text-primary-400">{{ orderResult.license_key }}</code>
                <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(orderResult.license_key)" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Domain verification tab -->
    <UCard v-if="activeTab === 'domain'">
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">正版域名验证</span>
      </template>
      <div class="space-y-4">
        <p class="text-sm text-gray-500">输入域名查询是否为正版授权，验证程序的合法性</p>
        <div class="flex gap-2">
          <UInput v-model="verifyDomain" class="flex-1" placeholder="例如: example.com" @keyup.enter="handleVerifyDomain" />
          <UButton :loading="verifyingDomain" @click="handleVerifyDomain">验证</UButton>
        </div>

        <div v-if="domainResult" class="p-4 rounded-lg space-y-2 text-sm" :class="domainResult.status === 'authorized' ? 'bg-green-50 dark:bg-green-950' : domainResult.status === 'pirated' ? 'bg-red-50 dark:bg-red-950' : 'bg-gray-50 dark:bg-gray-800'">
          <template v-if="domainResult.status === 'authorized'">
            <div class="flex items-center gap-2 mb-2">
              <UIcon name="i-lucide-check-circle" class="w-5 h-5 text-green-500" />
              <span class="font-medium text-green-700 dark:text-green-400">正版授权</span>
            </div>
            <div v-if="domainResult.program_name" class="flex justify-between">
              <span class="text-gray-500">程序</span>
              <span class="text-gray-700 dark:text-gray-300">{{ domainResult.program_name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">授权码</span>
              <span class="text-gray-700 dark:text-gray-300">{{ domainResult.license_key }}</span>
            </div>
            <div v-if="domainResult.expires_at" class="flex justify-between">
              <span class="text-gray-500">到期时间</span>
              <span class="text-gray-700 dark:text-gray-300">{{ domainResult.expires_at?.substring(0, 10) || '永久' }}</span>
            </div>
          </template>
          <template v-else-if="domainResult.status === 'pirated'">
            <div class="flex items-center gap-2 mb-2">
              <UIcon name="i-lucide-alert-triangle" class="w-5 h-5 text-red-500" />
              <span class="font-medium text-red-700 dark:text-red-400">盗版警告</span>
            </div>
            <p class="text-red-600 dark:text-red-400">{{ domainResult.message }}</p>
          </template>
          <template v-else>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-help-circle" class="w-5 h-5 text-gray-400" />
              <span class="text-gray-600 dark:text-gray-400">{{ domainResult.message }}</span>
            </div>
          </template>
        </div>
      </div>
    </UCard>

    <!-- Footer links -->
    <div class="text-center text-xs text-gray-400 space-y-1">
      <p v-if="settingsStore.footerCopyright">{{ settingsStore.footerCopyright }}</p>
      <p v-if="settingsStore.policeRecord">{{ settingsStore.policeRecord }}</p>
      <NuxtLink to="/login" class="text-gray-400 hover:text-gray-600">管理后台</NuxtLink>
    </div>
  </div>
</template>
