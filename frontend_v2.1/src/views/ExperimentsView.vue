<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <AppHeader />

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- View Toggle and Actions -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center space-x-1 bg-gray-100 rounded-lg p-1">
          <button
            @click="viewMode = 'list'"
            :class="[
              'px-4 py-2 rounded-md text-sm font-medium transition-colors',
              viewMode === 'list' 
                ? 'bg-white text-gray-900 shadow-sm' 
                : 'text-gray-600 hover:text-gray-900'
            ]"
          >
            <div class="flex items-center gap-2">
              <QueueListIcon class="w-4 h-4" />
              Отобразить список
            </div>
          </button>
          <button
            @click="viewMode = 'calendar'"
            :class="[
              'px-4 py-2 rounded-md text-sm font-medium transition-colors',
              viewMode === 'calendar' 
                ? 'bg-white text-gray-900 shadow-sm' 
                : 'text-gray-600 hover:text-gray-900'
            ]"
          >
            <div class="flex items-center gap-2">
              <CalendarIcon class="w-4 h-4" />
              Отобразить календарь
            </div>
          </button>
        </div>

        <div class="flex gap-3">
          <button
            @click="showTemplates = true"
            class="inline-flex items-center px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors shadow-sm"
          >
            <DocumentIcon class="w-5 h-5 mr-2" />
            Сохраненные шаблоны
          </button>
          <button
            @click="showCreateForm = true"
            class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-sm"
          >
            <PlusIcon class="w-5 h-5 mr-2" />
            Новый эксперимент
          </button>
        </div>
      </div>

      <!-- Stats Overview -->
      <div v-if="viewMode === 'list'" class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <BeakerIcon class="w-8 h-8 text-blue-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Всего экспериментов</p>
              <p class="text-3xl font-bold text-gray-900">{{ experimentStore.stats.total }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <PlayIcon class="w-8 h-8 text-green-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Активные</p>
              <p class="text-3xl font-bold text-gray-900">{{ experimentStore.stats.active }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <DocumentIcon class="w-8 h-8 text-yellow-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Черновики</p>
              <p class="text-3xl font-bold text-gray-900">{{ experimentStore.stats.draft }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow-sm p-6 border border-gray-200">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <ClockIcon class="w-8 h-8 text-purple-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-500">Всего дней</p>
              <p class="text-3xl font-bold text-gray-900">{{ experimentStore.stats.totalDays }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters and Search -->
      <div v-if="viewMode === 'list'" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-8">
        <div class="flex flex-col sm:flex-row gap-4">
          <!-- Search -->
          <div class="flex-1">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Поиск экспериментов..."
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          <!-- Status Filter -->
          <div class="sm:w-48">
            <select
              v-model="statusFilter"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Все статусы</option>
              <option value="draft">Черновики</option>
              <option value="active">Активные</option>
              <option value="paused">Приостановленные</option>
              <option value="completed">Завершенные</option>
            </select>
          </div>

          <!-- Import/Export -->
          <div class="flex gap-2">
            <input
              ref="fileInput"
              type="file"
              accept=".json"
              @change="handleImport"
              class="hidden"
            />
            <button
              @click="triggerFileInput"
              class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <ArrowUpTrayIcon class="w-5 h-5 inline mr-2" />
              Импорт
            </button>
            <button
              @click="handleExportAll"
              class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <ArrowDownTrayIcon class="w-5 h-5 inline mr-2" />
              Экспорт всех
            </button>
          </div>
        </div>
      </div>

      <!-- List View Content -->
      <div v-if="viewMode === 'list'">
        <!-- Loading State -->
        <div v-if="experimentStore.loading" class="text-center py-12">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-2 text-gray-600">Загрузка экспериментов...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredExperiments.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
          <BeakerIcon class="w-16 h-16 text-gray-300 mx-auto mb-4" />
          <h3 class="text-lg font-medium text-gray-900 mb-2">
            {{ experimentStore.experiments.length === 0 ? 'Нет экспериментов' : 'Нет экспериментов, соответствующих вашим фильтрам' }}
          </h3>
          <p class="text-gray-500 mb-6">
            {{ experimentStore.experiments.length === 0 
              ? `Создайте свой первый эксперимент для ${chamberStore.selectedChamber?.name}` 
              : 'Попробуйте отрегулировать ваш поиск или критерии фильтра'
            }}
          </p>
          <button
            v-if="experimentStore.experiments.length === 0"
            @click="showCreateForm = true"
            class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <PlusIcon class="w-5 h-5 mr-2" />
            Создать эксперимент
          </button>
        </div>

        <!-- Experiments Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <ExperimentCard
            v-for="experiment in filteredExperiments"
            :key="experiment.id"
            :experiment="experiment"
            @edit="handleEdit"
            @duplicate="handleDuplicate"
            @export="handleExport"
            @delete="handleDelete"
            @status-change="handleStatusChange"
            @save-template="saveAsTemplate"
          />
        </div>
      </div>

      <!-- Calendar View Content -->
      <div v-else-if="viewMode === 'calendar'">
        <CalendarView 
          :experiments="experimentStore.experiments"
          @create-experiment="showCreateForm = true"
          @edit-experiment="handleEdit"
        />
      </div>
    </main>

    <!-- Experiment Templates Modal -->
    <ExperimentTemplates
      :is-open="showTemplates"
      @close="showTemplates = false"
      @use-template="handleUseTemplate"
    />

    <!-- Experiment Form Modal -->
    <ExperimentForm
      v-if="showCreateForm || editingExperiment"
      :experiment="editingExperiment"
      :chamber="chamberStore.selectedChamber!"
      @close="closeForm"
      @save="handleSave"
    />

    <!-- Delete Confirmation Modal -->
    <ConfirmDialog
      v-if="deletingExperiment"
      title="Удалить эксперимент"
      :message="`Вы уверены, что хотите удалить '${deletingExperiment.title}'? Это действие не может быть отменено.`"
      confirm-text="Удалить"
      confirm-class="bg-red-600 hover:bg-red-700"
      @confirm="confirmDelete"
      @cancel="deletingExperiment = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { 
  BeakerIcon, 
  PlusIcon, 
  QueueListIcon, 
  CalendarIcon, 
  PlayIcon, 
  DocumentIcon, 
  ClockIcon, 
  MagnifyingGlassIcon,
  ArrowUpTrayIcon,
  ArrowDownTrayIcon
} from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'
import { useExperimentStore } from '@/stores/experiment'
import { useToastStore } from '@/stores/toast'
import AppHeader from '@/components/AppHeader.vue'
import ExperimentCard from '@/components/ExperimentCard.vue'
import ExperimentForm from '@/components/ExperimentForm.vue'
import CalendarView from '@/components/CalendarView.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import ExperimentTemplates from '@/components/ExperimentTemplates.vue'
import type { Experiment, ExperimentStatus } from '@/types'
import { useRouter } from 'vue-router'

const router = useRouter()
const chamberStore = useChamberStore()
const experimentStore = useExperimentStore()
const toastStore = useToastStore()

const viewMode = ref<'list' | 'calendar'>('list')
const searchQuery = ref('')
const statusFilter = ref<ExperimentStatus | ''>('')
const showCreateForm = ref(false)
const showTemplates = ref(false)
const editingExperiment = ref<Experiment | null>(null)
const deletingExperiment = ref<Experiment | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

// Load experiments when chamber is selected
onMounted(async () => {
  if (chamberStore.selectedChamber) {
    await experimentStore.fetchExperiments(chamberStore.selectedChamber.id)
  }
})

// Reload experiments when chamber changes
watch(() => chamberStore.selectedChamber, async (newChamber) => {
  if (newChamber) {
    await experimentStore.fetchExperiments(newChamber.id)
  }
})

// Filtered experiments based on search and status
const filteredExperiments = computed(() => {
  let filtered = experimentStore.experiments

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(exp => 
      exp.title.toLowerCase().includes(query) ||
      exp.description.toLowerCase().includes(query)
    )
  }

  // Status filter
  if (statusFilter.value) {
    filtered = filtered.filter(exp => exp.status === statusFilter.value)
  }

  return filtered
})

function handleEdit(experiment: Experiment) {
  editingExperiment.value = experiment
}

function handleDuplicate(experiment: Experiment) {
  const duplicated = experimentStore.duplicateExperiment(experiment)
  editingExperiment.value = { ...duplicated, id: '', created_at: '', updated_at: '' } as Experiment
}

function handleExport(experiment: Experiment) {
  const data = {
    ...experiment,
    exported_at: new Date().toISOString(),
    version: '2.0'
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${experiment.title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_experiment.json`
  a.click()
  URL.revokeObjectURL(url)
  
  toastStore.success('Эксперимент экспортирован', `Экспортирован ${experiment.title}`)
}

function handleExportAll() {
  const data = {
    chamber: chamberStore.selectedChamber,
    experiments: experimentStore.experiments,
    exported_at: new Date().toISOString(),
    version: '2.0'
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${chamberStore.selectedChamber?.name.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_all_experiments_${new Date().toISOString().split('T')[0]}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  toastStore.success('Эксперименты экспортированы', 'Экспортированы все эксперименты')
}

async function handleImport(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    // Handle single experiment or multiple experiments
    const experiments = Array.isArray(data.experiments) ? data.experiments : [data]
    
    for (const exp of experiments) {
      // Clean up imported data
      delete exp.id
      delete exp.created_at
      delete exp.updated_at
      exp.status = 'draft'
      exp.title = exp.title + ' (Imported)'
      exp.chamber_id = chamberStore.selectedChamber!.id
      
      await experimentStore.createExperiment(exp)
    }
    
    toastStore.success('Эксперименты импортированы', `Импортировано ${experiments.length} эксперимент(ов)`)
    
    // Clear file input
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  } catch (error: any) {
    toastStore.error('Ошибка', error.message || 'Неверный формат файла')
  }
}

function handleDelete(experiment: Experiment) {
  deletingExperiment.value = experiment
}

async function confirmDelete() {
  if (!deletingExperiment.value) return
  
  try {
    await experimentStore.deleteExperiment(deletingExperiment.value.id)
    toastStore.success('Эксперимент удален', `Удален ${deletingExperiment.value.title}`)
    deletingExperiment.value = null
  } catch (error: any) {
    toastStore.error('Ошибка', error.message || 'Не удалось удалить эксперимент')
  }
}

async function handleStatusChange(experiment: Experiment, status: ExperimentStatus) {
  try {
    await experimentStore.updateExperiment(experiment.id, { ...experiment, status })
    toastStore.success('Статус обновлен', `${experiment.title} теперь ${status}`)
  } catch (error: any) {
    toastStore.error('Ошибка', error.message || 'Не удалось обновить статус')
  }
}

async function handleSave(data: any) {
  console.log(data)
  try {
    if (editingExperiment.value?.id) {
      await experimentStore.updateExperiment(editingExperiment.value.id, data)
      toastStore.success('Experiment Updated', 'Changes saved successfully')
    } else {
      await experimentStore.createExperiment(data)
      toastStore.success('Experiment Created', 'New experiment created successfully')
    }
    closeForm()
  } catch (error: any) {
    toastStore.error('Save Failed', error.message || 'Failed to save experiment')
  }
}

function closeForm() {
  showCreateForm.value = false
  editingExperiment.value = null
}

function triggerFileInput() {
  fileInput.value?.click()
}

function handleUseTemplate(template: Experiment) {
  console.log(template)
  template.id = ''
  template.created_at = ''
  template.updated_at = ''
  editingExperiment.value = template
}

function saveAsTemplate(experiment: Experiment) {
  experiment.id = ''
  experiment.created_at = ''
  experiment.updated_at = ''
  // Save to localStorage using the same logic as ExperimentTemplates component
  try {
    const STORAGE_KEY = 'experiment_templates'
    const saved = localStorage.getItem(STORAGE_KEY)
    const templates = saved ? JSON.parse(saved) : []
    templates.unshift(experiment)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(templates))
    toastStore.success('Шаблон сохранен', `Создан шаблон из ${experiment.title}`)
    router.go(0)
  } catch (error: any) {
    toastStore.error('Ошибка', 'Не удалось сохранить шаблон')
  }
}
</script> 