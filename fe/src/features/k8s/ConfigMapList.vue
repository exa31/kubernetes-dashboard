<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import { computed, onMounted, ref } from 'vue'

import EnvEditor from '@/features/k8s/EnvEditor.vue'
import { useK8sStore } from '@/stores'
import type { ConfigMapDetail, ConfigMapItem } from '@/types'

const k8sStore = useK8sStore()
const { configmaps, selectedNamespace, isLoading, isActionLoading } = storeToRefs(k8sStore)

const searchQuery = ref('')
const activeConfigMapDetail = ref<ConfigMapDetail | null>(null)
const isEditorOpen = ref(false)

// Create Dialog
const isCreateOpen = ref(false)
const newCMName = ref('')
const newCMEnvText = ref('CONFIG_KEY=value\nLOG_LEVEL=info')

onMounted(() => {
  k8sStore.fetchConfigMaps()
})

const filteredConfigMaps = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return configmaps.value
  return configmaps.value.filter((c) => c.name.toLowerCase().includes(q))
})

const openConfigMap = async (item: ConfigMapItem) => {
  try {
    const detail = await k8sStore.getConfigMapDetail(item.name)
    activeConfigMapDetail.value = detail
    isEditorOpen.value = true
  } catch {
    // handled in store
  }
}

const deleteConfigMap = async (item: ConfigMapItem) => {
  if (confirm(`Are you sure you want to delete configmap '${item.name}' from namespace '${selectedNamespace.value}'?`)) {
    await k8sStore.deleteConfigMap(item.name)
  }
}

const createConfigMap = async () => {
  if (!newCMName.value.trim()) return

  const lines = newCMEnvText.value.split('\n')
  const data: Record<string, string> = {}
  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const eqIdx = trimmed.indexOf('=')
    if (eqIdx > 0) {
      const k = trimmed.slice(0, eqIdx).trim()
      let v = trimmed.slice(eqIdx + 1).trim()
      if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
        v = v.slice(1, -1)
      }
      data[k] = v
    }
  }

  await k8sStore.saveConfigMap({
    name: newCMName.value.trim(),
    namespace: selectedNamespace.value,
    data,
  })

  isCreateOpen.value = false
  newCMName.value = ''
}
</script>

<template>
  <div class="space-y-4">
    <!-- Top toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <i class="pi pi-file text-sky-500"></i>
          <span>ConfigMaps</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Manage application configurations and non-sensitive environment variables in <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ selectedNamespace }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText
            v-model="searchQuery"
            placeholder="Search configmaps..."
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
          @click="k8sStore.fetchConfigMaps()"
        />

        <Button
          label="Create ConfigMap"
          icon="pi pi-plus"
          size="small"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          @click="isCreateOpen = true"
        />
      </div>
    </div>

    <!-- PrimeVue DataTable for ConfigMaps -->
    <div class="w-full bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
      <DataTable
        :value="filteredConfigMaps"
        :loading="isLoading"
        striped-rows
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
        table-style="min-width: 100%"
        class="p-datatable-sm w-full"
        row-hover
        @row-click="(e) => openConfigMap(e.data)"
      >
        <!-- Name Column -->
        <Column field="name" header="Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1 cursor-pointer">
              <div class="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-500 flex items-center justify-center font-bold shrink-0">
                <i class="pi pi-file text-xs"></i>
              </div>
              <div>
                <div class="font-semibold text-slate-900 dark:text-slate-100 font-mono text-sm hover:text-sky-600 transition-colors">
                  {{ data.name }}
                </div>
                <div class="text-xs text-slate-400 mt-0.5 truncate font-mono">
                  {{ data.keys.slice(0, 8).join(', ') }}{{ data.keys.length > 8 ? ` +${data.keys.length - 8} more` : '' }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Variables / Key count -->
        <Column field="key_count" header="Variables" sortable style="width: 140px">
          <template #body="{ data }">
            <span class="font-semibold font-mono text-slate-800 dark:text-slate-200">
              {{ data.key_count }}
            </span>
            <span class="text-xs text-slate-400 ml-1">keys</span>
          </template>
        </Column>

        <!-- Age Column -->
        <Column field="age" header="Age" sortable style="width: 120px">
          <template #body="{ data }">
            <span class="text-xs text-slate-500 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <!-- Actions Column -->
        <Column header="Actions" header-style="text-align: right" body-style="text-align: right" style="width: 180px">
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5" @click.stop>
              <Button
                label="Edit Env"
                icon="pi pi-file-edit"
                size="small"
                class="btn-blue text-xs px-3 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="openConfigMap(data)"
              />
              <Button
                icon="pi pi-trash"
                size="small"
                class="btn-rose text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :disabled="isActionLoading"
                title="Delete configmap"
                @click="deleteConfigMap(data)"
              />
            </div>
          </template>
        </Column>

        <!-- Empty state -->
        <template #empty>
          <div class="py-16 text-center text-slate-400">
            <i class="pi pi-file text-4xl mb-3 text-slate-300 dark:text-slate-700"></i>
            <h3 class="font-semibold text-slate-700 dark:text-slate-300">No ConfigMaps Found</h3>
            <p class="text-xs text-slate-500 mt-1">There are no ConfigMaps in namespace {{ selectedNamespace }}.</p>
            <Button
              label="Create ConfigMap"
              icon="pi pi-plus"
              size="small"
              class="mt-4 btn-emerald text-xs shadow-xs cursor-pointer"
              @click="isCreateOpen = true"
            />
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Env Editor Dialog Modal -->
    <Dialog
      v-model:visible="isEditorOpen"
      modal
      :show-header="false"
      class="w-[95vw] max-w-6xl h-[85vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
      :pt="{
        root: { class: 'border-none p-0 overflow-hidden' },
        content: { class: 'p-0 h-full overflow-hidden' }
      }"
    >
      <EnvEditor
        v-if="activeConfigMapDetail"
        resource-type="configmap"
        :detail="activeConfigMapDetail"
        @close="isEditorOpen = false"
        @saved="() => k8sStore.fetchConfigMaps()"
      />
    </Dialog>

    <!-- Create ConfigMap Dialog -->
    <Dialog
      v-model:visible="isCreateOpen"
      modal
      :show-header="false"
      class="w-[90vw] max-w-xl rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
      :pt="{
        root: { class: 'border-none p-0 overflow-hidden' },
        content: { class: 'p-0 overflow-hidden bg-white dark:bg-slate-900' }
      }"
    >
      <!-- Custom Header -->
      <div class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-sky-500/10 text-sky-500 flex items-center justify-center font-bold text-base shrink-0">
            <i class="pi pi-file-edit"></i>
          </div>
          <div>
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100">
              Create New ConfigMap
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Store non-sensitive application environment variables and configuration
            </p>
          </div>
        </div>
        <button
          type="button"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          @click="isCreateOpen = false"
        >
          <i class="pi pi-times"></i>
        </button>
      </div>

      <form class="p-6 space-y-4" @submit.prevent="createConfigMap">
        <div>
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
            ConfigMap Name *
          </label>
          <InputText
            v-model="newCMName"
            placeholder="e.g. app-config"
            class="w-full font-mono text-sm"
            required
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
            Variables (.env format)
          </label>
          <textarea
            v-model="newCMEnvText"
            rows="6"
            placeholder="KEY=value&#10;LOG_LEVEL=debug"
            class="w-full p-3 font-mono text-xs bg-slate-950 text-emerald-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
          ></textarea>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <Button label="Cancel" severity="secondary" text @click="isCreateOpen = false" />
          <Button
            type="submit"
            label="Create ConfigMap"
            icon="pi pi-check"
            class="btn-emerald text-xs shadow-xs cursor-pointer"
            :loading="isActionLoading"
          />
        </div>
      </form>
    </Dialog>
  </div>
</template>

<style scoped>
</style>
