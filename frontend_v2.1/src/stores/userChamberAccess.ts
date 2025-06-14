import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types/auth'
import type { Chamber } from '@/types'
import api from '@/services/api'

export interface UserWithChamberAccess {
  user: User
  chambers: Chamber[]
}

export const useUserChamberAccessStore = defineStore('userChamberAccess', () => {
  // State
  const usersWithAccess = ref<UserWithChamberAccess[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchAllUsersWithChamberAccess() {
    loading.value = true
    error.value = null
    try {
      const response = await api.getAllUsersWithChamberAccess()
      if (response.success && response.data) {
        usersWithAccess.value = response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getUserChamberAccess(userId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.getUserChamberAccess(userId)
      if (response.success && response.data) {
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function setUserChamberAccess(userId: string, chamberIds: string[]) {
    loading.value = true
    error.value = null
    try {
      const response = await api.setUserChamberAccess(userId, chamberIds)
      if (response.success) {
        // Refresh the users with access list
        await fetchAllUsersWithChamberAccess()
        return true
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function grantChamberAccess(userId: string, chamberId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.grantChamberAccess(userId, chamberId)
      if (response.success) {
        // Refresh the users with access list
        await fetchAllUsersWithChamberAccess()
        return true
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function revokeChamberAccess(userId: string, chamberId: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.revokeChamberAccess(userId, chamberId)
      if (response.success) {
        // Refresh the users with access list
        await fetchAllUsersWithChamberAccess()
        return true
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function hasChamberAccess(userId: string, chamberId: string) {
    try {
      const response = await api.hasChamberAccess(userId, chamberId)
      if (response.success && response.data) {
        return response.data.has_access
      }
      return false
    } catch (err) {
      console.error('Error checking chamber access:', err)
      return false
    }
  }

  return {
    // State
    usersWithAccess,
    loading,
    error,
    // Actions
    fetchAllUsersWithChamberAccess,
    getUserChamberAccess,
    setUserChamberAccess,
    grantChamberAccess,
    revokeChamberAccess,
    hasChamberAccess
  }
}) 