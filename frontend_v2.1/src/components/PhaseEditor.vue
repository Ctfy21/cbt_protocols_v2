<template>
  <div class="border border-gray-200 rounded-lg p-4">
    <div class="flex items-start justify-between mb-4">
      <div class="flex-1">
        <input
          v-model="localPhase.title"
          type="text"
          class="text-xl font-medium text-gray-900 bg-transparent border-0 p-0 focus:ring-0 w-full"
          placeholder="Phase title"
          @input="updatePhase"
        />
        <textarea
          v-model="localPhase.description"
          rows="2"
          class="mt-1 text-lg text-gray-500 bg-transparent border-0 p-0 focus:ring-0 w-full resize-none"
          placeholder="Phase description..."
          @input="updatePhase"
        ></textarea>
      </div>
      <button
        @click="$emit('remove')"
        class="ml-2 text-gray-400 hover:text-red-600"
      >
        <TrashIcon class="w-6 h-6" />
      </button>
    </div>

    <!-- Duration -->
    <div class="mb-4">
      <label class="block text-lg font-medium text-gray-700 mt-4">
        Duration (days)
      </label>
      <input
        v-model.number="localPhase.duration_days"
        type="number"
        min="1"
        class="w-full px-3 py-2 text-lg border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        @input="updatePhase"
      />
      <label class="block text-lg font-medium text-gray-700 mt-4">Start Day Time (HH:MM)</label>
      <input
        v-model="startDayTime"
        type="time"
        class="w-full px-3 py-2 text-lg border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        @input="updateStartDayTime"
      />
    </div>

    <!-- Charts -->
    <div class="space-y-4 mb-4">


      <!-- Day Duration Schedule -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
            <ClockIcon class="w-6 h-6 text-purple-500" />
            <span>Day Duration Schedule</span>
          </div>
          <button
            @click.prevent="toggleScheduleMode('dayDuration')"
            class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
          >
            {{ scheduleMode.dayDuration === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
          </button>
        </div>
        
        <!-- Chart Mode -->
        <Chart
          v-if="scheduleMode.dayDuration === 'chart'"
          v-model="dayDurationSchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="24"
          :step="1"
          :unit="' hours'"
          @update:model-value="updateDayDurationSchedule"
          title="Day Duration Schedule"
        />
        
        <!-- Form Mode -->
        <div v-else class="bg-gray-50 p-4 rounded-lg">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Day Duration (hours) - Applied to all days
          </label>
          <input
            v-model.number="formValues.dayDuration"
            type="number"
            min="0"
            max="24"
            step="1"
            class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @input="updateFormValue('dayDuration', formValues.dayDuration)"
          />
          <span class="ml-2 text-sm text-gray-500">hours</span>
        </div>
      </div>


      <!-- Day Charts -->
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <SunIcon class="w-6 h-6 text-yellow-500" />
          <span>Day Schedule</span>
        </div>
        
        <!-- Temperature Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">Temperature Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('temperatureDay')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.temperatureDay === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.temperatureDay === 'chart'"
            v-model="temperatureDaySchedule" 
            :duration="localPhase.duration_days"
            :min="10"
            :max="30"
            :step="1"
            :unit="'°C'"
            @update:model-value="updateTemperatureDaySchedule"
            title="Temperature Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Temperature (°C) - Applied to all days
            </label>
            <input
              v-model.number="formValues.temperatureDay"
              type="number"
              min="10"
              max="30"
              step="1"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('temperatureDay', formValues.temperatureDay)"
            />
            <span class="ml-2 text-sm text-gray-500">°C</span>
          </div>
        </div>

        <!-- Humidity Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">Humidity Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('humidityDay')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.humidityDay === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.humidityDay === 'chart'"
            v-model="humidityDaySchedule" 
            :duration="localPhase.duration_days"
            :min="0"
            :max="100"
            :step="5"
            :unit="'%'"
            @update:model-value="updateHumidityDaySchedule"
            title="Humidity Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Humidity (%) - Applied to all days
            </label>
            <input
              v-model.number="formValues.humidityDay"
              type="number"
              min="0"
              max="100"
              step="5"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('humidityDay', formValues.humidityDay)"
            />
            <span class="ml-2 text-sm text-gray-500">%</span>
          </div>
        </div>

        <!-- CO2 Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">CO2 Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('co2Day')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.co2Day === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.co2Day === 'chart'"
            v-model="co2DaySchedule" 
            :duration="localPhase.duration_days"
            :min="0"
            :max="2000"
            :step="100"
            :unit="'ppm'"
            @update:model-value="updateCO2DaySchedule"
            title="CO2 Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              CO2 (ppm) - Applied to all days
            </label>
            <input
              v-model.number="formValues.co2Day"
              type="number"
              min="0"
              max="2000"
              step="100"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('co2Day', formValues.co2Day)"
            />
            <span class="ml-2 text-sm text-gray-500">ppm</span>
          </div>
        </div>
      </div>

      <!-- Night Charts -->
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <MoonIcon class="w-6 h-6 text-blue-500" />
          <span>Night Schedule</span>
        </div>

        <!-- Temperature Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">Temperature Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('temperatureNight')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.temperatureNight === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.temperatureNight === 'chart'"
            v-model="temperatureNightSchedule" 
            :duration="localPhase.duration_days"
            :min="10"
            :max="30"
            :step="1"
            :unit="'°C'"
            @update:model-value="updateTemperatureNightSchedule"
            title="Temperature Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Temperature (°C) - Applied to all days
            </label>
            <input
              v-model.number="formValues.temperatureNight"
              type="number"
              min="10"
              max="30"
              step="1"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('temperatureNight', formValues.temperatureNight)"
            />
            <span class="ml-2 text-sm text-gray-500">°C</span>
          </div>
        </div>

        <!-- Humidity Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">Humidity Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('humidityNight')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.humidityNight === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.humidityNight === 'chart'"
            v-model="humidityNightSchedule" 
            :duration="localPhase.duration_days"
            :min="0"
            :max="100"
            :step="5"
            :unit="'%'"
            @update:model-value="updateHumidityNightSchedule"
            title="Humidity Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Humidity (%) - Applied to all days
            </label>
            <input
              v-model.number="formValues.humidityNight"
              type="number"
              min="0"
              max="100"
              step="5"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('humidityNight', formValues.humidityNight)"
            />
            <span class="ml-2 text-sm text-gray-500">%</span>
          </div>
        </div>

        <!-- CO2 Schedule -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">CO2 Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('co2Night')"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ scheduleMode.co2Night === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="scheduleMode.co2Night === 'chart'"
            v-model="co2NightSchedule" 
            :duration="localPhase.duration_days"
            :min="0"
            :max="2000"
            :step="100"
            :unit="'ppm'"
            @update:model-value="updateCO2NightSchedule"
            title="CO2 Schedule"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              CO2 (ppm) - Applied to all days
            </label>
            <input
              v-model.number="formValues.co2Night"
              type="number"
              min="0"
              max="2000"
              step="100"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('co2Night', formValues.co2Night)"
            />
            <span class="ml-2 text-sm text-gray-500">ppm</span>
          </div>
        </div>
      </div>

      <!-- Lamp Intensity Charts -->
      <div v-if="chamber.lamps && chamber.lamps.length > 0" class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <LightBulbIcon class="w-6 h-6 text-yellow-400" />
          <span>Lamp Intensity Schedules</span>
        </div>
        <div v-for="lamp in chamber.lamps" :key="lamp.entity_id" class="space-y-2">
          <div class="flex items-center justify-between">
            <h4 class="text-md font-medium text-gray-700">{{ lamp.name }} Intensity Schedule</h4>
            <button
              @click.prevent="toggleScheduleMode('lampIntensity', lamp.entity_id)"
              class="px-3 py-1 text-sm bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
            >
              {{ (scheduleMode.lampIntensity[lamp.entity_id] || 'chart') === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
            </button>
          </div>
          
          <Chart
            v-if="(scheduleMode.lampIntensity[lamp.entity_id] || 'chart') === 'chart'"
            v-model="lampIntensitySchedules[lamp.entity_id]" 
            :duration="localPhase.duration_days"
            :min="lamp.intensity_min"
            :max="lamp.intensity_max"
            :step="5"
            :unit="'%'"
            @update:model-value="(schedule) => updateLampIntensitySchedule(lamp.entity_id, schedule)"
            :title="`${lamp.name} Intensity Schedule`"
          />
          
          <div v-else class="bg-gray-50 p-4 rounded-lg">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ lamp.name }} Intensity (%) - Applied to all days
            </label>
            <input
              v-model.number="formValues.lampIntensity[lamp.entity_id]"
              type="number"
              :min="lamp.intensity_min"
              :max="lamp.intensity_max"
              step="5"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @input="updateFormValue('lampIntensity', formValues.lampIntensity[lamp.entity_id], lamp.entity_id)"
            />
            <span class="ml-2 text-sm text-gray-500">%</span>
          </div>
        </div>
      </div>

      <!-- Watering Zones -->
      <div v-if="chamber.watering_zones && chamber.watering_zones.length > 0" class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <BeakerIcon class="w-6 h-6 text-blue-500" />
          <span>Watering Zones</span>
        </div>
        
        <div v-for="(zone, index) in chamber.watering_zones" :key="`zone-${index}`" class="border border-gray-200 rounded-lg p-4 space-y-4">
          <h5 class="text-lg font-medium text-gray-900">{{ zone.name }}</h5>
          
          <!-- Start Time Schedule -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <h5 class="text-sm font-medium text-gray-700">Start Time</h5>
              <button
                @click.prevent="toggleScheduleMode('wateringZone', undefined, `zone_${index}`, 'start_time')"
                class="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
              >
                {{ ((scheduleMode.wateringZones[`zone_${index}`]?.start_time) || 'chart') === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
              </button>
            </div>
            
            <Chart
              v-if="((scheduleMode.wateringZones[`zone_${index}`]?.start_time) || 'chart') === 'chart'"
              v-model="wateringZoneSchedules[`zone_${index}`].start_time_schedule"
              :duration="localPhase.duration_days"
              :min="0"
              :max="23"
              :step="1"
              :unit="':00'"
              @update:model-value="(schedule) => updateWateringZoneSchedule(`zone_${index}`, 'start_time', schedule)"
              :title="`${zone.name} - Start Time`"
            />
            
            <div v-else class="bg-gray-50 p-3 rounded-lg">
              <label class="block text-xs font-medium text-gray-700 mb-2">
                Start Time (hour) - Applied to all days
              </label>
              <input
                v-model.number="formValues.wateringZones[`zone_${index}`].start_time"
                type="number"
                min="0"
                max="23"
                step="1"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="updateFormValue('wateringZone', formValues.wateringZones[`zone_${index}`].start_time, undefined, `zone_${index}`, 'start_time')"
              />
              <span class="ml-1 text-xs text-gray-500">:00</span>
            </div>
          </div>
          
          <!-- Period Schedule -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <h5 class="text-sm font-medium text-gray-700">Period Between Watering</h5>
              <button
                @click.prevent="toggleScheduleMode('wateringZone', undefined, `zone_${index}`, 'period')"
                class="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
              >
                {{ ((scheduleMode.wateringZones[`zone_${index}`]?.period) || 'chart') === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
              </button>
            </div>
            
            <Chart
              v-if="((scheduleMode.wateringZones[`zone_${index}`]?.period) || 'chart') === 'chart'"
              v-model="wateringZoneSchedules[`zone_${index}`].period_schedule"
              :duration="localPhase.duration_days"
              :min="1"
              :max="24"
              :step="1"
              :unit="' hours'"
              @update:model-value="(schedule) => updateWateringZoneSchedule(`zone_${index}`, 'period', schedule)"
              :title="`${zone.name} - Period Between Watering`"
            />
            
            <div v-else class="bg-gray-50 p-3 rounded-lg">
              <label class="block text-xs font-medium text-gray-700 mb-2">
                Period (hours) - Applied to all days
              </label>
              <input
                v-model.number="formValues.wateringZones[`zone_${index}`].period"
                type="number"
                min="1"
                max="24"
                step="1"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="updateFormValue('wateringZone', formValues.wateringZones[`zone_${index}`].period, undefined, `zone_${index}`, 'period')"
              />
              <span class="ml-1 text-xs text-gray-500">hours</span>
            </div>
          </div>
          
          <!-- Pause Between Schedule -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <h5 class="text-sm font-medium text-gray-700">Pause Between Cycles</h5>
              <button
                @click.prevent="toggleScheduleMode('wateringZone', undefined, `zone_${index}`, 'pause_between')"
                class="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
              >
                {{ ((scheduleMode.wateringZones[`zone_${index}`]?.pause_between) || 'chart') === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
              </button>
            </div>
            
            <Chart
              v-if="((scheduleMode.wateringZones[`zone_${index}`]?.pause_between) || 'chart') === 'chart'"
              v-model="wateringZoneSchedules[`zone_${index}`].pause_between_schedule"
              :duration="localPhase.duration_days"
              :min="0"
              :max="24"
              :step="1"
              :unit="' hours'"
              @update:model-value="(schedule) => updateWateringZoneSchedule(`zone_${index}`, 'pause_between', schedule)"
              :title="`${zone.name} - Pause Between Cycles`"
            />
            
            <div v-else class="bg-gray-50 p-3 rounded-lg">
              <label class="block text-xs font-medium text-gray-700 mb-2">
                Pause Between (hours) - Applied to all days
              </label>
              <input
                v-model.number="formValues.wateringZones[`zone_${index}`].pause_between"
                type="number"
                min="0"
                max="24"
                step="1"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="updateFormValue('wateringZone', formValues.wateringZones[`zone_${index}`].pause_between, undefined, `zone_${index}`, 'pause_between')"
              />
              <span class="ml-1 text-xs text-gray-500">hours</span>
            </div>
          </div>
          
          <!-- Duration Schedule -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <h5 class="text-sm font-medium text-gray-700">Watering Duration</h5>
              <button
                @click.prevent="toggleScheduleMode('wateringZone', undefined, `zone_${index}`, 'duration')"
                class="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition-colors"
              >
                {{ ((scheduleMode.wateringZones[`zone_${index}`]?.duration) || 'chart') === 'chart' ? 'Switch to Form' : 'Switch to Chart' }}
              </button>
            </div>
            
            <Chart
              v-if="((scheduleMode.wateringZones[`zone_${index}`]?.duration) || 'chart') === 'chart'"
              v-model="wateringZoneSchedules[`zone_${index}`].duration_schedule"
              :duration="localPhase.duration_days"
              :min="0"
              :max="300"
              :step="10"
              :unit="' sec'"
              @update:model-value="(schedule) => updateWateringZoneSchedule(`zone_${index}`, 'duration', schedule)"
              :title="`${zone.name} - Watering Duration`"
            />
            
            <div v-else class="bg-gray-50 p-3 rounded-lg">
              <label class="block text-xs font-medium text-gray-700 mb-2">
                Duration (seconds) - Applied to all days
              </label>
              <input
                v-model.number="formValues.wateringZones[`zone_${index}`].duration"
                type="number"
                min="0"
                max="300"
                step="10"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="updateFormValue('wateringZone', formValues.wateringZones[`zone_${index}`].duration, undefined, `zone_${index}`, 'duration')"
              />
              <span class="ml-1 text-xs text-gray-500">sec</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onBeforeMount } from 'vue'
import { TrashIcon, SunIcon, MoonIcon, LightBulbIcon, ClockIcon, BeakerIcon } from '@heroicons/vue/24/outline'
import Chart from './Chart.vue'
import type { Phase, Chamber, WateringZoneSchedule } from '@/types'

interface Props {
  phase: Phase
  phaseIndex: number
  chamber: Chamber
}

const props = defineProps<Props>()

const emit = defineEmits<{
  update: [phase: Phase]
  remove: []
}>()

// Local phase data
const localPhase = reactive({ ...props.phase })



// Temperature schedule for the chart
const temperatureDaySchedule = ref<Record<number, number>>({})
const humidityDaySchedule = ref<Record<number, number>>({})
const co2DaySchedule = ref<Record<number, number>>({})
const temperatureNightSchedule = ref<Record<number, number>>({})
const humidityNightSchedule = ref<Record<number, number>>({})
const co2NightSchedule = ref<Record<number, number>>({})

// Lamp intensity schedules
const lampIntensitySchedules = reactive<Record<string, Record<number, number>>>({})

// Watering zone schedules
const wateringZoneSchedules = reactive<Record<string, {
  start_time_schedule: Record<number, number>
  period_schedule: Record<number, number>
  pause_between_schedule: Record<number, number>
  duration_schedule: Record<number, number>
}>>({})

// Environmental parameters
const startDayTime = ref<string>('09:00')

// Day duration schedule
const dayDurationSchedule = ref<Record<number, number>>({})

// Mode states for each schedule type (chart/form)
const scheduleMode = reactive({
  dayDuration: 'chart',
  temperatureDay: 'chart',
  temperatureNight: 'chart',
  humidityDay: 'chart',
  humidityNight: 'chart',
  co2Day: 'chart',
  co2Night: 'chart',
  lampIntensity: {} as Record<string, string>, // entity_id -> mode
  wateringZones: {} as Record<string, Record<string, string>> // zone_key -> schedule_type -> mode
} as Record<string, any>)

// Form values for single value mode
const formValues = reactive({
  dayDuration: 12,
  temperatureDay: 25,
  temperatureNight: 20,
  humidityDay: 60,
  humidityNight: 70,
  co2Day: 800,
  co2Night: 400,
  lampIntensity: {} as Record<string, number>,
  wateringZones: {} as Record<string, {
    start_time: number,
    period: number,
    pause_between: number,
    duration: number
  }>
} as Record<string, any>)

// Initialize values from phase data
onBeforeMount(() => {
  initializeValues()
  
})

watch(() => props.phase, () => {
  Object.assign(localPhase, props.phase)
}, { deep: true })

function initializeValues() {
  // Initialize start_day if not set
  const dayStartEntity = findInputNumberByType('day_start')
  if (dayStartEntity) {
    if (!localPhase.start_day) {
      localPhase.start_day = {}
    }
    if (!localPhase.start_day[dayStartEntity]) {
      localPhase.start_day[dayStartEntity] = { entity_id: dayStartEntity, value: 9 }
    }
    // Set start day time from start_day
    const startHour = localPhase.start_day[dayStartEntity].value
    startDayTime.value = `${startHour.toString().padStart(2, '0')}:00`
  }

  // Create default schedules for the entire duration
  const defaultTemperatureDaySchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 25])
  )
  const defaultHumidityDaySchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 60])
  )
  const defaultCO2DaySchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 800])
  )
  const defaultTemperatureNightSchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 20])
  )
  const defaultHumidityNightSchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 70])
  )
  const defaultCO2NightSchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 400])
  )
  const defaultDayDurationSchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 12])
  )
  const defaultLampIntensitySchedule = Object.fromEntries(
    Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 80])
  )


  // Initialize schedules from phase data if exists, otherwise use defaults
  const dayDurationEntity = findInputNumberByType('day_duration')
  if (dayDurationEntity) {
    dayDurationSchedule.value = props.phase.work_day_schedule?.[dayDurationEntity]?.schedule || defaultDayDurationSchedule
  }

  // Initialize temperature schedules
  const tempDayEntity = findInputNumberByType('temp_day')
  const tempNightEntity = findInputNumberByType('temp_night')
  if (tempDayEntity) {
    temperatureDaySchedule.value = props.phase.temperature_day_schedule?.[tempDayEntity]?.schedule || defaultTemperatureDaySchedule
  }
  if (tempNightEntity) {
    temperatureNightSchedule.value = props.phase.temperature_night_schedule?.[tempNightEntity]?.schedule || defaultTemperatureNightSchedule
  }

  // Initialize humidity schedules
  const humidityDayEntity = findInputNumberByType('humidity_day')
  const humidityNightEntity = findInputNumberByType('humidity_night')
  if (humidityDayEntity) {
    humidityDaySchedule.value = props.phase.humidity_day_schedule?.[humidityDayEntity]?.schedule || defaultHumidityDaySchedule
  }
  if (humidityNightEntity) {
    humidityNightSchedule.value = props.phase.humidity_night_schedule?.[humidityNightEntity]?.schedule || defaultHumidityNightSchedule
  }

  // Initialize CO2 schedules
  const co2DayEntity = findInputNumberByType('co2_day')
  const co2NightEntity = findInputNumberByType('co2_night')
  if (co2DayEntity) {
    co2DaySchedule.value = props.phase.co2_day_schedule?.[co2DayEntity]?.schedule || defaultCO2DaySchedule
  }
  if (co2NightEntity) {
    co2NightSchedule.value = props.phase.co2_night_schedule?.[co2NightEntity]?.schedule || defaultCO2NightSchedule
  }

  // Initialize lamp intensity schedules
  if(props.chamber.lamps && props.chamber.lamps.length > 0) {
    props.chamber.lamps.forEach((lamp) => {
      lampIntensitySchedules[lamp.entity_id] = props.phase.light_intensity_schedule?.[lamp.entity_id]?.schedule || defaultLampIntensitySchedule
      // Initialize form values and modes
      formValues.lampIntensity[lamp.entity_id] = 80
      scheduleMode.lampIntensity[lamp.entity_id] = 'chart'
    })
  }

  // Initialize watering zones
  if (props.chamber.watering_zones && props.chamber.watering_zones.length > 0) {
    props.chamber.watering_zones.forEach((zone, index) => {
      const zoneKey = `zone_${index}`
      
      // Default schedules for watering zones
      const defaultStartTimeSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 8]) // Default 8:00
      )
      const defaultPeriodSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 12]) // Default 12 hours
      )
      const defaultPauseBetweenSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 2]) // Default 2 hours
      )
      const defaultDurationSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, 60]) // Default 60 sec
      )
      
      // Initialize from phase data or use defaults
      wateringZoneSchedules[zoneKey] = {
        start_time_schedule: props.phase.watering_zones?.[zoneKey]?.start_time_schedule || defaultStartTimeSchedule,
        period_schedule: props.phase.watering_zones?.[zoneKey]?.period_schedule || defaultPeriodSchedule,
        pause_between_schedule: props.phase.watering_zones?.[zoneKey]?.pause_between_schedule || defaultPauseBetweenSchedule,
        duration_schedule: props.phase.watering_zones?.[zoneKey]?.duration_schedule || defaultDurationSchedule
      }
      
      // Initialize form values and modes
      formValues.wateringZones[zoneKey] = {
        start_time: 8,
        period: 12,
        pause_between: 2,
        duration: 60
      }
      scheduleMode.wateringZones[zoneKey] = {
        start_time: 'chart',
        period: 'chart',
        pause_between: 'chart',
        duration: 'chart'
      }
    })
  }

  // Update all schedules to ensure they are saved in phase
  if (dayDurationEntity) {
    updateDayDurationSchedule(dayDurationSchedule.value)
  }

  if (tempDayEntity) {
    updateTemperatureDaySchedule(temperatureDaySchedule.value)
  }
  if (tempNightEntity) {
    updateTemperatureNightSchedule(temperatureNightSchedule.value)
  }

  if (humidityDayEntity) {
    updateHumidityDaySchedule(humidityDaySchedule.value)
  }
  if (humidityNightEntity) {
    updateHumidityNightSchedule(humidityNightSchedule.value)
  }

  if (co2DayEntity) {
    updateCO2DaySchedule(co2DaySchedule.value)
  }
  if (co2NightEntity) {
    updateCO2NightSchedule(co2NightSchedule.value)
  }

  // Update lamp intensity schedules
  if(props.chamber.lamps && props.chamber.lamps.length > 0) {
    props.chamber.lamps.forEach((lamp) => {
      updateLampIntensitySchedule(lamp.entity_id, lampIntensitySchedules[lamp.entity_id])
    })
  }

  // Update watering zone schedules
  if (props.chamber.watering_zones && props.chamber.watering_zones.length > 0) {
    props.chamber.watering_zones.forEach((zone, index) => {
      const zoneKey = `zone_${index}`
      updateWateringZoneSchedule(zoneKey, 'start_time', wateringZoneSchedules[zoneKey].start_time_schedule)
      updateWateringZoneSchedule(zoneKey, 'period', wateringZoneSchedules[zoneKey].period_schedule)
      updateWateringZoneSchedule(zoneKey, 'pause_between', wateringZoneSchedules[zoneKey].pause_between_schedule)
      updateWateringZoneSchedule(zoneKey, 'duration', wateringZoneSchedules[zoneKey].duration_schedule)
    })
  }
}

function findInputNumberByType(type: string): string | null {
  const inputNumber = props.chamber.input_numbers.find((inputNum) => inputNum.type === type)
  return inputNumber?.entity_id || null
}

function updatePhase() {
  emit('update', { ...localPhase })
}

function updateTemperatureDaySchedule(schedule: Record<number, number>) {
  temperatureDaySchedule.value = schedule
  const tempDayEntity = findInputNumberByType('temperature_day')
  if (tempDayEntity) {
    if (!localPhase.temperature_day_schedule) {
      localPhase.temperature_day_schedule = {}
    }
    localPhase.temperature_day_schedule[tempDayEntity] = { entity_id: tempDayEntity, schedule }
  }
  updatePhase()
}

function updateHumidityDaySchedule(schedule: Record<number, number>) {
  humidityDaySchedule.value = schedule
  const humidityDayEntity = findInputNumberByType('humidity_day')
  if (humidityDayEntity) {
    if (!localPhase.humidity_day_schedule) {
      localPhase.humidity_day_schedule = {}
    }
    localPhase.humidity_day_schedule[humidityDayEntity] = { entity_id: humidityDayEntity, schedule }
  }
  updatePhase()
}

function updateCO2DaySchedule(schedule: Record<number, number>) {
  co2DaySchedule.value = schedule
  const co2DayEntity = findInputNumberByType('co2_day')
  if (co2DayEntity) {
    if (!localPhase.co2_day_schedule) {
      localPhase.co2_day_schedule = {}
    }
    localPhase.co2_day_schedule[co2DayEntity] = { entity_id: co2DayEntity, schedule }
  }
  updatePhase()
}

function updateTemperatureNightSchedule(schedule: Record<number, number>) {
  temperatureNightSchedule.value = schedule
  const tempNightEntity = findInputNumberByType('temperature_night')
  if (tempNightEntity) {
    if (!localPhase.temperature_night_schedule) {
      localPhase.temperature_night_schedule = {}
    }
    localPhase.temperature_night_schedule[tempNightEntity] = { entity_id: tempNightEntity, schedule }
  }
  updatePhase()
}

function updateHumidityNightSchedule(schedule: Record<number, number>) {
  humidityNightSchedule.value = schedule
  const humidityNightEntity = findInputNumberByType('humidity_night')
  if (humidityNightEntity) {
    if (!localPhase.humidity_night_schedule) {
      localPhase.humidity_night_schedule = {}
    }
    localPhase.humidity_night_schedule[humidityNightEntity] = { entity_id: humidityNightEntity, schedule }
  }
  updatePhase()
}

function updateCO2NightSchedule(schedule: Record<number, number>) {
  co2NightSchedule.value = schedule
  const co2NightEntity = findInputNumberByType('co2_night')
  if (co2NightEntity) {
    if (!localPhase.co2_night_schedule) {
      localPhase.co2_night_schedule = {}
    }
    localPhase.co2_night_schedule[co2NightEntity] = { entity_id: co2NightEntity, schedule }
  }
  updatePhase()
}

function updateLampIntensitySchedule(entity_id: string, schedule: Record<number, number>) {
  lampIntensitySchedules[entity_id] = schedule
  if (!localPhase.light_intensity_schedule) {
    localPhase.light_intensity_schedule = {}
  }
  if (!localPhase.light_intensity_schedule[entity_id]) {
    localPhase.light_intensity_schedule[entity_id] = { entity_id, schedule: {} }
  }
  localPhase.light_intensity_schedule[entity_id].schedule = schedule
  updatePhase()
}

function updateDayDurationSchedule(schedule: Record<number, number>) {
  dayDurationSchedule.value = schedule
  const dayDurationEntity = findInputNumberByType('day_duration')
  if (dayDurationEntity) {
    if (!localPhase.work_day_schedule) {
      localPhase.work_day_schedule = {}
    }
    localPhase.work_day_schedule[dayDurationEntity] = { entity_id: dayDurationEntity, schedule }
  }
  updatePhase()
}

function updateStartDayTime() {
  const timeValue = startDayTime.value
  if (timeValue) {
    const [hours, _] = timeValue.split(':').map(Number)
    const dayStartEntity = findInputNumberByType('day_start')
    if (dayStartEntity) {
      if (!localPhase.start_day) {
        localPhase.start_day = {}
      }
      localPhase.start_day[dayStartEntity] = { entity_id: dayStartEntity, value: hours }
      updatePhase()
    }
  }
}

function updateWateringZoneSchedule(zoneKey: string, scheduleType: 'start_time' | 'period' | 'pause_between' | 'duration', schedule: Record<number, number>) {
  const zoneIndex = parseInt(zoneKey.replace('zone_', ''))
  const zone = props.chamber.watering_zones[zoneIndex]
  if (!zone) return
  
  // Update local state
  wateringZoneSchedules[zoneKey][`${scheduleType}_schedule`] = schedule
  
  // Update phase data
  if (!localPhase.watering_zones) {
    localPhase.watering_zones = {}
  }
  
  if (!localPhase.watering_zones[zoneKey]) {
    localPhase.watering_zones[zoneKey] = {
      name: zone.name,
      start_time_entity_id: zone.start_time_entity_id,
      period_entity_id: zone.period_entity_id,
      pause_between_entity_id: zone.pause_between_entity_id,
      duration_entity_id: zone.duration_entity_id,
      start_time_schedule: {},
      period_schedule: {},
      pause_between_schedule: {},
      duration_schedule: {}
    }
  }
  
  localPhase.watering_zones[zoneKey][`${scheduleType}_schedule`] = schedule
  updatePhase()
}

// Mode switching functions
function toggleScheduleMode(type: string, entityId?: string, zoneKey?: string, scheduleType?: string) {
  if (zoneKey && scheduleType) {
    // Watering zone schedule
    if (!scheduleMode.wateringZones[zoneKey]) {
      scheduleMode.wateringZones[zoneKey] = {}
    }
    const currentMode = scheduleMode.wateringZones[zoneKey][scheduleType] || 'chart'
    scheduleMode.wateringZones[zoneKey][scheduleType] = currentMode === 'chart' ? 'form' : 'chart'
    
    if (scheduleMode.wateringZones[zoneKey][scheduleType] === 'form') {
      // Switch to form mode - create uniform schedule from form value
      const formValue = formValues.wateringZones[zoneKey]?.[scheduleType as keyof typeof formValues.wateringZones[string]] || getDefaultWateringValue(scheduleType)
      const uniformSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, formValue])
      )
      updateWateringZoneSchedule(zoneKey, scheduleType as any, uniformSchedule)
    }
  } else if (entityId) {
    // Lamp intensity schedule
    const currentMode = scheduleMode.lampIntensity[entityId] || 'chart'
    scheduleMode.lampIntensity[entityId] = currentMode === 'chart' ? 'form' : 'chart'
    
    if (scheduleMode.lampIntensity[entityId] === 'form') {
      // Switch to form mode - create uniform schedule from form value
      const formValue = formValues.lampIntensity[entityId] || 80
      const uniformSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, formValue])
      )
      updateLampIntensitySchedule(entityId, uniformSchedule)
    }
  } else {
    // Environmental schedule
    const currentMode = scheduleMode[type] || 'chart'
    scheduleMode[type] = currentMode === 'chart' ? 'form' : 'chart'
    
    if (scheduleMode[type] === 'form') {
      // Switch to form mode - create uniform schedule from form value
      const formValue = formValues[type]
      const uniformSchedule = Object.fromEntries(
        Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, formValue])
      )
      
      // Call appropriate update function
      switch (type) {
        case 'dayDuration': updateDayDurationSchedule(uniformSchedule); break
        case 'temperatureDay': updateTemperatureDaySchedule(uniformSchedule); break
        case 'temperatureNight': updateTemperatureNightSchedule(uniformSchedule); break
        case 'humidityDay': updateHumidityDaySchedule(uniformSchedule); break
        case 'humidityNight': updateHumidityNightSchedule(uniformSchedule); break
        case 'co2Day': updateCO2DaySchedule(uniformSchedule); break
        case 'co2Night': updateCO2NightSchedule(uniformSchedule); break
      }
    }
  }
}

function updateFormValue(type: string, value: number, entityId?: string, zoneKey?: string, scheduleType?: string) {
  if (zoneKey && scheduleType) {
    // Watering zone form value
    if (!formValues.wateringZones[zoneKey]) {
      formValues.wateringZones[zoneKey] = {
        start_time: 8,
        period: 12,
        pause_between: 2,
        duration: 60
      }
    }
    (formValues.wateringZones[zoneKey] as any)[scheduleType] = value
    
    // Update schedule with uniform value
    const uniformSchedule = Object.fromEntries(
      Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, value])
    )
    updateWateringZoneSchedule(zoneKey, scheduleType as any, uniformSchedule)
  } else if (entityId) {
    // Lamp intensity form value
    formValues.lampIntensity[entityId] = value
    
    // Update schedule with uniform value
    const uniformSchedule = Object.fromEntries(
      Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, value])
    )
    updateLampIntensitySchedule(entityId, uniformSchedule)
  } else {
    // Environmental form value
    formValues[type] = value
    
    // Update schedule with uniform value
    const uniformSchedule = Object.fromEntries(
      Array.from({ length: localPhase.duration_days }, (_, i) => [i+1, value])
    )
    
    // Call appropriate update function
    switch (type) {
      case 'dayDuration': updateDayDurationSchedule(uniformSchedule); break
      case 'temperatureDay': updateTemperatureDaySchedule(uniformSchedule); break
      case 'temperatureNight': updateTemperatureNightSchedule(uniformSchedule); break
      case 'humidityDay': updateHumidityDaySchedule(uniformSchedule); break
      case 'humidityNight': updateHumidityNightSchedule(uniformSchedule); break
      case 'co2Day': updateCO2DaySchedule(uniformSchedule); break
      case 'co2Night': updateCO2NightSchedule(uniformSchedule); break
    }
  }
}

function getDefaultWateringValue(scheduleType: string): number {
  switch (scheduleType) {
    case 'start_time': return 8
    case 'period': return 12
    case 'pause_between': return 2
    case 'duration': return 60
    default: return 0
  }
}
</script>