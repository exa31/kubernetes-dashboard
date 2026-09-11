<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { computed, ref, watch } from 'vue'

import { useAuthStore, useK8sStore } from '@/stores'
import type { DeploymentRevision } from '@/types'

const props = defineProps<{
  visible: boolean
  deploymentName: string
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'rolledBack'): void
}>()

const authStore = useAuthStore()
const k8sStore = useK8sStore()
const confirm = useConfirm()
const toast = useToast()

const revisions = ref<DeploymentRevision[]>([])
const isLoading = ref(false)
const isRollingBack = ref(false)

const canMutate = computed(() => authStore.canMutateNamespace(props.namespace))
const hasPreviousRevision = computed(() => revisions.value.some((r) => !r.is_current))

watch(
  () => [props.visible, props.deploymentName],
  async ([vis, name]) => {
    if (vis && name) {
      await loadHistory()
    }
  },
  { immediate: true }
)

async function loadHistory() {
  if (!props.deploymentName) return
  isLoading.value = true
  try {
    revisions.value = await k8sStore.getDeploymentHistory(props.deploymentName, props.namespace)
  } catch (err: unknown) {
    revisions.value = []
    toast.add({
      severity: 'error',
      summary: 'Failed to load history',
      detail: err instanceof Error ? err.message : 'Unknown error',
      life: 4000
    })
  } finally {
    isLoading.value = false
  }
}

function handleRollback(targetRevision?: number) {
  const isPrevious = !targetRevision || targetRevision === 0
  const title = isPrevious
    ? 'Rollback to Previous Version'
    : `Rollback to Revision #${targetRevision}`
  const message = isPrevious
    ? `Rollback deployment '${props.deploymentName}' to its immediate previous revision? Pods will restart using that configuration.`
    : `Rollback deployment '${props.deploymentName}' to revision #${targetRevision}? Pods will restart using that revision's template.`

  confirm.require({
    message,
    header: title,
    icon: 'pi pi-undo',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Confirm Rollback',
      severity: 'warn'
    },
    accept: async () => {
      isRollingBack.value = true
      try {
        const res = await k8sStore.rollbackDeployment(
          props.deploymentName,
          targetRevision,
          props.namespace
        )
        toast.add({
          severity: 'success',
          summary: 'Rollback Succeeded',
          detail: res.message || `Rolled back to revision #${res.to_revision}`,
          life: 4000
        })
        emit('rolledBack')
        await loadHistory()
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Rollback Failed',
          detail: err instanceof Error ? err.message : 'Unknown error occurred',
          life: 5000
        })
      } finally {
        isRollingBack.value = false
      }
    }
  })
}

function formatChangeCause(cause?: string): string {
  if (!cause || cause === '<none>') {
    return 'None'
  }
  return cause
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
    <div
      class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-500 flex items-center justify-center font-bold text-lg shrink-0"
        >
          <i class="pi pi-undo"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
              Rollout History & Rollback: {{ deploymentName }}
            </h2>
            <Tag value="Deployments" severity="info" class="text-[10px] font-mono px-2 py-0.5" />
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Namespace:
            <span class="font-mono text-slate-700 dark:text-slate-300 font-semibold">{{
              namespace
            }}</span>
            &bull; Inspect previous revisions and rollback pod template
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

    <!-- Modal Body -->
    <div class="p-6 space-y-4">
      <!-- Action & Summary Toolbar -->
      <div
        class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-slate-50 dark:bg-slate-950/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-800/80"
      >
        <div class="flex items-center gap-2.5">
          <div
            class="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0"
          >
            <i class="pi pi-info-circle text-sm"></i>
          </div>
          <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
            Revisions correspond to ReplicaSets created during rollout updates. Rolling back safely
            restores the pod template spec.
          </p>
        </div>

        <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
          <Button
            label="Refresh"
            icon="pi pi-refresh"
            size="small"
            severity="secondary"
            outlined
            :loading="isLoading"
            class="text-xs px-3 py-1.5 rounded-lg cursor-pointer"
            @click="loadHistory"
          />

          <Button
            label="Rollback to Previous (Undo)"
            icon="pi pi-undo"
            size="small"
            severity="warn"
            :loading="isRollingBack"
            :disabled="!canMutate || !hasPreviousRevision || isLoading"
            :title="
              !canMutate
                ? 'Read-only: cannot rollback'
                : !hasPreviousRevision
                  ? 'No previous revision available'
                  : 'Rollback to immediate previous revision'
            "
            class="text-xs px-3.5 py-1.5 rounded-lg cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
            @click="handleRollback(0)"
          />
        </div>
      </div>

      <!-- Revisions Table -->
      <div
        class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden bg-white dark:bg-slate-950"
      >
        <DataTable
          :value="revisions"
          :loading="isLoading"
          striped-rows
          responsive-layout="scroll"
          class="p-datatable-sm w-full"
        >
          <!-- Revision Column -->
          <Column field="revision" header="Rev" style="width: 7rem">
            <template #body="{ data }">
              <div class="flex items-center gap-2">
                <span
                  class="font-mono text-xs font-bold px-2 py-0.5 rounded-md bg-slate-200 dark:bg-slate-800 text-slate-800 dark:text-slate-200"
                >
                  #{{ data.revision }}
                </span>
                <Tag
                  v-if="data.is_current"
                  value="Current"
                  severity="success"
                  class="text-[10px] font-mono uppercase px-1.5 py-0.2"
                />
              </div>
            </template>
          </Column>

          <!-- ReplicaSet & Images -->
          <Column header="Containers & Images" style="min-width: 16rem">
            <template #body="{ data }">
              <div class="space-y-1">
                <div class="flex items-center gap-1.5 text-[11px] font-mono text-slate-400">
                  <i class="pi pi-box text-[10px]"></i>
                  <span>{{ data.replicaset }}</span>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="img in data.images"
                    :key="img"
                    class="font-mono text-xs text-sky-400 bg-sky-500/10 px-2 py-0.5 rounded border border-sky-500/20 max-w-sm truncate"
                    :title="img"
                  >
                    {{ img }}
                  </span>
                </div>
              </div>
            </template>
          </Column>

          <!-- Change Cause -->
          <Column field="change_cause" header="Change Cause" style="min-width: 12rem">
            <template #body="{ data }">
              <span
                class="text-xs"
                :class="
                  data.change_cause && data.change_cause !== '<none>'
                    ? 'text-slate-700 dark:text-slate-300 font-medium'
                    : 'text-slate-400 dark:text-slate-500 italic'
                "
              >
                {{ formatChangeCause(data.change_cause) }}
              </span>
            </template>
          </Column>

          <!-- Age -->
          <Column field="age" header="Age" style="width: 7rem">
            <template #body="{ data }">
              <span
                class="text-xs font-mono text-slate-500 dark:text-slate-400"
                :title="new Date(data.created_at).toLocaleString()"
              >
                {{ data.age }}
              </span>
            </template>
          </Column>

          <!-- Action -->
          <Column header="Action" style="width: 10rem; text-align: right">
            <template #body="{ data }">
              <div class="flex justify-end">
                <Button
                  v-if="!data.is_current"
                  label="Rollback to this"
                  icon="pi pi-undo"
                  size="small"
                  class="btn-amber text-xs px-2.5 py-1 rounded-lg cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!canMutate || isRollingBack"
                  :title="
                    !canMutate
                      ? 'Read-only: cannot rollback'
                      : `Rollback to revision #${data.revision}`
                  "
                  @click="handleRollback(data.revision)"
                />
                <span
                  v-else
                  class="text-[11px] font-semibold text-emerald-400 flex items-center gap-1.5 px-2 py-1 rounded bg-emerald-500/10 border border-emerald-500/20"
                >
                  <i class="pi pi-check text-[10px]"></i>
                  Active Version
                </span>
              </div>
            </template>
          </Column>

          <template #empty>
            <div class="py-10 text-center text-slate-400">
              <i class="pi pi-history text-3xl mb-2 text-slate-500"></i>
              <h3 class="font-semibold text-slate-200">No Rollout History Found</h3>
              <p class="text-xs text-slate-500 mt-1">
                No replica sets or previous versions available for this deployment.
              </p>
            </div>
          </template>
        </DataTable>
      </div>
    </div>
  </Dialog>
</template>
