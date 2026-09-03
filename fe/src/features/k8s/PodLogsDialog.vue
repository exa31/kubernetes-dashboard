<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { PodItem } from '@/types'
import { logger } from '@/utils'

const props = defineProps<{
  visible: boolean
  deploymentName?: string
  podName?: string
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

const k8sStore = useK8sStore()

// State
const pods = ref<PodItem[]>([])
const selectedPod = ref<string>('')
const selectedContainer = ref<string>('')
const tailLines = ref<number>(250)
const showTimestamps = ref<boolean>(false)
const autoRefresh = ref<boolean>(false)
const logSearch = ref<string>('')
const logsText = ref<string>('')
const isLoadingPods = ref<boolean>(false)
const isLoadingLogs = ref<boolean>(false)
const terminalRef = ref<HTMLDivElement | null>(null)

let timerId: number | null = null

const tailOptions = [
  { label: '100 lines', value: 100 },
  { label: '250 lines', value: 250 },
  { label: '500 lines', value: 500 },
  { label: '1000 lines', value: 1000 },
]

const availableContainers = computed(() => {
  const current = pods.value.find((p) => p.name === selectedPod.value)
  return current?.containers || []
})

const filteredLines = computed(() => {
  if (!logsText.value) return []
  const allLines = logsText.value.split('\n')
  const q = logSearch.value.trim().toLowerCase()
  if (!q) return allLines
  return allLines.filter((l) => l.toLowerCase().includes(q))
})

async function loadPods() {
  if (!props.deploymentName) return
  isLoadingPods.value = true
  try {
    const list = await k8sStore.getDeploymentPods(props.deploymentName, props.namespace)
    pods.value = list
    if (list.length > 0 && !list.some((p) => p.name === selectedPod.value)) {
      selectedPod.value = list[0].name
      if (list[0].containers.length > 0) {
        selectedContainer.value = list[0].containers[0]
      }
    }
  } catch (err) {
    logger.error('Failed to load pods for deployment', err)
  } finally {
    isLoadingPods.value = false
  }
}

async function loadLogs() {
  if (!selectedPod.value) return
  isLoadingLogs.value = true
  try {
    const res = await k8sStore.getPodLogs(
      selectedPod.value,
      {
        container: selectedContainer.value,
        tail_lines: tailLines.value,
        timestamps: showTimestamps.value,
      },
      props.namespace,
    )
    logsText.value = res.logs || '(No log output returned)'
    await nextTick()
    scrollToBottom()
  } catch (err) {
    logger.error('Failed to fetch pod logs', err)
    logsText.value = `Error fetching logs: ${err instanceof Error ? err.message : 'Unknown error'}`
  } finally {
    isLoadingLogs.value = false
  }
}

function scrollToBottom() {
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

function copyLogs() {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(logsText.value)
  }
}

// Watchers
watch(
  () => [props.visible, props.podName, props.deploymentName],
  ([open]) => {
    if (open) {
      if (props.podName) {
        selectedPod.value = props.podName
        loadLogs()
      } else if (props.deploymentName) {
        loadPods().then(() => {
          loadLogs()
        })
      }
    } else {
      if (timerId !== null) {
        clearInterval(timerId)
        timerId = null
      }
    }
  },
)

watch(selectedPod, (newPod) => {
  const p = pods.value.find((item) => item.name === newPod)
  if (p && p.containers.length > 0) {
    selectedContainer.value = p.containers[0]
  }
  if (props.visible) {
    loadLogs()
  }
})

watch([selectedContainer, tailLines, showTimestamps], () => {
  if (props.visible) {
    loadLogs()
  }
})

watch(autoRefresh, (enabled) => {
  if (timerId !== null) {
    clearInterval(timerId)
    timerId = null
  }
  if (enabled) {
    timerId = window.setInterval(() => {
      if (props.visible) {
        loadLogs()
      }
    }, 3000)
  }
})

onBeforeUnmount(() => {
  if (timerId !== null) {
    clearInterval(timerId)
  }
})

function closeDialog() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[95vw] max-w-6xl h-[85vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 h-full flex flex-col overflow-hidden bg-slate-950 text-slate-100 rounded-xl' }
    }"
  >
    <!-- Header Controls -->
    <div class="p-4 bg-slate-900 border-b border-slate-800 flex flex-wrap items-center justify-between gap-3 shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-emerald-500/10 text-emerald-400 flex items-center justify-center font-bold text-sm">
          <i class="pi pi-terminal"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-bold text-sm text-slate-100 font-mono">{{ deploymentName || podName }}</h2>
            <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">{{ namespace }}</span>
          </div>
          <p class="text-[11px] text-slate-400 mt-0.5">Live streaming logs from application pods</p>
        </div>
      </div>

      <!-- Actions / Close -->
      <div class="flex items-center gap-2">
        <Button
          icon="pi pi-copy"
          label="Copy"
          severity="secondary"
          size="small"
          text
          class="text-xs text-slate-300 hover:text-white"
          @click="copyLogs"
        />
        <Button
          icon="pi pi-arrow-down"
          label="Bottom"
          severity="secondary"
          size="small"
          text
          class="text-xs text-slate-300 hover:text-white"
          @click="scrollToBottom"
        />
        <Button
          icon="pi pi-times"
          severity="secondary"
          text
          rounded
          size="small"
          class="text-slate-400 hover:text-white"
          @click="closeDialog"
        />
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="px-4 py-2.5 bg-slate-900/80 border-b border-slate-800/80 flex flex-wrap items-center justify-between gap-3 text-xs shrink-0">
      <div class="flex flex-wrap items-center gap-2.5">
        <!-- Pod Selector -->
        <div class="flex items-center gap-1.5">
          <span class="text-slate-400">Pod:</span>
          <Select
            v-model="selectedPod"
            :options="pods"
            option-label="name"
            option-value="name"
            placeholder="Select Pod"
            class="text-xs py-1 px-2 font-mono w-56 bg-slate-800 border-slate-700 text-slate-200"
            :loading="isLoadingPods"
          >
            <template #option="{ option }">
              <div class="flex items-center justify-between gap-2 font-mono text-xs w-full">
                <span>{{ option.name }}</span>
                <span
                  class="text-[10px] px-1.5 py-0.5 rounded font-bold"
                  :class="option.phase === 'Running' ? 'bg-emerald-950 text-emerald-400' : 'bg-amber-950 text-amber-400'"
                >
                  {{ option.phase }}
                </span>
              </div>
            </template>
          </Select>
        </div>

        <!-- Container Selector (if multiple) -->
        <div v-if="availableContainers.length > 1" class="flex items-center gap-1.5">
          <span class="text-slate-400">Container:</span>
          <Select
            v-model="selectedContainer"
            :options="availableContainers"
            class="text-xs py-1 px-2 font-mono w-40 bg-slate-800 border-slate-700 text-slate-200"
          />
        </div>

        <!-- Tail Lines -->
        <div class="flex items-center gap-1.5">
          <span class="text-slate-400">Lines:</span>
          <Select
            v-model="tailLines"
            :options="tailOptions"
            option-label="label"
            option-value="value"
            class="text-xs py-1 px-2 w-32 bg-slate-800 border-slate-700 text-slate-200"
          />
        </div>

        <!-- Timestamps toggle -->
        <button
          type="button"
          class="px-2.5 py-1.5 rounded text-xs font-mono flex items-center gap-1.5 border transition-colors cursor-pointer"
          :class="showTimestamps ? 'bg-sky-600/30 text-sky-300 border-sky-500/50' : 'bg-slate-800 text-slate-400 border-slate-700 hover:text-slate-200'"
          @click="showTimestamps = !showTimestamps"
        >
          <i class="pi pi-clock text-[10px]"></i>
          <span>Timestamps</span>
        </button>

        <!-- Auto Refresh toggle -->
        <button
          type="button"
          class="px-2.5 py-1.5 rounded text-xs font-mono flex items-center gap-1.5 border transition-colors cursor-pointer"
          :class="autoRefresh ? 'bg-emerald-600/30 text-emerald-300 border-emerald-500/50' : 'bg-slate-800 text-slate-400 border-slate-700 hover:text-slate-200'"
          @click="autoRefresh = !autoRefresh"
        >
          <span class="w-2 h-2 rounded-full" :class="autoRefresh ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'"></span>
          <span>Live (3s)</span>
        </button>

        <Button
          icon="pi pi-refresh"
          size="small"
          class="bg-sky-600 hover:bg-sky-500 text-white border-none text-xs px-2.5 py-1.5 rounded-lg shadow-xs cursor-pointer"
          :loading="isLoadingLogs"
          @click="loadLogs"
        />
      </div>

      <!-- Search in logs -->
      <IconField class="w-48">
        <InputIcon class="pi pi-search text-xs" />
        <InputText
          v-model="logSearch"
          placeholder="Filter logs..."
          class="w-full text-xs py-1 bg-slate-800 border-slate-700 text-slate-200 placeholder-slate-500 rounded-md"
        />
      </IconField>
    </div>

    <!-- Terminal Log Body -->
    <div
      ref="terminalRef"
      class="flex-1 overflow-auto p-4 font-mono text-xs text-slate-200 bg-slate-950 select-text leading-relaxed"
    >
      <div v-if="filteredLines.length === 0" class="text-slate-500 italic py-8 text-center">
        No log output found matching criteria.
      </div>

      <div
        v-for="(line, idx) in filteredLines"
        :key="idx"
        class="flex hover:bg-slate-900/60 rounded px-1.5 py-0.5 group"
      >
        <span class="select-none text-slate-600 group-hover:text-slate-400 w-12 text-right pr-3 shrink-0 font-mono">
          {{ idx + 1 }}
        </span>
        <span class="break-all whitespace-pre-wrap flex-1">{{ line }}</span>
      </div>
    </div>

    <!-- Footer Stats -->
    <div class="px-4 py-2 bg-slate-900 border-t border-slate-800 flex items-center justify-between text-[11px] text-slate-400 font-mono shrink-0">
      <div>
        <span>Showing {{ filteredLines.length }} lines</span>
        <span v-if="logSearch" class="ml-2 text-amber-400">(filtered by "{{ logSearch }}")</span>
      </div>
      <div>
        <span>Pod: <strong class="text-slate-200">{{ selectedPod || 'none' }}</strong></span>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
</style>
