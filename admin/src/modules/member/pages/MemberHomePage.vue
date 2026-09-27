<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight, Image, ShieldCheck } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { apiFetch, apiFetchBlob, apiFetchEnvelope } from '@/lib/api'
import { useSitePresentation } from '@/lib/site-presentation'
import { setPageSEO } from '@/lib/seo'
import { useAuthStore } from '@/stores/auth'
import MemberUploadPanel from '../components/MemberUploadPanel.vue'

interface MediaVariant { url: string }
interface MediaItem {
  id: number
  original_name: string
  content_type: string
  width: number
  height: number
  variants: Record<string, MediaVariant>
}
interface PlanData { name: string }
interface UsageData { usage: Record<string, number>; limits: { storage_bytes: number } }

const { t, locale } = useI18n()
const auth = useAuthStore()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const recent = ref<MediaItem[]>([])
const previews = ref<Record<number, string>>({})
const plan = ref<PlanData>()
const usage = ref<UsageData>()
const loading = ref(true)
const error = ref(false)
const recentError = ref(false)
const storagePercent = computed(() => {
  const limit = usage.value?.limits.storage_bytes ?? 0
  return limit > 0 ? Math.min(100, (usage.value?.usage.storage ?? 0) / limit * 100) : 0
})

function formatBytes(bytes: number) {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) { value /= 1024; unit += 1 }
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)} ${units[unit]}`
}

function releasePreviews() {
  Object.values(previews.value).forEach((url) => URL.revokeObjectURL(url))
  previews.value = {}
}

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

async function loadRecent() {
  if (!auth.token) return
  releasePreviews()
  recentError.value = false
  try {
    const response = await apiFetchEnvelope<MediaItem[]>('/api/v1/media?page=1&per_page=6', {}, auth.token)
    recent.value = response.data
    await Promise.allSettled(response.data.map(async (item) => {
      const original = item.variants?.original?.url
      if (!original || !auth.token) return
      const blob = await apiFetchBlob(original, auth.token)
      previews.value[item.id] = URL.createObjectURL(blob)
    }))
  } catch {
    recentError.value = true
  }
}

async function loadOverview() {
  if (!auth.token) { loading.value = false; return }
  try {
    const [subscription, ownUsage] = await Promise.all([
      apiFetch<{ plan: PlanData }>('/api/v1/subscription', {}, auth.token),
      apiFetch<UsageData>('/api/v1/me/usage', {}, auth.token),
    ])
    plan.value = subscription.plan
    usage.value = ownUsage
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
  await loadRecent()
}

onMounted(() => {
  void loadSitePresentation()
  setPageSEO({ title: `${t('member.home.heading')} · FastImg`, description: t('member.home.description'), path: '/' })
  void loadOverview()
})
onBeforeUnmount(releasePreviews)
</script>

<template>
  <div class="flex flex-col gap-9">
    <section class="max-w-3xl">
      <p class="mb-2 text-sm text-muted-foreground">{{ t('member.home.eyebrow') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.home.heading') }}</h1>
      <p class="mt-3 text-muted-foreground">{{ t('member.home.description') }}</p>
    </section>

    <MemberUploadPanel @updated="loadRecent" />

    <Alert v-if="error" variant="destructive">
      <AlertTitle>{{ t('member.home.errors.overview') }}</AlertTitle>
      <AlertDescription>{{ t('member.errors.loadDescription') }}</AlertDescription>
    </Alert>

    <section class="grid gap-4 sm:grid-cols-2" :aria-label="t('member.home.accountSummary')">
      <Card>
        <CardHeader class="pb-3">
          <CardDescription class="flex items-center gap-2"><ShieldCheck />{{ t('member.home.currentPlan') }}</CardDescription>
          <CardTitle v-if="!loading && plan" class="text-xl">{{ plan.name }}</CardTitle>
          <Skeleton v-else-if="loading" class="h-7 w-36" />
          <CardTitle v-else-if="auth.isAuthenticated" class="text-base text-muted-foreground">{{ t('member.home.planUnavailable') }}</CardTitle>
          <CardTitle v-else class="text-base text-muted-foreground">{{ t('member.home.loginToViewUsage') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <RouterLink v-if="auth.isAuthenticated" to="/plans" class="text-sm text-primary hover:underline">{{ t('member.home.viewPlans') }} <ArrowRight class="inline size-4" /></RouterLink>
          <RouterLink v-else :to="{ name: 'login', query: { redirect: '/plans' } }" class="text-sm text-primary hover:underline">{{ t('member.home.loginToViewUsage') }} <ArrowRight class="inline size-4" /></RouterLink>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="pb-3">
          <CardDescription class="flex items-center gap-2"><Image />{{ t('member.home.storage') }}</CardDescription>
          <CardTitle v-if="!loading && usage" class="text-xl">{{ formatBytes(usage.usage.storage ?? 0) }} <span class="text-base font-normal text-muted-foreground">/ {{ usage.limits.storage_bytes ? formatBytes(usage.limits.storage_bytes) : t('member.plans.unlimited') }}</span></CardTitle>
          <Skeleton v-else-if="loading" class="h-7 w-44" />
          <CardTitle v-else-if="auth.isAuthenticated" class="text-base text-muted-foreground">{{ t('member.home.usageUnavailable') }}</CardTitle>
          <CardTitle v-else class="text-base text-muted-foreground">{{ t('member.home.loginToViewUsage') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="auth.isAuthenticated" class="h-2 overflow-hidden rounded-full bg-muted"><div class="h-full rounded-full bg-primary transition-all" :style="{ width: `${storagePercent}%` }" /></div>
          <RouterLink v-else :to="{ name: 'login', query: { redirect: '/plans' } }" class="text-sm text-primary hover:underline">{{ t('member.home.loginToViewUsage') }}</RouterLink>
        </CardContent>
      </Card>
    </section>

    <section class="flex flex-col gap-4">
      <div class="flex items-end justify-between gap-4">
        <div><h2 class="text-xl font-semibold">{{ t('member.home.recentTitle') }}</h2><p class="mt-1 text-sm text-muted-foreground">{{ auth.isAuthenticated ? t('member.home.recentDescription') : t('member.home.guestRecentDescription') }}</p></div>
        <Button v-if="auth.isAuthenticated" variant="outline" as-child><RouterLink to="/media">{{ t('member.home.viewAll') }}<ArrowRight data-icon="inline-end" /></RouterLink></Button>
      </div>
      <Alert v-if="recentError" variant="destructive"><AlertDescription>{{ t('member.home.errors.recentMedia') }}</AlertDescription></Alert>
      <div v-else-if="recent.length" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        <RouterLink v-for="item in recent" :key="item.id" to="/media" class="group overflow-hidden rounded-lg border bg-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <div class="aspect-square bg-muted/40"><img :src="previews[item.id] || fallbackImageURL" :alt="item.original_name" class="size-full object-cover transition-transform group-hover:scale-105" @error="handleImageError" /></div>
          <div class="flex items-center justify-between gap-2 p-2"><span class="truncate text-xs" :title="item.original_name">{{ item.original_name }}</span><Badge variant="outline" class="shrink-0 text-[10px]">{{ item.content_type.replace('image/', '').toUpperCase() }}</Badge></div>
        </RouterLink>
      </div>
      <div v-else-if="!loading && auth.isAuthenticated" class="rounded-lg border border-dashed px-5 py-10 text-center text-sm text-muted-foreground">{{ t('member.home.emptyRecent') }}</div>
      <div v-else-if="!loading" class="rounded-lg border border-dashed px-5 py-10 text-center text-sm text-muted-foreground">
        <p>{{ t('member.home.guestRecent') }}</p>
        <RouterLink :to="{ name: 'login', query: { redirect: '/media' } }" class="mt-2 inline-block text-primary hover:underline">{{ t('member.actions.login') }}</RouterLink>
      </div>
      <div v-else class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6"><Skeleton v-for="i in 6" :key="i" class="aspect-square rounded-lg" /></div>
    </section>
  </div>
</template>
