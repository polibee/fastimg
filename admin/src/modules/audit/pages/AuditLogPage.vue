<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RefreshCw, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Input } from '@/components/ui/input'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi, type AuditLog, type ResourceListMeta } from '@/generated/api'
import { auditOutcome, formatAuditMetadata, parseAuditMetadata } from '@/lib/audit-log'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const entries = ref<AuditLog[]>([])
const meta = ref<ResourceListMeta>({ page: 1, per_page: 20, total: 0, last_page: 1 })
const pageSize = ref('20')
const actionFilter = ref('all')
const userFilter = ref('')
const selectedEntry = ref<AuditLog | null>(null)
const loading = ref(true)
const error = ref('')
const cleanupOpen = ref(false)
const cleanupMode = ref<'retention' | 'selected' | 'filtered' | 'all'>('retention')
const cleanupDays = ref('365')
const cleanupConfirmation = ref('')
const selectedIds = ref<number[]>([])
const cleaning = ref(false)
const cleanupResult = ref<number | null>(null)
const canCleanup = computed(() => auth.can('admin.settings.manage'))
const auditActionOptions = ['auth.login', 'auth.refresh', 'auth.logout', 'auth.logout_all', 'billing.payment.create', 'billing.webhook.receive', 'media.upload', 'settings.update', 'moderation.report.create', 'admin.create', 'admin.update', 'admin.delete']

function localizedError(value: unknown) {
  return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('errors.unknown')
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function actionLabel(action: string) {
  const key = `auth.auditActions.${action.replaceAll('.', '_')}`
  const translated = t(key)
  return translated === key ? action : translated
}

function outcomeLabel(entry: AuditLog) {
  const outcome = auditOutcome(entry.metadata)
  return outcome === 'error' ? t('auth.auditOutcomeError') : outcome === 'success' ? t('auth.auditOutcomeSuccess') : t('auth.auditOutcomeUnknown')
}

function outcomeVariant(entry: AuditLog) {
  return auditOutcome(entry.metadata) === 'error' ? 'destructive' : 'secondary'
}

function auditSummary(entry: AuditLog) {
  const metadata = parseAuditMetadata(entry.metadata)
  const status = typeof metadata.response === 'object' && metadata.response !== null
    ? (metadata.response as Record<string, unknown>).status
    : undefined
  return t(auditOutcome(entry.metadata) === 'error' ? 'auth.auditSummaryError' : 'auth.auditSummarySuccess', { action: actionLabel(entry.action), status: status || '' })
}

async function load(page = 1) {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ page: String(page), per_page: pageSize.value })
    if (actionFilter.value !== 'all') params.set('action', actionFilter.value)
    if (userFilter.value.trim()) params.set('user_id', userFilter.value.trim())
    const response = await generatedApi.auditLogs(auth.token, params)
    entries.value = response.data
    meta.value = response.meta
    selectedIds.value = []
  } catch (value) {
    entries.value = []
    error.value = localizedError(value)
  } finally {
    loading.value = false
  }
}

function changePageSize(value: unknown) {
  pageSize.value = String(value)
  meta.value.per_page = Number(pageSize.value)
  void load(1)
}

function submitFilters() { void load(1) }
function resetFilters() {
  actionFilter.value = 'all'
  userFilter.value = ''
  void load(1)
}

const selectedCount = computed(() => selectedIds.value.length)
const hasActiveFilter = computed(() => actionFilter.value !== 'all' || userFilter.value.trim() !== '')
const allVisibleSelected = computed(() => entries.value.length > 0 && entries.value.every((entry) => selectedIds.value.includes(entry.id)))

function toggleEntry(id: number) {
  selectedIds.value = selectedIds.value.includes(id)
    ? selectedIds.value.filter((value) => value !== id)
    : [...selectedIds.value, id]
}

function toggleAllVisible() {
  selectedIds.value = allVisibleSelected.value ? [] : entries.value.map((entry) => entry.id)
}

function openCleanup(mode: 'retention' | 'selected' | 'filtered' | 'all' = 'retention') {
  cleanupMode.value = mode
  cleanupConfirmation.value = ''
  cleanupOpen.value = true
}

async function cleanupLogs() {
  if (!auth.token) return
  const days = Number(cleanupDays.value)
  if (cleanupMode.value === 'retention' && (!Number.isInteger(days) || days < 1 || days > 3650)) {
    error.value = t('auth.auditRetentionInvalid')
    return
  }
  if (cleanupMode.value !== 'retention' && cleanupConfirmation.value !== 'DELETE') {
    error.value = t('auth.auditCleanupConfirmationInvalid')
    return
  }
  if (cleanupMode.value === 'selected' && selectedIds.value.length === 0) {
    error.value = t('auth.auditCleanupSelectionRequired')
    return
  }
  cleaning.value = true
  error.value = ''
  try {
    const request = cleanupMode.value === 'retention'
      ? { mode: 'retention' as const, retention_days: days }
      : cleanupMode.value === 'selected'
        ? { mode: 'selected' as const, ids: selectedIds.value, confirmation: cleanupConfirmation.value }
        : cleanupMode.value === 'filtered'
          ? { mode: 'filtered' as const, action: actionFilter.value === 'all' ? '' : actionFilter.value, user_id: userFilter.value.trim(), confirmation: cleanupConfirmation.value }
          : { mode: 'all' as const, confirmation: cleanupConfirmation.value }
    const result = await generatedApi.cleanupAuditLogs(request, auth.token)
    cleanupResult.value = result.deleted
    selectedIds.value = []
    cleanupConfirmation.value = ''
    cleanupOpen.value = false
    await load(1)
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    cleaning.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">{{ t('auth.auditLogs') }}</h1>
        <p class="text-sm text-muted-foreground">{{ t('auth.auditDescription') }}</p>
      </div>
      <div class="flex flex-wrap gap-2"><Button v-if="canCleanup" variant="outline" :disabled="loading" @click="openCleanup()"><Trash2 data-icon="inline-start" />{{ t('auth.auditCleanup') }}</Button><Button v-if="canCleanup" variant="outline" :disabled="loading || selectedCount === 0" @click="openCleanup('selected')">{{ t('auth.auditCleanupSelected', { count: selectedCount }) }}</Button><Button v-if="canCleanup" variant="outline" :disabled="loading || !hasActiveFilter" @click="openCleanup('filtered')">{{ t('auth.auditCleanupFiltered') }}</Button><Button variant="outline" :disabled="loading" @click="load"><RefreshCw data-icon="inline-start" />{{ t('resource.refresh') }}</Button></div>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-center sm:justify-between"><div><CardTitle>{{ t('auth.auditRecent') }}</CardTitle><CardDescription>{{ t('auth.auditRecentDescription', { count: meta.total }) }}</CardDescription></div><form class="flex w-full flex-wrap gap-2 sm:w-auto" @submit.prevent="submitFilters"><Select :model-value="actionFilter" @update:model-value="(value) => { actionFilter = String(value); submitFilters() }"><SelectTrigger class="w-48" :aria-label="t('auth.auditActionFilter')"><SelectValue :placeholder="t('auth.auditActionAll')" /></SelectTrigger><SelectContent><SelectItem value="all">{{ t('auth.auditActionAll') }}</SelectItem><SelectItem v-for="action in auditActionOptions" :key="action" :value="action">{{ actionLabel(action) }}</SelectItem></SelectContent></Select><Input v-model="userFilter" class="w-28" :placeholder="t('auth.auditUserFilter')" :aria-label="t('auth.auditUserFilter')" /><Button type="submit" size="sm">{{ t('resource.search') }}</Button><Button type="button" variant="ghost" size="sm" @click="resetFilters">{{ t('resource.clearSelection') }}</Button></form></CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3"><Skeleton v-for="item in 5" :key="item" class="h-10" /></div>
        <Empty v-else-if="!entries.length"><EmptyHeader><EmptyTitle>{{ t('states.emptyTitle') }}</EmptyTitle><EmptyDescription>{{ t('auth.auditEmpty') }}</EmptyDescription></EmptyHeader></Empty>
        <Table v-else><TableHeader><TableRow><TableHead class="w-10"><input type="checkbox" :checked="allVisibleSelected" :aria-label="t('auth.auditSelectAll')" @click.stop="toggleAllVisible" /></TableHead><TableHead>{{ t('auth.auditAction') }}</TableHead><TableHead>{{ t('auth.auditOutcome') }}</TableHead><TableHead>{{ t('auth.auditUser') }}</TableHead><TableHead>{{ t('auth.auditTime') }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="entry in entries" :key="entry.id" class="cursor-pointer" @click="selectedEntry = entry"><TableCell @click.stop><input type="checkbox" :checked="selectedIds.includes(entry.id)" :aria-label="t('auth.auditSelectRow', { id: entry.id })" @change="toggleEntry(entry.id)" /></TableCell><TableCell><div class="font-medium">{{ actionLabel(entry.action) }}</div><div class="text-xs text-muted-foreground">{{ auditSummary(entry) }}</div></TableCell><TableCell><Badge :variant="outcomeVariant(entry)">{{ outcomeLabel(entry) }}</Badge></TableCell><TableCell>{{ entry.user_id || t('auth.auditSystemUser') }}</TableCell><TableCell>{{ formatDate(entry.created_at) }}</TableCell></TableRow></TableBody></Table>
        <div v-if="!loading && meta.total > 0" class="mt-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><div class="flex items-center gap-2"><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select></div><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
      </CardContent>
    </Card>
    <Dialog :open="selectedEntry !== null" @update:open="(open) => { if (!open) selectedEntry = null }"><DialogContent class="max-h-[calc(100vh-2rem)] overflow-y-auto"><DialogHeader><DialogTitle>{{ selectedEntry ? actionLabel(selectedEntry.action) : t('auth.auditDetail') }}</DialogTitle><DialogDescription>{{ selectedEntry ? auditSummary(selectedEntry) : '' }} · {{ selectedEntry ? formatDate(selectedEntry.created_at) : '' }}</DialogDescription></DialogHeader><dl v-if="selectedEntry" class="grid gap-3 text-sm"><div class="flex justify-between gap-4"><dt class="text-muted-foreground">{{ t('auth.auditUser') }}</dt><dd>{{ selectedEntry.user_id || t('auth.auditSystemUser') }}</dd></div><div class="flex justify-between gap-4"><dt class="text-muted-foreground">{{ t('auth.auditOutcome') }}</dt><dd><Badge :variant="outcomeVariant(selectedEntry)">{{ outcomeLabel(selectedEntry) }}</Badge></dd></div><details v-if="formatAuditMetadata(selectedEntry.metadata)" class="rounded-md bg-muted p-3"><summary class="cursor-pointer text-sm font-medium">{{ t('auth.auditTechnicalDetails') }}</summary><pre class="mt-3 max-h-[50vh] overflow-auto whitespace-pre-wrap break-words text-xs">{{ formatAuditMetadata(selectedEntry.metadata) }}</pre></details><p v-else class="text-muted-foreground">{{ t('auth.auditNoMetadata') }}</p></dl></DialogContent></Dialog>
    <Dialog v-model:open="cleanupOpen"><DialogContent class="max-h-[calc(100vh-2rem)] overflow-y-auto"><DialogHeader><DialogTitle>{{ t('auth.auditCleanup') }}</DialogTitle><DialogDescription>{{ t(cleanupMode === 'retention' ? 'auth.auditCleanupDescription' : 'auth.auditCleanupDestructiveDescription') }}</DialogDescription></DialogHeader><div class="grid gap-3"><div class="grid gap-2"><label class="text-sm font-medium">{{ t('auth.auditCleanupMode') }}</label><Select v-model="cleanupMode"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="retention">{{ t('auth.auditCleanupRetentionMode') }}</SelectItem><SelectItem value="selected">{{ t('auth.auditCleanupSelectedMode', { count: selectedCount }) }}</SelectItem><SelectItem value="filtered">{{ t('auth.auditCleanupFilteredMode') }}</SelectItem><SelectItem value="all">{{ t('auth.auditCleanupAllMode') }}</SelectItem></SelectContent></Select></div><div v-if="cleanupMode === 'retention'" class="grid gap-2"><label for="audit-retention-days" class="text-sm font-medium">{{ t('auth.auditRetentionDays') }}</label><Input id="audit-retention-days" v-model="cleanupDays" type="number" min="1" max="3650" /></div><Alert v-else variant="destructive"><AlertTitle>{{ t('auth.auditCleanupAllWarningTitle') }}</AlertTitle><AlertDescription>{{ t(cleanupMode === 'selected' ? 'auth.auditCleanupSelectedWarning' : cleanupMode === 'filtered' ? 'auth.auditCleanupFilteredWarning' : 'auth.auditCleanupAllWarning') }}</AlertDescription></Alert><div v-if="cleanupMode !== 'retention'" class="grid gap-2"><label for="audit-cleanup-confirmation" class="text-sm font-medium">{{ t('auth.auditCleanupConfirmationLabel') }}</label><Input id="audit-cleanup-confirmation" v-model="cleanupConfirmation" :placeholder="t('auth.auditCleanupConfirmationPlaceholder')" autocomplete="off" /><p class="text-xs text-muted-foreground">{{ t('auth.auditCleanupConfirmationHint') }}</p></div></div><DialogFooter><Button variant="outline" @click="cleanupOpen = false">{{ t('resource.cancel') }}</Button><Button :disabled="cleaning || (cleanupMode !== 'retention' && cleanupConfirmation !== 'DELETE') || (cleanupMode === 'selected' && selectedCount === 0)" @click="cleanupLogs">{{ t(cleanupMode === 'all' ? 'auth.auditCleanupAllConfirm' : 'auth.auditCleanupConfirm') }}</Button></DialogFooter></DialogContent></Dialog>
    <Alert v-if="cleanupResult !== null"><AlertTitle>{{ t('auth.auditCleanupDone') }}</AlertTitle><AlertDescription>{{ t('auth.auditCleanupResult', { count: cleanupResult }) }}</AlertDescription></Alert>
  </div>
</template>
