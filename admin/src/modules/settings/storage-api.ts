import { apiFetch } from '@/lib/api'

export type StorageConfigField = {
  name: string
  label: string
  secret: boolean
  required: boolean
  placeholder?: string
}

export type StorageProviderDefinition = {
  code: string
  label: string
  description: string
  registration_url?: string
  adapter_available: boolean
  fields: StorageConfigField[]
}

export type StorageConnection = {
  id: number
  provider_code: string
  name: string
  enabled: boolean
  is_primary: boolean
  status: string
  public_base_url: string
  path_prefix: string
  default_visibility: string
  signed_url_ttl_seconds: number
  last_checked_at?: string
  last_success_at?: string
  last_error_code?: string
  last_error_message?: string
  config: Record<string, string>
}

export type StorageOverview = {
  providers: Record<string, StorageProviderDefinition>
  connections: StorageConnection[]
}

export type StorageStatistics = {
  connection_id: number
  provider_code: string
  name: string
  status: string
  is_primary: boolean
  object_count: number
  stored_bytes: number
  allowed_request_count: number
  estimated_bandwidth_bytes: number
  media_count: number
  object_status_counts: Record<string, number>
  variant_counts: Record<string, number>
  orphan_object_count: number
  orphan_bytes: number
  daily: StorageDailyStatistics[]
  health: StorageHealthStatistics
  data_source: string
  generated_at: string
}

export type StorageDailyStatistics = {
  date: string
  upload_count: number
  upload_bytes: number
  allowed_request_count: number
  estimated_bandwidth_bytes: number
}

export type StorageHealthStatistics = {
  last_status: string
  last_latency_ms?: number
  last_checked_at?: string
  error_count_24h: number
}

export type StorageStatisticsOverview = {
  generated_at: string
  connections: StorageStatistics[]
}

export type StorageConnectionInput = {
  name: string
  enabled: boolean
  is_primary: boolean
  public_base_url: string
  path_prefix: string
  default_visibility: string
  signed_url_ttl_seconds: number
  config: Record<string, string>
}

export function storageConnections(token: string) {
  return apiFetch<StorageOverview>('/api/v1/admin/storage/connections', {}, token)
}

export function storageStatistics(token: string) {
  return apiFetch<StorageStatisticsOverview>('/api/v1/admin/storage/statistics', {}, token)
}

export function updateStorageConnection(provider: string, input: StorageConnectionInput, token: string) {
  return apiFetch<StorageConnection>(`/api/v1/admin/storage/connections/${encodeURIComponent(provider)}`, { method: 'PUT', body: JSON.stringify(input) }, token)
}

export function testStorageConnection(provider: string, token: string) {
  return apiFetch<StorageConnection>(`/api/v1/admin/storage/connections/${encodeURIComponent(provider)}/test`, { method: 'POST', body: JSON.stringify({}) }, token)
}
