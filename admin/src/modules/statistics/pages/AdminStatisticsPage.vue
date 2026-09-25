<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { BarChart3, RefreshCw } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { generatedApi, type AdminOverview } from '@/generated/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n(); const auth = useAuthStore(); const loading = ref(true); const overview = ref<AdminOverview>(); const error = ref(false)
const metrics = ['users', 'media', 'albums', 'folders', 'orders', 'payment_transactions'] as const
async function load() { if (!auth.token) return; loading.value = true; error.value = false; try { overview.value = await generatedApi.overview(auth.token) } catch { error.value = true } finally { loading.value = false } }
onMounted(load)
</script>
<template><div class="flex flex-col gap-6"><div class="flex items-center justify-between"><div class="flex items-center gap-2"><BarChart3 class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('statistics.title') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('statistics.description') }}</p></div></div><Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button></div><p v-if="error" class="text-sm text-destructive">{{ t('statistics.loadFailed') }}</p><div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"><Card v-for="metric in metrics" :key="metric"><CardHeader><CardTitle class="text-sm font-medium">{{ t(`statistics.metrics.${metric}`) }}</CardTitle></CardHeader><CardContent><Skeleton v-if="loading" class="h-9 w-20" /><p v-else class="text-3xl font-semibold">{{ overview?.[metric] ?? 0 }}</p></CardContent></Card></div></div></template>
