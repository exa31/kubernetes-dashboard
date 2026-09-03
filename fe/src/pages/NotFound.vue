<script setup lang="ts">
import Button from 'primevue/button'
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center p-4 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-50 text-center">
    <div class="max-w-md w-full">
      <div class="text-sky-500 font-bold text-9xl tracking-tighter mb-4 opacity-50 dark:opacity-20">
        404
      </div>
      <h1 class="text-3xl font-bold tracking-tight mb-3">Page not found</h1>
      <p class="text-slate-500 dark:text-slate-400 mb-8">
        Sorry, we couldn't find the page you're looking for. It might have been moved or deleted.
      </p>
      <router-link to="/">
        <Button label="Back to Dashboard" icon="pi pi-home" class="bg-sky-600 hover:bg-sky-500 border-none px-6 py-2.5 rounded-xl font-semibold text-white shadow-xs cursor-pointer" />
      </router-link>
    </div>
  </div>
</template>
