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
                  <div v-for="entity in assignedEntitiesList" :key="entity.id" 
                       class="flex items-center justify-between p-2 bg-green-50 border border-green-200 rounded-md">
                    <div>
                      <p class="text-sm font-medium text-gray-900">{{ entity.name }}</p>
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
                  <div v-for="entity in unassignedEntitiesList" :key="entity.entity_id" 
                       class="flex items-center justify-between p-2 bg-gray-50 border border-gray-200 rounded-md">
                    <div>
                      <p class="text-sm font-medium text-gray-900">{{ entity.name }}</p>
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
                  
                  <div v-if="Object.keys(configState.day_duration).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.day_duration" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('day_duration', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('day_duration', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="Object.keys(configState.day_start).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.day_start" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('day_start', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('day_start', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.temperature.day || Object.keys(configState.temperature.day).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.temperature.day" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('temperature_day', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('temperature_day', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.humidity.day || Object.keys(configState.humidity.day).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.humidity.day" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('humidity_day', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('humidity_day', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.co2.day || Object.keys(configState.co2.day).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.co2.day" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('co2_day', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('co2_day', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.temperature.night || Object.keys(configState.temperature.night).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.temperature.night" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('temperature_night', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('temperature_night', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.humidity.night || Object.keys(configState.humidity.night).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.humidity.night" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('humidity_night', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('humidity_night', String(entityId), newEntityId)"
                        />
                      </div>
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
                  
                  <div v-if="!configState.co2.night || Object.keys(configState.co2.night).length === 0" 
                       class="text-center py-4 text-gray-500 text-sm">
                    Нет назначенных сущностей
                  </div>
                  
                  <div v-else class="space-y-3">
                    <div v-for="(value, entityId) in configState.co2.night" :key="entityId" 
                         class="flex items-center gap-3 p-3 bg-gray-50 rounded-md">
                      <div class="flex-1">
                        <EntitySelector
                          :model-value="String(entityId)"
                          :available-entities="availableEntitiesForType('co2_night', String(entityId))"
                          @update:model-value="(newEntityId: string) => updateEntityMapping('co2_night', String(entityId), newEntityId)"
                        />
                      </div>
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
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-semibold text-gray-900 flex items-center">
                <LightBulbIcon class="w-5 h-5 mr-2 text-yellow-400" />
                Настройки освещения
              </h2>
              <button
                @click="addEntityMapping('lamp')"
                class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
              >
                + Добавить лампу
              </button>
            </div>
  
            <div v-if="Object.keys(configState.lamps).length === 0" 
                 class="text-center py-8 text-gray-500">
              Лампы не настроены
            </div>
  
            <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="(lamp, entityId) in configState.lamps" :key="entityId" 
                   class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-start justify-between mb-3">
                  <h4 class="font-medium text-gray-900">{{ lamp.name || getEntityName(entityId.toString()) }}</h4>
                  <button
                    @click="removeEntityMapping('lamp', entityId.toString())"
                    class="p-1 text-red-600 hover:text-red-800"
                  >
                    <TrashIcon class="w-4 h-4" />
                  </button>
                </div>
                
                <div class="space-y-3">
                  <div>
                    <EntitySelector
                      :model-value="entityId.toString()"
                      :available-entities="availableEntitiesForType('lamp', entityId.toString())"
                      @update:model-value="(newEntityId: string) => updateEntityMapping('lamp', entityId.toString(), newEntityId)"
                    />
                  </div>
                  
                  <div class="text-xs text-gray-500">
                    <p>Мин: {{ lamp.min || 0 }}%</p>
                    <p>Макс: {{ lamp.max || 100 }}%</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
  
          <!-- Watering Zones Configuration -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-semibold text-gray-900 flex items-center">
                <BeakerIcon class="w-5 h-5 mr-2 text-blue-500" />
                Зоны полива
              </h2>
              <button
                @click="addWateringZone"
                class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
              >
                + Добавить зону
              </button>
            </div>
  
            <div v-if="configState.watering_zones.length === 0" 
                 class="text-center py-8 text-gray-500">
              Зоны полива не настроены
            </div>
  
            <div v-else class="space-y-4">
              <div v-for="(zone, index) in configState.watering_zones" :key="index" 
                   class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-start justify-between mb-4">
                  <div>
                    <input
                      v-model="zone.name"
                      type="text"
                      placeholder="Название зоны"
                      class="text-lg font-medium text-gray-900 bg-transparent border-0 p-0 focus:ring-0"
                    />
                  </div>
                  <button
                    @click="removeWateringZone(index)"
                    class="p-1 text-red-600 hover:text-red-800"
                  >
                    <TrashIcon class="w-4 h-4" />
                  </button>
                </div>
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <div class="flex items-center justify-between mb-1">
                      <label class="block text-sm font-medium text-gray-700">
                        Entity время начала
                      </label>
                      <button
                        @click="addWateringZoneEntity(index, 'start_time_entity_id')"
                        class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700"
                      >
                        + Добавить
                      </button>
                    </div>
                    <div v-if="!zone.start_time_entity_id || Object.keys(zone.start_time_entity_id).length === 0" 
                         class="text-center py-2 text-gray-500 text-xs border border-dashed border-gray-300 rounded">
                      Нет назначенных сущностей
                    </div>
                    <div v-else class="space-y-2">
                      <div v-for="(value, entityId) in zone.start_time_entity_id" :key="entityId" 
                           class="flex items-center gap-2">
                        <div class="flex-1">
                          <EntitySelector
                            :model-value="entityId"
                            :available-entities="availableEntitiesForWateringZone('start_time', entityId)"
                            @update:model-value="(newId: string) => updateWateringZoneEntity(index, 'start_time_entity_id', newId, entityId)"
                          />
                        </div>
                        <button
                          @click="removeWateringZoneEntity(index, 'start_time_entity_id', entityId)"
                          class="p-1 text-red-600 hover:text-red-800"
                        >
                          <TrashIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <div class="flex items-center justify-between mb-1">
                      <label class="block text-sm font-medium text-gray-700">
                        Entity период
                      </label>
                      <button
                        @click="addWateringZoneEntity(index, 'period_entity_id')"
                        class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700"
                      >
                        + Добавить
                      </button>
                    </div>
                    <div v-if="!zone.period_entity_id || Object.keys(zone.period_entity_id).length === 0" 
                         class="text-center py-2 text-gray-500 text-xs border border-dashed border-gray-300 rounded">
                      Нет назначенных сущностей
                    </div>
                    <div v-else class="space-y-2">
                      <div v-for="(value, entityId) in zone.period_entity_id" :key="entityId" 
                           class="flex items-center gap-2">
                        <div class="flex-1">
                          <EntitySelector
                            :model-value="entityId"
                            :available-entities="availableEntitiesForWateringZone('period', entityId)"
                            @update:model-value="(newId: string) => updateWateringZoneEntity(index, 'period_entity_id', newId, entityId)"
                          />
                        </div>
                        <button
                          @click="removeWateringZoneEntity(index, 'period_entity_id', entityId)"
                          class="p-1 text-red-600 hover:text-red-800"
                        >
                          <TrashIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <div class="flex items-center justify-between mb-1">
                      <label class="block text-sm font-medium text-gray-700">
                        Entity пауза между циклами
                      </label>
                      <button
                        @click="addWateringZoneEntity(index, 'pause_between_entity_id')"
                        class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700"
                      >
                        + Добавить
                      </button>
                    </div>
                    <div v-if="!zone.pause_between_entity_id || Object.keys(zone.pause_between_entity_id).length === 0" 
                         class="text-center py-2 text-gray-500 text-xs border border-dashed border-gray-300 rounded">
                      Нет назначенных сущностей
                    </div>
                    <div v-else class="space-y-2">
                      <div v-for="(value, entityId) in zone.pause_between_entity_id" :key="entityId" 
                           class="flex items-center gap-2">
                        <div class="flex-1">
                          <EntitySelector
                            :model-value="entityId"
                            :available-entities="availableEntitiesForWateringZone('pause_between', entityId)"
                            @update:model-value="(newId: string) => updateWateringZoneEntity(index, 'pause_between_entity_id', newId, entityId)"
                          />
                        </div>
                        <button
                          @click="removeWateringZoneEntity(index, 'pause_between_entity_id', entityId)"
                          class="p-1 text-red-600 hover:text-red-800"
                        >
                          <TrashIcon class="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <div class="flex items-center justify-between mb-1">
                      <label class="block text-sm font-medium text-gray-700">
                        Entity продолжительность
                      </label>
                      <button
                        @click="addWateringZoneEntity(index, 'duration_entity_id')"
                        class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700"
                      >
                        + Добавить
                      </button>
                    </div>
                    <div v-if="!zone.duration_entity_id || Object.keys(zone.duration_entity_id).length === 0" 
                         class="text-center py-2 text-gray-500 text-xs border border-dashed border-gray-300 rounded">
                      Нет назначенных сущностей
                    </div>
                    <div v-else class="space-y-2">
                      <div v-for="(value, entityId) in zone.duration_entity_id" :key="entityId" 
                           class="flex items-center gap-2">
                        <div class="flex-1">
                          <EntitySelector
                            :model-value="entityId"
                            :available-entities="availableEntitiesForWateringZone('duration', entityId)"
                            @update:model-value="(newId: string) => updateWateringZoneEntity(index, 'duration_entity_id', newId, entityId)"
                          />
                        </div>
                        <button
                          @click="removeWateringZoneEntity(index, 'duration_entity_id', entityId)"
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
                v-for="entity in unassignedEntitiesList"
                :key="entity.entity_id"
                @click="selectEntityForAssignment(entity)"
                class="p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-colors"
              >
                <p class="font-medium text-gray-900">{{ entity.name }}</p>
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
  
            <div v-if="unassignedEntitiesList.length === 0" class="text-center py-8 text-gray-500">
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
  import EntitySelector from '@/components/EntitySelector.vue'
  import { InputNumber, WateringZone, ChamberConfig } from '@/types'
  
  const route = useRoute()
  const chamberStore = useChamberStore()
  const toastStore = useToastStore()
  
  const chamberId = computed(() => route.params.id as string)
  const chamber = computed(() => chamberStore.chambers.find(c => c.id === chamberId.value))
  
  const loading = ref(false)
  const error = ref<string | null>(null)
  const config = ref<any>(null)
  
  // Use reactive state for better reactivity tracking
  const configState = reactive({
    lamps: {} as Record<string, InputNumber>,
    watering_zones: [] as WateringZone[],
    unrecognised_entities: {} as Record<string, InputNumber>,
    day_duration: {} as Record<string, InputNumber>,
    day_start: {} as Record<string, InputNumber>,
    temperature: { day: {} as Record<string, InputNumber>, night: {} as Record<string, InputNumber> },
    humidity: { day: {} as Record<string, InputNumber>, night: {} as Record<string, InputNumber> },
    co2: { day: {} as Record<string, InputNumber>, night: {} as Record<string, InputNumber> }
  })
  
  // Entity assignment modal
  const showEntityModal = ref(false)
  const pendingAssignment = ref({ type: '', entityId: '', zoneIndex: -1 })
  
  const hasChanges = computed(() => {
    if (!config.value) return false
    
    // Compare entity assignments only (not values)
    const compareEntityKeys = (obj1: any, obj2: any) => {
      const keys1 = Object.keys(obj1 || {}).sort()
      const keys2 = Object.keys(obj2 || {}).sort()
      return JSON.stringify(keys1) !== JSON.stringify(keys2)
    }
    
    // Check if entity assignments have changed
    if (compareEntityKeys(config.value.day_duration, configState.day_duration)) return true
    if (compareEntityKeys(config.value.day_start, configState.day_start)) return true
    if (compareEntityKeys(config.value.temperature?.day, configState.temperature.day)) return true
    if (compareEntityKeys(config.value.temperature?.night, configState.temperature.night)) return true
    if (compareEntityKeys(config.value.humidity?.day, configState.humidity.day)) return true
    if (compareEntityKeys(config.value.humidity?.night, configState.humidity.night)) return true
    if (compareEntityKeys(config.value.co2?.day, configState.co2.day)) return true
    if (compareEntityKeys(config.value.co2?.night, configState.co2.night)) return true
    
    // Check if lamps have changed
    let originalLampIds: string[] = []
    if (Array.isArray(config.value.lamps)) {
      originalLampIds = config.value.lamps.map((l: any) => l.entity_id).sort()
    } else if (config.value.lamps && typeof config.value.lamps === 'object') {
      originalLampIds = Object.keys(config.value.lamps).sort()
    }
    const currentLampIds = Object.keys(configState.lamps).sort()
    if (JSON.stringify(originalLampIds) !== JSON.stringify(currentLampIds)) return true
    
    // Check if watering zones have changed
    if (JSON.stringify(config.value.watering_zones) !== JSON.stringify(configState.watering_zones)) return true
    
    return false
  })
  
  // Get assigned entities with their types
  const assignedEntitiesList = computed(() => {
    const assigned: Array<{id: string, entity_id: string, name: string, assigned_to: string}> = []
    
    // Helper function to add assigned entity
    const addAssigned = (entityId: string, assignedTo: string) => {
      const entity = configState.unrecognised_entities[entityId]
      if (entity) {
        assigned.push({
          id: `${entityId}_${assignedTo}`,
          entity_id: entityId,
          name: entity.name,
          assigned_to: assignedTo
        })
      }
    }
    
    // Day duration
    Object.keys(configState.day_duration).forEach(entityId => {
      addAssigned(entityId, 'Продолжительность дня')
    })
    
    // Day start
    Object.keys(configState.day_start).forEach(entityId => {
      addAssigned(entityId, 'Начало дня')
    })
    
    // Temperature day/night
    Object.keys(configState.temperature.day || {}).forEach(entityId => {
      addAssigned(entityId, 'Температура (день)')
    })
    
    Object.keys(configState.temperature.night || {}).forEach(entityId => {
      addAssigned(entityId, 'Температура (ночь)')
    })
    
    // Humidity day/night
    Object.keys(configState.humidity.day || {}).forEach(entityId => {
      addAssigned(entityId, 'Влажность (день)')
    })
    
    Object.keys(configState.humidity.night || {}).forEach(entityId => {
      addAssigned(entityId, 'Влажность (ночь)')
    })
    
    // CO2 day/night
    Object.keys(configState.co2.day || {}).forEach(entityId => {
      addAssigned(entityId, 'CO2 (день)')
    })
    
    Object.keys(configState.co2.night || {}).forEach(entityId => {
      addAssigned(entityId, 'CO2 (ночь)')
    })
    
    // Lamps
    Object.keys(configState.lamps || {}).forEach(entityId => {
      addAssigned(entityId, 'Освещение')
    })
    
    // Watering zones entities
    configState.watering_zones.forEach(zone => {
      if (zone.start_time_entity_id) addAssigned(String(Object.keys(zone.start_time_entity_id)[0]), `Полив: ${zone.name} - Время начала`)
      if (zone.period_entity_id) addAssigned(String(Object.keys(zone.period_entity_id)[0]), `Полив: ${zone.name} - Период`)
      if (zone.pause_between_entity_id) addAssigned(String(Object.keys(zone.pause_between_entity_id)[0]), `Полив: ${zone.name} - Пауза`)
      if (zone.duration_entity_id) addAssigned(String(Object.keys(zone.duration_entity_id)[0]), `Полив: ${zone.name} - Продолжительность`)
    })
    
    return assigned
  })
  
  // Get unassigned entities
  const unassignedEntitiesList = computed(() => {
    const assignedEntityIds = new Set(assignedEntitiesList.value.map(a => a.entity_id))
    return Object.values(configState.unrecognised_entities).filter((entity: InputNumber) => 
      !assignedEntityIds.has(entity.entity_id)
    )
  })
  
  // Get available entities for a specific type (including current assignment)
  function availableEntitiesForType(type: string, currentEntityId: string): InputNumber[] {
    const available = [...unassignedEntitiesList.value]
    
    // Add current entity if it exists
    const currentEntity = configState.unrecognised_entities[currentEntityId]
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
        
        // Reset config state and populate with data
        configState.unrecognised_entities = { ...(response.data.unrecognised_entities || {}) }
        configState.watering_zones = [...(response.data.watering_zones || [])]
        
        // Convert lamps array to object
        configState.lamps = {}
        if (response.data.lamps && Array.isArray(response.data.lamps)) {
          response.data.lamps.forEach((lamp: any) => {
            configState.lamps[lamp.entity_id] = lamp
          })
        } else if (response.data.lamps && typeof response.data.lamps === 'object') {
          // Handle case where lamps is already an object
          configState.lamps = { ...response.data.lamps }
        }
        
        // Convert number values to empty objects for UI
        configState.day_duration = Object.keys(response.data.day_duration || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.day_start = Object.keys(response.data.day_start || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.temperature.day = Object.keys(response.data.temperature?.day || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.temperature.night = Object.keys(response.data.temperature?.night || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.humidity.day = Object.keys(response.data.humidity?.day || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.humidity.night = Object.keys(response.data.humidity?.night || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.co2.day = Object.keys(response.data.co2?.day || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
        
        configState.co2.night = Object.keys(response.data.co2?.night || {}).reduce((acc, key) => {
          acc[key] = {}
          return acc
        }, {} as Record<string, any>)
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
      // Prepare config for saving - restore original values for assigned entities
      const configToSave: ChamberConfig = {
        lamps: configState.lamps,
        watering_zones: configState.watering_zones,
        unrecognised_entities: configState.unrecognised_entities,
        day_duration: configState.day_duration,
        day_start: configState.day_start,
        temperature: {
          day: configState.temperature.day,
          night: configState.temperature.night
        },
        humidity: {
          day: configState.humidity.day,
          night: configState.humidity.night
        },
        co2: {
          day: configState.co2.day,
          night: configState.co2.night
        }
      }
      
      const response = await api.updateChamberConfig(chamberId.value, configToSave)
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
    const entity = configState.unrecognised_entities[entityId]
    if (entity) {
      return entity.name || entityId
    }
    
    const lamp = configState.lamps[entityId]
    if (lamp) {
      return lamp.name || entityId
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
      'co2_night': 'CO2 (ночь)',
      'lamp': 'Освещение'
    }
    return typeNames[type] || type
  }
  
  function addEntityMapping(type: string): void {
    pendingAssignment.value = { type, entityId: '', zoneIndex: -1 }
    showEntityModal.value = true
  }
  
  function selectEntityForAssignment(entity: InputNumber): void {
    const type = pendingAssignment.value.type
    
    // Add to appropriate config section (value will be null or empty object)
    if (type === 'day_duration') {
      configState.day_duration[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'day_duration',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'day_start') {
      configState.day_start[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'day_start',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'temperature_day') {
      configState.temperature.day[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'temperature_day',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'temperature_night') {
      configState.temperature.night[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'temperature_night',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'humidity_day') {
      configState.humidity.day[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'humidity_day',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'humidity_night') {
      configState.humidity.night[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'humidity_night',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type === 'lamp') {
      // Create lamp configuration
      configState.lamps[entity.entity_id] = {
        entity_id: entity.entity_id,
        name: entity.name,
        type: 'lamp',
        min: entity.min || 0,
        max: entity.max || 100,
        step: entity.step || 1,
        value: entity.value || 0,
        unit: entity.unit || ''
      }
    } else if (type.startsWith('watering_zone_')) {
      // Handle watering zone entity assignment
      const zoneIndex = pendingAssignment.value.zoneIndex;
      const field = type.replace('watering_zone_', '') as 'start_time_entity_id' | 'period_entity_id' | 'pause_between_entity_id' | 'duration_entity_id';
      
      if (zoneIndex >= 0 && configState.watering_zones[zoneIndex]) {
        if (!configState.watering_zones[zoneIndex][field]) {
          configState.watering_zones[zoneIndex][field] = {};
        }
        
        configState.watering_zones[zoneIndex][field][entity.entity_id] = {
          entity_id: entity.entity_id,
          name: entity.name,
          type: 'watering_zone',
          min: entity.min || 0,
          max: entity.max || 100,
          step: entity.step || 1,
          value: entity.value || 0,
          unit: entity.unit || ''
        };
      }
    }
    
    closeEntityModal()
  }
  
  function closeEntityModal(): void {
    showEntityModal.value = false
    pendingAssignment.value = { type: '', entityId: '', zoneIndex: -1 }
  }
  
  function deleteUnrecognisedEntity(entityId: string): void {
    delete configState.unrecognised_entities[entityId]
    }

  function updateEntityMapping(type: string, oldEntityId: string, newEntityId: string): void {
    if (oldEntityId === newEntityId) return
    
    // Update the mapping by removing old and adding new
    if (type === 'day_duration') {
      delete configState.day_duration[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.day_duration[newEntityId] = configState.day_duration[oldEntityId]
    } else if (type === 'day_start') {
      delete configState.day_start[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.day_start[newEntityId] = configState.day_start[oldEntityId]
    } else if (type === 'temperature_day') {
      delete configState.temperature.day[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.temperature.day[newEntityId] = configState.temperature.day[oldEntityId]
    } else if (type === 'temperature_night') {
      delete configState.temperature.night[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.temperature.night[newEntityId] = configState.temperature.night[oldEntityId]
    } else if (type === 'humidity_day') {
      delete configState.humidity.day[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.humidity.day[newEntityId] = configState.humidity.day[oldEntityId]
    } else if (type === 'humidity_night') {
      delete configState.humidity.night[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.humidity.night[newEntityId] = configState.humidity.night[oldEntityId]
    } else if (type === 'co2_day') {
      delete configState.co2.day[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.co2.day[newEntityId] = configState.co2.day[oldEntityId]
    } else if (type === 'co2_night') {
      delete configState.co2.night[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      configState.co2.night[newEntityId] = configState.co2.night[oldEntityId]
    } else if (type === 'lamp') {
      const lampData = configState.lamps[oldEntityId]
      deleteUnrecognisedEntity(oldEntityId)
      if (lampData) {
        configState.lamps[newEntityId] = {
          ...lampData,
          entity_id: newEntityId,
          name: lampData.name,
          type: 'lamp',
          min: lampData.min,
          max: lampData.max,
          step: lampData.step,
          value: lampData.value,
          unit: lampData.unit
        }
      }
    }
  }
  
  function addUnrecognisedEntity(entity: InputNumber): void {
    configState.unrecognised_entities[entity.entity_id] = {
      entity_id: entity.entity_id,
      name: entity.name,
      min: entity.min,
      max: entity.max,
      step: entity.step,
      value: entity.value,
      unit: entity.unit,
      type: entity.type
    }
  }

  function removeEntityMapping(type: string, entityId: string): void {
    // Simply remove from the appropriate mapping
    if (type === 'day_duration') {
      addUnrecognisedEntity(configState.day_duration[entityId])
      delete configState.day_duration[entityId]
    } else if (type === 'day_start') {
      addUnrecognisedEntity(configState.day_start[entityId])
      delete configState.day_start[entityId]
    } else if (type === 'temperature_day') {
      addUnrecognisedEntity(configState.temperature.day[entityId])
      delete configState.temperature.day[entityId]
    } else if (type === 'temperature_night') {
      addUnrecognisedEntity(configState.temperature.night[entityId])
      delete configState.temperature.night[entityId]
    } else if (type === 'humidity_day') {
      addUnrecognisedEntity(configState.humidity.day[entityId])
      delete configState.humidity.day[entityId]
    } else if (type === 'humidity_night') {
      addUnrecognisedEntity(configState.humidity.night[entityId])
      delete configState.humidity.night[entityId]
    } else if (type === 'co2_day') {
      addUnrecognisedEntity(configState.co2.day[entityId])
      delete configState.co2.day[entityId]
    } else if (type === 'co2_night') {
      addUnrecognisedEntity(configState.co2.night[entityId])
      delete configState.co2.night[entityId]
    } else if (type === 'lamp') {
        const lampData = configState.lamps[entityId]
        if (lampData) {
        addUnrecognisedEntity(lampData)
        delete configState.lamps[entityId]
        }
    }
  }
  
  // Watering zones functions
  function addWateringZone(): void {
    configState.watering_zones.push({
      name: `Зона ${configState.watering_zones.length + 1}`,
      start_time_entity_id: {},
      period_entity_id: {},
      pause_between_entity_id: {},
      duration_entity_id: {}
    })
  }
  
  function removeWateringZone(index: number): void {
    configState.watering_zones.splice(index, 1)
  }
  
  function updateWateringZoneEntity(zoneIndex: number, field: string, newEntityId: string, oldEntityId: string = ''): void {
    type WateringZoneField = 'start_time_entity_id' | 'period_entity_id' | 'pause_between_entity_id' | 'duration_entity_id';

    if (configState.watering_zones[zoneIndex] && (['start_time_entity_id', 'period_entity_id', 'pause_between_entity_id', 'duration_entity_id'] as string[]).includes(field)) {
      // If oldEntityId is provided and different from newEntityId, remove the old entity
      if (oldEntityId && oldEntityId !== newEntityId && configState.watering_zones[zoneIndex][field as WateringZoneField][oldEntityId]) {
        addUnrecognisedEntity(configState.watering_zones[zoneIndex][field as WateringZoneField][oldEntityId])
        delete configState.watering_zones[zoneIndex][field as WateringZoneField][oldEntityId];
      }
      
      // Add the new entity if it's not empty
      deleteUnrecognisedEntity(newEntityId)
      if (newEntityId) {
        configState.watering_zones[zoneIndex][field as WateringZoneField][newEntityId] = {
          entity_id: newEntityId,
          name: getEntityName(newEntityId),
          type: 'watering_zone',
          min: 0,
          max: 0,
          step: 0,
          value: 0,
          unit: ''
        };
      }
    }
  }
  
  function addWateringZoneEntity(zoneIndex: number, field: string): void {
    pendingAssignment.value = { type: `watering_zone_${field}`, entityId: '', zoneIndex: zoneIndex };
    showEntityModal.value = true;
  }
  
  function removeWateringZoneEntity(zoneIndex: number, field: string, entityId: string): void {
    type WateringZoneField = 'start_time_entity_id' | 'period_entity_id' | 'pause_between_entity_id' | 'duration_entity_id';
    
    if (configState.watering_zones[zoneIndex] && 
        configState.watering_zones[zoneIndex][field as WateringZoneField] && 
        configState.watering_zones[zoneIndex][field as WateringZoneField][entityId]) {
      addUnrecognisedEntity(configState.watering_zones[zoneIndex][field as WateringZoneField][entityId])
      delete configState.watering_zones[zoneIndex][field as WateringZoneField][entityId];
    }
  }
  
  function availableEntitiesForWateringZone(type: string, currentEntityId: string): InputNumber[] {
    const available = [...unassignedEntitiesList.value]
    
    // Add current entity if it exists and not empty
    if (currentEntityId) {
      const currentEntity = configState.unrecognised_entities[currentEntityId]
      if (currentEntity) {
        available.unshift(currentEntity)
      }
    }
    
    // Add empty option at the beginning
    available.unshift({
    entity_id: '',
    name: 'Не выбрано',
    type: 'unrecognised',
    min: 0,
    max: 0,
    step: 0,
    value: 0,
    unit: ''
    })
    
    return available
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