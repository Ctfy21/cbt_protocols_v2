<!-- <template>
    <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div class="max-w-md w-full space-y-8">
        <div>
          <div class="flex justify-center">
            <BeakerIcon class="w-16 h-16 text-blue-600" />
          </div>
          <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Create a new account
          </h2>
          <p class="mt-2 text-center text-sm text-gray-600">
            Or
            <router-link to="/login" class="font-medium text-blue-600 hover:text-blue-500">
              sign in to existing account
            </router-link>
          </p>
        </div>
        
        <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
          <div class="space-y-4">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700">
                Full Name
              </label>
              <input
                id="name"
                v-model="form.name"
                name="name"
                type="text"
                autocomplete="name"
                required
                class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="John Doe"
              />
            </div>
            
            <div>
              <label for="email" class="block text-sm font-medium text-gray-700">
                Email Address
              </label>
              <input
                id="email"
                v-model="form.email"
                name="email"
                type="email"
                autocomplete="email"
                required
                class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="john@example.com"
              />
            </div>
            
            <div>
              <label for="password" class="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                id="password"
                v-model="form.password"
                name="password"
                type="password"
                autocomplete="new-password"
                required
                minlength="6"
                class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Minimum 6 characters"
              />
            </div>
            
            <div>
              <label for="confirmPassword" class="block text-sm font-medium text-gray-700">
                Confirm Password
              </label>
              <input
                id="confirmPassword"
                v-model="form.confirmPassword"
                name="confirmPassword"
                type="password"
                autocomplete="new-password"
                required
                class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Confirm your password"
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
              {{ loading ? 'Creating account...' : 'Create account' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref, reactive } from 'vue'
  import { useRouter } from 'vue-router'
  import { BeakerIcon, XCircleIcon } from '@heroicons/vue/24/outline'
  import { useAuthStore } from '@/stores/auth'
  import { useToastStore } from '@/stores/toast'
  
  const router = useRouter()
  const authStore = useAuthStore()
  const toastStore = useToastStore()
  
  const form = reactive({
    name: '',
    email: '',
    password: '',
    confirmPassword: ''
  })
  
  const loading = ref(false)
  const error = ref('')
  
  async function handleSubmit() {
    loading.value = true
    error.value = ''
    
    // Validate passwords match
    if (form.password !== form.confirmPassword) {
      error.value = 'Passwords do not match'
      loading.value = false
      return
    }
    
    // Validate password length
    if (form.password.length < 6) {
      error.value = 'Password must be at least 6 characters'
      loading.value = false
      return
    }
    
    try {
      await authStore.register({
        name: form.name,
        email: form.email,
        password: form.password
      })
      
      toastStore.success('Welcome!', 'Account created successfully')
      router.push('/')
    } catch (err: any) {
      console.log(err)
      error.value = err.response?.data?.error || 'Failed to create account'
    } finally {
      loading.value = false
    }
  }
  </script> -->