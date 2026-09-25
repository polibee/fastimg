<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle, Check, Image, LoaderCircle, ShieldCheck } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { apiFetch } from '@/lib/api'
import { setPageSEO } from '@/lib/seo'
import { useAuthStore } from '@/stores/auth'
import { createMemberOrder, type MemberPlanPrice } from '@/modules/billing/api'
import { useRouter } from 'vue-router'

interface PlanEntitlements {
  storage_bytes: number
  max_file_bytes: number
  daily_uploads: number
  monthly_api_uploads: number
  monthly_bandwidth_bytes: number
  transform_count: number
  api_rate_per_minute: number
  token_limit: number
  ads_enabled: boolean
  watermark_enabled: boolean
}

interface Plan {
  id: number
  code: string
  name: string
  description: string
  price_amount: number
  currency: string
  billing_period: string
  entitlements: PlanEntitlements
  prices?: MemberPlanPrice[]
}

interface SubscriptionResponse {
  subscription: { id: number; plan_id: number; status: string; starts_at: string | null; ends_at: string | null }
  plan: Plan
  snapshot_available: boolean
}

interface UsageResponse {
  period_key: string
  usage: Record<string, number>
  limits: PlanEntitlements
  bandwidth_metered: boolean
  administrator?: boolean
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const plans = ref<Plan[]>([])
const subscription = ref<SubscriptionResponse>()
const usage = ref<UsageResponse>()
const loading = ref(true)
const error = ref(false)
const bandwidthMeteringEnabled = computed(() => usage.value?.bandwidth_metered ?? false)
const checkoutLoading = ref<number | null>(null)

const currentPlanID = computed(() => subscription.value?.plan?.id ?? subscription.value?.subscription?.plan_id)
const currentPlan = computed(() => subscription.value?.plan)
const storagePercent = computed(() => quotaPercent(usage.value?.usage.storage ?? 0, usage.value?.limits.storage_bytes ?? 0))
const bandwidthPercent = computed(() => quotaPercent(usage.value?.usage.bandwidth ?? 0, usage.value?.limits.monthly_bandwidth_bytes ?? 0))

onMounted(async () => {
  setPageSEO({ title: `${t('member.plans.title')} · FastImg`, description: t('member.plans.guestDescription'), path: '/plans' })
  try {
    plans.value = await apiFetch<Plan[]>('/api/v1/plans')
    if (auth.token) {
      const [subscriptionData, usageData] = await Promise.all([
        apiFetch<SubscriptionResponse>('/api/v1/subscription', {}, auth.token),
        apiFetch<UsageResponse>('/api/v1/me/usage', {}, auth.token),
      ])
      subscription.value = subscriptionData
      usage.value = usageData
    }
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
})

function quotaPercent(used: number, limit: number) {
  if (!limit) return 0
  return Math.min(100, Math.max(0, (used / limit) * 100))
}

function formatBytes(bytes: number, zeroMeansUnlimited = false) {
  if (!bytes && zeroMeansUnlimited) return t('member.plans.unlimited')
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit += 1
  }
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)} ${units[unit]}`
}

function formatCount(value: number) {
  return value ? new Intl.NumberFormat(locale.value).format(value) : t('member.plans.unlimited')
}

function formatPrice(plan: Plan) {
  if (plan.price_amount === 0) return t('member.plans.free')
  try {
    return new Intl.NumberFormat(locale.value, { style: 'currency', currency: plan.currency }).format(plan.price_amount / 100)
  } catch {
    return `${(plan.price_amount / 100).toFixed(2)} ${plan.currency}`
  }
}

function entitlementRows(plan: Plan) {
  const entitlements = plan.entitlements
  return [
    { label: t('member.plans.storage'), value: formatBytes(entitlements.storage_bytes, true) },
    { label: t('member.plans.maxFile'), value: formatBytes(entitlements.max_file_bytes, true) },
    { label: t('member.plans.dailyUploads'), value: formatCount(entitlements.daily_uploads) },
    { label: t('member.plans.apiUploads'), value: formatCount(entitlements.monthly_api_uploads) },
    { label: t('member.plans.bandwidth'), value: formatBytes(entitlements.monthly_bandwidth_bytes, true) },
    { label: t('member.plans.transforms'), value: formatCount(entitlements.transform_count) },
    { label: t('member.plans.apiRate'), value: formatCount(entitlements.api_rate_per_minute) },
    { label: t('member.plans.tokens'), value: formatCount(entitlements.token_limit) },
    { label: t('member.plans.ads'), value: entitlements.ads_enabled ? t('member.plans.adsEnabled') : t('member.plans.adsDisabled') },
    { label: t('member.plans.watermark'), value: entitlements.watermark_enabled ? t('member.plans.watermarkEnabled') : t('member.plans.watermarkDisabled') },
  ]
}

async function startCheckout(plan: Plan) {
  if (!auth.token || plan.price_amount === 0) return
  const price = plan.prices?.[0]
  if (!price) return
  checkoutLoading.value = plan.id
  try {
    const order = await createMemberOrder({ plan_id: plan.id, price_id: price.id, currency: price.currency, billing_period: price.billing_period }, auth.token)
    await router.push({ name: 'member-checkout', params: { orderId: order.id } })
  } finally {
    checkoutLoading.value = null
  }
}
</script>

<template>
  <div class="flex flex-col gap-8">
    <section class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div class="max-w-2xl">
        <p class="mb-2 text-sm text-muted-foreground">{{ t('member.plans.eyebrow') }}</p>
        <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.plans.heading') }}</h1>
        <p class="mt-3 text-muted-foreground">{{ auth.isAuthenticated ? t('member.plans.description') : t('member.plans.guestDescription') }}</p>
      </div>
      <Badge v-if="currentPlan" variant="secondary" class="w-fit gap-1.5 px-3 py-1.5">
        <ShieldCheck />{{ t('member.plans.currentPlan', { name: currentPlan.name }) }}
      </Badge>
    </section>

    <Alert v-if="error" variant="destructive">
      <AlertCircle />
      <AlertTitle>{{ t('member.errors.loadTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.errors.loadDescription') }}</AlertDescription>
    </Alert>

    <template v-if="loading">
      <div class="grid gap-4 md:grid-cols-2">
        <Skeleton class="h-44" />
        <Skeleton class="h-44" />
      </div>
      <div class="grid gap-4 lg:grid-cols-3">
        <Skeleton v-for="index in 3" :key="index" class="h-96" />
      </div>
    </template>

    <template v-else-if="!error">
      <section v-if="usage" class="grid gap-4 md:grid-cols-2" :aria-label="t('member.plans.usageTitle')">
        <Alert v-if="usage.administrator" class="md:col-span-2">
          <ShieldCheck />
          <AlertTitle>{{ t('member.plans.administratorTitle') }}</AlertTitle>
          <AlertDescription>{{ t('member.plans.administratorDescription') }}</AlertDescription>
        </Alert>
        <Card>
          <CardHeader class="pb-3">
            <CardDescription class="flex items-center gap-2"><Image />{{ t('member.plans.storageUsed') }}</CardDescription>
            <CardTitle class="text-2xl">{{ formatBytes(usage.usage.storage ?? 0) }} <span class="text-base font-normal text-muted-foreground">/ {{ formatBytes(usage.limits.storage_bytes, true) }}</span></CardTitle>
          </CardHeader>
          <CardContent class="flex flex-col gap-2">
            <Progress :model-value="storagePercent" :aria-label="t('member.plans.storageUsed')" />
            <p class="text-sm text-muted-foreground">{{ t('member.plans.lifetimeBalance') }}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-3">
            <CardDescription class="flex items-center gap-2"><LoaderCircle />{{ t('member.plans.bandwidthUsed') }}</CardDescription>
            <CardTitle class="text-2xl">
              <template v-if="bandwidthMeteringEnabled">{{ formatBytes(usage.usage.bandwidth ?? 0) }} <span class="text-base font-normal text-muted-foreground">/ {{ formatBytes(usage.limits.monthly_bandwidth_bytes, true) }}</span></template>
              <span v-else>{{ t('member.plans.bandwidthNotMetered') }}</span>
            </CardTitle>
          </CardHeader>
          <CardContent class="flex flex-col gap-2">
            <Progress v-if="bandwidthMeteringEnabled" :model-value="bandwidthPercent" :aria-label="t('member.plans.bandwidthUsed')" />
            <p class="text-sm text-muted-foreground">{{ bandwidthMeteringEnabled ? t('member.plans.monthPeriod', { period: usage.period_key }) : t('member.plans.bandwidthNotMeteredHelp') }}</p>
          </CardContent>
        </Card>
      </section>
      <Card v-else>
        <CardHeader>
          <CardTitle>{{ t('member.plans.loginToViewUsage') }}</CardTitle>
          <CardDescription>{{ t('member.plans.loginToViewUsageHelp') }}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button as-child><RouterLink :to="{ name: 'login', query: { redirect: '/plans' } }">{{ t('member.actions.login') }}</RouterLink></Button>
        </CardContent>
      </Card>

      <section class="grid gap-4 lg:grid-cols-3" :aria-label="t('member.plans.availablePlans')">
        <Card v-for="plan in plans" :key="plan.id" class="flex flex-col" :class="plan.id === currentPlanID ? 'border-primary shadow-sm' : ''">
          <CardHeader>
            <div class="flex items-start justify-between gap-3">
              <div>
                <CardTitle class="text-xl">{{ plan.name }}</CardTitle>
                <CardDescription class="mt-2 min-h-10">{{ plan.description }}</CardDescription>
              </div>
              <Badge v-if="plan.id === currentPlanID" variant="default">{{ t('member.plans.current') }}</Badge>
            </div>
            <p class="pt-3 text-3xl font-semibold tracking-tight">{{ formatPrice(plan) }}<span v-if="plan.price_amount > 0" class="ml-1 text-sm font-normal text-muted-foreground">/ {{ t(`member.plans.periods.${plan.billing_period}`, plan.billing_period) }}</span></p>
          </CardHeader>
          <CardContent class="flex flex-1 flex-col gap-4">
            <ul class="flex flex-col gap-3 border-t pt-4">
              <li v-for="item in entitlementRows(plan)" :key="item.label" class="flex items-start justify-between gap-3 text-sm">
                <span class="text-muted-foreground">{{ item.label }}</span>
                <span class="flex items-center gap-1 text-right font-medium"><Check />{{ item.value }}</span>
              </li>
            </ul>
            <Button class="mt-auto w-full" :variant="plan.id === currentPlanID ? 'secondary' : 'outline'" :disabled="plan.id === currentPlanID || plan.price_amount === 0 || !plan.prices?.length || checkoutLoading === plan.id" @click="startCheckout(plan)">
              {{ checkoutLoading === plan.id ? t('member.billing.creatingOrder') : plan.id === currentPlanID ? t('member.plans.current') : plan.price_amount === 0 ? t('member.plans.free') : plan.prices?.length ? t('member.billing.choosePlan') : t('member.plans.purchaseUnavailable') }}
            </Button>
          </CardContent>
        </Card>
      </section>
      <p class="text-sm text-muted-foreground">{{ t('member.plans.purchaseNotice') }}</p>
    </template>
  </div>
</template>
