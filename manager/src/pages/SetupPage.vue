<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '../stores/setup'
import { setupApi } from '../api/setup'
import { getErrorMessage } from '../api/client'

const router = useRouter()
const setupStore = useSetupStore()

const steps = ['Organization', 'Administrator Account', 'Complete']
const activeStep = ref(0)
const error = ref('')
const isLoading = ref(false)
const showPassword = ref(false)
const countdown = ref(20)

// Form state
const organizationName = ref('')
const adminEmail = ref('')
const adminName = ref('')
const adminPassword = ref('')
const confirmPassword = ref('')

function generatePassword(length = 10): string {
  const upper = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  const lower = 'abcdefghijklmnopqrstuvwxyz'
  const numbers = '0123456789'
  const all = upper + lower + numbers

  const password = [
    upper[Math.floor(Math.random() * upper.length)],
    lower[Math.floor(Math.random() * lower.length)],
    numbers[Math.floor(Math.random() * numbers.length)],
  ]

  for (let i = 3; i < length; i++) {
    password.push(all[Math.floor(Math.random() * all.length)])
  }

  return password.sort(() => Math.random() - 0.5).join('')
}

async function handleNext() {
  error.value = ''

  if (activeStep.value === 0) {
    if (!organizationName.value.trim()) {
      error.value = 'Organization name is required'
      return
    }
    activeStep.value++
  } else if (activeStep.value === 1) {
    if (!adminEmail.value.trim()) {
      error.value = 'Admin email is required'
      return
    }
    if (!adminName.value.trim()) {
      error.value = 'Admin name is required'
      return
    }
    if (!adminPassword.value) {
      error.value = 'Password is required'
      return
    }
    if (adminPassword.value.length < 8) {
      error.value = 'Password must be at least 8 characters'
      return
    }
    if (adminPassword.value !== confirmPassword.value) {
      error.value = 'Passwords do not match'
      return
    }
    await handleSubmit()
  }
}

function handleBack() {
  error.value = ''
  activeStep.value--
}

async function handleSubmit() {
  error.value = ''
  isLoading.value = true

  try {
    await setupApi.createAdmin({
      organization_name: organizationName.value,
      admin_email: adminEmail.value,
      admin_name: adminName.value,
      admin_password: adminPassword.value,
    })
    activeStep.value = 2 // Go to success step
  } catch (err) {
    error.value = getErrorMessage(err)
  } finally {
    isLoading.value = false
  }
}

async function handleGoToLogin() {
  await setupStore.checkSetupStatus()
  router.push('/login')
}

function handleGeneratePassword() {
  const newPassword = generatePassword(10)
  adminPassword.value = newPassword
  confirmPassword.value = newPassword
}

function handleCopyPassword() {
  navigator.clipboard.writeText(adminPassword.value)
}

// Countdown timer for success page
watch(countdown, (newVal) => {
  if (activeStep.value === 2 && newVal > 0) {
    setTimeout(() => countdown.value--, 1000)
  } else if (activeStep.value === 2 && newVal === 0) {
    handleGoToLogin()
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
    <div class="max-w-lg w-full bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8">
      <h1 class="text-3xl font-bold text-center mb-2">Welcome to AI Together</h1>
      <p class="text-gray-600 dark:text-gray-400 text-center mb-6">Let's set up your organization</p>

      <!-- Stepper -->
      <div class="mb-6">
        <div class="flex items-center justify-center">
          <div
            v-for="(step, index) in steps"
            :key="index"
            class="flex items-center"
          >
            <div
              class="flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium"
              :class="{
                'bg-blue-600 text-white': index <= activeStep,
                'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400': index > activeStep
              }"
            >
              {{ index + 1 }}
            </div>
            <span
              v-if="index < steps.length - 1"
              class="w-16 h-1 mx-1"
              :class="{
                'bg-blue-600': index < activeStep,
                'bg-gray-200 dark:bg-gray-700': index >= activeStep
              }"
            ></span>
          </div>
        </div>
        <div class="flex justify-center mt-2 text-sm">
          <div
            v-for="(step, index) in steps"
            :key="`label-${index}`"
            class="w-24 text-center"
            :class="{
              'text-blue-600 font-medium': index <= activeStep,
              'text-gray-500': index > activeStep
            }"
          >
            {{ step }}
          </div>
        </div>
      </div>

      <div
        v-if="error"
        class="bg-red-100 dark:bg-red-900 border border-red-400 dark:border-red-700 text-red-700 dark:text-red-200 px-4 py-3 rounded mb-4"
      >
        {{ error }}
      </div>

      <!-- Step 0: Organization -->
      <div v-if="activeStep === 0">
        <h2 class="text-xl font-semibold mb-4">Organization Information</h2>
        <div>
          <label for="orgName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Organization Name</label>
          <input
            id="orgName"
            v-model="organizationName"
            type="text"
            required
            autofocus
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            placeholder="Enter the name of your organization or team"
          />
        </div>
      </div>

      <!-- Step 1: Admin Account -->
      <div v-if="activeStep === 1">
        <h2 class="text-xl font-semibold mb-4">Administrator Account</h2>
        <div class="space-y-4">
          <div>
            <label for="adminEmail" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Admin Email</label>
            <input
              id="adminEmail"
              v-model="adminEmail"
              type="email"
              required
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div>
            <label for="adminName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Admin Name</label>
            <input
              id="adminName"
              v-model="adminName"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <div class="relative">
            <label for="adminPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Password</label>
            <div class="relative">
              <input
                id="adminPassword"
                v-model="adminPassword"
                :type="showPassword ? 'text' : 'password'"
                required
                class="w-full px-3 py-2 pr-20 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
              <div class="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1">
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="p-1 text-gray-500 hover:text-gray-700"
                >
                  {{ showPassword ? '👁️' : '👁️‍🗨️' }}
                </button>
                <button
                  v-if="adminPassword"
                  type="button"
                  @click="handleCopyPassword"
                  class="p-1 text-gray-500 hover:text-gray-700"
                  title="Copy password"
                >
                  📋
                </button>
              </div>
            </div>
            <p class="text-sm text-gray-500 mt-1">Minimum 8 characters</p>
          </div>
          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Confirm Password</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              :type="showPassword ? 'text' : 'password'"
              required
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
          </div>
          <button
            type="button"
            @click="handleGeneratePassword"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Generate Secure Password
          </button>
        </div>
      </div>

      <!-- Step 2: Complete -->
      <div v-if="activeStep === 2" class="text-center">
        <h2 class="text-2xl font-bold text-green-600 dark:text-green-400 mb-4">✓ Setup Complete!</h2>
        <p class="text-gray-600 dark:text-gray-400 mb-4">
          Your organization has been set up successfully. You can now log in with your administrator account.
        </p>
        <div class="bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700 rounded-lg p-4 mb-4 text-left">
          <p class="text-sm mb-1"><strong>Email:</strong> {{ adminEmail }}</p>
          <p class="text-sm"><strong>Password:</strong> {{ adminPassword }}</p>
        </div>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          Redirecting to login in {{ countdown }} second{{ countdown !== 1 ? 's' : '' }}...
        </p>
        <button
          @click="handleGoToLogin"
          class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md transition-colors"
        >
          Go to Login Now
        </button>
      </div>

      <!-- Navigation buttons -->
      <div v-if="activeStep < 2" class="flex justify-between mt-6">
        <button
          @click="handleBack"
          :disabled="activeStep === 0 || isLoading"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Back
        </button>
        <button
          @click="handleNext"
          :disabled="isLoading"
          class="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
        >
          <span v-if="isLoading" class="inline-block animate-spin mr-2">⟳</span>
          {{ activeStep === 1 ? 'Create Account' : 'Next' }}
        </button>
      </div>
    </div>
  </div>
</template>
