<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { BarChart3, RefreshCw } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { generatedApi, type AdminOverview, type AdminTrendReport } from '@/generated/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const loading = ref(true)
const error = ref(false)
const overview = ref<AdminOverview>()
const trends = ref<AdminTrendReport>()
const today = new Date().toISOString().slice(0, 10)
const thirtyDaysAgo = new Date(Date.now() - 29 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10)
const from = ref(thirtyDaysAgo)
const to = ref(today)
const metrics = ['users', 'media', 'albums', 'folders', 'orders', 'payment_transactions'] as const
const trendColumns = ['users', 'media', 'albums', 'orders', 'payment_transactions'] as const
const maxActivity = computed(() => Math.max(1, ...(trends.value?.points || []).flatMap((point) => trendColumns.map((key) => point[key]))))

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = false
  try {
    const query = new URLSearchParams({ from: from.value, to: to.value })
    const [summary, report] = await Promise.all([generatedApi.overview(auth.token), generatedApi.statisticsTrends(query, auth.token)])
    overview.value = summary
    trends.value = report
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-2"><BarChart3 class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('statistics.title') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('statistics.description') }}</p></div></div>
      <Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ t('statistics.loadFailed') }}</p>
    <Card>
      <CardHeader><CardTitle>{{ t('statistics.range') }}</CardTitle></CardHeader>
      <CardContent class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <label class="flex flex-1 flex-col gap-2 text-sm font-medium">{{ t('statistics.from') }}<Input v-model="from" type="date" /></label>
        <label class="flex flex-1 flex-col gap-2 text-sm font-medium">{{ t('statistics.to') }}<Input v-model="to" type="date" /></label>
        <Button :disabled="loading || !from || !to" @click="load">{{ t('resource.refresh') }}</Button>
      </CardContent>
    </Card>
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"><Card v-for="metric in metrics" :key="metric"><CardHeader><CardTitle class="text-sm font-medium">{{ t(`statistics.metrics.${metric}`) }}</CardTitle></CardHeader><CardContent><Skeleton v-if="loading" class="h-9 w-20" /><p v-else class="text-3xl font-semibold">{{ overview?.[metric] ?? 0 }}</p></CardContent></Card></div>
    <Card>
      <CardHeader><CardTitle>{{ t('statistics.trend') }}</CardTitle><CardDescription>{{ trends?.from }} — {{ trends?.to }}</CardDescription></CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3"><Skeleton v-for="item in 6" :key="item" class="h-8 w-full" /></div>
        <p v-else-if="!trends?.points.length" class="text-sm text-muted-foreground">{{ t('statistics.empty') }}</p>
        <div v-else class="overflow-x-auto"><table class="w-full min-w-[760px] text-sm"><thead><tr class="border-b text-left text-muted-foreground"><th class="px-3 py-2">{{ t('statistics.from') }}</th><th v-for="column in trendColumns" :key="column" class="px-3 py-2">{{ t(`statistics.metrics.${column}`) }}</th><th class="px-3 py-2">{{ t('statistics.bandwidth') }}</th></tr></thead><tbody><tr v-for="point in trends.points" :key="point.date" class="border-b last:border-0"><td class="px-3 py-2 font-medium">{{ point.date }}</td><td v-for="column in trendColumns" :key="column" class="px-3 py-2"><div class="flex items-center gap-2"><span class="w-8">{{ point[column] }}</span><span class="h-1.5 rounded-full bg-primary/70" :style="{ width: `${Math.max(point[column] / maxActivity * 100, point[column] ? 4 : 0)}px` }" /></div></td><td class="px-3 py-2">{{ point.bandwidth_bytes.toLocaleString() }} {{ t('statistics.bytes') }}</td></tr></tbody></table></div>
      </CardContent>
    </Card>
  </div>
</template>
