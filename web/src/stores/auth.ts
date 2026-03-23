import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authAPI, type AuthUser, type ChangePasswordPayload, type LoginPayload } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const initialized = ref(false)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  const clearAuth = () => {
    user.value = null
    initialized.value = true
  }

  const fetchCurrentUser = async (silent = false) => {
    if (!silent) {
      loading.value = true
    }

    try {
      const currentUser = await authAPI.me()
      user.value = currentUser
      initialized.value = true
      return currentUser
    } catch (error: any) {
      if (error?.response?.status === 401) {
        clearAuth()
        return null
      }

      initialized.value = true
      throw error
    } finally {
      if (!silent) {
        loading.value = false
      }
    }
  }

  const ensureInitialized = async () => {
    if (initialized.value) {
      return user.value
    }
    return fetchCurrentUser()
  }

  const login = async (payload: LoginPayload) => {
    await authAPI.login(payload)
    return fetchCurrentUser(true)
  }

  const logout = async () => {
    try {
      await authAPI.logout()
    } finally {
      clearAuth()
    }
  }

  const changePassword = async (payload: ChangePasswordPayload) => {
    await authAPI.changePassword(payload)
  }

  return {
    user,
    loading,
    initialized,
    isLoggedIn,
    isAdmin,
    clearAuth,
    ensureInitialized,
    fetchCurrentUser,
    login,
    logout,
    changePassword
  }
})

export const clearAuthState = () => {
  try {
    useAuthStore().clearAuth()
  } catch {
    // Pinia may not be active during early module evaluation.
  }
}
