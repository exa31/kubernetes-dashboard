<script setup lang="ts">
import BrandLogo from '@/components/brand/BrandLogo.vue'

defineProps<{
  isSidebarOpen: boolean
}>()

defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <aside
    :class="[
      'bg-white dark:bg-slate-950 border-r border-slate-200 dark:border-slate-800 transition-all duration-300 flex flex-col',
      'fixed inset-y-0 left-0 z-30 md:static',
      isSidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full md:translate-x-0 md:w-20'
    ]"
  >
    <div class="h-16 flex items-center px-4 border-b border-slate-200 dark:border-slate-800 shrink-0">
      <router-link to="/overview" class="flex items-center">
        <BrandLogo :collapsed="!isSidebarOpen" size="md" badge="v2.0" />
      </router-link>
    </div>
    <div class="flex-1 overflow-y-auto overflow-x-hidden p-3 space-y-1.5">
      <div v-if="isSidebarOpen" class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 mt-2 px-3">Kubernetes Cluster</div>
      <router-link
        to="/overview"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-emerald-50 dark:bg-emerald-950/30 !text-emerald-600 dark:!text-emerald-400 font-semibold"
      >
        <i class="pi pi-server text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Cluster Overview</span>
      </router-link>

      <div v-if="isSidebarOpen" class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 mt-3 px-3">Config & Workloads</div>
      <router-link
        to="/secrets"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-amber-50 dark:bg-amber-950/30 !text-amber-600 dark:!text-amber-400 font-semibold"
      >
        <i class="pi pi-lock text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Secrets (Env)</span>
      </router-link>
      <router-link
        to="/configmaps"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-sky-50 dark:bg-sky-950/30 !text-sky-600 dark:!text-sky-400 font-semibold"
      >
        <i class="pi pi-file text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">ConfigMaps</span>
      </router-link>
      <router-link
        to="/workloads"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-blue-50 dark:bg-blue-950/30 !text-blue-600 dark:!text-blue-400 font-semibold"
      >
        <i class="pi pi-server text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Workloads & Pods</span>
      </router-link>
      <router-link
        to="/cronjobs"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-amber-50 dark:bg-amber-950/30 !text-amber-600 dark:!text-amber-400 font-semibold"
      >
        <i class="pi pi-clock text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">CronJobs</span>
      </router-link>

      <div v-if="isSidebarOpen" class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 mt-4 px-3">Networking</div>
      <router-link
        to="/services"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-teal-50 dark:bg-teal-950/30 !text-teal-600 dark:!text-teal-400 font-semibold"
      >
        <i class="pi pi-share-alt text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Services</span>
      </router-link>
      <router-link
        to="/ingresses"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-cyan-50 dark:bg-cyan-950/30 !text-cyan-600 dark:!text-cyan-400 font-semibold"
      >
        <i class="pi pi-globe text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Ingresses</span>
      </router-link>

      <div v-if="isSidebarOpen" class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 mt-4 px-3">Storage</div>
      <router-link
        to="/storage"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-teal-50 dark:bg-teal-950/30 !text-teal-600 dark:!text-teal-400 font-semibold"
      >
        <i class="pi pi-database text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Storage & Volumes</span>
      </router-link>

      <div v-if="isSidebarOpen" class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 mt-4 px-3">Management</div>
      <router-link
        to="/users"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors text-slate-600 dark:text-slate-300"
        active-class="bg-emerald-50 dark:bg-emerald-950/40 !text-emerald-600 dark:!text-emerald-400 font-semibold"
      >
        <i class="pi pi-users text-base"></i>
        <span v-if="isSidebarOpen" class="font-medium text-sm">Access & Users</span>
      </router-link>
    </div>
  </aside>
</template>
