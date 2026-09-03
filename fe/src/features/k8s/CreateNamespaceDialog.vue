<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { computed, ref } from 'vue'

import { useK8sStore } from '@/stores'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'created', name: string): void
}>()

const k8sStore = useK8sStore()
const name = ref('')
const switchImmediately = ref(true)
const labelKey = ref('')
const labelValue = ref('')
const labels = ref<Record<string, string>>({})
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

const isValidName = computed(() => {
  const trimmed = name.value.trim()
  if (!trimmed) return false
  const regex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
  return regex.test(trimmed) && trimmed.length <= 63
})

function addLabel() {
  const k = labelKey.value.trim()
  const v = labelValue.value.trim()
  if (!k) return
  labels.value[k] = v
  labelKey.value = ''
  labelValue.value = ''
}

function removeLabel(k: string) {
  delete labels.value[k]
}

async function handleCreate() {
  if (!isValidName.value) return
  isSubmitting.value = true
  errorMessage.value = null
  try {
    const nsName = name.value.trim().toLowerCase()
    await k8sStore.createNamespace({
      name: nsName,
      labels: Object.keys(labels.value).length > 0 ? labels.value : undefined,
    })
    if (switchImmediately.value) {
      k8sStore.setNamespace(nsName)
    }
    emit('created', nsName)
    handleClose()
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to create namespace'
  } finally {
    isSubmitting.value = false
  }
}

function handleClose() {
  name.value = ''
  labels.value = {}
  labelKey.value = ''
  labelValue.value = ''
  errorMessage.value = null
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[90vw] max-w-md rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
    }"
    @update:visible="emit('update:visible', $event)"
  >
    <!-- Custom Header -->
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-900/90">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-sky-500/15 border border-sky-500/30 flex items-center justify-center text-sky-400 shrink-0">
          <i class="pi pi-folder-plus text-base"></i>
        </div>
        <div>
          <h3 class="text-sm font-bold text-slate-900 dark:text-slate-100">Create Namespace</h3>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Provision a new isolated cluster namespace
          </p>
        </div>
      </div>

      <button
        type="button"
        class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
        @click="handleClose"
      >
        <i class="pi pi-times text-xs"></i>
      </button>
    </div>

    <!-- Content -->
    <form class="p-6 space-y-4" @submit.prevent="handleCreate">
      <div v-if="errorMessage" class="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
        <i class="pi pi-exclamation-circle text-sm"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
          Namespace Name <span class="text-rose-500">*</span>
        </label>
        <InputText
          v-model="name"
          placeholder="e.g. production, staging-app"
          class="w-full text-xs rounded-xl bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 py-2 font-mono"
          autofocus
        />
        <p class="text-[11px] text-slate-400 mt-1 font-mono">
          Must be lowercase alphanumeric or hyphens (DNS-1123 label).
        </p>
      </div>

      <!-- Optional Labels -->
      <div class="space-y-2">
        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
          Labels (Optional)
        </label>
        <div class="flex items-center gap-2">
          <InputText
            v-model="labelKey"
            placeholder="Key (e.g. env)"
            class="flex-1 text-xs rounded-xl bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 py-1.5 font-mono"
          />
          <InputText
            v-model="labelValue"
            placeholder="Value (e.g. prod)"
            class="flex-1 text-xs rounded-xl bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 py-1.5 font-mono"
          />
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-800 text-slate-200 border border-slate-700 hover:bg-slate-700 transition cursor-pointer"
            @click="addLabel"
          >
            Add
          </button>
        </div>

        <!-- Render active labels -->
        <div v-if="Object.keys(labels).length > 0" class="flex flex-wrap gap-1.5 mt-2">
          <span
            v-for="(val, key) in labels"
            :key="key"
            class="px-2 py-0.5 rounded text-[11px] font-mono bg-sky-500/10 text-sky-400 border border-sky-500/30 flex items-center gap-1.5"
          >
            <span>{{ key }}: {{ val }}</span>
            <button type="button" class="hover:text-rose-400 cursor-pointer" @click="removeLabel(String(key))">
              &times;
            </button>
          </span>
        </div>
      </div>

      <!-- Quick Switch Checkbox -->
      <div class="flex items-center gap-2 pt-1">
        <input
          id="switch-ns"
          v-model="switchImmediately"
          type="checkbox"
          class="rounded border-slate-700 text-sky-500 focus:ring-0 cursor-pointer"
        />
        <label for="switch-ns" class="text-xs text-slate-600 dark:text-slate-300 cursor-pointer select-none">
          Switch to this namespace immediately upon creation
        </label>
      </div>

      <!-- Footer Actions -->
      <div class="pt-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-end gap-2.5">
        <Button
          label="Cancel"
          size="small"
          class="btn-slate text-xs px-3.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
          @click="handleClose"
        />
        <Button
          label="Create Namespace"
          icon="pi pi-check"
          size="small"
          class="btn-sky text-xs px-3.5 py-1.5 rounded-lg active:scale-95 cursor-pointer font-semibold"
          :disabled="!isValidName || isSubmitting"
          :loading="isSubmitting"
          @click="handleCreate"
        />
      </div>
    </form>
  </Dialog>
</template>
