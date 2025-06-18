<template>
  <div class="bg-white rounded-lg shadow-sm border p-4">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-lg font-semibold text-gray-900">{{ experiment.title }}</h3>
      <span 
        class="px-2 py-1 text-xs font-medium rounded-full"
        :class="statusClasses"
      >
        {{ statusText }}
      </span>
    </div>
    
    <div class="mb-4">
      <div class="flex justify-between text-sm text-gray-600 mb-1">
        <span>Прогресс</span>
        <span>{{ Math.round(progress.progressPercent) }}%</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2">
        <div 
          class="bg-blue-600 h-2 rounded-full transition-all duration-300"
          :style="{ width: `${progress.progressPercent}%` }"
        ></div>
      </div>
    </div>
    
    <div class="grid grid-cols-2 gap-4 text-sm">
      <div>
        <span class="text-gray-600 block">Текущая фаза:</span>
        <span class="font-medium">
          {{ progress.currentPhase !== null ? `Фаза ${progress.currentPhase + 1}` : 'Не определена' }}
        </span>
      </div>
      <div>
        <span class="text-gray-600 block">Времени осталось:</span>
        <span class="font-medium">{{ formatTimeRemaining(progress.timeRemaining) }}</span>
      </div>
    </div>
    
    <div v-if="progress.currentPhase !== null && experiment.phases[progress.currentPhase]" class="mt-3 pt-3 border-t">
      <div class="text-sm">
        <span class="text-gray-600">Текущая фаза:</span>
        <span class="font-medium ml-1">{{ experiment.phases[progress.currentPhase].title }}</span>
      </div>
      <div class="text-xs text-gray-500 mt-1">
        {{ experiment.phases[progress.currentPhase].description }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { Experiment } from '@/types'
import experimentTracker from '@/services/experimentTracker'

interface Props {
  experiment: Experiment
}

const props = defineProps<Props>()

const progress = ref({
  currentPhase: null as number | null,
  progressPercent: 0,
  timeRemaining: null as number | null,
  isCompleted: false
})

let updateInterval: number | null = null

const statusClasses = computed(() => {
  switch (props.experiment.status) {
    case 'active':
      return 'bg-green-100 text-green-800'
    case 'completed':
      return 'bg-blue-100 text-blue-800'
    case 'paused':
      return 'bg-yellow-100 text-yellow-800'
    case 'draft':
      return 'bg-gray-100 text-gray-800'
    case 'archived':
      return 'bg-purple-100 text-purple-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
})

const statusText = computed(() => {
  switch (props.experiment.status) {
    case 'active':
      return 'Активный'
    case 'completed':
      return 'Завершен'
    case 'paused':
      return 'Приостановлен'
    case 'draft':
      return 'Черновик'
    case 'archived':
      return 'Архивирован'
    default:
      return props.experiment.status
  }
})

function updateProgress() {
  progress.value = experimentTracker.getExperimentProgress(props.experiment)
}

function formatTimeRemaining(timeMs: number | null): string {
  if (timeMs === null || timeMs <= 0) {
    return 'Завершен'
  }
  
  const days = Math.floor(timeMs / (1000 * 60 * 60 * 24))
  const hours = Math.floor((timeMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((timeMs % (1000 * 60 * 60)) / (1000 * 60))
  
  if (days > 0) {
    return `${days}д ${hours}ч`
  } else if (hours > 0) {
    return `${hours}ч ${minutes}м`
  } else {
    return `${minutes}м`
  }
}

onMounted(() => {
  updateProgress()
  // Обновляем прогресс каждые 30 секунд
  updateInterval = window.setInterval(updateProgress, 30000)
})

onUnmounted(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
})
</script> 