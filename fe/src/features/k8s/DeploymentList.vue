<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { computed, onMounted, ref, watch } from 'vue'

import DeploymentEditorDialog from '@/features/k8s/DeploymentEditorDialog.vue'
import PodLogsDialog from '@/features/k8s/PodLogsDialog.vue'
import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'
import WebTerminalDialog from '@/features/k8s/WebTerminalDialog.vue'
import { useK8sStore } from '@/stores'
import type { DaemonSetItem, DeploymentItem, PodItem, StatefulSetItem } from '@/types'

const k8sStore = useK8sStore()
const {
  deployments,
  statefulsets,
  daemonsets,
  pods,
  podMetrics,
  selectedNamespace,
  isLoading,
  isActionLoading,
} = storeToRefs(k8sStore)

// Active tab: deployments | statefulsets | daemonsets | pods
const activeTab = ref<'deployments' | 'statefulsets' | 'daemonsets' | 'pods'>('deployments')

const searchQuery = ref('')
const restartNotification = ref<{ title: string; message: string } | null>(null)

// Dialog states
const isLogsOpen = ref(false)
const isEditorOpen = ref(false)
const isTerminalOpen = ref(false)
const isYamlOpen = ref(false)

const selectedDeploymentName = ref('')
const selectedPod = ref<PodItem | null>(null)
const selectedYamlResource = ref({ kind: 'Deployment', name: '', namespace: '' })
const isScaling = ref<Record<string, boolean>>({})

function getUsageColor(pct: number) {
  if (pct >= 90) return 'text-rose-400'
  if (pct >= 70) return 'text-amber-400'
  return 'text-emerald-400'
}

function getUsageBarColor(pct: number) {
  if (pct >= 90) return 'bg-rose-500'
  if (pct >= 70) return 'bg-amber-500'
  return 'bg-emerald-500'
}

onMounted(() => {
  fetchAllWorkloads()
})

watch(selectedNamespace, () => {
  fetchAllWorkloads()
})

function fetchAllWorkloads() {
  k8sStore.fetchDeployments()
  k8sStore.fetchStatefulSets()
  k8sStore.fetchDaemonSets()
  k8sStore.fetchPods()
  k8sStore.fetchPodMetrics()
}

// Quick scale for deployments
const quickScaleDeployment = async (item: DeploymentItem, newReplicas: number) => {
  if (newReplicas < 0) return
  isScaling.value[item.name] = true
  try {
    await k8sStore.scaleDeployment(item.name, newReplicas)
    showNotification(item.name, `Scaled deployment to ${newReplicas} replicas`)
  } catch (err: unknown) {
    alert(`Failed to scale deployment: ${err instanceof Error ? err.message : 'Unknown error'}`)
  } finally {
    isScaling.value[item.name] = false
  }
}

// Quick scale for statefulsets
const quickScaleStatefulSet = async (item: StatefulSetItem, newReplicas: number) => {
  if (newReplicas < 0) return
  isScaling.value[item.name] = true
  try {
    await k8sStore.scaleStatefulSet(item.name, newReplicas)
    showNotification(item.name, `Scaled statefulset to ${newReplicas} replicas`)
  } catch (err: unknown) {
    alert(`Failed to scale statefulset: ${err instanceof Error ? err.message : 'Unknown error'}`)
  } finally {
    isScaling.value[item.name] = false
  }
}

function showNotification(title: string, message: string) {
  restartNotification.value = { title, message }
  setTimeout(() => {
    restartNotification.value = null
  }, 5000)
}

// Filtering
const filteredDeployments = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return deployments.value
  return deployments.value.filter(
    (d) =>
      d.name.toLowerCase().includes(q) ||
      d.images.some((img) => img.toLowerCase().includes(q)) ||
      d.env_secrets.some((s) => s.toLowerCase().includes(q)),
  )
})

const filteredStatefulSets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return statefulsets.value
  return statefulsets.value.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.images.some((img) => img.toLowerCase().includes(q)),
  )
})

const filteredDaemonSets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return daemonsets.value
  return daemonsets.value.filter(
    (d) =>
      d.name.toLowerCase().includes(q) ||
      d.images.some((img) => img.toLowerCase().includes(q)),
  )
})

const filteredPods = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return pods.value
  return pods.value.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.phase.toLowerCase().includes(q) ||
      (p.status_reason && p.status_reason.toLowerCase().includes(q)) ||
      (p.node && p.node.toLowerCase().includes(q)) ||
      (p.ip && p.ip.includes(q)),
  )
})

// Dialog openers
const openLogsForDeployment = (item: DeploymentItem) => {
  selectedDeploymentName.value = item.name
  selectedPod.value = null
  isLogsOpen.value = true
}

const openLogsForPod = (pod: PodItem) => {
  selectedDeploymentName.value = ''
  selectedPod.value = pod
  isLogsOpen.value = true
}

const openTerminalForPod = (pod: PodItem) => {
  selectedPod.value = pod
  isTerminalOpen.value = true
}

const openEditor = (item: DeploymentItem) => {
  selectedDeploymentName.value = item.name
  isEditorOpen.value = true
}

const openYamlModal = (kind: string, name: string) => {
  selectedYamlResource.value = {
    kind,
    name,
    namespace: selectedNamespace.value,
  }
  isYamlOpen.value = true
}

const restartDeployment = async (item: DeploymentItem) => {
  if (confirm(`Trigger rollout restart for deployment '${item.name}'? Pods will restart sequentially.`)) {
    try {
      const res = await k8sStore.restartDeployment(item.name)
      showNotification(item.name, res.message || 'Rollout restart initiated')
    } catch {
      // handled in store
    }
  }
}

const restartStatefulSet = async (item: StatefulSetItem) => {
  if (confirm(`Trigger rollout restart for statefulset '${item.name}'? Pods will restart sequentially.`)) {
    try {
      await k8sStore.restartStatefulSet(item.name)
      showNotification(item.name, `Rollout restart initiated for ${item.name}`)
    } catch {
      // handled in store
    }
  }
}

const restartDaemonSet = async (item: DaemonSetItem) => {
  if (confirm(`Trigger rollout restart for daemonset '${item.name}' across all nodes?`)) {
    try {
      await k8sStore.restartDaemonSet(item.name)
      showNotification(item.name, `Rollout restart initiated for ${item.name}`)
    } catch {
      // handled in store
    }
  }
}

const deletePodConfirm = async (pod: PodItem) => {
  if (confirm(`Delete / Redeploy pod '${pod.name}'? The controller will automatically recreate it.`)) {
    try {
      await k8sStore.deletePod(pod.name)
      showNotification(pod.name, `Pod '${pod.name}' terminated and recreating`)
    } catch {
      // handled in store
    }
  }
}

function getPhaseColor(phase: string, reason?: string) {
  const r = (reason || phase).toLowerCase()
  if (r.includes('crashloop') || r.includes('oom') || r.includes('failed') || r.includes('error')) {
    return 'danger'
  }
  if (r.includes('pending') || r.includes('containercreating')) {
    return 'warn'
  }
  if (r.includes('terminating')) {
    return 'secondary'
  }
  if (phase.toLowerCase() === 'running') {
    return 'success'
  }
  return 'info'
}
</script>

<template>
  <div class="space-y-4">
    <!-- Top Header -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-objects-column text-sky-500"></i>
          <span>Workload Management</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Monitor Deployments, StatefulSets, DaemonSets, and inspect/redeploy Pods in <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ selectedNamespace }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition flex items-center gap-1.5"
          :disabled="isLoading"
          @click="fetchAllWorkloads"
        >
          <i class="pi pi-refresh text-xs" :class="{ 'pi-spin': isLoading }"></i>
          <span>Refresh</span>
        </button>
      </div>
    </div>

    <!-- Notification Banner -->
    <div
      v-if="restartNotification"
      class="p-4 rounded-xl bg-sky-500/10 border border-sky-500/30 text-sky-400 text-xs flex items-center justify-between shadow-sm animate-fade-in"
    >
      <div class="flex items-center gap-2">
        <i class="pi pi-check-circle text-base text-sky-400"></i>
        <span><strong>{{ restartNotification.title }}:</strong> {{ restartNotification.message }}</span>
      </div>
      <button class="opacity-70 hover:opacity-100" @click="restartNotification = null">
        <i class="pi pi-times"></i>
      </button>
    </div>

    <!-- Workload Navigation Tabs -->
    <div class="flex items-center gap-2 border-b border-slate-200 dark:border-slate-800 pb-2">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2"
        :class="activeTab === 'deployments'
          ? 'bg-sky-500/15 text-sky-400 border border-sky-500/30 shadow-sm'
          : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'"
        @click="activeTab = 'deployments'"
      >
        <i class="pi pi-server text-xs"></i>
        <span>Deployments</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-sky-500/20 text-sky-300">
          {{ deployments.length }}
        </span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2"
        :class="activeTab === 'statefulsets'
          ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30 shadow-sm'
          : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'"
        @click="activeTab = 'statefulsets'"
      >
        <i class="pi pi-database text-xs"></i>
        <span>StatefulSets</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-indigo-500/20 text-indigo-300">
          {{ statefulsets.length }}
        </span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2"
        :class="activeTab === 'daemonsets'
          ? 'bg-teal-500/15 text-teal-400 border border-teal-500/30 shadow-sm'
          : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'"
        @click="activeTab = 'daemonsets'"
      >
        <i class="pi pi-clone text-xs"></i>
        <span>DaemonSets</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-teal-500/20 text-teal-300">
          {{ daemonsets.length }}
        </span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2"
        :class="activeTab === 'pods'
          ? 'bg-purple-500/15 text-purple-400 border border-purple-500/30 shadow-sm'
          : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'"
        @click="activeTab = 'pods'"
      >
        <i class="pi pi-box text-xs"></i>
        <span>Pods Deep-Dive</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-purple-500/20 text-purple-300">
          {{ pods.length }}
        </span>
      </button>
    </div>

    <!-- Search Toolbar -->
    <div class="flex items-center justify-between gap-4">
      <IconField icon-position="left" class="w-full sm:w-80">
        <InputIcon class="pi pi-search text-xs" />
        <InputText
          v-model="searchQuery"
          :placeholder="`Search ${activeTab}...`"
          class="w-full text-xs rounded-xl bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 py-2 text-slate-800 dark:text-slate-200"
        />
      </IconField>
    </div>

    <!-- TAB 1: Deployments -->
    <div v-if="activeTab === 'deployments'" class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
      <DataTable
        :value="filteredDeployments"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <!-- Name -->
        <Column field="name" header="Name" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400 shrink-0">
                <i class="pi pi-server text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{ data.name }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Replicas & Quick Scale -->
        <Column field="ready_replicas" header="Replicas" sortable style="min-width: 10rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2">
              <span
                class="px-2 py-0.5 rounded text-xs font-mono font-bold"
                :class="data.ready_replicas === data.replicas ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30' : 'bg-amber-500/10 text-amber-400 border border-amber-500/30'"
              >
                {{ data.ready_replicas }}/{{ data.replicas }}
              </span>

              <!-- Quick scale buttons -->
              <div class="flex items-center gap-1 bg-slate-800/80 p-0.5 rounded-lg border border-slate-700">
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30"
                  :disabled="isScaling[data.name] || data.replicas <= 0"
                  title="Scale down (-1)"
                  @click="quickScaleDeployment(data, data.replicas - 1)"
                >
                  <i class="pi pi-minus"></i>
                </button>
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30"
                  :disabled="isScaling[data.name]"
                  title="Scale up (+1)"
                  @click="quickScaleDeployment(data, data.replicas + 1)"
                >
                  <i class="pi pi-plus"></i>
                </button>
              </div>
            </div>
          </template>
        </Column>

        <!-- Images -->
        <Column header="Containers & Images" style="min-width: 15rem">
          <template #body="{ data }">
            <div class="space-y-1">
              <div
                v-for="img in data.images"
                :key="img"
                class="text-xs font-mono text-slate-300 truncate max-w-xs"
                :title="img"
              >
                {{ img }}
              </div>
            </div>
          </template>
        </Column>

        <!-- Env Secrets -->
        <Column header="Attached Secrets" style="min-width: 12rem">
          <template #body="{ data }">
            <div class="flex flex-wrap gap-1">
              <span
                v-for="sec in data.env_secrets"
                :key="sec"
                class="px-2 py-0.5 rounded text-[10px] font-mono bg-amber-500/10 text-amber-300 border border-amber-500/30"
              >
                {{ sec }}
              </span>
              <span v-if="data.env_secrets.length === 0" class="text-xs text-slate-500 font-mono">None</span>
            </div>
          </template>
        </Column>

        <!-- Age -->
        <Column field="age" header="Age" sortable style="min-width: 6rem">
          <template #body="{ data }">
            <span class="text-xs text-slate-400 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <!-- Actions -->
        <Column header="Actions" align-frozen="right" style="min-width: 18rem; text-align: right">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <!-- Logs -->
              <Button
                label="Logs"
                icon="pi pi-align-left"
                size="small"
                class="btn-emerald text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="View live container logs"
                @click="openLogsForDeployment(data)"
              />

              <!-- YAML -->
              <Button
                label="YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Inspect & Edit YAML"
                @click="openYamlModal('Deployment', data.name)"
              />

              <!-- Edit -->
              <Button
                label="Edit"
                icon="pi pi-file-edit"
                size="small"
                class="btn-blue text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Edit replicas & containers"
                @click="openEditor(data)"
              />

              <!-- Rollout Restart -->
              <Button
                icon="pi pi-refresh"
                size="small"
                class="btn-amber text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :loading="isActionLoading"
                title="Trigger Rollout Restart"
                @click="restartDeployment(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-server text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No Deployments Found</h3>
            <p class="text-xs text-slate-500 mt-1">No deployments found matching filter in {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 2: StatefulSets -->
    <div v-if="activeTab === 'statefulsets'" class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
      <DataTable
        :value="filteredStatefulSets"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <Column field="name" header="Name" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400 shrink-0">
                <i class="pi pi-database text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{ data.name }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <Column field="ready_replicas" header="Replicas" sortable style="min-width: 10rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2">
              <span
                class="px-2 py-0.5 rounded text-xs font-mono font-bold"
                :class="data.ready_replicas === data.replicas ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30' : 'bg-amber-500/10 text-amber-400 border border-amber-500/30'"
              >
                {{ data.ready_replicas }}/{{ data.replicas }}
              </span>

              <!-- Quick scale -->
              <div class="flex items-center gap-1 bg-slate-800/80 p-0.5 rounded-lg border border-slate-700">
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30"
                  :disabled="isScaling[data.name] || data.replicas <= 0"
                  title="Scale down (-1)"
                  @click="quickScaleStatefulSet(data, data.replicas - 1)"
                >
                  <i class="pi pi-minus"></i>
                </button>
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30"
                  :disabled="isScaling[data.name]"
                  title="Scale up (+1)"
                  @click="quickScaleStatefulSet(data, data.replicas + 1)"
                >
                  <i class="pi pi-plus"></i>
                </button>
              </div>
            </div>
          </template>
        </Column>

        <Column header="Images" style="min-width: 16rem">
          <template #body="{ data }">
            <div class="space-y-1">
              <div v-for="img in data.images" :key="img" class="text-xs font-mono text-slate-300">
                {{ img }}
              </div>
            </div>
          </template>
        </Column>

        <Column field="age" header="Age" sortable style="min-width: 6rem">
          <template #body="{ data }">
            <span class="text-xs text-slate-400 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <Column header="Actions" align-frozen="right" style="min-width: 14rem; text-align: right">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <Button
                label="YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Inspect & Edit YAML"
                @click="openYamlModal('StatefulSet', data.name)"
              />

              <Button
                label="Restart"
                icon="pi pi-refresh"
                size="small"
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :loading="isActionLoading"
                title="Rollout Restart"
                @click="restartStatefulSet(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-database text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No StatefulSets Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no StatefulSets in {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 3: DaemonSets -->
    <div v-if="activeTab === 'daemonsets'" class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
      <DataTable
        :value="filteredDaemonSets"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <Column field="name" header="Name" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-teal-500/10 border border-teal-500/30 flex items-center justify-center text-teal-400 shrink-0">
                <i class="pi pi-clone text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{ data.name }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <Column header="Pod Status" style="min-width: 12rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2 text-xs font-mono">
              <span class="px-2 py-0.5 rounded bg-teal-500/10 text-teal-300 border border-teal-500/30">
                Ready: {{ data.number_ready }}/{{ data.desired_number_scheduled }}
              </span>
              <span class="text-slate-400 text-[11px]">
                ({{ data.number_available }} available)
              </span>
            </div>
          </template>
        </Column>

        <Column header="Images" style="min-width: 16rem">
          <template #body="{ data }">
            <div class="space-y-1">
              <div v-for="img in data.images" :key="img" class="text-xs font-mono text-slate-300">
                {{ img }}
              </div>
            </div>
          </template>
        </Column>

        <Column field="age" header="Age" sortable style="min-width: 6rem">
          <template #body="{ data }">
            <span class="text-xs text-slate-400 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <Column header="Actions" align-frozen="right" style="min-width: 14rem; text-align: right">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <Button
                label="YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Inspect & Edit YAML"
                @click="openYamlModal('DaemonSet', data.name)"
              />

              <Button
                label="Restart"
                icon="pi pi-refresh"
                size="small"
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :loading="isActionLoading"
                title="Rollout Restart"
                @click="restartDaemonSet(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-clone text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No DaemonSets Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no DaemonSets in {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 4: Pods Deep-Dive -->
    <div v-if="activeTab === 'pods'" class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
      <DataTable
        :value="filteredPods"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <!-- Pod Name -->
        <Column field="name" header="Pod Name" sortable style="min-width: 16rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400 shrink-0">
                <i class="pi pi-box text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs font-mono">{{ data.name }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Phase & Status Reason -->
        <Column field="phase" header="Phase & Reason" sortable style="min-width: 10rem">
          <template #body="{ data }">
            <div class="flex items-center gap-1.5">
              <Tag
                :value="data.status_reason || data.phase"
                :severity="getPhaseColor(data.phase, data.status_reason)"
                class="text-[11px] font-mono px-2 py-0.5"
              />
            </div>
          </template>
        </Column>

        <!-- Ready Fraction -->
        <Column field="ready" header="Ready" sortable style="min-width: 6rem">
          <template #body="{ data }">
            <span class="text-xs font-mono text-slate-300 font-semibold">{{ data.ready }}</span>
          </template>
        </Column>

        <!-- Restarts -->
        <Column field="restarts" header="Restarts" sortable style="min-width: 7rem">
          <template #body="{ data }">
            <span
              class="px-2 py-0.5 rounded text-xs font-mono font-bold"
              :class="data.restarts > 0 ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30' : 'text-slate-400'"
            >
              {{ data.restarts }}
            </span>
          </template>
        </Column>

        <!-- Node & IP -->
        <Column header="Placement & IP" style="min-width: 12rem">
          <template #body="{ data }">
            <div class="text-xs font-mono">
              <div class="text-slate-300">{{ data.node || 'unassigned' }}</div>
              <div class="text-[11px] text-emerald-400">{{ data.ip || 'no-ip' }}</div>
            </div>
          </template>
        </Column>

        <!-- Live Metrics Usage (CPU & Memory) -->
        <Column header="Live Resource Usage" style="min-width: 13rem">
          <template #body="{ data }">
            <div v-if="podMetrics[data.name]" class="space-y-1.5 py-0.5">
              <!-- CPU Meter -->
              <div>
                <div class="flex items-center justify-between text-[10px] font-mono mb-0.5">
                  <span class="text-slate-400">CPU: <b class="text-slate-200">{{ podMetrics[data.name]?.cpu_usage }}</b></span>
                  <span class="font-bold" :class="getUsageColor(podMetrics[data.name]?.cpu_percent || 0)">
                    {{ Math.round(podMetrics[data.name]?.cpu_percent || 0) }}%
                  </span>
                </div>
                <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="getUsageBarColor(podMetrics[data.name]?.cpu_percent || 0)"
                    :style="{ width: `${Math.min(100, podMetrics[data.name]?.cpu_percent || 5)}%` }"
                  ></div>
                </div>
              </div>

              <!-- Memory Meter -->
              <div>
                <div class="flex items-center justify-between text-[10px] font-mono mb-0.5">
                  <span class="text-slate-400">Mem: <b class="text-slate-200">{{ podMetrics[data.name]?.memory_usage }}</b></span>
                  <span class="font-bold" :class="getUsageColor(podMetrics[data.name]?.memory_percent || 0)">
                    {{ Math.round(podMetrics[data.name]?.memory_percent || 0) }}%
                  </span>
                </div>
                <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="getUsageBarColor(podMetrics[data.name]?.memory_percent || 0)"
                    :style="{ width: `${Math.min(100, podMetrics[data.name]?.memory_percent || 10)}%` }"
                  ></div>
                </div>
              </div>
            </div>
            <div v-else class="text-[11px] font-mono text-slate-500 italic">
              Telemetry sync...
            </div>
          </template>
        </Column>

        <!-- Age -->
        <Column field="age" header="Age" sortable style="min-width: 6rem">
          <template #body="{ data }">
            <span class="text-xs text-slate-400 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <!-- Actions -->
        <Column header="Actions" align-frozen="right" style="min-width: 18rem; text-align: right">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <!-- Web Terminal Button -->
              <Button
                label="Shell"
                icon="pi pi-terminal"
                size="small"
                class="btn-emerald text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Open interactive in-browser shell"
                @click="openTerminalForPod(data)"
              />

              <!-- Logs Button -->
              <Button
                label="Logs"
                icon="pi pi-align-left"
                size="small"
                class="btn-blue text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="View pod container logs"
                @click="openLogsForPod(data)"
              />

              <!-- YAML Button -->
              <Button
                label="YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="View Pod Manifest"
                @click="openYamlModal('Pod', data.name)"
              />

              <!-- Redeploy / Delete Pod -->
              <Button
                icon="pi pi-trash"
                size="small"
                class="btn-rose text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :loading="isActionLoading"
                title="Redeploy / Delete Pod (triggers restart)"
                @click="deletePodConfirm(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-box text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No Pods Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no pods matching filter in {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Modals -->
    <!-- Pod Logs Dialog Modal -->
    <PodLogsDialog
      v-model:visible="isLogsOpen"
      :deployment-name="selectedDeploymentName"
      :pod-name="selectedPod?.name"
      :namespace="selectedNamespace"
    />

    <!-- Deployment Editor Dialog Modal -->
    <DeploymentEditorDialog
      v-model:visible="isEditorOpen"
      :deployment-name="selectedDeploymentName"
      :namespace="selectedNamespace"
      @saved="k8sStore.fetchDeployments()"
    />

    <!-- Interactive Web Terminal Modal -->
    <WebTerminalDialog
      v-model:visible="isTerminalOpen"
      :pod-name="selectedPod?.name || ''"
      :namespace="selectedNamespace"
      :containers="selectedPod?.containers || []"
    />

    <!-- Live Resource YAML Modal -->
    <ResourceYamlDialog
      v-model:visible="isYamlOpen"
      :kind="selectedYamlResource.kind"
      :name="selectedYamlResource.name"
      :namespace="selectedYamlResource.namespace"
      @applied="fetchAllWorkloads"
    />
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
