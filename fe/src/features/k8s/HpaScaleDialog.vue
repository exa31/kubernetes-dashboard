<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { HPADetail, KedaHTTPScaledObject } from '@/types'
import { logger } from '@/utils'

const props = withDefaults(
  defineProps<{
    visible: boolean
    workloadName: string
    workloadKind?: 'Deployment' | 'StatefulSet'
    namespace: string
    currentReplicas?: number
  }>(),
  {
    workloadKind: 'Deployment',
    currentReplicas: 1
  }
)

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'saved'): void
}>()

const k8sStore = useK8sStore()
const toast = useToast()
const confirm = useConfirm()

// Active state
const scalingMode = ref<'keda' | 'hpa'>('keda')
const existingKeda = ref<KedaHTTPScaledObject | null>(null)
const existingHpa = ref<HPADetail | null>(null)
const isLoading = ref<boolean>(false)
const isSaving = ref<boolean>(false)
const isDeleting = ref<boolean>(false)
const errorMessage = ref<string | null>(null)

// KEDA HTTP Fields
const kedaMinReplicas = ref<number>(0)
const kedaMaxReplicas = ref<number>(3)
const kedaConcurrency = ref<number>(30)
const kedaScaledownPeriod = ref<number>(300)
const kedaTargetService = ref<string>('')
const kedaTargetPort = ref<number>(8080)
const kedaHosts = ref<string>('')

// HPA Fields
const hpaMinReplicas = ref<number>(1)
const hpaMaxReplicas = ref<number>(5)
const hpaTargetCpu = ref<number | null>(80)
const hpaTargetMemory = ref<number | null>(null)

const isAutoscalingActive = computed(() => !!existingKeda.value || !!existingHpa.value)

async function loadAutoscalerConfig() {
  if (!props.workloadName || !props.namespace) return
  isLoading.value = true
  errorMessage.value = null
  try {
    const res = await k8sStore.getWorkloadAutoscaler(
      props.workloadName,
      props.workloadKind,
      props.namespace
    )

    if (res.type === 'keda-http' && res.keda_http) {
      scalingMode.value = 'keda'
      existingKeda.value = res.keda_http
      existingHpa.value = null
      kedaMinReplicas.value = res.keda_http.min_replicas ?? 0
      kedaMaxReplicas.value = res.keda_http.max_replicas ?? 3
      kedaConcurrency.value = res.keda_http.concurrency ?? 30
      kedaScaledownPeriod.value = res.keda_http.scaledown_period ?? 300
      kedaTargetService.value = res.keda_http.target_service || props.workloadName
      kedaTargetPort.value = res.keda_http.target_port || 8080
      kedaHosts.value = (res.keda_http.hosts || []).join(', ')
    } else if (res.type === 'hpa' && res.hpa) {
      scalingMode.value = 'hpa'
      existingHpa.value = res.hpa
      existingKeda.value = null
      hpaMinReplicas.value = res.hpa.min_replicas ?? 1
      hpaMaxReplicas.value = res.hpa.max_replicas ?? 5
      hpaTargetCpu.value = res.hpa.target_cpu ?? 80
      hpaTargetMemory.value = res.hpa.target_memory ?? null
    } else {
      // Default to KEDA HTTP (User's cluster standard!)
      scalingMode.value = 'keda'
      existingKeda.value = null
      existingHpa.value = null
      kedaMinReplicas.value = 0 // Scale to Zero
      kedaMaxReplicas.value = Math.max(3, (props.currentReplicas || 1) * 3)
      kedaConcurrency.value = 30
      kedaScaledownPeriod.value = 300
      kedaTargetService.value = props.workloadName
      kedaTargetPort.value = 8080
      kedaHosts.value = ''

      hpaMinReplicas.value = Math.max(1, props.currentReplicas || 1)
      hpaMaxReplicas.value = Math.max(5, (props.currentReplicas || 1) * 3)
      hpaTargetCpu.value = 80
      hpaTargetMemory.value = null
    }
  } catch (err: unknown) {
    logger.error('Failed to load autoscaling config', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load autoscaling details'
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.visible,
  (open) => {
    if (open) {
      loadAutoscalerConfig()
    }
  }
)

function applyKedaPreset(min: number, max: number, concurrency: number, scaledown: number = 300) {
  kedaMinReplicas.value = min
  kedaMaxReplicas.value = max
  kedaConcurrency.value = concurrency
  kedaScaledownPeriod.value = scaledown
}

function applyHpaPreset(min: number, max: number, cpu: number, mem: number | null = null) {
  hpaMinReplicas.value = min
  hpaMaxReplicas.value = max
  hpaTargetCpu.value = cpu
  hpaTargetMemory.value = mem
}

async function handleSave() {
  isSaving.value = true
  errorMessage.value = null
  try {
    if (scalingMode.value === 'keda') {
      if (kedaMinReplicas.value < 0) {
        toast.add({
          severity: 'warn',
          summary: 'Validation Warning',
          detail: 'Min replicas cannot be negative',
          life: 3000
        })
        isSaving.value = false
        return
      }
      if (kedaMaxReplicas.value < kedaMinReplicas.value || kedaMaxReplicas.value < 1) {
        toast.add({
          severity: 'warn',
          summary: 'Validation Warning',
          detail: 'Max replicas must be at least 1 and >= Min replicas',
          life: 3000
        })
        isSaving.value = false
        return
      }

      const hostsArr = kedaHosts.value
        ? kedaHosts.value.split(',').map((h) => h.trim()).filter(Boolean)
        : []

      const payload = {
        name: existingKeda.value?.name || `${props.workloadName}-http-scale`,
        namespace: props.namespace,
        target_kind: props.workloadKind,
        target_name: props.workloadName,
        target_service: kedaTargetService.value || props.workloadName,
        target_port: kedaTargetPort.value || 80,
        min_replicas: kedaMinReplicas.value,
        max_replicas: kedaMaxReplicas.value,
        concurrency: kedaConcurrency.value || 30,
        scaledown_period: kedaScaledownPeriod.value || 300,
        hosts: hostsArr
      }

      await k8sStore.saveKedaHTTPScaledObject(payload)

      toast.add({
        severity: 'success',
        summary: 'KEDA HTTP Scaling Saved',
        detail: `Per-request autoscaling active (${kedaMinReplicas.value} - ${kedaMaxReplicas.value} pods, target ${kedaConcurrency.value} req/pod)`,
        life: 4000
      })
    } else {
      // HPA Mode
      if (hpaMinReplicas.value < 1) {
        toast.add({
          severity: 'warn',
          summary: 'Validation Warning',
          detail: 'Min replicas must be at least 1 for HPA',
          life: 3000
        })
        isSaving.value = false
        return
      }
      if (hpaMaxReplicas.value < hpaMinReplicas.value) {
        toast.add({
          severity: 'warn',
          summary: 'Validation Warning',
          detail: 'Max replicas must be greater than or equal to Min replicas',
          life: 3000
        })
        isSaving.value = false
        return
      }

      const payload = {
        name: existingHpa.value?.name || `${props.workloadName}-hpa`,
        namespace: props.namespace,
        target_kind: props.workloadKind,
        target_name: props.workloadName,
        min_replicas: hpaMinReplicas.value,
        max_replicas: hpaMaxReplicas.value,
        target_cpu: hpaTargetCpu.value ?? undefined,
        target_memory: hpaTargetMemory.value ?? undefined
      }

      await k8sStore.saveHPA(payload)

      toast.add({
        severity: 'success',
        summary: 'HPA Autoscaling Saved',
        detail: `Resource autoscaling active (${hpaMinReplicas.value} - ${hpaMaxReplicas.value} pods)`,
        life: 4000
      })
    }

    emit('saved')
    emit('update:visible', false)
  } catch (err: unknown) {
    logger.error('Failed to save autoscaling config', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to save autoscaling config'
  } finally {
    isSaving.value = false
  }
}

function confirmDeleteAutoscaler() {
  confirm.require({
    message: `Remove autoscaling for '${props.workloadName}'? The workload will revert to static replica scaling.`,
    header: 'Disable Autoscaling',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Disable Autoscaling',
      severity: 'danger'
    },
    accept: async () => {
      isDeleting.value = true
      try {
        if (existingKeda.value) {
          await k8sStore.deleteKedaHTTPScaledObject(existingKeda.value.name, props.namespace)
        }
        if (existingHpa.value) {
          await k8sStore.deleteHPA(existingHpa.value.name, props.namespace)
        }

        toast.add({
          severity: 'info',
          summary: 'Autoscaling Disabled',
          detail: `Autoscaling removed for ${props.workloadName}.`,
          life: 4000
        })
        existingKeda.value = null
        existingHpa.value = null
        emit('saved')
        emit('update:visible', false)
      } catch (err: unknown) {
        logger.error('Failed to delete autoscaler', err)
        toast.add({
          severity: 'error',
          summary: 'Failed to Disable',
          detail: err instanceof Error ? err.message : 'Unknown error',
          life: 5000
        })
      } finally {
        isDeleting.value = false
      }
    }
  })
}

function closeDialog() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[95vw] max-w-2xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
    }"
    @update:visible="(val) => emit('update:visible', val)"
  >
    <!-- Header -->
    <div
      class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-500 flex items-center justify-center font-bold text-lg shrink-0"
        >
          <i class="pi pi-bolt"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
              Autoscale Settings: {{ workloadName }}
            </h2>
            <Tag :value="workloadKind" severity="info" class="text-[10px] font-mono px-2 py-0.5" />
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Namespace:
            <span class="font-mono text-slate-700 dark:text-slate-300 font-semibold">{{
              namespace
            }}</span>
            &bull; Event-Driven (KEDA) & Resource Autoscaling
          </p>
        </div>
      </div>

      <Button
        icon="pi pi-times"
        severity="secondary"
        text
        rounded
        size="small"
        class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
        @click="closeDialog"
      />
    </div>

    <!-- Error Alert -->
    <div v-if="errorMessage" class="p-4 bg-rose-500/10 border-b border-rose-500/20">
      <Message severity="error" :closable="false" class="text-xs">
        {{ errorMessage }}
      </Message>
    </div>

    <!-- Body -->
    <div class="p-6 space-y-6 max-h-[75vh] overflow-y-auto bg-slate-50/50 dark:bg-slate-950">
      <!-- Loading State -->
      <div v-if="isLoading" class="py-12 text-center text-slate-400 space-y-3">
        <i class="pi pi-spin pi-spinner text-3xl text-cyan-500"></i>
        <p class="text-xs font-mono">Loading autoscaler details from cluster...</p>
      </div>

      <div v-else class="space-y-6">
        <!-- Status Banner -->
        <div
          class="p-4 rounded-xl border flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-xs"
          :class="
            isAutoscalingActive
              ? 'bg-emerald-500/10 border-emerald-500/30 dark:bg-emerald-950/20'
              : 'bg-slate-100 dark:bg-slate-800/40 border-slate-200 dark:border-slate-800'
          "
        >
          <div class="flex items-center gap-3">
            <span
              class="w-3 h-3 rounded-full flex shrink-0"
              :class="isAutoscalingActive ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"
            ></span>
            <div>
              <div class="text-xs font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
                <span>{{ isAutoscalingActive ? 'Autoscaling Active' : 'Autoscaling Disabled' }}</span>
                <span
                  v-if="existingKeda"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-cyan-500/20 text-cyan-400 font-bold border border-cyan-500/30"
                >
                  ⚡ KEDA HTTP: {{ existingKeda.min_replicas }} - {{ existingKeda.max_replicas }} Pods
                </span>
                <span
                  v-else-if="existingHpa"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-bold border border-emerald-500/30"
                >
                  📊 HPA: {{ existingHpa.min_replicas }} - {{ existingHpa.max_replicas }} Pods
                </span>
              </div>
              <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                {{
                  existingKeda
                    ? `Managed by KEDA HTTPScaledObject '${existingKeda.name}' with per-request concurrency scaling.`
                    : existingHpa
                    ? `Managed by Kubernetes HPA '${existingHpa.name}' with CPU/Memory metrics.`
                    : 'Currently using manual replica scaling without auto-scaling rules.'
                }}
              </p>
            </div>
          </div>

          <!-- Live metrics snapshot if active -->
          <div
            v-if="isAutoscalingActive"
            class="flex items-center gap-3 shrink-0 bg-white/50 dark:bg-slate-900/60 px-3 py-1.5 rounded-lg border border-emerald-500/20 text-xs font-mono"
          >
            <div>
              <span class="text-[10px] text-slate-400 block uppercase">Current Pods</span>
              <span class="font-bold text-emerald-400">{{ currentReplicas }}</span>
            </div>
            <div class="h-6 w-px bg-slate-700/40"></div>
            <div v-if="existingKeda">
              <span class="text-[10px] text-slate-400 block uppercase">Target Concurrency</span>
              <span class="font-bold text-cyan-400">{{ existingKeda.concurrency ?? 30 }} req/pod</span>
            </div>
            <div v-else-if="existingHpa?.target_cpu">
              <span class="text-[10px] text-slate-400 block uppercase">CPU Target</span>
              <span class="font-bold text-cyan-400">{{ existingHpa.target_cpu }}%</span>
            </div>
          </div>
        </div>

        <!-- Mode Switcher (KEDA HTTP vs HPA) -->
        <div class="flex items-center gap-2 p-1.5 rounded-xl bg-slate-200/70 dark:bg-slate-800/80 border border-slate-300 dark:border-slate-700">
          <button
            type="button"
            class="flex-1 py-2 px-3 rounded-lg text-xs font-semibold flex items-center justify-center gap-2 transition cursor-pointer"
            :class="
              scalingMode === 'keda'
                ? 'bg-white dark:bg-slate-900 text-cyan-500 dark:text-cyan-400 shadow-sm border border-slate-200 dark:border-slate-700'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
            "
            @click="scalingMode = 'keda'"
          >
            <i class="pi pi-bolt text-cyan-500"></i>
            <span>KEDA HTTP (Per-Request & Scale-to-Zero)</span>
            <span class="text-[9px] font-mono px-1.5 py-0.2 rounded bg-cyan-500/20 text-cyan-400 font-bold uppercase">Default</span>
          </button>

          <button
            type="button"
            class="flex-1 py-2 px-3 rounded-lg text-xs font-semibold flex items-center justify-center gap-2 transition cursor-pointer"
            :class="
              scalingMode === 'hpa'
                ? 'bg-white dark:bg-slate-900 text-sky-500 dark:text-sky-400 shadow-sm border border-slate-200 dark:border-slate-700'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
            "
            @click="scalingMode = 'hpa'"
          >
            <i class="pi pi-chart-line text-sky-500"></i>
            <span>Resource HPA (CPU / RAM %)</span>
          </button>
        </div>

        <!-- TAB 1: KEDA HTTP (Per-Request Scaling) -->
        <div v-if="scalingMode === 'keda'" class="space-y-5">
          <!-- Boundaries: Min / Max -->
          <div
            class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
          >
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <i class="pi pi-arrows-alt text-cyan-500"></i>
                Pod Boundaries (Scale Range)
              </h3>
              <span class="text-[11px] text-cyan-500 font-mono">Supports Scale to Zero (0 pods)</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Min Replicas (Allows 0!) -->
              <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800 min-w-0">
                <div class="flex items-center justify-between">
                  <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                    Min Replicas (Scale to 0)
                  </label>
                  <span v-if="kedaMinReplicas === 0" class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-cyan-500/20 text-cyan-400 font-bold">
                    Scale-to-Zero ON
                  </span>
                </div>
                <p class="text-[11px] text-slate-500 mt-0.5">Pod count when idle (0 = idle pods shut down).</p>

                <!-- Custom Stepper (No Overflow!) -->
                <div class="flex items-center gap-2 mt-3 w-full">
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="kedaMinReplicas <= 0"
                    @click="kedaMinReplicas = Math.max(0, kedaMinReplicas - 1)"
                  >
                    <i class="pi pi-minus text-xs"></i>
                  </button>
                  <input
                    v-model.number="kedaMinReplicas"
                    type="number"
                    min="0"
                    max="100"
                    class="flex-1 min-w-0 h-9 text-center font-mono font-bold text-sm rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                  />
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="kedaMinReplicas >= 100"
                    @click="kedaMinReplicas = Math.min(100, kedaMinReplicas + 1)"
                  >
                    <i class="pi pi-plus text-xs"></i>
                  </button>
                </div>
              </div>

              <!-- Max Replicas -->
              <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Max Replicas (Cap Limit)
                </label>
                <p class="text-[11px] text-slate-500 mt-0.5">Maximum pods allowed under heavy traffic.</p>

                <!-- Custom Stepper (No Overflow!) -->
                <div class="flex items-center gap-2 mt-3 w-full">
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="kedaMaxReplicas <= kedaMinReplicas"
                    @click="kedaMaxReplicas = Math.max(kedaMinReplicas, kedaMaxReplicas - 1)"
                  >
                    <i class="pi pi-minus text-xs"></i>
                  </button>
                  <input
                    v-model.number="kedaMaxReplicas"
                    type="number"
                    :min="kedaMinReplicas"
                    max="500"
                    class="flex-1 min-w-0 h-9 text-center font-mono font-bold text-sm rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                  />
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="kedaMaxReplicas >= 500"
                    @click="kedaMaxReplicas = Math.min(500, kedaMaxReplicas + 1)"
                  >
                    <i class="pi pi-plus text-xs"></i>
                  </button>
                </div>
              </div>
            </div>

            <!-- KEDA Presets -->
            <div class="pt-2 border-t border-slate-100 dark:border-slate-800/80 flex flex-wrap items-center gap-2">
              <span class="text-[11px] text-slate-500 font-medium mr-1">KEDA Presets:</span>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyKedaPreset(0, 3, 30, 300)"
              >
                ⚡ 0 - 3 pods (30 req/pod, Scale to 0)
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyKedaPreset(0, 5, 50, 300)"
              >
                ⚡ 0 - 5 pods (50 req/pod, Scale to 0)
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyKedaPreset(1, 10, 30, 300)"
              >
                ⚡ 1 - 10 pods (Always On)
              </button>
            </div>
          </div>

          <!-- Per-Request Triggers (Concurrency & Cooldown) -->
          <div
            class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
          >
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <i class="pi pi-sliders-v text-cyan-500"></i>
                Request Concurrency & Cooldown
              </h3>
              <span class="text-[11px] text-slate-500">HTTP Addon Metric Tuning</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Concurrency Target -->
              <div class="space-y-1.5 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Target Concurrency (Requests / Pod)
                </label>
                <p class="text-[11px] text-slate-500">When concurrent requests per pod exceed this, KEDA scales up.</p>
                <div class="flex items-center gap-2 mt-2">
                  <input
                    v-model.number="kedaConcurrency"
                    type="number"
                    min="1"
                    max="1000"
                    placeholder="30"
                    class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                  />
                  <span class="text-xs font-mono text-slate-400 shrink-0">req/pod</span>
                </div>
              </div>

              <!-- Cooldown Period -->
              <div class="space-y-1.5 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Idle Scale-down Delay (Cooldown)
                </label>
                <p class="text-[11px] text-slate-500">Seconds to wait after requests stop before scaling down/to-zero.</p>
                <div class="flex items-center gap-2 mt-2">
                  <input
                    v-model.number="kedaScaledownPeriod"
                    type="number"
                    min="10"
                    max="3600"
                    placeholder="300"
                    class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                  />
                  <span class="text-xs font-mono text-slate-400 shrink-0">sec (5 min)</span>
                </div>
              </div>
            </div>

            <!-- Target Service & Port Routing -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-3 border-t border-slate-100 dark:border-slate-800/80">
              <div class="space-y-1.5 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Target Service Name
                </label>
                <InputText
                  v-model="kedaTargetService"
                  :placeholder="workloadName"
                  class="w-full font-mono text-xs py-2"
                />
              </div>

              <div class="space-y-1.5 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Target Service Port
                </label>
                <input
                  v-model.number="kedaTargetPort"
                  type="number"
                  min="1"
                  max="65535"
                  placeholder="8080"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                />
              </div>
            </div>

            <!-- Optional Ingress Hosts -->
            <div class="space-y-1.5 min-w-0">
              <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                Routing Hosts (Optional Ingress Domains)
              </label>
              <InputText
                v-model="kedaHosts"
                placeholder="e.g. app.eka-dev.cloud, api.eka-dev.cloud (comma separated)"
                class="w-full font-mono text-xs py-2"
              />
              <p class="text-[11px] text-slate-500">KEDA HTTP Interceptor routes requests with these Host headers.</p>
            </div>
          </div>
        </div>

        <!-- TAB 2: Resource HPA (CPU / Memory) -->
        <div v-else class="space-y-5">
          <!-- Scaling Boundaries -->
          <div
            class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
          >
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <i class="pi pi-arrows-alt text-sky-500"></i>
                Pod Boundaries (HPA Limits)
              </h3>
              <span class="text-[11px] text-slate-500">Min pod count for HPA is 1</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Min Replicas -->
              <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Min Replicas
                </label>
                <p class="text-[11px] text-slate-500 mt-0.5">Cluster will never scale below this value.</p>
                <div class="flex items-center gap-2 mt-3 w-full">
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="hpaMinReplicas <= 1"
                    @click="hpaMinReplicas = Math.max(1, hpaMinReplicas - 1)"
                  >
                    <i class="pi pi-minus text-xs"></i>
                  </button>
                  <input
                    v-model.number="hpaMinReplicas"
                    type="number"
                    min="1"
                    max="100"
                    class="flex-1 min-w-0 h-9 text-center font-mono font-bold text-sm rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
                  />
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="hpaMinReplicas >= 100"
                    @click="hpaMinReplicas = Math.min(100, hpaMinReplicas + 1)"
                  >
                    <i class="pi pi-plus text-xs"></i>
                  </button>
                </div>
              </div>

              <!-- Max Replicas -->
              <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800 min-w-0">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Max Replicas
                </label>
                <p class="text-[11px] text-slate-500 mt-0.5">Cluster will never scale above this value.</p>
                <div class="flex items-center gap-2 mt-3 w-full">
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="hpaMaxReplicas <= hpaMinReplicas"
                    @click="hpaMaxReplicas = Math.max(hpaMinReplicas, hpaMaxReplicas - 1)"
                  >
                    <i class="pi pi-minus text-xs"></i>
                  </button>
                  <input
                    v-model.number="hpaMaxReplicas"
                    type="number"
                    :min="hpaMinReplicas"
                    max="500"
                    class="flex-1 min-w-0 h-9 text-center font-mono font-bold text-sm rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
                  />
                  <button
                    type="button"
                    class="w-9 h-9 rounded-lg flex items-center justify-center text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 transition cursor-pointer disabled:opacity-30 shrink-0"
                    :disabled="hpaMaxReplicas >= 500"
                    @click="hpaMaxReplicas = Math.min(500, hpaMaxReplicas + 1)"
                  >
                    <i class="pi pi-plus text-xs"></i>
                  </button>
                </div>
              </div>
            </div>

            <!-- Quick presets -->
            <div class="pt-2 border-t border-slate-100 dark:border-slate-800/80 flex flex-wrap items-center gap-2">
              <span class="text-[11px] text-slate-500 font-medium mr-1">HPA Presets:</span>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-sky-500/10 hover:text-sky-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyHpaPreset(1, 5, 80)"
              >
                1 - 5 pods (80% CPU)
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-sky-500/10 hover:text-sky-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyHpaPreset(2, 10, 75)"
              >
                2 - 10 pods (75% CPU)
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-sky-500/10 hover:text-sky-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
                @click="applyHpaPreset(3, 20, 70)"
              >
                3 - 20 pods (70% CPU)
              </button>
            </div>
          </div>

          <!-- Metric Targets (CPU & Memory) -->
          <div
            class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
          >
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <i class="pi pi-chart-line text-sky-500"></i>
                Autoscaling Metric Triggers
              </h3>
              <span class="text-[11px] text-slate-500">Scale out when average usage exceeds target</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Target CPU -->
              <div class="space-y-1.5 min-w-0">
                <div class="flex items-center justify-between">
                  <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                    Target CPU Utilization (%)
                  </label>
                  <span class="text-[10px] font-mono text-sky-500 font-semibold">Recommended</span>
                </div>
                <p class="text-[11px] text-slate-500">Target average CPU percentage across all pods.</p>
                <div class="flex items-center gap-2 mt-1">
                  <input
                    v-model.number="hpaTargetCpu"
                    type="number"
                    min="1"
                    max="100"
                    placeholder="80"
                    class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
                  />
                  <span class="text-xs font-mono text-slate-400">%</span>
                </div>
              </div>

              <!-- Target Memory -->
              <div class="space-y-1.5 min-w-0">
                <div class="flex items-center justify-between">
                  <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                    Target Memory Utilization (%)
                  </label>
                  <span class="text-[10px] font-mono text-slate-400">Optional</span>
                </div>
                <p class="text-[11px] text-slate-500">Leave empty to scale on CPU only.</p>
                <div class="flex items-center gap-2 mt-1">
                  <input
                    v-model.number="hpaTargetMemory"
                    type="number"
                    min="1"
                    max="100"
                    placeholder="Optional (e.g. 75)"
                    class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
                  />
                  <span class="text-xs font-mono text-slate-400">%</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Information Note -->
        <div class="flex items-start gap-2.5 p-3.5 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-xs text-cyan-400">
          <i class="pi pi-info-circle text-sm mt-0.5 shrink-0"></i>
          <p class="leading-relaxed text-[11px]">
            <strong class="font-semibold text-cyan-300">{{ scalingMode === 'keda' ? 'KEDA HTTP Scaling' : 'Kubernetes HPA' }}:</strong>
            {{
              scalingMode === 'keda'
                ? 'Monitors inbound HTTP traffic via KEDA Interceptor. When request concurrency rises, pods scale up to Max limit. If there is no traffic, pods safely scale down to 0 to save CPU and Memory.'
                : 'Monitors pod resource utilization (CPU & RAM). Scales replicas out when average load exceeds the target threshold.'
            }}
          </p>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="px-6 py-4 bg-slate-50 dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0"
    >
      <div>
        <Button
          v-if="isAutoscalingActive"
          label="Disable Autoscaling"
          icon="pi pi-trash"
          severity="danger"
          text
          size="small"
          class="text-xs font-semibold text-rose-400 hover:text-rose-300 cursor-pointer"
          :loading="isDeleting"
          @click="confirmDeleteAutoscaler"
        />
      </div>

      <div class="flex items-center gap-2">
        <Button
          label="Cancel"
          severity="secondary"
          text
          size="small"
          class="text-xs cursor-pointer"
          @click="closeDialog"
        />
        <Button
          :label="isAutoscalingActive ? 'Update Autoscaler' : 'Enable Autoscaling'"
          icon="pi pi-check"
          size="small"
          class="btn-sky text-xs font-semibold px-4 py-2 cursor-pointer"
          :loading="isSaving"
          @click="handleSave"
        />
      </div>
    </div>
  </Dialog>
</template>
