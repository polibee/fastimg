// The isolated development previews cannot reach a Windows loopback listener
// through their own origin. Keep the backend explicit for the two reserved
// FastImg preview ports instead of silently calling the frontend server.
const previewPorts = new Set(['53083', '53084'])
const API_BASE_URL = import.meta.env?.VITE_API_BASE_URL ?? (previewPorts.has(globalThis.location?.port ?? '') ? 'http://127.0.0.1:53085' : '')

/**
 * Public machine-readable endpoints are served by the backend. During local
 * development the member/admin UI and backend intentionally use different
 * ports, so links such as sitemap.xml must not fall back to the SPA entry.
 */
export function publicEndpointURL(path: string) {
  const base = String(API_BASE_URL || globalThis.location?.origin || '').replace(/\/+$/, '')
  return `${base}${path}`
}

const LOCALIZED_ERROR_CODES = new Set([
  'VALIDATION_ERROR',
  'AUTH_INVALID_CREDENTIALS',
  'AUTH_UNAUTHORIZED',
  'AUTH_SESSION_STORE_UNAVAILABLE',
  'AUTH_RATE_LIMITED',
  'AUTH_RATE_LIMIT_STORE_UNAVAILABLE',
  'AUTH_VERIFICATION_RESEND_RATE_LIMITED',
  'AUTH_REGISTRATION_DISABLED',
  'AUTH_REGISTRATION_FAILED',
  'AUTH_CAPTCHA_REQUIRED',
  'AUTH_CAPTCHA_FAILED',
  'AUTH_CAPTCHA_UNAVAILABLE',
  'AUTH_EMAIL_UNVERIFIED',
  'AUTH_EMAIL_TAKEN',
  'AUTH_EMAIL_NOT_ALLOWED',
  'AUTH_EMAIL_UNAVAILABLE',
  'AUTH_EMAIL_VERIFICATION_EXPIRED',
  'RBAC_FORBIDDEN',
  'RBAC_PERMISSIONS_ERROR',
  'RBAC_ROLE_NOT_FOUND',
  'RBAC_PERMISSION_NOT_FOUND',
  'RBAC_USER_NOT_FOUND',
  'RBAC_SYSTEM_ROLE',
  'RBAC_LAST_ADMIN',
  'RESOURCE_NOT_FOUND',
  'RELATION_NOT_FOUND',
  'RELATION_FORBIDDEN',
  'SETTINGS_INVALID',
  'SETTINGS_ERROR',
  'INTERNAL_ERROR',
  'PLAN_NOT_FOUND',
  'SUBSCRIPTION_UPDATE_FAILED',
  'SUBSCRIPTION_UNAVAILABLE',
  'GATEWAY_UNAVAILABLE',
  'PAYMENT_PROVIDER_UNAVAILABLE',
  'PAYMENT_PROVIDER_REJECTED',
  'ORDER_NOT_CANCELLABLE',
  'AUDIT_CLEANUP_CONFIRMATION_REQUIRED',
  'AUDIT_CLEANUP_SELECTION_REQUIRED',
  'AUDIT_CLEANUP_SELECTION_TOO_LARGE',
  'AUDIT_CLEANUP_FILTER_REQUIRED',
  'STORAGE_CONNECTIONS_UNAVAILABLE',
  'STORAGE_CONNECTION_NOT_FOUND',
  'STORAGE_CONFIGURATION_INVALID',
  'STORAGE_PROVIDER_ADAPTER_PENDING',
  'STORAGE_STATISTICS_UNAVAILABLE',
  'PUBLIC_ALBUM_NOT_FOUND',
  'PUBLIC_ALBUM_MEDIA_NOT_FOUND',
  'PUBLIC_ALBUM_CONTENT_UNAVAILABLE',
  'QUEUE_TASKS_UNAVAILABLE',
  'QUEUE_TASK_INVALID',
  'QUEUE_TASK_NOT_FOUND',
  'QUEUE_TASK_RETRY_FAILED',
])

export function errorMessageKey(code?: string) {
  return code && LOCALIZED_ERROR_CODES.has(code) ? `errors.${code}` : 'errors.unknown'
}

export class ApiError extends Error {
  readonly status: number
  readonly code?: string

  constructor(
    message: string,
    status: number,
    code?: string,
  ) {
    super(message)
    this.status = status
    this.code = code
  }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(init.headers)
  if (!(typeof FormData !== 'undefined' && init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers, credentials: 'include' })
  const payload = await response.json().catch(() => null) as { data?: T; code?: string; message?: string } | null
  if (!response.ok) {
    throw new ApiError(payload?.message ?? 'Request failed', response.status, payload?.code)
  }

  return (payload?.data ?? payload) as T
}

export async function apiFetchBlob(path: string, token?: string): Promise<Blob> {
  const headers = new Headers()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${API_BASE_URL}${path}`, { headers, credentials: 'include' })
  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { code?: string; message?: string } | null
    throw new ApiError(payload?.message ?? 'Request failed', response.status, payload?.code)
  }
  return response.blob()
}

export async function apiFetchEnvelope<T>(path: string, init: RequestInit = {}, token?: string): Promise<{ data: T; meta?: Record<string, unknown> }> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers, credentials: 'include' })
  const payload = await response.json().catch(() => null) as { data?: T; meta?: Record<string, unknown>; code?: string; message?: string } | null
  if (!response.ok) {
    throw new ApiError(payload?.message ?? 'Request failed', response.status, payload?.code)
  }

  return { data: (payload?.data ?? payload) as T, meta: payload?.meta }
}

export async function apiDownload(path: string, init: RequestInit = {}, token?: string): Promise<Blob> {
  const headers = new Headers(init.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers, credentials: 'include' })
  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { code?: string; message?: string } | null
    throw new ApiError(payload?.message ?? 'Download failed', response.status, payload?.code)
  }
  return response.blob()
}
