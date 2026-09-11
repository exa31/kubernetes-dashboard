<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { HPADetail } from '@/types'
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

const existingHpa = ref<HPADetail | null>(null)
const isLoading = ref<boolean>(false)
const isSaving = ref<boolean>(false)
const isDeleting = ref<boolean>(false)
const errorMessage = ref<string | null>(null)

// Form fields
const minReplicas = ref<number>(1)
const maxReplicas = ref<number>(5)
const targetCpu = ref<number | null>(80)
const targetMemory = ref<number | null>(null)

const isAutoscalingActive = computed(() => !!existingHpa.value)

async function loadHpaConfig() {
  if (!props.workloadName || !props.namespace) return
  isLoading.value = true
  errorMessage.value = null
  try {
    const data = await k8sStore.getHPAForWorkload(
      props.workloadName,
      props.workloadKind,
      props.namespace
    )
    existingHpa.value = data
    if (data) {
      minReplicas.value = data.min_replicas ?? 1
      maxReplicas.value = data.max_replicas ?? 5
      targetCpu.value = data.target_cpu ?? 80
      targetMemory.value = data.target_memory ?? null
    } else {
      // Defaults for new configuration
      minReplicas.value = Math.max(1, props.currentReplicas || 1)
      maxReplicas.value = Math.max(5, (props.currentReplicas || 1) * 3)
      targetCpu.value = 80
      targetMemory.value = null
    }
  } catch (err: unknown) {
    logger.error('Failed to load HPA config', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load HPA details'
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.visible,
  (open) => {
    if (open) {
      loadHpaConfig()
    }
  }
)

function applyPreset(min: number, max: number, cpu: number, mem: number | null = null) {
  minReplicas.value = min
  maxReplicas.value = max
  targetCpu.value = cpu
  targetMemory.value = mem
}

async function handleSave() {
  if (minReplicas.value < 1) {
    toast.add({
      severity: 'warn',
      summary: 'Validation Warning',
      detail: 'Min replicas must be at least 1',
      life: 3000
    })
    return
  }

  if (maxReplicas.value < minReplicas.value) {
    toast.add({
      severity: 'warn',
      summary: 'Validation Warning',
      detail: 'Max replicas must be greater than or equal to Min replicas',
      life: 3000
    })
    return
  }

  isSaving.value = true
  errorMessage.value = null
  try {
    const payload = {
      name: existingHpa.value?.name || `${props.workloadName}-hpa`,
      namespace: props.namespace,
      target_kind: props.workloadKind,
      target_name: props.workloadName,
      min_replicas: minReplicas.value,
      max_replicas: maxReplicas.value,
      target_cpu: targetCpu.value ?? undefined,
      target_memory: targetMemory.value ?? undefined
    }

    await k8sStore.saveHPA(payload)

    toast.add({
      severity: 'success',
      summary: 'Autoscaling Updated',
      detail: `Max/Min scaling configured for ${props.workloadName} (${minReplicas.value} - ${maxReplicas.value} pods)`,
      life: 4000
    })

    emit('saved')
    emit('update:visible', false)
  } catch (err: unknown) {
    logger.error('Failed to save HPA', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to save autoscaling config'
  } finally {
    isSaving.value = false
  }
}

function confirmDeleteHpa() {
  if (!existingHpa.value) return

  confirm.require({
    message: `Are you sure you want to remove autoscaling for '${props.workloadName}'? The workload will revert to manual replica management.`,
    header: 'Disable Autoscaling',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger text-xs font-semibold',
    rejectClass: 'p-button-secondary text-xs',
    accept: async () => {
      if (!existingHpa.value) return
      isDeleting.value = true
      try {
        await k8sStore.deleteHPA(existingHpa.value.name, props.namespace)
        toast.add({
          severity: 'info',
          summary: 'Autoscaling Disabled',
          detail: `Autoscaling removed for ${props.workloadName}. Workload will maintain current replica count.`,
          life: 4000
        })
        existingHpa.value = null
        emit('saved')
        emit('update:visible', false)
      } catch (err: unknown) {
        logger.error('Failed to delete HPA', err)
        toast.add({
          severity: 'error',
          summary: 'Failed to Disable Autoscaling',
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
      class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-500 flex items-center justify-center font-bold text-lg shrink-0"
        >
          <i class="pi pi-sliders-h"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
              Max & Min Scaling (HPA): {{ workloadName }}
            </h2>
            <Tag :value="workloadKind" severity="info" class="text-[10px] font-mono px-2 py-0.5" />
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Namespace:
            <span class="font-mono text-slate-700 dark:text-slate-300 font-semibold">{{
              namespace
            }}</span>
            &bull; Configure Horizontal Pod Autoscaler boundaries & metric thresholds
          </p>
        </div>
      </div>

      <Button
        icon="pi pi-times"
        severity="secondary"
        text
        rounded
        size="small"
        class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
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
        <p class="text-xs font-mono">Reading HorizontalPodAutoscaler configuration...</p>
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
                <span>{{ isAutoscalingActive ? 'Autoscaling is Active' : 'Autoscaling is Disabled' }}</span>
                <span
                  v-if="isAutoscalingActive"
                  class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-emerald-500/20 text-emerald-400 font-bold"
                >
                  {{ existingHpa?.min_replicas }} - {{ existingHpa?.max_replicas }} Pods
                </span>
              </div>
              <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                {{
                  isAutoscalingActive
                    ? `Managed by HPA '${existingHpa?.name}'. Kubernetes dynamically adjusts pods between limits.`
                    : 'Currently using static/manual replica scaling.'
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
              <span class="font-bold text-emerald-400">{{ existingHpa?.current_replicas ?? currentReplicas }}</span>
            </div>
            <div class="h-6 w-px bg-slate-700/40"></div>
            <div v-if="existingHpa?.target_cpu">
              <span class="text-[10px] text-slate-400 block uppercase">CPU (Cur/Target)</span>
              <span class="font-bold text-cyan-400">
                {{ existingHpa.current_cpu !== undefined ? `${existingHpa.current_cpu}%` : 'N/A' }} / {{ existingHpa.target_cpu }}%
              </span>
            </div>
          </div>
        </div>

        <!-- Scaling Boundaries (Min / Max Replicas) -->
        <div
          class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
        >
          <div class="flex items-center justify-between">
            <h3 class="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <i class="pi pi-arrows-alt text-cyan-500"></i>
              Scaling Boundaries (Pod Limits)
            </h3>
            <span class="text-[11px] text-slate-500">Limits pod scaling range</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Min Replicas -->
            <div class="space-y-1.5 p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800">
              <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                Min Replicas (Minimum Pods)
              </label>
              <p class="text-[11px] text-slate-500">Cluster will never scale below this value.</p>
              <InputNumber
                v-model="minReplicas"
                show-buttons
                button-layout="horizontal"
                :min="1"
                :max="100"
                class="w-full font-mono text-sm mt-2"
                :pt="{
                  input: { class: 'text-center font-bold text-slate-900 dark:text-slate-100' }
                }"
              />
            </div>

            <!-- Max Replicas -->
            <div class="space-y-1.5 p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800">
              <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                Max Replicas (Maximum Pods)
              </label>
              <p class="text-[11px] text-slate-500">Cluster will never scale above this value.</p>
              <InputNumber
                v-model="maxReplicas"
                show-buttons
                button-layout="horizontal"
                :min="minReplicas"
                :max="500"
                class="w-full font-mono text-sm mt-2"
                :pt="{
                  input: { class: 'text-center font-bold text-slate-900 dark:text-slate-100' }
                }"
              />
            </div>
          </div>

          <!-- Quick presets -->
          <div class="pt-2 border-t border-slate-100 dark:border-slate-800/80 flex flex-wrap items-center gap-2">
            <span class="text-[11px] text-slate-500 font-medium mr-1">Quick Presets:</span>
            <button
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
              @click="applyPreset(1, 5, 80)"
            >
              1 - 5 pods (80% CPU)
            </button>
            <button
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
              @click="applyPreset(2, 10, 75)"
            >
              2 - 10 pods (75% CPU)
            </button>
            <button
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 hover:bg-cyan-500/10 hover:text-cyan-400 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition cursor-pointer"
              @click="applyPreset(3, 20, 70)"
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
              <i class="pi pi-chart-line text-cyan-500"></i>
              Autoscaling Metric Triggers
            </h3>
            <span class="text-[11px] text-slate-500">Scale out when average usage exceeds target</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Target CPU -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Target CPU Utilization (%)
                </label>
                <span class="text-[10px] font-mono text-cyan-500 font-semibold">Recommended</span>
              </div>
              <p class="text-[11px] text-slate-500">Target average CPU percentage across all pods.</p>
              <InputNumber
                v-model="targetCpu"
                :min="1"
                :max="100"
                suffix="%"
                placeholder="e.g. 80%"
                class="w-full font-mono text-sm"
              />
            </div>

            <!-- Target Memory -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
                  Target Memory Utilization (%)
                </label>
                <span class="text-[10px] font-mono text-slate-400">Optional</span>
              </div>
              <p class="text-[11px] text-slate-500">Leave empty to scale on CPU only.</p>
              <InputNumber
                v-model="targetMemory"
                :min="1"
                :max="100"
                suffix="%"
                placeholder="e.g. 75% (Optional)"
                class="w-full font-mono text-sm"
              />
            </div>
          </div>
        </div>

        <!-- Information Note -->
        <div class="flex items-start gap-2.5 p-3 rounded-lg bg-sky-500/10 border border-sky-500/20 text-xs text-sky-400">
          <i class="pi pi-info-circle text-sm mt-0.5 shrink-0"></i>
          <p class="leading-relaxed text-[11px]">
            The Horizontal Pod Autoscaler automatically increases or decreases the number of Pods in response to CPU/Memory utilization. If workload load spikes, K8s scales up to the <strong>Max ({{ maxReplicas }})</strong> limit. When traffic cools down, it safely scales back down to <strong>Min ({{ minReplicas }})</strong>.
          </p>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="px-6 py-4 bg-slate-50 dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between"
    >
      <div>
        <Button
          v-if="isAutoscalingActive"
          label="Disable Autoscaling"
          icon="pi pi-trash"
          severity="danger"
          text
          size="small"
          class="text-xs font-semibold text-rose-400 hover:text-rose-300"
          :loading="isDeleting"
          @click="confirmDeleteHpa"
        />
      </div>

      <div class="flex items-center gap-2">
        <Button
          label="Cancel"
          severity="secondary"
          text
          size="small"
          class="text-xs"
          @click="closeDialog"
        />
        <Button
          :label="isAutoscalingActive ? 'Update Scaling Rule' : 'Enable Max/Min Autoscaling'"
          icon="pi pi-check"
          size="small"
          class="btn-sky text-xs font-semibold px-4 py-2"
          :loading="isSaving"
          @click="handleSave"
        />
      </div>
    </div>
  </Dialog>
</template>
