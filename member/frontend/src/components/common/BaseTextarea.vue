<template>
  <textarea
    v-bind="$attrs"
    class="base-textarea"
    :value="modelValue"
    autocorrect="off"
    autocapitalize="none"
    spellcheck="false"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue?: string
  }>(),
  {
    modelValue: '',
  },
)

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

useAttrs()

const onInput = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}
</script>
