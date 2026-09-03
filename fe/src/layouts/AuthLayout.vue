<script setup lang="ts">
import BrandLogo from '@/components/brand/BrandLogo.vue'
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 flex items-center justify-center p-4 relative overflow-hidden">
    <!-- Ambient glowing aura -->
    <div class="absolute w-[500px] h-[500px] rounded-full bg-gradient-to-tr from-sky-500/10 via-indigo-500/10 to-transparent blur-3xl pointer-events-none -top-32 -left-32"></div>
    <div class="absolute w-[400px] h-[400px] rounded-full bg-gradient-to-br from-purple-500/10 via-emerald-500/10 to-transparent blur-3xl pointer-events-none -bottom-32 -right-32"></div>

    <div class="w-full max-w-md relative z-10">
      <div class="flex flex-col items-center text-center mb-8">
        <BrandLogo size="lg" badge="ENTERPRISE" class="mb-3" />
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Cloud-Native Kubernetes Control Plane & Observability
        </p>
      </div>
      
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl shadow-sky-500/5 overflow-hidden">
        <router-view />
      </div>
    </div>
  </div>
</template>
