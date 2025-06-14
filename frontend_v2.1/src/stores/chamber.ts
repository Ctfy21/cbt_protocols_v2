import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Chamber } from '@/types'
import api from '@/services/api'
import { useAuthStore } from '@/stores/auth'

export const useChamberStore = defineStore('chamber', () => {
  // State
  const chambers = ref<Chamber[]>([])
  const selectedChamber = ref<Chamber | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const onlineChambers = computed(() => 
    chambers.value.filter(c => c.status === 'online')
  )

  const offlineChambers = computed(() => 
    chambers.value.filter(c => c.status === 'offline')
  )

  // Actions
  async function fetchChambers() {
    loading.value = true
    error.value = null
    try {
      const authStore = useAuthStore()
      
      // If user is admin, fetch all chambers
      if (authStore.isAdmin) {
        const response = await api.getChambers()
        if (response.success && response.data) {
          chambers.value = response.data
        }
      } else {
        // For regular users, fetch only chambers they have access to
        if (authStore.user?.id) {
          const response = await api.getUserChamberAccess(authStore.user.id)
          if (response.success && response.data) {
            chambers.value = response.data
          }
        } else {
          chambers.value = []
        }
      }
    } catch (err) {
      error.value = api.formatError(err)
    } finally {
      loading.value = false
    }
  }

  async function fetchChamber(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.getChamber(id)
      if (response.success && response.data) {
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
    } finally {
      loading.value = false
    }
  }

  async function registerChamber(data: Partial<Chamber>) {
    loading.value = true
    error.value = null
    try {
      const response = await api.registerChamber(data)
      if (response.success && response.data) {
        chambers.value.push(response.data)
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  function selectChamber(chamber: Chamber | null) {
    selectedChamber.value = chamber
    if (chamber) {
      localStorage.setItem('selected_chamber', JSON.stringify(chamber))
    } else {
      localStorage.removeItem('selected_chamber')
    }
  }

  function loadSelectedChamber() {
    const saved = localStorage.getItem('selected_chamber')
    if (saved) {
      try {
        selectedChamber.value = JSON.parse(saved)
      } catch (err) {
        console.error('Failed to parse saved chamber:', err)
        localStorage.removeItem('selected_chamber')
      }
    }
  }

  return {
    // State
    chambers,
    selectedChamber,
    loading,
    error,
    // Getters
    onlineChambers,
    offlineChambers,
    // Actions
    fetchChambers,
    fetchChamber,
    registerChamber,
    selectChamber,
    loadSelectedChamber
  }
}) 