<script setup lang="ts">
import { ref, onMounted } from 'vue'
import dayjs from 'dayjs'
import { usersApi, type UpdateUserRequest } from '../api/users'
import type { User } from '../types/models'

interface UserFormData {
  email: string
  name: string
  password: string
  role: 'manager' | 'member'
}

const initialFormData: UserFormData = {
  email: '',
  name: '',
  password: '',
  role: 'member',
}

const users = ref<User[]>([])
const isLoading = ref(false)
const openDialog = ref(false)
const editingUser = ref<User | null>(null)
const formData = ref<UserFormData>({ ...initialFormData })
const deleteConfirm = ref<User | null>(null)
const error = ref<string | null>(null)
const isSaving = ref(false)

async function fetchUsers() {
  isLoading.value = true
  try {
    const response = await usersApi.listUsers()
    users.value = response.users
  } finally {
    isLoading.value = false
  }
}

function handleOpenDialog(user?: User) {
  if (user) {
    editingUser.value = user
    formData.value = {
      email: user.email,
      name: user.name,
      password: '',
      role: user.role as 'manager' | 'member',
    }
  } else {
    editingUser.value = null
    formData.value = { ...initialFormData }
  }
  error.value = null
  openDialog.value = true
}

function handleCloseDialog() {
  openDialog.value = false
  editingUser.value = null
  formData.value = { ...initialFormData }
  error.value = null
}

async function handleSubmit() {
  error.value = null

  if (!formData.value.name.trim()) {
    error.value = 'Name is required'
    return
  }

  isSaving.value = true
  try {
    if (editingUser.value) {
      const updateData: UpdateUserRequest = {
        name: formData.value.name,
        role: formData.value.role,
      }
      if (formData.value.password.trim()) {
        updateData.password = formData.value.password
      }
      await usersApi.updateUser(editingUser.value.id, updateData)
    } else {
      if (!formData.value.email.trim()) {
        error.value = 'Email is required'
        return
      }
      if (!formData.value.password.trim() || formData.value.password.length < 6) {
        error.value = 'Password must be at least 6 characters'
        return
      }
      await usersApi.createUser(formData.value)
    }
    handleCloseDialog()
    await fetchUsers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    isSaving.value = false
  }
}

function handleDelete(user: User) {
  deleteConfirm.value = user
}

async function confirmDelete() {
  if (!deleteConfirm.value) return

  isSaving.value = true
  try {
    await usersApi.deleteUser(deleteConfirm.value.id)
    deleteConfirm.value = null
    await fetchUsers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">User Management</h1>
      <div class="flex gap-2">
        <button
          @click="fetchUsers"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        >
          ⟳ Refresh
        </button>
        <button
          @click="handleOpenDialog()"
          class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors"
        >
          + Add User
        </button>
      </div>
    </div>

    <div
      v-if="error"
      class="mb-4 p-4 rounded-lg bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200"
    >
      {{ error }}
      <button @click="error = null" class="float-right text-current opacity-70 hover:opacity-100">×</button>
    </div>

    <div v-if="isLoading" class="flex justify-center items-center py-20">
      <div class="animate-spin text-4xl">⟳</div>
    </div>

    <div v-else class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Email</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Role</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="user in users" :key="user.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ user.id }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ user.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ user.email }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="inline-flex px-2 py-1 text-xs font-semibold rounded-full"
                  :class="{
                    'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200': user.role === 'manager',
                    'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300': user.role === 'member'
                  }"
                >
                  {{ user.role }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                {{ user.created_at ? dayjs(user.created_at).format('YYYY-MM-DD') : '-' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="handleOpenDialog(user)" class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3">Edit</button>
                <button @click="handleDelete(user)" class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300">Delete</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
    <div
      v-if="openDialog"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-semibold">{{ editingUser ? 'Edit User' : 'Add User' }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div
            v-if="error"
            class="p-3 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 text-sm"
          >
            {{ error }}
          </div>
          <div v-if="!editingUser">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Email</label>
            <input
              v-model="formData.email"
              type="email"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Name</label>
            <input
              v-model="formData.name"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ editingUser ? 'New Password (leave blank to keep current)' : 'Password' }}
            </label>
            <input
              v-model="formData.password"
              type="password"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
            <p v-if="!editingUser" class="text-sm text-gray-500 mt-1">Minimum 6 characters</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Role</label>
            <select
              v-model="formData.role"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            >
              <option value="member">Member</option>
              <option value="manager">Manager</option>
            </select>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2">
          <button
            @click="handleCloseDialog"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleSubmit"
            :disabled="isSaving"
            class="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
          >
            {{ isSaving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div
      v-if="deleteConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-semibold">Confirm Delete</h2>
        </div>
        <div class="px-6 py-4">
          <p class="text-gray-700 dark:text-gray-300">
            Are you sure you want to delete user "{{ deleteConfirm.name }}" ({{ deleteConfirm.email }})?
            This action cannot be undone.
          </p>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2">
          <button
            @click="deleteConfirm = null"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="confirmDelete"
            :disabled="isSaving"
            class="bg-red-600 hover:bg-red-700 disabled:bg-gray-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
          >
            {{ isSaving ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
