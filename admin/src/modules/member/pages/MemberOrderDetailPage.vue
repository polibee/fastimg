<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { getMemberOrder, type MemberOrder } from '@/modules/billing/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n(); const route = useRoute(); const router = useRouter(); const auth = useAuthStore(); const order = ref<MemberOrder>(); const loading = ref(true)
function statusLabel(status: string) { return t(`member.billing.statuses.${status}`, status) }
onMounted(async () => { if (!auth.token) { await router.replace({ name: 'login', query: { redirect: route.fullPath } }); return }; try { order.value = await getMemberOrder(String(route.params.id), auth.token) } finally { loading.value = false } })
</script>
<template><div class="mx-auto flex w-full max-w-3xl flex-col gap-6"><Button variant="ghost" class="w-fit" @click="router.push('/orders')"><ArrowLeft class="mr-2 size-4" />{{ t('member.billing.backToOrders') }}</Button><Skeleton v-if="loading" class="h-64" /><Card v-else-if="order"><CardHeader><CardTitle>{{ order.public_order_no }}</CardTitle></CardHeader><CardContent class="flex flex-col gap-3"><div class="flex justify-between"><span>{{ t('member.billing.status') }}</span><strong>{{ statusLabel(order.status) }}</strong></div><div class="flex justify-between"><span>{{ t('member.billing.amount') }}</span><strong>{{ (order.total_amount_minor / 100).toFixed(2) }} {{ order.currency }}</strong></div><Button v-if="order.status === 'pending_payment'" @click="router.push({ name: 'member-checkout', params: { orderId: order.id } })">{{ t('member.billing.continueCheckout') }}</Button></CardContent></Card></div></template>
