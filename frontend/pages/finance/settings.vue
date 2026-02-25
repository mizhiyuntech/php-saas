<script setup lang="ts">
const { get, put } = useApi()
const toast = useToast()

useHead({ title: '支付配置' })

const loading = ref(true)
const saving = ref(false)
const activePayment = ref('wechat')

const paymentTabs = [
  { label: '微信支付', value: 'wechat', icon: 'i-lucide-smartphone' },
  { label: '支付宝', value: 'alipay', icon: 'i-lucide-wallet' },
  { label: '易支付', value: 'epay', icon: 'i-lucide-credit-card' }
]

interface PaymentData {
  enabled: boolean
  config: Record<string, string>
  return_url?: string
  notify_url?: string
  auth_dir?: string
}

const paymentConfigs = reactive<Record<string, PaymentData>>({
  wechat: {
    enabled: false,
    config: { app_id: '', mch_id: '', api_key: '', cert_path: '' },
    return_url: '',
    notify_url: '',
    auth_dir: ''
  },
  alipay: {
    enabled: false,
    config: { app_id: '', private_key: '', alipay_public_key: '' },
    return_url: '',
    notify_url: ''
  },
  epay: {
    enabled: false,
    config: { api_url: '', merchant_id: '', api_key: '' },
    return_url: '',
    notify_url: ''
  }
})

async function fetchConfigs() {
  loading.value = true
  try {
    const res = await get('/api/finance/payment-configs')
    if (res.code === 0 && res.data) {
      for (const item of res.data) {
        const type = item.payment_type
        if (paymentConfigs[type]) {
          paymentConfigs[type].enabled = item.enabled
          if (item.config && typeof item.config === 'object') {
            paymentConfigs[type].config = { ...paymentConfigs[type].config, ...item.config }
          }
          if (item.return_url) paymentConfigs[type].return_url = item.return_url
          if (item.notify_url) paymentConfigs[type].notify_url = item.notify_url
          if (item.auth_dir) paymentConfigs[type].auth_dir = item.auth_dir
        }
      }
    }
  } finally {
    loading.value = false
  }
}

async function saveConfig(type: string) {
  saving.value = true
  try {
    const data = paymentConfigs[type]
    const res = await put('/api/finance/payment-configs', {
      payment_type: type,
      enabled: data.enabled,
      config: data.config
    })
    if (res.code === 0) {
      toast.add({ title: '保存成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    saving.value = false
  }
}

function copyText(text: string) {
  navigator.clipboard.writeText(text)
  toast.add({ title: '已复制到剪贴板', color: 'success' })
}

onMounted(fetchConfigs)
</script>

<template>
  <div class="space-y-4">
    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">支付配置</span>
      </template>

      <div class="flex gap-2 mb-6">
        <UButton
          v-for="tab in paymentTabs"
          :key="tab.value"
          :variant="activePayment === tab.value ? 'solid' : 'outline'"
          :color="activePayment === tab.value ? 'primary' : 'neutral'"
          :icon="tab.icon"
          @click="activePayment = tab.value"
        >
          {{ tab.label }}
        </UButton>
      </div>

      <!-- WeChat Pay -->
      <div v-show="activePayment === 'wechat'" class="space-y-4">
        <div class="flex items-center gap-2">
          <USwitch v-model="paymentConfigs.wechat.enabled" />
          <span class="text-sm text-gray-700 dark:text-gray-300">启用微信支付</span>
        </div>
        <UFormField label="AppID">
          <UInput v-model="paymentConfigs.wechat.config.app_id" placeholder="微信AppID" />
        </UFormField>
        <UFormField label="商户号 (MchID)">
          <UInput v-model="paymentConfigs.wechat.config.mch_id" placeholder="微信商户号" />
        </UFormField>
        <UFormField label="API密钥">
          <UInput v-model="paymentConfigs.wechat.config.api_key" type="password" placeholder="商户API密钥" />
        </UFormField>

        <div v-if="paymentConfigs.wechat.return_url" class="space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">支付授权相关链接（请复制到微信支付商户平台）</p>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">支付授权目录：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.wechat.auth_dir }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.wechat.auth_dir || '')" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">回调通知地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.wechat.notify_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.wechat.notify_url || '')" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">支付回跳地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.wechat.return_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.wechat.return_url || '')" />
          </div>
        </div>

        <UButton :loading="saving" @click="saveConfig('wechat')">保存配置</UButton>
      </div>

      <!-- Alipay -->
      <div v-show="activePayment === 'alipay'" class="space-y-4">
        <div class="flex items-center gap-2">
          <USwitch v-model="paymentConfigs.alipay.enabled" />
          <span class="text-sm text-gray-700 dark:text-gray-300">启用支付宝</span>
        </div>
        <UFormField label="AppID">
          <UInput v-model="paymentConfigs.alipay.config.app_id" placeholder="支付宝AppID" />
        </UFormField>
        <UFormField label="应用私钥">
          <UTextarea v-model="paymentConfigs.alipay.config.private_key" placeholder="RSA2私钥" :rows="3" />
        </UFormField>
        <UFormField label="支付宝公钥">
          <UTextarea v-model="paymentConfigs.alipay.config.alipay_public_key" placeholder="支付宝公钥" :rows="3" />
        </UFormField>

        <div v-if="paymentConfigs.alipay.return_url" class="space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">支付相关链接（请复制到支付宝开放平台）</p>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">回调通知地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.alipay.notify_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.alipay.notify_url || '')" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">支付回跳地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.alipay.return_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.alipay.return_url || '')" />
          </div>
        </div>

        <UButton :loading="saving" @click="saveConfig('alipay')">保存配置</UButton>
      </div>

      <!-- EPay -->
      <div v-show="activePayment === 'epay'" class="space-y-4">
        <div class="flex items-center gap-2">
          <USwitch v-model="paymentConfigs.epay.enabled" />
          <span class="text-sm text-gray-700 dark:text-gray-300">启用易支付</span>
        </div>
        <UFormField label="接口地址">
          <UInput v-model="paymentConfigs.epay.config.api_url" placeholder="如 https://pay.example.com" />
        </UFormField>
        <UFormField label="商户ID">
          <UInput v-model="paymentConfigs.epay.config.merchant_id" placeholder="易支付商户ID" />
        </UFormField>
        <UFormField label="商户密钥">
          <UInput v-model="paymentConfigs.epay.config.api_key" type="password" placeholder="商户密钥" />
        </UFormField>

        <div v-if="paymentConfigs.epay.return_url" class="space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">支付相关链接（请复制到易支付商户后台）</p>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">回调通知地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.epay.notify_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.epay.notify_url || '')" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 w-20">支付回跳地址：</span>
            <code class="text-xs flex-1 bg-white dark:bg-gray-900 px-2 py-1 rounded">{{ paymentConfigs.epay.return_url }}</code>
            <UButton icon="i-lucide-copy" variant="ghost" color="neutral" size="xs" @click="copyText(paymentConfigs.epay.return_url || '')" />
          </div>
        </div>

        <UButton :loading="saving" @click="saveConfig('epay')">保存配置</UButton>
      </div>
    </UCard>
  </div>
</template>
