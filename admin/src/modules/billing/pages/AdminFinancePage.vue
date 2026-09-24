<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RefreshCw, WalletCards } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import type { BillingAdminKind } from '@/modules/billing/api'

const props = defineProps<{ kind: BillingAdminKind }>()
const { t, locale } = useI18n()
const auth = useAuthStore()
const rows = ref<Record<string, unknown>[]>([])
const total = ref(0)
const loading = ref(true)
const error = ref(false)

const config = computed(() => {
  const values: Record<BillingAdminKind, { title: string; description: string; endpoint: string; columns: string[] }> = {
    orders: { title: t('billing.admin.orders'), description: t('billing.admin.ordersDescription'), endpoint: '/api/v1/admin/orders', columns: ['public_order_no', 'user_id', 'status', 'total_amount_minor', 'currency', 'created_at'] },
    transactions: { title: t('billing.admin.transactions'), description: t('billing.admin.transactionsDescription'), endpoint: '/api/v1/admin/payment-transactions', columns: ['order_id', 'provider_code', 'type', 'direction', 'amount_minor', 'currency', 'status'] },
    events: { title: t('billing.admin.events'), description: t('billing.admin.eventsDescription'), endpoint: '/api/v1/admin/payment-webhook-events', columns: ['provider_code', 'event_id', 'event_type', 'signature_valid', 'processing_status', 'created_at'] },
    refunds: { title: t('billing.admin.refunds'), description: t('billing.admin.refundsDescription'), endpoint: '/api/v1/admin/refunds', columns: ['order_id', 'provider_code', 'amount_minor', 'currency', 'status', 'requested_at'] },
  }
  return values[props.kind]
})

function label(column: string) {
  return t(`billing.fields.${column}`, column)
}

function display(value: unknown, column: string) {
  if (value === null || value === undefined || value === '') return '—'
  if (column.endsWith('_at')) return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(String(value)))
  if (column.endsWith('_minor')) return `${(Number(value) / 100).toFixed(2)}`
  return String(value)
}

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = false
  try {
    const response = await apiFetchEnvelope<Record<string, unknown>[]>(`${config.value.endpoint}?page=1&per_page=50`, {}, auth.token)
    rows.value = response.data
    total.value = Number(response.meta?.total ?? response.data.length)
  } catch {
    rows.value = []
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-end justify-between gap-3">
      <div><div class="flex items-center gap-2"><WalletCards class="size-5 text-primary" /><h1 class="text-2xl font-semibold tracking-tight">{{ config.title }}</h1></div><p class="mt-1 text-sm text-muted-foreground">{{ config.description }}</p></div>
      <Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('billing.actions.refresh') }}</Button>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('billing.admin.loadFailed') }}</AlertTitle><AlertDescription>{{ t('billing.admin.loadFailedDescription') }}</AlertDescription></Alert>
    <Card>
      <CardHeader><CardTitle>{{ config.title }}</CardTitle><CardDescription>{{ t('billing.admin.total', { count: total }) }}</CardDescription></CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3"><Skeleton v-for="item in 5" :key="item" class="h-10" /></div>
        <Empty v-else-if="!rows.length"><EmptyHeader><EmptyTitle>{{ t('billing.admin.empty') }}</EmptyTitle><EmptyDescription>{{ t('billing.admin.emptyDescription') }}</EmptyDescription></EmptyHeader></Empty>
        <Table v-else><TableHeader><TableRow><TableHead v-for="column in config.columns" :key="column">{{ label(column) }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="(row, index) in rows" :key="String(row.id ?? index)"><TableCell v-for="column in config.columns" :key="column" class="max-w-64 truncate">{{ display(row[column], column) }}</TableCell></TableRow></TableBody></Table>
      </CardContent>
    </Card>
  </div>
</template>
