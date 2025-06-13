<template>
  <div class="fixed inset-0 bg-white/60 overflow-y-auto z-50">
    <div class="min-h-screen px-4 py-8 flex items-center justify-center">
      <div class="bg-white rounded-lg shadow-xl w-[80%] mx-auto border border-gray-200">
        <!-- Header -->
        <div class="px-8 py-6 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-3xl font-medium text-gray-900">
              {{ experiment ? 'Edit Experiment' : 'New Experiment' }}
            </h3>
            <button
              @click="$emit('close')"
              class="text-gray-400 hover:text-gray-600"
            >
              <XMarkIcon class="w-8 h-8" />
            </button>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="p-8">
          <!-- Basic Info -->
          <div class="space-y-8">
            <div>
              <h4 class="text-2xl font-medium text-gray-900 mb-6">Basic Information</h4>
              <div class="grid grid-cols-1 gap-6">
                <div>
                  <label class="block text-xl font-medium text-gray-700 mb-2">
                    Title
                  </label>
                  <input
                    v-model="form.title"
                    type="text"
                    required
                    class="w-full px-4 py-3 text-xl border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder="e.g., Tomato Growth Study"
                  />
                </div>
                <div>
                  <label class="block text-xl font-medium text-gray-700 mb-2">
                    Description
                  </label>
                  <textarea
                    v-model="form.description"
                    rows="4"
                    class="w-full px-4 py-3 text-xl border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder="Describe the experiment..."
                  ></textarea>
                </div>
                <div class="grid grid-cols-2 gap-6">
                  <div>
                    <label class="block text-xl font-medium text-gray-700 mb-2">
                      Start Date
                    </label>
                    <input
                      v-model="form.start_date"
                      type="date"
                      required
                      class="w-full px-4 py-3 text-xl border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>
                  <div>
                    <label class="block text-xl font-medium text-gray-700 mb-2">
                      Status
                    </label>
                    <select
                      v-model="form.status"
                      class="w-full px-4 py-3 text-xl border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="draft">Draft</option>
                      <option value="active">Active</option>
                      <option value="paused">Paused</option>
                      <option value="completed">Completed</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            <!-- Phases -->
            <div>
              <div class="flex items-center justify-between mb-6">
                <h4 class="text-2xl font-medium text-gray-900">Experiment Phases</h4>
                <button
                  type="button"
                  @click="addPhase"
                  class="text-xl text-blue-600 hover:text-blue-700 font-medium"
                >
                  + Add Phase
                </button>
              </div>

              <div v-if="form.phases.length === 0" class="text-center py-12 bg-gray-50 rounded-lg">
                <p class="text-xl text-gray-500">No phases added yet</p>
                <button
                  type="button"
                  @click="addPhase"
                  class="mt-4 text-xl text-blue-600 hover:text-blue-700 font-medium"
                >
                  Add your first phase
                </button>
              </div>

              <div v-else class="space-y-4">
                <PhaseEditor
                  v-for="(phase, index) in form.phases"
                  :key="index"
                  :phase="phase"
                  :phase-index="index"
                  :chamber="chamber"
                  @update="updatePhase(index, $event)"
                  @remove="removePhase(index)"
                >
                </PhaseEditor>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="mt-8 pt-8 border-t border-gray-200 flex justify-end gap-4">
            <button
              type="button"
              @click="$emit('close')"
              class="px-6 py-3 text-xl text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="loading || form.phases.length === 0"
              class="px-6 py-3 text-xl bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="loading" class="flex items-center">
                <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-white mr-2"></div>
                Saving...
              </span>
              <span v-else>{{ experiment ? 'Update' : 'Create' }} Experiment</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import PhaseEditor from './PhaseEditor.vue'
import type { Experiment, Chamber, Phase, ExperimentStatus, ScheduleItem } from '@/types'

interface Props {
  experiment?: Experiment | null
  chamber: Chamber
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
  save: [data: any]
}>()

const loading = ref(false)

const form = reactive({
  title: '',
  description: '',
  status: 'draft' as ExperimentStatus,
  chamber_id: props.chamber.id,
  start_date: new Date().toISOString().split('T')[0],
  phases: [] as Phase[],
  schedule: [] as ScheduleItem[]
})

// Initialize form with experiment data if editing
watch(() => props.experiment, (experiment) => {
  if (experiment) {
    form.title = experiment.title
    form.description = experiment.description
    form.status = experiment.status
    form.chamber_id = experiment.chamber_id
    form.start_date = experiment.start_date ? new Date(experiment.start_date).toISOString().split('T')[0] : new Date().toISOString().split('T')[0]
    form.phases = experiment.phases.map((p: Phase) => ({ ...p })) || []
    form.schedule = experiment.schedule || []
  }
}, { immediate: true })

function addPhase() {
  form.phases.push({
    title: `Phase ${form.phases.length + 1}`,
    description: '',
    duration_days: 7,
  })
  
  // Update schedule when adding a new phase
  updateSchedule()
}

function updatePhase(index: number, phase: Phase) {
  form.phases[index] = phase
  // Update schedule when phase is modified
  updateSchedule()
}

function removePhase(index: number) {
  form.phases.splice(index, 1)
  // Update schedule when phase is removed
  updateSchedule()
}

function updateSchedule() {
  const schedule: ScheduleItem[] = []
  let currentDate = new Date(form.start_date)
  
  form.phases.forEach((phase, index) => {
    const startDate = new Date(currentDate)
    const endDate = new Date(currentDate)
    endDate.setDate(endDate.getDate() + phase.duration_days)
    
    schedule.push({
      phase_index: index,
      start_timestamp: Math.floor(startDate.getTime()/1000),
      end_timestamp: Math.floor(endDate.getTime()/1000)
    })
    
    currentDate = endDate
  })
  
  form.schedule = schedule
}

// function addWateringZone(phase: Phase) {
//   const key = `zone_${Object.keys(phase.watering_zones).length + 1}`
//   phase.watering_zones[key] = {
//     name: '',
//     start_time_entity_id: '',
//     period_entity_id: '',
//     pause_between_entity_id: '',
//     duration_entity_id: ''
//   }
// }

// function removeWateringZone(phase: Phase, key: string) {
//   delete phase.watering_zones[key]
// }

async function handleSubmit() {
  loading.value = true
  try {
    // Update schedule before submitting
    updateSchedule()
    
    // Convert form data to API format
    const data = {
      ...form,
      start_date: new Date(form.start_date).toISOString()
    }
    
    emit('save', data)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style> 