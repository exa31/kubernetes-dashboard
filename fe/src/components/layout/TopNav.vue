<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import BrandLogo from '@/components/brand/BrandLogo.vue'
import { useDarkMode, useK8sRealtime } from '@/composables'
import { ApplyYamlDialog, EventsDrawer, NamespaceSelector } from '@/features/k8s'
import { useAuthStore, useK8sStore } from '@/stores'

defineProps<{
  isSidebarOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-sidebar'): void
}>()

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const k8sStore = useK8sStore()
const { user } = storeToRefs(authStore)
const { clusterInfo } = storeToRefs(k8sStore)
const { isDark, toggle: toggleTheme } = useDarkMode()
const { isConnected, connect, disconnect } = useK8sRealtime()

const isEventsOpen = ref(false)
const isApplyYamlOpen = ref(false)

const warningCount = computed(() => k8sStore.events.filter((e) => e.type === 'Warning').length)

onMounted(() => {
  connect()
  k8sStore.fetchEvents()
  k8sStore.fetchClusterInfo()
})

onBeforeUnmount(() => {
  disconnect()
})

const routeTitles: Record<string, string> = {
  '/': 'Cluster Overview',
  '/overview': 'Cluster Overview',
  '/secrets': 'Secrets (Env)',
  '/configmaps': 'ConfigMaps',
  '/workloads': 'Workloads & Pods',
  '/cronjobs': 'CronJobs',
  '/services': 'Services',
  '/ingresses': 'Ingresses',
  '/storage': 'Storage & Volumes',
  '/users': 'Access & Users',
}

const breadcrumbs = computed(() => {
  const { path } = route
  if (routeTitles[path]) return routeTitles[path]
  const segment = path.replace(/^\//, '').split('/')[0]
  if (!segment) return 'Cluster Overview'
  return segment.charAt(0).toUpperCase() + segment.slice(1).toLowerCase()
})

const handleLogout = async () => {
  await authStore.logout()
  router.push('/auth/login')
}
</script>

<template>
  <header class="h-16 flex items-center justify-between px-3 sm:px-6 bg-white/80 dark:bg-slate-950/80 backdrop-blur-md border-b border-slate-200 dark:border-slate-800 z-10 sticky top-0 min-w-0">
    <div class="flex items-center gap-2 sm:gap-4 min-w-0 shrink">
      <button
        class="p-2 -ml-1 sm:-ml-2 rounded-md text-slate-500 hover:text-slate-900 dark:hover:text-slate-50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer shrink-0"
        @click="emit('toggle-sidebar')"
      >
        <i :class="isSidebarOpen ? 'pi pi-times text-lg' : 'pi pi-bars text-lg'"></i>
      </button>

      <!-- Breadcrumbs -->
      <div class="hidden sm:flex items-center text-sm font-medium text-slate-500 min-w-0 truncate">
        <BrandLogo size="xs" :show-text="false" class="mr-2" />
        <span class="font-bold text-slate-800 dark:text-slate-200">KubeNexus</span>
        <i class="pi pi-chevron-right text-[10px] mx-2 text-slate-400 shrink-0"></i>
        <span class="text-sky-600 dark:text-sky-400 font-semibold truncate">{{ breadcrumbs }}</span>
      </div>
    </div>

    <div class="flex items-center gap-1.5 sm:gap-2.5 shrink-0">
      <!-- Namespace Selector -->
      <NamespaceSelector />

      <!-- Apply YAML Quick Action Button -->
      <button
        type="button"
        class="btn-emerald flex items-center gap-1.5 text-xs font-semibold px-2.5 sm:px-3 py-1.5 rounded-lg shadow-xs cursor-pointer shrink-0 whitespace-nowrap active:scale-95 transition-all"
        title="Apply or validate raw Kubernetes YAML manifest"
        @click="isApplyYamlOpen = true"
      >
        <i class="pi pi-code text-xs sm:text-sm"></i>
        <span class="hidden sm:inline">Apply YAML</span>
      </button>

      <!-- Realtime Live Sync status pill -->
      <div
        class="hidden sm:flex items-center gap-1.5 px-2 sm:px-2.5 py-1 rounded-full text-xs font-medium border font-mono transition-colors shrink-0"
        :class="isConnected ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20' : 'bg-slate-100 dark:bg-slate-800 text-slate-400 border-slate-200 dark:border-slate-700'"
        :title="isConnected ? 'Live Real-Time Sync Connected (SSE) - Cluster 103.150.226.122' : 'Connecting to Live Sync...'"
      >
        <span class="w-1.5 h-1.5 rounded-full" :class="isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
        <span class="hidden xl:inline">{{ isConnected ? 'Live Sync' : 'Connecting...' }}</span>
      </div>

      <!-- Cluster connection pill (Visible on extra-wide screens) -->
      <div class="hidden 2xl:flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800/60 font-mono shrink-0 max-w-[220px]">
        <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shrink-0"></span>
        <span class="truncate">{{ clusterInfo?.connected ? (clusterInfo?.endpoint ? 'Cluster: ' + clusterInfo.endpoint.replace('https://', '') : 'Cluster Connected') : 'Cluster Mode' }}</span>
      </div>

      <!-- Events / Alert Bell button -->
      <button
        type="button"
        class="relative w-8 h-8 sm:w-9 sm:h-9 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer border border-transparent hover:border-slate-200 dark:hover:border-slate-700 shrink-0"
        title="Cluster Events & Alerts"
        @click="isEventsOpen = true"
      >
        <i class="pi pi-bell text-sm sm:text-base"></i>
        <span
          v-if="warningCount > 0"
          class="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-amber-500 text-[9px] font-bold text-slate-950 font-mono"
        >
          {{ warningCount > 9 ? '9+' : warningCount }}
        </span>
      </button>

      <!-- Dark / Light Mode Toggle Button -->
      <button
        type="button"
        class="w-8 h-8 sm:w-9 sm:h-9 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer border border-transparent hover:border-slate-200 dark:hover:border-slate-700 shrink-0"
        :title="isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
        @click="toggleTheme"
      >
        <i :class="isDark ? 'pi pi-sun text-amber-400' : 'pi pi-moon text-sky-600'" class="text-sm sm:text-base"></i>
      </button>

      <div class="h-5 sm:h-6 w-px bg-slate-200 dark:bg-slate-700 mx-0.5 shrink-0 hidden sm:block"></div>

      <!-- User Menu -->
      <div class="flex items-center gap-2 shrink-0">
        <div class="w-8 h-8 rounded-full bg-sky-500/10 text-sky-600 dark:text-sky-400 border border-sky-500/20 flex items-center justify-center font-bold text-xs font-mono shrink-0 select-none">
          {{ (user?.name || 'A').charAt(0).toUpperCase() }}
        </div>

        <div class="hidden xl:block text-right">
          <div class="text-xs font-semibold text-slate-800 dark:text-slate-200 truncate max-w-[130px]">
            {{ user?.name || 'Kubernetes Admin' }}
          </div>
          <div class="text-[11px] text-slate-400 font-mono truncate max-w-[130px]">
            {{ user?.email || 'admin@kubeenv.local' }}
          </div>
        </div>

        <button
          class="p-1.5 sm:p-2 rounded-lg text-slate-500 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer shrink-0"
          title="Sign Out"
          @click="handleLogout"
        >
          <i class="pi pi-sign-out text-sm sm:text-base"></i>
        </button>
      </div>
    </div>

    <!-- Events & Alerts Drawer -->
    <EventsDrawer v-model:visible="isEventsOpen" />

    <!-- Apply YAML Dialog -->
    <ApplyYamlDialog v-model:visible="isApplyYamlOpen" />
  </header>
</template>

<style scoped>
</style>
