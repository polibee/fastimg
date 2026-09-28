<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Archive, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiFetch, apiFetchEnvelope, ApiError, errorMessageKey } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

type PlanPrice = {
  id: number
  version: string
  currency: string
  amount_minor: number
  billing_period: 'monthly' | 'yearly'
  trial_days: number
  status: 'active' | 'archived' | 'draft'
  effective_from: string
  effective_to?: string | null
}

const props = defineProps<{ planId: string }>()
const { t } = useI18n()
const auth = useAuthStore()
const prices = ref<PlanPrice[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref(false)
const draft = reactive({ amount: '', currency: 'CNY', billingPeriod: 'monthly', trialDays: '0' })

function localizedError(value: unknown) {
  return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('plans.pricing.unknownError')
}

function formatMoney(minor: number, currency: string) {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency, minimumFractionDigits: 2 }).format(minor / 100)
}

function formatDate(value?: string | null) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    prices.value = (await apiFetchEnvelope<PlanPrice[]>(`/api/v1/admin/plans/${encodeURIComponent(props.planId)}/prices`, {}, auth.token)).data
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    loading.value = false
  }
}

async function addPrice() {
  if (!auth.token) return
  const amount = Number.parseFloat(draft.amount)
  const trialDays = Number.parseInt(draft.trialDays || '0', 10)
  if (!Number.isFinite(amount) || amount < 0 || !Number.isInteger(trialDays) || trialDays < 0) {
    error.value = t('plans.pricing.invalidAmount')
    return
  }
  saving.value = true
  error.value = ''
  success.value = false
  try {
    await apiFetch<PlanPrice>(`/api/v1/admin/plans/${encodeURIComponent(props.planId)}/prices`, {
      method: 'POST',
      body: JSON.stringify({
        amount_minor: Math.round(amount * 100),
        currency: draft.currency,
        billing_period: draft.billingPeriod,
        trial_days: trialDays,
      }),
    }, auth.token)
    draft.amount = ''
    draft.trialDays = '0'
    success.value = true
    await load()
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    saving.value = false
  }
}

async function archivePrice(price: PlanPrice) {
  if (!auth.token || price.status !== 'active') return
  saving.value = true
  error.value = ''
  try {
    await apiFetch<PlanPrice>(`/api/v1/admin/plans/${encodeURIComponent(props.planId)}/prices/${price.id}/archive`, { method: 'POST', body: JSON.stringify({}) }, auth.token)
    await load()
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    saving.value = false
  }
}

async function deletePrice(price: PlanPrice) {
  if (!auth.token || price.status === 'active') return
  if (!window.confirm(t('plans.pricing.deleteConfirm'))) return
  saving.value = true
  error.value = ''
  try {
    await apiFetch<PlanPrice>(`/api/v1/admin/plans/${encodeURIComponent(props.planId)}/prices/${price.id}`, { method: 'DELETE' }, auth.token)
    await load()
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex items-start justify-between gap-3">
        <div>
          <CardTitle>{{ t('plans.pricing.title') }}</CardTitle>
          <CardDescription>{{ t('plans.pricing.description') }}</CardDescription>
        </div>
        <Button type="button" variant="ghost" size="icon" :aria-label="t('plans.pricing.refresh')" :disabled="loading || saving" @click="load"><RefreshCw :class="loading ? 'animate-spin' : ''" /></Button>
      </div>
    </CardHeader>
    <CardContent class="grid gap-5">
      <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('plans.pricing.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
      <Alert v-if="success"><AlertTitle>{{ t('plans.pricing.savedTitle') }}</AlertTitle><AlertDescription>{{ t('plans.pricing.savedDescription') }}</AlertDescription></Alert>

      <div class="grid gap-3 rounded-lg border bg-muted/20 p-4 sm:grid-cols-4">
        <div class="grid gap-2 sm:col-span-1"><Label for="plan-price-amount">{{ t('plans.pricing.amount') }}</Label><Input id="plan-price-amount" v-model="draft.amount" type="number" min="0" step="0.01" :placeholder="t('plans.pricing.amountPlaceholder')" /></div>
        <div class="grid gap-2"><Label>{{ t('plans.pricing.currency') }}</Label><Select v-model="draft.currency"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="CNY">CNY</SelectItem><SelectItem value="USD">USD</SelectItem><SelectItem value="EUR">EUR</SelectItem></SelectContent></Select></div>
        <div class="grid gap-2"><Label>{{ t('plans.pricing.period') }}</Label><Select v-model="draft.billingPeriod"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="monthly">{{ t('plans.pricing.monthly') }}</SelectItem><SelectItem value="yearly">{{ t('plans.pricing.yearly') }}</SelectItem></SelectContent></Select></div>
        <div class="grid gap-2"><Label for="plan-price-trial">{{ t('plans.pricing.trialDays') }}</Label><Input id="plan-price-trial" v-model="draft.trialDays" type="number" min="0" max="365" step="1" /></div>
        <div class="sm:col-span-4"><Button type="button" :disabled="saving || !draft.amount" @click="addPrice"><Plus data-icon="inline-start" />{{ t('plans.pricing.add') }}</Button></div>
      </div>

      <p class="text-xs leading-5 text-muted-foreground">{{ t('plans.pricing.immutableNotice') }}</p>
      <div v-if="!loading && !prices.length" class="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">{{ t('plans.pricing.empty') }}</div>
      <div v-else-if="!loading" class="overflow-x-auto rounded-lg border">
        <table class="w-full text-left text-sm">
          <thead class="border-b bg-muted/30"><tr><th class="px-3 py-2 font-medium">{{ t('plans.pricing.price') }}</th><th class="px-3 py-2 font-medium">{{ t('plans.pricing.period') }}</th><th class="px-3 py-2 font-medium">{{ t('plans.pricing.status') }}</th><th class="px-3 py-2 font-medium">{{ t('plans.pricing.effectiveFrom') }}</th><th class="px-3 py-2 font-medium text-right">{{ t('plans.pricing.actions') }}</th></tr></thead>
          <tbody><tr v-for="price in prices" :key="price.id" class="border-b last:border-0"><td class="px-3 py-2 font-medium">{{ formatMoney(price.amount_minor, price.currency) }}</td><td class="px-3 py-2">{{ price.billing_period === 'monthly' ? t('plans.pricing.monthly') : t('plans.pricing.yearly') }}<span v-if="price.trial_days" class="ml-1 text-xs text-muted-foreground">· {{ t('plans.pricing.trialDaysShort', { days: price.trial_days }) }}</span></td><td class="px-3 py-2">{{ price.status === 'active' ? t('plans.pricing.active') : t('plans.pricing.archived') }}</td><td class="px-3 py-2 text-muted-foreground">{{ formatDate(price.effective_from) }}</td><td class="px-3 py-2 text-right"><Button v-if="price.status === 'active'" type="button" variant="ghost" size="sm" :disabled="saving" @click="archivePrice(price)"><Archive data-icon="inline-start" />{{ t('plans.pricing.archive') }}</Button><Button v-else type="button" variant="ghost" size="sm" :disabled="saving" @click="deletePrice(price)"><Trash2 data-icon="inline-start" />{{ t('plans.pricing.delete') }}</Button></td></tr></tbody>
        </table>
      </div>
    </CardContent>
  </Card>
</template>
