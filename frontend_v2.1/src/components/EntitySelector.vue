<template>
    <div class="flex items-center gap-2">
        <p>{{ modelValue }}</p>
    </div>
</template>

<script setup lang="ts">
import type { InputNumber } from '@/types/index'

defineProps<{
  modelValue: string
  availableEntities: InputNumber[]
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script> 