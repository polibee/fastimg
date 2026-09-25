<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Activity, RefreshCw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Input } from '@/components/ui/input'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi, type MediaAccessLog, type ResourceListMeta } from '@/generated/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const entries = ref<MediaAccessLog[]>([])
const meta = ref<ResourceListMeta>({ page: 1, per_page: 20, total: 0, last_page: 1 })
const pageSize = ref('20')
const mediaFilter = ref('')
const resultFilter = ref('all')
const modeFilter = ref('all')
const loading = ref(true)
const error = ref('')

function formatDate(value: string) {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function localizedError(value: unknown) {
  return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('errors.unknown')
}

async function load(page = 1) {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ page: String(page), per_page: pageSize.value })
    if (mediaFilter.value.trim()) params.set('media_id', mediaFilter.value.trim())
    if (resultFilter.value !== 'all') params.set('result', resultFilter.value)
    if (modeFilter.value !== 'all') params.set('delivery_mode', modeFilter.value)
    const response = await generatedApi.mediaAccessLogs(auth.token, params)
    entries.value = response.data
    meta.value = response.meta
  } catch (value) {
    entries.value = []
    error.value = localizedError(value)
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  mediaFilter.value = ''
  resultFilter.value = 'all'
  modeFilter.value = 'all'
  void load(1)
}

function changePageSize(value: unknown) {
  pageSize.value = String(value)
  void load(1)
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <div class="flex items-center gap-2"><Activity class="size-5 text-primary" /><h1 class="text-2xl font-semibold tracking-tight">{{ t('auth.mediaAccessLogs') }}</h1></div>
        <p class="mt-1 text-sm text-muted-foreground">{{ t('auth.mediaAccessDescription') }}</p>
      </div>
      <Button variant="outline" size="sm" :disabled="loading" @click="load(meta.page)"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card>
      <CardHeader>
        <CardTitle>{{ t('auth.mediaAccessLogs') }}</CardTitle>
        <CardDescription>{{ t('auth.auditRecentDescription', { count: meta.total }) }}</CardDescription>
        <form class="flex flex-wrap items-center gap-2 pt-2" @submit.prevent="load(1)">
          <Input v-model="mediaFilter" class="w-28" :placeholder="t('auth.mediaAccessMedia')" :aria-label="t('auth.mediaAccessMedia')" />
          <Select v-model="resultFilter"><SelectTrigger class="w-32" :aria-label="t('auth.mediaAccessResult')"><SelectValue :placeholder="t('auth.mediaAccessResult')" /></SelectTrigger><SelectContent><SelectItem value="all">{{ t('auth.mediaAccessAll') }}</SelectItem><SelectItem value="allowed">{{ t('auth.mediaAccessAllowed') }}</SelectItem><SelectItem value="denied">{{ t('auth.mediaAccessDenied') }}</SelectItem></SelectContent></Select>
          <Select v-model="modeFilter"><SelectTrigger class="w-32" :aria-label="t('auth.mediaAccessMode')"><SelectValue :placeholder="t('auth.mediaAccessMode')" /></SelectTrigger><SelectContent><SelectItem value="all">{{ t('auth.mediaAccessAll') }}</SelectItem><SelectItem v-for="mode in ['off', 'referer', 'signed', 'hybrid', 'share']" :key="mode" :value="mode">{{ mode }}</SelectItem></SelectContent></Select>
          <Button type="submit" size="sm">{{ t('resource.search') }}</Button><Button type="button" variant="ghost" size="sm" @click="resetFilters">{{ t('resource.clearSelection') }}</Button>
        </form>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3"><Skeleton v-for="item in 5" :key="item" class="h-10" /></div>
        <Empty v-else-if="!entries.length"><EmptyHeader><EmptyTitle>{{ t('states.emptyTitle') }}</EmptyTitle><EmptyDescription>{{ t('auth.mediaAccessEmpty') }}</EmptyDescription></EmptyHeader></Empty>
        <Table v-else><TableHeader><TableRow><TableHead>{{ t('auth.mediaAccessMedia') }}</TableHead><TableHead>{{ t('auth.mediaAccessVariant') }}</TableHead><TableHead>{{ t('auth.mediaAccessMode') }}</TableHead><TableHead>{{ t('auth.mediaAccessResult') }}</TableHead><TableHead>{{ t('auth.mediaAccessReferer') }}</TableHead><TableHead>{{ t('auth.mediaAccessTime') }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="entry in entries" :key="entry.id"><TableCell class="font-medium">{{ entry.media_asset_id }}</TableCell><TableCell>{{ entry.variant }}</TableCell><TableCell>{{ entry.delivery_mode }}</TableCell><TableCell>{{ entry.result }}</TableCell><TableCell>{{ entry.referer_host || '—' }}</TableCell><TableCell>{{ formatDate(entry.accessed_at) }}</TableCell></TableRow></TableBody></Table>
        <div v-if="!loading && meta.total > 0" class="mt-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
      </CardContent>
    </Card>
  </div>
</template>
