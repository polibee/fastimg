<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle, Link2, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetch, apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface ShareLink {
  id: number
  media_id: number
  url: string
  token_prefix: string
  status: string
  expires_at: string | null
  created_at: string
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const items = ref<ShareLink[]>([])
const loading = ref(true)
const error = ref(false)
const revokingID = ref(0)
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
    const response = await apiFetchEnvelope<ShareLink[]>(`/api/v1/share-links?page=${page}&per_page=${pageSize.value}`, {}, auth.token)
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

async function revoke(item: ShareLink) {
  if (!auth.token) return
  revokingID.value = item.id
  try {
    await apiFetch(`/api/v1/share-links/${item.id}`, { method: 'DELETE' }, auth.token)
    item.status = 'revoked'
  } catch {
    error.value = true
  } finally {
    revokingID.value = 0
  }
}

function formatDate(value: string | null) {
  if (!value) return t('member.shareLinks.never')
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

onMounted(() => void load())
</script>

<template>
  <div class="flex flex-col gap-7">
    <section>
      <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Link2 />{{ t('member.shareLinks.workspace') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.shareLinks.title') }}</h1>
      <p class="mt-3 max-w-2xl text-muted-foreground">{{ t('member.shareLinks.description') }}</p>
    </section>

    <Alert v-if="error" variant="destructive">
      <AlertCircle />
      <AlertTitle>{{ t('member.shareLinks.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.shareLinks.errors.loadFailed') }}</AlertDescription>
    </Alert>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <Skeleton v-for="index in 3" :key="index" class="h-40 rounded-lg" />
    </div>
    <div v-else-if="items.length" class="grid gap-4 md:grid-cols-2">
      <Card v-for="item in items" :key="item.id">
        <CardHeader class="gap-3">
          <div class="flex items-center justify-between gap-3">
            <CardTitle class="text-base">{{ t('member.shareLinks.mediaLabel', { id: item.media_id }) }}</CardTitle>
            <Badge :variant="item.status === 'active' ? 'secondary' : 'outline'">{{ item.status === 'active' ? t('member.shareLinks.active') : t('member.shareLinks.revoked') }}</Badge>
          </div>
          <CardDescription>{{ t('member.shareLinks.tokenPrefix', { prefix: item.token_prefix }) }}</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-between gap-4 text-sm">
          <div>
            <p class="text-muted-foreground">{{ t('member.shareLinks.expires') }}</p>
            <p class="mt-1 font-medium">{{ formatDate(item.expires_at) }}</p>
          </div>
          <Button v-if="item.status === 'active'" variant="outline" size="sm" :disabled="revokingID === item.id" @click="revoke(item)">
            <Trash2 data-icon="inline-start" />{{ t('member.shareLinks.revoke') }}
          </Button>
        </CardContent>
      </Card>
    </div>
    <Card v-else>
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon"><Link2 /></EmptyMedia>
          <EmptyTitle>{{ t('member.shareLinks.empty') }}</EmptyTitle>
          <EmptyDescription>{{ t('member.shareLinks.emptyDescription') }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </Card>
    <div v-if="!loading && meta.total > 0" class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
  </div>
</template>
