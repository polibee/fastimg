<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle, Copy, KeyRound, RefreshCw, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetch, apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface PersonalToken {
  id: number
  name: string
  prefix: string
  scopes: string[]
  status: string
  expires_at: string | null
  created_at: string
  last_used_at: string | null
  usage_count: number
}

interface CreatedToken extends PersonalToken {
  token: string
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const items = ref<PersonalToken[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref(false)
const name = ref('')
const expiryMode = ref<'permanent' | 'custom'>('permanent')
const expiresAt = ref('')
const createdToken = ref<CreatedToken | null>(null)
const busyID = ref(0)
const meta = ref({ page: 1, per_page: 20, total: 0 })
const pageSize = ref('20')
const expiryMinimum = computed(() => {
  const value = new Date(Date.now() + 60_000)
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${value.getFullYear()}-${pad(value.getMonth() + 1)}-${pad(value.getDate())}T${pad(value.getHours())}:${pad(value.getMinutes())}`
})

async function load(page = 1) {
  if (!auth.token) {
    loading.value = false
    error.value = true
    return
  }
  loading.value = true
  error.value = false
  try {
    const response = await apiFetchEnvelope<PersonalToken[]>(`/api/v1/tokens?page=${page}&per_page=${pageSize.value}`, {}, auth.token)
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

async function createToken() {
  if (!auth.token || !name.value.trim() || (expiryMode.value === 'custom' && !expiresAt.value)) return
  saving.value = true
  error.value = false
  try {
    const expiry = expiryMode.value === 'permanent' ? null : new Date(expiresAt.value).toISOString()
    const response = await apiFetch<CreatedToken>('/api/v1/tokens', {
      method: 'POST',
      body: JSON.stringify({
        name: name.value.trim(),
        scopes: [],
        expires_at: expiry,
      }),
    }, auth.token)
    createdToken.value = response
    name.value = ''
    expiryMode.value = 'permanent'
    expiresAt.value = ''
    await load()
  } catch {
    error.value = true
  } finally {
    saving.value = false
  }
}

async function deleteToken(item: PersonalToken) {
  if (!auth.token || !window.confirm(t('member.tokens.confirmDelete'))) return
  busyID.value = item.id
  try {
    await apiFetch(`/api/v1/tokens/${item.id}`, { method: 'DELETE' }, auth.token)
    items.value = items.value.filter((candidate) => candidate.id !== item.id)
  } catch {
    error.value = true
  } finally {
    busyID.value = 0
  }
}

async function rotateToken(item: PersonalToken) {
  if (!auth.token || !window.confirm(t('member.tokens.confirmRotate'))) return
  busyID.value = item.id
  try {
    createdToken.value = await apiFetch<CreatedToken>(`/api/v1/tokens/${item.id}/rotate`, { method: 'POST' }, auth.token)
    await load()
  } catch {
    error.value = true
  } finally {
    busyID.value = 0
  }
}

async function copyCreatedToken() {
  if (!createdToken.value) return
  await navigator.clipboard.writeText(createdToken.value.token)
}

function formatDate(value: string | null) {
  if (!value) return t('member.tokens.never')
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

onMounted(() => void load())
</script>

<template>
  <div class="flex flex-col gap-7">
    <section>
      <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><KeyRound />{{ t('member.tokens.workspace') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.tokens.title') }}</h1>
      <p class="mt-3 max-w-2xl text-muted-foreground">{{ t('member.tokens.description') }}</p>
    </section>

    <Alert v-if="error" variant="destructive">
      <AlertCircle />
      <AlertTitle>{{ t('member.tokens.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.tokens.errors.operationFailed') }}</AlertDescription>
    </Alert>

    <Alert v-if="createdToken" class="border-primary/30">
      <KeyRound />
      <AlertTitle>{{ t('member.tokens.createdTitle') }}</AlertTitle>
      <AlertDescription class="space-y-3">
        <p>{{ t('member.tokens.createdDescription') }}</p>
        <div class="flex gap-2">
          <Input readonly :model-value="createdToken.token" aria-label="Personal API Token" />
          <Button variant="outline" size="icon" :aria-label="t('member.tokens.copy')" @click="copyCreatedToken"><Copy /></Button>
        </div>
      </AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle>{{ t('member.tokens.createTitle') }}</CardTitle>
        <CardDescription>{{ t('member.tokens.createDescription') }}</CardDescription>
      </CardHeader>
    <CardContent class="grid gap-5 md:grid-cols-[1fr_180px_240px_auto] md:items-end">
        <label class="grid gap-2 text-sm font-medium">
          {{ t('member.tokens.name') }}
          <Input v-model="name" :placeholder="t('member.tokens.namePlaceholder')" maxlength="120" />
        </label>
        <label class="grid gap-2 text-sm font-medium">
          {{ t('member.tokens.expiryMode') }}
          <select v-model="expiryMode" class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
            <option value="permanent">{{ t('member.tokens.permanent') }}</option>
            <option value="custom">{{ t('member.tokens.customExpiry') }}</option>
          </select>
        </label>
        <label v-if="expiryMode === 'custom'" class="grid gap-2 text-sm font-medium">
          {{ t('member.tokens.expires') }}
          <Input v-model="expiresAt" type="datetime-local" :min="expiryMinimum" />
        </label>
        <div v-else class="hidden md:block" aria-hidden="true" />
        <Button :disabled="saving || !name.trim() || (expiryMode === 'custom' && !expiresAt)" @click="createToken">{{ saving ? t('member.tokens.creating') : t('member.tokens.create') }}</Button>
        <div class="md:col-span-3 grid gap-2">
          <p class="text-sm font-medium">{{ t('member.tokens.fixedScopesTitle') }}</p>
          <p class="text-xs text-muted-foreground">{{ t('member.tokens.fixedScopes') }}</p>
        </div>
      </CardContent>
    </Card>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2">
      <Skeleton v-for="index in 2" :key="index" class="h-44 rounded-lg" />
    </div>
    <div v-else-if="items.length" class="grid gap-4 md:grid-cols-2">
      <Card v-for="item in items" :key="item.id">
        <CardHeader class="gap-3">
          <div class="flex items-center justify-between gap-3">
            <CardTitle class="text-base">{{ item.name }}</CardTitle>
            <Badge :variant="item.status === 'active' ? 'secondary' : 'outline'">{{ item.status === 'active' ? t('member.tokens.active') : t('member.tokens.revoked') }}</Badge>
          </div>
          <CardDescription>{{ item.prefix }}… · {{ t('member.tokens.createdAt', { date: formatDate(item.created_at) }) }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 text-sm">
          <div class="flex flex-wrap gap-2"><Badge v-for="scope in item.scopes" :key="scope" variant="outline">{{ scope }}</Badge></div>
          <div class="grid grid-cols-2 gap-3 text-muted-foreground">
            <span>{{ t('member.tokens.expires') }}: <strong class="font-medium text-foreground">{{ formatDate(item.expires_at) }}</strong></span>
            <span>{{ t('member.tokens.lastUsed') }}: <strong class="font-medium text-foreground">{{ formatDate(item.last_used_at) }}</strong></span>
          </div>
          <div class="flex justify-end gap-2">
            <Button v-if="item.status === 'active'" variant="outline" size="sm" :disabled="busyID === item.id" @click="rotateToken(item)"><RefreshCw data-icon="inline-start" />{{ t('member.tokens.rotate') }}</Button>
            <Button variant="outline" size="sm" :disabled="busyID === item.id" @click="deleteToken(item)"><Trash2 data-icon="inline-start" />{{ t('member.tokens.delete') }}</Button>
          </div>
        </CardContent>
      </Card>
    </div>
    <Card v-else>
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon"><KeyRound /></EmptyMedia>
          <EmptyTitle>{{ t('member.tokens.empty') }}</EmptyTitle>
          <EmptyDescription>{{ t('member.tokens.emptyDescription') }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </Card>
    <div v-if="!loading && meta.total > 0" class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
  </div>
</template>
