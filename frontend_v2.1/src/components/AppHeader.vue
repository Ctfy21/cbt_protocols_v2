<template>
  <header class="bg-white shadow-sm border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center py-4">
        <div class="flex items-center space-x-3">
          <div class="flex items-center space-x-2">
            <BeakerIcon class="w-8 h-8 text-blue-600" />
            <div>
              <h1 class="text-xl font-bold text-gray-900">Environmental Control Lab</h1>
              <div class="flex items-center gap-2 text-sm text-gray-600">
                <div class="flex items-center gap-1">
                  <div :class="[
                    'w-2 h-2 rounded-full',
                    chamberStore.selectedChamber?.status === 'online' ? 'bg-green-500' : 'bg-gray-400'
                  ]"></div>
                  <span>{{ chamberStore.selectedChamber?.name || 'No chamber selected' }}</span>
                </div>
                <span v-if="chamberStore.selectedChamber?.location" class="text-gray-400">•</span>
                <span v-if="chamberStore.selectedChamber?.location">{{ chamberStore.selectedChamber.location }}</span>
              </div>
            </div>
          </div>
          <span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded-full">
            v0.2 pre-alpha
          </span>
        </div>
        
        <div class="flex items-center space-x-4">
          <router-link
            to="/chambers"
            class="text-gray-600 hover:text-gray-900 text-sm font-medium"
          >
            Switch Chamber
          </router-link>
          
          <!-- User Menu -->
          <div class="relative">
            <button
              @click.stop="showUserMenu = !showUserMenu"
              class="flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <div class="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                <UserIcon class="w-5 h-5 text-gray-600" />
              </div>
              <ChevronDownIcon class="ml-1 w-4 h-4 text-gray-500" />
            </button>
            
            <!-- Dropdown Menu -->
            <Transition
              enter-active-class="transition ease-out duration-100"
              enter-from-class="transform opacity-0 scale-95"
              enter-to-class="transform opacity-100 scale-100"
              leave-active-class="transition ease-in duration-75"
              leave-from-class="transform opacity-100 scale-100"
              leave-to-class="transform opacity-0 scale-95"
            >
              <div
                v-if="showUserMenu"
                v-click-outside="() => showUserMenu = false"
                class="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-100"
              >
                <div class="px-4 py-3">
                  <p class="text-sm">Signed in as</p>
                  <p class="text-sm font-medium text-gray-900 truncate">
                    {{ authStore.userEmail }}
                  </p>
                </div>
                
                <div class="py-1">
                  <router-link
                    to="/profile"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <UserIcon class="w-4 h-4 mr-3" />
                      My Profile
                    </div>
                  </router-link>
                  <router-link
                    to="/experiments"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <BeakerIcon class="w-4 h-4 mr-3" />
                      Experiments
                    </div>
                  </router-link>
                  <router-link
                    to="/api-tokens"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <KeyIcon class="w-4 h-4 mr-3" />
                      API Tokens
                    </div>
                  </router-link>
                </div>

                <!-- Admin Menu Items -->
                <div v-if="authStore.isAdmin" class="py-1">
                  <router-link
                    to="/admin/users"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <UsersIcon class="w-4 h-4 mr-3" />
                      Manage Users
                    </div>
                  </router-link>
                </div>
                
                <div class="py-1">
                  <button
                    @click="handleLogout"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <ArrowRightStartOnRectangleIcon class="w-4 h-4 mr-3" />
                      Sign out
                    </div>
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { 
  BeakerIcon, 
  UserIcon, 
  ChevronDownIcon,
  ArrowRightStartOnRectangleIcon,
  KeyIcon,
  UsersIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const router = useRouter()
const chamberStore = useChamberStore()
const authStore = useAuthStore()
const toastStore = useToastStore()

const showUserMenu = ref(false)

async function handleLogout() {
  showUserMenu.value = false
  try {
    await authStore.logout()
    toastStore.info('Signed out', 'You have been logged out successfully')
    router.push('/login')
  } catch (error) {
    toastStore.error('Error', 'Failed to log out')
  }
}

// Click outside directive
interface ClickOutsideElement extends HTMLElement {
  clickOutsideEvent?: (event: MouseEvent) => void
}

const vClickOutside = {
  mounted(el: ClickOutsideElement, binding: any) {
    el.clickOutsideEvent = function(event: MouseEvent) {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value()
      }
    }
    document.addEventListener('click', el.clickOutsideEvent)
  },
  unmounted(el: ClickOutsideElement) {
    if (el.clickOutsideEvent) {
      document.removeEventListener('click', el.clickOutsideEvent)
    }
  }
}
</script>