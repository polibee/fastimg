<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { Archive, Download, RefreshCw, Trash2, Upload } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetch, apiFetchBlob, apiFetchEnvelope, ApiError, errorMessageKey } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

type BackupJob = { id: number; kind: string; status: string; file_name?: string; size_bytes: number; manifest_json?: string; created_at?: string; completed_at?: string | null; error_code?: string; error_message?: string }
type Preview = { format_version: number; file_count: number; expanded_bytes: number; files: string[]; settings_excluded: string[] }

const { t } = useI18n()
const auth = useAuthStore()
const jobs = ref<BackupJob[]>([])
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const selectedFile = ref<File>()
const restoreJob = ref<BackupJob>()
const preview = ref<Preview>()
const confirmation = ref('')
const meta = ref({ page: 1, per_page: 20, total: 0, last_page: 1 })
const pageSize = ref('20')
let pollTimer: ReturnType<typeof window.setTimeout> | undefined

function handleFileChange(event: Event) {
  selectedFile.value = (event.target as HTMLInputElement).files?.[0]
}

function message(value: unknown) {
  return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('backups.unknownError')
}
function formatBytes(value: number) {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
  return `${(value / 1024 ** index).toFixed(index ? 2 : 0)} ${units[index]}`
}
function statusLabel(status: string) { return t(`backups.status.${status}`, status) }
function dateLabel(value?: string | null) { return value ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : '—' }
function jobError(job: BackupJob) {
  if (job.error_code) {
    const key = errorMessageKey(job.error_code)
    if (key !== 'errors.unknown') return t(key)
  }
  return job.error_message || t('backups.unknownError')
}
function schedulePoll() {
  if (pollTimer) window.clearTimeout(pollTimer)
  if (!jobs.value.some((job) => job.status === 'queued' || job.status === 'running' || job.status === 'restoring')) return
  pollTimer = window.setTimeout(() => {
    pollTimer = undefined
    void load(meta.value.page, false)
  }, 1500)
}

function changePageSize(value: unknown) {
  pageSize.value = String(value)
  void load(1)
}

async function load(page = 1, showLoading = true) {
  if (!auth.token) return
  if (showLoading) loading.value = true
  error.value = ''
  try {
    const response = await apiFetchEnvelope<BackupJob[]>(`/api/v1/admin/backups?page=${page}&per_page=${pageSize.value}`, {}, auth.token)
    jobs.value = response.data
    meta.value = { page: Number(response.meta?.page || page), per_page: Number(response.meta?.per_page || pageSize.value), total: Number(response.meta?.total || 0), last_page: Number(response.meta?.last_page || 1) }
  } catch (value) { error.value = message(value) } finally { if (showLoading) loading.value = false; schedulePoll() }
}
async function createBackup() {
  if (!auth.token) return
  busy.value = true; error.value = ''
  try { await apiFetch<BackupJob>('/api/v1/admin/backups', { method: 'POST', body: JSON.stringify({}) }, auth.token); await load(meta.value.page) } catch (value) { error.value = message(value) } finally { busy.value = false }
}
async function download(job: BackupJob) {
  if (!auth.token) return
  try {
    const blob = await apiFetchBlob(`/api/v1/admin/backups/${job.id}/download`, auth.token)
    const url = URL.createObjectURL(blob); const anchor = document.createElement('a'); anchor.href = url; anchor.download = job.file_name || `fastimg-backup-${job.id}.tar.zst`; anchor.click(); URL.revokeObjectURL(url)
  } catch (value) { error.value = message(value) }
}
async function deleteJob(job: BackupJob) {
  if (!auth.token || !window.confirm(t('backups.deleteConfirm'))) return
  busy.value = true
  try { await apiFetch<boolean>(`/api/v1/admin/backups/${job.id}`, { method: 'DELETE' }, auth.token); await load(meta.value.page) } catch (value) { error.value = message(value) } finally { busy.value = false }
}
async function validateUpload() {
  if (!auth.token || !selectedFile.value) return
  busy.value = true; error.value = ''; preview.value = undefined; restoreJob.value = undefined
  try {
    const form = new FormData(); form.append('file', selectedFile.value)
    const result = await apiFetch<{ job: BackupJob; preview: Preview }>('/api/v1/admin/backups/validate', { method: 'POST', body: form }, auth.token)
    restoreJob.value = result.job; preview.value = result.preview; confirmation.value = ''
  } catch (value) { error.value = message(value) } finally { busy.value = false }
}
async function restore() {
  if (!auth.token || !restoreJob.value || confirmation.value !== 'RESTORE_FASTIMG_BACKUP') return
  if (!window.confirm(t('backups.restoreConfirm'))) return
  busy.value = true; error.value = ''
  try { await apiFetch<BackupJob>('/api/v1/admin/backups/restore', { method: 'POST', body: JSON.stringify({ id: restoreJob.value.id, mode: 'new_server', confirmation: confirmation.value }) }, auth.token); await load() } catch (value) { error.value = message(value) } finally { busy.value = false }
}
onMounted(() => { void load() })
onBeforeUnmount(() => { if (pollTimer) window.clearTimeout(pollTimer) })
</script>

<template>
  <div class="grid gap-6">
    <div><h1 class="text-2xl font-semibold tracking-tight">{{ t('backups.title') }}</h1><p class="text-sm text-muted-foreground">{{ t('backups.description') }}</p></div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('backups.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <div class="grid gap-6 xl:grid-cols-2">
      <Card><CardHeader><div class="flex items-start justify-between gap-3"><div><CardTitle class="flex items-center gap-2"><Archive class="size-5" />{{ t('backups.createTitle') }}</CardTitle><CardDescription>{{ t('backups.createDescription') }}</CardDescription></div><Button variant="outline" size="icon" :disabled="loading || busy" :aria-label="t('backups.refresh')" @click="load"><RefreshCw :class="loading ? 'animate-spin' : ''" /></Button></div></CardHeader><CardContent class="grid gap-4"><Button :disabled="busy" @click="createBackup"><Archive data-icon="inline-start" />{{ t('backups.create') }}</Button><p class="text-xs leading-5 text-muted-foreground">{{ t('backups.secretNotice') }}</p></CardContent></Card>
      <Card><CardHeader><CardTitle class="flex items-center gap-2"><Upload class="size-5" />{{ t('backups.restoreTitle') }}</CardTitle><CardDescription>{{ t('backups.restoreDescription') }}</CardDescription></CardHeader><CardContent class="grid gap-4"><div class="grid gap-2"><Label for="backup-file">{{ t('backups.file') }}</Label><Input id="backup-file" type="file" accept=".zst,.tar.zst" @change="handleFileChange" /></div><Button variant="outline" :disabled="busy || !selectedFile" @click="validateUpload"><Upload data-icon="inline-start" />{{ t('backups.validate') }}</Button><div v-if="preview" class="grid gap-3 rounded-lg border bg-muted/20 p-4 text-sm"><div class="grid grid-cols-2 gap-2"><span>{{ t('backups.fileCount') }}</span><strong>{{ preview.file_count }}</strong><span>{{ t('backups.expandedBytes') }}</span><strong>{{ formatBytes(preview.expanded_bytes) }}</strong><span>{{ t('backups.excludedSecrets') }}</span><strong>{{ preview.settings_excluded.length }}</strong></div><p class="text-xs text-muted-foreground">{{ t('backups.newServerOnly') }}</p><div class="grid gap-2"><Label for="restore-confirmation">{{ t('backups.confirmation') }}</Label><Input id="restore-confirmation" v-model="confirmation" placeholder="RESTORE_FASTIMG_BACKUP" autocomplete="off" /></div><Button :disabled="busy || confirmation !== 'RESTORE_FASTIMG_BACKUP'" @click="restore">{{ t('backups.restore') }}</Button></div></CardContent></Card>
    </div>
    <Card><CardHeader><CardTitle>{{ t('backups.historyTitle') }}</CardTitle></CardHeader><CardContent><div v-if="!loading && !jobs.length" class="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">{{ t('backups.empty') }}</div><div v-else class="overflow-x-auto"><table class="w-full text-left text-sm"><thead class="border-b"><tr><th class="px-3 py-2">{{ t('backups.file') }}</th><th class="px-3 py-2">{{ t('backups.statusLabel') }}</th><th class="px-3 py-2">{{ t('backups.size') }}</th><th class="px-3 py-2">{{ t('backups.created') }}</th><th class="px-3 py-2 text-right">{{ t('backups.actions') }}</th></tr></thead><tbody><template v-for="job in jobs" :key="job.id"><tr class="border-b last:border-0"><td class="px-3 py-2 font-medium">{{ job.file_name || `#${job.id}` }}</td><td class="px-3 py-2">{{ statusLabel(job.status) }}</td><td class="px-3 py-2">{{ formatBytes(job.size_bytes) }}</td><td class="px-3 py-2 text-muted-foreground">{{ dateLabel(job.created_at) }}</td><td class="flex justify-end gap-1 px-3 py-2"><Button v-if="job.status === 'ready'" variant="ghost" size="sm" @click="download(job)"><Download data-icon="inline-start" />{{ t('backups.download') }}</Button><Button variant="ghost" size="sm" :disabled="busy" @click="deleteJob(job)"><Trash2 data-icon="inline-start" />{{ t('backups.delete') }}</Button></td></tr><tr v-if="job.status === 'failed' || job.status === 'restore_failed'" class="border-b"><td colspan="5" class="px-3 pb-3 text-sm text-destructive">{{ jobError(job) }}</td></tr></template></tbody></table></div><div v-if="!loading && meta.total > 0" class="mt-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div></CardContent></Card>
  </div>
</template>
