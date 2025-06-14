<template>
    <div class="min-h-screen bg-gray-50">
      <!-- Header -->
      <AppHeader />
  
      <!-- Main Content -->
      <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="mb-6">
          <h1 class="text-2xl font-bold text-gray-900">Мой профиль</h1>
        </div>
  
        <div class="bg-white shadow overflow-hidden sm:rounded-lg">
          <!-- User Info -->
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Информация о пользователе</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">Личные данные и настройки учетной записи.</p>
          </div>
          
          <div class="border-t border-gray-200">
            <dl>
              <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt class="text-sm font-medium text-gray-500">Полное имя</dt>
                <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  <div v-if="!editingName" class="flex items-center justify-between">
                    <span>{{ authStore.user?.name }}</span>
                    <button
                      @click="startEditingName"
                      class="text-blue-600 hover:text-blue-700"
                    >
                      <PencilIcon class="w-4 h-4" />
                    </button>
                  </div>
                  <form v-else @submit.prevent="updateName" class="flex items-center gap-2">
                    <input
                      v-model="newName"
                      type="text"
                      class="flex-1 px-3 py-1 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="Введите ваше имя"
                    />
                    <button
                      type="submit"
                      :disabled="updatingName"
                      class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                    >
                      Сохранить
                    </button>
                    <button
                      type="button"
                      @click="cancelEditingName"
                      class="px-3 py-1 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
                    >
                      Отмена
                    </button>
                  </form>
                </dd>
              </div>
              
              <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt class="text-sm font-medium text-gray-500">Email адрес</dt>
                <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  {{ authStore.user?.email }}
                </dd>
              </div>
              
              <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt class="text-sm font-medium text-gray-500">Роль учетной записи</dt>
                <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded-full',
                    authStore.user?.role === 'admin' 
                      ? 'bg-purple-100 text-purple-800' 
                      : 'bg-gray-100 text-gray-800'
                  ]">
                    {{ authStore.user?.role }}
                  </span>
                </dd>
              </div>
              
              <div class="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt class="text-sm font-medium text-gray-500">Участник с</dt>
                <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  {{ formatDate(authStore.user?.created_at) }}
                </dd>
              </div>
              
              <div class="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt class="text-sm font-medium text-gray-500">Последний вход</dt>
                <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                  {{ formatDate(authStore.user?.last_login) }}
                </dd>
              </div>
            </dl>
          </div>
        </div>
  
        <!-- Change Password -->
        <div class="mt-6 bg-white shadow overflow-hidden sm:rounded-lg">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900">Безопасность</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">Обновить ваш пароль.</p>
          </div>
          
          <div class="border-t border-gray-200 px-4 py-5 sm:px-6">
            <button
              v-if="!changingPassword"
              @click="changingPassword = true"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Изменить пароль
            </button>
            
            <form v-else @submit.prevent="updatePassword" class="space-y-4 max-w-md">
              <div>
                <label for="currentPassword" class="block text-sm font-medium text-gray-700">
                  Текущий пароль
                </label>
                <input
                  id="currentPassword"
                  v-model="passwordForm.currentPassword"
                  type="password"
                  required
                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                />
              </div>
              
              <div>
                <label for="newPassword" class="block text-sm font-medium text-gray-700">
                  Новый пароль
                </label>
                <input
                  id="newPassword"
                  v-model="passwordForm.newPassword"
                  type="password"
                  required
                  minlength="6"
                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                />
              </div>
              
              <div>
                <label for="confirmNewPassword" class="block text-sm font-medium text-gray-700">
                  Подтвердить новый пароль
                </label>
                <input
                  id="confirmNewPassword"
                  v-model="passwordForm.confirmNewPassword"
                  type="password"
                  required
                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                />
              </div>
              
              <div v-if="passwordError" class="rounded-md bg-red-50 p-4">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <XCircleIcon class="h-5 w-5 text-red-400" />
                  </div>
                  <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800">{{ passwordError }}</h3>
                  </div>
                </div>
              </div>
              
              <div class="flex gap-2">
                <button
                  type="submit"
                  :disabled="updatingPassword"
                  class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {{ updatingPassword ? 'Обновление...' : 'Обновить пароль' }}
                </button>
                <button
                  type="button"
                  @click="cancelPasswordChange"
                  class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  Отмена
                </button>
              </div>
            </form>
          </div>
        </div>
      </main>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref, reactive } from 'vue'
  import { format } from 'date-fns'
  import { PencilIcon, XCircleIcon } from '@heroicons/vue/24/outline'
  import { useAuthStore } from '@/stores/auth'
  import { useToastStore } from '@/stores/toast'
  import AppHeader from '@/components/AppHeader.vue'
  
  const authStore = useAuthStore()
  const toastStore = useToastStore()
  
  // Name editing
  const editingName = ref(false)
  const newName = ref('')
  const updatingName = ref(false)
  
  // Password change
  const changingPassword = ref(false)
  const updatingPassword = ref(false)
  const passwordError = ref('')
  const passwordForm = reactive({
    currentPassword: '',
    newPassword: '',
    confirmNewPassword: ''
  })
  
  function formatDate(date: string | undefined): string {
    if (!date) return 'Never'
    try {
      return format(new Date(date), 'MMM d, yyyy')
    } catch {
      return 'Invalid date'
    }
  }
  
  function startEditingName() {
    newName.value = authStore.user?.name || ''
    editingName.value = true
  }
  
  function cancelEditingName() {
    editingName.value = false
    newName.value = ''
  }
  
  async function updateName() {
    if (!newName.value.trim()) return
    
    updatingName.value = true
    try {
      await authStore.updateProfile({ name: newName.value.trim() })
      toastStore.success('Имя обновлено', 'Имя обновлено успешно')
      editingName.value = false
    } catch (error) {
      toastStore.error('Ошибка', 'Не удалось обновить имя')
    } finally {
      updatingName.value = false
    }
  }
  
  function cancelPasswordChange() {
    changingPassword.value = false
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmNewPassword = ''
    passwordError.value = ''
  }
  
  async function updatePassword() {
    passwordError.value = ''
    
    // Validate passwords match
    if (passwordForm.newPassword !== passwordForm.confirmNewPassword) {
      passwordError.value = 'Новые пароли не совпадают'
      return
    }
    
    // Validate password length
    if (passwordForm.newPassword.length < 6) {
      passwordError.value = 'Пароль должен быть не менее 6 символов'
      return
    }
    
    updatingPassword.value = true
    try {
      await authStore.changePassword(passwordForm.currentPassword, passwordForm.newPassword)
      toastStore.success('Пароль обновлен', 'Пароль обновлен успешно')
      cancelPasswordChange()
    } catch (error: any) {
      passwordError.value = error.response?.data?.error || 'Не удалось обновить пароль'
    } finally {
      updatingPassword.value = false
    }
  }
  </script>