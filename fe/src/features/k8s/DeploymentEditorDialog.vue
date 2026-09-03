<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { ContainerDetail, DeploymentDetail } from '@/types'
import { logger } from '@/utils'

const props = defineProps<{
  visible: boolean
  deploymentName: string
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'saved'): void
}>()

const k8sStore = useK8sStore()

const deployment = ref<DeploymentDetail | null>(null)
const replicas = ref<number>(1)
const containers = ref<ContainerDetail[]>([])
const isLoading = ref<boolean>(false)
const isSaving = ref<boolean>(false)
const statusMessage = ref<{ text: string; severity: 'success' | 'error' } | null>(null)

async function loadDeployment() {
  if (!props.deploymentName) return
  isLoading.value = true
  statusMessage.value = null
  try {
    const data = await k8sStore.getDeploymentDetail(props.deploymentName, props.namespace)
    deployment.value = data
    replicas.value = data.replicas
    containers.value = JSON.parse(JSON.stringify(data.containers))
  } catch (err: unknown) {
    logger.error('Failed to load deployment details', err)
    statusMessage.value = {
      text: err instanceof Error ? err.message : 'Failed to load deployment',
      severity: 'error',
    }
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.visible,
  (open) => {
    if (open) {
      loadDeployment()
    }
  },
)

function addEnvVar(containerIndex: number) {
  if (!containers.value[containerIndex].env) {
    containers.value[containerIndex].env = []
  }
  containers.value[containerIndex].env.push({
    name: 'NEW_KEY',
    value: 'value',
  })
}

function removeEnvVar(containerIndex: number, envIndex: number) {
  containers.value[containerIndex].env.splice(envIndex, 1)
}

async function saveChanges() {
  isSaving.value = true
  statusMessage.value = null
  try {
    await k8sStore.updateDeployment(
      props.deploymentName,
      {
        replicas: replicas.value,
        containers: containers.value,
      },
      props.namespace,
    )
    statusMessage.value = {
      text: `Deployment '${props.deploymentName}' updated and rollout restart triggered!`,
      severity: 'success',
    }
    emit('saved')
  } catch (err: unknown) {
    logger.error('Failed to update deployment', err)
    let msg = 'Failed to update deployment'
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      if (res?.data?.message) msg = res.data.message
    } else if (err instanceof Error) {
      msg = err.message
    }
    statusMessage.value = { text: msg, severity: 'error' }
  } finally {
    isSaving.value = false
  }
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
    class="w-[90vw] max-w-4xl h-[80vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 h-full flex flex-col overflow-hidden' }
    }"
  >
    <!-- Header -->
    <div class="p-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-500 flex items-center justify-center font-bold text-sm">
          <i class="pi pi-server"></i>
        </div>
        <div>
          <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
            Edit Deployment: {{ deploymentName }}
          </h2>
          <p class="text-xs text-slate-500">Namespace: <span class="font-mono text-slate-700 dark:text-slate-300">{{ namespace }}</span></p>
        </div>
      </div>

      <Button
        icon="pi pi-times"
        severity="secondary"
        text
        rounded
        size="small"
        @click="closeDialog"
      />
    </div>

    <!-- Alert / Message -->
    <div v-if="statusMessage" class="p-4 shrink-0">
      <Message :severity="statusMessage.severity" :closable="false" class="text-xs">
        {{ statusMessage.text }}
      </Message>
    </div>

    <!-- Body -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6 bg-slate-50/50 dark:bg-slate-950">
      <div v-if="isLoading" class="py-12 text-center text-slate-400">
        <i class="pi pi-spin pi-spinner text-3xl mb-2"></i>
        <p class="text-xs">Loading deployment details...</p>
      </div>

      <div v-else class="space-y-6">
        <!-- Replicas section -->
        <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
            Pod Replicas (Scaling)
          </label>
          <div class="flex items-center gap-3 mt-2">
            <InputNumber
              v-model="replicas"
              show-buttons
              button-layout="horizontal"
              :min="0"
              :max="100"
              class="w-48 font-mono text-sm"
            />
            <span class="text-xs text-slate-500">Currently {{ deployment?.ready_replicas ?? 0 }} ready pods</span>
          </div>
        </div>

        <!-- Containers section -->
        <div
          v-for="(c, cIdx) in containers"
          :key="c.name"
          class="p-5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4"
        >
          <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
            <div class="flex items-center gap-2">
              <i class="pi pi-box text-sky-500 text-sm"></i>
              <span class="font-bold text-sm font-mono text-slate-900 dark:text-slate-100">{{ c.name }}</span>
            </div>
          </div>

          <!-- Container Image -->
          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Container Image
            </label>
            <InputText
              v-model="c.image"
              placeholder="e.g. nginx:alpine or ghcr.io/..."
              class="w-full font-mono text-xs py-2"
            />
          </div>

          <!-- Container Environment Variables -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                Direct Container Environment Variables
              </label>
              <Button
                label="Add Env"
                icon="pi pi-plus"
                size="small"
                text
                class="text-xs py-0.5 text-sky-600 dark:text-sky-400 font-semibold"
                @click="addEnvVar(cIdx)"
              />
            </div>

            <div class="space-y-2">
              <div
                v-for="(env, eIdx) in c.env"
                :key="eIdx"
                class="flex items-center gap-2"
              >
                <InputText
                  v-model="env.name"
                  placeholder="KEY"
                  class="w-1/3 font-mono text-xs py-1.5"
                />
                <span class="text-slate-400 font-mono">=</span>
                <InputText
                  v-model="env.value"
                  placeholder="VALUE"
                  class="flex-1 font-mono text-xs py-1.5"
                />
                <Button
                  icon="pi pi-trash"
                  size="small"
                  severity="danger"
                  text
                  rounded
                  class="text-xs"
                  @click="removeEnvVar(cIdx, eIdx)"
                />
              </div>

              <div v-if="!c.env || c.env.length === 0" class="text-xs text-slate-400 italic py-2">
                No direct container environment variables configured.
              </div>
            </div>
          </div>

          <!-- Mounted EnvFrom Sources Info -->
          <div v-if="c.env_from && c.env_from.length > 0" class="pt-2 border-t border-slate-100 dark:border-slate-800/60">
            <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider block mb-2">
              Mounted Config / Secret Sources:
            </span>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="ef in c.env_from"
                :key="ef.name"
                class="text-xs px-2.5 py-1 rounded-md font-mono flex items-center gap-1.5 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300"
              >
                <i :class="ef.type === 'secret' ? 'pi pi-lock text-amber-500' : 'pi pi-file text-sky-500'" class="text-[10px]"></i>
                <span>{{ ef.name }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 flex items-center justify-end gap-2 shrink-0">
      <Button label="Cancel" severity="secondary" text size="small" @click="closeDialog" />
      <Button
        label="Save & Rollout Restart"
        icon="pi pi-check"
        size="small"
        class="btn-emerald text-xs shadow-xs cursor-pointer"
        :loading="isSaving"
        @click="saveChanges"
      />
    </div>
  </Dialog>
</template>

<style scoped>
</style>
