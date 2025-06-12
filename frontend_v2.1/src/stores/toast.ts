import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function show(toast: Omit<Toast, 'id'>) {
    const id = Date.now().toString()
    const newToast: Toast = {
      id,
      duration: 5000,
      ...toast
    }
    
    toasts.value.push(newToast)
    
    if (newToast.duration && newToast.duration > 0) {
      setTimeout(() => {
        remove(id)
      }, newToast.duration)
    }
  }

  function remove(id: string) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index > -1) {
      toasts.value.splice(index, 1)
    }
  }

  function success(title: string, message?: string) {
    show({ type: 'success', title, message })
  }

  function error(title: string, message?: string) {
    show({ type: 'error', title, message })
  }

  function warning(title: string, message?: string) {
    show({ type: 'warning', title, message })
  }

  function info(title: string, message?: string) {
    show({ type: 'info', title, message })
  }

  return {
    toasts,
    show,
    remove,
    success,
    error,
    warning,
    info
  }
}) 