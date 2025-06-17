<template>
    <div class="min-h-screen bg-gray-50">
      <!-- Header -->
      <AppHeader />
  
      <!-- Main Content -->
      <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Page Header -->
        <div class="mb-6">
          <nav class="mb-4">
            <ol class="flex items-center space-x-2 text-sm">
              <li>
                <router-link to="/chambers" class="text-gray-500 hover:text-gray-700">
                  Климатические камеры
                </router-link>
              </li>
              <li>
                <ChevronRightIcon class="w-4 h-4 text-gray-400" />
              </li>
              <li class="text-gray-900 font-medium">Конфигурация камеры</li>
            </ol>
          </nav>
  
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-2xl font-bold text-gray-900">
                Конфигурация: {{ chamber?.name || 'Загрузка...' }}
              </h1>
              <p class="text-gray-600 mt-1">Управление настройками климатической камеры</p>
            </div>
  
            <div class="flex items-center space-x-3">
              <button
                @click="refreshConfig"
                :disabled="loading"
                class="inline-flex items-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
              >
                <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
                Обновить
              </button>
              <button
                @click="saveConfig"
                :disabled="loading || !hasChanges"
                class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                <CheckIcon class="w-4 h-4 mr-2" />
                Сохранить изменения
              </button>
            </div>
          </div>
        </div>
  
        <!-- Loading State -->
        <div v-if="loading && !config" class="text-center py-12">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-2 text-gray-600">Загрузка конфигурации...</p>
        </div>
  
        <!-- Error State -->
        <div v-else-if="error" class="text-center py-12">
          <ExclamationCircleIcon class="w-16 h-16 text-red-400 mx-auto mb-4" />
          <h3 class="text-lg font-medium text-gray-900 mb-2">Ошибка загрузки конфигурации</h3>
          <p class="text-gray-500">{{ error }}</p>
          <button
            @click="refreshConfig"
            class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Попробовать снова
          </button>
        </div>
  
        <!-- Configuration Content -->
        <div v-else-if="config" class="space-y-8">
          <!-- Climate Control Settings -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-6">Настройки климат-контроля</h2>
  
            <!-- Day Settings -->
            <div class="mb-8">
              <h3 class="text-md font-medium text-gray-700 mb-4 flex items-center">
                <SunIcon class="w-5 h-5 mr-2 text-yellow-500" />
                Дневной режим
              </h3>
              
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <!-- Day Duration -->
                <div v-if="Object.keys(editableConfig.day_duration).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Продолжительность дня (часы)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.day_duration" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.day_duration[entityId]"
                      type="number"
                      min="0"
                      max="24"
                      step="0.5"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- Day Start -->
                <div v-if="Object.keys(editableConfig.day_start).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Начало дня (час)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.day_start" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.day_start[entityId]"
                      type="number"
                      min="0"
                      max="23"
                      step="1"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- Temperature Day -->
                <div v-if="editableConfig.temperature.day && Object.keys(editableConfig.temperature.day).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Температура (°C)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.temperature.day" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.temperature.day[entityId]"
                      type="number"
                      min="10"
                      max="40"
                      step="0.5"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- Humidity Day -->
                <div v-if="editableConfig.humidity.day && Object.keys(editableConfig.humidity.day).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Влажность (%)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.humidity.day" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.humidity.day[entityId]"
                      type="number"
                      min="0"
                      max="100"
                      step="5"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- CO2 Day -->
                <div v-if="editableConfig.co2.day && Object.keys(editableConfig.co2.day).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    CO2 (ppm)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.co2.day" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.co2.day[entityId]"
                      type="number"
                      min="0"
                      max="5000"
                      step="100"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
              </div>
            </div>
  
            <!-- Night Settings -->
            <div>
              <h3 class="text-md font-medium text-gray-700 mb-4 flex items-center">
                <MoonIcon class="w-5 h-5 mr-2 text-blue-500" />
                Ночной режим
              </h3>
              
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <!-- Temperature Night -->
                <div v-if="editableConfig.temperature.night && Object.keys(editableConfig.temperature.night).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Температура (°C)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.temperature.night" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.temperature.night[entityId]"
                      type="number"
                      min="10"
                      max="40"
                      step="0.5"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- Humidity Night -->
                <div v-if="editableConfig.humidity.night && Object.keys(editableConfig.humidity.night).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Влажность (%)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.humidity.night" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.humidity.night[entityId]"
                      type="number"
                      min="0"
                      max="100"
                      step="5"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
  
                <!-- CO2 Night -->
                <div v-if="editableConfig.co2.night && Object.keys(editableConfig.co2.night).length > 0">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    CO2 (ppm)
                  </label>
                  <div v-for="(value, entityId) in editableConfig.co2.night" :key="entityId" class="mb-2">
                    <input
                      v-model.number="editableConfig.co2.night[entityId]"
                      type="number"
                      min="0"
                      max="5000"
                      step="100"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <span class="text-xs text-gray-500">{{ getEntityName(String(entityId)) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
  
          <!-- Lamps Configuration -->
          <div v-if="editableConfig.lamps.length > 0" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-6 flex items-center">
              <LightBulbIcon class="w-5 h-5 mr-2 text-yellow-400" />
              Настройки освещения
            </h2>
  
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <div v-for="(lamp, index) in editableConfig.lamps" :key="lamp.entity_id" class="border border-gray-200 rounded-lg p-4">
                <h4 class="font-medium text-gray-900 mb-3">{{ lamp.name || lamp.friendly_name }}</h4>
                
                <div class="space-y-3">
                  <div>
                    <label class="block text-sm text-gray-600 mb-1">Текущее значение</label>
                    <div class="flex items-center">
                      <input
                        v-model.number="editableConfig.lamps[index].current_value"
                        type="range"
                        :min="lamp.intensity_min"
                        :max="lamp.intensity_max"
                        step="5"
                        class="flex-1"
                      />
                      <span class="ml-3 text-sm font-medium text-gray-900 w-12 text-right">
                        {{ editableConfig.lamps[index].current_value }}%
                      </span>
                    </div>
                  </div>
                  
                  <div class="text-xs text-gray-500">
                    <p>Мин: {{ lamp.intensity_min }}%</p>
                    <p>Макс: {{ lamp.intensity_max }}%</p>
                    <p>ID: {{ lamp.entity_id }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
  
          <!-- Watering Zones Configuration -->
          <div v-if="editableConfig.watering_zones.length > 0" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-6 flex items-center">
              <BeakerIcon class="w-5 h-5 mr-2 text-blue-500" />
              Зоны полива
            </h2>
  
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div v-for="zone in editableConfig.watering_zones" :key="zone.name" class="border border-gray-200 rounded-lg p-4">
                <h4 class="font-medium text-gray-900 mb-3">{{ zone.name }}</h4>
                
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-600">Время начала:</span>
                    <span class="font-medium">{{ getEntityName(String(zone.start_time_entity_id)) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">Период:</span>
                    <span class="font-medium">{{ getEntityName(String(zone.period_entity_id)) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">Пауза между циклами:</span>
                    <span class="font-medium">{{ getEntityName(String(zone.pause_between_entity_id)) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">Продолжительность:</span>
                    <span class="font-medium">{{ getEntityName(String(zone.duration_entity_id)) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
  
          <!-- Unrecognized Entities -->
          <div v-if="editableConfig.unrecognised_entities.length > 0" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-6 flex items-center">
              <QuestionMarkCircleIcon class="w-5 h-5 mr-2 text-gray-500" />
              Нераспознанные сущности
            </h2>
  
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="entity in editableConfig.unrecognised_entities" :key="entity.entity_id" class="border border-gray-200 rounded-lg p-3">
                <h4 class="font-medium text-gray-900 text-sm">{{ entity.friendly_name }}</h4>
                <p class="text-xs text-gray-500 mt-1">{{ entity.entity_id }}</p>
                <div class="mt-2 flex items-center">
                  <input
                    :value="entity.value"
                    type="number"
                    :min="entity.min"
                    :max="entity.max"
                    :step="entity.step"
                    disabled
                    class="flex-1 px-2 py-1 text-sm border border-gray-200 rounded bg-gray-50"
                  />
                  <span class="ml-2 text-xs text-gray-500">{{ entity.unit }}</span>
                </div>
              </div>
            </div>
          </div>
  
          <!-- Metadata -->
          <div class="bg-gray-50 rounded-lg p-6 text-sm text-gray-600">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <span class="font-medium">Последнее обновление:</span>
                <span class="ml-2">{{ formatDate(config.updated_at) }}</span>
              </div>
              <div v-if="config.synced_at">
                <span class="font-medium">Синхронизировано:</span>
                <span class="ml-2">{{ formatDate(config.synced_at) }}</span>
              </div>
              <div v-if="hasChanges" class="text-orange-600 font-medium">
                Есть несохраненные изменения
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref, reactive, computed, onMounted, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { format } from 'date-fns'
  import {
    ChevronRightIcon,
    ArrowPathIcon,
    CheckIcon,
    ExclamationCircleIcon,
    SunIcon,
    MoonIcon,
    LightBulbIcon,
    BeakerIcon,
    QuestionMarkCircleIcon
  } from '@heroicons/vue/24/outline'
  import { useChamberStore } from '@/stores/chamber'
  import { useToastStore } from '@/stores/toast'
  import AppHeader from '@/components/AppHeader.vue'
  import api from '@/services/api'
  
  const route = useRoute()
  const chamberStore = useChamberStore()
  const toastStore = useToastStore()
  
  const chamberId = computed(() => route.params.id as string)
  const chamber = computed(() => chamberStore.chambers.find(c => c.id === chamberId.value))
  
  const loading = ref(false)
  const error = ref<string | null>(null)
  const config = ref<any>(null)
  const editableConfig = reactive<any>({
    lamps: [],
    watering_zones: [],
    unrecognised_entities: [],
    day_duration: {},
    day_start: {},
    temperature: { day: {}, night: {} },
    humidity: { day: {}, night: {} },
    co2: { day: {}, night: {} }
  })
  
  const hasChanges = computed(() => {
    if (!config.value) return false
    return JSON.stringify(config.value) !== JSON.stringify(editableConfig)
  })
  
  async function loadConfig() {
    loading.value = true
    error.value = null
    
    try {
      // Load chamber if not already loaded
      if (!chamber.value) {
        await chamberStore.fetchChambers()
      }
      
      const response = await api.getChamberConfig(chamberId.value)
      if (response.success && response.data) {
        config.value = response.data
        
        // Deep copy to editable config
        Object.assign(editableConfig, {
          lamps: [...(response.data.lamps || [])],
          watering_zones: [...(response.data.watering_zones || [])],
          unrecognised_entities: [...(response.data.unrecognised_entities || [])],
          day_duration: { ...(response.data.day_duration || {}) },
          day_start: { ...(response.data.day_start || {}) },
          temperature: {
            day: { ...(response.data.temperature?.day || {}) },
            night: { ...(response.data.temperature?.night || {}) }
          },
          humidity: {
            day: { ...(response.data.humidity?.day || {}) },
            night: { ...(response.data.humidity?.night || {}) }
          },
          co2: {
            day: { ...(response.data.co2?.day || {}) },
            night: { ...(response.data.co2?.night || {}) }
          }
        })
      }
    } catch (err: any) {
      error.value = err.message || 'Не удалось загрузить конфигурацию'
      toastStore.error('Ошибка', error.value || 'Не удалось загрузить конфигурацию')
    } finally {
      loading.value = false
    }
  }
  
  async function saveConfig() {
    loading.value = true
    
    try {
      const response = await api.updateChamberConfig(chamberId.value, editableConfig)
      if (response.success && response.data) {
        config.value = response.data
        toastStore.success('Конфигурация сохранена', 'Изменения успешно применены')
        await loadConfig() // Reload to get updated timestamps
      }
    } catch (err: any) {
      toastStore.error('Ошибка', err.message || 'Не удалось сохранить конфигурацию')
    } finally {
      loading.value = false
    }
  }
  
  function refreshConfig() {
    loadConfig()
  }
  
  function formatDate(dateStr: string | undefined): string {
    if (!dateStr) return 'Неизвестно'
    try {
      return format(new Date(dateStr), 'dd.MM.yyyy HH:mm:ss')
    } catch {
      return 'Неверная дата'
    }
  }
  
  function getEntityName(entityId: string): string {
    // Try to find in unrecognised entities
    const entity = editableConfig.unrecognised_entities.find((e: any) => e.entity_id === entityId)
    if (entity) {
      return entity.friendly_name || entity.name || entityId
    }
    
    // Try to find in lamps
    const lamp = editableConfig.lamps.find((l: any) => l.entity_id === entityId)
    if (lamp) {
      return lamp.friendly_name || lamp.name || entityId
    }
    
    // Return entity ID if not found
    return entityId
  }
  
  // Watch for chamber changes
  watch(chamberId, () => {
    if (chamberId.value) {
      loadConfig()
    }
  })
  
  onMounted(() => {
    if (chamberId.value) {
      loadConfig()
    }
  })
  </script>