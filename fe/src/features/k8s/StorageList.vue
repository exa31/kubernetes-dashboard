<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { computed, onMounted, ref } from 'vue'

import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'
import { useK8sStore } from '@/stores'

const k8sStore = useK8sStore()
const activeTab = ref<'pvc' | 'pv'>('pvc')
const searchQuery = ref('')
const isYamlOpen = ref(false)
const selectedYaml = ref({ kind: 'PersistentVolumeClaim', name: '', namespace: '' })

function openYaml(kind: string, name: string, namespace = '') {
  selectedYaml.value = { kind, name, namespace }
  isYamlOpen.value = true
}

onMounted(() => {
  k8sStore.fetchPVCs()
  k8sStore.fetchPVs()
})

const filteredPVCs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return k8sStore.pvcs
  return k8sStore.pvcs.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.volume.toLowerCase().includes(q) ||
      p.storage_class.toLowerCase().includes(q),
  )
})

const filteredPVs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return k8sStore.pvs
  return k8sStore.pvs.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.claim.toLowerCase().includes(q) ||
      p.storage_class.toLowerCase().includes(q),
  )
})

function refresh() {
  k8sStore.fetchPVCs()
  k8sStore.fetchPVs()
}
</script>

<template>
  <div class="space-y-6 w-full">
    <!-- Header Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
          Storage & Volumes
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Monitor Persistent Volumes (PV) and Claims (PVC) across the cluster
        </p>
      </div>

      <!-- Tab Buttons & Refresh -->
      <div class="flex items-center gap-2">
        <div class="flex items-center bg-slate-100 dark:bg-slate-800/80 p-1 rounded-xl border border-slate-200 dark:border-slate-700/80">
          <button
            type="button"
            class="px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer"
            :class="
              activeTab === 'pvc'
                ? 'bg-teal-600 text-white font-semibold shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'
            "
            @click="activeTab = 'pvc'"
          >
            Volume Claims (PVC)
          </button>
          <button
            type="button"
            class="px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer"
            :class="
              activeTab === 'pv'
                ? 'bg-sky-600 text-white font-semibold shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'
            "
            @click="activeTab = 'pv'"
          >
            Cluster Volumes (PV)
          </button>
        </div>

        <Button
          icon="pi pi-refresh"
          size="small"
          class="btn-sky text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
          :loading="k8sStore.isLoading"
          title="Refresh storage"
          @click="refresh"
        />
      </div>
    </div>

    <!-- Search Toolbar -->
    <div class="flex items-center justify-between gap-4">
      <IconField class="w-full sm:w-80">
        <InputIcon class="pi pi-search" />
        <InputText
          v-model="searchQuery"
          :placeholder="activeTab === 'pvc' ? 'Search PVCs by name, volume...' : 'Search PVs by name, claim...'"
          class="w-full !rounded-xl text-sm"
        />
      </IconField>
    </div>

    <!-- PVC Table -->
    <div
      v-if="activeTab === 'pvc'"
      class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs overflow-hidden w-full"
    >
      <DataTable
        :value="filteredPVCs"
        :loading="k8sStore.isLoading"
        paginator
        :rows="10"
        class="w-full text-sm"
        table-style="min-width: 100%"
        striped-rows
      >
        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-database text-4xl mb-3 text-slate-300 dark:text-slate-600 block"></i>
            <p class="font-medium text-slate-600 dark:text-slate-300">No PersistentVolumeClaims found</p>
            <p class="text-xs text-slate-400 mt-1">There are no PVCs in namespace '{{ k8sStore.selectedNamespace }}'</p>
          </div>
        </template>

        <Column field="name" header="Claim Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-2.5 py-1">
              <div class="w-8 h-8 rounded-lg bg-teal-500/10 text-teal-600 dark:text-teal-400 flex items-center justify-center shrink-0">
                <i class="pi pi-database text-sm"></i>
              </div>
              <div>
                <div class="font-bold text-slate-900 dark:text-slate-100 font-mono text-xs">
                  {{ data.name }}
                </div>
                <div class="text-[11px] text-slate-400 font-mono">
                  Volume: {{ data.volume || 'Unbound' }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <Column field="status" header="Status" sortable style="width: 120px">
          <template #body="{ data }">
            <Tag
              :value="data.status"
              :severity="data.status === 'Bound' ? 'success' : data.status === 'Pending' ? 'warn' : 'danger'"
              class="font-mono text-xs font-semibold px-2 py-0.5"
            />
          </template>
        </Column>

        <Column field="capacity" header="Capacity" sortable style="width: 120px">
          <template #body="{ data }">
            <span class="font-mono text-xs font-bold text-slate-800 dark:text-slate-200">
              {{ data.capacity || 'N/A' }}
            </span>
          </template>
        </Column>

        <Column field="storage_class" header="Storage Class" sortable>
          <template #body="{ data }">
            <span class="font-mono text-xs text-slate-600 dark:text-slate-300">
              {{ data.storage_class || 'default' }}
            </span>
          </template>
        </Column>

        <Column field="access_modes" header="Access Modes">
          <template #body="{ data }">
            <div class="flex flex-wrap gap-1">
              <span
                v-for="mode in data.access_modes"
                :key="mode"
                class="text-[10px] px-2 py-0.5 rounded font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700"
              >
                {{ mode }}
              </span>
            </div>
          </template>
        </Column>

        <Column field="age" header="Age" sortable style="width: 100px">
          <template #body="{ data }">
            <span class="text-xs text-slate-500 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <Column header="Action" header-style="text-align: right" body-style="text-align: right" style="width: 100px">
          <template #body="{ data }">
            <Button
              label="YAML"
              icon="pi pi-code"
              size="small"
              class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
              @click.stop="openYaml('PersistentVolumeClaim', data.name, data.namespace)"
            />
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- PV Table -->
    <div
      v-else
      class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs overflow-hidden w-full"
    >
      <DataTable
        :value="filteredPVs"
        :loading="k8sStore.isLoading"
        paginator
        :rows="10"
        class="w-full text-sm"
        table-style="min-width: 100%"
        striped-rows
      >
        <template #empty>
          <div class="py-12 text-center text-slate-400">
            <i class="pi pi-server text-4xl mb-3 text-slate-300 dark:text-slate-600 block"></i>
            <p class="font-medium text-slate-600 dark:text-slate-300">No PersistentVolumes found</p>
            <p class="text-xs text-slate-400 mt-1">There are no cluster-level PVs provisioned</p>
          </div>
        </template>

        <Column field="name" header="Volume Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-2.5 py-1">
              <div class="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400 flex items-center justify-center shrink-0">
                <i class="pi pi-hdd text-sm"></i>
              </div>
              <div class="font-bold text-slate-900 dark:text-slate-100 font-mono text-xs">
                {{ data.name }}
              </div>
            </div>
          </template>
        </Column>

        <Column field="status" header="Status" sortable style="width: 120px">
          <template #body="{ data }">
            <Tag
              :value="data.status"
              :severity="data.status === 'Bound' ? 'success' : data.status === 'Available' ? 'info' : 'warn'"
              class="font-mono text-xs font-semibold px-2 py-0.5"
            />
          </template>
        </Column>

        <Column field="capacity" header="Capacity" sortable style="width: 120px">
          <template #body="{ data }">
            <span class="font-mono text-xs font-bold text-slate-800 dark:text-slate-200">
              {{ data.capacity }}
            </span>
          </template>
        </Column>

        <Column field="claim" header="Claim">
          <template #body="{ data }">
            <span class="font-mono text-xs text-sky-600 dark:text-sky-400 font-medium">
              {{ data.claim || 'None' }}
            </span>
          </template>
        </Column>

        <Column field="storage_class" header="Storage Class" sortable>
          <template #body="{ data }">
            <span class="font-mono text-xs text-slate-600 dark:text-slate-300">
              {{ data.storage_class || 'default' }}
            </span>
          </template>
        </Column>

        <Column field="reclaim_policy" header="Reclaim Policy" sortable style="width: 140px">
          <template #body="{ data }">
            <span class="text-xs text-slate-500 font-mono">{{ data.reclaim_policy }}</span>
          </template>
        </Column>

        <Column field="age" header="Age" sortable style="width: 100px">
          <template #body="{ data }">
            <span class="text-xs text-slate-500 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <Column header="Action" header-style="text-align: right" body-style="text-align: right" style="width: 100px">
          <template #body="{ data }">
            <Button
              label="YAML"
              icon="pi pi-code"
              size="small"
              class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
              @click.stop="openYaml('PersistentVolume', data.name)"
            />
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Resource YAML Dialog -->
    <ResourceYamlDialog
      v-model:visible="isYamlOpen"
      :kind="selectedYaml.kind"
      :name="selectedYaml.name"
      :namespace="selectedYaml.namespace"
      @applied="refresh"
    />
  </div>
</template>
