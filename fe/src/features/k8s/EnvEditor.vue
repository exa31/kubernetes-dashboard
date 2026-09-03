<script setup lang="ts">
import Button from 'primevue/button'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { ConfigMapDetail, SecretDetail } from '@/types'
import { logger } from '@/utils'

const props = defineProps<{
  resourceType: 'secret' | 'configmap'
  detail: SecretDetail | ConfigMapDetail
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved', payload: Record<string, string>): void
}>()

const k8sStore = useK8sStore()

// State
type EditorMode = 'table' | 'dotenv' | 'yaml'
const activeMode = ref<EditorMode>('table')
const searchQuery = ref('')
const showAllSecrets = ref(false)
const revealedKeys = ref<Record<string, boolean>>({})
const copySuccessKey = ref<string | null>(null)
const bannerMessage = ref<{ text: string; type: 'success' | 'error' | 'info' } | null>(null)
const selectedDeploymentToRestart = ref<string>('')

// Key-Value Rows
interface EnvRow {
  id: string
  key: string
  value: string
  isNew?: boolean
}

let rowCounter = 0
const rows = ref<EnvRow[]>([])
const rawDotEnv = ref('')

// Initialize rows from props.detail.data
// Synchronize rows -> rawDotEnv
function syncToDotEnv() {
  const lines = rows.value
    .filter((r) => r.key.trim() !== '')
    .map((r) => {
      let val = r.value
      // Wrap in quotes if contains spaces, special chars, or newlines
      if (val.includes('\n') || val.includes(' ') || val.includes('"') || val.includes('#')) {
        val = `"${val.replace(/"/g, '\\"')}"`
      }
      return `${r.key}=${val}`
    })
  rawDotEnv.value = lines.join('\n')
}

// Synchronize rawDotEnv -> rows (when user edits .env tab)
function syncFromDotEnv() {
  const lines = rawDotEnv.value.split('\n')
  const newRows: EnvRow[] = []

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    const equalIndex = trimmed.indexOf('=')
    if (equalIndex > 0) {
      const k = trimmed.slice(0, equalIndex).trim()
      let v = trimmed.slice(equalIndex + 1).trim()

      // Strip matching wrapping quotes
      if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
        v = v.slice(1, -1)
      }
      newRows.push({
        id: `row-${++rowCounter}`,
        key: k,
        value: v,
      })
    }
  }

  rows.value = newRows
}

function initFromData() {
  const data = props.detail?.data || {}
  const parsedRows: EnvRow[] = []
  for (const [key, value] of Object.entries(data)) {
    parsedRows.push({
      id: `row-${++rowCounter}`,
      key,
      value: String(value ?? ''),
    })
  }
  // Sort alphabetically by key
  parsedRows.sort((a, b) => a.key.localeCompare(b.key))
  rows.value = parsedRows
  syncToDotEnv()
}

watch(
  () => props.detail,
  () => {
    initFromData()
  },
  { immediate: true },
)

const onDotEnvChange = () => {
  syncFromDotEnv()
}

const onTabChange = (mode: EditorMode) => {
  if (activeMode.value === 'dotenv') {
    syncFromDotEnv()
  } else if (mode === 'dotenv') {
    syncToDotEnv()
  }
  activeMode.value = mode
}

// Filtered rows
const filteredRows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter((r) => r.key.toLowerCase().includes(q) || r.value.toLowerCase().includes(q))
})

// Add new row
const addRow = () => {
  const newRow: EnvRow = {
    id: `row-${++rowCounter}`,
    key: '',
    value: '',
    isNew: true,
  }
  rows.value.unshift(newRow)
  syncToDotEnv()
}

// Remove row
const removeRow = (id: string) => {
  rows.value = rows.value.filter((r) => r.id !== id)
  syncToDotEnv()
}

// Toggle mask
const toggleReveal = (key: string) => {
  revealedKeys.value[key] = !revealedKeys.value[key]
}

const toggleAllReveal = () => {
  showAllSecrets.value = !showAllSecrets.value
  const newState = showAllSecrets.value
  const keysMap: Record<string, boolean> = {}
  rows.value.forEach((r) => {
    keysMap[r.key] = newState
  })
  revealedKeys.value = keysMap
}

const isRevealed = (key: string) => {
  if (props.resourceType === 'configmap') return true
  if (showAllSecrets.value) return true
  return !!revealedKeys.value[key]
}

// Copy to clipboard
const copyToClipboard = async (text: string, keyIdentifier: string) => {
  try {
    await navigator.clipboard.writeText(text)
    copySuccessKey.value = keyIdentifier
    setTimeout(() => {
      if (copySuccessKey.value === keyIdentifier) {
        copySuccessKey.value = null
      }
    }, 2000)
  } catch (err) {
    logger.warn('Failed to copy to clipboard', err)
  }
}

// Export as .env file
const exportAsDotEnv = () => {
  syncToDotEnv()
  const blob = new Blob([rawDotEnv.value], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.detail.name}.env`
  a.click()
  URL.revokeObjectURL(url)
}

// Import from .env file
const fileInput = ref<HTMLInputElement | null>(null)
const triggerFileInput = () => {
  fileInput.value?.click()
}
const onFileSelected = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (event) => {
    rawDotEnv.value = String(event.target?.result || '')
    syncFromDotEnv()
    bannerMessage.value = {
      text: `Imported variables from ${file.name}`,
      type: 'success',
    }
  }
  reader.readAsText(file)
}

// Connected deployments
const connectedDeployments = computed(() => {
  const {name} = props.detail
  return k8sStore.deployments.filter((d) => {
    if (props.resourceType === 'secret') {
      return d.env_secrets.includes(name)
    }
    return d.env_configmaps.includes(name)
  })
})

// Auto-select first matching deployment if available
watch(
  connectedDeployments,
  (deps) => {
    if (deps.length > 0 && !selectedDeploymentToRestart.value) {
      selectedDeploymentToRestart.value = deps[0].name
    }
  },
  { immediate: true },
)

// YAML Manifest Preview
const yamlManifest = computed(() => {
  const kind = props.resourceType === 'secret' ? 'Secret' : 'ConfigMap'
  const {name} = props.detail
  const {namespace} = props.detail
  const dataMap = rows.value.reduce(
    (acc, r) => {
      if (r.key.trim()) acc[r.key.trim()] = r.value
      return acc
    },
    {} as Record<string, string>,
  )

  let out = `apiVersion: v1\nkind: ${kind}\nmetadata:\n  name: ${name}\n  namespace: ${namespace}\n`
  if (props.resourceType === 'secret') {
    out += `type: ${props.detail.type || 'Opaque'}\nstringData:\n`
  } else {
    out += `data:\n`
  }
  for (const [k, v] of Object.entries(dataMap)) {
    const formatted = v.includes('\n') ? `|\n    ${v.replace(/\n/g, '\n    ')}` : `"${v.replace(/"/g, '\\"')}"`
    out += `  ${k}: ${formatted}\n`
  }
  return out
})

// Save to Kubernetes
const isSaving = ref(false)
const saveChanges = async (restartDeploymentName?: string) => {
  isSaving.value = true
  bannerMessage.value = null
  try {
    // If in dotenv mode, sync first
    if (activeMode.value === 'dotenv') {
      syncFromDotEnv()
    }

    const payloadData: Record<string, string> = {}
    rows.value.forEach((r) => {
      const k = r.key.trim()
      if (k) {
        payloadData[k] = r.value
      }
    })

    if (props.resourceType === 'secret') {
      await k8sStore.saveSecret({
        name: props.detail.name,
        namespace: props.detail.namespace,
        type: props.detail.type || 'Opaque',
        data: payloadData,
        labels: props.detail.labels,
        annotations: props.detail.annotations,
      })
    } else {
      await k8sStore.saveConfigMap({
        name: props.detail.name,
        namespace: props.detail.namespace,
        data: payloadData,
        labels: props.detail.labels,
        annotations: props.detail.annotations,
      })
    }

    let msg = `Successfully saved ${props.resourceType} '${props.detail.name}'`

    // If restart requested
    if (restartDeploymentName) {
      await k8sStore.restartDeployment(restartDeploymentName, props.detail.namespace)
      msg += ` & initiated rollout restart for deployment '${restartDeploymentName}'!`
    }

    bannerMessage.value = { text: msg, type: 'success' }
    emit('saved', payloadData)
  } catch (err: unknown) {
    logger.error('Failed to save resource', err)
    let errText = 'Failed to save to Kubernetes cluster'
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      if (res?.data?.message) errText = res.data.message
    } else if (err instanceof Error) {
      errText = err.message
    }
    bannerMessage.value = {
      text: errText,
      type: 'error',
    }
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <div class="flex flex-col h-full bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl overflow-hidden">
    <!-- Hidden file input for .env import -->
    <input
      ref="fileInput"
      type="file"
      accept=".env,text/plain"
      class="hidden"
      @change="onFileSelected"
    />

    <!-- Header Section -->
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-4 bg-slate-50/70 dark:bg-slate-950/40">
      <div class="flex items-center gap-3 min-w-0">
        <div
class="w-10 h-10 rounded-lg flex items-center justify-center font-bold text-white shadow-sm shrink-0"
             :class="resourceType === 'secret' ? 'bg-gradient-to-tr from-amber-500 to-orange-500' : 'bg-gradient-to-tr from-sky-500 to-blue-600'">
          <i :class="resourceType === 'secret' ? 'pi pi-lock text-lg' : 'pi pi-file text-lg'"></i>
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="text-lg font-bold text-slate-900 dark:text-slate-100 truncate">
              {{ detail.name }}
            </h2>
            <span
class="text-xs px-2 py-0.5 rounded-full font-medium"
                  :class="resourceType === 'secret' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300' : 'bg-sky-100 text-sky-800 dark:bg-sky-950/60 dark:text-sky-300'">
              {{ resourceType === 'secret' ? (detail.type || 'Secret') : 'ConfigMap' }}
            </span>
          </div>
          <div class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 flex items-center gap-2">
            <span>Namespace: <strong class="text-slate-700 dark:text-slate-300">{{ detail.namespace }}</strong></span>
            <span>•</span>
            <span>{{ rows.length }} variables</span>
          </div>
        </div>
      </div>

      <!-- Action Toolbar -->
      <div class="flex items-center gap-2">
        <Button
          label="Export .env"
          icon="pi pi-download"
          size="small"
          severity="secondary"
          outlined
          @click="exportAsDotEnv"
        />
        <Button
          label="Import .env"
          icon="pi pi-upload"
          size="small"
          severity="secondary"
          outlined
          @click="triggerFileInput"
        />
        <Button
          label="Copy All"
          :icon="copySuccessKey === 'all' ? 'pi pi-check' : 'pi pi-copy'"
          size="small"
          severity="secondary"
          outlined
          @click="copyToClipboard(rawDotEnv, 'all')"
        />
        
        <div class="h-6 w-px bg-slate-200 dark:bg-slate-700 mx-1 hidden sm:block"></div>

        <Button
          label="Save"
          icon="pi pi-check"
          size="small"
          :loading="isSaving"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          @click="() => saveChanges()"
        />

        <!-- Save & Rollout Restart if connected deployment detected -->
        <Button
          v-if="connectedDeployments.length > 0"
          :label="`Save & Restart Pods (${connectedDeployments[0].name})`"
          icon="pi pi-refresh"
          size="small"
          :loading="isSaving"
          class="btn-amber text-xs shadow-xs cursor-pointer"
          @click="() => saveChanges(connectedDeployments[0].name)"
        />

        <Button
          icon="pi pi-times"
          size="small"
          severity="secondary"
          text
          rounded
          @click="emit('close')"
        />
      </div>
    </div>

    <!-- Notification Banner -->
    <div
      v-if="bannerMessage"
      class="px-6 py-2.5 text-sm flex items-center justify-between border-b transition-all"
      :class="{
        'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-200 border-emerald-200 dark:border-emerald-800/60': bannerMessage.type === 'success',
        'bg-rose-50 dark:bg-rose-950/40 text-rose-800 dark:text-rose-200 border-rose-200 dark:border-rose-800/60': bannerMessage.type === 'error',
        'bg-blue-50 dark:bg-blue-950/40 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-800/60': bannerMessage.type === 'info',
      }"
    >
      <div class="flex items-center gap-2">
        <i :class="bannerMessage.type === 'success' ? 'pi pi-check-circle' : 'pi pi-exclamation-circle'"></i>
        <span>{{ bannerMessage.text }}</span>
      </div>
      <button class="text-xs hover:opacity-75" @click="bannerMessage = null">Dismiss</button>
    </div>

    <!-- Mode Selector & Filter Toolbar -->
    <div class="px-6 py-3 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 bg-white dark:bg-slate-900">
      <!-- Tabs -->
      <div class="flex items-center p-1 bg-slate-100 dark:bg-slate-800 rounded-lg text-xs font-semibold">
        <button
          type="button"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-md transition-all cursor-pointer"
          :class="activeMode === 'table' ? 'bg-white dark:bg-slate-700 text-sky-600 dark:text-sky-400 font-bold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
          @click="onTabChange('table')"
        >
          <i class="pi pi-table"></i>
          <span>Key-Value Table (Rancher)</span>
        </button>
        <button
          type="button"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-md transition-all cursor-pointer"
          :class="activeMode === 'dotenv' ? 'bg-white dark:bg-slate-700 text-sky-600 dark:text-sky-400 font-bold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
          @click="onTabChange('dotenv')"
        >
          <i class="pi pi-code"></i>
          <span>Raw .env Bulk Editor</span>
        </button>
        <button
          type="button"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-md transition-all cursor-pointer"
          :class="activeMode === 'yaml' ? 'bg-white dark:bg-slate-700 text-sky-600 dark:text-sky-400 font-bold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
          @click="onTabChange('yaml')"
        >
          <i class="pi pi-align-left"></i>
          <span>YAML Manifest</span>
        </button>
      </div>

      <!-- Controls in Table Mode -->
      <div v-if="activeMode === 'table'" class="flex items-center gap-3">
        <!-- Search filter -->
        <IconField>
          <InputIcon class="pi pi-search text-xs" />
          <InputText
            v-model="searchQuery"
            placeholder="Filter keys or values..."
            class="py-1 text-xs w-56 rounded-lg"
          />
        </IconField>

        <!-- Toggle Mask All (for secrets) -->
        <Button
          v-if="resourceType === 'secret'"
          :label="showAllSecrets ? 'Mask All Values' : 'Reveal All Values'"
          :icon="showAllSecrets ? 'pi pi-eye-slash' : 'pi pi-eye'"
          size="small"
          severity="secondary"
          text
          @click="toggleAllReveal"
        />

        <Button
          label="Add Variable"
          icon="pi pi-plus"
          size="small"
          class="btn-sky text-xs shadow-xs cursor-pointer"
          @click="addRow"
        />
      </div>
    </div>

    <!-- Tab 1: Key-Value Table Editor (Rancher Style) -->
    <div v-if="activeMode === 'table'" class="flex-1 overflow-auto p-6">
      <div class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-800">
        <table class="w-full text-left text-sm border-collapse">
          <thead>
            <tr class="bg-slate-50 dark:bg-slate-950/60 border-b border-slate-200 dark:border-slate-800 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              <th class="py-3 px-4 w-12 text-center">#</th>
              <th class="py-3 px-4 w-2/5">Key / Variable Name</th>
              <th class="py-3 px-4">Value (Plaintext)</th>
              <th class="py-3 px-4 w-28 text-center">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800 bg-white dark:bg-slate-900">
            <tr
              v-for="(row, idx) in filteredRows"
              :key="idx"
              class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors group"
            >
              <!-- Row Index -->
              <td class="py-2.5 px-4 text-xs font-mono text-slate-400 text-center select-none">
                {{ idx + 1 }}
              </td>

              <!-- Key -->
              <td class="py-2.5 px-4">
                <input
                  v-model="row.key"
                  type="text"
                  placeholder="e.g. DATABASE_PASSWORD"
                  class="w-full font-mono text-xs font-semibold text-slate-800 dark:text-slate-100 bg-transparent border border-transparent hover:border-slate-300 dark:hover:border-slate-700 focus:border-sky-500 focus:bg-white dark:focus:bg-slate-900 rounded px-2 py-1.5 focus:outline-none transition-all"
                  @input="syncToDotEnv"
                />
              </td>

              <!-- Value -->
              <td class="py-2.5 px-4">
                <div class="relative flex items-center">
                  <input
                    v-model="row.value"
                    :type="isRevealed(row.key) ? 'text' : 'password'"
                    placeholder="Value..."
                    class="w-full font-mono text-xs text-slate-800 dark:text-slate-100 bg-transparent border border-transparent hover:border-slate-300 dark:hover:border-slate-700 focus:border-sky-500 focus:bg-white dark:focus:bg-slate-900 rounded pl-2 pr-16 py-1.5 focus:outline-none transition-all"
                    @input="syncToDotEnv"
                  />
                  <!-- Inline tools: Eye toggle & Copy -->
                  <div class="absolute right-1.5 flex items-center gap-1">
                    <button
                      v-if="resourceType === 'secret'"
                      type="button"
                      class="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                      :title="isRevealed(row.key) ? 'Mask value' : 'Show value'"
                      @click="toggleReveal(row.key)"
                    >
                      <i :class="isRevealed(row.key) ? 'pi pi-eye-slash text-xs' : 'pi pi-eye text-xs'"></i>
                    </button>
                    <button
                      type="button"
                      class="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                      :title="copySuccessKey === row.id ? 'Copied!' : 'Copy value'"
                      @click="copyToClipboard(row.value, row.id)"
                    >
                      <i :class="copySuccessKey === row.id ? 'pi pi-check text-xs text-emerald-500' : 'pi pi-copy text-xs'"></i>
                    </button>
                  </div>
                </div>
              </td>

              <!-- Actions -->
              <td class="py-2.5 px-4 text-center">
                <button
                  type="button"
                  class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer"
                  title="Remove variable"
                  @click="removeRow(row.id)"
                >
                  <i class="pi pi-trash text-xs"></i>
                </button>
              </td>
            </tr>

            <!-- Empty state -->
            <tr v-if="filteredRows.length === 0">
              <td colspan="4" class="py-12 text-center text-slate-400">
                <i class="pi pi-inbox text-3xl mb-2"></i>
                <p class="text-sm">No environment variables match your search.</p>
                <Button
                  label="Add Variable"
                  icon="pi pi-plus"
                  size="small"
                  class="mt-3"
                  @click="addRow"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Add button at bottom -->
      <div class="mt-4 flex justify-between items-center text-xs text-slate-500">
        <span>Showing {{ filteredRows.length }} of {{ rows.length }} variables</span>
        <button
          type="button"
          class="flex items-center gap-1.5 font-semibold text-sky-600 dark:text-sky-400 hover:underline cursor-pointer"
          @click="addRow"
        >
          <i class="pi pi-plus text-[10px]"></i>
          <span>Add New Variable</span>
        </button>
      </div>
    </div>

    <!-- Tab 2: Raw .env Bulk Editor -->
    <div v-else-if="activeMode === 'dotenv'" class="flex-1 flex flex-col p-6 overflow-hidden">
      <div class="mb-3 flex items-center justify-between text-xs text-slate-500">
        <span>
          Bulk edit environment variables directly in standard <code class="font-bold text-sky-500">.env</code> syntax. Edits automatically synchronize with the table view.
        </span>
        <span class="font-mono">{{ rawDotEnv.split('\n').filter(Boolean).length }} lines</span>
      </div>
      <div class="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 overflow-hidden relative shadow-inner">
        <textarea
          v-model="rawDotEnv"
          spellcheck="false"
          class="w-full h-full p-4 font-mono text-xs bg-slate-950 text-emerald-400 focus:outline-none resize-none leading-relaxed selection:bg-sky-600 selection:text-white"
          placeholder="KEY=VALUE&#10;ANOTHER_KEY=another_value"
          @input="onDotEnvChange"
        ></textarea>
      </div>
    </div>

    <!-- Tab 3: YAML Manifest -->
    <div v-else class="flex-1 flex flex-col p-6 overflow-hidden">
      <div class="mb-3 flex items-center justify-between text-xs text-slate-500">
        <span>
          Live Kubernetes manifest preview with <code class="font-bold text-sky-500">stringData</code>.
        </span>
        <Button
          label="Copy YAML"
          icon="pi pi-copy"
          size="small"
          class="btn-sky text-xs px-2.5 py-1.5 rounded-lg shadow-xs cursor-pointer"
          @click="copyToClipboard(yamlManifest, 'yaml')"
        />
      </div>
      <div class="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 overflow-auto bg-slate-950 p-4">
        <pre class="font-mono text-xs text-sky-300 leading-relaxed">{{ yamlManifest }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
