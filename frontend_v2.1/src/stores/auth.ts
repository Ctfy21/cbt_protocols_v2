import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import type { User, LoginCredentials, RegisterData } from '@/types/auth'
import experimentTracker from '@/services/experimentTracker'

const TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const USER_KEY = 'auth_user'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const initialized = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const Username = computed(() => user.value?.username || '') 

  // Actions
  async function login(credentials: LoginCredentials) {
    loading.value = true
    error.value = null
    try {
      const response = await api.login(credentials)
      if (response.success && response.data) {
        setAuthData(response.data)
        // Запускаем отслеживание экспериментов после успешного логина
        experimentTracker.startTracking()
        return response.data
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(data: RegisterData) {
    loading.value = true
    error.value = null
    try {
      const response = await api.register(data)
      if (response.success && response.data) {
        setAuthData(response.data)
        return response.data
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      if (token.value) {
        await api.logout()
      }
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      clearAuthData()
      // Останавливаем отслеживание при выходе
      experimentTracker.stopTracking()
      const router = useRouter()
      router.push('/login')
    }
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await api.refreshToken(refreshToken.value)
      if (response.success && response.data) {
        setAuthData(response.data)
        return response.data
      }
    } catch (err) {
      clearAuthData()
      throw err
    }
  }

  async function fetchCurrentUser() {
    if (!token.value) {
      return null
    }

    loading.value = true
    error.value = null
    try {
      const response = await api.getCurrentUser()
      if (response.success && response.data) {
        user.value = response.data
        saveToStorage()
        return response.data
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch user'
      if (err.response?.status === 401) {
        clearAuthData()
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data: { name?: string }) {
    loading.value = true
    error.value = null
    try {
      const response = await api.updateProfile(data)
      if (response.success && response.data) {
        user.value = response.data
        saveToStorage()
        return response.data
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update profile'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function changePassword(currentPassword: string, newPassword: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.changePassword(currentPassword, newPassword)
      if (response.success) {
        return true
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to change password'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Helper functions
  function setAuthData(authData: any) {
    user.value = authData.user
    token.value = authData.token
    refreshToken.value = authData.refresh_token
    
    // Set axios default header
    api.setAuthToken(authData.token)
    
    // Save to localStorage
    saveToStorage()
  }

  function clearAuthData() {
    user.value = null
    token.value = null
    refreshToken.value = null
    
    // Remove axios default header
    api.setAuthToken(null)
    
    // Clear localStorage
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    
  }

  function saveToStorage() {
    if (token.value) {
      localStorage.setItem(TOKEN_KEY, token.value)
    }
    if (refreshToken.value) {
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken.value)
    }
    if (user.value) {
      localStorage.setItem(USER_KEY, JSON.stringify(user.value))
    }
  }

  function loadFromStorage() {
    const savedToken = localStorage.getItem(TOKEN_KEY)
    const savedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    const savedUser = localStorage.getItem(USER_KEY)

    if (savedToken && savedRefreshToken && savedUser) {
      try {
        token.value = savedToken
        refreshToken.value = savedRefreshToken
        user.value = JSON.parse(savedUser)
        
        // Set axios default header
        api.setAuthToken(savedToken)
        
        return true
      } catch (err) {
        console.error('Failed to parse saved user:', err)
        clearAuthData()
        return false
      }
    }
    
    return false
  }

  async function initialize() {
    if (initialized.value) return

    initialized.value = true
    // Try to load from storage
    if (loadFromStorage()) {
      // Set the token in axios headers
      api.setAuthToken(token.value)
      
              // Verify token is still valid
        try {
          await fetchCurrentUser()
          // Если токен валиден, запускаем отслеживание
          experimentTracker.startTracking()
        } catch (err) {
        console.log('Token validation failed, attempting refresh...')
        // Token might be expired, try to refresh
        if (refreshToken.value) {
          try {
            await refreshAccessToken()
            // After successful refresh, try fetching user again
            await fetchCurrentUser()
            // Запускаем отслеживание после успешного обновления токена
            experimentTracker.startTracking()
          } catch (refreshErr) {
            console.error('Token refresh failed:', refreshErr)
            // Both failed, clear auth data
            clearAuthData()
          }
        } else {
          console.log('No refresh token available')
          clearAuthData()
        }
      }
    } else {
      console.log('No auth data found in storage')
    }
  }

  return {
    // State
    user,
    token,
    refreshToken,
    loading,
    error,
    initialized,
    
    // Getters
    isAuthenticated,
    isAdmin,
    Username,
    
    // Actions
    login,
    register,
    logout,
    refreshAccessToken,
    fetchCurrentUser,
    updateProfile,
    changePassword,
    initialize,
    loadFromStorage,
    clearAuthData
  }
})