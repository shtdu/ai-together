<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { profileApi } from '../api/profile'
import { getErrorMessage } from '../api/client'

const authStore = useAuthStore()

const name = ref(authStore.user?.name || '')
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const message = ref<{ type: 'success' | 'error'; text: string } | null>(null)
const isUpdatingProfile = ref(false)
const isChangingPassword = ref(false)

async function handleUpdateProfile() {
  if (!name.value.trim()) {
    message.value = { type: 'error', text: 'Name cannot be empty.' }
    return
  }

  isUpdatingProfile.value = true
  message.value = null

  try {
    const response = await profileApi.updateProfile({ name: name.value.trim() })
    // Update the auth store with the new name
    if (authStore.user) {
      authStore.user.name = response.user.name
    }
    message.value = { type: 'success', text: 'Profile updated successfully.' }
  } catch (error) {
    message.value = { type: 'error', text: getErrorMessage(error) }
  } finally {
    isUpdatingProfile.value = false
  }
}

async function handleUpdatePassword() {
  if (newPassword.value !== confirmPassword.value) {
    message.value = { type: 'error', text: 'New passwords do not match.' }
    return
  }
  if (newPassword.value.length < 6) {
    message.value = { type: 'error', text: 'New password must be at least 6 characters.' }
    return
  }

  isChangingPassword.value = true
  message.value = null

  try {
    await profileApi.changePassword({
      current_password: currentPassword.value,
      new_password: newPassword.value,
    })
    message.value = { type: 'success', text: 'Password changed successfully.' }
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (error) {
    message.value = { type: 'error', text: getErrorMessage(error) }
  } finally {
    isChangingPassword.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">Profile</h1>

    <div
      v-if="message"
      class="mb-4 p-4 rounded-lg"
      :class="{
        'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-200': message.type === 'success',
        'bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200': message.type === 'error'
      }"
    >
      {{ message.text }}
      <button @click="message = null" class="float-right text-current opacity-70 hover:opacity-100">×</button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Account Information -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 class="text-xl font-semibold mb-4">Account Information</h2>
        <form @submit.prevent="handleUpdateProfile">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Email</label>
            <input
              :value="authStore.user?.email || ''"
              disabled
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-700 cursor-not-allowed"
            />
            <p class="text-sm text-gray-500 mt-1">Email cannot be changed</p>
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Name</label>
            <input
              v-model="name"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Role</label>
            <input
              :value="authStore.user?.role === 'manager' ? 'Admin' : 'Member'"
              disabled
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-700 cursor-not-allowed"
            />
          </div>
          <button
            type="submit"
            :disabled="isUpdatingProfile"
            class="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
          >
            {{ isUpdatingProfile ? 'Saving...' : 'Update Profile' }}
          </button>
        </form>
      </div>

      <!-- Change Password -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 class="text-xl font-semibold mb-4">Change Password</h2>
        <form @submit.prevent="handleUpdatePassword">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Current Password</label>
            <input
              v-model="currentPassword"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div class="border-t border-gray-200 dark:border-gray-700 my-4"></div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">New Password</label>
            <input
              v-model="newPassword"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Confirm New Password</label>
            <input
              v-model="confirmPassword"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <button
            type="submit"
            :disabled="isChangingPassword"
            class="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
          >
            {{ isChangingPassword ? 'Changing...' : 'Change Password' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
