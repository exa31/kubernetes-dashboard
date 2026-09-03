<script setup lang="ts">
import Badge from 'primevue/badge'
import Button from 'primevue/button'
import Drawer from 'primevue/drawer'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const k8sStore = useK8sStore()
const isLoading = ref(false)
const searchQuery = ref('')
const typeFilter = ref<'all' | 'Warning' | 'Normal'>('all')
const scopeFilter = ref<'current' | 'all'>('current')
const autoRefresh = ref(true)

let timer: ReturnType<typeof setInterval> | null = null

async function loadEvents() {
  isLoading.value = true
  try {
    const ns = scopeFilter.value === 'current' ? k8sStore.selectedNamespace : ''
    await k8sStore.fetchEventsFeed(ns || undefined)
  } finally {
    isLoading.value = false
  }
}

function startPolling() {
  stopPolling()
  if (autoRefresh.value && props.visible) {
    timer = setInterval(() => {
      const ns = scopeFilter.value === 'current' ? k8sStore.selectedNamespace : ''
      k8sStore.fetchEventsFeed(ns || undefined)
    }, 5000)
  }
}

function stopPolling() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      loadEvents()
      startPolling()
    } else {
      stopPolling()
    }
  },
)

watch(scopeFilter, () => {
  loadEvents()
})

watch(autoRefresh, (val) => {
  if (val) startPolling()
  else stopPolling()
})

onMounted(() => {
  if (props.visible) {
    loadEvents()
    startPolling()
  }
})

onBeforeUnmount(() => {
  stopPolling()
})

// Use eventsFeed if available, fallback to store events
const allEvents = computed(() => {
  return k8sStore.eventsFeed && k8sStore.eventsFeed.length > 0
    ? k8sStore.eventsFeed
    : k8sStore.events
})

const warningCount = computed(() => allEvents.value.filter((e) => e.type === 'Warning').length)
const normalCount = computed(() => allEvents.value.filter((e) => e.type === 'Normal').length)

const filteredEvents = computed(() => {
  let list = allEvents.value

  if (typeFilter.value !== 'all') {
    list = list.filter((e) => e.type === typeFilter.value)
  }

  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (e) =>
        e.involved_object.toLowerCase().includes(q) ||
        e.reason.toLowerCase().includes(q) ||
        e.message.toLowerCase().includes(q) ||
        (e.namespace && e.namespace.toLowerCase().includes(q)),
    )
  }

  return list
})
</script>

<template>
  <Drawer
    :visible="visible"
    modal
    position="right"
    class="!w-full sm:!w-[520px] md:!w-[560px] max-w-[95vw] bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 border-l border-slate-200 dark:border-slate-800 p-0 shadow-2xl"
    header="Cluster Events & Alerts Feed"
    @update:visible="(val) => emit('update:visible', val)"
  >
    <template #header>
      <div class="flex items-center justify-between w-full pr-4 min-w-0">
        <div class="flex items-center gap-2.5 min-w-0 flex-1 mr-2">
          <div class="w-9 h-9 rounded-xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-500 shrink-0">
            <i class="pi pi-bell text-base"></i>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 truncate">Live Cluster Events</h3>
              <span
                v-if="autoRefresh"
                class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shrink-0"
                title="Live 5s streaming active"
              ></span>
            </div>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate">
              {{ scopeFilter === 'current' ? `Namespace: ${k8sStore.selectedNamespace}` : 'All Cluster Namespaces' }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <!-- Auto-refresh switch -->
          <button
            type="button"
            class="px-2 py-1 rounded text-[11px] font-mono border transition flex items-center gap-1 cursor-pointer shrink-0"
            :class="autoRefresh ? 'bg-emerald-500/10 text-emerald-500 dark:text-emerald-400 border-emerald-500/30' : 'bg-slate-100 dark:bg-slate-800 text-slate-400 border-slate-300 dark:border-slate-700'"
            title="Toggle 5s live auto-refresh"
            @click="autoRefresh = !autoRefresh"
          >
            <i class="pi pi-bolt text-[10px]"></i>
            <span>{{ autoRefresh ? 'Live' : 'Paused' }}</span>
          </button>

          <Button
            icon="pi pi-refresh"
            size="small"
            class="btn-amber text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer font-bold shrink-0"
            :loading="isLoading"
            title="Refresh events now"
            @click="loadEvents"
          />
          <Badge v-if="warningCount > 0" :value="warningCount" severity="warn" class="shrink-0" />
        </div>
      </div>
    </template>

    <!-- Filter Bar & Search -->
    <div class="p-4 border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-slate-900/60 space-y-3 shrink-0">
      <!-- Search -->
      <div class="relative">
        <InputText
          v-model="searchQuery"
          placeholder="Filter events by object, reason, or message..."
          class="w-full text-xs rounded-xl bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 py-1.5 pr-8 font-mono"
        />
        <button
          v-if="searchQuery"
          type="button"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white text-xs cursor-pointer"
          @click="searchQuery = ''"
        >
          &times;
        </button>
      </div>

      <!-- Filter pills -->
      <div class="flex items-center justify-between flex-wrap gap-2 text-xs">
        <div class="flex items-center gap-1.5 flex-wrap">
          <!-- All -->
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="typeFilter === 'all' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-400 hover:text-white'"
            @click="typeFilter = 'all'"
          >
            All ({{ allEvents.length }})
          </button>

          <!-- Warnings -->
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition flex items-center gap-1 cursor-pointer"
            :class="typeFilter === 'Warning' ? 'bg-amber-500/20 text-amber-300 font-semibold border border-amber-500/40' : 'text-slate-400 hover:text-amber-400'"
            @click="typeFilter = 'Warning'"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>Warnings ({{ warningCount }})</span>
          </button>

          <!-- Normal -->
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition flex items-center gap-1 cursor-pointer"
            :class="typeFilter === 'Normal' ? 'bg-sky-500/20 text-sky-300 font-semibold border border-sky-500/40' : 'text-slate-400 hover:text-sky-400'"
            @click="typeFilter = 'Normal'"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-sky-400"></span>
            <span>Normal ({{ normalCount }})</span>
          </button>
        </div>

        <!-- Scope switch -->
        <div class="flex items-center bg-slate-200 dark:bg-slate-800 p-0.5 rounded-lg border border-slate-300 dark:border-slate-700 shrink-0">
          <button
            type="button"
            class="px-2 py-0.5 rounded text-[11px] transition font-medium cursor-pointer"
            :class="scopeFilter === 'current' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white font-semibold shadow-xs' : 'text-slate-500 dark:text-slate-400 hover:text-white'"
            @click="scopeFilter = 'current'"
          >
            Current NS
          </button>
          <button
            type="button"
            class="px-2 py-0.5 rounded text-[11px] transition font-medium cursor-pointer"
            :class="scopeFilter === 'all' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-white font-semibold shadow-xs' : 'text-slate-500 dark:text-slate-400 hover:text-white'"
            @click="scopeFilter = 'all'"
          >
            Cluster-Wide
          </button>
        </div>
      </div>
    </div>

    <!-- Events List -->
    <div class="p-4 space-y-3 overflow-y-auto max-h-[calc(100vh-170px)]">
      <div v-if="isLoading && filteredEvents.length === 0" class="flex flex-col items-center justify-center py-12 text-slate-400">
        <i class="pi pi-spin pi-spinner text-2xl mb-2 text-amber-500"></i>
        <p class="text-sm">Loading cluster events feed...</p>
      </div>

      <div v-else-if="filteredEvents.length === 0" class="text-center py-12 text-slate-400">
        <i class="pi pi-check-circle text-4xl mb-3 text-emerald-500"></i>
        <p class="text-base font-semibold text-slate-700 dark:text-slate-200">No matching events</p>
        <p class="text-xs text-slate-500 mt-1">
          {{ searchQuery ? 'No events matching your filter search.' : 'Everything in this scope is running stably.' }}
        </p>
      </div>

      <div
        v-for="(evt, idx) in filteredEvents"
        :key="idx"
        class="p-3.5 rounded-xl border transition-all min-w-0"
        :class="
          evt.type === 'Warning'
            ? 'bg-amber-500/5 dark:bg-amber-500/10 border-amber-500/30 shadow-xs'
            : 'bg-slate-50 dark:bg-slate-800/40 border-slate-200 dark:border-slate-800'
        "
      >
        <div class="flex items-start justify-between gap-2 mb-1.5 min-w-0">
          <div class="flex items-center gap-1.5 flex-wrap min-w-0 flex-1">
            <Tag
              :value="evt.reason"
              :severity="evt.type === 'Warning' ? 'warn' : 'info'"
              class="font-mono text-[10px] px-2 py-0.5 shrink-0"
            />
            <span class="text-xs font-semibold font-mono text-slate-800 dark:text-slate-200 break-all min-w-0">
              {{ evt.involved_object }}
            </span>
            <span
              v-if="evt.namespace"
              class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 shrink-0"
            >
              {{ evt.namespace }}
            </span>
          </div>
          <span class="text-[11px] text-slate-400 font-mono shrink-0 whitespace-nowrap pl-1">{{ evt.age }}</span>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed font-mono break-all [overflow-wrap:anywhere] min-w-0">
          {{ evt.message }}
        </p>

        <div v-if="evt.count > 1" class="mt-2 flex items-center justify-end">
          <span class="text-[10px] px-2 py-0.5 rounded bg-slate-200 dark:bg-slate-700/60 text-slate-600 dark:text-slate-400 font-mono font-semibold shrink-0">
            Seen {{ evt.count }} times
          </span>
        </div>
      </div>
    </div>
  </Drawer>
</template>
