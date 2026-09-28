import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { generatedApi, type AuthUser, type LoginResponse } from '@/generated/api'
import { hasAnyPermission, hasPermission } from '@/lib/permissions'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | undefined>()
  const user = ref<AuthUser>()
  const restored = ref(false)
  let refreshTimer: ReturnType<typeof setTimeout> | undefined
  let refreshInFlight: Promise<string | undefined> | undefined

  const isAuthenticated = computed(() => Boolean(token.value && user.value))
  const can = (permission: string) => hasPermission(user.value?.permissions, permission)
  const canAny = (permissions: string[]) => hasAnyPermission(user.value?.permissions, permissions)

  async function login(email: string, password: string, turnstileToken?: string) {
    const response: LoginResponse = await generatedApi.login({ email, password, turnstile_token: turnstileToken })
    token.value = response.access_token
    user.value = response.user
    scheduleRefresh(response.access_token)
  }

  async function fetchCurrentUser() {
    const tokenValue = token.value
    if (!tokenValue) return
    user.value = await generatedApi.currentUser(tokenValue)
  }

  async function restore() {
    if (restored.value) return
    restored.value = true
    try {
      const refreshed = await refreshAccessToken(false)
      if (!refreshed) return
      await fetchCurrentUser()
    } catch {
      token.value = undefined
      user.value = undefined
      clearRefreshTimer()
    }
  }

  async function logout() {
    const tokenValue = token.value
    if (tokenValue) {
      await generatedApi.logout(tokenValue)
    }
    token.value = undefined
    user.value = undefined
    clearRefreshTimer()
  }

  async function logoutAll() {
    const tokenValue = token.value
    if (tokenValue) {
      await generatedApi.logoutAll(tokenValue)
    }
    token.value = undefined
    user.value = undefined
    clearRefreshTimer()
  }

  function clearRefreshTimer() {
    if (refreshTimer) clearTimeout(refreshTimer)
    refreshTimer = undefined
  }

  function tokenExpiry(accessToken: string) {
    try {
      const encoded = accessToken.split('.')[1]
      if (!encoded) return 0
      const normalized = encoded.replace(/-/g, '+').replace(/_/g, '/')
      const payload = JSON.parse(globalThis.atob(normalized + '='.repeat((4 - normalized.length % 4) % 4))) as { exp?: number }
      return typeof payload.exp === 'number' ? payload.exp * 1000 : 0
    } catch {
      return 0
    }
  }

  function scheduleRefresh(accessToken: string) {
    clearRefreshTimer()
    const expiresAt = tokenExpiry(accessToken)
    const delay = expiresAt > 0 ? Math.max(30_000, expiresAt - Date.now() - 120_000) : 45 * 60_000
    refreshTimer = setTimeout(() => { void refreshAccessToken(true) }, delay)
  }

  async function refreshAccessToken(clearOnFailure: boolean) {
    if (refreshInFlight) return refreshInFlight
    refreshInFlight = (async () => {
      try {
        const response = await generatedApi.refresh()
        token.value = response.access_token
        scheduleRefresh(response.access_token)
        return response.access_token
      } catch {
        if (clearOnFailure) {
          token.value = undefined
          user.value = undefined
          clearRefreshTimer()
        }
        return undefined
      } finally {
        refreshInFlight = undefined
      }
    })()
    return refreshInFlight
  }

  return { token, user, isAuthenticated, can, canAny, login, fetchCurrentUser, restore, logout, logoutAll }
})
