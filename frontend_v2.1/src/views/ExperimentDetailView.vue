<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <AppHeader />

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-gray-600">Loading experiment...</p>
      </div>

      <!-- Experiment Not Found -->
      <div v-else-if="!experiment" class="text-center py-12">
        <ExclamationCircleIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 mb-2">Experiment not found</h3>
        <router-link
          to="/experiments"
          class="text-blue-600 hover:text-blue-700"
        >
          Back to experiments
        </router-link>
      </div>

      <!-- Experiment Details -->
      <div v-else>
        <!-- Breadcrumb -->
        <nav class="mb-6">
          <ol class="flex items-center space-x-2 text-sm">
            <li>
              <router-link to="/experiments" class="text-gray-500 hover:text-gray-700">
                Experiments
              </router-link>
            </li>
            <li>
              <ChevronRightIcon class="w-4 h-4 text-gray-400" />
            </li>
            <li class="text-gray-900 font-medium">{{ experiment.title }}</li>
          </ol>
        </nav>

        <!-- Experiment Info Card -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ experiment.title }}</h1>
              <p class="text-gray-600">{{ experiment.description }}</p>
            </div>
            <div class="flex items-center gap-2">
              <div :class="statusClasses">
                {{ experiment.status }}
              </div>
              <button
                @click="editExperiment"
                class="p-2 text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
              >
                <PencilIcon class="w-5 h-5" />
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div>
              <span class="text-gray-500">Start Date:</span>
              <span class="ml-2 font-medium">{{ formatDate(experiment.start_date) }}</span>
            </div>
            <div>
              <span class="text-gray-500">End Date:</span>
              <span class="ml-2 font-medium">{{ formatDate(calculateEndDate()) }}</span>
            </div>
            <div>
              <span class="text-gray-500">Total Duration:</span>
              <span class="ml-2 font-medium">{{ totalDuration }} days</span>
            </div>
          </div>

          <!-- Progress Bar -->
          <div v-if="experiment.status === 'active' && progress >= 0" class="mt-4">
            <div class="flex items-center justify-between text-sm text-gray-600 mb-2">
              <span>Progress</span>
              <span>{{ Math.round(progress) }}%</span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-3">
              <div 
                class="bg-blue-600 h-3 rounded-full transition-all duration-300"
                :style="{ width: `${progress}%` }"
              ></div>
            </div>
          </div>
        </div>

        <!-- Phases -->
        <div class="space-y-6">
          <h2 class="text-lg font-semibold text-gray-900">Experiment Phases</h2>
          
          <div v-if="experiment.phases.length === 0" class="text-center py-8 bg-white rounded-lg shadow-sm border border-gray-200">
            <p class="text-gray-500">No phases defined</p>
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="(phase, index) in experiment.phases"
              :key="index"
              class="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
            >
              <div class="flex items-start justify-between mb-4">
                <div>
                  <h3 class="text-lg font-medium text-gray-900">{{ phase.title }}</h3>
                  <p class="text-sm text-gray-600 mt-1">{{ phase.description }}</p>
                </div>
                <div class="text-sm text-gray-500">
                  {{ phase.duration_days }} days
                </div>
              </div>


              <!-- Light Settings -->
              <div v-if="phase.light_intensity_schedule && Object.keys(phase.light_intensity_schedule).length > 0" class="mt-4 pt-4 border-t border-gray-100">
                <h4 class="text-sm font-medium text-gray-700 mb-2">Light Settings</h4>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
                  <div v-for="light in Object.values(phase.light_intensity_schedule)" :key="light.entity_id">
                    <span class="text-gray-500">{{ getLampName(light.entity_id) }}:</span>
                  </div>
                </div>
              </div>

              <!-- Watering Zones -->
              <div v-if="phase.watering_zones && Object.keys(phase.watering_zones).length > 0" class="mt-4 pt-4 border-t border-gray-100">
                <h4 class="text-sm font-medium text-gray-700 mb-2">Watering Zones</h4>
                <div class="space-y-2">
                  <div v-for="(zone, zoneKey) in phase.watering_zones" :key="zoneKey" class="text-sm">
                    <span class="font-medium">{{ zone.name }}:</span>
                    <div class="ml-4 text-gray-600">
                      <span v-if="getFirstScheduleValue(zone.start_time_schedule)">
                        Start: {{ getFirstScheduleValue(zone.start_time_schedule) }}:00,
                      </span>
                      <span v-if="getFirstScheduleValue(zone.period_schedule)">
                        Period: {{ getFirstScheduleValue(zone.period_schedule) }}h,
                      </span>
                      <span v-if="getFirstScheduleValue(zone.duration_schedule)">
                        Duration: {{ getFirstScheduleValue(zone.duration_schedule) }}s
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Edit Form Modal -->
    <ExperimentForm
      v-if="showEditForm"
      :experiment="experiment"
      :chamber="chamberStore.selectedChamber!"
      @close="showEditForm = false"
      @save="handleSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { format } from 'date-fns'
import { 
  ExclamationCircleIcon,
  ChevronRightIcon,
  PencilIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useExperimentStore } from '@/stores/experiment'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import ExperimentForm from '@/components/ExperimentForm.vue'
import type { Phase } from '@/types'

const route = useRoute()
const router = useRouter()
const chamberStore = useChamberStore()
const experimentStore = useExperimentStore()
const toastStore = useToastStore()

const loading = ref(true)
const showEditForm = ref(false)
const experiment = computed(() => experimentStore.currentExperiment)

onMounted(async () => {
  const id = route.params.id as string
  try {
    await experimentStore.fetchExperiment(id)
  } catch (error) {
    toastStore.error('Error', 'Failed to load experiment')
  } finally {
    loading.value = false
  }
})

const statusClasses = computed(() => {
  const baseClasses = 'px-3 py-1 text-sm font-medium rounded-full'
  
  switch (experiment.value?.status) {
    case 'active':
      return `${baseClasses} bg-green-100 text-green-800`
    case 'paused':
      return `${baseClasses} bg-yellow-100 text-yellow-800`
    case 'completed':
      return `${baseClasses} bg-blue-100 text-blue-800`
    case 'draft':
    default:
      return `${baseClasses} bg-gray-100 text-gray-800`
  }
})

const totalDuration = computed(() => {
  return experiment.value?.phases?.reduce((sum, phase) => sum + (phase.duration_days || 0), 0) || 0
})

const progress = computed(() => {
  if (!experiment.value || experiment.value.status !== 'active' || !experiment.value.start_date) {
    return -1
  }
  
  const now = new Date()
  const start = new Date(experiment.value.start_date)
  const endDate = calculateEndDate()
  
  if (!endDate) return -1
  
  const end = new Date(endDate)
  const total = end.getTime() - start.getTime()
  const elapsed = now.getTime() - start.getTime()
  
  return Math.min(100, Math.max(0, (elapsed / total) * 100))
})

function formatDate(date: string | null | undefined): string {
  if (!date) return 'Not set'
  try {
    return format(new Date(date), 'MMM d, yyyy')
  } catch {
    return 'Invalid date'
  }
}

function calculateEndDate(): string | null {
  if (!experiment.value?.start_date) return null
  
  const start = new Date(experiment.value.start_date)
  const totalDays = experiment.value.phases?.reduce((sum, phase) => sum + (phase.duration_days || 0), 0) || 0
  
  if (totalDays === 0) return null
  
  const end = new Date(start)
  end.setDate(end.getDate() + totalDays)
  
  return end.toISOString()
}

function getLampName(entityId: string): string {
  const lamp = chamberStore.selectedChamber?.lamps.find(l => l.entity_id === entityId)
  return lamp?.friendly_name || entityId
}

function getFirstScheduleValue(schedule?: Record<number, number>): number | null {
  if (!schedule || Object.keys(schedule).length === 0) return null
  const firstDay = Math.min(...Object.keys(schedule).map(Number))
  return schedule[firstDay] || null
}

function editExperiment() {
  showEditForm.value = true
}

async function handleSave(data: any) {
  try {
    if (experiment.value) {
      await experimentStore.updateExperiment(experiment.value.id, data)
      toastStore.success('Experiment Updated', 'Changes saved successfully')
      showEditForm.value = false
    }
  } catch (error: any) {
    toastStore.error('Save Failed', error.message || 'Failed to save experiment')
  }
}
</script>