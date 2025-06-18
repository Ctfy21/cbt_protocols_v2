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
import { onMounted, onUnmounted } from 'vue'
import { useChamberStore } from '@/stores/chamber'
import { useAuthStore } from '@/stores/auth'
import ToastNotifications from '@/components/ToastNotifications.vue'
import experimentTracker from '@/services/experimentTracker'

const chamberStore = useChamberStore()
const authStore = useAuthStore()

onMounted(() => {
  // Load saved chamber selection
  chamberStore.loadSelectedChamber()
  
  // Запускаем отслеживание экспериментов если пользователь авторизован
  if (authStore.isAuthenticated) {
    experimentTracker.startTracking()
  }
})

onUnmounted(() => {
  // Останавливаем отслеживание при размонтировании
  experimentTracker.stopTracking()
})
</script>

<style>
@import './style.css';
</style> 