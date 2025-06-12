<template>
  <div class="border border-gray-200 rounded-lg p-4">
    <div class="flex items-start justify-between mb-4">
      <div class="flex-1">
        <input
          v-model="localPhase.title"
          type="text"
          class="text-sm font-medium text-gray-900 bg-transparent border-0 p-0 focus:ring-0 w-full"
          placeholder="Phase title"
          @input="updatePhase"
        />
        <textarea
          v-model="localPhase.description"
          rows="2"
          class="mt-1 text-sm text-gray-500 bg-transparent border-0 p-0 focus:ring-0 w-full resize-none"
          placeholder="Phase description..."
          @input="updatePhase"
        ></textarea>
      </div>
      <button
        @click="$emit('remove')"
        class="ml-2 text-gray-400 hover:text-red-600"
      >
        <TrashIcon class="w-4 h-4" />
      </button>
    </div>

    <!-- Duration -->
    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700 mb-1">
        Duration (days)
      </label>
      <input
        v-model.number="localPhase.duration_days"
        type="number"
        min="1"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        @input="updatePhase"
      />
    </div>

    <!-- Environmental Parameters -->
    <div class="space-y-4">
      <div>
        <h5 class="text-sm font-medium text-gray-700 mb-2">Environmental Parameters</h5>
        <div class="grid grid-cols-2 gap-4">
          <!-- Temperature -->
          <div>
            <label class="block text-xs text-gray-600 mb-1">Day Temperature (°C)</label>
            <input
              v-model.number="tempDay"
              type="number"
              step="0.1"
              min="0"
              max="50"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
          <div>
            <label class="block text-xs text-gray-600 mb-1">Night Temperature (°C)</label>
            <input
              v-model.number="tempNight"
              type="number"
              step="0.1"
              min="0"
              max="50"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>

          <!-- Humidity -->
          <div>
            <label class="block text-xs text-gray-600 mb-1">Day Humidity (%)</label>
            <input
              v-model.number="humidityDay"
              type="number"
              step="1"
              min="0"
              max="100"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
          <div>
            <label class="block text-xs text-gray-600 mb-1">Night Humidity (%)</label>
            <input
              v-model.number="humidityNight"
              type="number"
              step="1"
              min="0"
              max="100"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>

          <!-- CO2 -->
          <div>
            <label class="block text-xs text-gray-600 mb-1">Day CO2 (ppm)</label>
            <input
              v-model.number="co2Day"
              type="number"
              step="10"
              min="0"
              max="2000"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
          <div>
            <label class="block text-xs text-gray-600 mb-1">Night CO2 (ppm)</label>
            <input
              v-model.number="co2Night"
              type="number"
              step="10"
              min="0"
              max="2000"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
        </div>
      </div>

      <!-- Light Settings -->
      <div>
        <h5 class="text-sm font-medium text-gray-700 mb-2">Light Settings</h5>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-gray-600 mb-1">Day Start (HH:MM)</label>
            <input
              v-model="dayStart"
              type="time"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
          <div>
            <label class="block text-xs text-gray-600 mb-1">Day Duration (hours)</label>
            <input
              v-model.number="dayDuration"
              type="number"
              step="0.5"
              min="0"
              max="24"
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
              @input="updateEnvironmentParams"
            />
          </div>
        </div>

        <!-- Lamp Intensities -->
        <div v-if="chamber.lamps && chamber.lamps.length > 0" class="mt-3">
          <label class="block text-xs text-gray-600 mb-1">Lamp Intensities (%)</label>
          <div class="grid grid-cols-2 gap-2">
            <div v-for="lamp in chamber.lamps" :key="lamp.entity_id" class="flex items-center gap-2">
              <span class="text-xs text-gray-600">{{ lamp.friendly_name }}:</span>
              <input
                v-model.number="lampIntensities[lamp.entity_id]"
                type="number"
                step="1"
                :min="lamp.intensity_min"
                :max="lamp.intensity_max"
                class="flex-1 px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500"
                @input="updateLampIntensities"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { TrashIcon } from '@heroicons/vue/24/outline'
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

// Environmental parameters
const tempDay = ref(24)
const tempNight = ref(20)
const humidityDay = ref(70)
const humidityNight = ref(60)
const co2Day = ref(500)
const co2Night = ref(400)
const dayStart = ref('09:00')
const dayDuration = ref(12)
const lampIntensities = reactive<Record<string, number>>({})

// Initialize values from phase data
onMounted(() => {
  initializeValues()
})

watch(() => props.phase, () => {
  Object.assign(localPhase, props.phase)
  initializeValues()
}, { deep: true })

function initializeValues() {
  // Extract values from input_numbers
  const inputNumbers = props.phase.input_numbers || {}
  
  // Temperature
  const tempDayEntity = findInputNumberByType('temp_day')
  const tempNightEntity = findInputNumberByType('temp_night')
  if (tempDayEntity) tempDay.value = inputNumbers[tempDayEntity]?.value || 24
  if (tempNightEntity) tempNight.value = inputNumbers[tempNightEntity]?.value || 20
  
  // Humidity
  const humidityDayEntity = findInputNumberByType('humidity_day')
  const humidityNightEntity = findInputNumberByType('humidity_night')
  if (humidityDayEntity) humidityDay.value = inputNumbers[humidityDayEntity]?.value || 70
  if (humidityNightEntity) humidityNight.value = inputNumbers[humidityNightEntity]?.value || 60
  
  // CO2
  const co2DayEntity = findInputNumberByType('co2_day')
  const co2NightEntity = findInputNumberByType('co2_night')
  if (co2DayEntity) co2Day.value = inputNumbers[co2DayEntity]?.value || 500
  if (co2NightEntity) co2Night.value = inputNumbers[co2NightEntity]?.value || 400
  
  // Light schedule
  const dayStartEntity = findInputNumberByType('day_start')
  const dayDurationEntity = findInputNumberByType('day_duration')
  if (dayStartEntity) {
    const startValue = inputNumbers[dayStartEntity]?.value || 9
    dayStart.value = `${Math.floor(startValue).toString().padStart(2, '0')}:${Math.round((startValue % 1) * 60).toString().padStart(2, '0')}`
  }
  if (dayDurationEntity) dayDuration.value = inputNumbers[dayDurationEntity]?.value || 12
  
  // Lamp intensities
  Object.entries(props.phase.light_intensity || {}).forEach(([entity_id, li]) => {
    lampIntensities[entity_id] = li.intensity
  })
}

function findInputNumberByType(type: string): string | null {
  const inputNumber = props.chamber.input_numbers.find((inputNum) => inputNum.type === type)
  return inputNumber?.entity_id || null
}

function updatePhase() {
  emit('update', { ...localPhase })
}

function updateEnvironmentParams() {
  const inputNumbers: Record<string, any> = {}
  
  // Temperature
  const tempDayEntity = findInputNumberByType('temp_day')
  const tempNightEntity = findInputNumberByType('temp_night')
  if (tempDayEntity) inputNumbers[tempDayEntity] = { entity_id: tempDayEntity, value: tempDay.value }
  if (tempNightEntity) inputNumbers[tempNightEntity] = { entity_id: tempNightEntity, value: tempNight.value }
  
  // Humidity
  const humidityDayEntity = findInputNumberByType('humidity_day')
  const humidityNightEntity = findInputNumberByType('humidity_night')
  if (humidityDayEntity) inputNumbers[humidityDayEntity] = { entity_id: humidityDayEntity, value: humidityDay.value }
  if (humidityNightEntity) inputNumbers[humidityNightEntity] = { entity_id: humidityNightEntity, value: humidityNight.value }
  
  // CO2
  const co2DayEntity = findInputNumberByType('co2_day')
  const co2NightEntity = findInputNumberByType('co2_night')
  if (co2DayEntity) inputNumbers[co2DayEntity] = { entity_id: co2DayEntity, value: co2Day.value }
  if (co2NightEntity) inputNumbers[co2NightEntity] = { entity_id: co2NightEntity, value: co2Night.value }
  
  // Light schedule
  const dayStartEntity = findInputNumberByType('day_start')
  const dayDurationEntity = findInputNumberByType('day_duration')
  if (dayStartEntity && dayStart.value) {
    const [hours, minutes] = dayStart.value.split(':').map(Number)
    inputNumbers[dayStartEntity] = { entity_id: dayStartEntity, value: hours + minutes / 60 }
  }
  if (dayDurationEntity) inputNumbers[dayDurationEntity] = { entity_id: dayDurationEntity, value: dayDuration.value }
  
  localPhase.input_numbers = inputNumbers
  updatePhase()
}

function updateLampIntensities() {
  localPhase.light_intensity = Object.fromEntries(Object.entries(lampIntensities).map(([entity_id, intensity]) => [entity_id, { entity_id, intensity }]))
  updatePhase()
}
</script> 