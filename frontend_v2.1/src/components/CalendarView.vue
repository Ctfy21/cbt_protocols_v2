<template>
  <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <button
          @click="previousMonth"
          class="p-2 hover:bg-gray-100 rounded-md transition-colors"
        >
          <ChevronLeftIcon class="w-5 h-5" />
        </button>
        <h3 class="text-lg font-semibold text-gray-900">
          {{ format(currentMonth, 'MMMM yyyy') }}
        </h3>
        <button
          @click="nextMonth"
          class="p-2 hover:bg-gray-100 rounded-md transition-colors"
        >
          <ChevronRightIcon class="w-5 h-5" />
        </button>
      </div>
      <button
        @click="goToToday"
        class="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
      >
        Сегодня
      </button>
    </div>

    <!-- Calendar Grid -->
    <div class="grid grid-cols-7 gap-px bg-gray-200">
      <!-- Weekday Headers -->
      <div
        v-for="day in weekDays"
        :key="day"
        class="bg-gray-50 p-2 text-center text-xs font-medium text-gray-700"
      >
        {{ day }}
      </div>

      <!-- Calendar Days -->
      <div
        v-for="(day, index) in calendarDays"
        :key="index"
        :class="[
          'bg-white p-2 min-h-[100px] relative',
          day.isCurrentMonth ? '' : 'bg-gray-50',
          day.isToday ? 'ring-2 ring-blue-500' : ''
        ]"
      >
        <div class="text-xs text-gray-500 mb-1">
          {{ day.date.getDate() }}
        </div>
        
        <!-- Experiment Events -->
        <div class="space-y-1">
          <div
            v-for="event in day.events"
            :key="event.id"
            @click="$emit('edit-experiment', event.experiment)"
            :class="[
              'text-xs p-1 rounded cursor-pointer hover:opacity-80 transition-opacity',
              getEventColorClass(event.type, event.experiment.status)
            ]"
          >
            <div class="font-medium truncate">{{ event.experiment.title }}</div>
            <div class="text-xs opacity-75">{{ event.label }}</div>
          </div>
        </div>

        <!-- Add Event Button -->
        <button
          v-if="day.isCurrentMonth && day.events.length === 0"
          @click="$emit('create-experiment', { defaultStartDate: day.date })"
          class="absolute inset-0 w-full h-full opacity-0 hover:opacity-100 bg-gray-50 flex items-center justify-center transition-opacity"
        >
          <PlusIcon class="w-4 h-4 text-gray-400" />
        </button>
      </div>
    </div>

    <!-- Legend -->
    <div class="mt-4 flex items-center gap-4 text-xs">
      <div class="flex items-center gap-1">
        <div class="w-3 h-3 bg-green-200 rounded"></div>
        <span>Активен</span>
      </div>
      <div class="flex items-center gap-1">
        <div class="w-3 h-3 bg-yellow-200 rounded"></div>
        <span>Приостановлен</span>
      </div>
      <div class="flex items-center gap-1">
        <div class="w-3 h-3 bg-gray-200 rounded"></div>
        <span>Черновик</span>
      </div>
      <div class="flex items-center gap-1">
        <div class="w-3 h-3 bg-blue-200 rounded"></div>
        <span>Завершен</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  format, 
  startOfMonth, 
  endOfMonth, 
  startOfWeek, 
  endOfWeek,
  eachDayOfInterval,
  isSameMonth,
  isSameDay,
  addMonths,
  subMonths
} from 'date-fns'
import { ChevronLeftIcon, ChevronRightIcon, PlusIcon } from '@heroicons/vue/24/outline'
import type { Experiment, ExperimentStatus } from '@/types'

interface Props {
  experiments: Experiment[]
}

const props = defineProps<Props>()

// const emit = defineEmits<{
//   'create-experiment': [options: { defaultStartDate: Date }]
//   'edit-experiment': [experiment: Experiment]
// }>()

const currentMonth = ref(new Date())
const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

interface CalendarEvent {
  id: string
  experiment: Experiment
  type: 'start' | 'end' | 'ongoing'
  label: string
}

interface CalendarDay {
  date: Date
  isCurrentMonth: boolean
  isToday: boolean
  events: CalendarEvent[]
}

const calendarDays = computed((): CalendarDay[] => {
  const start = startOfWeek(startOfMonth(currentMonth.value))
  const end = endOfWeek(endOfMonth(currentMonth.value))
  const days = eachDayOfInterval({ start, end })
  const today = new Date()

  return days.map(date => {
    const events: CalendarEvent[] = []
    
    // Check each experiment for events on this day
    props.experiments.forEach(experiment => {
      if (!experiment.start_date) return
      
      const startDate = new Date(experiment.start_date)
      const endDate = calculateEndDate(experiment)
      
      if (isSameDay(date, startDate)) {
        events.push({
          id: `${experiment.id}-start`,
          experiment,
          type: 'start',
          label: 'Начало'
        })
      } else if (endDate && isSameDay(date, endDate)) {
        events.push({
          id: `${experiment.id}-end`,
          experiment,
          type: 'end',
          label: 'Конец'
        })
      } else if (
        experiment.status === 'active' &&
        date >= startDate &&
        endDate &&
        date <= endDate
      ) {
        // Show ongoing experiments on the first day of each month
        if (date.getDate() === 1) {
          events.push({
            id: `${experiment.id}-ongoing`,
            experiment,
            type: 'ongoing',
            label: 'Продолжается'
          })
        }
      }
    })

    return {
      date,
      isCurrentMonth: isSameMonth(date, currentMonth.value),
      isToday: isSameDay(date, today),
      events
    }
  })
})

function calculateEndDate(experiment: Experiment): Date | null {
  if (!experiment.start_date) return null
  
  const totalDays = experiment.phases?.reduce((sum, phase) => sum + (phase.duration_days || 0), 0) || 0
  if (totalDays === 0) return null
  
  const start = new Date(experiment.start_date)
  const end = new Date(start)
  end.setDate(end.getDate() + totalDays - 1)
  
  return end
}

function getEventColorClass(_: string, status: ExperimentStatus): string {
  const baseClasses = 'text-xs'
  
  switch (status) {
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
}

function previousMonth() {
  currentMonth.value = subMonths(currentMonth.value, 1)
}

function nextMonth() {
  currentMonth.value = addMonths(currentMonth.value, 1)
}

function goToToday() {
  currentMonth.value = new Date()
}
</script> 