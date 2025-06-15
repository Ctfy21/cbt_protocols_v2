<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <UsersIcon class="w-8 h-8 text-blue-600" />
            <h1 class="text-2xl font-bold text-gray-900">Управление пользователями</h1>
          </div>
          <div class="flex items-center space-x-3">
            <button
              @click="$router.back()"
              class="inline-flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <ArrowLeftIcon class="w-4 h-4 mr-2" />
              Назад
            </button>
            <button
              @click="refreshData"
              :disabled="loading"
              class="inline-flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
            >
              <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
              Обновить
            </button>
            <button
              @click="showCreateUserForm = true"
              class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <PlusIcon class="w-4 h-4 mr-2" />
              Добавить пользователя
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
              <p class="text-sm font-medium text-gray-500">Всего пользователей</p>
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
              <p class="text-sm font-medium text-gray-500">Активные пользователи</p>
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
              <p class="text-sm font-medium text-gray-500">Администраторы</p>
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
              <p class="text-sm font-medium text-gray-500">Всего климатических камер</p>
              <p class="text-3xl font-bold text-gray-900">{{ chamberStore.chambers.length }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Chamber Access Overview -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-900">Обзор доступа к климатическим камерам</h2>
          <button
            @click="showBulkAssignModal = true"
            class="inline-flex items-center px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors text-sm"
          >
            <Cog6ToothIcon class="w-4 h-4 mr-2" />
            Массовое назначение
          </button>
        </div>
        
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Chamber Assignment Summary -->
          <div>
            <h3 class="text-sm font-medium text-gray-700 mb-3">Статистика назначений</h3>
            <div class="space-y-2">
              <div v-for="chamber in chamberStore.chambers" :key="chamber.id" class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div class="flex items-center">
                  <div :class="[
                    'w-3 h-3 rounded-full mr-3',
                    chamber.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                  ]"></div>
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                    <p class="text-xs text-gray-500">{{ chamber.location || 'Неизвестное местоположение' }}</p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="text-sm font-medium text-gray-900">{{ getUsersForChamber(chamber.id).length }} пользователей</p>
                  <p class="text-xs text-gray-500">{{ chamber.status }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Quick Access Matrix -->
          <div>
            <h3 class="text-sm font-medium text-gray-700 mb-3">Быстрый просмотр доступа</h3>
            <div class="overflow-x-auto">
              <table class="min-w-full">
                <thead>
                  <tr>
                    <th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider pb-2">Пользователь</th>
                    <th v-for="chamber in chamberStore.chambers.slice(0, 3)" :key="chamber.id" 
                        class="text-center text-xs font-medium text-gray-500 uppercase tracking-wider pb-2">
                      {{ chamber.name.substring(0, 8) }}...
                    </th>
                  </tr>
                </thead>
                <tbody class="space-y-1">
                  <tr v-for="userWithAccess in users.slice(0, 5)" :key="userWithAccess.user.id" class="border-b border-gray-100">
                    <td class="py-2 text-sm text-gray-900">{{ userWithAccess.user.username }}</td>
                    <td v-for="chamber in chamberStore.chambers.slice(0, 3)" :key="chamber.id" class="py-2 text-center">
                      <div :class="[
                        'w-3 h-3 rounded-full mx-auto',
                        userWithAccess.chambers.some(c => c.id === chamber.id) ? 'bg-green-500' : 'bg-gray-300'
                      ]"></div>
                    </td>
                  </tr>
                </tbody>
              </table>
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
                placeholder="Поиск пользователей..."
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
              <option value="">Все роли</option>
              <option value="admin">Администратор</option>
              <option value="user">Пользователь</option>
            </select>
          </div>

          <!-- Status Filter -->
          <div class="sm:w-48">
            <select
              v-model="statusFilter"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Все статусы</option>
              <option value="true">Активный</option>
              <option value="false">Неактивный</option>
            </select>
          </div>

          <!-- Chamber Filter -->
          <div class="sm:w-48">
            <select
              v-model="chamberFilter"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Все климатические камеры</option>
              <option v-for="chamber in chamberStore.chambers" :key="chamber.id" :value="chamber.id">
                {{ chamber.name }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading && users.length === 0" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-gray-600">Загрузка пользователей...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="text-center py-12">
        <ExclamationCircleIcon class="w-16 h-16 text-red-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Ошибка загрузки пользователей</h3>
        <p class="text-gray-500">{{ error }}</p>
        <button
          @click="refreshData"
          class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Попробовать снова
        </button>
      </div>

      <!-- Users Table -->
      <div v-else class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Пользователь
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Роль
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Статус
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Доступ к климатической камере
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Последний вход
                </th>
                <th scope="col" class="relative px-6 py-3">
                  <span class="sr-only">Действия</span>
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
                          {{ userWithAccess.user.username.charAt(0).toUpperCase() }}
                        </span>
                      </div>
                    </div>
                    <div class="ml-4">
                      <div class="text-sm font-medium text-gray-900">{{ userWithAccess.user.username }}</div>
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
                <td class="px-6 py-4">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="flex flex-wrap gap-1">
                        <span v-if="userWithAccess.chambers.length === 0" class="text-sm text-gray-500 italic">
                          Нет доступа
                        </span>
                        <span 
                          v-else
                          v-for="chamber in userWithAccess.chambers.slice(0, 3)" 
                          :key="chamber.id"
                          :class="[
                            'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium',
                            chamber.status === 'online' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                          ]"
                        >
                          {{ chamber.name }}
                        </span>
                        <span 
                          v-if="userWithAccess.chambers.length > 3"
                          class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                        >
                          +{{ userWithAccess.chambers.length - 3 }} больше
                        </span>
                      </div>
                      <p class="text-xs text-gray-500 mt-1">{{ userWithAccess.chambers.length }} климатических камер всего</p>
                    </div>
                    <button
                      @click="manageChamberAccess(userWithAccess)"
                      class="ml-2 inline-flex items-center px-3 py-1 bg-blue-100 text-blue-800 rounded-md hover:bg-blue-200 transition-colors text-sm font-medium"
                    >
                      <Cog6ToothIcon class="w-4 h-4 mr-1" />
                      Управление
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
                      Редактировать
                    </button>
                    <button
                      v-if="userWithAccess.user.is_active"
                      @click="toggleUserStatus(userWithAccess.user, false)"
                      class="text-yellow-600 hover:text-yellow-900"
                    >
                      Деактивировать
                    </button>
                    <button
                      v-else
                      @click="toggleUserStatus(userWithAccess.user, true)"
                      class="text-green-600 hover:text-green-900"
                    >
                      Активировать
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
          <h3 class="text-lg font-medium text-gray-900 mb-2">Пользователи не найдены</h3>
          <p class="text-gray-500">{{ users.length === 0 ? 'В системе пока нет пользователей.' : 'Пользователи не соответствуют вашим текущим фильтрам.' }}</p>
        </div>
      </div>
    </main>

    <!-- Create User Modal -->
    <div v-if="showCreateUserForm" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">Создать нового пользователя</h2>
            <button
              @click="closeCreateUserForm"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <form @submit.prevent="createUser" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Username</label>
              <input
                v-model="newUserForm.username"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Полное имя"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
              <input
                v-model="newUserForm.password"
                type="password"
                required
                minlength="6"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Минимум 6 символов"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Роль</label>
              <select
                v-model="newUserForm.role"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="user">Пользователь</option>
                <option value="admin">Администратор</option>
              </select>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                @click="closeCreateUserForm"
                class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                type="submit"
                :disabled="creatingUser"
                class="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
              >
                {{ creatingUser ? 'Создание...' : 'Создать пользователя' }}
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
            <h2 class="text-xl font-semibold text-gray-900">Редактировать пользователя</h2>
            <button
              @click="closeEditUserForm"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <form @submit.prevent="updateUser" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Имя</label>
              <input
                v-model="editUserForm.username"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Роль</label>
              <select
                v-model="editUserForm.role"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="user">Пользователь</option>
                <option value="admin">Администратор</option>
              </select>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                @click="closeEditUserForm"
                class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                type="submit"
                :disabled="updatingUser"
                class="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
              >
                {{ updatingUser ? 'Обновление...' : 'Обновить пользователя' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Chamber Access Modal -->
    <div v-if="managingAccessUser" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">
              Управление доступом к климатической камере - {{ managingAccessUser.user.username }}
            </h2>
            <button
              @click="closeChamberAccessModal"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <!-- Current Access Summary -->
          <div class="mb-6 p-4 bg-blue-50 rounded-lg">
            <h3 class="text-sm font-medium text-blue-900 mb-2">Текущий доступ</h3>
            <div v-if="managingAccessUser.chambers.length === 0" class="text-sm text-blue-700">
              У этого пользователя нет доступа к климатическим камерам
            </div>
            <div v-else class="flex flex-wrap gap-2">
              <span 
                v-for="chamber in managingAccessUser.chambers" 
                :key="chamber.id"
                class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800"
              >
                {{ chamber.name }}
                <span :class="[
                  'ml-2 w-2 h-2 rounded-full',
                  chamber.status === 'online' ? 'bg-green-500' : 'bg-gray-400'
                ]"></span>
              </span>
            </div>
          </div>

          <!-- Loading State -->
          <div v-if="chamberStore.loading" class="text-center py-8">
            <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            <p class="mt-2 text-sm text-gray-600">Загрузка климатических камер...</p>
          </div>

          <!-- Chamber Selection -->
          <div v-else class="space-y-4">
            <div class="border border-gray-200 rounded-lg">
              <div class="p-4 bg-gray-50 border-b border-gray-200">
                <div class="flex items-center justify-between">
                  <h3 class="text-lg font-medium text-gray-900">Доступные климатические камеры</h3>
                  <div class="flex space-x-2">
                    <button
                      @click="selectAllChambers"
                      class="px-3 py-1 text-sm bg-green-100 text-green-700 rounded-md hover:bg-green-200 transition-colors"
                    >
                      Выбрать все
                    </button>
                    <button
                      @click="clearAllChambers"
                      class="px-3 py-1 text-sm bg-red-100 text-red-700 rounded-md hover:bg-red-200 transition-colors"
                    >
                      Очистить все
                    </button>
                  </div>
                </div>
              </div>
              
              <div class="p-4">
                <div v-if="chamberStore.chambers.length === 0" class="text-sm text-gray-500 italic text-center py-8">
                  Нет доступных климатических камер
                </div>
                <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div
                    v-for="chamber in chamberStore.chambers"
                    :key="chamber.id"
                    class="relative"
                  >
                    <label 
                      :for="`chamber-${chamber.id}`"
                      :class="[
                        'flex items-center p-4 border-2 rounded-lg cursor-pointer transition-colors',
                        selectedChamberIds.includes(chamber.id) 
                          ? 'border-blue-500 bg-blue-50' 
                          : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                      ]"
                    >
                      <input
                        :id="`chamber-${chamber.id}`"
                        type="checkbox"
                        :value="chamber.id"
                        v-model="selectedChamberIds"
                        class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                      <div class="ml-3 flex-1">
                        <div class="flex items-center justify-between">
                          <div>
                            <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                            <p class="text-xs text-gray-500">{{ chamber.location || 'Неизвестное местоположение' }}</p>
                            <p class="text-xs text-gray-400 mt-1">{{ formatUrl(chamber.ha_url) }}</p>
                          </div>
                          <div class="text-right">
                            <div :class="[
                              'px-2 py-1 text-xs font-medium rounded-full',
                              chamber.status === 'online' 
                                ? 'bg-green-100 text-green-800' 
                                : 'bg-gray-100 text-gray-800'
                            ]">
                              {{ chamber.status }}
                            </div>
                            <p class="text-xs text-gray-500 mt-1">
                              {{ getUsersForChamber(chamber.id).length }} пользователей
                            </p>
                          </div>
                        </div>
                      </div>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Selection Summary -->
          <div class="mt-6 p-4 bg-gray-50 rounded-lg">
            <h3 class="text-sm font-medium text-gray-700 mb-2">Итоги выбора</h3>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-600">
                Выбрано {{ selectedChamberIds.length }} из {{ chamberStore.chambers.length }} климатических камер
              </span>
              <div v-if="selectedChamberIds.length !== managingAccessUser.chambers.length || 
                         !managingAccessUser.chambers.every(c => selectedChamberIds.includes(c.id))" 
                   class="text-sm text-orange-600 font-medium">
                Несохраненные изменения
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex justify-end space-x-3 mt-6 pt-6 border-t border-gray-200">
            <button
              @click="closeChamberAccessModal"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            >
              Отмена
            </button>
            <button
              @click="saveChamberAccess"
              :disabled="savingAccess"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              <span v-if="savingAccess">Сохранение...</span>
              <span v-else>Сохранить изменения</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bulk Assign Modal -->
    <div v-if="showBulkAssignModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">Массовое назначение климатических камер</h2>
            <button
              @click="showBulkAssignModal = false"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- User Selection -->
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-4">Выберите пользователей</h3>
              <div class="border border-gray-200 rounded-lg max-h-80 overflow-y-auto">
                <div v-for="userWithAccess in users" :key="userWithAccess.user.id" class="p-3 border-b border-gray-100">
                  <label class="flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      :value="userWithAccess.user.id"
                      v-model="bulkSelectedUsers"
                      class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <div class="ml-3">
                      <p class="text-sm font-medium text-gray-900">{{ userWithAccess.user.username }}</p>
                      <p class="text-xs text-gray-500">{{ userWithAccess.chambers.length }} климатических камер</p>
                    </div>
                  </label>
                </div>
              </div>
            </div>

            <!-- Chamber Selection -->
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-4">Выберите климатические камеры</h3>
              <div class="border border-gray-200 rounded-lg max-h-80 overflow-y-auto">
                <div v-for="chamber in chamberStore.chambers" :key="chamber.id" class="p-3 border-b border-gray-100">
                  <label class="flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      :value="chamber.id"
                      v-model="bulkSelectedChambers"
                      class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <div class="ml-3 flex-1">
                      <div class="flex items-center justify-between">
                        <div>
                          <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                          <p class="text-xs text-gray-500">{{ chamber.location || 'Неизвестное местоположение' }}</p>
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
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Bulk Action Type -->
          <div class="mt-6">
            <h3 class="text-lg font-medium text-gray-900 mb-4">Действие</h3>
            <div class="space-y-2">
              <label class="flex items-center">
                <input
                  type="radio"
                  value="grant"
                  v-model="bulkActionType"
                  class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <span class="ml-2 text-sm text-gray-900">Предоставить доступ к выбранным климатическим камерам</span>
              </label>
              <label class="flex items-center">
                <input
                  type="radio"
                  value="revoke"
                  v-model="bulkActionType"
                  class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <span class="ml-2 text-sm text-gray-900">Отозвать доступ к выбранным климатическим камерам</span>
              </label>
              <label class="flex items-center">
                <input
                  type="radio"
                  value="replace"
                  v-model="bulkActionType"
                  class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <span class="ml-2 text-sm text-gray-900">Заменить весь доступ выбранными климатическими камерами</span>
              </label>
            </div>
          </div>

          <!-- Summary -->
          <div class="mt-6 p-4 bg-gray-50 rounded-lg">
            <h3 class="text-sm font-medium text-gray-700 mb-2">Итоги</h3>
            <p class="text-sm text-gray-600">
              {{ bulkActionType === 'grant' ? 'Предоставить доступ к' : 
                 bulkActionType === 'revoke' ? 'Отозвать доступ к' : 'Заменить доступ на' }}
              <strong>{{ bulkSelectedChambers.length }} климатических камер</strong>
              для <strong>{{ bulkSelectedUsers.length }} пользователей</strong>
            </p>
          </div>

          <!-- Actions -->
          <div class="flex justify-end space-x-3 mt-6 pt-6 border-t border-gray-200">
            <button
              @click="showBulkAssignModal = false"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            >
              Отмена
            </button>
            <button
              @click="executeBulkAssignment"
              :disabled="bulkSelectedUsers.length === 0 || bulkSelectedChambers.length === 0 || !bulkActionType"
              class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 transition-colors"
            >
              Выполнить массовое назначение
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
  ArrowLeftIcon,
  Cog6ToothIcon
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
const chamberFilter = ref('')

// User Creation
const showCreateUserForm = ref(false)
const creatingUser = ref(false)
const newUserForm = reactive({
  username: '',
  password: '',
  role: 'user' as 'user' | 'admin'
})

// User Editing
const editingUser = ref<User | null>(null)
const updatingUser = ref(false)
const editUserForm = reactive({
  username: '',
  role: 'user' as 'user' | 'admin'
})

// Chamber Access Management
const managingAccessUser = ref<UserWithChamberAccess | null>(null)
const selectedChamberIds = ref<string[]>([])
const savingAccess = ref(false)

// Bulk Assignment
const showBulkAssignModal = ref(false)
const bulkSelectedUsers = ref<string[]>([])
const bulkSelectedChambers = ref<string[]>([])
const bulkActionType = ref<'grant' | 'revoke' | 'replace'>('grant')

// Computed
const filteredUsers = computed(() => {
  let filtered = users.value

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(userWithAccess => 
      userWithAccess.user.username.toLowerCase().includes(query)
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

  // Chamber filter
  if (chamberFilter.value) {
    filtered = filtered.filter(userWithAccess => 
      userWithAccess.chambers.some(chamber => chamber.id === chamberFilter.value)
    )
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

function formatUrl(url: string): string {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname + (urlObj.port ? ':' + urlObj.port : '')
  } catch {
    return url
  }
}

function getUsersForChamber(chamberId: string): UserWithChamberAccess[] {
  return users.value.filter(userWithAccess => 
    userWithAccess.chambers.some(chamber => chamber.id === chamberId)
  )
}

// User Creation
function closeCreateUserForm() {
  showCreateUserForm.value = false
  newUserForm.username = ''
  newUserForm.password = ''
  newUserForm.role = 'user'
}

async function createUser() {
  creatingUser.value = true
  try {
    const response = await api.api.post('/users', {
      username: newUserForm.username,
      password: newUserForm.password,
      role: newUserForm.role
    })
    
    if (response.data.success) {
      toastStore.success('Пользователь создан', `Пользователь ${newUserForm.username} успешно создан`)
      closeCreateUserForm()
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Ошибка', api.formatError(err))
  } finally {
    creatingUser.value = false
  }
}

// User Editing
function editUser(user: User) {
  editingUser.value = user
  editUserForm.username = user.username
  editUserForm.role = user.role
}

function closeEditUserForm() {
  editingUser.value = null
  editUserForm.username = ''
  editUserForm.role = 'user'
}

async function updateUser() {
  if (!editingUser.value) return
  
  updatingUser.value = true
  try {
    const response = await api.api.put(`/users/${editingUser.value.id}`, {
      username: editUserForm.username,
      role: editUserForm.role
    })
    
    if (response.data.success) {
      toastStore.success('Пользователь обновлен', `Пользователь ${editUserForm.username} успешно обновлен`)
      closeEditUserForm()
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Ошибка', api.formatError(err))
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
        `Пользователь ${isActive ? 'активирован' : 'деактивирован'}`, 
        `Пользователь ${user.username} успешно ${action}н`
      )
      await refreshData()
    }
  } catch (err: any) {
    toastStore.error('Ошибка', api.formatError(err))
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

function selectAllChambers() {
  selectedChamberIds.value = chamberStore.chambers.map(c => c.id)
}

function clearAllChambers() {
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
    
    toastStore.success('Доступ обновлен', `Доступ к климатической камере обновлен для ${managingAccessUser.value.user.username}`)
    closeChamberAccessModal()
    await refreshData()
  } catch (err) {
    toastStore.error('Ошибка', 'Не удалось обновить доступ к климатической камере')
  } finally {
    savingAccess.value = false
  }
}

// Bulk Assignment
async function executeBulkAssignment() {
  if (bulkSelectedUsers.value.length === 0 || bulkSelectedChambers.value.length === 0) return

  try {
    const promises = bulkSelectedUsers.value.map(async (userId) => {
      if (bulkActionType.value === 'replace') {
        // Replace all access with selected chambers
        return userChamberAccessStore.setUserChamberAccess(userId, bulkSelectedChambers.value)
      } else if (bulkActionType.value === 'grant') {
        // Grant access to selected chambers
        const currentUser = users.value.find(u => u.user.id === userId)
        if (currentUser) {
          const currentChamberIds = currentUser.chambers.map(c => c.id)
          const newChamberIds = [...new Set([...currentChamberIds, ...bulkSelectedChambers.value])]
          return userChamberAccessStore.setUserChamberAccess(userId, newChamberIds)
        }
      } else if (bulkActionType.value === 'revoke') {
        // Revoke access to selected chambers
        const currentUser = users.value.find(u => u.user.id === userId)
        if (currentUser) {
          const currentChamberIds = currentUser.chambers.map(c => c.id)
          const newChamberIds = currentChamberIds.filter(id => !bulkSelectedChambers.value.includes(id))
          return userChamberAccessStore.setUserChamberAccess(userId, newChamberIds)
        }
      }
    })

    await Promise.all(promises)
    
    toastStore.success(
      'Массовое назначение выполнено', 
      `Обновлен доступ для ${bulkSelectedUsers.value.length} пользователей`
    )
    
    showBulkAssignModal.value = false
    bulkSelectedUsers.value = []
    bulkSelectedChambers.value = []
    bulkActionType.value = 'grant'
    await refreshData()
  } catch (err) {
    toastStore.error('Ошибка', 'Не удалось выполнить массовое назначение')
  }
}

// Initialize
onMounted(async () => {
  await refreshData()
})
</script>