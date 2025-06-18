<template>
  <header class="bg-white shadow-sm border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center py-4">
        <div class="flex items-center space-x-3">
          <div class="flex items-center space-x-2">
            <BeakerIcon class="w-8 h-8 text-blue-600" />
            <div>
              <h1 class="text-xl font-bold text-gray-900">Лаборатория управления микроклиматом</h1>
              <div class="flex items-center gap-2 text-sm text-gray-600">
                <!-- Current Chamber Status -->
                <div v-if="chamberStore.selectedChamber" class="flex items-center gap-1">
                  <div :class="[
                    'w-2 h-2 rounded-full',
                    chamberStore.selectedChamber?.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                  ]"></div>
                  <span class="font-medium">{{ chamberStore.selectedChamber.name }}</span>
                  <span v-if="chamberStore.selectedChamber.location" class="text-gray-400">•</span>
                  <span v-if="chamberStore.selectedChamber.location">{{ chamberStore.selectedChamber.location }}</span>
                </div>
                
                <!-- No Chamber Selected -->
                <div v-else-if="chamberStore.chambers.length > 0" class="flex items-center gap-1">
                  <div class="w-2 h-2 rounded-full bg-gray-400"></div>
                  <span class="text-gray-500">Камера не выбрана</span>
                  <span class="text-gray-400">•</span>
                  <router-link 
                    to="/chambers" 
                    class="text-blue-600 hover:text-blue-800 font-medium"
                  >
                    Выбрать камеру
                  </router-link>
                </div>
                
                <!-- No Access -->
                <div v-else-if="!authStore.isAdmin" class="flex items-center gap-1">
                  <LockClosedIcon class="w-4 h-4 text-gray-400" />
                  <span class="text-gray-500">Нет доступа к камерам</span>
                </div>
                
                <!-- Admin View -->
                <div v-else class="flex items-center gap-1">
                  <CogIcon class="w-4 h-4 text-gray-400" />
                  <span class="text-gray-500">Администрирование</span>
                </div>
              </div>
            </div>
          </div>
          <span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded-full">
            v0.9 альфа
          </span>
        </div>
        
        <div class="flex items-center space-x-4">
          <!-- Chamber Access Info for Regular Users -->
          <div v-if="!authStore.isAdmin && chamberStore.chambers.length > 0" class="hidden lg:flex items-center space-x-2">
            <div class="flex items-center space-x-1">
              <HomeIcon class="w-4 h-4 text-gray-500" />
              <span class="text-sm text-gray-600">
                {{ chamberStore.chambers.length }} {{ pluralize(chamberStore.chambers.length, 'камера', 'камеры', 'камер') }}
              </span>
            </div>
            <div class="flex items-center space-x-1">
              <div class="w-2 h-2 rounded-full bg-green-500"></div>
              <span class="text-sm text-gray-600">{{ chamberStore.onlineChambers.length }} онлайн</span>
            </div>
          </div>

          <!-- Quick Actions -->
          <div class="flex items-center space-x-2">
            <!-- Chamber Selector Dropdown for Users -->
            <div v-if="!authStore.isAdmin && chamberStore.chambers.length > 1" class="relative">
              <button
                @click.stop="showChamberDropdown = !showChamberDropdown"
                class="flex items-center px-3 py-2 text-sm text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                <HomeIcon class="w-4 h-4 mr-2" />
                {{ chamberStore.selectedChamber ? 'Сменить камеру' : 'Выбрать камеру' }}
                <ChevronDownIcon class="w-4 h-4 ml-1" />
              </button>
              
              <!-- Chamber Dropdown -->
              <Transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
              >
                <div
                  v-if="showChamberDropdown"
                  v-click-outside="() => showChamberDropdown = false"
                  class="absolute right-0 mt-2 w-80 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-50"
                >
                  <div class="p-4">
                    <h3 class="text-sm font-medium text-gray-900 mb-3">Доступные камеры</h3>
                    <div class="space-y-2 max-h-60 overflow-y-auto">
                      <button
                        v-for="chamber in chamberStore.chambers"
                        :key="chamber.id"
                        @click="selectChamber(chamber)"
                        :class="[
                          'w-full text-left p-3 rounded-lg border transition-colors',
                          chamberStore.selectedChamber?.id === chamber.id
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                        ]"
                      >
                        <div class="flex items-center justify-between">
                          <div>
                            <p class="text-sm font-medium text-gray-900">{{ chamber.name }}</p>
                            <p class="text-xs text-gray-500">{{ chamber.location || 'Местоположение не указано' }}</p>
                          </div>
                          <div class="flex items-center space-x-2">
                            <div :class="[
                              'w-2 h-2 rounded-full',
                              chamber.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                            ]"></div>
                            <span :class="[
                              'text-xs px-2 py-1 rounded-full',
                              chamber.status === 'online'
                                ? 'bg-green-100 text-green-800'
                                : 'bg-red-100 text-red-800'
                            ]">
                              {{ chamber.status === 'online' ? 'Онлайн' : 'Оффлайн' }}
                            </span>
                          </div>
                        </div>
                      </button>
                    </div>
                    <div class="mt-3 pt-3 border-t border-gray-200">
                      <router-link
                        to="/chambers"
                        @click="showChamberDropdown = false"
                        class="block w-full text-center px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                      >
                        Просмотреть все камеры
                      </router-link>
                    </div>
                  </div>
                </div>
              </Transition>
            </div>

            <!-- Change Chamber Link -->
            <router-link
              v-else-if="!authStore.isAdmin"
              to="/chambers"
              class="text-gray-600 hover:text-gray-900 text-sm font-medium"
            >
              {{ chamberStore.chambers.length === 0 ? 'Получить доступ к камерам' : 'Управление камерами' }}
            </router-link>

            <!-- Admin Chamber Link -->
            <router-link
              v-else
              to="/chambers"
              class="text-gray-600 hover:text-gray-900 text-sm font-medium"
            >
              Управление климатическими камерами
            </router-link>
          </div>
          
          <!-- User Menu -->
          <div class="relative">
            <button
              @click.stop="showUserMenu = !showUserMenu"
              class="flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <div class="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <span class="text-white font-medium text-sm">
                  {{ authStore.user?.username.charAt(0).toUpperCase() }}
                </span>
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
                class="absolute right-0 mt-2 w-64 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-100 z-50"
              >
                <!-- User Info -->
                <div class="px-4 py-3">
                  <p class="text-sm text-gray-500">Вы вошли как</p>
                  <p class="text-sm font-medium text-gray-900 truncate">
                    {{ authStore.user?.username }}
                  </p>
                  <div class="flex items-center mt-1">
                    <span :class="[
                      'px-2 py-1 text-xs font-medium rounded-full',
                      authStore.user?.role === 'admin' 
                        ? 'bg-purple-100 text-purple-800' 
                        : 'bg-blue-100 text-blue-800'
                    ]">
                      {{ authStore.user?.role === 'admin' ? 'Администратор' : 'Пользователь' }}
                    </span>
                  </div>
                </div>
                
                <!-- User Actions -->
                <div class="py-1">
                  <router-link
                    to="/profile"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <UserIcon class="w-4 h-4 mr-3" />
                      Мой профиль
                    </div>
                  </router-link>
                  <router-link
                    to="/experiments"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <BeakerIcon class="w-4 h-4 mr-3" />
                      Мои эксперименты
                    </div>
                  </router-link>
                  <router-link
                    to="/chambers"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <HomeIcon class="w-4 h-4 mr-3" />
                      {{ authStore.isAdmin ? 'Управление камерами' : 'Мои камеры' }}
                    </div>
                  </router-link>
                </div>

                <!-- API Tokens (for all users) -->
                <div class="py-1">
                  <router-link
                    to="/api-tokens"
                    @click="showUserMenu = false"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <KeyIcon class="w-4 h-4 mr-3" />
                      API токены
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
                      Управление пользователями
                    </div>
                  </router-link>
                </div>
                
                <!-- Logout -->
                <div class="py-1">
                  <button
                    @click="handleLogout"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    <div class="flex items-center">
                      <ArrowRightStartOnRectangleIcon class="w-4 h-4 mr-3" />
                      Выйти
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
  UsersIcon,
  HomeIcon,
  LockClosedIcon,
  CogIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import type { Chamber } from '@/types'

const router = useRouter()
const chamberStore = useChamberStore()
const authStore = useAuthStore()
const toastStore = useToastStore()

const showUserMenu = ref(false)
const showChamberDropdown = ref(false)

async function handleLogout() {
  showUserMenu.value = false
  try {
    await authStore.logout()
    toastStore.info('Выход выполнен', 'Вы успешно вышли из системы')
    router.push('/login')
  } catch (error) {
    toastStore.error('Ошибка', 'Не удалось выйти из системы')
  }
}

function selectChamber(chamber: Chamber) {
  chamberStore.selectChamber(chamber)
  showChamberDropdown.value = false
  toastStore.success('Камера выбрана', `Выбрана ${chamber.name}`)
}

function pluralize(count: number, one: string, few: string, many: string): string {
  if (count % 10 === 1 && count % 100 !== 11) {
    return one
  } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
    return few
  } else {
    return many
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