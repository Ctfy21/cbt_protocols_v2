import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import './style.css'
import App from './App.vue'
import routes from './router'
import { useAuthStore } from './stores/auth'

const app = createApp(App)
const pinia = createPinia()
const router = createRouter({
  history: createWebHistory(),
  routes
})

// Use Pinia before router to ensure stores are available
app.use(pinia)

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Initialize auth store if not already done
  if (!authStore.initialized) {
    await authStore.initialize()
  }
  
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)
  const requiresAdmin = to.matched.some(record => record.meta.requiresAdmin === true)
  const isAuthenticated = authStore.isAuthenticated
  const isAdmin = authStore.isAdmin
  
  if (requiresAuth && !isAuthenticated) {
    // Redirect to login with return url
    next({
      path: '/login',
      query: { redirect: to.fullPath }
    })
  } else if (requiresAdmin && !isAdmin) {
    // Redirect non-admin users to home
    next('/')
  } else if (!requiresAuth && isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    // Redirect authenticated users away from login/register
    next('/')
  } else {
    next()
  }
})

// Add response interceptor for token refresh
import api from './services/api'

api.api.interceptors.response.use(
  response => response,
  async error => {
    const authStore = useAuthStore()
    const originalRequest = error.config
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        await authStore.refreshAccessToken()
        // Retry the original request with new token
        originalRequest.headers['Authorization'] = `Bearer ${authStore.token}`
        return api.api(originalRequest)
      } catch (refreshError) {
        // Refresh failed, redirect to login
        authStore.clearAuthData()
        router.push('/login')
        return Promise.reject(refreshError)
      }
    }
    
    return Promise.reject(error)
  }
)

app.use(router)
app.mount('#app')