<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ClipboardList } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { listMemberOrders, type MemberOrder } from '@/modules/billing/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const orders = ref<MemberOrder[]>([])
const loading = ref(true)

function formatDate(value: string) { return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) }
function statusLabel(status: string) { return t(`member.billing.statuses.${status}`, status) }
onMounted(async () => {
  if (!auth.token) { await router.replace({ name: 'login', query: { redirect: '/orders' } }); return }
  try { orders.value = (await listMemberOrders(auth.token)).data } finally { loading.value = false }
})
</script>

<template>
  <div class="flex flex-col gap-6"><div><div class="flex items-center gap-2"><ClipboardList class="size-5 text-primary" /><h1 class="text-2xl font-semibold tracking-tight">{{ t('member.billing.ordersTitle') }}</h1></div><p class="mt-1 text-sm text-muted-foreground">{{ t('member.billing.ordersDescription') }}</p></div>
    <Skeleton v-if="loading" class="h-64" />
    <Empty v-else-if="!orders.length"><EmptyHeader><EmptyTitle>{{ t('member.billing.noOrders') }}</EmptyTitle><EmptyDescription>{{ t('member.billing.noOrdersDescription') }}</EmptyDescription></EmptyHeader><Button as-child><RouterLink to="/plans">{{ t('member.billing.viewPlans') }}</RouterLink></Button></Empty>
    <Card v-else v-for="order in orders" :key="order.id"><CardHeader class="flex flex-row items-center justify-between"><div><CardTitle>{{ order.public_order_no }}</CardTitle><p class="mt-1 text-sm text-muted-foreground">{{ formatDate(order.created_at) }}</p></div><Button as-child variant="outline"><RouterLink :to="{ name: 'member-order-detail', params: { id: order.id } }">{{ t('member.billing.viewOrder') }}</RouterLink></Button></CardHeader><CardContent class="flex justify-between text-sm"><span>{{ statusLabel(order.status) }}</span><strong>{{ (order.total_amount_minor / 100).toFixed(2) }} {{ order.currency }}</strong></CardContent></Card>
  </div>
</template>
