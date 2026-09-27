<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { HardDrive, RefreshCw, Settings2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ApiError, errorMessageKey } from '@/lib/api'
import { storageConnections, storageStatistics, type StorageConnection, type StorageOverview, type StorageStatistics } from '@/modules/settings/storage-api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const overview = ref<StorageOverview>()
const statistics = ref<StorageStatistics[]>([])
const loading = ref(true)
const error = ref('')
const enabledConnections = computed(() => (overview.value?.connections || []).filter((connection) => connection.enabled))
const primaryConnection = computed(() => enabledConnections.value.find((connection) => connection.is_primary))
const statisticsByProvider = computed(() => Object.fromEntries(statistics.value.map((item) => [item.provider_code, item])))

function providerTitle(connection: StorageConnection) { return t(`storage.providers.${connection.provider_code}.title`, connection.name) }
function statusLabel(status: string) { return t(`storage.${status}`, status) }
function statusClass(status: string) {
  if (status === 'healthy') return 'text-emerald-600'
  if (status === 'degraded') return 'text-amber-600'
  if (status === 'error' || status === 'incomplete') return 'text-destructive'
  return 'text-muted-foreground'
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let size = value
  let unit = -1
  do { size /= 1024; unit += 1 } while (size >= 1024 && unit < units.length - 1)
  return `${size.toFixed(size >= 10 ? 0 : 1)} ${units[unit]}`
}

function formatCount(value: number) { return value.toLocaleString() }
function labelForStatus(status: string) { return t(`storage.objectStatuses.${status}`, status) }
function labelForVariant(variant: string) { return t(`storage.variants.${variant}`, variant) }

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    const [connections, usage] = await Promise.all([storageConnections(auth.token), storageStatistics(auth.token)])
    overview.value = connections
    statistics.value = usage.connections
  } catch (cause) {
    error.value = cause instanceof ApiError && cause.code ? t(errorMessageKey(cause.code)) : t('storage.loadFailed')
  } finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-2"><HardDrive class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('storage.title') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('storage.description') }}</p></div></div>
      <div class="flex gap-2"><Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button><Button variant="outline" size="sm" as-child><RouterLink to="/admin/settings"><Settings2 class="mr-2 size-4" />{{ t('storage.openSettings') }}</RouterLink></Button></div>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <div class="grid gap-4 sm:grid-cols-3">
      <Card><CardHeader><CardTitle class="text-sm font-medium">{{ t('storage.enabledConnections') }}</CardTitle></CardHeader><CardContent><Skeleton v-if="loading" class="h-9 w-16" /><p v-else class="text-3xl font-semibold">{{ enabledConnections.length }}</p></CardContent></Card>
      <Card><CardHeader><CardTitle class="text-sm font-medium">{{ t('storage.primaryConnection') }}</CardTitle></CardHeader><CardContent><Skeleton v-if="loading" class="h-9 w-28" /><p v-else class="text-lg font-semibold">{{ primaryConnection ? providerTitle(primaryConnection) : t('storage.noPrimary') }}</p></CardContent></Card>
      <Card><CardHeader><CardTitle class="text-sm font-medium">{{ t('storage.providerAdapter') }}</CardTitle></CardHeader><CardContent><Skeleton v-if="loading" class="h-9 w-20" /><p v-else class="text-lg font-semibold">{{ enabledConnections.filter((item) => overview?.providers[item.provider_code]?.adapter_available).length }} / {{ enabledConnections.length }} {{ t('storage.available') }}</p></CardContent></Card>
    </div>
    <Card>
      <CardHeader><CardTitle>{{ t('storage.overview') }}</CardTitle><CardDescription>{{ t('storage.description') }}</CardDescription></CardHeader>
      <CardContent>
        <div v-if="loading" class="grid gap-3"><Skeleton v-for="item in 3" :key="item" class="h-16 w-full" /></div>
        <div v-else-if="!overview?.connections.length" class="text-sm text-muted-foreground">{{ t('storage.noPrimary') }}</div>
        <div v-else class="divide-y rounded-md border"><div v-for="connection in overview.connections" :key="connection.provider_code" class="flex flex-col gap-2 p-4 sm:flex-row sm:items-center sm:justify-between"><div><div class="flex items-center gap-2"><span class="font-medium">{{ providerTitle(connection) }}</span><span v-if="connection.is_primary" class="rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary">{{ t('storage.primary') }}</span></div><p class="mt-1 text-xs text-muted-foreground">{{ connection.public_base_url || connection.config.endpoint || t('storage.notConfigured') }}</p></div><div class="flex items-center gap-3 text-sm"><span :class="['font-medium', statusClass(connection.status)]">{{ statusLabel(connection.status) }}</span><span v-if="connection.last_error_message" class="max-w-md text-xs text-muted-foreground">{{ connection.last_error_message }}</span></div></div></div>
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>{{ t('storage.statisticsTitle') }}</CardTitle><CardDescription>{{ t('storage.statisticsDescription') }}</CardDescription></CardHeader>
      <CardContent>
        <div v-if="loading" class="grid gap-3 sm:grid-cols-2"><Skeleton v-for="item in 2" :key="item" class="h-28 w-full" /></div>
        <p v-else-if="!enabledConnections.length" class="text-sm text-muted-foreground">{{ t('storage.statisticsEmpty') }}</p>
        <div v-else class="grid gap-4 xl:grid-cols-2">
          <section v-for="connection in enabledConnections" :key="connection.provider_code" class="rounded-lg border p-4">
            <div class="flex items-start justify-between gap-3"><div><h3 class="font-medium">{{ providerTitle(connection) }}</h3><p class="mt-1 text-xs text-muted-foreground">{{ t('storage.statisticsSource') }}</p></div><span :class="['text-sm font-medium', statusClass(connection.status)]">{{ statusLabel(connection.status) }}</span></div>
            <div v-if="statisticsByProvider[connection.provider_code]" class="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
              <div><p class="text-muted-foreground">{{ t('storage.objectCount') }}</p><p class="mt-1 text-lg font-semibold">{{ formatCount(statisticsByProvider[connection.provider_code].object_count) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.storedBytes') }}</p><p class="mt-1 text-lg font-semibold">{{ formatBytes(statisticsByProvider[connection.provider_code].stored_bytes) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.allowedRequests') }}</p><p class="mt-1 text-lg font-semibold">{{ formatCount(statisticsByProvider[connection.provider_code].allowed_request_count) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.estimatedBandwidth') }}</p><p class="mt-1 text-lg font-semibold">{{ formatBytes(statisticsByProvider[connection.provider_code].estimated_bandwidth_bytes) }}</p></div>
            </div>
            <div v-if="statisticsByProvider[connection.provider_code]" class="mt-4 grid gap-3 text-sm sm:grid-cols-4">
              <div><p class="text-muted-foreground">{{ t('storage.mediaCount') }}</p><p class="mt-1 font-semibold">{{ formatCount(statisticsByProvider[connection.provider_code].media_count) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.orphanObjects') }}</p><p class="mt-1 font-semibold">{{ formatCount(statisticsByProvider[connection.provider_code].orphan_object_count) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.orphanBytes') }}</p><p class="mt-1 font-semibold">{{ formatBytes(statisticsByProvider[connection.provider_code].orphan_bytes) }}</p></div>
              <div><p class="text-muted-foreground">{{ t('storage.healthErrors24h') }}</p><p class="mt-1 font-semibold">{{ formatCount(statisticsByProvider[connection.provider_code].health.error_count_24h) }}</p></div>
            </div>
            <div v-if="statisticsByProvider[connection.provider_code]" class="mt-4 grid gap-4 lg:grid-cols-2">
              <div>
                <h4 class="text-sm font-medium">{{ t('storage.objectBreakdown') }}</h4>
                <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span v-for="(count, status) in statisticsByProvider[connection.provider_code].object_status_counts" :key="`status-${status}`">{{ labelForStatus(status) }}: {{ formatCount(count) }}</span>
                </div>
                <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span v-for="(count, variant) in statisticsByProvider[connection.provider_code].variant_counts" :key="`variant-${variant}`">{{ labelForVariant(variant) }}: {{ formatCount(count) }}</span>
                </div>
              </div>
              <div>
                <h4 class="text-sm font-medium">{{ t('storage.healthTitle') }}</h4>
                <p class="mt-2 text-xs text-muted-foreground">{{ t('storage.lastHealth') }}: {{ statisticsByProvider[connection.provider_code].health.last_status || t('storage.notAvailable') }}<span v-if="statisticsByProvider[connection.provider_code].health.last_latency_ms != null"> · {{ statisticsByProvider[connection.provider_code].health.last_latency_ms }} ms</span></p>
              </div>
            </div>
            <div v-if="statisticsByProvider[connection.provider_code]" class="mt-5 overflow-x-auto">
              <h4 class="text-sm font-medium">{{ t('storage.trendTitle') }}</h4>
              <table class="mt-2 w-full text-left text-xs">
                <thead class="border-b text-muted-foreground"><tr><th class="py-2 pr-4 font-medium">{{ t('storage.date') }}</th><th class="py-2 pr-4 font-medium">{{ t('storage.uploads') }}</th><th class="py-2 pr-4 font-medium">{{ t('storage.allowedRequests') }}</th><th class="py-2 font-medium">{{ t('storage.estimatedBandwidth') }}</th></tr></thead>
                <tbody><tr v-for="day in statisticsByProvider[connection.provider_code].daily.slice(-7)" :key="day.date" class="border-b last:border-0"><td class="py-2 pr-4">{{ day.date }}</td><td class="py-2 pr-4">{{ formatCount(day.upload_count) }} · {{ formatBytes(day.upload_bytes) }}</td><td class="py-2 pr-4">{{ formatCount(day.allowed_request_count) }}</td><td class="py-2">{{ formatBytes(day.estimated_bandwidth_bytes) }}</td></tr></tbody>
              </table>
            </div>
          </section>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
