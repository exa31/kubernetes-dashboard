<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import Dialog from 'primevue/dialog'
import { k8sApi } from '@/api'
import { useToast } from 'primevue/usetoast'

interface Props {
  visible: boolean
  kind: string
  name: string
  namespace?: string
}

const props = withDefaults(defineProps<Props>(), {
  namespace: 'default',
})

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'applied'): void
}>()

const toast = useToast()

const yamlContent = ref('')
const originalYaml = ref('')
const apiVersion = ref('')
const isLoading = ref(false)
const isApplying = ref(false)
const isValidating = ref(false)
const isEditing = ref(false)
const activeTab = ref<'editor' | 'diff'>('editor')
const validationResult = ref<{ status: 'success' | 'error'; message: string } | null>(null)

interface DiffLine {
  type: 'added' | 'removed' | 'unchanged'
  text: string
  oldNum?: number
  newNum?: number
}

const diffLines = computed<DiffLine[]>(() => {
  const oldLines = originalYaml.value.split('\n')
  const newLines = yamlContent.value.split('\n')
  const n = oldLines.length
  const m = newLines.length

  if (n + m > 3500) {
    return newLines.map((line, idx) => ({
      type: line === oldLines[idx] ? 'unchanged' : 'added',
      text: line,
      newNum: idx + 1,
    }))
  }

  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = 0; i < n; i++) {
    for (let j = 0; j < m; j++) {
      if (oldLines[i] === newLines[j]) {
        dp[i + 1][j + 1] = dp[i][j] + 1
      } else {
        dp[i + 1][j + 1] = Math.max(dp[i + 1][j], dp[i][j + 1])
      }
    }
  }

  const result: DiffLine[] = []
  let i = n
  let j = m
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      result.unshift({
        type: 'unchanged',
        text: oldLines[i - 1],
        oldNum: i,
        newNum: j,
      })
      i--
      j--
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      result.unshift({
        type: 'added',
        text: newLines[j - 1],
        newNum: j,
      })
      j--
    } else if (i > 0 && (j === 0 || dp[i][j - 1] < dp[i - 1][j])) {
      result.unshift({
        type: 'removed',
        text: oldLines[i - 1],
        oldNum: i,
      })
      i--
    }
  }

  return result
})

const diffSummary = computed(() => {
  let added = 0
  let removed = 0
  for (const line of diffLines.value) {
    if (line.type === 'added') added++
    if (line.type === 'removed') removed++
  }
  return { added, removed }
})

const lineCount = computed(() => {
  if (!yamlContent.value) return 1
  return yamlContent.value.split('\n').length
})

const lineNumbers = computed(() => {
  return Array.from({ length: lineCount.value }, (_, i) => i + 1)
})

const isModified = computed(() => {
  return yamlContent.value !== originalYaml.value
})

async function fetchManifest() {
  if (!props.kind || !props.name) return
  isLoading.value = true
  validationResult.value = null
  isEditing.value = false

  try {
    const res = await k8sApi.getResourceYAML(props.kind, props.namespace, props.name)
    yamlContent.value = res.yaml
    originalYaml.value = res.yaml
    apiVersion.value = res.api_version
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || 'Failed to fetch resource YAML'
    toast.add({
      severity: 'error',
      summary: 'Error Fetching YAML',
      detail: msg,
      life: 5000,
    })
  } finally {
    isLoading.value = false
  }
}

async function handleValidateDryRun() {
  if (!yamlContent.value.trim()) return
  isValidating.value = true
  validationResult.value = null

  try {
    const res = await k8sApi.applyYAML(yamlContent.value, props.namespace, true)
    if (res.error_count > 0) {
      const errItem = res.results.find((r) => r.status === 'error')
      validationResult.value = {
        status: 'error',
        message: errItem?.message || 'YAML validation failed during dry run',
      }
    } else {
      validationResult.value = {
        status: 'success',
        message: `Validation successful: ${res.success_count} resource(s) valid.`,
      }
    }
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || 'Dry-run validation failed'
    validationResult.value = {
      status: 'error',
      message: msg,
    }
  } finally {
    isValidating.value = false
  }
}

async function handleApply() {
  if (!yamlContent.value.trim()) return
  isApplying.value = true
  validationResult.value = null

  try {
    const res = await k8sApi.applyYAML(yamlContent.value, props.namespace, false)
    if (res.error_count > 0) {
      const errItem = res.results.find((r) => r.status === 'error')
      toast.add({
        severity: 'error',
        summary: 'Apply Failed',
        detail: errItem?.message || 'Failed to apply manifest changes',
        life: 5000,
      })
    } else {
      toast.add({
        severity: 'success',
        summary: 'Manifest Applied',
        detail: `${props.kind} ${props.name} updated successfully`,
        life: 4000,
      })
      originalYaml.value = yamlContent.value
      isEditing.value = false
      emit('applied')
    }
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || 'Failed to apply YAML'
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: msg,
      life: 5000,
    })
  } finally {
    isApplying.value = false
  }
}

function handleCopy() {
  navigator.clipboard.writeText(yamlContent.value)
  toast.add({
    severity: 'info',
    summary: 'Copied',
    detail: 'YAML copied to clipboard',
    life: 2000,
  })
}

function handleDownload() {
  const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.kind.toLowerCase()}-${props.name}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

function handleReset() {
  yamlContent.value = originalYaml.value
  validationResult.value = null
  isEditing.value = false
}

function handleClose() {
  emit('update:visible', false)
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      fetchManifest()
    }
  },
)
</script>

<template>
  <Dialog
    :visible="props.visible"
    modal
    :show-header="false"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden w-[90vw] max-w-4xl rounded-2xl shadow-2xl bg-slate-950 z-50' },
      content: { class: 'p-0 overflow-hidden bg-slate-950 flex flex-col max-h-[88vh]' },
    }"
    @update:visible="handleClose"
  >
    <!-- Custom Dialog Header -->
    <div
      class="flex items-center justify-between px-6 py-4 bg-slate-900 border-b border-slate-800 shrink-0"
    >
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
          <i class="pi pi-file-edit text-base"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="font-bold text-white text-base tracking-wide">{{ props.kind }} Manifest</span>
            <span class="px-2 py-0.5 rounded text-[11px] font-mono bg-slate-800 text-slate-300 border border-slate-700">
              {{ apiVersion || 'k8s' }}
            </span>
            <span
              v-if="isModified"
              class="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-500/20 text-amber-300 border border-amber-500/40"
            >
              MODIFIED
            </span>
          </div>
          <div class="text-xs text-slate-400 flex items-center gap-2 mt-0.5 font-mono">
            <span class="text-amber-300/90">{{ props.name }}</span>
            <span>&bull;</span>
            <span>namespace: {{ props.namespace }}</span>
          </div>
        </div>
      </div>

      <!-- Action buttons -->
      <div class="flex items-center gap-2">
        <!-- View Mode Switcher -->
        <div class="flex items-center bg-slate-800 p-0.5 rounded-lg border border-slate-700">
          <button
            type="button"
            class="px-2.5 py-1 rounded text-xs transition font-medium flex items-center gap-1.5 cursor-pointer"
            :class="activeTab === 'editor' ? 'bg-slate-700 text-white shadow-xs font-semibold' : 'text-slate-400 hover:text-white'"
            @click="activeTab = 'editor'"
          >
            <i class="pi pi-code text-[11px]"></i>
            <span>Editor</span>
          </button>
          <button
            type="button"
            class="px-2.5 py-1 rounded text-xs transition font-medium flex items-center gap-1.5 cursor-pointer"
            :class="activeTab === 'diff' ? 'bg-slate-700 text-white shadow-xs font-semibold' : 'text-slate-400 hover:text-white'"
            @click="activeTab = 'diff'"
          >
            <i class="pi pi-arrows-h text-[11px]"></i>
            <span>Diff</span>
            <span
              v-if="isModified"
              class="px-1.5 py-0.2 rounded text-[10px] font-mono font-bold bg-amber-500/20 text-amber-300"
            >
              +{{ diffSummary.added }} -{{ diffSummary.removed }}
            </span>
          </button>
        </div>

        <!-- Edit Mode Toggle -->
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium transition flex items-center gap-1.5 border cursor-pointer"
          :class="isEditing
            ? 'bg-amber-500 text-slate-950 border-amber-400 font-semibold'
            : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-700 hover:text-white'"
          @click="isEditing = !isEditing"
        >
          <i :class="isEditing ? 'pi pi-lock' : 'pi pi-pencil'" class="text-xs"></i>
          <span>{{ isEditing ? 'Editing Enabled' : 'Edit Mode' }}</span>
        </button>

        <!-- Copy -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition cursor-pointer"
          title="Copy to clipboard"
          @click="handleCopy"
        >
          <i class="pi pi-copy text-xs"></i>
        </button>

        <!-- Download -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition cursor-pointer"
          title="Download .yaml"
          @click="handleDownload"
        >
          <i class="pi pi-download text-xs"></i>
        </button>

        <!-- Reset -->
        <button
          v-if="isModified"
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-amber-900/40 text-amber-400 border border-slate-700 hover:border-amber-700/50 transition cursor-pointer"
          title="Reset changes"
          @click="handleReset"
        >
          <i class="pi pi-undo text-xs"></i>
        </button>

        <!-- Close -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-rose-900/40 text-slate-400 hover:text-rose-400 border border-slate-700 hover:border-rose-700/50 transition ml-1 cursor-pointer"
          title="Close"
          @click="handleClose"
        >
          <i class="pi pi-times text-xs"></i>
        </button>
      </div>
    </div>

    <!-- Validation Result Banner -->
    <div
      v-if="validationResult"
      class="px-6 py-2.5 text-xs font-mono flex items-center justify-between border-b"
      :class="validationResult.status === 'success'
        ? 'bg-emerald-950/60 border-emerald-800 text-emerald-300'
        : 'bg-rose-950/60 border-rose-800 text-rose-300'"
    >
      <div class="flex items-center gap-2">
        <i :class="validationResult.status === 'success' ? 'pi pi-check-circle' : 'pi pi-exclamation-circle'"></i>
        <span>{{ validationResult.message }}</span>
      </div>
      <button class="text-xs opacity-70 hover:opacity-100 cursor-pointer" @click="validationResult = null">
        <i class="pi pi-times"></i>
      </button>
    </div>

    <!-- VIEW 1: Editor Area with Line Numbers -->
    <div v-if="activeTab === 'editor'" class="relative flex flex-1 overflow-hidden bg-[#0c1220] min-h-[380px] max-h-[560px]">
      <!-- Loading State -->
      <div v-if="isLoading" class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center z-10">
        <div class="flex flex-col items-center gap-3">
          <i class="pi pi-spin pi-spinner text-2xl text-amber-400"></i>
          <span class="text-xs text-slate-400 font-mono">Fetching {{ props.kind }} manifest...</span>
        </div>
      </div>

      <!-- Line numbers gutter -->
      <div
        class="w-12 bg-slate-900/60 border-r border-slate-800/80 select-none py-3 px-2 text-right font-mono text-xs text-slate-600 overflow-hidden leading-relaxed shrink-0"
      >
        <div v-for="n in lineNumbers" :key="n">{{ n }}</div>
      </div>

      <!-- Textarea / Editor -->
      <div class="flex-1 overflow-auto relative">
        <textarea
          v-model="yamlContent"
          :readonly="!isEditing"
          spellcheck="false"
          class="w-full h-full min-h-[380px] bg-transparent text-slate-200 font-mono text-xs p-3 leading-relaxed border-none outline-none resize-none selection:bg-amber-500/30 whitespace-pre"
          :class="{ 'opacity-90 cursor-default': !isEditing, 'cursor-text': isEditing }"
          placeholder="Loading YAML manifest..."
        ></textarea>
      </div>
    </div>

    <!-- VIEW 2: Visual Side-by-Side / Unified Diff Viewer -->
    <div v-else class="flex-1 overflow-auto bg-[#0c1220] min-h-[380px] max-h-[560px] font-mono text-xs p-2">
      <div v-if="!isModified" class="py-16 text-center text-slate-500">
        <i class="pi pi-check-circle text-3xl mb-2 text-emerald-500"></i>
        <h4 class="font-bold text-slate-300">No Changes Detected</h4>
        <p class="text-xs text-slate-500 mt-1">The working copy exactly matches the active cluster manifest.</p>
      </div>

      <div v-else class="space-y-0.5">
        <div
          v-for="(diff, idx) in diffLines"
          :key="idx"
          class="flex items-stretch font-mono text-xs leading-relaxed px-2 py-0.5 rounded-sm"
          :class="{
            'bg-emerald-950/40 text-emerald-300 border-l-2 border-emerald-500': diff.type === 'added',
            'bg-rose-950/40 text-rose-300 border-l-2 border-rose-500 opacity-80': diff.type === 'removed',
            'text-slate-400 hover:bg-slate-900/40': diff.type === 'unchanged',
          }"
        >
          <!-- Line markers -->
          <div class="w-16 flex items-center justify-between text-[10px] text-slate-600 select-none pr-3 shrink-0">
            <span class="w-7 text-right">{{ diff.oldNum || '' }}</span>
            <span class="w-7 text-right">{{ diff.newNum || '' }}</span>
          </div>

          <!-- Diff Sign -->
          <span class="w-5 text-center shrink-0 font-bold" :class="{
            'text-emerald-400': diff.type === 'added',
            'text-rose-400': diff.type === 'removed',
            'text-slate-700': diff.type === 'unchanged',
          }">
            {{ diff.type === 'added' ? '+' : diff.type === 'removed' ? '-' : ' ' }}
          </span>

          <!-- Line text -->
          <span class="whitespace-pre overflow-x-auto flex-1">{{ diff.text }}</span>
        </div>
      </div>
    </div>

    <!-- Dialog Footer Actions -->
    <div
      class="flex items-center justify-between px-6 py-3.5 bg-slate-900 border-t border-slate-800 shrink-0"
    >
      <div class="text-xs text-slate-500 font-mono flex items-center gap-2">
        <span>Lines: {{ lineCount }}</span>
        <span>&bull;</span>
        <span>Mode: {{ isEditing ? 'Editable' : 'Read-Only' }}</span>
      </div>

      <div class="flex items-center gap-3">
        <button
          type="button"
          class="px-4 py-2 rounded-xl text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-300 transition border border-slate-700"
          @click="handleClose"
        >
          Close
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-xl text-xs font-medium bg-sky-600/20 hover:bg-sky-600/30 text-sky-400 border border-sky-500/40 transition flex items-center gap-1.5"
          :disabled="isValidating || !isModified"
          @click="handleValidateDryRun"
        >
          <i v-if="isValidating" class="pi pi-spin pi-spinner text-xs"></i>
          <i v-else class="pi pi-shield text-xs"></i>
          <span>Validate (Dry Run)</span>
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-xl text-xs font-bold bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 text-slate-950 transition shadow-lg shadow-emerald-500/20 flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="isApplying || !isModified"
          @click="handleApply"
        >
          <i v-if="isApplying" class="pi pi-spin pi-spinner text-xs"></i>
          <i v-else class="pi pi-check text-xs"></i>
          <span>Apply Changes</span>
        </button>
      </div>
    </div>
  </Dialog>
</template>
