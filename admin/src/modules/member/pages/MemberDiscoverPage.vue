<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Flag, Image, LoaderCircle, RefreshCw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { apiFetch, apiFetchEnvelope, ApiError } from '@/lib/api'
import { setPageSEO } from '@/lib/seo'
import { useSitePresentation } from '@/lib/site-presentation'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

interface DiscoveryStatus { enabled: boolean; submissions_enabled: boolean }
interface DiscoveryItem {
  id: number
  original_name: string
  content_type: string
  width: number
  height: number
  size_bytes: number
  created_at: string
  thumbnail_url: string
  original_url: string
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const status = ref<DiscoveryStatus>({ enabled: true, submissions_enabled: true })
const items = ref<DiscoveryItem[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const failed = ref(false)
const page = ref(1)
const total = ref(0)
const reportingID = ref<number | null>(null)
const reportReason = ref('copyright')
const reportDescription = ref('')
const reportFailed = ref(false)
const reportedID = ref<number | null>(null)

const hasMore = computed(() => items.value.length < total.value)

onMounted(async () => {
  void loadSitePresentation()
  setPageSEO({ title: `${t('member.discover.title')} · FastImg`, description: t('member.discover.description'), path: '/discover' })
  await loadFeed(true)
})

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

async function loadFeed(reset = false) {
  if (reset) {
    loading.value = true
    failed.value = false
    page.value = 1
    items.value = []
  } else {
    loadingMore.value = true
  }
  try {
    const statusResponse = await apiFetch<DiscoveryStatus>('/api/v1/discovery/status')
    status.value = statusResponse
    if (!statusResponse.enabled) return
    const response = await apiFetchEnvelope<DiscoveryItem[]>(`/api/v1/discovery/feed?page=${page.value}&per_page=24`)
    items.value = reset ? response.data : [...items.value, ...response.data]
    total.value = Number(response.meta?.total ?? items.value.length)
  } catch {
    failed.value = true
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function loadMore() {
  page.value += 1
  await loadFeed()
}

function formatDate(value: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(value))
}

function formatSize(value: number) {
  if (!value) return ''
  const units = ['B', 'KB', 'MB', 'GB']
  let amount = value
  let unit = 0
  while (amount >= 1024 && unit < units.length - 1) { amount /= 1024; unit += 1 }
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(amount)} ${units[unit]}`
}

function openReport(item: DiscoveryItem) {
  if (!auth.isAuthenticated) {
    void router.push({ name: 'login', query: { redirect: '/discover' } })
    return
  }
  reportingID.value = item.id
  reportReason.value = 'copyright'
  reportDescription.value = ''
  reportFailed.value = false
}

async function submitReport() {
  if (!auth.token || !reportingID.value) return
  reportFailed.value = false
  try {
    await apiFetch(`/api/v1/media/${reportingID.value}/reports`, {
      method: 'POST',
      body: JSON.stringify({ reason: reportReason.value, description: reportDescription.value.trim() }),
    }, auth.token)
    reportedID.value = reportingID.value
    reportingID.value = null
  } catch (error) {
    reportFailed.value = error instanceof ApiError && error.code === 'REPORT_ALREADY_SUBMITTED'
  }
}
</script>

<template>
  <section class="space-y-8">
    <header class="max-w-2xl space-y-3">
      <p class="text-sm font-medium text-primary">{{ t('member.discover.eyebrow') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.discover.heading') }}</h1>
      <p class="text-muted-foreground">{{ t('member.discover.description') }}</p>
    </header>

    <Alert v-if="!status.enabled" class="max-w-2xl">
      <Image class="size-4" />
      <AlertTitle>{{ t('member.discover.disabledTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.discover.disabledDescription') }}</AlertDescription>
    </Alert>
    <Alert v-else-if="failed" variant="destructive" class="max-w-2xl">
      <AlertTitle>{{ t('member.discover.errorTitle') }}</AlertTitle>
      <AlertDescription class="flex items-center justify-between gap-4">
        <span>{{ t('member.discover.errorDescription') }}</span>
        <Button variant="outline" size="sm" @click="loadFeed(true)"><RefreshCw class="size-4" />{{ t('member.discover.retry') }}</Button>
      </AlertDescription>
    </Alert>

    <div v-else-if="loading" class="flex items-center gap-2 text-sm text-muted-foreground">
      <LoaderCircle class="size-4 animate-spin" />{{ t('member.discover.loading') }}
    </div>
    <div v-else-if="items.length" class="columns-1 gap-4 sm:columns-2 lg:columns-3 xl:columns-4">
      <article v-for="item in items" :key="item.id" class="group mb-4 break-inside-avoid overflow-hidden rounded-lg border bg-card shadow-sm">
        <a :href="item.original_url" target="_blank" rel="noreferrer" class="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset">
          <img :src="item.thumbnail_url" :alt="item.original_name" loading="lazy" class="block h-auto w-full bg-muted object-cover transition-opacity group-hover:opacity-90" @error="handleImageError" />
        </a>
        <div class="space-y-2 px-3 py-3">
          <div class="flex items-start justify-between gap-2">
            <p class="min-w-0 truncate text-sm font-medium" :title="item.original_name">{{ item.original_name }}</p>
            <Button variant="ghost" size="icon" class="size-8 shrink-0" :aria-label="t('member.discover.report')" :title="t('member.discover.report')" @click="openReport(item)"><Flag class="size-4" /></Button>
          </div>
          <p class="text-xs text-muted-foreground">{{ item.width }} × {{ item.height }} · {{ formatSize(item.size_bytes) }} · {{ formatDate(item.created_at) }}</p>
          <p v-if="reportedID === item.id" class="text-xs text-primary">{{ t('member.discover.reportSubmitted') }}</p>
        </div>
      </article>
    </div>
    <div v-else class="rounded-lg border border-dashed p-10 text-center text-muted-foreground">
      <Image class="mx-auto mb-3 size-8" />
      <p class="font-medium text-foreground">{{ t('member.discover.emptyTitle') }}</p>
      <p class="mt-1 text-sm">{{ t('member.discover.emptyDescription') }}</p>
    </div>

    <div v-if="hasMore && !loading" class="flex justify-center">
      <Button variant="outline" :disabled="loadingMore" @click="loadMore">
        <LoaderCircle v-if="loadingMore" class="size-4 animate-spin" />
        {{ t('member.discover.loadMore') }}
      </Button>
    </div>

    <div v-if="reportingID" class="max-w-xl space-y-4 rounded-lg border bg-card p-5">
      <div>
        <h2 class="font-semibold">{{ t('member.discover.reportTitle') }}</h2>
        <p class="mt-1 text-sm text-muted-foreground">{{ t('member.discover.reportDescription') }}</p>
      </div>
      <label class="block space-y-2 text-sm">
        <span class="font-medium">{{ t('member.discover.reportReason') }}</span>
        <select v-model="reportReason" class="flex h-10 w-full rounded-md border bg-background px-3 text-sm">
          <option value="copyright">{{ t('member.discover.reasons.copyright') }}</option>
          <option value="adult">{{ t('member.discover.reasons.adult') }}</option>
          <option value="illegal">{{ t('member.discover.reasons.illegal') }}</option>
          <option value="other">{{ t('member.discover.reasons.other') }}</option>
        </select>
      </label>
      <Textarea v-model="reportDescription" :placeholder="t('member.discover.reportPlaceholder')" maxlength="2000" />
      <Alert v-if="reportFailed" variant="destructive"><AlertDescription>{{ t('member.discover.reportFailed') }}</AlertDescription></Alert>
      <div class="flex justify-end gap-2">
        <Button variant="ghost" @click="reportingID = null">{{ t('member.discover.cancel') }}</Button>
        <Button @click="submitReport">{{ t('member.discover.submitReport') }}</Button>
      </div>
    </div>
  </section>
</template>
