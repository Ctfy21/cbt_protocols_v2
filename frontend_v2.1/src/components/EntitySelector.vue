<template>
    <div class="flex items-center gap-2">
        <p>{{ modelValue }}</p>
    </div>
</template>

<script setup lang="ts">
// Entity interfaces
interface Entity {
  entity_id: string
  friendly_name: string
  name: string
  min: number
  max: number
  step: number
  value: number
  unit: string
}

defineProps<{
  modelValue: string
  availableEntities: Entity[]
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script> 