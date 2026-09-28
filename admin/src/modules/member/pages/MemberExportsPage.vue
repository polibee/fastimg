<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Archive, Download, RefreshCw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetch, apiFetchBlob, apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface ExportJob { id: number; status: string; file_name?: string; size_bytes?: number; media_count?: number; created_at?: string; error_code?: string }
interface ExportMeta { page: number; per_page: number; total: number; last_page: number }
const { t } = useI18n(); const auth = useAuthStore(); const jobs = ref<ExportJob[]>([]); const loading = ref(true); const busy = ref(false); const error = ref(''); const pageSize = ref('20'); const meta = ref<ExportMeta>({ page: 1, per_page: 20, total: 0, last_page: 0 })
function date(value?: string) { return value ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : '—' }
function bytes(value = 0) { return value ? `${(value / 1024 / 1024).toFixed(2)} MB` : '—' }
async function load(page = meta.value.page) { if (!auth.token) return; loading.value = true; try { const response = await apiFetchEnvelope<ExportJob[]>(`/api/v1/me/exports?page=${page}&per_page=${pageSize.value}`, {}, auth.token); jobs.value = response.data; if (response.meta) meta.value = response.meta as unknown as ExportMeta } catch { error.value = 'exports.errors.loadFailed' } finally { loading.value = false } }
async function changePageSize(value: unknown) { pageSize.value = String(value); await load(1) }
async function create() { if (!auth.token) return; busy.value = true; error.value = ''; try { await apiFetch<ExportJob>('/api/v1/me/exports', { method: 'POST', body: JSON.stringify({}) }, auth.token); await load(1) } catch { error.value = 'exports.errors.createFailed' } finally { busy.value = false } }
async function download(job: ExportJob) { if (!auth.token) return; try { const blob = await apiFetchBlob(`/api/v1/me/exports/${job.id}/download`, auth.token); const url = URL.createObjectURL(blob); const anchor = document.createElement('a'); anchor.href = url; anchor.download = job.file_name || `fastimg-export-${job.id}.zip`; anchor.click(); URL.revokeObjectURL(url) } catch { error.value = 'exports.errors.downloadFailed' } }
onMounted(load)
</script>

<template>
  <div class="grid gap-6"><div class="flex flex-wrap items-start justify-between gap-4"><div><p class="text-sm font-medium text-primary">{{ t('exports.eyebrow') }}</p><h1 class="mt-2 text-3xl font-semibold tracking-tight">{{ t('exports.title') }}</h1><p class="mt-2 text-muted-foreground">{{ t('exports.description') }}</p></div><div class="flex gap-2"><Button variant="outline" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('exports.refresh') }}</Button><Button :disabled="busy" @click="create"><Archive class="mr-2 size-4" />{{ busy ? t('exports.creating') : t('exports.create') }}</Button></div></div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('exports.errorTitle') }}</AlertTitle><AlertDescription>{{ t(error) }}</AlertDescription></Alert>
    <Card><CardHeader><CardTitle>{{ t('exports.jobsTitle') }}</CardTitle><CardDescription>{{ t('exports.jobsDescription') }}</CardDescription></CardHeader><CardContent><div v-if="loading" class="text-sm text-muted-foreground">{{ t('exports.loading') }}</div><div v-else-if="!jobs.length" class="py-8 text-center text-sm text-muted-foreground">{{ t('exports.empty') }}</div><div v-else class="grid gap-3"><div v-for="job in jobs" :key="job.id" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border px-4 py-3"><div><p class="font-medium">{{ job.file_name || t('exports.pendingFile') }}</p><p class="text-xs text-muted-foreground">{{ t(`exports.status.${job.status}`, job.status) }} · {{ job.media_count ?? 0 }} {{ t('exports.mediaCount') }} · {{ bytes(job.size_bytes) }} · {{ date(job.created_at) }}</p></div><Button v-if="job.status === 'ready'" variant="outline" size="sm" @click="download(job)"><Download class="mr-2 size-4" />{{ t('exports.download') }}</Button></div></div><div v-if="!loading && meta.total > 0" class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div></CardContent></Card>
  </div>
</template>
