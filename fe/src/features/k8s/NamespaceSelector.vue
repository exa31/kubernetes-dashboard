<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Select from 'primevue/select'
import { onMounted, ref } from 'vue'

import CreateNamespaceDialog from './CreateNamespaceDialog.vue'
import { useK8sStore } from '@/stores'

const k8sStore = useK8sStore()
const { namespaces, selectedNamespace, isLoading } = storeToRefs(k8sStore)
const showCreateDialog = ref(false)

onMounted(() => {
  k8sStore.fetchNamespaces()
})

const onSelectChange = (val: string) => {
  if (val) {
    k8sStore.setNamespace(val)
  }
}

const onNamespaceCreated = async (name: string) => {
  await k8sStore.fetchNamespaces()
  k8sStore.setNamespace(name)
}
</script>

<template>
  <div class="flex items-center gap-1.5 sm:gap-2 shrink-0">
    <div class="flex items-center gap-1 text-xs font-semibold text-slate-500 dark:text-slate-400 hidden lg:flex">
      <i class="pi pi-box text-sky-500 text-xs"></i>
      <span>Namespace:</span>
    </div>

    <Select
      :model-value="selectedNamespace"
      :options="namespaces"
      option-label="name"
      option-value="name"
      filter
      filter-placeholder="Search..."
      placeholder="Namespace"
      :loading="isLoading"
      class="text-xs font-mono font-semibold w-32 sm:w-40 md:w-48 h-9 flex items-center bg-slate-100 dark:bg-slate-800/80 border-slate-200 dark:border-slate-700/60"
      @update:model-value="onSelectChange"
    >
      <template #value="slotProps">
        <div v-if="slotProps.value" class="flex items-center gap-2 font-mono text-xs font-bold text-slate-800 dark:text-slate-100">
          <span class="w-2 h-2 rounded-full bg-sky-500"></span>
          <span>{{ slotProps.value }}</span>
        </div>
        <span v-else class="text-xs text-slate-400">Select Namespace</span>
      </template>

      <template #option="{ option }">
        <div class="flex items-center justify-between gap-3 font-mono text-xs w-full py-0.5">
          <span class="font-medium text-slate-800 dark:text-slate-200">{{ option.name }}</span>
          <span
            class="text-[10px] px-1.5 py-0.5 rounded font-bold"
            :class="option.status === 'Active' ? 'bg-emerald-50 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'"
          >
            {{ option.status }}
          </span>
        </div>
      </template>

      <template #footer>
        <div class="p-1.5 border-t border-slate-200 dark:border-slate-700/60 bg-slate-50 dark:bg-slate-900/90">
          <button
            type="button"
            class="w-full py-1.5 px-2 rounded-lg text-xs font-semibold bg-sky-500/10 hover:bg-sky-500/20 text-sky-600 dark:text-sky-400 flex items-center justify-center gap-1.5 transition cursor-pointer"
            @click="showCreateDialog = true"
          >
            <i class="pi pi-plus text-[10px]"></i>
            <span>Create Namespace</span>
          </button>
        </div>
      </template>
    </Select>

    <button
      type="button"
      class="w-8 h-8 rounded-lg flex items-center justify-center bg-slate-100 hover:bg-sky-500/10 dark:bg-slate-800 dark:hover:bg-sky-500/20 text-slate-500 hover:text-sky-500 dark:text-slate-400 dark:hover:text-sky-400 border border-slate-200 dark:border-slate-700/60 transition cursor-pointer shrink-0"
      title="Create New Namespace"
      @click="showCreateDialog = true"
    >
      <i class="pi pi-plus text-xs"></i>
    </button>

    <CreateNamespaceDialog
      v-model:visible="showCreateDialog"
      @created="onNamespaceCreated"
    />
  </div>
</template>

<style scoped>
</style>
