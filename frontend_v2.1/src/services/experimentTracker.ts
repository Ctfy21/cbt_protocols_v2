import type { Experiment } from '@/types'
import api from './api'
import { useExperimentStore } from '@/stores/experiment'
import { useToastStore } from '@/stores/toast'

class ExperimentTracker {
  private static instance: ExperimentTracker
  private trackingInterval: number | null = null
  private readonly TRACKING_INTERVAL = 60000 // Проверка каждую минуту
  private readonly COMPLETION_GRACE_PERIOD = 5 * 60 * 1000 // 5 минут задержки перед автозавершением

  private constructor() {}

  static getInstance(): ExperimentTracker {
    if (!ExperimentTracker.instance) {
      ExperimentTracker.instance = new ExperimentTracker()
    }
    return ExperimentTracker.instance
  }

  /**
   * Запускает отслеживание активных экспериментов
   */
  startTracking() {
    if (this.trackingInterval) {
      return // Уже запущено
    }

    console.log('Starting experiment tracking...')
    this.trackingInterval = window.setInterval(() => {
      this.checkExperiments()
    }, this.TRACKING_INTERVAL)

    // Проверяем сразу при запуске
    this.checkExperiments()
  }

  /**
   * Останавливает отслеживание экспериментов
   */
  stopTracking() {
    if (this.trackingInterval) {
      clearInterval(this.trackingInterval)
      this.trackingInterval = null
      console.log('Stopped experiment tracking')
    }
  }

  /**
   * Проверяет все активные эксперименты на предмет завершения
   */
  private async checkExperiments() {
    try {
      const experimentStore = useExperimentStore()
      const toastStore = useToastStore()
      
      // Получаем актуальный список экспериментов
      await experimentStore.fetchExperiments()
      
      const activeExperiments = experimentStore.activeExperiments
      console.log(`Checking ${activeExperiments.length} active experiments...`)

      for (const experiment of activeExperiments) {
        const isCompleted = this.isExperimentCompleted(experiment)
        
        if (isCompleted) {
          console.log(`Experiment "${experiment.title}" has completed, updating status...`)
          
          try {
            // Обновляем статус эксперимента на completed через специальный endpoint
            const response = await api.updateExperimentStatus(experiment.id, 'completed')
            
            if (response.success && response.data) {
              // Обновляем эксперимент в store
              const index = experimentStore.experiments.findIndex(e => e.id === experiment.id)
              if (index !== -1) {
                experimentStore.experiments[index] = response.data
              }
              
              toastStore.success(`Эксперимент "${experiment.title}" автоматически завершен`)
            }
          } catch (error) {
            console.error(`Failed to update experiment ${experiment.id}:`, error)
            toastStore.error(`Ошибка при обновлении статуса эксперимента "${experiment.title}"`)
          }
        }
      }
    } catch (error) {
      console.error('Error during experiment tracking:', error)
    }
  }

  /**
   * Проверяет, завершен ли эксперимент
   */
  private isExperimentCompleted(experiment: Experiment): boolean {
    if (!experiment.schedule || experiment.schedule.length === 0) {
      return false
    }

    const now = Date.now()
    
    // Находим последний элемент расписания
    const lastScheduleItem = experiment.schedule.reduce((latest, item) => {
      return item.end_timestamp > latest.end_timestamp ? item : latest
    })

    // Проверяем, прошло ли время окончания последней фазы + период ожидания
    const completionTime = (lastScheduleItem.end_timestamp * 1000) + this.COMPLETION_GRACE_PERIOD
    return now >= completionTime
  }

  /**
   * Возвращает информацию о прогрессе эксперимента
   */
  getExperimentProgress(experiment: Experiment): {
    currentPhase: number | null
    progressPercent: number
    timeRemaining: number | null
    isCompleted: boolean
  } {
    if (!experiment.schedule || experiment.schedule.length === 0) {
      return {
        currentPhase: null,
        progressPercent: 0,
        timeRemaining: null,
        isCompleted: false
      }
    }

    const now = Date.now()
    const totalDuration = this.getTotalExperimentDuration(experiment)
    const startTime = this.getExperimentStartTime(experiment)
    
    if (!startTime) {
      return {
        currentPhase: null,
        progressPercent: 0,
        timeRemaining: totalDuration,
        isCompleted: false
      }
    }

    const elapsed = now - startTime
    const progressPercent = Math.min(100, Math.max(0, (elapsed / totalDuration) * 100))
    const timeRemaining = Math.max(0, totalDuration - elapsed)
    const isCompleted = this.isExperimentCompleted(experiment)

    // Определяем текущую фазу
    let currentPhase: number | null = null
    for (const scheduleItem of experiment.schedule) {
      const phaseStart = scheduleItem.start_timestamp * 1000
      const phaseEnd = scheduleItem.end_timestamp * 1000
      
      if (now >= phaseStart && now < phaseEnd) {
        currentPhase = scheduleItem.phase_index
        break
      }
    }

    return {
      currentPhase,
      progressPercent,
      timeRemaining,
      isCompleted
    }
  }

  /**
   * Возвращает общую продолжительность эксперимента в миллисекундах
   */
  private getTotalExperimentDuration(experiment: Experiment): number {
    if (!experiment.schedule || experiment.schedule.length === 0) {
      return 0
    }

    const startTime = this.getExperimentStartTime(experiment)
    const endTime = this.getExperimentEndTime(experiment)

    if (!startTime || !endTime) {
      return 0
    }

    return endTime - startTime
  }

  /**
   * Возвращает время начала эксперимента
   */
  private getExperimentStartTime(experiment: Experiment): number | null {
    if (!experiment.schedule || experiment.schedule.length === 0) {
      return null
    }

    const firstScheduleItem = experiment.schedule.reduce((earliest, item) => {
      return item.start_timestamp < earliest.start_timestamp ? item : earliest
    })

    return firstScheduleItem.start_timestamp * 1000
  }

  /**
   * Возвращает время окончания эксперимента
   */
  private getExperimentEndTime(experiment: Experiment): number | null {
    if (!experiment.schedule || experiment.schedule.length === 0) {
      return null
    }

    const lastScheduleItem = experiment.schedule.reduce((latest, item) => {
      return item.end_timestamp > latest.end_timestamp ? item : latest
    })

    return lastScheduleItem.end_timestamp * 1000
  }
}

export default ExperimentTracker.getInstance() 