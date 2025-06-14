<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="text-center">
      <div class="flex justify-center mb-8">
        <BeakerIcon class="w-24 h-24 text-blue-600" />
      </div>
      <h1 class="text-4xl font-bold text-gray-900 mb-4">
        Добро пожаловать в систему управления экспериментами в климатических камерах
      </h1>
      <p class="text-xl text-gray-600 mb-8">
        Управление и мониторинг экспериментов в климатических камерах
      </p>
      <div class="space-x-4">
        <router-link
          to="/chambers"
          class="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-md"
        >
          <HomeIcon class="w-5 h-5 mr-2" />
          Управление климатическими камерами
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { BeakerIcon, HomeIcon } from '@heroicons/vue/24/outline'

</script> 