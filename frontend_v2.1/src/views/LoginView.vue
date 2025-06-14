<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="max-w-md w-full space-y-8">
      <div>
        <div class="flex justify-center">
          <BeakerIcon class="w-16 h-16 text-blue-600" />
        </div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Вход в систему
        </h2>
        <!-- REMOVED: Register link paragraph -->
      </div>
      
      <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
        <div class="rounded-md shadow-sm -space-y-px">
          <div>
            <label for="email" class="sr-only">Email адрес</label>
            <input
              id="email"
              v-model="form.email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="Email адрес"
            />
          </div>
          <div>
            <label for="password" class="sr-only">Пароль</label>
            <input
              id="password"
              v-model="form.password"
              name="password"
              type="password"
              autocomplete="current-password"
              required
              class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="Пароль"
            />
          </div>
        </div>

        <div v-if="error" class="rounded-md bg-red-50 p-4">
          <div class="flex">
            <div class="flex-shrink-0">
              <XCircleIcon class="h-5 w-5 text-red-400" />
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800">{{ error }}</h3>
            </div>
          </div>
        </div>

        <div>
          <button
            type="submit"
            :disabled="loading"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="loading" class="absolute left-0 inset-y-0 flex items-center pl-3">
              <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
            </span>
            {{ loading ? 'Вход...' : 'Вход' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { BeakerIcon, XCircleIcon } from '@heroicons/vue/24/outline'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const toastStore = useToastStore()

const form = reactive({
  email: '',
  password: ''
})

const loading = ref(false)
const error = ref('')

onMounted(() => {
  // Check if already authenticated
  if (authStore.isAuthenticated) {
    router.push(route.query.redirect as string || '/')
  }
})

async function handleSubmit() {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.login({
      email: form.email,
      password: form.password
    })
    
    toastStore.success('Добро пожаловать!', `Вход выполнен как ${authStore.userName}`)
    
    // Redirect to intended page or home
    const redirectTo = route.query.redirect as string || '/'
    router.push(redirectTo)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Неверный email или пароль'
  } finally {
    loading.value = false
  }
}
</script>