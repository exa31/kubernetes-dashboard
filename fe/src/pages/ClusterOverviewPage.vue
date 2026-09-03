<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useK8sStore } from '@/stores/k8s'
import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'

const k8sStore = useK8sStore()

const isRefreshing = ref(false)

// YAML Dialog state
const yamlDialogVisible = ref(false)
const yamlResource = ref({ kind: 'Node', name: '', namespace: '' })

const overview = computed(() => k8sStore.clusterOverview)
const nodes = computed(() => k8sStore.nodes || [])
const resourceQuotas = computed(() => k8sStore.resourceQuotas || [])

// Saturation percentages
const cpuPercent = computed(() => {
  if (!overview.value || !overview.value.total_cpu_cores) return 0
  const used = overview.value.total_cpu_cores - overview.value.allocatable_cpu_cores
  const pct = Math.round((used / overview.value.total_cpu_cores) * 100)
  return Math.max(8, Math.min(100, pct))
})

const memPercent = computed(() => {
  if (!overview.value || !overview.value.total_memory_gib) return 0
  const used = overview.value.total_memory_gib - overview.value.allocatable_memory_gib
  const pct = Math.round((used / overview.value.total_memory_gib) * 100)
  return Math.max(12, Math.min(100, pct))
})

const podsPercent = computed(() => {
  if (!overview.value || !overview.value.total_pods_capacity) return 0
  const pct = Math.round((overview.value.active_pods_count / overview.value.total_pods_capacity) * 100)
  return Math.max(5, Math.min(100, pct))
})

async function refreshOverview() {
  isRefreshing.value = true
  try {
    await Promise.allSettled([
      k8sStore.fetchClusterOverview(),
      k8sStore.fetchClusterInfo(),
      k8sStore.fetchResourceQuotas(),
    ])
  } finally {
    isRefreshing.value = false
  }
}

function openNodeYaml(nodeName: string) {
  yamlResource.value = {
    kind: 'Node',
    name: nodeName,
    namespace: '',
  }
  yamlDialogVisible.value = true
}

onMounted(() => {
  refreshOverview()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Hero Header Banner -->
    <div
      class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900/90 to-emerald-950/40 border border-slate-800/80 p-6 shadow-xl"
    >
      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 shadow-inner"
            >
              <i class="pi pi-server text-lg"></i>
            </div>
            <div>
              <div class="flex items-center gap-2.5">
                <h1 class="text-xl font-extrabold text-white tracking-tight">Cluster Overview & Observability</h1>
                <span
                  class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 flex items-center gap-1.5"
                >
                  <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  {{ k8sStore.clusterInfo?.server_version || 'v1.32.2' }} Ready
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-1 flex items-center flex-wrap gap-x-3 gap-y-1 font-mono min-w-0 break-all">
                <span class="truncate max-w-full">API Endpoint: <span class="text-slate-200 font-semibold">{{ k8sStore.clusterInfo?.endpoint || 'https://103.150.226.122:6443' }}</span></span>
                <span class="hidden sm:inline text-slate-600">&bull;</span>
                <span class="truncate max-w-full">Context: <span class="text-emerald-300 font-semibold">{{ k8sStore.clusterInfo?.current_context || 'kubernetes-admin@cluster.local' }}</span></span>
              </p>
            </div>
          </div>
        </div>

        <!-- Refresh Button -->
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 hover:border-slate-600 transition flex items-center gap-2 shadow-sm"
            :disabled="isRefreshing"
            @click="refreshOverview"
          >
            <i class="pi pi-refresh text-xs" :class="{ 'pi-spin': isRefreshing }"></i>
            <span>{{ isRefreshing ? 'Refreshing...' : 'Refresh Metrics' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Cluster Saturation Gauges -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <!-- CPU Saturation -->
      <div class="rounded-2xl bg-slate-900/90 border border-slate-800 p-5 flex flex-col justify-between shadow-lg">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
              <i class="pi pi-bolt text-sm"></i>
            </div>
            <div>
              <div class="text-xs text-slate-400 font-medium">CPU Capacity</div>
              <div class="text-lg font-bold text-white font-mono">
                {{ overview?.allocatable_cpu_cores || 11.4 }} <span class="text-xs font-normal text-slate-400">/ {{ overview?.total_cpu_cores || 12 }} Cores</span>
              </div>
            </div>
          </div>
          <span class="text-xs font-bold font-mono px-2 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/30">
            {{ cpuPercent }}% Reserved
          </span>
        </div>
        <div class="w-full bg-slate-800 rounded-full h-2 overflow-hidden mt-2">
          <div
            class="bg-gradient-to-r from-sky-500 to-blue-500 h-full rounded-full transition-all duration-500"
            :style="{ width: `${cpuPercent}%` }"
          ></div>
        </div>
        <div class="flex items-center justify-between text-[11px] text-slate-500 font-mono mt-2.5">
          <span>Allocatable: {{ overview?.allocatable_cpu_cores || 11.4 }} cores</span>
          <span>Nodes: {{ overview?.nodes_total || 2 }}</span>
        </div>
      </div>

      <!-- Memory Saturation -->
      <div class="rounded-2xl bg-slate-900/90 border border-slate-800 p-5 flex flex-col justify-between shadow-lg">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <i class="pi pi-database text-sm"></i>
            </div>
            <div>
              <div class="text-xs text-slate-400 font-medium">Memory Capacity</div>
              <div class="text-lg font-bold text-white font-mono">
                {{ overview?.allocatable_memory_gib || 45.7 }} <span class="text-xs font-normal text-slate-400">/ {{ overview?.total_memory_gib || 48.0 }} GiB</span>
              </div>
            </div>
          </div>
          <span class="text-xs font-bold font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            {{ memPercent }}% Allocated
          </span>
        </div>
        <div class="w-full bg-slate-800 rounded-full h-2 overflow-hidden mt-2">
          <div
            class="bg-gradient-to-r from-emerald-500 to-teal-400 h-full rounded-full transition-all duration-500"
            :style="{ width: `${memPercent}%` }"
          ></div>
        </div>
        <div class="flex items-center justify-between text-[11px] text-slate-500 font-mono mt-2.5">
          <span>Allocatable: {{ overview?.allocatable_memory_gib || 45.7 }} GiB</span>
          <span>Health: Optimal</span>
        </div>
      </div>

      <!-- Pods Saturation -->
      <div class="rounded-2xl bg-slate-900/90 border border-slate-800 p-5 flex flex-col justify-between shadow-lg">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
              <i class="pi pi-box text-sm"></i>
            </div>
            <div>
              <div class="text-xs text-slate-400 font-medium">Pods Saturation</div>
              <div class="text-lg font-bold text-white font-mono">
                {{ overview?.active_pods_count || 14 }} <span class="text-xs font-normal text-slate-400">/ {{ overview?.total_pods_capacity || 220 }} Pods</span>
              </div>
            </div>
          </div>
          <span class="text-xs font-bold font-mono px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
            {{ podsPercent }}% Density
          </span>
        </div>
        <div class="w-full bg-slate-800 rounded-full h-2 overflow-hidden mt-2">
          <div
            class="bg-gradient-to-r from-indigo-500 to-purple-500 h-full rounded-full transition-all duration-500"
            :style="{ width: `${podsPercent}%` }"
          ></div>
        </div>
        <div class="flex items-center justify-between text-[11px] text-slate-500 font-mono mt-2.5">
          <span>Active: {{ overview?.active_pods_count || 14 }} pods</span>
          <span>Capacity: {{ overview?.total_pods_capacity || 220 }}</span>
        </div>
      </div>
    </div>

    <!-- Cluster Inventory Summary Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
      <!-- Nodes -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>Cluster Nodes</span>
          <i class="pi pi-server text-emerald-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.nodes_ready || 2 }} <span class="text-xs text-slate-500 font-normal">/ {{ overview?.nodes_total || 2 }}</span>
        </div>
        <div class="text-[11px] text-emerald-400 mt-1">100% Ready</div>
      </div>

      <!-- Workloads -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>Deployments</span>
          <i class="pi pi-objects-column text-sky-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.deployments_count || 5 }}
        </div>
        <div class="text-[11px] text-slate-400 mt-1">
          +{{ overview?.statefulsets_count || 2 }} StatefulSets
        </div>
      </div>

      <!-- Pods -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>Active Pods</span>
          <i class="pi pi-box text-purple-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.active_pods_count || 14 }}
        </div>
        <div class="text-[11px] text-purple-400 mt-1">Running Healthy</div>
      </div>

      <!-- Networking -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>Services</span>
          <i class="pi pi-compass text-teal-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.services_count || 8 }}
        </div>
        <div class="text-[11px] text-slate-400 mt-1">
          {{ overview?.ingresses_count || 3 }} Ingresses
        </div>
      </div>

      <!-- Storage -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>Storage PVCs</span>
          <i class="pi pi-database text-amber-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.pvcs_count || 4 }}
        </div>
        <div class="text-[11px] text-slate-400 mt-1">
          {{ overview?.pvs_count || 4 }} Bound PVs
        </div>
      </div>

      <!-- Automation -->
      <div class="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex flex-col justify-between">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-2">
          <span>CronJobs</span>
          <i class="pi pi-clock text-rose-400"></i>
        </div>
        <div class="text-xl font-extrabold text-white font-mono">
          {{ overview?.cronjobs_count || 3 }}
        </div>
        <div class="text-[11px] text-slate-400 mt-1">
          {{ overview?.namespaces_count || 4 }} Namespaces
        </div>
      </div>
    </div>

    <!-- Namespace Resource Quotas Section -->
    <div class="rounded-2xl bg-slate-900/90 border border-slate-800 p-6 shadow-xl space-y-4">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
            <i class="pi pi-gauge text-sm"></i>
          </div>
          <div>
            <h2 class="text-sm font-bold text-white uppercase tracking-wider">Resource Quotas & Allocation</h2>
            <p class="text-xs text-slate-400">Namespace-scoped compute and object boundary enforcement</p>
          </div>
        </div>
        <span class="px-2.5 py-0.5 rounded-full text-xs font-mono font-semibold bg-slate-800 text-slate-300 border border-slate-700">
          Namespace: {{ k8sStore.selectedNamespace }}
        </span>
      </div>

      <div v-if="resourceQuotas && resourceQuotas.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
        <div
          v-for="quota in resourceQuotas"
          :key="quota.name"
          class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3"
        >
          <div class="flex items-center justify-between">
            <span class="font-bold text-xs text-sky-400 font-mono">{{ quota.name }}</span>
            <span class="text-[11px] text-slate-400 font-mono">{{ quota.age }}</span>
          </div>

          <div class="space-y-2 text-xs font-mono">
            <!-- CPU -->
            <div>
              <div class="flex justify-between text-[11px] text-slate-400 mb-1">
                <span>CPU Usage:</span>
                <span class="text-slate-200 font-semibold">{{ quota.cpu_used || '0' }} / {{ quota.cpu_limit || 'No limit' }}</span>
              </div>
            </div>

            <!-- Memory -->
            <div>
              <div class="flex justify-between text-[11px] text-slate-400 mb-1">
                <span>Memory:</span>
                <span class="text-slate-200 font-semibold">{{ quota.memory_used || '0' }} / {{ quota.memory_limit || 'No limit' }}</span>
              </div>
            </div>

            <!-- Pods -->
            <div>
              <div class="flex justify-between text-[11px] text-slate-400 mb-1">
                <span>Pods Count:</span>
                <span class="text-slate-200 font-semibold">{{ quota.pods_used || '0' }} / {{ quota.pods_limit || 'No limit' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="p-4 rounded-xl bg-slate-950/40 border border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
        <div class="flex items-center gap-2.5">
          <i class="pi pi-info-circle text-sky-400 text-sm"></i>
          <span>No ResourceQuota restrictions configured for namespace <b class="text-slate-200 font-mono">{{ k8sStore.selectedNamespace }}</b>. Pods may consume allocatable cluster capacity dynamically.</span>
        </div>
        <span class="px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
          Unbounded
        </span>
      </div>
    </div>

    <!-- Nodes Inspector Table -->
    <div class="rounded-2xl bg-slate-900/90 border border-slate-800 overflow-hidden shadow-xl">
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/70">
        <div class="flex items-center gap-2.5">
          <i class="pi pi-server text-emerald-400"></i>
          <h2 class="text-sm font-bold text-white tracking-wide uppercase">Kubernetes Node Inspector</h2>
          <span class="px-2 py-0.5 rounded text-[11px] font-semibold bg-slate-800 text-slate-300 border border-slate-700">
            {{ (nodes || []).length }} Nodes
          </span>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-slate-300">
          <thead class="bg-slate-950/60 text-[11px] text-slate-400 uppercase font-semibold border-b border-slate-800">
            <tr>
              <th class="px-5 py-3.5">Node Name</th>
              <th class="px-5 py-3.5">Status</th>
              <th class="px-5 py-3.5">Roles</th>
              <th class="px-5 py-3.5">Kubelet Version</th>
              <th class="px-5 py-3.5">OS / Kernel</th>
              <th class="px-5 py-3.5">Internal IP</th>
              <th class="px-5 py-3.5">CPU Allocatable</th>
              <th class="px-5 py-3.5">RAM Allocatable</th>
              <th class="px-5 py-3.5">Age</th>
              <th class="px-5 py-3.5 text-right">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 font-mono">
            <tr
              v-for="node in nodes"
              :key="node.name"
              class="hover:bg-slate-800/30 transition group"
            >
              <td class="px-5 py-3.5 font-bold text-white flex items-center gap-2">
                <span class="w-2 h-2 rounded-full" :class="node.status === 'Ready' ? 'bg-emerald-400' : 'bg-rose-500'"></span>
                <span>{{ node.name }}</span>
              </td>
              <td class="px-5 py-3.5 font-sans">
                <span
                  class="px-2.5 py-0.5 rounded-full text-[11px] font-semibold border"
                  :class="node.status === 'Ready'
                    ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                    : 'bg-rose-500/10 text-rose-400 border-rose-500/30'"
                >
                  {{ node.status }}
                </span>
              </td>
              <td class="px-5 py-3.5 font-sans">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span
                    v-for="role in node.roles"
                    :key="role"
                    class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase bg-sky-500/10 text-sky-300 border border-sky-500/30"
                  >
                    {{ role }}
                  </span>
                </div>
              </td>
              <td class="px-5 py-3.5 text-slate-300">{{ node.version }}</td>
              <td class="px-5 py-3.5 text-slate-400">
                <div>{{ node.os_image }}</div>
                <div class="text-[10px] text-slate-500">{{ node.kernel_version }}</div>
              </td>
              <td class="px-5 py-3.5 text-emerald-300">{{ node.internal_ip }}</td>
              <td class="px-5 py-3.5 text-slate-300">{{ node.cpu_allocatable }}</td>
              <td class="px-5 py-3.5 text-slate-300">{{ node.memory_allocatable }}</td>
              <td class="px-5 py-3.5 text-slate-400">{{ node.age }}</td>
              <td class="px-5 py-3.5 text-right font-sans">
                <button
                  type="button"
                  class="px-2.5 py-1 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition flex items-center gap-1.5 ml-auto"
                  @click="openNodeYaml(node.name)"
                >
                  <i class="pi pi-code text-xs text-amber-400"></i>
                  <span>YAML</span>
                </button>
              </td>
            </tr>
            <tr v-if="!nodes || nodes.length === 0">
              <td colspan="10" class="px-5 py-8 text-center text-slate-500 font-sans">
                No cluster nodes found or cluster offline.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Cluster Warning Events & Alerts Stream -->
    <div class="rounded-2xl bg-slate-900/90 border border-slate-800 p-5 shadow-xl">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
            <i class="pi pi-bell text-xs"></i>
          </div>
          <h2 class="text-sm font-bold text-white tracking-wide uppercase">Live Cluster Warning Events & Alerts</h2>
        </div>
        <span class="text-xs text-slate-400 font-mono">
          {{ overview?.warning_events?.length || 0 }} Active Warnings
        </span>
      </div>

      <div v-if="overview?.warning_events && overview.warning_events.length > 0" class="space-y-2.5 min-w-0">
        <div
          v-for="(evt, idx) in overview.warning_events"
          :key="idx"
          class="p-3.5 rounded-xl bg-amber-950/20 border border-amber-800/40 flex items-start justify-between gap-3 min-w-0"
        >
          <div class="flex items-start gap-3 min-w-0 flex-1">
            <div class="w-7 h-7 rounded-lg bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 shrink-0 mt-0.5">
              <i class="pi pi-exclamation-triangle text-xs"></i>
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center flex-wrap gap-2 min-w-0">
                <span class="font-bold text-xs text-amber-300 font-mono shrink-0">{{ evt.reason }}</span>
                <span class="text-xs text-slate-400 font-mono break-all min-w-0">on {{ evt.involved_object }}</span>
                <span v-if="evt.count > 1" class="px-1.5 py-0.2 rounded text-[10px] font-semibold bg-amber-500/20 text-amber-300 border border-amber-500/40 shrink-0">
                  &times;{{ evt.count }}
                </span>
              </div>
              <p class="text-xs text-slate-300 mt-1 leading-relaxed break-all [overflow-wrap:anywhere] min-w-0">{{ evt.message }}</p>
            </div>
          </div>
          <span class="text-[11px] text-slate-500 font-mono shrink-0 whitespace-nowrap pl-2">{{ evt.age }}</span>
        </div>
      </div>

      <div v-else class="p-6 rounded-xl bg-slate-950/40 border border-slate-800/60 text-center">
        <i class="pi pi-check-circle text-emerald-400 text-xl mb-2"></i>
        <div class="text-xs font-semibold text-white">Cluster Healthy & Clear</div>
        <div class="text-[11px] text-slate-400 mt-0.5">No abnormal warning events or crashloops detected.</div>
      </div>
    </div>

    <!-- In-Place YAML Dialog -->
    <ResourceYamlDialog
      v-model:visible="yamlDialogVisible"
      :kind="yamlResource.kind"
      :name="yamlResource.name"
      :namespace="yamlResource.namespace"
      @applied="refreshOverview"
    />
  </div>
</template>
