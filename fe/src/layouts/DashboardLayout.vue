<script setup lang="ts">
import { ref } from "vue";

import Sidebar from "@/components/layout/Sidebar.vue";
import TopNav from "@/components/layout/TopNav.vue";

const isSidebarOpen = ref(true);
</script>

<template>
  <div
    class="flex h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 overflow-hidden relative"
  >
    <!-- Mobile Backdrop -->
    <div 
      v-if="isSidebarOpen" 
      class="fixed inset-0 z-20 bg-slate-900/50 md:hidden backdrop-blur-sm"
      @click="isSidebarOpen = false"
    ></div>

    <Sidebar :is-sidebar-open="isSidebarOpen" @close="isSidebarOpen = false" />

    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <TopNav
        :is-sidebar-open="isSidebarOpen"
        @toggle-sidebar="isSidebarOpen = !isSidebarOpen"
      />

      <main class="flex-1 overflow-auto p-6">
        <div class="w-full h-full">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>
