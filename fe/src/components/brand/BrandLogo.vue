<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
    showText?: boolean
    collapsed?: boolean
    badge?: string
  }>(),
  {
    size: 'md',
    showText: true,
    collapsed: false,
    badge: 'v2.0',
  },
)

const iconSizeClass = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'w-6 h-6'
    case 'sm':
      return 'w-7 h-7'
    case 'md':
      return 'w-9 h-9'
    case 'lg':
      return 'w-11 h-11'
    case 'xl':
      return 'w-14 h-14'
    default:
      return 'w-9 h-9'
  }
})

const textTitleClass = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'text-xs'
    case 'sm':
      return 'text-sm'
    case 'md':
      return 'text-base font-extrabold'
    case 'lg':
      return 'text-xl font-black'
    case 'xl':
      return 'text-2xl font-black'
    default:
      return 'text-base font-extrabold'
  }
})
</script>

<template>
  <div class="flex items-center gap-2.5 select-none shrink-0 group">
    <!-- Brand SVG Icon -->
    <div
      class="relative flex items-center justify-center shrink-0 transition-transform duration-300 group-hover:scale-105"
      :class="iconSizeClass"
    >
      <svg
        viewBox="0 0 48 48"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="w-full h-full drop-shadow-md"
      >
        <defs>
          <!-- Background Gradient -->
          <linearGradient id="kn-bg-grad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#0284c7" />
            <stop offset="50%" stop-color="#2563eb" />
            <stop offset="100%" stop-color="#7c3aed" />
          </linearGradient>

          <!-- Outer Glow Gradient -->
          <linearGradient id="kn-glow-grad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#38bdf8" />
            <stop offset="100%" stop-color="#c084fc" />
          </linearGradient>

          <!-- Core Radial Gradient -->
          <radialGradient id="kn-core-grad" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#34d399" />
            <stop offset="100%" stop-color="#059669" />
          </radialGradient>
        </defs>

        <!-- Outer Hexagon Shield -->
        <path
          d="M24 3 L42 13.5 L42 34.5 L24 45 L6 34.5 L6 13.5 Z"
          fill="url(#kn-bg-grad)"
          stroke="url(#kn-glow-grad)"
          stroke-width="1.8"
          stroke-linejoin="round"
        />

        <!-- Inner Isometric Node Box -->
        <!-- Top Face -->
        <polygon points="24,11 35,17.5 24,24 13,17.5" fill="#ffffff" fill-opacity="0.25" />
        <!-- Left Face -->
        <polygon points="13,17.5 24,24 24,36.5 13,30" fill="#000000" fill-opacity="0.25" />
        <!-- Right Face -->
        <polygon points="24,24 35,17.5 35,30 24,36.5" fill="#ffffff" fill-opacity="0.12" />

        <!-- Satellite Cluster Nodes -->
        <circle cx="24" cy="11" r="2.2" fill="#38bdf8" stroke="#ffffff" stroke-width="1" />
        <circle cx="35" cy="17.5" r="2.2" fill="#818cf8" stroke="#ffffff" stroke-width="1" />
        <circle cx="35" cy="30" r="2.2" fill="#c084fc" stroke="#ffffff" stroke-width="1" />
        <circle cx="24" cy="36.5" r="2.2" fill="#34d399" stroke="#ffffff" stroke-width="1" />
        <circle cx="13" cy="30" r="2.2" fill="#38bdf8" stroke="#ffffff" stroke-width="1" />
        <circle cx="13" cy="17.5" r="2.2" fill="#818cf8" stroke="#ffffff" stroke-width="1" />

        <!-- Center Active Nexus Quantum Core -->
        <circle cx="24" cy="24" r="4.2" fill="url(#kn-core-grad)" stroke="#ffffff" stroke-width="1.5" />
        <circle cx="24" cy="24" r="1.6" fill="#ffffff" />
      </svg>
    </div>

    <!-- Brand Text -->
    <div v-if="showText && !collapsed" class="flex items-center gap-2 min-w-0">
      <div class="tracking-tight leading-none" :class="textTitleClass">
        <span class="text-slate-900 dark:text-white">Kube</span><span class="text-transparent bg-clip-text bg-gradient-to-r from-sky-400 via-cyan-300 to-emerald-400">Nexus</span>
      </div>

      <!-- Version / Edition Badge -->
      <span
        v-if="badge"
        class="text-[10px] px-1.5 py-0.5 rounded-md font-mono font-bold bg-sky-500/10 text-sky-600 dark:text-sky-400 border border-sky-500/20"
      >
        {{ badge }}
      </span>
    </div>
  </div>
</template>
