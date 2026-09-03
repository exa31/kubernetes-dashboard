<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { ContainerDetail, CronJobDetail } from '@/types'
import { CRON_PRESETS, describeCron } from '@/utils'

const props = defineProps<{
  visible: boolean
  cronJob: CronJobDetail | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'saved'): void
}>()

const k8sStore = useK8sStore()

const schedule = ref('')
const isSuspended = ref(false)
const containers = ref<ContainerDetail[]>([])
const isSaving = ref(false)
const errorMessage = ref<string | null>(null)

watch(
  () => props.cronJob,
  (cj) => {
    if (cj) {
      schedule.value = cj.schedule
      isSuspended.value = cj.suspend
      containers.value = JSON.parse(JSON.stringify(cj.containers || []))
      errorMessage.value = null
    }
  },
  { immediate: true },
)

const scheduleDescription = computed(() => describeCron(schedule.value))

function setPreset(val: string) {
  schedule.value = val
}

function addEnv(containerIndex: number) {
  if (!containers.value[containerIndex].env) {
    containers.value[containerIndex].env = []
  }
  containers.value[containerIndex].env.push({ name: '', value: '' })
}

function removeEnv(containerIndex: number, envIndex: number) {
  containers.value[containerIndex].env.splice(envIndex, 1)
}

async function handleSave() {
  if (!props.cronJob) return
  isSaving.value = true
  errorMessage.value = null
  try {
    await k8sStore.updateCronJob(props.cronJob.name, {
      schedule: schedule.value,
      suspend: isSuspended.value,
      containers: containers.value,
    })
    emit('saved')
    emit('update:visible', false)
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to update CronJob'
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[95vw] max-w-3xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
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
          <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
            Edit CronJob: {{ cronJob?.name }}
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Modify schedule expression, container images, or env configurations
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

    <div v-if="cronJob" class="p-6 space-y-6 max-h-[75vh] overflow-y-auto">
      <!-- Error Alert -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-300 text-xs flex items-center gap-2"
      >
        <i class="pi pi-exclamation-circle text-base shrink-0"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Schedule Section -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
            <i class="pi pi-clock text-amber-500"></i>
            <span>Schedule Expression</span>
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
              'px-2.5 py-1 text-xs rounded-md border font-mono transition-all cursor-pointer',
              schedule === preset.value
                ? 'bg-sky-600 text-white border-sky-600 shadow-xs'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-200 dark:hover:bg-slate-700'
            ]"
            @click="setPreset(preset.value)"
          >
            {{ preset.label }}
          </button>
        </div>

        <InputText
          v-model="schedule"
          placeholder="* * * * *"
          class="font-mono text-sm w-full"
        />
        <p class="text-[11px] text-slate-400">
          Format: <code>minute hour day-of-month month day-of-week</code> (e.g. <code>0 0 1 * *</code>)
        </p>
      </div>

      <!-- Suspend Toggle -->
      <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <div>
          <div class="text-xs font-bold text-slate-800 dark:text-slate-200">Suspend Schedule Execution</div>
          <div class="text-[11px] text-slate-400 mt-0.5">
            When suspended, automatic scheduled executions are paused without deleting the job template.
          </div>
        </div>
        <button
          type="button"
          :class="[
            'px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors cursor-pointer flex items-center gap-1.5',
            isSuspended
              ? 'bg-amber-500 text-white'
              : 'bg-emerald-600 text-white'
          ]"
          @click="isSuspended = !isSuspended"
        >
          <i :class="isSuspended ? 'pi pi-pause' : 'pi pi-play'"></i>
          <span>{{ isSuspended ? 'Suspended (Paused)' : 'Active (Running)' }}</span>
        </button>
      </div>

      <!-- Containers & Environment Variables -->
      <div class="space-y-4">
        <label class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
          <i class="pi pi-box text-sky-500"></i>
          <span>Container Image & Environment</span>
        </label>

        <div
          v-for="(c, cIdx) in containers"
          :key="c.name || cIdx"
          class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 space-y-3"
        >
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <span class="text-[11px] font-semibold text-slate-500 block mb-1">Container Name</span>
              <InputText v-model="c.name" class="text-xs font-mono w-full" disabled />
            </div>
            <div>
              <span class="text-[11px] font-semibold text-slate-500 block mb-1">Container Image</span>
              <InputText v-model="c.image" class="text-xs font-mono w-full" placeholder="e.g. alpine:latest" />
            </div>
          </div>

          <!-- Environment Variables Table -->
          <div class="pt-2">
            <div class="flex items-center justify-between mb-2">
              <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Environment Variables</span>
              <Button
                label="Add Variable"
                icon="pi pi-plus"
                size="small"
                text
                class="text-xs !py-1 text-sky-600 dark:text-sky-400 font-semibold"
                @click="addEnv(cIdx)"
              />
            </div>

            <div v-if="!c.env || c.env.length === 0" class="text-xs text-slate-400 italic py-2 text-center bg-slate-50 dark:bg-slate-900 rounded-lg">
              No direct environment variables configured.
            </div>

            <div v-else class="space-y-2 max-h-48 overflow-y-auto pr-1">
              <div
                v-for="(envItem, eIdx) in c.env"
                :key="eIdx"
                class="flex items-center gap-2"
              >
                <InputText
                  v-model="envItem.name"
                  placeholder="KEY_NAME"
                  class="text-xs font-mono flex-1"
                />
                <InputText
                  v-model="envItem.value"
                  placeholder="Value"
                  class="text-xs font-mono flex-1"
                />
                <button
                  type="button"
                  class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors cursor-pointer shrink-0"
                  title="Remove Variable"
                  @click="removeEnv(cIdx, eIdx)"
                >
                  <i class="pi pi-trash text-xs"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
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
          label="Save Changes"
          icon="pi pi-check"
          size="small"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          :loading="isSaving"
          @click="handleSave"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
</style>
