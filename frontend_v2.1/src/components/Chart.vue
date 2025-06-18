<template>
  <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
    <div class="mb-4">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <button
            @click.prevent="clearChart"
            class="px-3 py-2 text-base text-gray-600 hover:text-gray-800 border border-gray-300 rounded hover:bg-gray-50 transition-colors"
          >
            Очистить
          </button>
        </div>
      </div>
      
      <div class="relative" style="height: 800px;">
        <div 
          ref="chartContainer"
          class="relative w-full h-full border-2 border-gray-200 rounded-lg overflow-hidden bg-gradient-to-br from-gray-50 to-white"
        >
          <!-- Grid Pattern -->
          <div class="absolute inset-0 grid-pattern opacity-30"></div>
          
          <!-- Y-axis labels -->
          <div
            v-for="temp in temperatureLabels"
            :key="temp"
            class="absolute text-base text-gray-600"
            :style="getYLabelStyle(temp)"
          >
            {{ temp }}{{ unit }}
          </div>
          
          <!-- X-axis labels -->
          <div
            v-for="day in dayLabels"
            :key="day"
            class="absolute text-base text-gray-600"
            :style="getXLabelStyle(day)"
          >
            {{ day }}
          </div>
          
          <!-- Clickable Points -->
          <div
            v-for="point in gridPoints"
            :key="`${point.day}-${point.temp}`"
            @click="handlePointClick(point)"
            @mouseenter="showTooltip($event, point)"
            @mouseleave="hideTooltip"
            class="absolute w-3 h-3 rounded-full cursor-pointer transition-all duration-200 z-10"
            :class="getPointClasses(point)"
            :style="getPointStyle(point)"
          ></div>
          
          <!-- Tooltip -->
          <div
            ref="tooltipRef"
            v-show="tooltip.show"
            class="absolute bg-gray-900 text-white px-3 py-2 rounded text-base pointer-events-none z-20 whitespace-nowrap transition-opacity duration-200 shadow-lg"
            :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
          >
            Day {{ tooltip.day }}: {{ tooltip.temp }} {{ unit }}
          </div>
        </div>
      </div>
      
      <!-- Legend -->
      <div class="mt-4 flex items-center gap-4 text-base text-gray-600">
        <div class="flex items-center gap-2">
          <div class="w-3 h-3 bg-gray-300 rounded-full opacity-40"></div>
          <span>Доступно</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-4 h-4 bg-red-500 rounded-full"></div>
          <span>Выбрано</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'

interface Props {
  modelValue?: Record<number, number>
  duration: number
  min?: number
  max?: number
  step?: number
  unit?: string
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  min: 10,
  max: 30,
  step: 1,
  unit: '',
  title: '',
  modelValue: () => ({})
})

const emit = defineEmits<{
  'update:modelValue': [value: Record<number, number>]
}>()

const chartContainer = ref<HTMLElement>()
const tooltipRef = ref<HTMLElement>()
const selectedPoints = ref<Record<number, number>>({})
const tooltip = ref({
  show: false,
  x: 0,
  y: 0,
  day: 0,
  temp: 0,
  arrowPosition: 'bottom' // 'top', 'bottom', 'left', 'right'
})

// Chart configuration
const chartPadding = { left: 100, right: 40, top: 50, bottom: 50 }

// Container width for responsive chart
const containerWidth = ref(0)

// Update container width on resize
onMounted(() => {

  if (chartContainer.value) {
    const resizeObserver = new ResizeObserver(entries => {
      for (const entry of entries) {
        containerWidth.value = entry.contentRect.width
      }
    })
    resizeObserver.observe(chartContainer.value)
  }
})

// Initialize selected points from modelValue
watch(() => props.modelValue, (newValue) => {
  if (newValue && typeof newValue === 'object' && Object.keys(newValue).length > 0) {
    selectedPoints.value = { ...newValue }
  } else {
    selectedPoints.value = {}
  }
}, { immediate: true })

// Add debounce function
function debounce(fn: Function, delay: number) {
  let timeoutId: ReturnType<typeof setTimeout>
  return function (...args: any[]) {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => fn(...args), delay)
  }
}

// Update modelValue when selectedPoints change
const debouncedEmit = debounce((newValue: Record<number, number>) => {
  emit('update:modelValue', { ...newValue })
}, 100)

watch(selectedPoints, (newValue) => {
  debouncedEmit(newValue)
}, { deep: true })

// Styles
const chartWidth = computed(() => containerWidth.value - chartPadding.left - chartPadding.right)
const chartHeight = computed(() => 790 - chartPadding.top - chartPadding.bottom)

// Temperature labels for Y-axis
const temperatureLabels = computed(() => {
  const labels = []
  for (let temp = props.min; temp <= props.max; temp += props.step * 2) {
    labels.push(temp)
  }
  return labels
})

// Day labels for X-axis
const dayLabels = computed(() => {
  const labels = []
  const availableWidth = chartWidth.value
  const minLabelWidth = 60 // минимальная ширина для одной метки в пикселях
  const maxLabels = Math.floor(availableWidth / minLabelWidth)
  
  // Вычисляем шаг между метками
  const step = Math.max(1, Math.ceil(props.duration / maxLabels))
  
  // Добавляем метки с вычисленным шагом
  for (let day = 1; day <= props.duration; day += step) {
    labels.push(day)
  }
  
  // Всегда добавляем последний день, если его еще нет
  if (labels[labels.length - 1] !== props.duration) {
    labels.push(props.duration)
  }
  
  return labels
})

// Grid points
const gridPoints = computed(() => {
  const points = []
  for (let day = 1; day <= props.duration; day++) {
    for (let temp = props.min; temp <= props.max; temp += props.step) {
      points.push({ day, temp })
    }
  }
  return points
})

// Sorted selected days for line drawing
// const sortedSelectedDays = computed(() => {
//   return Object.keys(selectedPoints.value)
//     .map(Number)
//     .sort((a, b) => a - b)
// })


function getYLabelStyle(temp: number) {
  const y = chartPadding.top + (1 - (temp - props.min) / (props.max - props.min)) * chartHeight.value
  return {
    top: `${y}px`,
    left: '15px',
    transform: 'translateY(-50%)'
  }
}

function getXLabelStyle(day: number) {
  const x = chartPadding.left + ((day - 1) / (props.duration - 1)) * chartWidth.value
  return {
    bottom: '10px',
    left: `${x}px`,
    transform: 'translateX(-50%)'
  }
}

function getPointStyle(point: { day: number; temp: number }) {
  const x = chartPadding.left + ((point.day - 1) / (props.duration - 1)) * chartWidth.value
  const y = chartPadding.top + (1 - (point.temp - props.min) / (props.max - props.min)) * chartHeight.value
  return {
    left: `${x}px`,
    top: `${y}px`,
    transform: 'translate(-50%, -50%)'
  }
}

function getPointClasses(point: { day: number; temp: number }) {
  const isSelected = selectedPoints.value[point.day] === point.temp
  return [
    isSelected ? 'bg-red-500 w-3 h-3 opacity-100' : 'bg-gray-300 opacity-40',
    'hover:opacity-100',
    isSelected ? 'hover:w-4 hover:h-4' : 'hover:w-3 hover:h-3 hover:bg-gray-400'
  ]
}

function handlePointClick(point: { day: number; temp: number }) {
  const newSelectedPoints = { ...selectedPoints.value }
  
  if (newSelectedPoints[point.day] === point.temp) {
    // Remove point if already selected
    delete newSelectedPoints[point.day]
  } else {
    // Add or update point
    newSelectedPoints[point.day] = point.temp
  }
  
  selectedPoints.value = newSelectedPoints
}

function showTooltip(event: MouseEvent, point: { day: number; temp: number }) {
  const rect = chartContainer.value?.getBoundingClientRect()
  if (!rect) return
  
  // Временно показываем tooltip для получения его размеров
  tooltip.value = {
    show: true,
    x: 0,
    y: 0,
    day: point.day,
    temp: point.temp,
    arrowPosition: 'bottom'
  }
  
  // Ждем следующий тик, чтобы tooltip отрендерился
  nextTick(() => {
    const tooltipRect = tooltipRef.value?.getBoundingClientRect()
    const tooltipWidth = tooltipRect?.width || 120
    const tooltipHeight = tooltipRect?.height || 40
    
    // Базовые позиции tooltip
    let x = event.clientX - rect.left + 7
    let y = event.clientY - rect.top - tooltipHeight - 7
    let arrowPosition = 'bottom'
    
    // Проверяем правую границу
    if (x + tooltipWidth > rect.width) {
      x = event.clientX - rect.left - tooltipWidth - 7
      arrowPosition = 'right'
    }
    
    // Проверяем левую границу
    if (x < 0) {
      x = event.clientX - rect.left + 7
      arrowPosition = 'left'
    }
    
    // Проверяем верхнюю границу
    if (y < 0) {
      y = event.clientY - rect.top + 7
      arrowPosition = 'top'
    }
    
    // Проверяем нижнюю границу
    if (y + tooltipHeight > rect.height) {
      y = event.clientY - rect.top - tooltipHeight - 7
      arrowPosition = 'bottom'
    }
    
    tooltip.value = {
      show: true,
      x: x,
      y: y,
      day: point.day,
      temp: point.temp,
      arrowPosition: arrowPosition
    }
  })
}

function hideTooltip() {
  tooltip.value.show = false
}

function clearChart() {
  selectedPoints.value = {}
}

</script>

<style scoped>
.grid-pattern {
  background-image: 
    linear-gradient(to right, rgba(229, 231, 235, 0.5) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(229, 231, 235, 0.5) 1px, transparent 1px);
  background-size: 25px 20px;
}
</style>