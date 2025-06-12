import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Experiment, ExperimentFormData, ExperimentStatus } from '@/types'
import api from '@/services/api'

export const useExperimentStore = defineStore('experiment', () => {
  // State
  const experiments = ref<Experiment[]>([])
  const currentExperiment = ref<Experiment | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const experimentsByStatus = computed(() => {
    const grouped: Record<ExperimentStatus, Experiment[]> = {
      draft: [],
      active: [],
      inactive: [],
      completed: [],
      paused: []
    }
    
    experiments.value.forEach(exp => {
      if (grouped[exp.status]) {
        grouped[exp.status].push(exp)
      }
    })
    
    return grouped
  })

  const activeExperiments = computed(() => 
    experiments.value.filter(e => e.status === 'active')
  )

  const stats = computed(() => {
    const total = experiments.value.length
    const active = experiments.value.filter(e => e.status === 'active').length
    const draft = experiments.value.filter(e => e.status === 'draft').length
    const totalDays = experiments.value.reduce((sum, e) => {
      return sum + (e.phases?.reduce((phaseSum, p) => phaseSum + (p.duration_days || 0), 0) || 0)
    }, 0)
    
    return { total, active, draft, totalDays }
  })

  // Actions
  async function fetchExperiments(chamberId?: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.getExperiments(chamberId ? { chamber_id: chamberId } : undefined)
      if (response.success && response.data) {
        experiments.value = response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
    } finally {
      loading.value = false
    }
  }

  async function fetchExperiment(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.getExperiment(id)
      if (response.success && response.data) {
        currentExperiment.value = response.data
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
    } finally {
      loading.value = false
    }
  }

  async function createExperiment(data: ExperimentFormData) {
    loading.value = true
    error.value = null
    try {
      const response = await api.createExperiment(data)
      if (response.success && response.data) {
        experiments.value.unshift(response.data)
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateExperiment(id: string, data: Partial<ExperimentFormData>) {
    loading.value = true
    error.value = null
    try {
      const response = await api.updateExperiment(id, data)
      if (response.success && response.data) {
        const index = experiments.value.findIndex(e => e.id === id)
        if (index !== -1) {
          experiments.value[index] = response.data
        }
        if (currentExperiment.value?.id === id) {
          currentExperiment.value = response.data
        }
        return response.data
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteExperiment(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await api.deleteExperiment(id)
      if (response.success) {
        experiments.value = experiments.value.filter(e => e.id !== id)
        if (currentExperiment.value?.id === id) {
          currentExperiment.value = null
        }
      }
    } catch (err) {
      error.value = api.formatError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  function duplicateExperiment(experiment: Experiment): ExperimentFormData {
    return {
      title: `${experiment.title} (Copy)`,
      description: experiment.description,
      status: 'draft',
      chamber_id: experiment.chamber_id,
      phases: experiment.phases.map(p => ({ ...p })),
      start_date: new Date().toISOString()
    }
  }

  return {
    // State
    experiments,
    currentExperiment,
    loading,
    error,
    // Getters
    experimentsByStatus,
    activeExperiments,
    stats,
    // Actions
    fetchExperiments,
    fetchExperiment,
    createExperiment,
    updateExperiment,
    deleteExperiment,
    duplicateExperiment
  }
}) 