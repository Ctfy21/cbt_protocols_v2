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


      <!-- Day Duration Chart -->
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <ClockIcon class="w-6 h-6 text-purple-500" />
          <span>Day Duration Schedule</span>
        </div>
        <Chart
          v-model="dayDurationSchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="24"
          :step="1"
          :unit="' hours'"
          @update:model-value="updateDayDurationSchedule"
          title="Day Duration Schedule"
        />
      </div>


      <!-- Day Charts -->
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <SunIcon class="w-6 h-6 text-yellow-500" />
          <span>Day Schedule</span>
        </div>
        
        <!-- Temperature Chart -->
        <Chart
          v-model="temperatureDaySchedule" 
          :duration="localPhase.duration_days"
          :min="10"
          :max="30"
          :step="1"
          :unit="'°C'"
          @update:model-value="updateTemperatureDaySchedule"
          title="Temperature Schedule"
        />

        <!-- Humidity Chart -->
        <Chart
          v-model="humidityDaySchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="100"
          :step="5"
          :unit="'%'"
          @update:model-value="updateHumidityDaySchedule"
          title="Humidity Schedule"
        />

        <!-- CO2 Chart -->
        <Chart
          v-model="co2DaySchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="2000"
          :step="100"
          :unit="'ppm'"
          @update:model-value="updateCO2DaySchedule"
          title="CO2 Schedule"
        />
      </div>

      <!-- Night Charts -->
      <div class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <MoonIcon class="w-6 h-6 text-blue-500" />
          <span>Night Schedule</span>
        </div>

        <!-- Temperature Chart -->
        <Chart
          v-model="temperatureNightSchedule" 
          :duration="localPhase.duration_days"
          :min="10"
          :max="30"
          :step="1"
          :unit="'°C'"
          @update:model-value="updateTemperatureNightSchedule"
          title="Temperature Schedule"
        />

        <!-- Humidity Chart -->
        <Chart
          v-model="humidityNightSchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="100"
          :step="5"
          :unit="'%'"
          @update:model-value="updateHumidityNightSchedule"
          title="Humidity Schedule"
        />

        <!-- CO2 Chart -->
        <Chart
          v-model="co2NightSchedule" 
          :duration="localPhase.duration_days"
          :min="0"
          :max="2000"
          :step="100"
          :unit="'ppm'"
          @update:model-value="updateCO2NightSchedule"
          title="CO2 Schedule"
        />
      </div>

      <!-- Lamp Intensity Charts -->
      <div v-if="chamber.lamps && chamber.lamps.length > 0" class="space-y-4">
        <div class="flex items-center gap-2 text-lg font-medium text-gray-700 mt-20">
          <LightBulbIcon class="w-6 h-6 text-yellow-400" />
          <span>Lamp Intensity Schedules</span>
        </div>
        <div v-for="lamp in chamber.lamps" :key="lamp.entity_id">
          <Chart
            v-model="lampIntensitySchedules[lamp.entity_id]" 
            :duration="localPhase.duration_days"
            :min="lamp.intensity_min"
            :max="lamp.intensity_max"
            :step="5"
            :unit="'%'"
            @update:model-value="(schedule) => updateLampIntensitySchedule(lamp.entity_id, schedule)"
            :title="`${lamp.name} Intensity Schedule`"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { TrashIcon, SunIcon, MoonIcon, LightBulbIcon, ClockIcon } from '@heroicons/vue/24/outline'
import Chart from './Chart.vue'
import type { Phase, Chamber } from '@/types'

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

// Environmental parameters
const startDayTime = ref<string>('09:00')

// Day duration schedule
const dayDurationSchedule = ref<Record<number, number>>({})

// Initialize values from phase data
onMounted(() => {
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
  const tempDayEntity = findInputNumberByType('temperature_day')
  const tempNightEntity = findInputNumberByType('temperature_night')
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
      lampIntensitySchedules[lamp.entity_id] = defaultLampIntensitySchedule
    })
  }

  updatePhase()
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
  const dayStartEntity = findInputNumberByType('day_start')
  if (dayStartEntity) {
    if (!localPhase.work_day_schedule) {
      localPhase.work_day_schedule = {}
    }
    localPhase.work_day_schedule[dayStartEntity] = { entity_id: dayStartEntity, schedule }
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
</script>