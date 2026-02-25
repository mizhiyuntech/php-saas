<script setup lang="ts">
const { get, put, post } = useApi()
const toast = useToast()

useHead({ title: 'SMTP邮件配置' })

const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const testEmail = ref('')

const form = reactive({
  host: '',
  port: 465,
  user: '',
  password: '',
  from_name: '',
  ssl: true
})

async function fetchConfig() {
  loading.value = true
  try {
    const res = await get('/api/smtp')
    if (res.code === 0 && res.data) {
      Object.assign(form, res.data)
    }
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!form.host || !form.user || !form.password) {
    toast.add({ title: '请填写完整的SMTP配置', color: 'error' })
    return
  }
  saving.value = true
  try {
    const res = await put('/api/smtp', form)
    if (res.code === 0) {
      toast.add({ title: '保存成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  if (!testEmail.value) {
    toast.add({ title: '请输入测试收件邮箱', color: 'error' })
    return
  }
  testing.value = true
  try {
    const res = await post('/api/smtp/test', { to: testEmail.value })
    if (res.code === 0) {
      toast.add({ title: '测试邮件发送成功', color: 'success' })
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    testing.value = false
  }
}

onMounted(fetchConfig)
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-4">
    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">SMTP邮件配置</span>
      </template>

      <div class="space-y-4">
        <UAlert
          title="说明"
          description="配置SMTP后，系统可以发送订单通知、授权码通知、验证码等邮件。支持QQ邮箱、163邮箱、阿里企业邮箱、腾讯企业邮箱等。"
          color="info"
          variant="subtle"
        />

        <div class="grid grid-cols-2 gap-4">
          <UFormField label="SMTP服务器" required>
            <UInput v-model="form.host" placeholder="如 smtp.qq.com" />
          </UFormField>
          <UFormField label="端口" required>
            <UInput v-model.number="form.port" type="number" placeholder="465" />
          </UFormField>
        </div>

        <UFormField label="邮箱账号" required>
          <UInput v-model="form.user" placeholder="your@email.com" />
        </UFormField>

        <UFormField label="邮箱密码/授权码" required>
          <UInput v-model="form.password" type="password" placeholder="SMTP密码或授权码" />
        </UFormField>

        <UFormField label="发件人名称">
          <UInput v-model="form.from_name" placeholder="如 鱼跃授权" />
        </UFormField>

        <div class="flex items-center gap-2">
          <USwitch v-model="form.ssl" />
          <span class="text-sm text-gray-700 dark:text-gray-300">启用SSL加密</span>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton :loading="saving" @click="handleSave">保存配置</UButton>
        </div>
      </template>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">发送测试</span>
      </template>
      <div class="flex items-end gap-3">
        <UFormField label="收件邮箱" class="flex-1">
          <UInput v-model="testEmail" placeholder="输入收件邮箱进行测试" />
        </UFormField>
        <UButton :loading="testing" variant="outline" @click="handleTest">发送测试邮件</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium text-gray-900 dark:text-white">常用SMTP配置参考</span>
      </template>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="text-left py-2 px-3 text-gray-500 font-medium">邮箱</th>
              <th class="text-left py-2 px-3 text-gray-500 font-medium">SMTP服务器</th>
              <th class="text-left py-2 px-3 text-gray-500 font-medium">端口</th>
              <th class="text-left py-2 px-3 text-gray-500 font-medium">SSL</th>
            </tr>
          </thead>
          <tbody class="text-gray-700 dark:text-gray-300">
            <tr class="border-b border-gray-100 dark:border-gray-800"><td class="py-2 px-3">QQ邮箱</td><td class="py-2 px-3">smtp.qq.com</td><td class="py-2 px-3">465</td><td class="py-2 px-3">是</td></tr>
            <tr class="border-b border-gray-100 dark:border-gray-800"><td class="py-2 px-3">163邮箱</td><td class="py-2 px-3">smtp.163.com</td><td class="py-2 px-3">465</td><td class="py-2 px-3">是</td></tr>
            <tr class="border-b border-gray-100 dark:border-gray-800"><td class="py-2 px-3">阿里企业邮箱</td><td class="py-2 px-3">smtp.qiye.aliyun.com</td><td class="py-2 px-3">465</td><td class="py-2 px-3">是</td></tr>
            <tr class="border-b border-gray-100 dark:border-gray-800"><td class="py-2 px-3">腾讯企业邮箱</td><td class="py-2 px-3">smtp.exmail.qq.com</td><td class="py-2 px-3">465</td><td class="py-2 px-3">是</td></tr>
            <tr><td class="py-2 px-3">Gmail</td><td class="py-2 px-3">smtp.gmail.com</td><td class="py-2 px-3">465</td><td class="py-2 px-3">是</td></tr>
          </tbody>
        </table>
      </div>
    </UCard>
  </div>
</template>
