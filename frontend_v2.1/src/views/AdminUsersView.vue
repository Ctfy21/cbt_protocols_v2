<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <UsersIcon class="w-8 h-8 text-blue-600" />
            <h1 class="text-2xl font-bold text-gray-900">User Management</h1>
          </div>
          <div class="flex items-center space-x-3">
            <button
              @click="$router.back()"
              class="inline-flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <ArrowLeftIcon class="w-4 h-4 mr-2" />
              Back
            </button>
            <button
              @click="refreshData"
              :disabled="loading"
              class="inline-flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
            >
              <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
              Refresh
            </button>
            <button
              @click="showCreateUserForm = true"
              class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <PlusIcon class="w-4 h-4 mr-2" />
              Add User
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <UsersIcon class="w-8 h-8 text-blue-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Total Users</p>
              <p class="text-3xl font-bold text-gray-900">{{ users.length }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <CheckCircleIcon class="w-8 h-8 text-green-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Active Users</p>
              <p class="text-3xl font-bold text-gray-900">{{ activeUsersCount }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <ShieldCheckIcon class="w-8 h-8 text-purple-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Admins</p>
              <p class="text-3xl font-bold text-gray-900">{{ adminUsersCount }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <HomeIcon class="w-8 h-8 text-orange-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Total Chambers</p>
              <p class="text-3xl font-bold text-gray-900">{{ chamberStore.chambers.length }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div class="flex flex-col sm:flex-row gap-4">
          <!-- Search -->
          <div class="flex-1">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search users..."
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          <!-- Role Filter -->
          <div class="sm:w-48">
            <select
              v-model="roleFilter"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">All Roles</option>
              <option value="admin">Admin</option>
              <option value="user">User</option>
            </select>
          </div>

          <!-- Status Filter -->
          <div class="sm:w-48">
            <select
              v-model="statusFilter"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">All Status</option>
              <option value="true">Active</option>
              <option value="false">Inactive</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading && users.length === 0" class="text-center py-12">
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

      <!-- Users Table -->
      <div v-else class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  User
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Role
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Chamber Access
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Login
                </th>
                <th scope="col" class="relative px-6 py-3">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="userWithAccess in filteredUsers" :key="userWithAccess.user.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <div class="flex-shrink-0">
                      <div class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
                        <span class="text-white font-medium text-sm">
                          {{ userWithAccess.user.name.charAt(0).toUpperCase() }}
                        </span>
                      </div>
                    </div>
                    <div class="ml-4">
                      <div class="text-sm font-medium text-gray-900">{{ userWithAccess.user.name }}</div>
                      <div class="text-sm text-gray-500">{{ userWithAccess.user.email }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    userWithAccess.user.role === 'admin' 
                      ? 'bg-purple-100 text-purple-800' 
                      : 'bg-gray-100 text-gray-800'
                  ]">
                    {{ userWithAccess.user.role }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    userWithAccess.user.is_active 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  ]">
                    {{ userWithAccess.user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <span class="text-sm text-gray-900">{{ userWithAccess.chambers.length }} chambers</span>
                    <button
                      @click="manageChamberAccess(userWithAccess)"
                      class="ml-2 text-blue-600 hover:text-blue-900 text-sm"
                    >
                      Manage
                    </button>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {{ formatDate(userWithAccess.user.last_login) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex justify-end space-x-3">
                    <button
                      @click="editUser(userWithAccess.user)"
                      class="text-blue-600 hover:text-blue-900"
                    >
                      Edit
                    </button>
                    <button
                      v-if="userWithAccess.user.is_active"
                      @click="toggleUserStatus(userWithAccess.user, false)"
                      class="text-yellow-600 hover:text-yellow-900"
                    >
                      Deactivate
                    </button>
                    <button
                      v-else
                      @click="toggleUserStatus(userWithAccess.user, true)"
                      class="text-green-600 hover:text-green-900"
                    >
                      Activate
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Empty State -->
        <div v-if="filteredUsers.length === 0" class="text-center py-12">
          <UsersIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
          <h3 class="text-lg font-medium text-gray-900 mb-2">No users found</h3>
          <p class="text-gray-500">{{ users.length === 0 ? 'No users in the system yet.' : 'No users match your current filters.' }}</p>
        </div>
      </div>
    </main>

    <!-- Create User Modal -->
    <div v-if="showCreateUserForm" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">Create New User</h2>
            <button
              @click="closeCreateUserForm"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <form @submit.prevent="createUser" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                v-model="newUserForm.name"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Full name"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input
                v-model="newUserForm.email"
                type="email"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="user@example.com"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
              <input
                v-model="newUserForm.password"
                type="password"
                required
                minlength="6"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Minimum 6 characters"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
              <select
                v-model="newUserForm.role"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                @click="closeCreateUserForm"
                class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                :disabled="creatingUser"
                class="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
              >
                {{ creatingUser ? 'Creating...' : 'Create User' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Edit User Modal -->
    <div v-if="editingUser" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">Edit User</h2>
            <button
              @click="closeEditUserForm"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <form @submit.prevent="updateUser" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                v-model="editUserForm.name"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input
                v-model="editUserForm.email"
                type="email"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
              <select
                v-model="editUserForm.role"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                @click="closeEditUserForm"
                class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                :disabled="updatingUser"
                class="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
              >
                {{ updatingUser ? 'Updating...' : 'Update User' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Chamber Access Modal -->
    <div v-if="managingAccessUser" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">
              Manage Chamber Access - {{ managingAccessUser.user.name }}
            </h2>
            <button
              @click="closeChamberAccessModal"
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
              @click="closeChamberAccessModal"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="saveChamberAccess"
              :disabled="savingAccess"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              <span v-if="savingAccess">Saving...</span>
              <span v-else>Save Changes</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { format } from 'date-fns'
import {
  UsersIcon,
  PlusIcon,
  ArrowPathIcon,
  ExclamationCircleIcon,
  XMarkIcon,
  MagnifyingGlassIcon,
  CheckCircleIcon,
  ShieldCheckIcon,
  HomeIcon,
  ArrowLeftIcon
} from '@heroicons/vue/24/outline'
import { useUserChamberAccessStore, type UserWithChamberAccess } from '@/stores/userChamberAccess'
import { useChamberStore } from '@/stores/chamber'
import { useToastStore } from '@/stores/toast'
import api from '@/services/api'
import type { User } from '@/types/auth'

const userChamberAccessStore = useUserChamberAccessStore()
const chamberStore = useChamberStore()
const toastStore = useToastStore()

// State
const users = ref<UserWithChamberAccess[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')
const roleFilter = ref('')
const statusFilter = ref('')

// User Creation
const showCreateUserForm = ref(false)
const creatingUser = ref(false)
const newUserForm = reactive({
  name: '',
  email: '',
  password: '',
  role: 'user' as 'user' | 'admin'
})

// User Editing
const editingUser = ref<User | null>(null)
const updatingUser = ref(false)
const editUserForm = reactive({
  name: '',
  email: '',
  role: 'user' as 'user' | 'admin'
})

// Chamber Access Management
const managingAccessUser = ref<UserWithChamberAccess | null>(null)
const selectedChamberIds = ref<string[]>([])
const savingAccess = ref(false)

// Computed
const filteredUsers = computed(() => {
  let filtered = users.value

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(userWithAccess => 
      userWithAccess.user.name.toLowerCase().includes(query) ||
      userWithAccess.user.email.toLowerCase().includes(query)
    )
  }

  // Role filter
  if (roleFilter.value) {
    filtered = filtered.filter(userWithAccess => userWithAccess.user.role === roleFilter.value)
  }

  // Status filter
  if (statusFilter.value) {
    const isActive = statusFilter.value === 'true'
    filtered = filtered.filter(userWithAccess => userWithAccess.user.is_active === isActive)
  }

  return filtered
})

const activeUsersCount = computed(() => 
  users.value.filter(userWithAccess => userWithAccess.user.is_active).length
)

const adminUsersCount = computed(() => 
  users.value.filter(userWithAccess => userWithAccess.user.role === 'admin').length
)

// Methods
async function refreshData() {
  loading.value = true
  error.value = null
  try {
    await userChamberAccessStore.fetchAllUsersWithChamberAccess()
    users.value = userChamberAccessStore.usersWithAccess
    
    // Load chambers if not already loaded
    if (chamberStore.chambers.length === 0) {
      await chamberStore.fetchChambers()
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load data'
  } finally {
    loading.value = false
  }
}

function formatDate(date: string | undefined): string {
  if (!date) return 'Never'
  try {
    return format(new Date(date), 'MMM d, yyyy')
  } catch {
    return 'Invalid date'
  }
}

// User Creation
function closeCreateUserForm() {
  showCreateUserForm.value = false
  newUserForm.name = ''
  newUserForm.email = ''
  newUserForm.password = ''
  newUserForm.role = 'user'
}

async function createUser() {
  creatingUser.value = true
  try {
    const response = await api.api.post('/users', {
      name: newUserForm.name,
      email: newUserForm.email,
      password: newUserForm.password,
      role: newUserForm.role
    })
    
    if (response.data.success) {
      toastStore.success('User Created', `Successfully created user ${newUserForm.name}`)
      closeCreateUserForm()
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Error', api.formatError(err))
  } finally {
    creatingUser.value = false
  }
}

// User Editing
function editUser(user: User) {
  editingUser.value = user
  editUserForm.name = user.name
  editUserForm.email = user.email
  editUserForm.role = user.role
}

function closeEditUserForm() {
  editingUser.value = null
  editUserForm.name = ''
  editUserForm.email = ''
  editUserForm.role = 'user'
}

async function updateUser() {
  if (!editingUser.value) return
  
  updatingUser.value = true
  try {
    const response = await api.api.put(`/users/${editingUser.value.id}`, {
      name: editUserForm.name,
      email: editUserForm.email,
      role: editUserForm.role
    })
    
    if (response.data.success) {
      toastStore.success('User Updated', `Successfully updated ${editUserForm.name}`)
      closeEditUserForm()
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Error', api.formatError(err))
  } finally {
    updatingUser.value = false
  }
}

async function toggleUserStatus(user: User, isActive: boolean) {
  try {
    const action = isActive ? 'activate' : 'deactivate'
    const endpoint = isActive ? `/users/${user.id}/activate` : `/users/${user.id}`
    const method = isActive ? 'post' : 'delete'
    
    const response = await api.api[method](endpoint)
    
    if (response.data.success) {
      toastStore.success(
        `User ${isActive ? 'Activated' : 'Deactivated'}`, 
        `Successfully ${action}d ${user.name}`
      )
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Error', api.formatError(err))
  }
}

// Chamber Access Management
function manageChamberAccess(userWithAccess: UserWithChamberAccess) {
  managingAccessUser.value = userWithAccess
  selectedChamberIds.value = userWithAccess.chambers.map(c => c.id)
  
  // Load chambers if not already loaded
  if (chamberStore.chambers.length === 0) {
    chamberStore.fetchChambers()
  }
}

function closeChamberAccessModal() {
  managingAccessUser.value = null
  selectedChamberIds.value = []
}

async function saveChamberAccess() {
  if (!managingAccessUser.value) return

  savingAccess.value = true
  try {
    await userChamberAccessStore.setUserChamberAccess(
      managingAccessUser.value.user.id,
      selectedChamberIds.value
    )
    
    toastStore.success('Access Updated', `Chamber access updated for ${managingAccessUser.value.user.name}`)
    closeChamberAccessModal()
    await refreshData()
  } catch (err) {
    toastStore.error('Error', 'Failed to update chamber access')
  } finally {
    savingAccess.value = false
  }
}

// Initialize
onMounted(async () => {
  await refreshData()
})
</script>