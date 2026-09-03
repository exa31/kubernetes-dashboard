<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import { ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { ServiceEndpoints } from '@/types'

const props = defineProps<{
  visible: boolean
  serviceName: string
  namespace: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

const k8sStore = useK8sStore()
const isLoading = ref(false)
const endpoints = ref<ServiceEndpoints | null>(null)
const error = ref<string | null>(null)

watch(
  () => [props.visible, props.serviceName, props.namespace],
  async ([vis]) => {
    if (vis && props.serviceName) {
      loadEndpoints()
    }
  },
  { immediate: true },
)

async function loadEndpoints() {
  isLoading.value = true
  error.value = null
  try {
    endpoints.value = await k8sStore.getServiceEndpoints(props.namespace, props.serviceName)
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch service endpoints'
  } finally {
    isLoading.value = false
  }
}

function handleClose() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[90vw] max-w-3xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
    }"
    @update:visible="emit('update:visible', $event)"
  >
    <!-- Custom Sleek Header -->
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-900/90">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400 shrink-0">
          <i class="pi pi-sitemap text-base"></i>
        </div>
        <div>
          <h3 class="text-sm font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>Routing Endpoints & Targets</span>
            <span class="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
              {{ serviceName }}
            </span>
          </h3>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Active backend Pod IPs and health check state in namespace <strong class="font-mono text-slate-600 dark:text-slate-300">{{ namespace }}</strong>
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
          title="Refresh Endpoints"
          :disabled="isLoading"
          @click="loadEndpoints"
        >
          <i class="pi pi-refresh text-xs" :class="{ 'pi-spin': isLoading }"></i>
        </button>
        <button
          type="button"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
          @click="handleClose"
        >
          <i class="pi pi-times text-xs"></i>
        </button>
      </div>
    </div>

    <!-- Body Content -->
    <div class="p-6 space-y-6 max-h-[75vh] overflow-y-auto">
      <!-- Error notice -->
      <div v-if="error" class="p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
        <i class="pi pi-exclamation-triangle text-sm"></i>
        <span>{{ error }}</span>
      </div>

      <!-- Loading skeleton -->
      <div v-if="isLoading" class="py-12 text-center text-slate-400 space-y-3">
        <i class="pi pi-spin pi-spinner text-3xl text-emerald-500"></i>
        <p class="text-xs font-mono">Resolving endpoint addresses from cluster...</p>
      </div>

      <div v-else-if="endpoints" class="space-y-5">
        <!-- Service Ports Mapping -->
        <div class="space-y-2">
          <h4 class="text-xs font-bold uppercase tracking-wider text-slate-500 flex items-center gap-2">
            <i class="pi pi-link text-xs text-sky-400"></i>
            <span>Exposed Ports</span>
          </h4>
          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5">
            <div
              v-for="p in endpoints.ports"
              :key="p.port"
              class="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 flex items-center justify-between"
            >
              <div>
                <span class="text-xs font-bold text-slate-900 dark:text-slate-100 font-mono">
                  {{ p.name || 'unnamed' }}
                </span>
                <div class="text-[11px] text-slate-400 font-mono">Protocol: {{ p.protocol }}</div>
              </div>
              <span class="px-2.5 py-1 rounded-lg text-xs font-mono font-bold bg-sky-500/10 text-sky-400 border border-sky-500/30">
                :{{ p.port }}
              </span>
            </div>
            <div v-if="endpoints.ports.length === 0" class="col-span-full text-xs text-slate-500 italic p-3">
              No exposed ports found on this endpoint.
            </div>
          </div>
        </div>

        <!-- Target Pods Breakdown -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-xs font-bold uppercase tracking-wider text-slate-500 flex items-center gap-2">
              <i class="pi pi-box text-xs text-emerald-400"></i>
              <span>Backend Pod Targets ({{ endpoints.targets.length }})</span>
            </h4>
            <span class="text-[11px] text-slate-400 font-mono">
              Ready: {{ endpoints.targets.filter(t => t.ready).length }} / {{ endpoints.targets.length }}
            </span>
          </div>

          <div v-if="endpoints.targets.length === 0" class="p-8 text-center rounded-2xl bg-amber-500/5 border border-amber-500/20 text-amber-300">
            <i class="pi pi-info-circle text-3xl mb-2 text-amber-400"></i>
            <h5 class="text-xs font-bold">No Active Pod Targets Found</h5>
            <p class="text-[11px] text-amber-400/80 mt-1 max-w-md mx-auto">
              This Service does not currently point to any running Pod. Check if the Service's selector matches your deployment labels.
            </p>
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="target in endpoints.targets"
              :key="target.ip"
              class="p-3.5 rounded-xl border transition flex items-center justify-between gap-4"
              :class="target.ready
                ? 'bg-emerald-500/5 border-emerald-500/30 hover:border-emerald-500/50'
                : 'bg-rose-500/5 border-rose-500/30 hover:border-rose-500/50'"
            >
              <div class="flex items-center gap-3">
                <div
                  class="w-8 h-8 rounded-lg flex items-center justify-center font-bold shrink-0"
                  :class="target.ready ? 'bg-emerald-500/20 text-emerald-400' : 'bg-rose-500/20 text-rose-400'"
                >
                  <i :class="target.ready ? 'pi pi-check text-xs' : 'pi pi-times text-xs'"></i>
                </div>
                <div>
                  <div class="text-xs font-bold text-slate-900 dark:text-slate-100 font-mono flex items-center gap-2">
                    <span>{{ target.pod_name || 'Direct IP Target' }}</span>
                  </div>
                  <div class="flex items-center gap-3 text-[11px] text-slate-400 font-mono mt-0.5">
                    <span>IP: <strong class="text-slate-300">{{ target.ip }}</strong></span>
                    <span>&bull;</span>
                    <span>Node: <strong class="text-slate-300">{{ target.node_name || 'unassigned' }}</strong></span>
                  </div>
                </div>
              </div>

              <Tag
                :value="target.ready ? 'Healthy / Ready' : 'Not Ready'"
                :severity="target.ready ? 'success' : 'danger'"
                class="font-mono text-[11px] px-2.5 py-0.5"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="px-6 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/90 flex justify-end">
      <Button
        label="Close"
        size="small"
        class="btn-slate text-xs px-4 py-1.5 rounded-lg active:scale-95 cursor-pointer"
        @click="handleClose"
      />
    </div>
  </Dialog>
</template>
