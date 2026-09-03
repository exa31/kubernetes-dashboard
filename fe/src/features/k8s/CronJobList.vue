<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { computed, onMounted, ref } from 'vue'

import { useK8sStore } from '@/stores'
import type { CronJobDetail, CronJobItem } from '@/types'
import { describeCron, formatDate } from '@/utils'

import CreateCronJobDialog from './CreateCronJobDialog.vue'
import CronJobEditorDialog from './CronJobEditorDialog.vue'
import CronJobHistoryDialog from './CronJobHistoryDialog.vue'
import ResourceYamlDialog from './ResourceYamlDialog.vue'

const k8sStore = useK8sStore()
const { cronjobs, selectedNamespace, isLoading, isActionLoading } = storeToRefs(k8sStore)

const searchQuery = ref('')
const isCreateOpen = ref(false)
const isEditorOpen = ref(false)
const isHistoryOpen = ref(false)
const isYamlOpen = ref(false)
const selectedYamlName = ref('')
const selectedCronJob = ref<CronJobDetail | null>(null)
const selectedCronJobName = ref('')
const notification = ref<{ type: 'success' | 'error'; message: string } | null>(null)

function openYaml(name: string) {
  selectedYamlName.value = name
  isYamlOpen.value = true
}

onMounted(() => {
  k8sStore.fetchCronJobs()
})

const filteredCronJobs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return cronjobs.value
  return cronjobs.value.filter(
    (cj) =>
      cj.name.toLowerCase().includes(q) ||
      cj.schedule.toLowerCase().includes(q) ||
      cj.image.toLowerCase().includes(q) ||
      describeCron(cj.schedule).toLowerCase().includes(q),
  )
})

async function openEditor(cj: CronJobItem) {
  try {
    const detail = await k8sStore.getCronJobDetail(cj.name, cj.namespace)
    selectedCronJob.value = detail
    isEditorOpen.value = true
  } catch {
    notification.value = {
      type: 'error',
      message: `Failed to load details for ${cj.name}`,
    }
  }
}

function openHistory(cj: CronJobItem) {
  selectedCronJobName.value = cj.name
  isHistoryOpen.value = true
}

async function handleRunNow(cj: CronJobItem) {
  try {
    const job = await k8sStore.triggerCronJobNow(cj.name, cj.namespace)
    notification.value = {
      type: 'success',
      message: `Job '${job.name}' triggered successfully for ${cj.name}!`,
    }
  } catch (err: unknown) {
    notification.value = {
      type: 'error',
      message: err instanceof Error ? err.message : 'Failed to trigger job',
    }
  }
}

async function handleToggleSuspend(cj: CronJobItem) {
  try {
    const isSuspended = await k8sStore.toggleSuspendCronJob(cj.name, cj.namespace)
    notification.value = {
      type: 'success',
      message: `${cj.name} is now ${isSuspended ? 'Suspended (Paused)' : 'Active'}`,
    }
  } catch (err: unknown) {
    notification.value = {
      type: 'error',
      message: err instanceof Error ? err.message : 'Failed to toggle suspend',
    }
  }
}

async function handleDelete(cj: CronJobItem) {
  if (!confirm(`Are you sure you want to delete CronJob '${cj.name}'?`)) {
    return
  }
  try {
    await k8sStore.deleteCronJob(cj.name, cj.namespace)
    notification.value = {
      type: 'success',
      message: `CronJob '${cj.name}' deleted successfully`,
    }
  } catch (err: unknown) {
    notification.value = {
      type: 'error',
      message: err instanceof Error ? err.message : 'Failed to delete CronJob',
    }
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Top toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-clock text-amber-500"></i>
          <span>CronJobs & Scheduled Tasks</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Automated batch tasks, schedules, and manual run executions in <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ selectedNamespace }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText
            v-model="searchQuery"
            placeholder="Search cronjobs & schedule..."
            class="text-sm w-64"
          />
        </IconField>

        <Button
          label="Refresh"
          icon="pi pi-refresh"
          severity="secondary"
          outlined
          size="small"
          :loading="isLoading"
          @click="k8sStore.fetchCronJobs()"
        />

        <Button
          label="Create CronJob"
          icon="pi pi-plus"
          size="small"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          @click="isCreateOpen = true"
        />
      </div>
    </div>

    <!-- Notification Banner -->
    <div
      v-if="notification"
      :class="[
        'p-3.5 rounded-xl border text-xs flex items-center justify-between transition-all shadow-sm',
        notification.type === 'success'
          ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300'
          : 'bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800 text-rose-700 dark:text-rose-300'
      ]"
    >
      <div class="flex items-center gap-2 font-medium">
        <i :class="notification.type === 'success' ? 'pi pi-check-circle' : 'pi pi-exclamation-circle'"></i>
        <span>{{ notification.message }}</span>
      </div>
      <button class="text-xs hover:underline cursor-pointer" @click="notification = null">Dismiss</button>
    </div>

    <!-- PrimeVue DataTable for CronJobs -->
    <div class="w-full rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden bg-white dark:bg-slate-950">
      <DataTable
        :value="filteredCronJobs"
        :loading="isLoading"
        striped-rows
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
        table-style="min-width: 100%"
        class="p-datatable-sm w-full"
      >
        <!-- Name Column -->
        <Column field="name" header="CronJob Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1">
              <div class="w-8 h-8 rounded-lg bg-amber-500/10 text-amber-500 flex items-center justify-center font-bold shrink-0">
                <i class="pi pi-clock text-xs"></i>
              </div>
              <div>
                <div
                  class="font-semibold text-slate-900 dark:text-slate-100 font-mono text-sm hover:text-sky-600 transition-colors cursor-pointer"
                  @click="openEditor(data)"
                >
                  {{ data.name }}
                </div>
                <div class="text-xs text-slate-400 mt-0.5 truncate font-mono">
                  {{ data.image || 'no image' }} &bull; Age: {{ data.age }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Schedule Column -->
        <Column field="schedule" header="Schedule" sortable>
          <template #body="{ data }">
            <div>
              <span class="px-2 py-0.5 rounded font-mono text-xs bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200 border border-slate-200 dark:border-slate-700/60 font-semibold">
                {{ data.schedule }}
              </span>
              <div class="text-[11px] text-slate-500 dark:text-slate-400 mt-1">
                {{ describeCron(data.schedule) }}
              </div>
            </div>
          </template>
        </Column>

        <!-- Status Column -->
        <Column header="Status" style="width: 150px">
          <template #body="{ data }">
            <div class="space-y-1">
              <Tag
                :value="data.suspend ? 'Suspended' : 'Active'"
                :severity="data.suspend ? 'warn' : 'success'"
                class="font-mono text-xs"
              />
              <div v-if="data.active_jobs > 0" class="text-[10px] text-sky-600 dark:text-sky-400 font-mono font-semibold">
                {{ data.active_jobs }} running
              </div>
            </div>
          </template>
        </Column>

        <!-- Last Schedule Column -->
        <Column header="Last Schedule" style="width: 170px">
          <template #body="{ data }">
            <div v-if="data.last_schedule_time" class="text-xs text-slate-600 dark:text-slate-300 font-mono">
              {{ formatDate(data.last_schedule_time) }}
            </div>
            <span v-else class="text-xs text-slate-400 italic">Never scheduled</span>
          </template>
        </Column>

        <!-- Actions Column -->
        <Column header="Actions" header-style="text-align: right" body-style="text-align: right" style="width: 250px">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <!-- Run Now Button -->
              <Button
                title="Trigger immediate execution"
                icon="pi pi-bolt"
                size="small"
                class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :loading="isActionLoading"
                @click="handleRunNow(data)"
              />

              <!-- Suspend/Resume Toggle Button -->
              <Button
                :title="data.suspend ? 'Resume automatic schedule' : 'Pause schedule'"
                :icon="data.suspend ? 'pi pi-play' : 'pi pi-pause'"
                size="small"
                :class="data.suspend ? 'btn-emerald' : 'btn-amber'"
                class="text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="handleToggleSuspend(data)"
              />

              <!-- History & Logs Button -->
              <Button
                title="View execution history & logs"
                icon="pi pi-history"
                size="small"
                class="btn-sky text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="openHistory(data)"
              />

              <!-- YAML Button -->
              <Button
                title="Inspect & Edit YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="openYaml(data.name)"
              />

              <!-- Edit Button -->
              <Button
                title="Edit schedule & container"
                icon="pi pi-pencil"
                size="small"
                class="btn-blue text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="openEditor(data)"
              />

              <!-- Delete Button -->
              <Button
                title="Delete CronJob"
                icon="pi pi-trash"
                size="small"
                class="btn-rose text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="handleDelete(data)"
              />
            </div>
          </template>
        </Column>

        <!-- Empty State -->
        <template #empty>
          <div class="py-16 text-center text-slate-400">
            <i class="pi pi-clock text-4xl mb-3 text-slate-300 dark:text-slate-700"></i>
            <h3 class="font-semibold text-slate-700 dark:text-slate-300">No CronJobs Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no CronJobs configured in namespace {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Dialogs -->
    <CreateCronJobDialog
      v-model:visible="isCreateOpen"
      :namespace="selectedNamespace"
      @created="k8sStore.fetchCronJobs()"
    />

    <CronJobEditorDialog
      v-model:visible="isEditorOpen"
      :cron-job="selectedCronJob"
      @saved="k8sStore.fetchCronJobs()"
    />

    <CronJobHistoryDialog
      v-model:visible="isHistoryOpen"
      :cron-job-name="selectedCronJobName"
      :namespace="selectedNamespace"
    />

    <!-- Resource YAML Dialog -->
    <ResourceYamlDialog
      v-model:visible="isYamlOpen"
      kind="CronJob"
      :name="selectedYamlName"
      :namespace="selectedNamespace"
      @applied="k8sStore.fetchCronJobs()"
    />
  </div>
</template>

<style scoped>
</style>
