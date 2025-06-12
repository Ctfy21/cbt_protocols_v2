<template>
  <header class="bg-white shadow-sm border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center py-4">
        <div class="flex items-center space-x-3">
          <div class="flex items-center space-x-2">
            <BeakerIcon class="w-8 h-8 text-blue-600" />
            <div>
              <h1 class="text-xl font-bold text-gray-900">Environmental Control Lab</h1>
              <div class="flex items-center gap-2 text-sm text-gray-600">
                <div class="flex items-center gap-1">
                  <div :class="[
                    'w-2 h-2 rounded-full',
                    chamberStore.selectedChamber?.status === 'online' ? 'bg-green-500' : 'bg-gray-400'
                  ]"></div>
                  <span>{{ chamberStore.selectedChamber?.name || 'No chamber selected' }}</span>
                </div>
                <span v-if="chamberStore.selectedChamber?.location" class="text-gray-400">•</span>
                <span v-if="chamberStore.selectedChamber?.location">{{ chamberStore.selectedChamber.location }}</span>
              </div>
            </div>
          </div>
          <span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded-full">
            v2.0
          </span>
        </div>
        
        <div class="flex items-center space-x-4">
          <router-link
            to="/chambers"
            class="text-gray-600 hover:text-gray-900 text-sm font-medium"
          >
            Switch Chamber
          </router-link>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { BeakerIcon } from '@heroicons/vue/24/outline'
import { useChamberStore } from '@/stores/chamber'

const chamberStore = useChamberStore()
</script> 