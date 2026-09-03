<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Card from 'primevue/card'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useK8sStore } from '@/stores'

const router = useRouter()
const k8sStore = useK8sStore()
const { clusterInfo, selectedNamespace, secrets, configmaps, deployments, isLoading } = storeToRefs(k8sStore)

onMounted(() => {
  k8sStore.fetchAllResources()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-compass text-sky-500"></i>
          <span>Cluster Overview</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Lightweight environment variable & workload management dashboard
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          label="Refresh Cluster Data"
          icon="pi pi-refresh"
          size="small"
          class="btn-sky text-xs px-3 py-1.5 rounded-lg active:scale-95 cursor-pointer"
          :loading="isLoading"
          @click="k8sStore.fetchAllResources()"
        />
      </div>
    </div>

    <!-- Cluster Connection Status Hero Banner -->
    <div class="p-6 rounded-2xl bg-gradient-to-r from-slate-900 via-sky-950 to-slate-900 border border-sky-900/40 text-white shadow-xl relative overflow-hidden">
      <div class="relative z-10 flex flex-wrap items-center justify-between gap-6">
        <div>
          <div class="flex items-center gap-2 mb-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
            <span class="text-xs font-semibold uppercase tracking-wider text-emerald-400">
              {{ clusterInfo?.connected ? 'Kubernetes Control Plane Active' : 'Cluster Mode' }}
            </span>
          </div>
          <h2 class="text-xl font-bold font-mono tracking-tight text-slate-100">
            {{ clusterInfo?.endpoint || 'https://103.150.226.122:6443' }}
          </h2>
          <div class="flex flex-wrap items-center gap-4 text-xs text-slate-300 mt-2 font-mono">
            <span>Version: <strong class="text-sky-300">{{ clusterInfo?.server_version || 'v1.32.2' }}</strong></span>
            <span>•</span>
            <span>Context: <strong class="text-sky-300">{{ clusterInfo?.current_context || 'default' }}</strong></span>
            <span>•</span>
            <span>Active Namespace: <strong class="text-amber-300">{{ selectedNamespace }}</strong></span>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <div class="text-right hidden sm:block">
            <div class="text-xs text-slate-400">Total Namespaces</div>
            <div class="text-2xl font-bold text-white">{{ clusterInfo?.namespace_count || 3 }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats Grid for Active Namespace -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <!-- Secrets Card -->
      <Card class="shadow-sm border border-slate-200 dark:border-slate-800 hover:border-amber-400/50 transition-all cursor-pointer group" @click="router.push('/secrets')">
        <template #content>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400">Secrets (Env)</div>
              <div class="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-1 font-mono">
                {{ secrets.length }}
              </div>
              <div class="text-xs text-slate-500 mt-1">
                Decoded variables in {{ selectedNamespace }}
              </div>
            </div>
            <div class="w-12 h-12 rounded-xl bg-amber-500/10 text-amber-500 flex items-center justify-center text-xl group-hover:scale-110 transition-transform">
              <i class="pi pi-lock"></i>
            </div>
          </div>
        </template>
      </Card>

      <!-- ConfigMaps Card -->
      <Card class="shadow-sm border border-slate-200 dark:border-slate-800 hover:border-sky-400/50 transition-all cursor-pointer group" @click="router.push('/configmaps')">
        <template #content>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-sky-600 dark:text-sky-400">ConfigMaps</div>
              <div class="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-1 font-mono">
                {{ configmaps.length }}
              </div>
              <div class="text-xs text-slate-500 mt-1">
                Plaintext configurations in {{ selectedNamespace }}
              </div>
            </div>
            <div class="w-12 h-12 rounded-xl bg-sky-500/10 text-sky-500 flex items-center justify-center text-xl group-hover:scale-110 transition-transform">
              <i class="pi pi-file"></i>
            </div>
          </div>
        </template>
      </Card>

      <!-- Deployments Card -->
      <Card class="shadow-sm border border-slate-200 dark:border-slate-800 hover:border-blue-400/50 transition-all cursor-pointer group" @click="router.push('/workloads')">
        <template #content>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-xs font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400">Deployments</div>
              <div class="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-1 font-mono">
                {{ deployments.length }}
              </div>
              <div class="text-xs text-slate-500 mt-1">
                Workloads with 1-click Rollout Restart
              </div>
            </div>
            <div class="w-12 h-12 rounded-xl bg-blue-500/10 text-blue-500 flex items-center justify-center text-xl group-hover:scale-110 transition-transform">
              <i class="pi pi-server"></i>
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Quick Environment Management Panel -->
    <div class="bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="font-bold text-base text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i class="pi pi-bolt text-sky-500"></i>
            <span>Quick Secrets Access in {{ selectedNamespace }}</span>
          </h3>
          <p class="text-xs text-slate-500 mt-0.5">Click any secret to open the modern Rancher-style Env Editor</p>
        </div>
        <Button
          label="View All Secrets"
          icon="pi pi-arrow-right"
          icon-pos="right"
          size="small"
          class="text-sky-600 dark:text-sky-400 font-semibold hover:underline bg-transparent border-none cursor-pointer"
          @click="router.push('/secrets')"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <div
          v-for="s in secrets.slice(0, 6)"
          :key="s.name"
          class="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 hover:border-sky-500 hover:shadow-md transition-all cursor-pointer flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/40"
          @click="router.push('/secrets')"
        >
          <div class="min-w-0 flex items-center gap-2.5">
            <i class="pi pi-lock text-amber-500 text-sm"></i>
            <div class="min-w-0">
              <div class="font-semibold text-xs font-mono truncate text-slate-900 dark:text-slate-100">
                {{ s.name }}
              </div>
              <div class="text-[11px] text-slate-400 font-mono mt-0.5">
                {{ s.key_count }} keys • {{ s.age }}
              </div>
            </div>
          </div>
          <i class="pi pi-chevron-right text-xs text-slate-400"></i>
        </div>

        <div
          v-if="secrets.length === 0"
          class="col-span-full py-8 text-center text-slate-400 text-xs italic"
        >
          No secrets found in this namespace. Select another namespace in the header.
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
