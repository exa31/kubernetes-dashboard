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

import ResourceYamlDialog from '@/features/k8s/ResourceYamlDialog.vue'
import { useK8sStore } from '@/stores'

const k8sStore = useK8sStore()
const { ingresses, selectedNamespace, isLoading } = storeToRefs(k8sStore)

const searchQuery = ref('')
const isYamlOpen = ref(false)
const selectedYamlName = ref('')

function openYaml(name: string) {
  selectedYamlName.value = name
  isYamlOpen.value = true
}

onMounted(() => {
  k8sStore.fetchIngresses()
})

const filteredIngresses = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return ingresses.value
  return ingresses.value.filter(
    (ing) =>
      ing.name.toLowerCase().includes(q) ||
      ing.class_name.toLowerCase().includes(q) ||
      ing.hosts.some((h) => h.toLowerCase().includes(q)) ||
      ing.rules.some((r) => r.service_name.toLowerCase().includes(q)),
  )
})
</script>

<template>
  <div class="space-y-4">
    <!-- Top toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-globe text-cyan-500"></i>
          <span>Ingresses & Routes</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Public hostnames, reverse proxy routing, and TLS certificates in <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ selectedNamespace }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText
            v-model="searchQuery"
            placeholder="Search ingresses & hosts..."
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
          @click="k8sStore.fetchIngresses()"
        />
      </div>
    </div>

    <!-- PrimeVue DataTable for Ingresses -->
    <div class="w-full rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden bg-white dark:bg-slate-950">
      <DataTable
        :value="filteredIngresses"
        :loading="isLoading"
        striped-rows
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
        table-style="min-width: 100%"
        class="p-datatable-sm w-full"
      >
        <!-- Name Column -->
        <Column field="name" header="Ingress Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1">
              <div class="w-8 h-8 rounded-lg bg-cyan-500/10 text-cyan-500 flex items-center justify-center font-bold shrink-0">
                <i class="pi pi-globe text-xs"></i>
              </div>
              <div>
                <div class="font-semibold text-slate-900 dark:text-slate-100 font-mono text-sm">
                  {{ data.name }}
                </div>
                <div class="flex items-center gap-1.5 text-xs text-slate-400 mt-0.5">
                  <span>Class: <strong class="text-slate-600 dark:text-slate-300 font-mono">{{ data.class_name || 'default' }}</strong></span>
                  <span>&bull;</span>
                  <span>Age: {{ data.age }}</span>
                </div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Hosts Column (Clickable!) -->
        <Column header="Hosts (Public URLs)">
          <template #body="{ data }">
            <div class="space-y-1">
              <div
                v-for="host in data.hosts"
                :key="host"
                class="flex items-center gap-1.5"
              >
                <a
                  :href="`https://${host}`"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="font-mono text-xs text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1"
                >
                  <span>{{ host }}</span>
                  <i class="pi pi-external-link text-[10px]"></i>
                </a>
              </div>
              <span v-if="data.hosts.length === 0" class="text-xs text-slate-400 italic">* (All hosts)</span>
            </div>
          </template>
        </Column>

        <!-- Address / LB Column -->
        <Column field="address" header="Address" style="width: 150px">
          <template #body="{ data }">
            <span v-if="data.address" class="font-mono text-xs text-slate-700 dark:text-slate-300">
              {{ data.address }}
            </span>
            <span v-else class="text-xs text-slate-400 italic">-</span>
          </template>
        </Column>

        <!-- Backend Routing Rules Column -->
        <Column header="Routing Paths">
          <template #body="{ data }">
            <div class="space-y-1 w-full">
              <div
                v-for="(r, idx) in data.rules"
                :key="idx"
                class="text-[11px] font-mono flex items-center gap-1 text-slate-700 dark:text-slate-300"
              >
                <span class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                  {{ r.path || '/' }}
                </span>
                <span class="text-slate-400">&rarr;</span>
                <span class="text-sky-600 dark:text-sky-400 font-semibold">{{ r.service_name }}:{{ r.service_port }}</span>
              </div>
              <span v-if="data.rules.length === 0" class="text-xs text-slate-400 italic">No explicit path rules</span>
            </div>
          </template>
        </Column>

        <!-- TLS Column -->
        <Column header="TLS / SSL" style="width: 140px">
          <template #body="{ data }">
            <Tag
              v-if="data.tls && data.tls.length > 0"
              value="TLS Enabled"
              icon="pi pi-lock"
              severity="success"
              class="font-mono text-xs"
            />
            <span v-else class="text-xs text-slate-400 italic">HTTP Only</span>
          </template>
        </Column>

        <!-- Actions Column -->
        <Column header="Actions" header-style="text-align: right" body-style="text-align: right" style="width: 100px">
          <template #body="{ data }">
            <Button
              label="YAML"
              icon="pi pi-code"
              size="small"
              class="btn-purple text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
              @click.stop="openYaml(data.name)"
            />
          </template>
        </Column>

        <!-- Empty state -->
        <template #empty>
          <div class="py-16 text-center text-slate-400">
            <i class="pi pi-globe text-4xl mb-3 text-slate-300 dark:text-slate-700"></i>
            <h3 class="font-semibold text-slate-700 dark:text-slate-300">No Ingresses Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no Ingress routes configured in namespace {{ selectedNamespace }}.</p>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Resource YAML Dialog -->
    <ResourceYamlDialog
      v-model:visible="isYamlOpen"
      kind="Ingress"
      :name="selectedYamlName"
      :namespace="selectedNamespace"
      @applied="k8sStore.fetchIngresses()"
    />
  </div>
</template>

<style scoped>
</style>
