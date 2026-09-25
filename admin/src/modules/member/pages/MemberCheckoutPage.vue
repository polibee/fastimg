<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AlertCircle, ArrowLeft, CreditCard } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { cancelMemberOrder, completeFakeMemberPayment, getMemberOrder, listMemberPaymentGateways, startMemberPayment, type MemberOrder } from '@/modules/billing/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const order = ref<MemberOrder>()
const gateways = ref<string[]>([])
const selectedGateway = ref('')
const payment = ref<{ status: string; checkout_url: string; provider_payment_id: string }>()
const loading = ref(true)
const paying = ref(false)
const canceling = ref(false)
const confirmingFake = ref(false)
type CheckoutError = 'order' | 'gateway' | 'payment' | 'cancel'
const error = ref<CheckoutError>()

function formatAmount(value = 0, currency = 'USD') {
  try { return new Intl.NumberFormat(locale.value, { style: 'currency', currency }).format(value / 100) } catch { return `${(value / 100).toFixed(2)} ${currency}` }
}

async function load() {
  if (!auth.token) { await router.replace({ name: 'login', query: { redirect: route.fullPath } }); return }
  loading.value = true
  error.value = undefined
  try {
    order.value = await getMemberOrder(String(route.params.orderId), auth.token)
  } catch { error.value = 'order'; loading.value = false; return }
  try {
    gateways.value = await listMemberPaymentGateways(auth.token)
    selectedGateway.value = order.value.gateway_code || gateways.value[0] || ''
  } catch {
    gateways.value = []
    selectedGateway.value = ''
    error.value = 'gateway'
  } finally { loading.value = false }
}

async function pay() {
  if (!auth.token || !order.value) return
  paying.value = true
  error.value = undefined
  try { payment.value = await startMemberPayment(String(order.value.id), selectedGateway.value, auth.token) } catch { error.value = 'payment' } finally { paying.value = false }
}

async function cancel() {
  if (!auth.token || !order.value) return
  canceling.value = true
  error.value = undefined
  try {
    await cancelMemberOrder(String(order.value.id), auth.token)
    order.value = { ...order.value, status: 'canceled' }
  } catch { error.value = 'cancel' } finally { canceling.value = false }
}

async function completeFake() {
  if (!auth.token || !order.value) return
  confirmingFake.value = true
  error.value = undefined
  try {
    await completeFakeMemberPayment(String(order.value.id), auth.token)
    payment.value = undefined
    await load()
  } catch { error.value = 'payment' } finally { confirmingFake.value = false }
}

function statusLabel(status: string) {
  return t(`member.billing.statuses.${status}`, status)
}

onMounted(load)
</script>

<template>
  <div class="mx-auto flex w-full max-w-3xl flex-col gap-6">
    <Button variant="ghost" class="w-fit" @click="router.push('/plans')"><ArrowLeft class="mr-2 size-4" />{{ t('member.billing.backToPlans') }}</Button>
    <Alert v-if="error" variant="destructive"><AlertCircle /><AlertTitle>{{ t(error === 'order' ? 'member.billing.orderErrorTitle' : error === 'gateway' ? 'member.billing.gatewayErrorTitle' : error === 'cancel' ? 'member.billing.cancelErrorTitle' : 'member.billing.paymentErrorTitle') }}</AlertTitle><AlertDescription>{{ t(error === 'order' ? 'member.billing.orderErrorDescription' : error === 'gateway' ? 'member.billing.gatewayErrorDescription' : error === 'cancel' ? 'member.billing.cancelErrorDescription' : 'member.billing.paymentErrorDescription') }}</AlertDescription></Alert>
    <Skeleton v-if="loading" class="h-64" />
    <Card v-else-if="order">
      <CardHeader><CardTitle class="flex items-center gap-2"><CreditCard class="size-5" />{{ t('member.billing.checkoutTitle') }}</CardTitle><CardDescription>{{ order.public_order_no }}</CardDescription></CardHeader>
      <CardContent class="flex flex-col gap-5">
        <div class="flex items-center justify-between border-b pb-4"><span class="text-muted-foreground">{{ t('member.billing.amount') }}</span><strong>{{ formatAmount(order.total_amount_minor, order.currency) }}</strong></div>
        <div class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('member.billing.status') }}</span><span>{{ statusLabel(order.status) }}</span></div>
        <label v-if="order.status === 'pending_payment'" class="grid gap-2 text-sm font-medium">
          {{ t('member.billing.gateway') }}
          <Select v-model="selectedGateway" :disabled="paying || gateways.length === 0">
            <SelectTrigger><SelectValue :placeholder="t('member.billing.gatewayPlaceholder')" /></SelectTrigger>
            <SelectContent><SelectItem v-for="gateway in gateways" :key="gateway" :value="gateway">{{ t(`member.billing.gateways.${gateway}`, gateway) }}</SelectItem></SelectContent>
          </Select>
        </label>
        <Button v-if="order.status === 'pending_payment'" :disabled="paying || !selectedGateway" @click="pay">{{ paying ? t('member.billing.startingPayment') : t('member.billing.payNow') }}</Button>
        <Button v-if="payment?.provider_payment_id && selectedGateway === 'fake' && order.status === 'pending_payment'" variant="secondary" :disabled="confirmingFake" @click="completeFake">{{ confirmingFake ? t('member.billing.confirmingFake') : t('member.billing.confirmFake') }}</Button>
        <Button v-if="order.status === 'pending_payment'" variant="ghost" :disabled="canceling" @click="cancel">{{ canceling ? t('member.billing.cancelingOrder') : t('member.billing.cancelOrder') }}</Button>
        <Alert v-if="payment"><AlertTitle>{{ t('member.billing.paymentCreated') }}</AlertTitle><AlertDescription>{{ t('member.billing.paymentPending') }} <a v-if="payment.checkout_url.startsWith('http')" class="underline" :href="payment.checkout_url" target="_blank" rel="noreferrer">{{ t('member.billing.openCheckout') }}</a></AlertDescription></Alert>
        <p class="text-xs text-muted-foreground">{{ t('member.billing.fulfillmentNotice') }}</p>
      </CardContent>
    </Card>
  </div>
</template>
