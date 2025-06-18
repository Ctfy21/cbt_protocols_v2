<template>
  <div v-if="isOpen" class="fixed inset-0 bg-white/70 bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white border border-gray-200 rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-900">Сохраненные шаблоны</h2>
        <button
          @click="$emit('close')"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <XMarkIcon class="w-6 h-6" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-6">
        <!-- Empty State -->
        <div v-if="templates.length === 0" class="text-center py-12">
          <DocumentIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
          <h3 class="text-lg font-medium text-gray-900 mb-2">Нет сохраненных шаблонов</h3>
          <p class="text-gray-500">
            Сохраните ваш первый шаблон из существующего эксперимента
          </p>
        </div>

        <!-- Templates Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-[60vh] overflow-y-auto">
          <div
            v-for="template in templates"
            :key="template.title"
            class="bg-gray-50 rounded-lg p-4 border border-gray-200 hover:border-blue-300 transition-colors"
          >
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1">
                <h3 class="font-medium text-gray-900 mb-1">{{ template.title }}</h3>
                <p class="text-sm text-gray-600 line-clamp-2">{{ template.description }}</p>
              </div>
              <button
                @click="deleteTemplate(template.title)"
                class="text-gray-400 hover:text-red-600 transition-colors ml-2"
                title="Удалить шаблон"
              >
                <TrashIcon class="w-4 h-4" />
              </button>
            </div>

            <div class="text-xs text-gray-500 mb-3">
              <div class="flex items-center gap-4">
                <span>{{ template.phases.length }} фаз</span>
                <span>{{ getTotalDays(template) }} дней</span>
              </div>
            </div>

            <div class="flex gap-2">
              <button
                @click="useTemplate(template)"
                class="flex-1 bg-blue-600 text-white px-3 py-2 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors"
              >
                Использовать шаблон
              </button>
              <button
                @click="exportTemplate(template)"
                class="px-3 py-2 border border-gray-300 text-gray-700 rounded-md text-sm hover:bg-gray-50 transition-colors"
                title="Экспорт шаблона"
              >
                <ArrowDownTrayIcon class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between p-6 border-t border-gray-200 bg-gray-50">
        <div class="flex gap-2">
          <input
            ref="fileInput"
            type="file"
            accept=".json"
            @change="handleImportTemplate"
            class="hidden"
          />
          <button
            @click="triggerFileInput"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <ArrowUpTrayIcon class="w-5 h-5 inline mr-2" />
            Импорт шаблона
          </button>
        </div>
        <button
          @click="$emit('close')"
          class="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
        >
          Закрыть
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { 
  XMarkIcon, 
  DocumentIcon, 
  TrashIcon, 
  ArrowDownTrayIcon, 
  ArrowUpTrayIcon 
} from '@heroicons/vue/24/outline'
import { useToastStore } from '@/stores/toast'
import type {Phase, ExperimentStatus, Experiment } from '@/types'
import { useChamberStore } from '@/stores/chamber'


const props = defineProps<Props>()
interface Props {
  isOpen: boolean
}


const emit = defineEmits<{
  close: []
  useTemplate: [template: Experiment]
}>()

const isOpen = computed(() => props.isOpen)

const chamberStore = useChamberStore()
const toastStore = useToastStore()
const templates = ref<Experiment[]>([])
const fileInput = ref<HTMLInputElement | null>(null)

const STORAGE_KEY = 'experiment_templates'

onMounted(() => {
  loadTemplates()
})

function loadTemplates() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      const all_templates = JSON.parse(saved)
      templates.value = all_templates.filter((t: Experiment) => t.chamber_id == chamberStore.selectedChamber?.id)
    }
  } catch (error) {
    console.error('Error loading templates:', error)
    templates.value = []
  }
}

function saveTemplates() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(templates.value))
  } catch (error) {
    console.error('Error saving templates:', error)
    toastStore.error('Ошибка', 'Не удалось сохранить шаблоны')
  }
}

function deleteTemplate(title: string) {
  templates.value = templates.value.filter((t: Experiment) => t.title !== title)
  saveTemplates()
  toastStore.success('Шаблон удален', 'Шаблон был успешно удален')
}

function useTemplate(template: Experiment) {
  emit('useTemplate', template)
  emit('close')
}

function exportTemplate(template: Experiment) {
  const data = {
    ...template,
    exported_at: new Date().toISOString(),
    version: '2.0'
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${template.title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_template.json`
  a.click()
  URL.revokeObjectURL(url)
  
  toastStore.success('Шаблон экспортирован', `Экспортирован шаблон ${template.title}`)
}

async function handleImportTemplate(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    // Validate template structure
    if (!data.title || !data.phases || !Array.isArray(data.phases)) {
      throw new Error('Неверный формат шаблона')
    }
    
    
    const template: Experiment = {
      id: '',
      created_at: '',
      updated_at: '',
      title: data.title,
      status: 'draft' as ExperimentStatus,
      chamber_id: '',
      description: data.description || '',
      phases: data.phases,
      schedule: data.schedule
    }
    
    templates.value.unshift(template)
    saveTemplates()
    toastStore.success('Шаблон импортирован', `Импортирован шаблон ${template.title}`)
    
    // Clear file input
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  } catch (error: any) {
    toastStore.error('Ошибка', error.message || 'Неверный формат файла')
  }
}

function getTotalDays(template: Experiment): number {
  return template.phases.reduce((sum: number, phase: Phase) => sum + (phase.duration_days || 0), 0)
}

function triggerFileInput() {
  fileInput.value?.click()
}

// Expose method to add template
function addTemplate(template: Experiment) {
  templates.value.unshift(template)
  saveTemplates()
}

defineExpose({ addTemplate })
</script> 