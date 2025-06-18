<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <HomeIcon class="w-8 h-8 text-blue-600" />
            <div>
              <h1 class="text-2xl font-bold text-gray-900">
                {{ authStore.isAdmin ? 'Управление климатическими камерами' : 'Мои климатические камеры' }}
              </h1>
              <p class="text-sm text-gray-600">
                {{ authStore.isAdmin 
                  ? 'Просмотр и управление всеми камерами системы' 
                  : 'Выберите климатическую камеру для работы с экспериментами' 
                }}
              </p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3">

            <button
              @click="router.push('/')"
              :disabled="chamberStore.loading"
              class="inline-flex items-center px-3 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
            >
              <HomeIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': chamberStore.loading }" />
              На главную
            </button>

            <!-- Refresh Button -->
            <button
              @click="refreshChambers"
              :disabled="chamberStore.loading"
              class="inline-flex items-center px-3 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
            >
              <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': chamberStore.loading }" />
              Обновить
            </button>

            <!-- Admin Button -->
            <div v-if="authStore.isAdmin" class="flex space-x-3">
              <button
                @click="goToUserAccessManagement"
                class="inline-flex items-center px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
              >
                <UsersIcon class="w-4 h-4 mr-2" />
                Управление доступом пользователей
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Stats for Admins -->
      <div v-if="authStore.isAdmin" class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <HomeIcon class="w-8 h-8 text-blue-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Всего камер</p>
              <p class="text-3xl font-bold text-gray-900">{{ chamberStore.chambers.length }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <CheckCircleIcon class="w-8 h-8 text-green-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Онлайн</p>
              <p class="text-3xl font-bold text-gray-900">{{ chamberStore.onlineChambers.length }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <XCircleIcon class="w-8 h-8 text-red-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Оффлайн</p>
              <p class="text-3xl font-bold text-gray-900">{{ chamberStore.offlineChambers.length }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <BeakerIcon class="w-8 h-8 text-purple-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Активные эксперименты</p>
              <p class="text-3xl font-bold text-gray-900">{{ activeExperimentsCount }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- User Access Info for Regular Users -->
      <div v-if="!authStore.isAdmin && chamberStore.chambers.length > 0" class="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-8">
        <div class="flex items-start">
          <div class="flex-shrink-0">
            <InformationCircleIcon class="w-6 h-6 text-blue-600" />
          </div>
          <div class="ml-3">
            <h3 class="text-lg font-medium text-blue-900">Доступные климатические камеры</h3>
            <div class="mt-2 text-sm text-blue-800">
              <p>У вас есть доступ к <strong>{{ chamberStore.chambers.length }}</strong> климатическим камерам.</p>
              <p class="mt-1">
                Онлайн: <strong>{{ chamberStore.onlineChambers.length }}</strong> | 
                Оффлайн: <strong>{{ chamberStore.offlineChambers.length }}</strong>
              </p>
            </div>
            <div class="mt-3">
              <router-link 
                to="/experiments" 
                class="inline-flex items-center px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors text-sm"
              >
                Перейти к экспериментам
                <ArrowRightIcon class="w-4 h-4 ml-1" />
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="chamberStore.loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-gray-600">Загрузка климатических камер...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="chamberStore.error" class="text-center py-12">
        <ExclamationCircleIcon class="w-16 h-16 text-red-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Ошибка загрузки климатических камер</h3>
        <p class="text-gray-500 mb-4">{{ chamberStore.error }}</p>
        <button
          @click="refreshChambers"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Повторить
        </button>
      </div>

      <!-- Empty State for Users -->
      <div v-else-if="!authStore.isAdmin && chamberStore.chambers.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
        <LockClosedIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Нет доступа к климатическим камерам</h3>
        <p class="text-gray-500 mb-6">
          У вас пока нет доступа к климатическим камерам. 
          Обратитесь к администратору для получения доступа.
        </p>
        <div class="bg-gray-50 rounded-lg p-4 max-w-md mx-auto">
          <h4 class="text-sm font-medium text-gray-900 mb-2">Как получить доступ:</h4>
          <ul class="text-sm text-gray-600 space-y-1">
            <li>• Обратитесь к системному администратору</li>
            <li>• Укажите, к каким камерам вам нужен доступ</li>
            <li>• Администратор настроит доступ в разделе управления пользователями</li>
          </ul>
        </div>
      </div>

      <!-- Empty State for Admins -->
      <div v-else-if="authStore.isAdmin && chamberStore.chambers.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
        <HomeIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Климатические камеры не зарегистрированы</h3>
        <p class="text-gray-500 mb-6">В системе пока нет зарегистрированных климатических камер</p>
        <div class="bg-gray-50 rounded-lg p-4 max-w-md mx-auto">
          <h4 class="text-sm font-medium text-gray-900 mb-2">Регистрация камер:</h4>
          <ul class="text-sm text-gray-600 space-y-1">
            <li>• Камеры регистрируются автоматически при запуске local_api_v2</li>
            <li>• Убедитесь, что local_api_v2 настроен правильно</li>
            <li>• Проверьте сетевое подключение камер</li>
          </ul>
        </div>
      </div>

      <!-- Chambers Grid -->
      <div v-else class="space-y-6">
        <!-- Filter for Admins -->
        <div v-if="authStore.isAdmin && chamberStore.chambers.length > 3" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div class="flex flex-col sm:flex-row gap-4">
            <div class="flex-1">
              <div class="relative">
                <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  v-model="searchQuery"
                  type="text"
                  placeholder="Поиск камер..."
                  class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
            <div class="sm:w-48">
              <select
                v-model="statusFilter"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">Все статусы</option>
                <option value="online">Онлайн</option>
                <option value="offline">Оффлайн</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Chamber Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div
            v-for="chamber in filteredChambers"
            :key="chamber.id"
            @click="selectAndNavigate(chamber)"
            :class="[
              'bg-white rounded-lg shadow-sm border-2 p-6 hover:shadow-md transition-all cursor-pointer',
              chamberStore.selectedChamber?.id === chamber.id 
                ? 'border-blue-500 ring-2 ring-blue-200' 
                : 'border-gray-200 hover:border-gray-300'
            ]"
          >
            <!-- Header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-900 mb-1">{{ chamber.name }}</h3>
                <p class="text-sm text-gray-500">{{ chamber.location || 'Местоположение не указано' }}</p>
              </div>
              <div class="flex flex-col items-end space-y-2">
                <div :class="[
                  'px-3 py-1 text-xs font-medium rounded-full',
                  chamber.status === 'online' 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800'
                ]">
                  {{ chamber.status === 'online' ? 'Онлайн' : 'Оффлайн' }}
                </div>
                <div v-if="chamberStore.selectedChamber?.id === chamber.id" 
                     class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
                  Выбрана
                </div>
              </div>
            </div>
            
            <!-- Details -->
            <div class="space-y-3 text-sm text-gray-600 mb-4">
              <div class="flex items-center">
                <GlobeAltIcon class="w-4 h-4 mr-2 text-gray-400" />
                <span>{{ formatUrl(chamber.ha_url) }}</span>
              </div>
              <div class="flex items-center">
                <ClockIcon class="w-4 h-4 mr-2 text-gray-400" />
                <span>Обновлено: {{ formatDate(chamber.last_heartbeat) }}</span>
              </div>
              <div class="flex items-center">
                <LightBulbIcon class="w-4 h-4 mr-2 text-gray-400" />
                <span>{{ chamber.config?.lamps?.length || 0 }} Нрстроек ламп</span>
              </div>
            </div>

            <!-- Chamber Details -->
            <div v-if="chamber.status === 'online'" class="grid grid-cols-1 gap-4 mb-4 p-3 bg-gray-50 rounded-lg">
              <div class="text-center">
                <p class="text-xs text-gray-500">Последнее обновление</p>
                <p class="text-lg font-semibold text-gray-900">
                  {{ formatRelativeTime(chamber.last_heartbeat) }}
                </p>
              </div>
            </div>

            <!-- Action Button -->
            <div class="mt-4 pt-4 border-t border-gray-100">
              <button
                @click.stop="selectAndNavigate(chamber)"
                :disabled="chamber.status === 'offline'"
                :class="[
                  'w-full px-4 py-2 rounded-md transition-colors text-sm font-medium',
                  chamber.status === 'online'
                    ? 'bg-blue-600 text-white hover:bg-blue-700'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed',
                  chamberStore.selectedChamber?.id === chamber.id
                    ? 'bg-green-600 hover:bg-green-700'
                    : ''
                ]"
              >
                {{ chamberStore.selectedChamber?.id === chamber.id 
                   ? 'Перейти к экспериментам' 
                   : chamber.status === 'online' 
                     ? 'Выбрать камеру' 
                     : 'Камера недоступна' 
                }}
              </button>
            </div>
            <div class="flex gap-2">
              <button
                @click.stop="selectAndNavigate(chamber)"
                :disabled="chamber.status === 'offline'"
                :class="[
                  'flex-1 px-4 py-2 rounded-md transition-colors text-sm font-medium',
                ]"
                >
              </button>
              
              <!-- Configuration button -->
              <router-link
                v-if="authStore.isAdmin && chamber.status === 'online'"
                :to="`/chambers/${chamber.id}/config`"
                @click.stop
                class="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors text-sm font-medium mt-5"
              >
                <Cog6ToothIcon class="w-4 h-4" />
              </router-link>
            </div>
          </div>
        </div>

        <!-- Help Text for Users -->
        <div v-if="!authStore.isAdmin && chamberStore.chambers.length > 0" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-3">Как работать с климатическими камерами</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm text-gray-600">
            <div>
              <h4 class="font-medium text-gray-900 mb-2">1. Выбор камеры</h4>
              <p>Кликните на любую доступную камеру, чтобы выбрать её для работы. Выбранная камера будет выделена синим цветом.</p>
            </div>
            <div>
              <h4 class="font-medium text-gray-900 mb-2">2. Создание экспериментов</h4>
              <p>После выбора камеры перейдите в раздел "Эксперименты" для создания и управления экспериментами.</p>
            </div>
            <div>
              <h4 class="font-medium text-gray-900 mb-2">3. Статус камеры</h4>
              <p>Зеленый статус означает, что камера онлайн и готова к работе. Красный статус означает, что камера недоступна.</p>
            </div>
            <div>
              <h4 class="font-medium text-gray-900 mb-2">4. Обновление данных</h4>
              <p>Используйте кнопку "Обновить" для получения актуальной информации о состоянии камер.</p>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { format, formatDistanceToNow } from 'date-fns'
import { ru } from 'date-fns/locale'
import { 
  HomeIcon, 
  UsersIcon,
  ExclamationCircleIcon,
  GlobeAltIcon,
  ClockIcon,
  LightBulbIcon,
  CheckCircleIcon,
  XCircleIcon,
  BeakerIcon,
  ArrowPathIcon,
  MagnifyingGlassIcon,
  InformationCircleIcon,
  LockClosedIcon,
  ArrowRightIcon,
  Cog6ToothIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useExperimentStore } from '@/stores/experiment'
import { useToastStore } from '@/stores/toast'
import { useAuthStore } from '@/stores/auth'
import type { Chamber } from '@/types'

const router = useRouter()
const chamberStore = useChamberStore()
const experimentStore = useExperimentStore()
const toastStore = useToastStore()
const authStore = useAuthStore()

const searchQuery = ref('')
const statusFilter = ref('')

// Computed
const filteredChambers = computed(() => {
  let filtered = chamberStore.chambers

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(chamber => 
      chamber.name.toLowerCase().includes(query) ||
      chamber.location?.toLowerCase().includes(query)
    )
  }

  // Status filter
  if (statusFilter.value) {
    filtered = filtered.filter(chamber => chamber.status === statusFilter.value)
  }

  return filtered
})

const activeExperimentsCount = computed(() => {
  return experimentStore.activeExperiments.length
})

// Methods
async function refreshChambers() {
  await chamberStore.fetchChambers()
  if (chamberStore.selectedChamber) {
    // Update selected chamber info if it exists in the new list
    const updatedChamber = chamberStore.chambers.find(c => c.id === chamberStore.selectedChamber?.id)
    if (updatedChamber) {
      chamberStore.selectChamber(updatedChamber)
    }
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

function formatDate(dateStr: string): string {
  try {
    return format(new Date(dateStr), 'dd.MM.yyyy HH:mm')
  } catch {
    return 'Неизвестно'
  }
}

function formatRelativeTime(dateStr: string): string {
  try {
    return formatDistanceToNow(new Date(dateStr), { addSuffix: true, locale: ru })
  } catch {
    return 'Неизвестно'
  }
}

function selectChamber(chamber: Chamber) {
  chamberStore.selectChamber(chamber)
  toastStore.success('Климатическая камера выбрана', `Выбрана ${chamber.name}`)
}

function selectAndNavigate(chamber: Chamber) {
  if (chamber.status === 'offline') {
    toastStore.warning('Камера недоступна', `Камера ${chamber.name} в настоящее время оффлайн`)
    return
  }
  
  selectChamber(chamber)
  router.push('/experiments')
}

function goToUserAccessManagement() {
  router.push('/admin/users')
}

// Initialize
onMounted(async () => {
  await refreshChambers()
  
  // Load experiments count if chamber is selected
  if (chamberStore.selectedChamber) {
    await experimentStore.fetchExperiments(chamberStore.selectedChamber.id)
  }
})
</script>