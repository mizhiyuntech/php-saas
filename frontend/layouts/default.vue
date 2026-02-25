<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const tabsStore = useTabsStore()
const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const toast = useToast()

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
  const idx = expandedGroups.value.indexOf(id)
  if (idx >= 0) {
    expandedGroups.value.splice(idx, 1)
  } else {
    expandedGroups.value.push(id)
  }
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
    <!-- Sidebar -->
    <aside class="w-60 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col shrink-0">
      <div class="h-14 flex items-center px-4 border-b border-gray-200 dark:border-gray-800">
        <span class="text-lg font-bold text-primary-600 dark:text-primary-400">{{ settingsStore.siteTitle }}</span>
      </div>

      <nav class="flex-1 overflow-y-auto p-2 space-y-0.5">
        <template v-for="item in menuItems" :key="item.id">
          <template v-if="item.children">
            <button
              class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              @click="toggleGroup(item.id)"
            >
              <UIcon :name="item.icon" class="w-4 h-4 shrink-0" />
              <span class="flex-1 text-left">{{ item.label }}</span>
              <UIcon
                :name="expandedGroups.includes(item.id) ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                class="w-4 h-4 shrink-0"
              />
            </button>
            <div v-show="expandedGroups.includes(item.id)" class="ml-4 space-y-0.5">
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
              class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
              :class="route.path === item.path
                ? 'bg-primary-50 dark:bg-primary-950 text-primary-600 dark:text-primary-400 font-medium'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              @click="navigateToMenu(item as any)"
            >
              <UIcon :name="item.icon" class="w-4 h-4 shrink-0" />
              <span>{{ item.label }}</span>
            </button>
          </template>
        </template>
      </nav>

      <div class="p-3 border-t border-gray-200 dark:border-gray-800">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-user" class="w-4 h-4 text-gray-500" />
            <span class="text-sm text-gray-600 dark:text-gray-400">{{ authStore.username }}</span>
          </div>
          <UButton
            icon="i-lucide-log-out"
            variant="ghost"
            color="neutral"
            size="xs"
            @click="handleLogout"
          />
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Tab bar -->
      <div class="h-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 flex items-center px-2 gap-1 overflow-x-auto shrink-0">
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
  </div>
</template>
