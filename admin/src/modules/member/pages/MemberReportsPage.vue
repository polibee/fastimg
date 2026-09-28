<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle, Flag } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface Report {
  id: number
  media_asset_id: number
  reason: string
  description: string | null
  status: string
  resolution: string | null
  created_at: string
  resolved_at: string | null
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const items = ref<Report[]>([])
const loading = ref(true)
const error = ref(false)
const meta = ref({ page: 1, per_page: 20, total: 0 })
const pageSize = ref('20')

async function load(page = 1) {
  if (!auth.token) {
    loading.value = false
    error.value = true
    return
  }
  loading.value = true
  error.value = false
  try {
    const response = await apiFetchEnvelope<Report[]>(`/api/v1/me/reports?page=${page}&per_page=${pageSize.value}`, {}, auth.token)
    items.value = response.data
    meta.value = { page: Number(response.meta?.page || page), per_page: Number(response.meta?.per_page || pageSize.value), total: Number(response.meta?.total || 0) }
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function changePageSize(value: unknown) {
  pageSize.value = String(value)
  void load(1)
}

function formatDate(value: string | null) {
  if (!value) return t('member.reports.notResolved')
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function statusLabel(value: string) {
  return t(`member.reports.status.${value}`, value)
}

onMounted(() => void load())
</script>

<template>
  <div class="flex flex-col gap-7">
    <section>
      <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Flag />{{ t('member.reports.workspace') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.reports.title') }}</h1>
      <p class="mt-3 max-w-2xl text-muted-foreground">{{ t('member.reports.description') }}</p>
    </section>
    <Alert v-if="error" variant="destructive">
      <AlertCircle /><AlertTitle>{{ t('member.reports.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.reports.errorDescription') }}</AlertDescription>
    </Alert>
    <div v-if="loading" class="grid gap-4 md:grid-cols-2"><Skeleton v-for="index in 3" :key="index" class="h-44 rounded-lg" /></div>
    <div v-else-if="items.length" class="grid gap-4 md:grid-cols-2">
      <Card v-for="item in items" :key="item.id">
        <CardHeader class="gap-3">
          <div class="flex items-center justify-between gap-3"><CardTitle class="text-base">{{ t('member.reports.mediaId', { id: item.media_asset_id }) }}</CardTitle><Badge :variant="item.status === 'resolved' ? 'secondary' : 'outline'">{{ statusLabel(item.status) }}</Badge></div>
          <CardDescription>{{ t('member.reports.reason') }}：{{ item.reason }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 text-sm">
          <p v-if="item.description" class="text-muted-foreground">{{ item.description }}</p>
          <dl class="grid gap-2 text-muted-foreground sm:grid-cols-2"><div><dt>{{ t('member.reports.submittedAt') }}</dt><dd class="font-medium text-foreground">{{ formatDate(item.created_at) }}</dd></div><div><dt>{{ t('member.reports.resolvedAt') }}</dt><dd class="font-medium text-foreground">{{ formatDate(item.resolved_at) }}</dd></div></dl>
          <p v-if="item.resolution" class="border-t pt-3"><span class="text-muted-foreground">{{ t('member.reports.resolution') }}：</span>{{ item.resolution }}</p>
        </CardContent>
      </Card>
    </div>
    <Card v-else><Empty><EmptyHeader><EmptyMedia variant="icon"><Flag /></EmptyMedia><EmptyTitle>{{ t('member.reports.empty') }}</EmptyTitle><EmptyDescription>{{ t('member.reports.emptyDescription') }}</EmptyDescription></EmptyHeader></Empty></Card>
    <div v-if="!loading && meta.total > 0" class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
      <p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p>
      <Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select>
      <Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination>
    </div>
  </div>
</template>
