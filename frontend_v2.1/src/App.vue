<template>
  <div id="app" class="min-h-screen bg-gray-50">
    <router-view />
    
    <!-- Toast Notifications -->
    <Teleport to="body">
      <ToastNotifications />
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useChamberStore } from '@/stores/chamber'
import ToastNotifications from '@/components/ToastNotifications.vue'

const chamberStore = useChamberStore()

onMounted(() => {
  // Load saved chamber selection
  chamberStore.loadSelectedChamber()
})
</script>

<style>
@import './style.css';
</style> 