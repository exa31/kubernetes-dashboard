<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { useK8sRealtime } from '@/composables'
import DeploymentEditorDialog from '@/features/k8s/DeploymentEditorDialog.vue'
import DeploymentRollbackDialog from '@/features/k8s/DeploymentRollbackDialog.vue'
import DeployWorkloadDialog from '@/features/k8s/DeployWorkloadDialog.vue'
import HpaScaleDialog from '@/features/k8s/HpaScaleDialog.vue'
import PodLogsDialog from '@/features/k8s/PodLogsDialog.vue'
import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'
import WebTerminalDialog from '@/features/k8s/WebTerminalDialog.vue'
import { useAuthStore, useK8sStore } from '@/stores'
import type {
  DaemonSetItem,
  DeploymentItem,
  HPAItem,
  KedaHTTPScaledObject,
  PodItem,
  StatefulSetItem
} from '@/types'

const authStore = useAuthStore()
const k8sStore = useK8sStore()
const confirm = useConfirm()
const toast = useToast()
const {
  deployments,
  statefulsets,
  daemonsets,
  pods,
  hpas,
  kedaHttpObjects,
  podMetrics,
  selectedNamespace,
  isLoading,
  isActionLoading
} = storeToRefs(k8sStore)

const canMutate = computed(() => authStore.canMutateNamespace(selectedNamespace.value))

// Active tab: deployments | statefulsets | daemonsets | pods | hpas
const activeTab = ref<'deployments' | 'statefulsets' | 'daemonsets' | 'pods' | 'hpas'>('deployments')

const searchQuery = ref('')
const restartNotification = ref<{ title: string; message: string } | null>(null)

// Dialog states
const isDeployOpen = ref(false)
const isLogsOpen = ref(false)
const isEditorOpen = ref(false)
const isTerminalOpen = ref(false)
const isYamlOpen = ref(false)
const isRollbackOpen = ref(false)
const isHpaScaleOpen = ref(false)

const selectedHpaTarget = ref<{
  name: string
  kind: 'Deployment' | 'StatefulSet'
  currentReplicas: number
}>({
  name: '',
  kind: 'Deployment',
  currentReplicas: 1
})

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

const { isConnected } = useK8sRealtime()

// Auto-refresh interval (in seconds: 5, 10, 30, or 0 = Off)
const autoRefreshInterval = ref<number>(5)
let pollingTimer: number | null = null

function refreshActiveWorkload() {
  switch (activeTab.value) {
    case 'pods':
      k8sStore.fetchPods()
      k8sStore.fetchPodMetrics()
      break
    case 'deployments':
      k8sStore.fetchDeployments()
      break
    case 'statefulsets':
      k8sStore.fetchStatefulSets()
      break
    case 'daemonsets':
      k8sStore.fetchDaemonSets()
      break
    case 'hpas':
      k8sStore.fetchHPAs()
      break
  }
}

function startPolling() {
  stopPolling()
  if (autoRefreshInterval.value <= 0) return
  pollingTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') {
      refreshActiveWorkload()
    }
  }, autoRefreshInterval.value * 1000)
}

function stopPolling() {
  if (pollingTimer !== null) {
    window.clearInterval(pollingTimer)
    pollingTimer = null
  }
}

function setAutoRefresh(sec: number) {
  autoRefreshInterval.value = sec
  startPolling()
}

function onVisibilityChange() {
  if (document.visibilityState === 'visible') {
    refreshActiveWorkload()
  }
}

onMounted(() => {
  fetchAllWorkloads()
  startPolling()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onBeforeUnmount(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', onVisibilityChange)
})

watch(selectedNamespace, () => {
  fetchAllWorkloads()
})

watch(activeTab, () => {
  refreshActiveWorkload()
})

function fetchAllWorkloads() {
  k8sStore.fetchDeployments()
  k8sStore.fetchStatefulSets()
  k8sStore.fetchDaemonSets()
  k8sStore.fetchPods()
  k8sStore.fetchPodMetrics()
  k8sStore.fetchHPAs()
}

// Quick scale for deployments
const quickScaleDeployment = async (item: DeploymentItem, newReplicas: number) => {
  if (newReplicas < 0) return
  isScaling.value[item.name] = true
  try {
    await k8sStore.scaleDeployment(item.name, newReplicas)
    const msg = `Scaled deployment to ${newReplicas} replicas`
    showNotification(item.name, msg)
    toast.add({
      severity: 'success',
      summary: 'Scale Success',
      detail: `${item.name}: ${msg}`,
      life: 3000
    })
  } catch (err: unknown) {
    toast.add({
      severity: 'error',
      summary: 'Scale Deployment Failed',
      detail: err instanceof Error ? err.message : 'Unknown error',
      life: 5000
    })
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
    const msg = `Scaled statefulset to ${newReplicas} replicas`
    showNotification(item.name, msg)
    toast.add({
      severity: 'success',
      summary: 'Scale Success',
      detail: `${item.name}: ${msg}`,
      life: 3000
    })
  } catch (err: unknown) {
    toast.add({
      severity: 'error',
      summary: 'Scale StatefulSet Failed',
      detail: err instanceof Error ? err.message : 'Unknown error',
      life: 5000
    })
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
      d.env_secrets.some((s) => s.toLowerCase().includes(q))
  )
})

const filteredStatefulSets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return statefulsets.value
  return statefulsets.value.filter(
    (s) => s.name.toLowerCase().includes(q) || s.images.some((img) => img.toLowerCase().includes(q))
  )
})

const filteredDaemonSets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return daemonsets.value
  return daemonsets.value.filter(
    (d) => d.name.toLowerCase().includes(q) || d.images.some((img) => img.toLowerCase().includes(q))
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
      (p.ip && p.ip.includes(q))
  )
})

const filteredHPAs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return hpas.value
  return hpas.value.filter(
    (h) =>
      h.name.toLowerCase().includes(q) ||
      h.target_name.toLowerCase().includes(q) ||
      h.target_kind.toLowerCase().includes(q)
  )
})

const hpaMap = computed<Record<string, HPAItem>>(() => {
  const map: Record<string, HPAItem> = {}
  for (const h of hpas.value) {
    map[`${h.target_kind}/${h.target_name}`] = h
    map[h.target_name] = h
  }
  return map
})

const kedaMap = computed<Record<string, KedaHTTPScaledObject>>(() => {
  const map: Record<string, KedaHTTPScaledObject> = {}
  for (const k of kedaHttpObjects.value) {
    map[`${k.target_kind}/${k.target_name}`] = k
    map[k.target_name] = k
  }
  return map
})

const filteredKedaObjects = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return kedaHttpObjects.value
  return kedaHttpObjects.value.filter(
    (k) =>
      k.name.toLowerCase().includes(q) ||
      k.target_name.toLowerCase().includes(q) ||
      k.target_service.toLowerCase().includes(q)
  )
})

function deleteKedaItem(item: KedaHTTPScaledObject) {
  confirm.require({
    message: `Are you sure you want to remove KEDA HTTP autoscaling for '${item.name}' (${item.target_name})?`,
    header: 'Delete KEDA Autoscaler',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Delete',
      severity: 'danger'
    },
    accept: async () => {
      try {
        await k8sStore.deleteKedaHTTPScaledObject(item.name, item.namespace)
        toast.add({
          severity: 'success',
          summary: 'Autoscaler Removed',
          detail: `KEDA Autoscaler '${item.name}' deleted successfully`,
          life: 3000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Delete Failed',
          detail: err instanceof Error ? err.message : 'Unknown error',
          life: 5000
        })
      }
    }
  })
}

function openHpaScale(
  name: string,
  kind: 'Deployment' | 'StatefulSet' = 'Deployment',
  currentReplicas: number = 1
) {
  selectedHpaTarget.value = { name, kind, currentReplicas }
  isHpaScaleOpen.value = true
}

function deleteHpaItem(item: HPAItem) {
  confirm.require({
    message: `Are you sure you want to remove autoscaling for '${item.name}' (${item.target_kind} ${item.target_name})?`,
    header: 'Delete HorizontalPodAutoscaler',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger text-xs font-semibold',
    rejectClass: 'p-button-secondary text-xs',
    accept: async () => {
      try {
        await k8sStore.deleteHPA(item.name, item.namespace)
        toast.add({
          severity: 'success',
          summary: 'Autoscaler Removed',
          detail: `HPA '${item.name}' deleted successfully`,
          life: 3000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Delete HPA Failed',
          detail: err instanceof Error ? err.message : 'Unknown error',
          life: 5000
        })
      }
    }
  })
}

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

const openRollbackDialog = (item: DeploymentItem) => {
  selectedDeploymentName.value = item.name
  isRollbackOpen.value = true
}

const openYamlModal = (kind: string, name: string) => {
  selectedYamlResource.value = {
    kind,
    name,
    namespace: selectedNamespace.value
  }
  isYamlOpen.value = true
}

const restartDeployment = (item: DeploymentItem) => {
  confirm.require({
    message: `Trigger rollout restart for deployment '${item.name}'? Pods will restart sequentially.`,
    header: 'Restart Deployment',
    icon: 'pi pi-refresh',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Restart',
      severity: 'warn'
    },
    accept: async () => {
      try {
        const res = await k8sStore.restartDeployment(item.name)
        const msg = res.message || 'Rollout restart initiated'
        showNotification(item.name, msg)
        toast.add({
          severity: 'info',
          summary: 'Rollout Restarted',
          detail: msg,
          life: 4000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Restart Failed',
          detail: err instanceof Error ? err.message : 'Failed to restart deployment',
          life: 5000
        })
      }
    }
  })
}

const restartStatefulSet = (item: StatefulSetItem) => {
  confirm.require({
    message: `Trigger rollout restart for statefulset '${item.name}'? Pods will restart sequentially.`,
    header: 'Restart StatefulSet',
    icon: 'pi pi-refresh',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Restart',
      severity: 'warn'
    },
    accept: async () => {
      try {
        await k8sStore.restartStatefulSet(item.name)
        const msg = `Rollout restart initiated for ${item.name}`
        showNotification(item.name, msg)
        toast.add({
          severity: 'info',
          summary: 'Rollout Restarted',
          detail: msg,
          life: 4000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Restart Failed',
          detail: err instanceof Error ? err.message : 'Failed to restart statefulset',
          life: 5000
        })
      }
    }
  })
}

const restartDaemonSet = (item: DaemonSetItem) => {
  confirm.require({
    message: `Trigger rollout restart for daemonset '${item.name}' across all nodes?`,
    header: 'Restart DaemonSet',
    icon: 'pi pi-refresh',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Restart',
      severity: 'warn'
    },
    accept: async () => {
      try {
        await k8sStore.restartDaemonSet(item.name)
        const msg = `Rollout restart initiated for ${item.name}`
        showNotification(item.name, msg)
        toast.add({
          severity: 'info',
          summary: 'Rollout Restarted',
          detail: msg,
          life: 4000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Restart Failed',
          detail: err instanceof Error ? err.message : 'Failed to restart daemonset',
          life: 5000
        })
      }
    }
  })
}

const deletePodConfirm = (pod: PodItem) => {
  confirm.require({
    message: `Delete / Redeploy pod '${pod.name}'? The controller will automatically recreate it.`,
    header: 'Redeploy Pod',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Redeploy',
      severity: 'danger'
    },
    accept: async () => {
      try {
        await k8sStore.deletePod(pod.name)
        const msg = `Pod '${pod.name}' terminated and recreating`
        showNotification(pod.name, msg)
        toast.add({
          severity: 'warn',
          summary: 'Pod Terminated',
          detail: msg,
          life: 4000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Redeploy Failed',
          detail: err instanceof Error ? err.message : 'Failed to delete pod',
          life: 5000
        })
      }
    }
  })
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
        <h1
          class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5"
        >
          <i class="pi pi-objects-column text-sky-500"></i>
          <span>Workload Management</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Monitor Deployments, StatefulSets, DaemonSets, and inspect/redeploy Pods in
          <strong class="text-slate-700 dark:text-slate-300 font-mono">{{
            selectedNamespace
          }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          v-if="canMutate"
          label="Deploy Workload"
          icon="pi pi-cloud-upload"
          size="small"
          class="btn-sky text-xs font-semibold shadow-md cursor-pointer px-3 py-1.5"
          @click="isDeployOpen = true"
        />

        <button
          type="button"
          class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition flex items-center gap-1.5 cursor-pointer"
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
        <span
          ><strong>{{ restartNotification.title }}:</strong> {{ restartNotification.message }}</span
        >
      </div>
      <button class="opacity-70 hover:opacity-100" @click="restartNotification = null">
        <i class="pi pi-times"></i>
      </button>
    </div>

    <!-- Read-Only Notice Banner -->
    <div
      v-if="!canMutate"
      class="p-3.5 rounded-xl bg-slate-900/80 border border-slate-800 flex items-center justify-between text-xs text-slate-300 shadow-sm"
    >
      <div class="flex items-center gap-2.5">
        <i class="pi pi-shield text-amber-400 text-sm"></i>
        <span>
          <b class="text-white uppercase font-mono">{{ authStore.userRole }}</b> Mode:
          {{
            authStore.isViewer
              ? 'Viewer role is restricted to read-only observability.'
              : 'Namespace ' + selectedNamespace + ' is outside your DevOps allowed boundary.'
          }}
          Mutating actions (scale, restart, edit, delete, shell) are disabled.
        </span>
      </div>
      <span
        class="px-2.5 py-0.5 rounded text-[10px] font-mono bg-slate-800 text-slate-400 border border-slate-700"
      >
        Read-Only
      </span>
    </div>

    <!-- Workload Navigation Tabs -->
    <div class="flex items-center gap-2 border-b border-slate-200 dark:border-slate-800 pb-2">
      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2"
        :class="
          activeTab === 'deployments'
            ? 'bg-sky-500/15 text-sky-400 border border-sky-500/30 shadow-sm'
            : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'
        "
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
        :class="
          activeTab === 'statefulsets'
            ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30 shadow-sm'
            : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'
        "
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
        :class="
          activeTab === 'daemonsets'
            ? 'bg-teal-500/15 text-teal-400 border border-teal-500/30 shadow-sm'
            : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'
        "
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
        :class="
          activeTab === 'pods'
            ? 'bg-purple-500/15 text-purple-400 border border-purple-500/30 shadow-sm'
            : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'
        "
        @click="activeTab = 'pods'"
      >
        <i class="pi pi-box text-xs"></i>
        <span>Pods Deep-Dive</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-purple-500/20 text-purple-300">
          {{ pods.length }}
        </span>
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-xl text-xs font-semibold transition flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === 'hpas'
            ? 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/30 shadow-sm'
            : 'text-slate-600 dark:text-slate-400 hover:text-white hover:bg-slate-800/40'
        "
        @click="activeTab = 'hpas'"
      >
        <i class="pi pi-bolt text-xs text-cyan-400"></i>
        <span>Autoscalers (KEDA & HPA)</span>
        <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-cyan-500/20 text-cyan-300 font-bold">
          {{ kedaHttpObjects.length + hpas.length }}
        </span>
      </button>
    </div>

    <!-- Search & Live Sync Toolbar -->
    <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
      <IconField icon-position="left" class="w-full sm:w-80">
        <InputIcon class="pi pi-search text-xs" />
        <InputText
          v-model="searchQuery"
          :placeholder="`Search ${activeTab}...`"
          class="w-full text-xs rounded-xl bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 py-2 text-slate-800 dark:text-slate-200"
        />
      </IconField>

      <!-- Live Sync & Auto-Refresh Controls -->
      <div class="flex items-center gap-2 self-end sm:self-auto">
        <!-- Live Status Pill -->
        <div
          class="flex items-center gap-2 px-2.5 py-1.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs shadow-xs"
        >
          <span class="relative flex h-2 w-2">
            <span
              class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"
              :class="
                isConnected
                  ? 'bg-emerald-400'
                  : autoRefreshInterval > 0
                    ? 'bg-sky-400'
                    : 'bg-slate-400'
              "
            ></span>
            <span
              class="relative inline-flex rounded-full h-2 w-2"
              :class="
                isConnected
                  ? 'bg-emerald-500'
                  : autoRefreshInterval > 0
                    ? 'bg-sky-500'
                    : 'bg-slate-500'
              "
            ></span>
          </span>
          <span class="text-[11px] font-mono font-medium text-slate-600 dark:text-slate-300">
            {{
              isConnected ? 'Live Stream' : autoRefreshInterval > 0 ? 'Live Polling' : 'Sync Paused'
            }}
          </span>
        </div>

        <!-- Interval Selector -->
        <div
          class="flex items-center bg-slate-100 dark:bg-slate-800/80 p-0.5 rounded-xl border border-slate-200 dark:border-slate-700/60 text-xs"
        >
          <button
            v-for="sec in [5, 10, 30, 0]"
            :key="sec"
            type="button"
            class="px-2 py-1 rounded-lg text-[11px] font-mono transition font-medium cursor-pointer"
            :class="
              autoRefreshInterval === sec
                ? 'bg-white dark:bg-slate-700 text-sky-500 dark:text-sky-400 shadow-xs'
                : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
            "
            @click="setAutoRefresh(sec)"
          >
            {{ sec === 0 ? 'Off' : `${sec}s` }}
          </button>
        </div>

        <!-- Manual Refresh Button -->
        <Button
          icon="pi pi-refresh"
          size="small"
          severity="secondary"
          outlined
          :loading="isLoading"
          title="Force refresh now"
          class="text-xs px-2.5 py-1.5 rounded-xl cursor-pointer"
          @click="fetchAllWorkloads"
        />
      </div>
    </div>

    <!-- TAB 1: Deployments -->
    <div
      v-if="activeTab === 'deployments'"
      class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90"
    >
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
              <div
                class="w-7 h-7 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400 shrink-0"
              >
                <i class="pi pi-server text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{
                  data.name
                }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Replicas & Quick Scale -->
        <Column field="ready_replicas" header="Replicas" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2">
              <span
                class="px-2 py-0.5 rounded text-xs font-mono font-bold shrink-0"
                :class="
                  data.ready_replicas === data.replicas
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                    : 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
                "
              >
                {{ data.ready_replicas }}/{{ data.replicas }}
              </span>

              <!-- KEDA HTTP badge if active -->
              <span
                v-if="kedaMap[data.name]"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 flex items-center gap-1 cursor-pointer hover:bg-cyan-500/20 transition shrink-0"
                title="KEDA HTTP Autoscaling active: click to configure"
                @click="openHpaScale(data.name, 'Deployment', data.replicas)"
              >
                <i class="pi pi-bolt text-[9px] text-cyan-400"></i>
                <span>{{ kedaMap[data.name].min_replicas }}-{{ kedaMap[data.name].max_replicas }} (KEDA)</span>
              </span>

              <!-- HPA badge if active -->
              <span
                v-else-if="hpaMap[data.name]"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-sky-500/10 text-sky-400 border border-sky-500/30 flex items-center gap-1 cursor-pointer hover:bg-sky-500/20 transition shrink-0"
                title="HPA Autoscaling active: click to configure"
                @click="openHpaScale(data.name, 'Deployment', data.replicas)"
              >
                <i class="pi pi-sliders-h text-[9px]"></i>
                <span>{{ hpaMap[data.name].min_replicas }}-{{ hpaMap[data.name].max_replicas }} (HPA)</span>
              </span>

              <!-- Quick scale buttons -->
              <div
                class="flex items-center gap-1 bg-slate-800/80 p-0.5 rounded-lg border border-slate-700 shrink-0"
              >
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30 disabled:cursor-not-allowed"
                  :disabled="!canMutate || isScaling[data.name] || data.replicas <= 0"
                  :title="!canMutate ? 'Read-only: insufficient permissions' : 'Scale down (-1)'"
                  @click="quickScaleDeployment(data, data.replicas - 1)"
                >
                  <i class="pi pi-minus"></i>
                </button>
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30 disabled:cursor-not-allowed"
                  :disabled="!canMutate || isScaling[data.name]"
                  :title="!canMutate ? 'Read-only: insufficient permissions' : 'Scale up (+1)'"
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
              <span v-if="data.env_secrets.length === 0" class="text-xs text-slate-500 font-mono"
                >None</span
              >
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
        <Column header="Actions" align-frozen="right" style="min-width: 27rem; text-align: right">
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

              <!-- Autoscale / HPA -->
              <Button
                label="Scale"
                icon="pi pi-sliders-h"
                size="small"
                class="btn-cyan text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!canMutate"
                :title="
                  !canMutate ? 'Read-only: cannot scale deployment' : 'Configure Max/Min Autoscaling (HPA)'
                "
                @click="openHpaScale(data.name, 'Deployment', data.replicas)"
              />

              <!-- Edit -->
              <Button
                label="Edit"
                icon="pi pi-file-edit"
                size="small"
                class="btn-blue text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!canMutate"
                :title="
                  !canMutate ? 'Read-only: cannot edit deployment' : 'Edit replicas & containers'
                "
                @click="openEditor(data)"
              />

              <!-- Rollback / History -->
              <Button
                label="Rollback"
                icon="pi pi-undo"
                size="small"
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="View rollout history & rollback revision"
                @click="openRollbackDialog(data)"
              />

              <!-- Rollout Restart -->
              <Button
                icon="pi pi-refresh"
                size="small"
                class="btn-rose text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :loading="isActionLoading"
                :disabled="!canMutate"
                :title="
                  !canMutate ? 'Read-only: cannot restart deployment' : 'Trigger Rollout Restart'
                "
                @click="restartDeployment(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-server text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No Deployments Found</h3>
            <p class="text-xs text-slate-500 mt-1">
              No deployments found matching filter in {{ selectedNamespace }}.
            </p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 2: StatefulSets -->
    <div
      v-if="activeTab === 'statefulsets'"
      class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90"
    >
      <DataTable
        :value="filteredStatefulSets"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <Column field="name" header="Name" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div
                class="w-7 h-7 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400 shrink-0"
              >
                <i class="pi pi-database text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{
                  data.name
                }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <Column field="ready_replicas" header="Replicas" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2">
              <span
                class="px-2 py-0.5 rounded text-xs font-mono font-bold shrink-0"
                :class="
                  data.ready_replicas === data.replicas
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                    : 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
                "
              >
                {{ data.ready_replicas }}/{{ data.replicas }}
              </span>

              <!-- HPA badge if active -->
              <span
                v-if="hpaMap[data.name]"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 flex items-center gap-1 cursor-pointer hover:bg-cyan-500/20 transition shrink-0"
                title="Autoscaling active: click to manage max/min scaling"
                @click="openHpaScale(data.name, 'StatefulSet', data.replicas)"
              >
                <i class="pi pi-sliders-h text-[9px]"></i>
                <span>{{ hpaMap[data.name].min_replicas }}-{{ hpaMap[data.name].max_replicas }}</span>
              </span>

              <!-- Quick scale -->
              <div
                class="flex items-center gap-1 bg-slate-800/80 p-0.5 rounded-lg border border-slate-700 shrink-0"
              >
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30 disabled:cursor-not-allowed"
                  :disabled="!canMutate || isScaling[data.name] || data.replicas <= 0"
                  :title="!canMutate ? 'Read-only: insufficient permissions' : 'Scale down (-1)'"
                  @click="quickScaleStatefulSet(data, data.replicas - 1)"
                >
                  <i class="pi pi-minus"></i>
                </button>
                <button
                  type="button"
                  class="w-5 h-5 rounded flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition text-[10px] disabled:opacity-30 disabled:cursor-not-allowed"
                  :disabled="!canMutate || isScaling[data.name]"
                  :title="!canMutate ? 'Read-only: insufficient permissions' : 'Scale up (+1)'"
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

        <Column header="Actions" align-frozen="right" style="min-width: 19rem; text-align: right">
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

              <!-- Autoscale / HPA -->
              <Button
                label="Scale"
                icon="pi pi-sliders-h"
                size="small"
                class="btn-cyan text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!canMutate"
                :title="
                  !canMutate
                    ? 'Read-only: cannot scale statefulset'
                    : 'Configure Max/Min Autoscaling (HPA)'
                "
                @click="openHpaScale(data.name, 'StatefulSet', data.replicas)"
              />

              <Button
                label="Restart"
                icon="pi pi-refresh"
                size="small"
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :loading="isActionLoading"
                :disabled="!canMutate"
                :title="!canMutate ? 'Read-only: cannot restart statefulset' : 'Rollout Restart'"
                @click="restartStatefulSet(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-database text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No StatefulSets Found</h3>
            <p class="text-xs text-slate-500 mt-1">
              There are no StatefulSets in {{ selectedNamespace }}.
            </p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 3: DaemonSets -->
    <div
      v-if="activeTab === 'daemonsets'"
      class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90"
    >
      <DataTable
        :value="filteredDaemonSets"
        :loading="isLoading"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <Column field="name" header="Name" sortable style="min-width: 14rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2.5">
              <div
                class="w-7 h-7 rounded-lg bg-teal-500/10 border border-teal-500/30 flex items-center justify-center text-teal-400 shrink-0"
              >
                <i class="pi pi-clone text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs">{{
                  data.name
                }}</span>
                <div class="text-[11px] text-slate-400 font-mono">{{ data.namespace }}</div>
              </div>
            </div>
          </template>
        </Column>

        <Column header="Pod Status" style="min-width: 12rem">
          <template #body="{ data }">
            <div class="flex items-center gap-2 text-xs font-mono">
              <span
                class="px-2 py-0.5 rounded bg-teal-500/10 text-teal-300 border border-teal-500/30"
              >
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
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :loading="isActionLoading"
                :disabled="!canMutate"
                :title="!canMutate ? 'Read-only: cannot restart daemonset' : 'Rollout Restart'"
                @click="restartDaemonSet(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-clone text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No DaemonSets Found</h3>
            <p class="text-xs text-slate-500 mt-1">
              There are no DaemonSets in {{ selectedNamespace }}.
            </p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 4: Pods Deep-Dive -->
    <div
      v-if="activeTab === 'pods'"
      class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90"
    >
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
              <div
                class="w-7 h-7 rounded-lg bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400 shrink-0"
              >
                <i class="pi pi-box text-xs"></i>
              </div>
              <div>
                <span class="font-bold text-slate-900 dark:text-slate-100 text-xs font-mono">{{
                  data.name
                }}</span>
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
              :class="
                data.restarts > 0
                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                  : 'text-slate-400'
              "
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
                  <span class="text-slate-400"
                    >CPU: <b class="text-slate-200">{{ podMetrics[data.name]?.cpu_usage }}</b></span
                  >
                  <span
                    class="font-bold"
                    :class="getUsageColor(podMetrics[data.name]?.cpu_percent || 0)"
                  >
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
                  <span class="text-slate-400"
                    >Mem:
                    <b class="text-slate-200">{{ podMetrics[data.name]?.memory_usage }}</b></span
                  >
                  <span
                    class="font-bold"
                    :class="getUsageColor(podMetrics[data.name]?.memory_percent || 0)"
                  >
                    {{ Math.round(podMetrics[data.name]?.memory_percent || 0) }}%
                  </span>
                </div>
                <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="getUsageBarColor(podMetrics[data.name]?.memory_percent || 0)"
                    :style="{
                      width: `${Math.min(100, podMetrics[data.name]?.memory_percent || 10)}%`
                    }"
                  ></div>
                </div>
              </div>
            </div>
            <div v-else class="text-[11px] font-mono text-slate-500 italic">Telemetry sync...</div>
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
                class="btn-emerald text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!canMutate"
                :title="
                  !canMutate
                    ? 'Read-only: terminal shell restricted'
                    : 'Open interactive in-browser shell'
                "
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
                class="btn-rose text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                :loading="isActionLoading"
                :disabled="!canMutate"
                :title="
                  !canMutate
                    ? 'Read-only: cannot delete or redeploy pod'
                    : 'Redeploy / Delete Pod (triggers restart)'
                "
                @click="deletePodConfirm(data)"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-box text-3xl mb-2 text-slate-500"></i>
            <h3 class="font-semibold text-slate-200">No Pods Found</h3>
            <p class="text-xs text-slate-500 mt-1">
              There are no pods matching filter in {{ selectedNamespace }}.
            </p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 5: Autoscalers (KEDA & HPA) -->
    <div v-if="activeTab === 'hpas'" class="space-y-6">
      <!-- Section 1: KEDA HTTP Autoscalers (Per-Request / Concurrency) -->
      <div class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
        <div class="px-5 py-3.5 bg-slate-50 dark:bg-slate-800/40 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="w-7 h-7 rounded-lg bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
              <i class="pi pi-bolt text-xs"></i>
            </div>
            <div>
              <h3 class="text-xs font-bold text-slate-800 dark:text-slate-100 uppercase tracking-wider font-mono">
                KEDA HTTP Autoscalers (Per-Request & Scale-to-Zero)
              </h3>
              <p class="text-[11px] text-slate-400">Event-driven scaling based on HTTP request concurrency</p>
            </div>
          </div>
          <Tag :value="`${filteredKedaObjects.length} Active`" severity="info" class="text-[10px] font-mono px-2 py-0.5" />
        </div>

        <DataTable
          :value="filteredKedaObjects"
          :loading="isLoading"
          responsive-layout="scroll"
          class="p-datatable-sm"
        >
          <!-- Name & Workload -->
          <Column field="name" header="Autoscaler / Target Workload" sortable style="min-width: 16rem">
            <template #body="{ data }">
              <div class="flex items-center gap-2.5">
                <div
                  class="w-8 h-8 rounded-xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400 shrink-0"
                >
                  <i class="pi pi-bolt text-sm"></i>
                </div>
                <div>
                  <div class="font-bold text-slate-900 dark:text-slate-100 text-xs font-mono">
                    {{ data.name }}
                  </div>
                  <div class="text-[11px] text-slate-400 flex items-center gap-1.5 mt-0.5 font-mono">
                    <span>Workload:</span>
                    <span class="px-1.5 py-0.2 rounded bg-slate-800 text-cyan-300 font-semibold text-[10px]">
                      {{ data.target_workload || `${data.target_kind}/${data.target_name}` }}
                    </span>
                  </div>
                </div>
              </div>
            </template>
          </Column>

          <!-- Scaling Range -->
          <Column header="Scaling Range (Min - Max)" style="min-width: 14rem">
            <template #body="{ data }">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span
                    class="px-2 py-0.5 rounded text-xs font-mono font-bold bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 flex items-center gap-1"
                  >
                    <span>{{ data.min_replicas }} - {{ data.max_replicas }} Pods</span>
                  </span>
                  <span
                    v-if="data.min_replicas === 0"
                    class="text-[9px] font-mono px-1.5 py-0.2 rounded bg-emerald-500/20 text-emerald-400 font-bold border border-emerald-500/30 uppercase"
                  >
                    Scale-to-Zero
                  </span>
                </div>
                <div class="text-[11px] text-slate-400 font-mono">
                  Cooldown: <b class="text-slate-200">{{ data.scaledown_period }}s</b>
                </div>
              </div>
            </template>
          </Column>

          <!-- Metric Trigger -->
          <Column header="Request Metric Trigger" style="min-width: 14rem">
            <template #body="{ data }">
              <div class="space-y-1 text-xs font-mono">
                <div class="flex items-center gap-1.5 text-cyan-400 font-semibold">
                  <i class="pi pi-sliders-v text-[10px]"></i>
                  <span>{{ data.concurrency ?? 30 }} req/pod target</span>
                </div>
                <div class="text-[11px] text-slate-400">
                  Service: <b class="text-slate-300">{{ data.target_service }}:{{ data.target_port }}</b>
                </div>
              </div>
            </template>
          </Column>

          <!-- Status -->
          <Column header="Status" style="min-width: 8rem">
            <template #body="{ data }">
              <Tag
                :value="data.ready ? 'Ready / Active' : 'Initializing'"
                :severity="data.ready ? 'success' : 'warn'"
                class="text-[10px] font-mono px-2 py-0.5"
              />
            </template>
          </Column>

          <!-- Age -->
          <Column field="age" header="Age" sortable style="min-width: 6rem">
            <template #body="{ data }">
              <span class="text-xs text-slate-400 font-mono">{{ data.age }}</span>
            </template>
          </Column>

          <!-- Actions -->
          <Column header="Actions" align-frozen="right" style="min-width: 14rem; text-align: right">
            <template #body="{ data }">
              <div class="flex items-center justify-end gap-1.5">
                <Button
                  label="Configure"
                  icon="pi pi-sliders-h"
                  size="small"
                  class="btn-cyan text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!canMutate"
                  title="Configure KEDA Autoscaler"
                  @click="openHpaScale(data.target_name, data.target_kind || 'Deployment', 1)"
                />
                <Button
                  label="YAML"
                  icon="pi pi-code"
                  size="small"
                  class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                  title="View YAML Manifest"
                  @click="openYamlModal('HTTPScaledObject', data.name)"
                />
                <Button
                  icon="pi pi-trash"
                  size="small"
                  class="btn-rose text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!canMutate"
                  title="Delete KEDA Autoscaler"
                  @click="deleteKedaItem(data)"
                />
              </div>
            </template>
          </Column>

          <template #empty>
            <div class="py-8 text-center text-slate-400">
              <i class="pi pi-bolt text-2xl mb-1 text-slate-500"></i>
              <p class="text-xs text-slate-400 font-mono">No KEDA HTTP Autoscalers in {{ selectedNamespace }}</p>
            </div>
          </template>
        </DataTable>
      </div>

      <!-- Section 2: Standard HPA (Resource Metrics) -->
      <div class="border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-white dark:bg-slate-900/90">
        <div class="px-5 py-3.5 bg-slate-50 dark:bg-slate-800/40 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="w-7 h-7 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
              <i class="pi pi-chart-line text-xs"></i>
            </div>
            <div>
              <h3 class="text-xs font-bold text-slate-800 dark:text-slate-100 uppercase tracking-wider font-mono">
                Standard HPA (CPU / Memory Resource Metrics)
              </h3>
              <p class="text-[11px] text-slate-400">Kubernetes HorizontalPodAutoscalers</p>
            </div>
          </div>
          <Tag :value="`${filteredHPAs.length} Active`" severity="info" class="text-[10px] font-mono px-2 py-0.5" />
        </div>

        <DataTable
          :value="filteredHPAs"
          :loading="isLoading"
          responsive-layout="scroll"
          class="p-datatable-sm"
        >
          <!-- HPA Name & Target -->
          <Column field="name" header="Autoscaler / Target Workload" sortable style="min-width: 16rem">
            <template #body="{ data }">
              <div class="flex items-center gap-2.5">
                <div
                  class="w-8 h-8 rounded-xl bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400 shrink-0"
                >
                  <i class="pi pi-sliders-h text-sm"></i>
                </div>
                <div>
                  <div class="font-bold text-slate-900 dark:text-slate-100 text-xs font-mono">
                    {{ data.name }}
                  </div>
                  <div class="text-[11px] text-slate-400 flex items-center gap-1.5 mt-0.5 font-mono">
                    <span>Target:</span>
                    <span class="px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 font-semibold text-[10px]">
                      {{ data.target_kind }}/{{ data.target_name }}
                    </span>
                  </div>
                </div>
              </div>
            </template>
          </Column>

          <!-- Min & Max Boundaries -->
          <Column header="Scaling Range (Min - Max)" style="min-width: 14rem">
            <template #body="{ data }">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span
                    class="px-2 py-0.5 rounded text-xs font-mono font-bold bg-sky-500/10 text-sky-400 border border-sky-500/30"
                  >
                    {{ data.min_replicas }} - {{ data.max_replicas }} Pods
                  </span>
                  <span class="text-[11px] text-slate-400 font-mono">
                    (Current: <b class="text-white">{{ data.current_replicas }}</b>)
                  </span>
                </div>
                <div class="w-36 h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-sky-500 rounded-full transition-all"
                    :style="{
                      width: `${Math.min(100, Math.max(10, (data.current_replicas / (data.max_replicas || 1)) * 100))}%`
                    }"
                  ></div>
                </div>
              </div>
            </template>
          </Column>

          <!-- Metrics Targets & Current -->
          <Column header="Target vs Current Metrics" style="min-width: 16rem">
            <template #body="{ data }">
              <div class="space-y-1 text-xs font-mono">
                <div v-if="data.target_cpu" class="flex items-center gap-2">
                  <span class="text-[11px] text-slate-400">CPU:</span>
                  <span
                    class="font-semibold"
                    :class="
                      data.current_cpu && data.current_cpu > data.target_cpu
                        ? 'text-amber-400'
                        : 'text-emerald-400'
                    "
                  >
                    {{ data.current_cpu !== undefined ? `${data.current_cpu}%` : 'N/A' }}
                  </span>
                  <span class="text-slate-500">/</span>
                  <span class="text-slate-300">{{ data.target_cpu }}% target</span>
                </div>
                <div v-if="data.target_memory" class="flex items-center gap-2">
                  <span class="text-[11px] text-slate-400">Mem:</span>
                  <span class="font-semibold text-emerald-400">
                    {{ data.current_memory !== undefined ? `${data.current_memory}%` : 'N/A' }}
                  </span>
                  <span class="text-slate-500">/</span>
                  <span class="text-slate-300">{{ data.target_memory }}% target</span>
                </div>
                <div v-if="!data.target_cpu && !data.target_memory" class="text-slate-500 text-[11px]">
                  No resource metrics configured
                </div>
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
          <Column header="Actions" align-frozen="right" style="min-width: 14rem; text-align: right">
            <template #body="{ data }">
              <div class="flex items-center justify-end gap-1.5">
                <Button
                  label="Configure"
                  icon="pi pi-sliders-h"
                  size="small"
                  class="btn-sky text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!canMutate"
                  title="Configure Max/Min Scaling"
                  @click="openHpaScale(data.target_name, data.target_kind || 'Deployment', data.current_replicas)"
                />
                <Button
                  label="YAML"
                  icon="pi pi-code"
                  size="small"
                  class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                  title="View HPA YAML"
                  @click="openYamlModal('HorizontalPodAutoscaler', data.name)"
                />
                <Button
                  icon="pi pi-trash"
                  size="small"
                  class="btn-rose text-xs px-2 py-1.5 rounded-lg active:scale-95 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!canMutate"
                  title="Delete HPA"
                  @click="deleteHpaItem(data)"
                />
              </div>
            </template>
          </Column>

          <template #empty>
            <div class="py-8 text-center text-slate-400">
              <i class="pi pi-sliders-h text-2xl mb-1 text-slate-500"></i>
              <p class="text-xs text-slate-400 font-mono">No standard Resource HPAs in {{ selectedNamespace }}</p>
            </div>
          </template>
        </DataTable>
      </div>
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

    <!-- Deployment Rollback & Revision History Modal -->
    <DeploymentRollbackDialog
      v-model:visible="isRollbackOpen"
      :deployment-name="selectedDeploymentName"
      :namespace="selectedNamespace"
      @rolled-back="k8sStore.fetchDeployments()"
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

    <!-- HPA Max/Min Scale Dialog Modal -->
    <HpaScaleDialog
      v-model:visible="isHpaScaleOpen"
      :workload-name="selectedHpaTarget.name"
      :workload-kind="selectedHpaTarget.kind"
      :namespace="selectedNamespace"
      :current-replicas="selectedHpaTarget.currentReplicas"
      @saved="fetchAllWorkloads"
    />

    <!-- Cloud Run Style Deploy Workload Dialog -->
    <DeployWorkloadDialog
      v-model:visible="isDeployOpen"
      :namespace="selectedNamespace"
      @deployed="
        (name) => {
          showNotification(name, `Workload '${name}' deployed successfully!`)
          toast.add({
            severity: 'success',
            summary: 'Workload Deployed',
            detail: `Workload '${name}' deployed successfully in Cloud Run style.`,
            life: 4000
          })
          fetchAllWorkloads()
        }
      "
    />
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
