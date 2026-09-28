<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { publicEndpointURL } from '@/lib/api'

interface StatusData { status: string; updated_at: string; components: Array<{ name: string; status: string }>; maintenance: Array<{ id: number; title: string; body?: string; severity?: string; starts_at?: string; ends_at?: string }> }
const { t } = useI18n(); const data = ref<StatusData>(); const loading = ref(true); const error = ref(false)
async function load() { loading.value = true; error.value = false; try { const response = await fetch(publicEndpointURL('/api/v1/status')); if (!response.ok) throw new Error('status'); const payload = await response.json() as { data: StatusData }; data.value = payload.data } catch { error.value = true } finally { loading.value = false } }
onMounted(load)
</script>

<template>
  <div class="mx-auto grid max-w-4xl gap-6">
    <div><p class="text-sm font-medium text-primary">{{ t('statusPage.eyebrow') }}</p><h1 class="mt-2 text-3xl font-semibold tracking-tight">{{ t('statusPage.title') }}</h1><p class="mt-2 text-muted-foreground">{{ t('statusPage.description') }}</p></div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('statusPage.errorTitle') }}</AlertTitle><AlertDescription><Button variant="outline" size="sm" class="mt-3" @click="load">{{ t('statusPage.retry') }}</Button></AlertDescription></Alert>
    <Card v-if="!error"><CardHeader><CardTitle>{{ t('statusPage.components') }}</CardTitle></CardHeader><CardContent><div v-if="loading" class="text-sm text-muted-foreground">{{ t('statusPage.loading') }}</div><div v-else class="grid gap-3 sm:grid-cols-2"><div v-for="component in data?.components ?? []" :key="component.name" class="flex items-center justify-between rounded-lg border px-4 py-3"><span>{{ t(`statusPage.componentNames.${component.name}`, component.name) }}</span><Badge :variant="component.status === 'operational' ? 'secondary' : 'destructive'">{{ t(`statusPage.statuses.${component.status}`, component.status) }}</Badge></div></div></CardContent></Card>
    <Card v-if="!error && (data?.maintenance?.length ?? 0) > 0"><CardHeader><CardTitle>{{ t('statusPage.maintenance') }}</CardTitle></CardHeader><CardContent class="grid gap-4"><article v-for="item in data?.maintenance" :key="item.id" class="border-l-2 border-primary pl-4"><h2 class="font-medium">{{ item.title }}</h2><p v-if="item.body" class="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{{ item.body }}</p></article></CardContent></Card>
  </div>
</template>
