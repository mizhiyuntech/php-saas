<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const tabsStore = useTabsStore()
const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const toast = useToast()

const sidebarCollapsed = ref(false)
const mobileMenuOpen = ref(false)

const menuItems = [
  { id: 'dashboard', label: '仪表盘', icon: 'i-lucide-layout-dashboard', path: '/' },
  { id: 'programs', label: '程序管理', icon: 'i-lucide-box', path: '/programs' },
  { id: 'licenses', label: '授权管理', icon: 'i-lucide-key-round', path: '/licenses' },
  { id: 'injection', label: '在线注入', icon: 'i-lucide-upload', path: '/injection' },
  {
    id: 'finance',
    label: '财务管理',
    icon: 'i-lucide-wallet',
    children: [
      { id: 'finance-settings', label: '支付配置', icon: 'i-lucide-credit-card', path: '/finance/settings' },
      { id: 'finance-orders', label: '订单查询', icon: 'i-lucide-receipt', path: '/finance/orders' }
    ]
  },
  { id: 'settings', label: '系统设置', icon: 'i-lucide-settings', path: '/settings' }
]

const expandedGroups = ref<string[]>(['finance'])

function toggleGroup(id: string) {
  if (sidebarCollapsed.value) {
    sidebarCollapsed.value = false
    if (!expandedGroups.value.includes(id)) {
      expandedGroups.value.push(id)
    }
    return
  }
  const idx = expandedGroups.value.indexOf(id)
  if (idx >= 0) {
    expandedGroups.value.splice(idx, 1)
  } else {
    expandedGroups.value.push(id)
  }
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

const breadcrumbItems = computed(() => {
  const items: { label: string; to?: string }[] = [{ label: '首页', to: '/' }]
  for (const item of menuItems) {
    if (item.children) {
      for (const child of item.children) {
        if (child.path === route.path) {
          items.push({ label: item.label })
          items.push({ label: child.label })
          return items
        }
      }
    } else if (item.path === route.path) {
      items.push({ label: item.label })
      return items
    }
  }
  return items
})

function navigateToMenu(item: { id: string; label: string; path: string }) {
  tabsStore.addTab({
    id: item.id,
    title: item.label,
    path: item.path,
    closable: item.id !== 'dashboard'
  })
  router.push(item.path)
  mobileMenuOpen.value = false
}

function switchTab(tab: { id: string; path: string }) {
  tabsStore.setActive(tab.id)
  router.push(tab.path)
}

function closeTab(id: string) {
  const redirectPath = tabsStore.removeTab(id)
  if (redirectPath) {
    router.push(redirectPath)
  }
}

const showPassword = ref(false)
const pwdLoading = ref(false)
const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

function openChangePassword() {
  Object.assign(pwdForm, { old_password: '', new_password: '', confirm_password: '' })
  showPassword.value = true
}

async function handleChangePassword() {
  if (!pwdForm.old_password || !pwdForm.new_password) {
    toast.add({ title: '请填写完整', color: 'error' })
    return
  }
  if (pwdForm.new_password.length < 6) {
    toast.add({ title: '新密码长度不能少于6位', color: 'error' })
    return
  }
  if (pwdForm.new_password !== pwdForm.confirm_password) {
    toast.add({ title: '两次密码输入不一致', color: 'error' })
    return
  }
  pwdLoading.value = true
  try {
    const { post } = useApi()
    const res = await post('/api/auth/change-password', {
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password
    })
    if (res.code === 0) {
      toast.add({ title: '密码修改成功', color: 'success' })
      showPassword.value = false
    } else {
      toast.add({ title: res.message, color: 'error' })
    }
  } finally {
    pwdLoading.value = false
  }
}

function handleLogout() {
  authStore.logout()
  toast.add({ title: '已退出登录', color: 'info' })
  router.push('/login')
}

watch(() => route.path, (newPath) => {
  for (const item of menuItems) {
    if (item.children) {
      for (const child of item.children) {
        if (child.path === newPath) {
          tabsStore.addTab({ id: child.id, title: child.label, path: child.path, closable: true })
          return
        }
      }
    } else if (item.path === newPath) {
      tabsStore.addTab({ id: item.id, title: item.label, path: item.path, closable: item.id !== 'dashboard' })
      return
    }
  }
}, { immediate: true })
</script>

<template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-950">
    <!-- Mobile overlay backdrop -->
    <div
      v-if="mobileMenuOpen"
      class="fixed inset-0 z-30 bg-black/50 lg:hidden"
      @click="mobileMenuOpen = false"
    />

    <!-- Sidebar -->
    <aside
      class="bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col shrink-0 transition-all duration-200 z-40"
      :class="[
        sidebarCollapsed ? 'w-14' : 'w-60',
        mobileMenuOpen ? 'fixed inset-y-0 left-0' : 'hidden lg:flex'
      ]"
    >
      <!-- Sidebar header -->
      <div class="h-14 flex items-center border-b border-gray-200 dark:border-gray-800" :class="sidebarCollapsed ? 'justify-center px-2' : 'px-4'">
        <span v-if="!sidebarCollapsed" class="text-lg font-bold text-primary-600 dark:text-primary-400 truncate">{{ settingsStore.siteTitle }}</span>
        <UButton
          v-if="sidebarCollapsed"
          icon="i-lucide-panel-right-open"
          variant="ghost"
          color="neutral"
          size="xs"
          @click="toggleSidebar"
        />
      </div>

      <!-- Nav -->
      <nav class="flex-1 overflow-y-auto p-2 space-y-0.5">
        <template v-for="item in menuItems" :key="item.id">
          <template v-if="item.children">
            <button
              class="w-full flex items-center rounded-md text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              :class="sidebarCollapsed ? 'justify-center px-2 py-2' : 'gap-2 px-3 py-2'"
              :title="sidebarCollapsed ? item.label : undefined"
              @click="toggleGroup(item.id)"
            >
              <UIcon :name="item.icon" class="w-4 h-4 shrink-0" />
              <template v-if="!sidebarCollapsed">
                <span class="flex-1 text-left">{{ item.label }}</span>
                <UIcon
                  :name="expandedGroups.includes(item.id) ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                  class="w-4 h-4 shrink-0"
                />
              </template>
            </button>
            <div v-if="!sidebarCollapsed" v-show="expandedGroups.includes(item.id)" class="ml-4 space-y-0.5">
              <button
                v-for="child in item.children"
                :key="child.id"
                class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
                :class="route.path === child.path
                  ? 'bg-primary-50 dark:bg-primary-950 text-primary-600 dark:text-primary-400 font-medium'
                  : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
                @click="navigateToMenu(child)"
              >
                <UIcon :name="child.icon" class="w-4 h-4 shrink-0" />
                <span>{{ child.label }}</span>
              </button>
            </div>
          </template>
          <template v-else>
            <button
              class="w-full flex items-center rounded-md text-sm transition-colors"
              :class="[
                sidebarCollapsed ? 'justify-center px-2 py-2' : 'gap-2 px-3 py-2',
                route.path === item.path
                  ? 'bg-primary-50 dark:bg-primary-950 text-primary-600 dark:text-primary-400 font-medium'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
              ]"
              :title="sidebarCollapsed ? item.label : undefined"
              @click="navigateToMenu(item as any)"
            >
              <UIcon :name="item.icon" class="w-4 h-4 shrink-0" />
              <span v-if="!sidebarCollapsed">{{ item.label }}</span>
            </button>
          </template>
        </template>
      </nav>

      <!-- Sidebar footer -->
      <div class="p-2 border-t border-gray-200 dark:border-gray-800">
        <template v-if="sidebarCollapsed">
          <div class="flex flex-col items-center gap-1">
            <UButton icon="i-lucide-lock" variant="ghost" color="neutral" size="xs" @click="openChangePassword" />
            <UButton icon="i-lucide-log-out" variant="ghost" color="neutral" size="xs" @click="handleLogout" />
          </div>
        </template>
        <template v-else>
          <div class="flex items-center justify-between px-1">
            <div class="flex items-center gap-2 min-w-0">
              <UIcon name="i-lucide-user" class="w-4 h-4 text-gray-500 shrink-0" />
              <span class="text-sm text-gray-600 dark:text-gray-400 truncate">{{ authStore.username }}</span>
            </div>
            <div class="flex items-center gap-0.5 shrink-0">
              <UButton icon="i-lucide-lock" variant="ghost" color="neutral" size="xs" @click="openChangePassword" />
              <UButton icon="i-lucide-log-out" variant="ghost" color="neutral" size="xs" @click="handleLogout" />
            </div>
          </div>
        </template>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col overflow-hidden min-w-0">
      <!-- Tab bar -->
      <div class="h-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 flex items-center px-2 gap-1 overflow-x-auto shrink-0">
        <!-- Mobile menu toggle -->
        <UButton
          icon="i-lucide-menu"
          variant="ghost"
          color="neutral"
          size="xs"
          class="lg:hidden shrink-0 mr-1"
          @click="toggleMobileMenu"
        />
        <!-- PC sidebar collapse toggle -->
        <UButton
          :icon="sidebarCollapsed ? 'i-lucide-panel-left-open' : 'i-lucide-panel-left-close'"
          variant="ghost"
          color="neutral"
          size="xs"
          class="hidden lg:inline-flex shrink-0 mr-1"
          @click="toggleSidebar"
        />
        <button
          v-for="tab in tabsStore.tabs"
          :key="tab.id"
          class="flex items-center gap-1 px-3 py-1.5 rounded text-xs whitespace-nowrap transition-colors group"
          :class="tabsStore.activeTab === tab.id
            ? 'bg-primary-50 dark:bg-primary-950 text-primary-600 dark:text-primary-400'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
          @click="switchTab(tab)"
        >
          <span>{{ tab.title }}</span>
          <span
            v-if="tab.closable"
            class="ml-1 w-4 h-4 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 opacity-0 group-hover:opacity-100 transition-opacity"
            @click.stop="closeTab(tab.id)"
          >
            <UIcon name="i-lucide-x" class="w-3 h-3" />
          </span>
        </button>
      </div>

      <!-- Breadcrumb -->
      <div class="px-4 py-2 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <UBreadcrumb :items="breadcrumbItems" />
      </div>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto p-4">
        <slot />
      </main>

      <!-- Footer -->
      <footer v-if="settingsStore.footerCopyright || settingsStore.policeRecord" class="px-4 py-2 text-center text-xs text-gray-400 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 shrink-0">
        <span v-if="settingsStore.footerCopyright">{{ settingsStore.footerCopyright }}</span>
        <span v-if="settingsStore.footerCopyright && settingsStore.policeRecord" class="mx-2">|</span>
        <span v-if="settingsStore.policeRecord">{{ settingsStore.policeRecord }}</span>
      </footer>
    </div>

    <!-- Change Password Modal -->
    <UModal v-model:open="showPassword">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">修改密码</span>
              <UButton icon="i-lucide-x" variant="ghost" color="neutral" size="xs" @click="showPassword = false" />
            </div>
          </template>

          <div class="space-y-4">
            <UFormField label="原密码">
              <UInput v-model="pwdForm.old_password" type="password" placeholder="请输入原密码" />
            </UFormField>
            <UFormField label="新密码">
              <UInput v-model="pwdForm.new_password" type="password" placeholder="至少6位" />
            </UFormField>
            <UFormField label="确认新密码">
              <UInput v-model="pwdForm.confirm_password" type="password" placeholder="再次输入新密码" />
            </UFormField>
          </div>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton variant="outline" color="neutral" @click="showPassword = false">取消</UButton>
              <UButton :loading="pwdLoading" @click="handleChangePassword">确定</UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
