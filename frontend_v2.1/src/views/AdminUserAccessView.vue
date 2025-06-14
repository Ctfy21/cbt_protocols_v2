<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <UsersIcon class="w-8 h-8 text-blue-600" />
            <h1 class="text-2xl font-bold text-gray-900">User Chamber Access Management</h1>
          </div>
          <button
            @click="refreshData"
            :disabled="loading"
            class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
            Refresh
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading State -->
      <div v-if="loading && usersWithAccess.length === 0" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-gray-600">Loading users...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="text-center py-12">
        <ExclamationCircleIcon class="w-16 h-16 text-red-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Error loading users</h3>
        <p class="text-gray-500">{{ error }}</p>
        <button
          @click="refreshData"
          class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Try Again
        </button>
      </div>

      <!-- Users List -->
      <div v-else-if="usersWithAccess.length > 0" class="space-y-6">
        <div
          v-for="userWithAccess in usersWithAccess"
          :key="userWithAccess.user.id"
          class="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
        >
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center space-x-3">
              <div class="flex-shrink-0">
                <div class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
                  <span class="text-white font-medium">{{ userWithAccess.user.name.charAt(0).toUpperCase() }}</span>
                </div>
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900">{{ userWithAccess.user.name }}</h3>
                <p class="text-sm text-gray-500">{{ userWithAccess.user.email }}</p>
                <div class="flex items-center mt-1">
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    userWithAccess.user.role === 'admin' 
                      ? 'bg-purple-100 text-purple-800' 
                      : 'bg-gray-100 text-gray-800'
                  ]">
                    {{ userWithAccess.user.role }}
                  </span>
                  <span :class="[
                    'ml-2 px-2 py-1 text-xs font-medium rounded-full',
                    userWithAccess.user.is_active 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  ]">
                    {{ userWithAccess.user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </div>
              </div>
            </div>
            <button
              @click="openEditModal(userWithAccess)"
              class="inline-flex items-center px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors text-sm"
            >
              <PencilIcon class="w-4 h-4 mr-2" />
              Edit Access
            </button>
          </div>

          <!-- Chamber Access -->
          <div class="mt-4">
            <h4 class="text-sm font-medium text-gray-900 mb-3">Chamber Access ({{ userWithAccess.chambers.length }})</h4>
            <div v-if="userWithAccess.chambers.length === 0" class="text-sm text-gray-500 italic">
              No chamber access assigned
            </div>
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              <div
                v-for="chamber in userWithAccess.chambers"
                :key="chamber.id"
                class="flex items-center p-3 bg-gray-50 rounded-lg"
              >
                <div class="flex-1">
                  <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                  <p class="text-xs text-gray-500">{{ chamber.location || 'No location' }}</p>
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
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
        <UsersIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">No users found</h3>
        <p class="text-gray-500">There are no users in the system yet.</p>
      </div>
    </main>

    <!-- Edit Modal -->
    <div v-if="editingUser" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">
              Edit Chamber Access - {{ editingUser.user.name }}
            </h2>
            <button
              @click="closeEditModal"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <!-- Loading State -->
          <div v-if="chamberStore.loading" class="text-center py-8">
            <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            <p class="mt-2 text-sm text-gray-600">Loading chambers...</p>
          </div>

          <!-- Chamber Selection -->
          <div v-else class="space-y-4">
            <div class="border border-gray-200 rounded-lg p-4">
              <h3 class="text-lg font-medium text-gray-900 mb-4">Available Chambers</h3>
              <div v-if="chamberStore.chambers.length === 0" class="text-sm text-gray-500 italic">
                No chambers available
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="chamber in chamberStore.chambers"
                  :key="chamber.id"
                  class="flex items-center p-3 border border-gray-200 rounded-lg hover:bg-gray-50"
                >
                  <input
                    :id="`chamber-${chamber.id}`"
                    type="checkbox"
                    :value="chamber.id"
                    v-model="selectedChamberIds"
                    class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <label :for="`chamber-${chamber.id}`" class="ml-3 flex-1 cursor-pointer">
                    <div class="flex items-center justify-between">
                      <div>
                        <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                        <p class="text-xs text-gray-500">{{ chamber.location || 'No location' }}</p>
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
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex justify-end space-x-3 mt-6 pt-6 border-t border-gray-200">
            <button
              @click="closeEditModal"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="saveAccess"
              :disabled="saving"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              <span v-if="saving">Saving...</span>
              <span v-else>Save Changes</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  UsersIcon,
  PencilIcon,
  ArrowPathIcon,
  ExclamationCircleIcon,
  XMarkIcon
} from '@heroicons/vue/24/outline'
import { useUserChamberAccessStore, type UserWithChamberAccess } from '@/stores/userChamberAccess'
import { useChamberStore } from '@/stores/chamber'
import { useToastStore } from '@/stores/toast'

const userChamberAccessStore = useUserChamberAccessStore()
const chamberStore = useChamberStore()
const toastStore = useToastStore()

// State
const editingUser = ref<UserWithChamberAccess | null>(null)
const selectedChamberIds = ref<string[]>([])
const saving = ref(false)

// Computed
const usersWithAccess = computed(() => userChamberAccessStore.usersWithAccess)
const loading = computed(() => userChamberAccessStore.loading)
const error = computed(() => userChamberAccessStore.error)

// Methods
async function refreshData() {
  try {
    await userChamberAccessStore.fetchAllUsersWithChamberAccess()
  } catch (err) {
    console.error('Error refreshing data:', err)
  }
}

function openEditModal(userWithAccess: UserWithChamberAccess) {
  editingUser.value = userWithAccess
  selectedChamberIds.value = userWithAccess.chambers.map(c => c.id)
  
  // Load chambers if not already loaded
  if (chamberStore.chambers.length === 0) {
    chamberStore.fetchChambers()
  }
}

function closeEditModal() {
  editingUser.value = null
  selectedChamberIds.value = []
}

async function saveAccess() {
  if (!editingUser.value) return

  saving.value = true
  try {
    await userChamberAccessStore.setUserChamberAccess(
      editingUser.value.user.id,
      selectedChamberIds.value
    )
    
    toastStore.success('Access Updated', `Chamber access updated for ${editingUser.value.user.name}`)
    closeEditModal()
  } catch (err) {
    toastStore.error('Error', 'Failed to update chamber access')
    console.error('Error saving access:', err)
  } finally {
    saving.value = false
  }
}

// Initialize
onMounted(async () => {
  await refreshData()
})
</script> 