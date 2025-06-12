<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <AppHeader />

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Experiment Calendar</h1>
        <p class="text-gray-600 mt-1">View and manage your experiment schedule</p>
      </div>

      <CalendarComponent
        :experiments="experimentStore.experiments"
        @create-experiment="handleCreateExperiment"
        @edit-experiment="handleEditExperiment"
      />
    </main>

    <!-- Experiment Form Modal -->
    <ExperimentForm
      v-if="showForm"
      :experiment="editingExperiment"
      :chamber="chamberStore.selectedChamber!"
      @close="closeForm"
      @save="handleSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChamberStore } from '@/stores/chamber'
import { useExperimentStore } from '@/stores/experiment'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import CalendarComponent from '@/components/CalendarView.vue'
import ExperimentForm from '@/components/ExperimentForm.vue'
import type { Experiment } from '@/types'

const router = useRouter()
const chamberStore = useChamberStore()
const experimentStore = useExperimentStore()
const toastStore = useToastStore()

const showForm = ref(false)
const editingExperiment = ref<Experiment | null>(null)
const defaultStartDate = ref<Date | null>(null)

onMounted(async () => {
  if (chamberStore.selectedChamber) {
    await experimentStore.fetchExperiments(chamberStore.selectedChamber.id)
  }
})

function handleCreateExperiment(options: { defaultStartDate: Date }) {
  editingExperiment.value = null
  defaultStartDate.value = options.defaultStartDate
  showForm.value = true
}

function handleEditExperiment(experiment: Experiment) {
  editingExperiment.value = experiment
  showForm.value = true
}

async function handleSave(data: any) {
  try {
    console.log(data)
    if (editingExperiment.value) {
      await experimentStore.updateExperiment(editingExperiment.value.id, data)
      toastStore.success('Experiment Updated', 'Changes saved successfully')
    } else {
      // Set default start date if provided
      if (defaultStartDate.value) {
        data.start_date = defaultStartDate.value.toISOString()
      }
      await experimentStore.createExperiment(data)
      toastStore.success('Experiment Created', 'New experiment created successfully')
    }
    closeForm()
  } catch (error: any) {
    toastStore.error('Save Failed', error.message || 'Failed to save experiment')
  }
}

function closeForm() {
  showForm.value = false
  editingExperiment.value = null
  defaultStartDate.value = null
}
</script> 