<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ClipboardList } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { listMemberOrders, type MemberOrder } from '@/modules/billing/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const orders = ref<MemberOrder[]>([])
const loading = ref(true)
const meta = ref({ page: 1, per_page: 20, total: 0 })
const pageSize = ref('20')

function formatDate(value: string) { return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) }
function statusLabel(status: string) { return t(`member.billing.statuses.${status}`, status) }
async function load(page = 1) {
  if (!auth.token) { await router.replace({ name: 'login', query: { redirect: '/orders' } }); return }
  loading.value = true
  try {
    const response = await listMemberOrders(auth.token, page, Number(pageSize.value))
    orders.value = response.data
    meta.value = { page: Number(response.meta?.page || page), per_page: Number(response.meta?.per_page || pageSize.value), total: Number(response.meta?.total || 0) }
  } finally { loading.value = false }
}
function changePageSize(value: unknown) { pageSize.value = String(value); void load(1) }
onMounted(() => void load())
</script>

<template>
  <div class="flex flex-col gap-6"><div><div class="flex items-center gap-2"><ClipboardList class="size-5 text-primary" /><h1 class="text-2xl font-semibold tracking-tight">{{ t('member.billing.ordersTitle') }}</h1></div><p class="mt-1 text-sm text-muted-foreground">{{ t('member.billing.ordersDescription') }}</p></div>
    <Skeleton v-if="loading" class="h-64" />
    <Empty v-else-if="!orders.length"><EmptyHeader><EmptyTitle>{{ t('member.billing.noOrders') }}</EmptyTitle><EmptyDescription>{{ t('member.billing.noOrdersDescription') }}</EmptyDescription></EmptyHeader><Button as-child><RouterLink to="/plans">{{ t('member.billing.viewPlans') }}</RouterLink></Button></Empty>
    <Card v-else v-for="order in orders" :key="order.id"><CardHeader class="flex flex-row items-center justify-between"><div><CardTitle>{{ order.public_order_no }}</CardTitle><p class="mt-1 text-sm text-muted-foreground">{{ formatDate(order.created_at) }}</p></div><Button as-child variant="outline"><RouterLink :to="{ name: 'member-order-detail', params: { id: order.id } }">{{ t('member.billing.viewOrder') }}</RouterLink></Button></CardHeader><CardContent class="flex justify-between text-sm"><span>{{ statusLabel(order.status) }}</span><strong>{{ (order.total_amount_minor / 100).toFixed(2) }} {{ order.currency }}</strong></CardContent></Card>
    <div v-if="!loading && meta.total > 0" class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
  </div>
</template>
