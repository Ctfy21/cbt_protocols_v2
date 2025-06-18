<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <AppHeader />

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Welcome Section -->
      <div class="bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg shadow-lg p-8 text-white mb-8">
        <div class="max-w-3xl">
          <h1 class="text-3xl font-bold mb-4">
            Добро пожаловать{{ authStore.user?.username ? `, ${authStore.user.username}` : '' }}!
          </h1>
          <p class="text-lg text-blue-100 mb-6">
            {{ authStore.isAdmin 
              ? 'Управляйте климатическими камерами и пользователями в вашей лаборатории' 
              : 'Создавайте и управляйте экспериментами в ваших климатических камерах' 
            }}
          </p>
          <div class="flex flex-wrap gap-3">
            <router-link
              v-if="!authStore.isAdmin && chamberStore.chambers.length > 0"
              to="/chambers"
              class="inline-flex items-center px-6 py-3 bg-white text-blue-600 rounded-lg hover:bg-blue-50 transition-colors font-medium"
            >
              <HomeIcon class="w-5 h-5 mr-2" />
              Выбрать климатическую камеру
            </router-link>
            <router-link
              v-if="chamberStore.selectedChamber"
              to="/experiments"
              class="inline-flex items-center px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-400 transition-colors font-medium"
            >
              <BeakerIcon class="w-5 h-5 mr-2" />
              Мои эксперименты
            </router-link>
            <router-link
              v-if="authStore.isAdmin"
              to="/admin/users"
              class="inline-flex items-center px-6 py-3 bg-white text-blue-600 rounded-lg hover:bg-blue-50 transition-colors font-medium"
            >
              <UsersIcon class="w-5 h-5 mr-2" />
              Управление пользователями
            </router-link>
          </div>
        </div>
      </div>

      <!-- Admin Dashboard -->
      <div v-if="authStore.isAdmin">
        <!-- System Overview -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <HomeIcon class="w-8 h-8 text-blue-600" />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-500">Климатические камеры</p>
                <p class="text-3xl font-bold text-gray-900">{{ chamberStore.chambers.length }}</p>
                <p class="text-sm text-gray-600">{{ chamberStore.onlineChambers.length }} онлайн</p>
              </div>
            </div>
          </div>

          <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <UsersIcon class="w-8 h-8 text-green-600" />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-500">Пользователи</p>
                <p class="text-3xl font-bold text-gray-900">{{ userStats.total }}</p>
                <p class="text-sm text-gray-600">{{ userStats.active }} активных</p>
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
                <p class="text-3xl font-bold text-gray-900">{{ experimentStore.activeExperiments.length }}</p>
                <p class="text-sm text-gray-600">Всего {{ experimentStore.experiments.length }}</p>
              </div>
            </div>
          </div>

          <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <ClockIcon class="w-8 h-8 text-orange-600" />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-500">Время системы</p>
                <p class="text-2xl font-bold text-gray-900">{{ currentTime }}</p>
                <p class="text-sm text-gray-600">{{ currentDate }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions for Admin -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Управление системой</h3>
            <div class="space-y-3">
              <router-link
                to="/admin/users"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div class="flex items-center">
                  <UsersIcon class="w-5 h-5 text-gray-600 mr-3" />
                  <span class="font-medium text-gray-900">Управление пользователями</span>
                </div>
                <ArrowRightIcon class="w-5 h-5 text-gray-400" />
              </router-link>
              <router-link
                to="/chambers"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div class="flex items-center">
                  <HomeIcon class="w-5 h-5 text-gray-600 mr-3" />
                  <span class="font-medium text-gray-900">Управление камерами</span>
                </div>
                <ArrowRightIcon class="w-5 h-5 text-gray-400" />
              </router-link>
              <router-link
                to="/api-tokens"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div class="flex items-center">
                  <KeyIcon class="w-5 h-5 text-gray-600 mr-3" />
                  <span class="font-medium text-gray-900">API токены</span>
                </div>
                <ArrowRightIcon class="w-5 h-5 text-gray-400" />
              </router-link>
            </div>
          </div>

          <!-- Recent Activity -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Статус камер</h3>
            <div class="space-y-3 max-h-64 overflow-y-auto">
              <div v-if="chamberStore.chambers.length === 0" class="text-center py-8 text-gray-500">
                Нет зарегистрированных камер
              </div>
              <div
                v-for="chamber in chamberStore.chambers.slice(0, 5)"
                :key="chamber.id"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
              >
                <div class="flex items-center">
                  <div :class="[
                    'w-3 h-3 rounded-full mr-3',
                    chamber.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                  ]"></div>
                  <div>
                    <p class="font-medium text-gray-900">{{ chamber.name }}</p>
                    <p class="text-sm text-gray-500">{{ chamber.location || 'Неизвестное местоположение' }}</p>
                  </div>
                </div>
                <div class="text-right">
                  <p :class="[
                    'text-sm font-medium',
                    chamber.status === 'online' ? 'text-green-600' : 'text-red-600'
                  ]">
                    {{ chamber.status === 'online' ? 'Онлайн' : 'Оффлайн' }}
                  </p>
                  <p class="text-xs text-gray-500">{{ formatRelativeTime(chamber.last_heartbeat) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- User Dashboard -->
      <div v-else>
        <!-- No Chamber Access -->
        <div v-if="chamberStore.chambers.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200 mb-8">
          <LockClosedIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
          <h3 class="text-xl font-medium text-gray-900 mb-4">Добро пожаловать в систему!</h3>
          <p class="text-gray-500 mb-6 max-w-md mx-auto">
            У вас пока нет доступа к климатическим камерам. 
            Обратитесь к администратору для получения доступа к нужным камерам.
          </p>
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-6 max-w-lg mx-auto">
            <h4 class="text-lg font-medium text-blue-900 mb-3">Как получить доступ:</h4>
            <ol class="text-sm text-blue-800 space-y-2 text-left">
              <li class="flex items-start">
                <span class="bg-blue-200 text-blue-800 rounded-full w-6 h-6 flex items-center justify-center mr-3 mt-0.5 text-xs font-bold">1</span>
                Обратитесь к системному администратору
              </li>
              <li class="flex items-start">
                <span class="bg-blue-200 text-blue-800 rounded-full w-6 h-6 flex items-center justify-center mr-3 mt-0.5 text-xs font-bold">2</span>
                Укажите, к каким камерам вам нужен доступ
              </li>
              <li class="flex items-start">
                <span class="bg-blue-200 text-blue-800 rounded-full w-6 h-6 flex items-center justify-center mr-3 mt-0.5 text-xs font-bold">3</span>
                Администратор настроит доступ через панель управления
              </li>
            </ol>
          </div>
        </div>

        <!-- User has chambers -->
        <div v-else>
          <!-- Chamber Overview -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <HomeIcon class="w-8 h-8 text-blue-600" />
                </div>
                <div class="ml-4">
                  <p class="text-sm font-medium text-gray-500">Доступные камеры</p>
                  <p class="text-3xl font-bold text-gray-900">{{ chamberStore.chambers.length }}</p>
                  <p class="text-sm text-gray-600">{{ chamberStore.onlineChambers.length }} онлайн</p>
                </div>
              </div>
            </div>

            <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <BeakerIcon class="w-8 h-8 text-green-600" />
                </div>
                <div class="ml-4">
                  <p class="text-sm font-medium text-gray-500">Мои эксперименты</p>
                  <p class="text-3xl font-bold text-gray-900">{{ experimentStore.experiments.length }}</p>
                  <p class="text-sm text-gray-600">{{ experimentStore.activeExperiments.length }} активных</p>
                </div>
              </div>
            </div>

            <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <div :class="[
                    'w-8 h-8 rounded-full flex items-center justify-center',
                    chamberStore.selectedChamber?.status === 'online' 
                      ? 'bg-green-100 text-green-600' 
                      : 'bg-gray-100 text-gray-600'
                  ]">
                    <CheckCircleIcon v-if="chamberStore.selectedChamber?.status === 'online'" class="w-5 h-5" />
                    <ExclamationCircleIcon v-else class="w-5 h-5" />
                  </div>
                </div>
                <div class="ml-4">
                  <p class="text-sm font-medium text-gray-500">Текущая камера</p>
                  <p class="text-lg font-bold text-gray-900">
                    {{ chamberStore.selectedChamber?.name || 'Не выбрана' }}
                  </p>
                  <p v-if="chamberStore.selectedChamber" class="text-sm text-gray-600">
                    {{ chamberStore.selectedChamber.status === 'online' ? 'Готова к работе' : 'Недоступна' }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Quick Actions for Users -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <!-- Current Chamber -->
            <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h3 class="text-lg font-semibold text-gray-900 mb-4">
                {{ chamberStore.selectedChamber ? 'Текущая камера' : 'Выберите камеру' }}
              </h3>
              
              <div v-if="chamberStore.selectedChamber" class="space-y-4">
                <div class="p-4 bg-gray-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <h4 class="font-medium text-gray-900">{{ chamberStore.selectedChamber.name }}</h4>
                    <div :class="[
                      'px-3 py-1 text-xs font-medium rounded-full',
                      chamberStore.selectedChamber.status === 'online' 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    ]">
                      {{ chamberStore.selectedChamber.status === 'online' ? 'Онлайн' : 'Оффлайн' }}
                    </div>
                  </div>
                  <p class="text-sm text-gray-600">{{ chamberStore.selectedChamber.location || 'Местоположение не указано' }}</p>
                  <div class="grid grid-cols-3 gap-4 mt-3 text-sm">
                    <div class="text-center">
                      <p class="font-medium text-gray-900">{{ chamberStore.selectedChamber.config?.lamps?.length || 0 }}</p>
                      <p class="text-gray-500">Лампы</p>
                    </div>
                    <div class="text-center">
                      <p class="font-medium text-gray-900">{{ chamberStore.selectedChamber.config?.watering_zones?.length || 0 }}</p>
                      <p class="text-gray-500">Полив</p>
                    </div>
                  </div>
                </div>
                <div class="space-y-2">
                  <router-link
                    to="/experiments"
                    class="flex items-center justify-between p-3 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors"
                  >
                    <div class="flex items-center">
                      <BeakerIcon class="w-5 h-5 text-blue-600 mr-3" />
                      <span class="font-medium text-blue-900">Управление экспериментами</span>
                    </div>
                    <ArrowRightIcon class="w-5 h-5 text-blue-600" />
                  </router-link>
                  <router-link
                    to="/chambers"
                    class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <div class="flex items-center">
                      <ArrowsRightLeftIcon class="w-5 h-5 text-gray-600 mr-3" />
                      <span class="font-medium text-gray-900">Сменить камеру</span>
                    </div>
                    <ArrowRightIcon class="w-5 h-5 text-gray-400" />
                  </router-link>
                </div>
              </div>

              <div v-else class="text-center py-8">
                <HomeIcon class="w-12 h-12 text-gray-300 mx-auto mb-4" />
                <p class="text-gray-500 mb-4">Выберите камеру для начала работы</p>
                <router-link
                  to="/chambers"
                  class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                  <HomeIcon class="w-4 h-4 mr-2" />
                  Выбрать камеру
                </router-link>
              </div>
            </div>

            <!-- Active Experiments Progress -->
            <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h3 class="text-lg font-semibold text-gray-900 mb-4">Активные эксперименты</h3>
              
              <div v-if="experimentStore.activeExperiments.length === 0" class="text-center py-8">
                <BeakerIcon class="w-12 h-12 text-gray-300 mx-auto mb-4" />
                <p class="text-gray-500 mb-4">У вас нет активных экспериментов</p>
                <router-link
                  v-if="chamberStore.selectedChamber"
                  to="/experiments"
                  class="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                >
                  <PlusIcon class="w-4 h-4 mr-2" />
                  Создать эксперимент
                </router-link>
              </div>

              <div v-else class="space-y-4 max-h-80 overflow-y-auto">
                <ExperimentProgress
                  v-for="experiment in experimentStore.activeExperiments.slice(0, 3)"
                  :key="experiment.id"
                  :experiment="experiment"
                  class="cursor-pointer"
                  @click="$router.push(`/experiments/${experiment.id}`)"
                />
                <div v-if="experimentStore.activeExperiments.length > 3" class="text-center pt-2">
                  <router-link
                    to="/experiments"
                    class="text-blue-600 hover:text-blue-800 text-sm font-medium"
                  >
                    Посмотреть все активные эксперименты ({{ experimentStore.activeExperiments.length }})
                  </router-link>
                </div>
              </div>
            </div>
          </div>

          <!-- Available Chambers -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-gray-900">Доступные климатические камеры</h3>
              <router-link
                to="/chambers"
                class="text-blue-600 hover:text-blue-800 text-sm font-medium"
              >
                Просмотреть все
              </router-link>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="chamber in chamberStore.chambers.slice(0, 6)"
                :key="chamber.id"
                @click="selectChamber(chamber)"
                :class="[
                  'p-4 border rounded-lg cursor-pointer transition-colors',
                  chamberStore.selectedChamber?.id === chamber.id
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                ]"
              >
                <div class="flex items-center justify-between mb-2">
                  <h4 class="font-medium text-gray-900">{{ chamber.name }}</h4>
                  <div :class="[
                    'w-3 h-3 rounded-full',
                    chamber.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                  ]"></div>
                </div>
                <p class="text-sm text-gray-500 mb-2">{{ chamber.location || 'Местоположение не указано' }}</p>
                <div class="flex justify-between text-xs text-gray-600">
                  <span>{{ chamber.status === 'online' ? 'Онлайн' : 'Оффлайн' }}</span>
                  <span>{{ formatRelativeTime(chamber.last_heartbeat) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { format, formatDistanceToNow } from 'date-fns'
import { ru } from 'date-fns/locale'
import {
  HomeIcon,
  BeakerIcon,
  UsersIcon,
  KeyIcon,
  ClockIcon,
  ArrowRightIcon,
  LockClosedIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  PlusIcon,
  ArrowsRightLeftIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useExperimentStore } from '@/stores/experiment'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import ExperimentProgress from '@/components/ExperimentProgress.vue'
import type { Chamber } from '@/types'

const chamberStore = useChamberStore()
const experimentStore = useExperimentStore()
const authStore = useAuthStore()
const toastStore = useToastStore()

const currentTime = ref('')
const currentDate = ref('')
const userStats = ref({ total: 0, active: 0 })

// Time updates
let timeInterval: NodeJS.Timeout

// const recentExperiments = computed(() => {
//   return experimentStore.experiments
//     .slice()
//     .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
//     .slice(0, 5)
// })

function updateTime() {
  const now = new Date()
  currentTime.value = format(now, 'HH:mm:ss')
  currentDate.value = format(now, 'dd.MM.yyyy')
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
  toastStore.success('Камера выбрана', `Выбрана ${chamber.name}`)
}

async function loadUserStats() {
  if (authStore.isAdmin) {
    try {
      // This would need to be implemented in the API
      // For now, using mock data
      userStats.value = { total: 5, active: 4 }
    } catch (error) {
      console.error('Failed to load user stats:', error)
    }
  }
}

onMounted(async () => {
  // Update time immediately and start interval
  updateTime()
  timeInterval = setInterval(updateTime, 1000)

  // Load data
  await chamberStore.fetchChambers()
  
  if (chamberStore.selectedChamber) {
    await experimentStore.fetchExperiments(chamberStore.selectedChamber.id)
  } else if (chamberStore.chambers.length > 0) {
    // Load experiments for all chambers for stats
    await experimentStore.fetchExperiments()
  }

  await loadUserStats()
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})
</script>