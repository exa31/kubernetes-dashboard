<script setup lang="ts">
import { ref, watch, nextTick, onBeforeUnmount } from 'vue'
import Dialog from 'primevue/dialog'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

interface Props {
  visible: boolean
  podName: string
  namespace: string
  containers?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  containers: () => [],
})

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const terminalContainer = ref<HTMLDivElement | null>(null)
const selectedContainer = ref<string>('')
const selectedShell = ref<string>('sh')
const connectionStatus = ref<'connecting' | 'connected' | 'disconnected'>('disconnected')
const isFullscreen = ref(false)

let term: Terminal | null = null
let fitAddon: FitAddon | null = null
let socket: WebSocket | null = null
let resizeObserver: ResizeObserver | null = null

// Initialize container selection
watch(
  () => props.containers,
  (newContainers) => {
    if (newContainers && newContainers.length > 0) {
      selectedContainer.value = newContainers[0]
    }
  },
  { immediate: true },
)

function connectTerminal() {
  if (!props.podName) return

  disconnectTerminal()
  connectionStatus.value = 'connecting'

  if (term) {
    term.reset()
  }

  // Derive WebSocket URL
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('access_token') || ''
  const containerParam = encodeURIComponent(selectedContainer.value || '')
  const shellParam = encodeURIComponent(selectedShell.value || 'sh')
  const wsUrl = `${protocol}//${window.location.host}/api/v1/k8s/ws/exec/${encodeURIComponent(props.namespace)}/${encodeURIComponent(props.podName)}?container=${containerParam}&shell=${shellParam}&token=${encodeURIComponent(token)}`

  socket = new WebSocket(wsUrl)
  socket.binaryType = 'arraybuffer'

  socket.onopen = () => {
    connectionStatus.value = 'connected'
    sendTerminalResize()
  }

  socket.onmessage = (event) => {
    if (term) {
      if (typeof event.data === 'string') {
        term.write(event.data)
      } else if (event.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(event.data))
      }
    }
  }

  socket.onerror = () => {
    connectionStatus.value = 'disconnected'
    if (term) {
      term.writeln('\r\n\x1b[31m[WebSocket Connection Error]\x1b[0m\r\n')
    }
  }

  socket.onclose = () => {
    connectionStatus.value = 'disconnected'
    if (term) {
      term.writeln('\r\n\x1b[33m[Connection closed]\x1b[0m\r\n')
    }
  }
}

function sendTerminalResize() {
  if (!term || !socket || socket.readyState !== WebSocket.OPEN) return
  const msg = JSON.stringify({
    type: 'resize',
    cols: term.cols,
    rows: term.rows,
  })
  socket.send(msg)
}

function initTerminal() {
  if (!terminalContainer.value) return

  if (term) {
    term.dispose()
    term = null
  }

  term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'bar',
    fontSize: 13,
    fontFamily: '"Fira Code", "JetBrains Mono", Consolas, "Liberation Mono", Menlo, Courier, monospace',
    theme: {
      background: '#090d16',
      foreground: '#e2e8f0',
      cursor: '#10b981',
      selectionBackground: '#1e293b',
      black: '#0f172a',
      red: '#ef4444',
      green: '#10b981',
      yellow: '#f59e0b',
      blue: '#3b82f6',
      magenta: '#ec4899',
      cyan: '#06b6d4',
      white: '#f8fafc',
    },
    convertEol: true,
    scrollback: 2000,
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)

  term.open(terminalContainer.value)
  fitAddon.fit()

  term.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(data)
    }
  })

  term.onResize(() => {
    sendTerminalResize()
  })

  resizeObserver = new ResizeObserver(() => {
    if (fitAddon && term) {
      try {
        fitAddon.fit()
        sendTerminalResize()
      } catch {}
    }
  })
  resizeObserver.observe(terminalContainer.value)

  connectTerminal()
}

function disconnectTerminal() {
  if (socket) {
    socket.close()
    socket = null
  }
}

function handleClose() {
  disconnectTerminal()
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  emit('update:visible', false)
}

function clearTerminal() {
  if (term) {
    term.clear()
  }
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    if (fitAddon) {
      fitAddon.fit()
      sendTerminalResize()
    }
  })
}

function changeShell(shell: string) {
  selectedShell.value = shell
  connectTerminal()
}

function changeContainer(c: string) {
  selectedContainer.value = c
  connectTerminal()
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      nextTick(() => {
        setTimeout(() => {
          initTerminal()
        }, 150)
      })
    } else {
      handleClose()
    }
  },
)

onBeforeUnmount(() => {
  handleClose()
  if (term) {
    term.dispose()
    term = null
  }
})
</script>

<template>
  <Dialog
    :visible="props.visible"
    modal
    :show-header="false"
    :pt="{
      root: {
        class: isFullscreen
          ? '!fixed !inset-0 !w-screen !h-screen !max-w-none !max-h-none !m-0 !rounded-none border-none p-0 bg-slate-950 overflow-hidden z-50'
          : 'border-none p-0 overflow-hidden w-[92vw] max-w-5xl rounded-2xl shadow-2xl bg-slate-950 z-50',
      },
      content: { class: 'p-0 overflow-hidden bg-slate-950 flex flex-col h-full' },
    }"
    @update:visible="handleClose"
  >
    <!-- Custom Terminal Header Bar -->
    <div
      class="flex items-center justify-between px-5 py-3.5 bg-slate-900/90 border-b border-slate-800 backdrop-blur shrink-0"
    >
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
          <i class="pi pi-terminal text-sm"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="font-bold text-white text-sm tracking-wide">Interactive Terminal</span>
            <span
              class="px-2 py-0.5 rounded text-[11px] font-medium border"
              :class="{
                'bg-emerald-500/10 text-emerald-400 border-emerald-500/30': connectionStatus === 'connected',
                'bg-amber-500/10 text-amber-400 border-amber-500/30': connectionStatus === 'connecting',
                'bg-rose-500/10 text-rose-400 border-rose-500/30': connectionStatus === 'disconnected',
              }"
            >
              <span class="inline-block w-1.5 h-1.5 rounded-full mr-1" :class="{
                'bg-emerald-400 animate-pulse': connectionStatus === 'connected',
                'bg-amber-400 animate-pulse': connectionStatus === 'connecting',
                'bg-rose-400': connectionStatus === 'disconnected',
              }"></span>
              {{ connectionStatus.toUpperCase() }}
            </span>
          </div>
          <div class="text-xs text-slate-400 flex items-center gap-2 mt-0.5 font-mono">
            <span>pod: {{ props.podName }}</span>
            <span>&bull;</span>
            <span>ns: {{ props.namespace }}</span>
          </div>
        </div>
      </div>

      <!-- Controls & Quick Selectors -->
      <div class="flex items-center gap-2">
        <!-- Container selector (if multiple) -->
        <div v-if="props.containers.length > 1" class="flex items-center gap-1.5 bg-slate-800/80 px-2.5 py-1 rounded-lg border border-slate-700">
          <span class="text-xs text-slate-400">Container:</span>
          <select
            :value="selectedContainer"
            class="bg-transparent text-xs text-slate-200 border-none outline-none cursor-pointer font-mono"
            @change="changeContainer(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="c in props.containers" :key="c" :value="c" class="bg-slate-900 text-white">
              {{ c }}
            </option>
          </select>
        </div>

        <!-- Shell Selector -->
        <div class="flex items-center gap-1.5 bg-slate-800/80 px-2.5 py-1 rounded-lg border border-slate-700">
          <span class="text-xs text-slate-400">Shell:</span>
          <select
            :value="selectedShell"
            class="bg-transparent text-xs text-slate-200 border-none outline-none cursor-pointer font-mono"
            @change="changeShell(($event.target as HTMLSelectElement).value)"
          >
            <option value="sh" class="bg-slate-900 text-white">/bin/sh</option>
            <option value="bash" class="bg-slate-900 text-white">/bin/bash</option>
          </select>
        </div>

        <!-- Reconnect -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition"
          title="Reconnect"
          @click="connectTerminal"
        >
          <i class="pi pi-refresh text-xs"></i>
        </button>

        <!-- Clear -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition"
          title="Clear Terminal"
          @click="clearTerminal"
        >
          <i class="pi pi-trash text-xs"></i>
        </button>

        <!-- Fullscreen Toggle -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition"
          :title="isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'"
          @click="toggleFullscreen"
        >
          <i :class="isFullscreen ? 'pi pi-window-minimize' : 'pi pi-window-maximize'" class="text-xs"></i>
        </button>

        <!-- Close -->
        <button
          type="button"
          class="p-2 rounded-lg bg-slate-800 hover:bg-rose-900/40 text-slate-400 hover:text-rose-400 border border-slate-700 hover:border-rose-700/50 transition ml-1"
          title="Close"
          @click="handleClose"
        >
          <i class="pi pi-times text-xs"></i>
        </button>
      </div>
    </div>

    <!-- Terminal Screen Body -->
    <div
      class="w-full bg-[#090d16] p-3 flex-1 overflow-hidden"
      :style="{ height: isFullscreen ? 'calc(100vh - 58px)' : '480px' }"
    >
      <div ref="terminalContainer" class="w-full h-full"></div>
    </div>

    <!-- Terminal Footer Status Info -->
    <div
      class="px-5 py-2 bg-slate-900/90 border-t border-slate-800 flex items-center justify-between text-xs text-slate-500 font-mono shrink-0"
    >
      <div class="flex items-center gap-4">
        <span>Press <kbd class="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700 text-[10px]">Ctrl+C</kbd> to interrupt</span>
        <span>Type <kbd class="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700 text-[10px]">exit</kbd> to close shell</span>
      </div>
      <div class="flex items-center gap-3">
        <span>Xterm.js Web Terminal</span>
        <span>&bull;</span>
        <span class="text-emerald-400">Kubernetes Remote Exec</span>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
:deep(.xterm) {
  padding: 4px;
}
:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
