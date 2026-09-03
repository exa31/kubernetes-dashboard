<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { computed, onMounted, ref } from 'vue'

import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'
import ServiceEndpointsDialog from '@/features/k8s/ServiceEndpointsDialog.vue'
import { useK8sStore } from '@/stores'
import type { ServiceDetail, ServiceItem } from '@/types'

const k8sStore = useK8sStore()
const { services, selectedNamespace, isLoading } = storeToRefs(k8sStore)

const searchQuery = ref('')
const selectedService = ref<ServiceDetail | null>(null)
const isDetailOpen = ref(false)
const isLoadingDetail = ref(false)

const isYamlOpen = ref(false)
const selectedYamlName = ref('')

const isEndpointsOpen = ref(false)
const selectedEndpointsName = ref('')

function openYaml(name: string) {
  selectedYamlName.value = name
  isYamlOpen.value = true
}

function openEndpoints(name: string) {
  selectedEndpointsName.value = name
  isEndpointsOpen.value = true
}

onMounted(() => {
  k8sStore.fetchServices()
})

const filteredServices = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return services.value
  return services.value.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.type.toLowerCase().includes(q) ||
      s.cluster_ip.toLowerCase().includes(q) ||
      (s.external_ip && s.external_ip.toLowerCase().includes(q)) ||
      s.ports.some((p) => String(p.port).includes(q) || p.name.toLowerCase().includes(q)),
  )
})

async function openDetail(svc: ServiceItem) {
  isLoadingDetail.value = true
  isDetailOpen.value = true
  try {
    const detail = await k8sStore.getServiceDetail(svc.name, svc.namespace)
    selectedService.value = detail
  } catch {
    selectedService.value = {
      ...svc,
      labels: {},
      annotations: {},
    }
  } finally {
    isLoadingDetail.value = false
  }
}

function copyToClipboard(text: string) {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text)
  }
}

function getTypeSeverity(type: string): 'info' | 'success' | 'warn' | 'secondary' {
  switch (type) {
    case 'ClusterIP':
      return 'info'
    case 'LoadBalancer':
      return 'success'
    case 'NodePort':
      return 'warn'
    case 'ExternalName':
      return 'secondary'
    default:
      return 'info'
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Top toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-share-alt text-teal-500"></i>
          <span>Services</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Cluster networking, internal endpoints, and port mappings in <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ selectedNamespace }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText
            v-model="searchQuery"
            placeholder="Search services..."
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
          @click="k8sStore.fetchServices()"
        />
      </div>
    </div>

    <!-- PrimeVue DataTable for Services -->
    <div class="w-full rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden bg-white dark:bg-slate-950">
      <DataTable
        :value="filteredServices"
        :loading="isLoading"
        striped-rows
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
        table-style="min-width: 100%"
        class="p-datatable-sm w-full"
        row-hover
        @row-click="(e) => openDetail(e.data)"
      >
        <!-- Name Column -->
        <Column field="name" header="Service Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1 cursor-pointer">
              <div class="w-8 h-8 rounded-lg bg-teal-500/10 text-teal-500 flex items-center justify-center font-bold shrink-0">
                <i class="pi pi-share-alt text-xs"></i>
              </div>
              <div>
                <div class="font-semibold text-slate-900 dark:text-slate-100 font-mono text-sm hover:text-teal-600 transition-colors">
                  {{ data.name }}
                </div>
                <div class="text-xs text-slate-400 mt-0.5">
                  Age: {{ data.age }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Type Column -->
        <Column field="type" header="Type" sortable style="width: 140px">
          <template #body="{ data }">
            <Tag
              :value="data.type"
              :severity="getTypeSeverity(data.type)"
              class="font-mono text-xs"
            />
          </template>
        </Column>

        <!-- IP / Target Column -->
        <Column header="Cluster IP / Target" style="width: 220px">
          <template #body="{ data }">
            <div v-if="data.cluster_ip" class="flex items-center gap-1.5 font-mono text-xs text-slate-700 dark:text-slate-300">
              <span>{{ data.cluster_ip }}</span>
              <button
                type="button"
                class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer text-[10px]"
                title="Copy Cluster IP"
                @click.stop="copyToClipboard(data.cluster_ip)"
              >
                <i class="pi pi-copy"></i>
              </button>
            </div>
            <div v-else-if="data.external_ip" class="font-mono text-xs text-slate-500 truncate max-w-xs" :title="data.external_ip">
              {{ data.external_ip }}
            </div>
            <span v-else class="text-slate-400 text-xs italic">None</span>
          </template>
        </Column>

        <!-- Port Mappings Column -->
        <Column header="Ports">
          <template #body="{ data }">
            <div class="flex flex-wrap gap-1">
              <span
                v-for="p in data.ports"
                :key="p.port"
                class="text-[11px] px-2 py-0.5 rounded font-mono bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700/60"
              >
                {{ p.port }}<span v-if="p.target_port && p.target_port !== String(p.port)" class="text-slate-400">&rarr;{{ p.target_port }}</span>/{{ p.protocol }}
              </span>
              <span v-if="data.ports.length === 0" class="text-xs text-slate-400 italic">None</span>
            </div>
          </template>
        </Column>

        <!-- Pod Selector Column -->
        <Column header="Pod Selector">
          <template #body="{ data }">
            <div v-if="data.selector && Object.keys(data.selector).length > 0" class="flex flex-wrap gap-1 max-w-xs">
              <span
                v-for="(val, key) in data.selector"
                :key="key"
                class="text-[11px] px-2 py-0.5 rounded font-mono bg-sky-50 dark:bg-sky-950/40 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800/50"
              >
                {{ key }}={{ val }}
              </span>
            </div>
            <span v-else class="text-xs text-slate-400 italic">None</span>
          </template>
        </Column>

        <!-- Actions Column -->
        <Column header="Actions" header-style="text-align: right" body-style="text-align: right" style="width: 250px">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5">
              <Button
                label="Endpoints"
                icon="pi pi-sitemap"
                size="small"
                class="btn-emerald text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                title="Inspect backend Pod IPs and health"
                @click.stop="openEndpoints(data.name)"
              />
              <Button
                label="YAML"
                icon="pi pi-code"
                size="small"
                class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click.stop="openYaml(data.name)"
              />
              <Button
                label="Detail"
                icon="pi pi-eye"
                size="small"
                class="btn-sky text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click.stop="openDetail(data)"
              />
            </div>
          </template>
        </Column>

        <!-- Empty state -->
        <template #empty>
          <div class="py-16 text-center text-slate-400">
            <i class="pi pi-share-alt text-4xl mb-3 text-slate-300 dark:text-slate-700"></i>
            <h3 class="font-semibold text-slate-700 dark:text-slate-300">No Services Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no Services in namespace {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Service Detail Dialog -->
    <Dialog
      v-model:visible="isDetailOpen"
      modal
      :show-header="false"
      class="w-[90vw] max-w-2xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
      :pt="{
        root: { class: 'border-none p-0 overflow-hidden' },
        content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
      }"
    >
      <!-- Custom Header -->
      <div class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-teal-500/10 text-teal-500 flex items-center justify-center font-bold text-base shrink-0">
            <i class="pi pi-compass"></i>
          </div>
          <div>
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100 font-mono">
              Service: {{ selectedService?.name }}
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Network endpoint routing, port mapping, and pod selector
            </p>
          </div>
        </div>
        <button
          type="button"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          @click="isDetailOpen = false"
        >
          <i class="pi pi-times"></i>
        </button>
      </div>

      <div v-if="selectedService" class="p-6 space-y-4 text-xs font-mono">
        <div class="grid grid-cols-2 gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
          <div>
            <span class="text-slate-400 block text-[10px] uppercase">Type</span>
            <span class="font-semibold text-slate-800 dark:text-slate-200">{{ selectedService.type }}</span>
          </div>
          <div>
            <span class="text-slate-400 block text-[10px] uppercase">Cluster IP</span>
            <span class="font-semibold text-slate-800 dark:text-slate-200">{{ selectedService.cluster_ip || 'None' }}</span>
          </div>
          <div v-if="selectedService.external_ip" class="col-span-2">
            <span class="text-slate-400 block text-[10px] uppercase">External Name / IP</span>
            <span class="font-semibold text-slate-800 dark:text-slate-200 break-all">{{ selectedService.external_ip }}</span>
          </div>
        </div>

        <!-- Port mappings -->
        <div>
          <h4 class="text-xs font-bold uppercase text-slate-500 mb-2">Port Mappings</h4>
          <div class="space-y-1">
            <div
              v-for="p in selectedService.ports"
              :key="p.port"
              class="p-2 rounded bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 flex items-center justify-between"
            >
              <span>Port {{ p.port }} &rarr; Target {{ p.target_port }} ({{ p.protocol }})</span>
              <span v-if="p.name" class="text-slate-400">{{ p.name }}</span>
            </div>
          </div>
        </div>

        <!-- Selector -->
        <div v-if="selectedService.selector && Object.keys(selectedService.selector).length > 0">
          <h4 class="text-xs font-bold uppercase text-slate-500 mb-2">Pod Selector</h4>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="(v, k) in selectedService.selector"
              :key="k"
              class="px-2 py-1 rounded bg-sky-50 dark:bg-sky-950/50 text-sky-600 dark:text-sky-400 border border-sky-200 dark:border-sky-800"
            >
              {{ k }}={{ v }}
            </span>
          </div>
        </div>
      </div>
    </Dialog>

    <!-- Resource YAML Dialog -->
    <ResourceYamlDialog
      v-model:visible="isYamlOpen"
      kind="Service"
      :name="selectedYamlName"
      :namespace="selectedNamespace"
      @applied="k8sStore.fetchServices()"
    />

    <!-- Service Endpoints Dialog -->
    <ServiceEndpointsDialog
      v-model:visible="isEndpointsOpen"
      :service-name="selectedEndpointsName"
      :namespace="selectedNamespace"
    />
  </div>
</template>

<style scoped>
</style>
