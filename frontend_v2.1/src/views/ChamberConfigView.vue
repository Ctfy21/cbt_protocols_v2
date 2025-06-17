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
        
        <!-- Available Entities Overview -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-6">Доступные сущности</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Assigned Entities -->
            <div>
              <h3 class="text-md font-medium text-gray-700 mb-3">Назначенные сущности</h3>
              <div class="space-y-2 max-h-60 overflow-y-auto">
                <div v-for="entity in assignedEntities" :key="entity.id" 
                     class="flex items-center justify-between p-2 bg-green-50 border border-green-200 rounded-md">
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ entity.friendly_name }}</p>
                    <p class="text-xs text-gray-500">{{ entity.entity_id }}</p>
                  </div>
                  <span class="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                    {{ entity.assigned_to }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Unassigned Entities -->
            <div>
              <h3 class="text-md font-medium text-gray-700 mb-3">Неназначенные сущности</h3>
              <div class="space-y-2 max-h-60 overflow-y-auto">
                <div v-for="entity in unassignedEntities" :key="entity.entity_id" 
                     class="flex items-center justify-between p-2 bg-gray-50 border border-gray-200 rounded-md">
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ entity.friendly_name }}</p>
                    <p class="text-xs text-gray-500">{{ entity.entity_id }}</p>
                  </div>
                  <span class="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                    Доступно
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Climate Control Settings -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-6">Настройки климат-контроля</h2>

          <!-- Day Settings -->
          <div class="mb-8">
            <h3 class="text-md font-medium text-gray-700 mb-4 flex items-center">
              <SunIcon class="w-5 h-5 mr-2 text-yellow-500" />
              Дневной режим
            </h3>
            
            <div class="space-y-6">
              <!-- Day Duration -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Продолжительность дня (часы)</h4>
                  <button
                    @click="addEntityMapping('day_duration')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="Object.keys(editableConfig.day_duration).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.day_duration" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('day_duration', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('day_duration', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.day_duration[entityId]"
                        type="number"
                        min="0"
                        max="24"
                        step="0.5"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">ч</span>
                    <button
                      @click="removeEntityMapping('day_duration', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Day Start -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Начало дня (час)</h4>
                  <button
                    @click="addEntityMapping('day_start')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="Object.keys(editableConfig.day_start).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.day_start" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('day_start', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('day_start', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.day_start[entityId]"
                        type="number"
                        min="0"
                        max="23"
                        step="1"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">:00</span>
                    <button
                      @click="removeEntityMapping('day_start', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Temperature Day -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Температура (°C)</h4>
                  <button
                    @click="addEntityMapping('temperature_day')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.temperature.day || Object.keys(editableConfig.temperature.day).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.temperature.day" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('temperature_day', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('temperature_day', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.temperature.day[entityId]"
                        type="number"
                        min="10"
                        max="40"
                        step="0.5"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">°C</span>
                    <button
                      @click="removeEntityMapping('temperature_day', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Humidity Day -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Влажность (%)</h4>
                  <button
                    @click="addEntityMapping('humidity_day')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.humidity.day || Object.keys(editableConfig.humidity.day).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.humidity.day" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('humidity_day', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('humidity_day', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.humidity.day[entityId]"
                        type="number"
                        min="0"
                        max="100"
                        step="5"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">%</span>
                    <button
                      @click="removeEntityMapping('humidity_day', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- CO2 Day -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">CO2 (ppm)</h4>
                  <button
                    @click="addEntityMapping('co2_day')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.co2.day || Object.keys(editableConfig.co2.day).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.co2.day" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('co2_day', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('co2_day', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.co2.day[entityId]"
                        type="number"
                        min="0"
                        max="5000"
                        step="100"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">ppm</span>
                    <button
                      @click="removeEntityMapping('co2_day', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
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
            
            <div class="space-y-6">
              <!-- Temperature Night -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Температура (°C)</h4>
                  <button
                    @click="addEntityMapping('temperature_night')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.temperature.night || Object.keys(editableConfig.temperature.night).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.temperature.night" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('temperature_night', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('temperature_night', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.temperature.night[entityId]"
                        type="number"
                        min="10"
                        max="40"
                        step="0.5"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">°C</span>
                    <button
                      @click="removeEntityMapping('temperature_night', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Humidity Night -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">Влажность (%)</h4>
                  <button
                    @click="addEntityMapping('humidity_night')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.humidity.night || Object.keys(editableConfig.humidity.night).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.humidity.night" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('humidity_night', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('humidity_night', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.humidity.night[entityId]"
                        type="number"
                        min="0"
                        max="100"
                        step="5"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">%</span>
                    <button
                      @click="removeEntityMapping('humidity_night', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- CO2 Night -->
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-900">CO2 (ppm)</h4>
                  <button
                    @click="addEntityMapping('co2_night')"
                    class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    + Добавить
                  </button>
                </div>
                
                <div v-if="!editableConfig.co2.night || Object.keys(editableConfig.co2.night).length === 0" 
                     class="text-center py-4 text-gray-500 text-sm">
                  Нет назначенных сущностей
                </div>
                
                <div v-else class="space-y-3">
                  <div v-for="(value, entityId) in editableConfig.co2.night" :key="entityId" 
                       class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                    <div class="flex-1">
                      <EntitySelector
                        :model-value="String(entityId)"
                        :available-entities="availableEntitiesForType('co2_night', String(entityId))"
                        @update:model-value="(newEntityId: string) => updateEntityMapping('co2_night', String(entityId), newEntityId)"
                      />
                    </div>
                    <div class="w-32">
                      <input
                        v-model.number="editableConfig.co2.night[entityId]"
                        type="number"
                        min="0"
                        max="5000"
                        step="100"
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <span class="text-sm text-gray-500 w-12">ppm</span>
                    <button
                      @click="removeEntityMapping('co2_night', String(entityId))"
                      class="p-1 text-red-600 hover:text-red-800"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
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

    <!-- Entity Assignment Modal -->
    <div v-if="showEntityModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div class="p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-900">
              Выберите сущность для {{ getTypeDisplayName(pendingAssignment.type) }}
            </h2>
            <button
              @click="closeEntityModal"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>

          <div class="space-y-3 max-h-60 overflow-y-auto">
            <div
              v-for="entity in unassignedEntities"
              :key="entity.entity_id"
              @click="selectEntityForAssignment(entity)"
              class="p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-colors"
            >
              <p class="font-medium text-gray-900">{{ entity.friendly_name }}</p>
              <p class="text-sm text-gray-500">{{ entity.entity_id }}</p>
              <div class="flex items-center mt-1 text-xs text-gray-400">
                <span>Мин: {{ entity.min }}</span>
                <span class="mx-2">•</span>
                <span>Макс: {{ entity.max }}</span>
                <span class="mx-2">•</span>
                <span>{{ entity.unit }}</span>
              </div>
            </div>
          </div>

          <div v-if="unassignedEntities.length === 0" class="text-center py-8 text-gray-500">
            Все сущности уже назначены
          </div>

          <div class="flex justify-end mt-6">
            <button
              @click="closeEntityModal"
              class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      </div>
    </div>
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
  TrashIcon,
  XMarkIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import api from '@/services/api'

// Entity interfaces
interface Entity {
  entity_id: string
  friendly_name: string
  name: string
  min: number
  max: number
  step: number
  value: number
  unit: string
}

// EntitySelector Component
const EntitySelector = {
  props: {
    modelValue: String,
    availableEntities: Array as () => Entity[]
  },
  emits: ['update:modelValue'],
  template: `
    <select 
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    >
      <option v-for="entity in availableEntities" :key="entity.entity_id" :value="entity.entity_id">
        {{ entity.friendly_name }} ({{ entity.entity_id }})
      </option>
    </select>
  `
}

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

// Entity assignment modal
const showEntityModal = ref(false)
const pendingAssignment = ref({ type: '', entityId: '' })

const hasChanges = computed(() => {
  if (!config.value) return false
  return JSON.stringify(config.value) !== JSON.stringify(editableConfig)
})

// Get assigned entities with their types
const assignedEntities = computed(() => {
  const assigned: Array<{id: string, entity_id: string, friendly_name: string, assigned_to: string}> = []
  
  // Day duration
  Object.keys(editableConfig.day_duration).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId,
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Продолжительность дня'
      })
    }
  })
  
  // Day start
  Object.keys(editableConfig.day_start).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId,
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Начало дня'
      })
    }
  })
  
  // Temperature day/night
  Object.keys(editableConfig.temperature.day || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_temp_day',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Температура (день)'
      })
    }
  })
  
  Object.keys(editableConfig.temperature.night || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_temp_night',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Температура (ночь)'
      })
    }
  })
  
  // Humidity day/night
  Object.keys(editableConfig.humidity.day || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_hum_day',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Влажность (день)'
      })
    }
  })
  
  Object.keys(editableConfig.humidity.night || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_hum_night',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'Влажность (ночь)'
      })
    }
  })
  
  // CO2 day/night
  Object.keys(editableConfig.co2.day || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_co2_day',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'CO2 (день)'
      })
    }
  })
  
  Object.keys(editableConfig.co2.night || {}).forEach(entityId => {
    const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
    if (entity) {
      assigned.push({
        id: entityId + '_co2_night',
        entity_id: entityId,
        friendly_name: entity.friendly_name,
        assigned_to: 'CO2 (ночь)'
      })
    }
  })
  
  return assigned
})

// Get unassigned entities
const unassignedEntities = computed((): Entity[] => {
  const assignedEntityIds = new Set(assignedEntities.value.map(a => a.entity_id))
  return editableConfig.unrecognised_entities.filter((entity: Entity) => 
    !assignedEntityIds.has(entity.entity_id)
  )
})

// Get available entities for a specific type (including current assignment)
function availableEntitiesForType(type: string, currentEntityId: string): Entity[] {
  const available = [...unassignedEntities.value]
  
  // Add current entity if it exists
  const currentEntity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === currentEntityId)
  if (currentEntity) {
    available.unshift(currentEntity)
  }
  
  return available
}

async function loadConfig(): Promise<void> {
  loading.value = true
  error.value = null
  
  try {
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

async function saveConfig(): Promise<void> {
  loading.value = true
  
  try {
    const response = await api.updateChamberConfig(chamberId.value, editableConfig)
    if (response.success && response.data) {
      config.value = response.data
      toastStore.success('Конфигурация сохранена', 'Изменения успешно применены')
      await loadConfig()
    }
  } catch (err: any) {
    toastStore.error('Ошибка', err.message || 'Не удалось сохранить конфигурацию')
  } finally {
    loading.value = false
  }
}

function refreshConfig(): void {
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
  const entity = editableConfig.unrecognised_entities.find((e: Entity) => e.entity_id === entityId)
  if (entity) {
    return entity.friendly_name || entity.name || entityId
  }
  
  const lamp = editableConfig.lamps.find((l: any) => l.entity_id === entityId)
  if (lamp) {
    return lamp.friendly_name || lamp.name || entityId
  }
  
  return entityId
}

function getTypeDisplayName(type: string): string {
  const typeNames: Record<string, string> = {
    'day_duration': 'Продолжительность дня',
    'day_start': 'Начало дня',
    'temperature_day': 'Температура (день)',
    'temperature_night': 'Температура (ночь)',
    'humidity_day': 'Влажность (день)',
    'humidity_night': 'Влажность (ночь)',
    'co2_day': 'CO2 (день)',
    'co2_night': 'CO2 (ночь)'
  }
  return typeNames[type] || type
}

function addEntityMapping(type: string): void {
  pendingAssignment.value = { type, entityId: '' }
  showEntityModal.value = true
}

function selectEntityForAssignment(entity: Entity): void {
  const type = pendingAssignment.value.type
  
  // Get default value based on type
  let defaultValue = 0
  switch (type) {
    case 'day_duration':
      defaultValue = 12
      break
    case 'day_start':
      defaultValue = 9
      break
    case 'temperature_day':
      defaultValue = 25
      break
    case 'temperature_night':
      defaultValue = 20
      break
    case 'humidity_day':
      defaultValue = 60
      break
    case 'humidity_night':
      defaultValue = 70
      break
    case 'co2_day':
      defaultValue = 800
      break
    case 'co2_night':
      defaultValue = 400
      break
  }
  
  // Add to appropriate config section
  if (type === 'day_duration') {
    editableConfig.day_duration[entity.entity_id] = defaultValue
  } else if (type === 'day_start') {
    editableConfig.day_start[entity.entity_id] = defaultValue
  } else if (type === 'temperature_day') {
    if (!editableConfig.temperature.day) {
      editableConfig.temperature.day = {}
    }
    editableConfig.temperature.day[entity.entity_id] = defaultValue
  } else if (type === 'temperature_night') {
    if (!editableConfig.temperature.night) {
      editableConfig.temperature.night = {}
    }
    editableConfig.temperature.night[entity.entity_id] = defaultValue
  } else if (type === 'humidity_day') {
    if (!editableConfig.humidity.day) {
      editableConfig.humidity.day = {}
    }
    editableConfig.humidity.day[entity.entity_id] = defaultValue
  } else if (type === 'humidity_night') {
    if (!editableConfig.humidity.night) {
      editableConfig.humidity.night = {}
    }
    editableConfig.humidity.night[entity.entity_id] = defaultValue
  } else if (type === 'co2_day') {
    if (!editableConfig.co2.day) {
      editableConfig.co2.day = {}
    }
    editableConfig.co2.day[entity.entity_id] = defaultValue
  } else if (type === 'co2_night') {
    if (!editableConfig.co2.night) {
      editableConfig.co2.night = {}
    }
    editableConfig.co2.night[entity.entity_id] = defaultValue
  }
  
  closeEntityModal()
}

function closeEntityModal(): void {
  showEntityModal.value = false
  pendingAssignment.value = { type: '', entityId: '' }
}

function updateEntityMapping(type: string, oldEntityId: string, newEntityId: string): void {
  if (oldEntityId === newEntityId) return
  
  // Get the current value
  let currentValue = 0
  
  if (type === 'day_duration') {
    currentValue = editableConfig.day_duration[oldEntityId]
    delete editableConfig.day_duration[oldEntityId]
    editableConfig.day_duration[newEntityId] = currentValue
  } else if (type === 'day_start') {
    currentValue = editableConfig.day_start[oldEntityId]
    delete editableConfig.day_start[oldEntityId]
    editableConfig.day_start[newEntityId] = currentValue
  } else if (type === 'temperature_day') {
    currentValue = editableConfig.temperature.day[oldEntityId]
    delete editableConfig.temperature.day[oldEntityId]
    editableConfig.temperature.day[newEntityId] = currentValue
  } else if (type === 'temperature_night') {
    currentValue = editableConfig.temperature.night[oldEntityId]
    delete editableConfig.temperature.night[oldEntityId]
    editableConfig.temperature.night[newEntityId] = currentValue
  } else if (type === 'humidity_day') {
    currentValue = editableConfig.humidity.day[oldEntityId]
    delete editableConfig.humidity.day[oldEntityId]
    editableConfig.humidity.day[newEntityId] = currentValue
  } else if (type === 'humidity_night') {
    currentValue = editableConfig.humidity.night[oldEntityId]
    delete editableConfig.humidity.night[oldEntityId]
    editableConfig.humidity.night[newEntityId] = currentValue
  } else if (type === 'co2_day') {
    currentValue = editableConfig.co2.day[oldEntityId]
    delete editableConfig.co2.day[oldEntityId]
    editableConfig.co2.day[newEntityId] = currentValue
  } else if (type === 'co2_night') {
    currentValue = editableConfig.co2.night[oldEntityId]
    delete editableConfig.co2.night[oldEntityId]
    editableConfig.co2.night[newEntityId] = currentValue
  }
}

function removeEntityMapping(type: string, entityId: string): void {
  if (type === 'day_duration') {
    delete editableConfig.day_duration[entityId]
  } else if (type === 'day_start') {
    delete editableConfig.day_start[entityId]
  } else if (type === 'temperature_day') {
    delete editableConfig.temperature.day[entityId]
  } else if (type === 'temperature_night') {
    delete editableConfig.temperature.night[entityId]
  } else if (type === 'humidity_day') {
    delete editableConfig.humidity.day[entityId]
  } else if (type === 'humidity_night') {
    delete editableConfig.humidity.night[entityId]
  } else if (type === 'co2_day') {
    delete editableConfig.co2.day[entityId]
  } else if (type === 'co2_night') {
    delete editableConfig.co2.night[entityId]
  }
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