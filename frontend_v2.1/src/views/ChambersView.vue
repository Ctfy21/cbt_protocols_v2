<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <HomeIcon class="w-8 h-8 text-blue-600" />
            <h1 class="text-2xl font-bold text-gray-900">Chamber Management</h1>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading State -->
      <div v-if="chamberStore.loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-gray-600">Loading chambers...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="chamberStore.error" class="text-center py-12">
        <ExclamationCircleIcon class="w-16 h-16 text-red-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Error loading chambers</h3>
        <p class="text-gray-500">{{ chamberStore.error }}</p>
        <button
          @click="chamberStore.fetchChambers()"
          class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Retry
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="chamberStore.chambers.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
        <HomeIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">No chambers registered</h3>
        <p class="text-gray-500 mb-6">Get started by registering your first chamber</p>
        <button
          @click="showRegisterForm = true"
          class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <PlusIcon class="w-5 h-5 mr-2" />
          Register Chamber
        </button>
      </div>

      <!-- Chambers Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="chamber in chamberStore.chambers"
          :key="chamber.id"
          @click="selectAndNavigate(chamber)"
          class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow cursor-pointer"
        >
          <div class="flex items-start justify-between mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">{{ chamber.name }}</h3>
              <p class="text-sm text-gray-500">{{ chamber.location || 'No location' }}</p>
            </div>
            <div :class="[
              'px-2 py-1 text-xs font-medium rounded-full',
              chamber.status === 'online' 
                ? 'bg-green-100 text-green-800' 
                : 'bg-gray-100 text-gray-800'
            ]">
              {{ chamber.status }}
            </div>
          </div>
          
          <div class="space-y-2 text-sm text-gray-600">
            <div class="flex items-center">
              <GlobeAltIcon class="w-4 h-4 mr-2" />
              <span>{{ formatUrl(chamber.ha_url) }}</span>
            </div>
            <div class="flex items-center">
              <ClockIcon class="w-4 h-4 mr-2" />
              <span>Last seen: {{ formatDate(chamber.last_heartbeat) }}</span>
            </div>
          </div>

          <div class="mt-4 pt-4 border-t border-gray-100">
            <button
              @click.stop="selectAndNavigate(chamber)"
              class="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors text-sm"
            >
              Select Chamber
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { format } from 'date-fns'
import { 
  HomeIcon, 
  PlusIcon, 
  ExclamationCircleIcon,
  GlobeAltIcon,
  ClockIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useToastStore } from '@/stores/toast'
import type { Chamber } from '@/types'

const router = useRouter()
const chamberStore = useChamberStore()
const toastStore = useToastStore()

const showRegisterForm = ref(false)

// Expose variables to template
defineExpose({
  chamberStore,
  showRegisterForm,
  formatUrl,
  formatDate,
  selectChamber,
  selectAndNavigate,
  onRegisterSuccess
})

onMounted(async () => {
  await chamberStore.fetchChambers()
})

function formatUrl(url: string): string {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname + (urlObj.port ? ':' + urlObj.port : '')
  } catch {
    return url
  }
}

function formatDate(dateStr: string): string {
  try {
    return format(new Date(dateStr), 'MMM d, yyyy HH:mm')
  } catch {
    return 'Unknown'
  }
}

function selectChamber(chamber: Chamber) {
  chamberStore.selectChamber(chamber)
  toastStore.success('Chamber Selected', `Selected ${chamber.name}`)
}

function selectAndNavigate(chamber: Chamber) {
  selectChamber(chamber)
  router.push('/experiments')
}

function onRegisterSuccess(chamber: Chamber) {
  showRegisterForm.value = false
  toastStore.success('Chamber Registered', `Successfully registered ${chamber.name}`)
  selectAndNavigate(chamber)
}
</script> 