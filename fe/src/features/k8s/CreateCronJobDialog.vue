<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import { CRON_PRESETS, describeCron } from '@/utils'

const props = defineProps<{
  visible: boolean
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'created'): void
}>()

const k8sStore = useK8sStore()

const name = ref('')
const schedule = ref('0 0 * * *')
const image = ref('')
const isSuspended = ref(false)
const isCreating = ref(false)
const errorMessage = ref<string | null>(null)

watch(
  () => props.visible,
  (vis) => {
    if (vis) {
      name.value = ''
      schedule.value = '0 0 * * *'
      image.value = 'busybox:latest'
      isSuspended.value = false
      errorMessage.value = null
    }
  },
)

const scheduleDescription = computed(() => describeCron(schedule.value))

function setPreset(val: string) {
  schedule.value = val
}

async function handleCreate() {
  if (!name.value.trim()) {
    errorMessage.value = 'CronJob name is required'
    return
  }
  if (!schedule.value.trim()) {
    errorMessage.value = 'Schedule expression is required'
    return
  }
  if (!image.value.trim()) {
    errorMessage.value = 'Container image is required'
    return
  }

  isCreating.value = true
  errorMessage.value = null
  try {
    await k8sStore.createCronJob({
      name: name.value.trim().toLowerCase(),
      namespace: props.namespace,
      schedule: schedule.value.trim(),
      suspend: isSuspended.value,
      containers: [
        {
          name: name.value.trim().toLowerCase(),
          image: image.value.trim(),
          env: [],
          env_from: [],
        },
      ],
    })
    emit('created')
    emit('update:visible', false)
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to create CronJob'
  } finally {
    isCreating.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[90vw] max-w-xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
    }"
    @update:visible="(val) => emit('update:visible', val)"
  >
    <!-- Custom Header -->
    <div class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-amber-500/10 text-amber-500 flex items-center justify-center font-bold text-base shrink-0">
          <i class="pi pi-clock"></i>
        </div>
        <div>
          <h2 class="font-bold text-base text-slate-900 dark:text-slate-100">
            Create New CronJob
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Schedule recurring automated batch tasks in namespace <span class="font-mono text-slate-700 dark:text-slate-300 font-semibold">{{ namespace }}</span>
          </p>
        </div>
      </div>
      <button
        type="button"
        class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="emit('update:visible', false)"
      >
        <i class="pi pi-times"></i>
      </button>
    </div>

    <div class="p-6 space-y-4">
      <div
        v-if="errorMessage"
        class="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-300 text-xs flex items-center gap-2"
      >
        <i class="pi pi-exclamation-circle text-base shrink-0"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <div>
        <label class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 block mb-1">
          CronJob Name
        </label>
        <InputText
          v-model="name"
          placeholder="e.g. daily-cleanup-job"
          class="font-mono text-sm w-full"
        />
      </div>

      <!-- Schedule -->
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <label class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 block">
            Schedule
          </label>
          <span class="text-xs font-semibold px-2 py-0.5 rounded bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400 font-mono">
            {{ scheduleDescription }}
          </span>
        </div>

        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="preset in CRON_PRESETS"
            :key="preset.value"
            type="button"
            :class="[
              'px-2 py-0.5 text-xs rounded-md border font-mono cursor-pointer',
              schedule === preset.value
                ? 'bg-sky-600 text-white border-sky-600 shadow-xs'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700'
            ]"
            @click="setPreset(preset.value)"
          >
            {{ preset.label }}
          </button>
        </div>

        <InputText
          v-model="schedule"
          placeholder="0 0 * * *"
          class="font-mono text-sm w-full"
        />
      </div>

      <!-- Container Image -->
      <div>
        <label class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 block mb-1">
          Container Image
        </label>
        <InputText
          v-model="image"
          placeholder="e.g. alpine:latest or busybox:latest"
          class="font-mono text-sm w-full"
        />
      </div>

      <!-- Initial Suspend -->
      <div class="flex items-center gap-2 pt-1">
        <input
          id="suspend-check"
          v-model="isSuspended"
          type="checkbox"
          class="rounded text-sky-600 focus:ring-sky-500 h-4 w-4"
        />
        <label for="suspend-check" class="text-xs text-slate-700 dark:text-slate-300 select-none cursor-pointer">
          Create as Suspended (paused initially)
        </label>
      </div>
      <div class="flex items-center justify-end gap-2 pt-4 border-t border-slate-200 dark:border-slate-800">
        <Button
          label="Cancel"
          severity="secondary"
          text
          size="small"
          @click="emit('update:visible', false)"
        />
        <Button
          label="Create CronJob"
          icon="pi pi-plus"
          size="small"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          :loading="isCreating"
          @click="handleCreate"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
</style>
