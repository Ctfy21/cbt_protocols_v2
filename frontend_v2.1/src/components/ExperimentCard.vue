<template>
  <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
    <!-- Header -->
    <div class="flex items-start justify-between mb-4">
      <div class="flex-1">
        <h3 class="text-lg font-semibold text-gray-900 mb-1">{{ experiment.title }}</h3>
        <p class="text-sm text-gray-500 line-clamp-2">{{ experiment.description }}</p>
      </div>
      <div :class="statusClasses">
        {{ experiment.status }}
      </div>
    </div>

    <!-- Info -->
    <div class="space-y-2 mb-4">
      <div class="flex items-center text-sm text-gray-600">
        <CalendarIcon class="w-4 h-4 mr-2" />
        <span>{{ formatDateRange }}</span>
      </div>
      <div class="flex items-center text-sm text-gray-600">
        <ChartBarIcon class="w-4 h-4 mr-2" />
        <span>{{ experiment.phases?.length || 0 }} фаз</span>
      </div>
      <div class="flex items-center text-sm text-gray-600">
        <ClockIcon class="w-4 h-4 mr-2" />
        <span>{{ totalDuration }} дней</span>
      </div>
    </div>

    <!-- Progress Bar (for active experiments) -->
    <div v-if="experiment.status === 'active' && progress >= 0" class="mb-4">
      <div class="flex items-center justify-between text-xs text-gray-600 mb-1">
        <span>Прогресс</span>
        <span>{{ Math.round(progress) }}%</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2">
        <div 
          class="bg-blue-600 h-2 rounded-full transition-all duration-300"
          :style="{ width: `${progress}%` }"
        ></div>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center justify-between pt-4 border-t border-gray-100">
      <!-- Status Actions -->
      <div class="flex items-center gap-1">
        <button
          v-if="experiment.status === 'draft'"
          @click="$emit('status-change', experiment, 'active')"
          class="p-2 text-green-600 hover:bg-green-50 rounded-md transition-colors"
          title="Начать эксперимент"
        >
          <PlayIcon class="w-4 h-4" />
        </button>
        <button
          v-if="experiment.status === 'active'"
          @click="$emit('status-change', experiment, 'paused')"
          class="p-2 text-yellow-600 hover:bg-yellow-50 rounded-md transition-colors"
          title="Приостановить эксперимент"
        >
          <PauseIcon class="w-4 h-4" />
        </button>
        <button
          v-if="experiment.status === 'paused'"
          @click="$emit('status-change', experiment, 'active')"
          class="p-2 text-green-600 hover:bg-green-50 rounded-md transition-colors"
          title="Продолжить эксперимент"
        >
          <PlayIcon class="w-4 h-4" />
        </button>
        <button
          v-if="experiment.status === 'active' || experiment.status === 'paused'"
          @click="$emit('status-change', experiment, 'completed')"
          class="p-2 text-red-600 hover:bg-red-50 rounded-md transition-colors"
          title="Завершить эксперимент"
        >
          <StopIcon class="w-4 h-4" />
        </button>
      </div>

      <!-- Menu Actions -->
      <div class="relative">
        <button
          @click="showMenu = !showMenu"
          class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-md transition-colors"
        >
          <EllipsisVerticalIcon class="w-4 h-4" />
        </button>
        
        <!-- Dropdown Menu -->
        <Transition
          enter-active-class="transition ease-out duration-100"
          enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100"
          leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100"
          leave-to-class="transform opacity-0 scale-95"
        >
          <div
            v-click-outside="() => showMenu = false"
            v-if="showMenu"            
            class="absolute right-0 bottom-full mb-2 w-48 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 divide-y divide-gray-100 z-10"
          >
            <div class="py-1">
              <button
                @click="handleAction('edit')"
                class="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
              >
                <PencilIcon class="w-4 h-4 mr-3" />
                Редактировать
              </button>
              <button
                @click="handleAction('duplicate')"
                class="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
              >
                <DocumentDuplicateIcon class="w-4 h-4 mr-3" />
                Дублировать
              </button>
              <button
                @click="handleAction('export')"
                class="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
              >
                <ArrowDownTrayIcon class="w-4 h-4 mr-3" />
                Экспорт
              </button>
            </div>
            <div class="py-1">
              <button
                @click="handleAction('delete')"
                class="flex items-center w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50"
              >
                <TrashIcon class="w-4 h-4 mr-3" />
                Удалить
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { format } from 'date-fns'
import {
  CalendarIcon,
  ChartBarIcon,
  ClockIcon,
  PlayIcon,
  PauseIcon,
  StopIcon,
  EllipsisVerticalIcon,
  PencilIcon,
  DocumentDuplicateIcon,
  ArrowDownTrayIcon,
  TrashIcon
} from '@heroicons/vue/24/outline'
import type { Experiment } from '@/types'

const props = defineProps<{
  experiment: Experiment
}>()

const emit = defineEmits(['edit', 'duplicate', 'export', 'delete', 'status-change'])

const showMenu = ref(false)

const statusClasses = computed(() => {
  const baseClasses = 'px-2 py-1 text-xs font-medium rounded-full'
  
  switch (props.experiment.status) {
    case 'active':
      return `${baseClasses} bg-green-100 text-green-800`
    case 'paused':
      return `${baseClasses} bg-yellow-100 text-yellow-800`
    case 'completed':
      return `${baseClasses} bg-blue-100 text-blue-800`
    case 'draft':
      return `${baseClasses} bg-gray-100 text-gray-800`
    default:
      return `${baseClasses} bg-gray-100 text-gray-800`
  }
})

const totalDuration = computed(() => {
  return props.experiment.phases?.reduce((sum, phase) => sum + (phase.duration_days || 0), 0) || 0
})

const formatDateRange = computed(() => {
  if (props.experiment.start_date) {
    try {
      const start = new Date(props.experiment.start_date)
      const endDate = props.experiment.end_date || calculateEndDate()
      const end = endDate ? new Date(endDate) : null
      
      if (end) {
        return `${format(start, 'MMM d')} - ${format(end, 'MMM d, yyyy')}`
      }
      return format(start, 'MMM d, yyyy')
    } catch {
      return 'Invalid date'
    }
  }
  return 'Not scheduled'
})

const progress = computed(() => {
  if (props.experiment.status !== 'active' || !props.experiment.start_date) {
    return -1
  }
  
  const now = new Date()
  const start = new Date(props.experiment.start_date)
  const endDate = props.experiment.end_date || calculateEndDate()
  
  if (!endDate) return -1
  
  const end = new Date(endDate)
  const total = end.getTime() - start.getTime()
  const elapsed = now.getTime() - start.getTime()
  
  return Math.min(100, Math.max(0, (elapsed / total) * 100))
})

function calculateEndDate(): string | null {
  if (!props.experiment.start_date) return null
  
  const start = new Date(props.experiment.start_date)
  const totalDays = props.experiment.phases?.reduce((sum, phase) => sum + (phase.duration_days || 0), 0) || 0
  
  if (totalDays === 0) return null
  
  const end = new Date(start)
  end.setDate(end.getDate() + totalDays)
  
  return end.toISOString()
}

function handleAction(action: 'edit' | 'duplicate' | 'export' | 'delete') {
  showMenu.value = false
  emit(action, props.experiment)
}

// Click outside directive
interface ClickOutsideElement extends HTMLElement {
  clickOutsideEvent?: (event: MouseEvent) => void
}

const vClickOutside = {
  mounted(el: ClickOutsideElement, binding: any) {
    el.clickOutsideEvent = function(event: MouseEvent) {
      // Get the menu button element
      const menuButton = el.parentElement?.querySelector('button')
      // Check if click is outside both the dropdown and the menu button
      if (!(el === event.target || el.contains(event.target as Node) || 
            menuButton === event.target || menuButton?.contains(event.target as Node))) {
        binding.value()
      }
    }
    document.addEventListener('click', el.clickOutsideEvent)
  },
  unmounted(el: ClickOutsideElement) {
    if (el.clickOutsideEvent) {
      document.removeEventListener('click', el.clickOutsideEvent)
    }
  }
}
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style> 