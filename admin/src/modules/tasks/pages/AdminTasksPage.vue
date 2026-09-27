<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ListTodo, RefreshCw, RotateCcw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi, type AdminTask, type ResourceListMeta } from '@/generated/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const entries = ref<AdminTask[]>([])
const meta = ref<ResourceListMeta>({ page: 1, per_page: 20, total: 0, last_page: 1 })
const pageSize = ref('20')
const loading = ref(true)
const error = ref('')
const retrying = ref('')

function formatDate(value: string) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

function localizedError(value: unknown) {
  return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('tasks.loadFailed')
}

async function load(page = 1) {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    const response = await generatedApi.failedTasks(auth.token, new URLSearchParams({ page: String(page), per_page: pageSize.value }))
    entries.value = response.data
    meta.value = response.meta
  } catch (value) {
    entries.value = []
    error.value = localizedError(value)
  } finally {
    loading.value = false
  }
}

async function retry(task: AdminTask) {
  if (!auth.token || retrying.value) return
  if (!window.confirm(t('tasks.retryConfirm'))) return
  retrying.value = task.uuid
  error.value = ''
  try {
    await generatedApi.retryFailedTask(task.uuid, auth.token)
    await load(meta.value.page)
  } catch (value) {
    error.value = value instanceof ApiError ? t(errorMessageKey(value.code)) : t('tasks.retryFailed')
  } finally {
    retrying.value = ''
  }
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
        <div class="flex items-center gap-2"><ListTodo class="size-5 text-primary" /><h1 class="text-2xl font-semibold tracking-tight">{{ t('tasks.title') }}</h1></div>
        <p class="mt-1 text-sm text-muted-foreground">{{ t('tasks.description') }}</p>
      </div>
      <Button variant="outline" size="sm" :disabled="loading" @click="load(meta.page)"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card>
      <CardHeader><CardTitle>{{ t('tasks.title') }}</CardTitle><CardDescription>{{ t('resource.total', { count: meta.total }) }}</CardDescription></CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3"><Skeleton v-for="item in 5" :key="item" class="h-10" /></div>
        <Empty v-else-if="!entries.length"><EmptyHeader><EmptyTitle>{{ t('tasks.empty') }}</EmptyTitle><EmptyDescription>{{ t('tasks.emptyDescription') }}</EmptyDescription></EmptyHeader></Empty>
        <Table v-else><TableHeader><TableRow><TableHead>{{ t('tasks.job') }}</TableHead><TableHead>{{ t('tasks.connection') }}</TableHead><TableHead>{{ t('tasks.queue') }}</TableHead><TableHead>{{ t('tasks.failedAt') }}</TableHead><TableHead class="text-right">{{ t('tasks.retry') }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="task in entries" :key="task.uuid"><TableCell class="max-w-[24rem] truncate font-medium" :title="task.uuid">{{ task.signature || t('tasks.unknown') }}</TableCell><TableCell>{{ task.connection }}</TableCell><TableCell>{{ task.queue }}</TableCell><TableCell>{{ formatDate(task.failed_at) }}</TableCell><TableCell class="text-right"><Button v-if="auth.can('admin.tasks.retry')" variant="outline" size="sm" :disabled="retrying === task.uuid" @click="retry(task)"><RotateCcw class="mr-2 size-4" />{{ retrying === task.uuid ? t('tasks.retrying') : t('tasks.retry') }}</Button></TableCell></TableRow></TableBody></Table>
        <div v-if="!loading && meta.total > 0" class="mt-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
      </CardContent>
    </Card>
  </div>
</template>
