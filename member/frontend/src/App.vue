<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
const applyTheme = () => {
  const userTheme = localStorage.getItem('theme')
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  const isDark = userTheme === 'dark' || (!userTheme && systemPrefersDark)

  document.documentElement.classList.toggle('dark', isDark)
}

onMounted(() => {
  applyTheme()

  // 可监听系统主题变化自动更新（可选）
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    applyTheme()
  })
})
</script>

<template>
  <RouterView />
</template>
