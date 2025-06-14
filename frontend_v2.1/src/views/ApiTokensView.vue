<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <AppHeader />

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">API Tokens</h1>
          <p class="text-gray-600 mt-1">Manage your API access tokens</p>
        </div>
        <button
          @click="showCreateForm = true"
          class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-sm"
        >
          <PlusIcon class="w-5 h-5 mr-2" />
          New Token
        </button>
      </div>

      <!-- Token List -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Expires</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                <th scope="col" class="relative px-6 py-3">
                  <span class="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="token in tokens" :key="token.id">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ token.name }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    token.type === 'service' ? 'bg-purple-100 text-purple-800' : 'bg-blue-100 text-blue-800'
                  ]">
                    {{ token.type }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ token.service_name || '-' }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {{ token.expires_at ? formatDate(token.expires_at) : 'Never' }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(token.created_at) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex justify-end space-x-3">
                    <button
                      @click="watchToken(token)"
                      class="text-blue-600 hover:text-blue-900"
                    >
                      Watch
                    </button>
                    <button
                      @click="deleteToken(token)"
                      class="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Create Token Modal -->
    <div v-if="showCreateForm" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">Create New API Token</h3>
        </div>
        
        <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
          <div>
            <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
            <input
              id="name"
              v-model="form.name"
              type="text"
              required
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              placeholder="My API Token"
            />
          </div>

          <div>
            <label for="type" class="block text-sm font-medium text-gray-700">Type</label>
            <select
              id="type"
              v-model="form.type"
              required
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
            >
              <option value="personal">Personal</option>
              <option value="service">Service</option>
            </select>
          </div>

          <div v-if="form.type === 'service'">
            <label for="serviceName" class="block text-sm font-medium text-gray-700">Service Name</label>
            <input
              id="serviceName"
              v-model="form.service_name"
              type="text"
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              placeholder="My Service"
            />
          </div>

          <div>
            <label for="permissions" class="block text-sm font-medium text-gray-700">Permissions</label>
            <div class="mt-2 space-y-2">
              <div v-for="permission in availablePermissions" :key="permission" class="flex items-center">
                <input
                  :id="permission"
                  v-model="form.permissions"
                  :value="permission"
                  type="checkbox"
                  class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label :for="permission" class="ml-2 block text-sm text-gray-700">
                  {{ permission }}
                </label>
              </div>
            </div>
          </div>

          <div>
            <label for="expiresAt" class="block text-sm font-medium text-gray-700">Expires At (Optional)</label>
            <input
              id="expiresAt"
              v-model="form.expires_at"
              type="date"
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
            />
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              @click="showCreateForm = false"
              class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="loading"
              class="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              {{ loading ? 'Creating...' : 'Create Token' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <ConfirmDialog
      v-if="deletingToken"
      title="Delete API Token"
      :message="`Are you sure you want to delete '${deletingToken.name}'? This action cannot be undone.`"
      confirm-text="Delete"
      confirm-class="bg-red-600 hover:bg-red-700"
      @confirm="confirmDelete"
      @cancel="deletingToken = null"
    />

    <!-- Watch Token Modal -->
    <div v-if="watchingToken" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
          <h3 class="text-lg font-medium text-gray-900">Watching Token: {{ watchingToken.name }}</h3>
          <button @click="stopWatching" class="text-gray-400 hover:text-gray-500">
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="p-6">
          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700">Permissions</h4>
              <div class="mt-2">
                <div v-for="permission in watchingToken.permissions" :key="permission" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 mr-2 mb-2">
                  {{ permission }}
                </div>
              </div>
            </div>
            <div>
              <div class="flex items-center justify-between">
                <h4 class="text-sm font-medium text-gray-700">Token Value</h4>
                <div class="flex items-center space-x-2">
                  <button
                    @click="copyToken(watchingToken.token)"
                    class="text-sm text-blue-600 hover:text-blue-900"
                  >
                    Copy
                  </button>
                  <button
                    @click="showToken = !showToken"
                    class="text-sm text-gray-600 hover:text-gray-900"
                  >
                    {{ showToken ? 'Hide' : 'Show' }}
                  </button>
                </div>
              </div>
              <div class="mt-2">
                <div class="relative">
                  <input
                    :type="showToken ? 'text' : 'password'"
                    :value="watchingToken.token"
                    readonly
                    class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 text-sm font-mono"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { format } from 'date-fns'
import { PlusIcon } from '@heroicons/vue/24/outline'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import api from '@/services/api'
import type { ApiToken } from '@/types/auth'

const toastStore = useToastStore()
const showCreateForm = ref(false)
const loading = ref(false)
const deletingToken = ref<ApiToken | null>(null)
const watchingToken = ref<ApiToken | null>(null)
const showToken = ref(false)
const tokens = ref<ApiToken[]>([])

const availablePermissions = [
  'admin',
]

const form = reactive({
  name: '',
  type: 'personal' as 'personal' | 'service',
  service_name: '',
  permissions: [] as string[],
  expires_at: format(new Date(), 'yyyy-MM-dd'),
})

function formatDate(date: string): string {
  try {
    return format(new Date(date), 'MMM d, yyyy HH:mm')
  } catch {
    return 'Invalid date'
  }
}

async function fetchTokens() {
  try {
    const response = await api.getApiTokens()
    if (response.success && response.data) {
      tokens.value = response.data
    }
  } catch (error: any) {
    toastStore.error('Error', error.message || 'Failed to fetch tokens')
  }
}

async function handleSubmit() {
  loading.value = true
  try {
    const response = await api.createApiToken({
      name: form.name,
      type: form.type,
      service_name: form.type === 'service' ? form.service_name : undefined,
      permissions: form.permissions,
      expires_at: String(Math.floor(new Date(form.expires_at).getTime() / 1000)),
    })
    
    if (response.success && response.data) {
      tokens.value.push(response.data)
      toastStore.success('Token Created', 'New API token created successfully')
      showCreateForm.value = false
      
      // Reset form
      form.name = ''
      form.type = 'personal'
      form.service_name = ''
      form.permissions = []
      form.expires_at = ''
    }
  } catch (error: any) {
    toastStore.error('Error', error.message || 'Failed to create token')
  } finally {
    loading.value = false
  }
}

function deleteToken(token: ApiToken) {
  deletingToken.value = token
}

async function confirmDelete() {
  if (!deletingToken.value) return
  
  try {
    const response = await api.deleteApiToken(deletingToken.value.id)
    if (response.success) {
      tokens.value = tokens.value.filter(t => t.id !== deletingToken.value!.id)
      toastStore.success('Token Deleted', 'API token deleted successfully')
      deletingToken.value = null
    }
  } catch (error: any) {
    toastStore.error('Error', error.message || 'Failed to delete token')
  }
}

function watchToken(token: ApiToken) {
  watchingToken.value = token
  showToken.value = false
}

async function stopWatching() {
  watchingToken.value = null
  showToken.value = false
}

async function copyToken(token: string) {
  try {
    await navigator.clipboard.writeText(token)
    toastStore.success('Copied', 'Token copied to clipboard')
  } catch (error) {
    toastStore.error('Error', 'Failed to copy token')
  }
}

onMounted(() => {
  fetchTokens()
})
</script> 