<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import { ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { JobItem } from '@/types'

import PodLogsDialog from './PodLogsDialog.vue'

const props = defineProps<{
  visible: boolean
  cronJobName: string
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

const k8sStore = useK8sStore()

const jobs = ref<JobItem[]>([])
const isLoading = ref(false)

// Pod logs modal state
const isLogsOpen = ref(false)
const selectedPodName = ref('')

watch(
  () => [props.visible, props.cronJobName],
  async ([vis, name]) => {
    if (vis && name) {
      await loadHistory()
    }
  },
  { immediate: true },
)

async function loadHistory() {
  if (!props.cronJobName) return
  isLoading.value = true
  try {
    jobs.value = await k8sStore.getCronJobJobs(props.cronJobName, props.namespace)
  } catch {
    jobs.value = []
  } finally {
    isLoading.value = false
  }
}

async function viewJobLogs(job: JobItem) {
  // Pods spawned by Job typically have the same prefix as Job Name
  selectedPodName.value = job.name
  isLogsOpen.value = true
}

function getStatusSeverity(status: string): 'success' | 'info' | 'danger' {
  switch (status) {
    case 'Complete':
      return 'success'
    case 'Running':
      return 'info'
    case 'Failed':
      return 'danger'
    default:
      return 'info'
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[95vw] max-w-4xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
    }"
    @update:visible="(val) => emit('update:visible', val)"
  >
    <!-- Custom Header -->
    <div class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-sky-500/10 text-sky-500 flex items-center justify-center font-bold text-base shrink-0">
          <i class="pi pi-history"></i>
        </div>
        <div>
          <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
            Execution History: {{ cronJobName }}
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Batch Jobs instantiated in namespace <span class="font-mono text-slate-700 dark:text-slate-300 font-semibold">{{ namespace }}</span>
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
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500 dark:text-slate-400">
          Historical batch Jobs instantiated by <strong class="font-mono text-slate-800 dark:text-slate-200">{{ cronJobName }}</strong> in namespace <strong class="font-mono text-slate-800 dark:text-slate-200">{{ namespace }}</strong>.
        </p>
        <Button
          label="Refresh"
          icon="pi pi-refresh"
          size="small"
          severity="secondary"
          outlined
          :loading="isLoading"
          class="text-xs"
          @click="loadHistory"
        />
      </div>

      <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden bg-white dark:bg-slate-950">
        <DataTable
          :value="jobs"
          :loading="isLoading"
          striped-rows
          paginator
          :rows="10"
          table-style="min-width: 100%"
          class="p-datatable-sm w-full"
        >
          <!-- Job Name -->
          <Column field="name" header="Job Name">
            <template #body="{ data }">
              <div class="flex items-center gap-2 font-mono text-xs font-semibold text-slate-800 dark:text-slate-200">
                <i class="pi pi-bolt text-amber-500"></i>
                <span>{{ data.name }}</span>
              </div>
            </template>
          </Column>

          <!-- Status -->
          <Column field="status" header="Status" style="width: 130px">
            <template #body="{ data }">
              <Tag
                :value="data.status"
                :severity="getStatusSeverity(data.status)"
                class="text-xs font-mono"
              />
            </template>
          </Column>

          <!-- Duration -->
          <Column field="duration" header="Duration" style="width: 120px">
            <template #body="{ data }">
              <span class="font-mono text-xs text-slate-600 dark:text-slate-300">
                {{ data.duration }}
              </span>
            </template>
          </Column>

          <!-- Age / Start Time -->
          <Column field="age" header="Age" style="width: 100px">
            <template #body="{ data }">
              <span class="text-xs text-slate-400">
                {{ data.age }}
              </span>
            </template>
          </Column>

          <!-- Action: View Logs -->
          <Column header="Logs" header-style="text-align: right" body-style="text-align: right" style="width: 120px">
            <template #body="{ data }">
              <Button
                label="Logs"
                icon="pi pi-terminal"
                size="small"
                class="bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-200 border-none text-xs py-1"
                @click="viewJobLogs(data)"
              />
            </template>
          </Column>

          <template #empty>
            <div class="py-12 text-center text-slate-400">
              <i class="pi pi-history text-3xl mb-2 text-slate-300 dark:text-slate-700"></i>
              <p class="text-xs">No execution history recorded for this CronJob yet.</p>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- Pod Logs Modal Viewer -->
    <PodLogsDialog
      v-model:visible="isLogsOpen"
      :pod-name="selectedPodName"
      :namespace="namespace"
    />
  </Dialog>
</template>

<style scoped>
</style>
