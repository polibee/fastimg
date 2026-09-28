<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { FileText, Plus, RefreshCw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, errorMessageKey } from '@/lib/api'
import { contentApi, type SitePage } from '@/modules/content/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const pages = ref<SitePage[]>([])
const loading = ref(true)
const error = ref('')

function statusLabel(status: SitePage['status']) { return t(`content.status.${status}`) }
function localError(value: unknown) { return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('content.loadFailed') }

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try { pages.value = (await contentApi.list(auth.token)).data } catch (value) { error.value = localError(value) } finally { loading.value = false }
}

async function toggle(page: SitePage) {
  if (!auth.token) return
  try {
    page.status === 'published' ? await contentApi.archive(page.id, auth.token) : await contentApi.publish(page.id, auth.token)
    await load()
  } catch (value) { error.value = localError(value) }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-3"><FileText class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('content.pagesTitle') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('content.pagesDescription') }}</p></div></div>
      <div class="flex gap-2"><Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button><Button size="sm" as-child><RouterLink to="/admin/content-pages/new"><Plus class="mr-2 size-4" />{{ t('content.newPage') }}</RouterLink></Button></div>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card><CardHeader><CardTitle>{{ t('content.pagesTitle') }}</CardTitle><CardDescription>{{ t('content.pagesDescription') }}</CardDescription></CardHeader><CardContent>
      <div v-if="loading" class="py-8 text-sm text-muted-foreground">{{ t('resource.loading') }}</div>
      <div v-else-if="!pages.length" class="py-8 text-sm text-muted-foreground">{{ t('content.empty') }}</div>
      <div v-else class="divide-y rounded-md border">
        <div v-for="page in pages" :key="page.id" class="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0"><div class="flex flex-wrap items-center gap-2"><h2 class="font-medium">{{ page.title }}</h2><span class="rounded-full bg-muted px-2 py-0.5 text-xs">{{ statusLabel(page.status) }}</span></div><p class="mt-1 text-xs text-muted-foreground"><code>/{{ page.slug }}</code><span v-if="page.excerpt" class="ml-2">{{ page.excerpt }}</span></p></div>
          <div class="flex shrink-0 gap-2"><Button variant="outline" size="sm" as-child><RouterLink :to="`/admin/content-pages/${page.id}/edit`">{{ t('content.edit') }}</RouterLink></Button><Button variant="ghost" size="sm" @click="toggle(page)">{{ page.status === 'published' ? t('content.archive') : t('content.publish') }}</Button></div>
        </div>
      </div>
    </CardContent></Card>
  </div>
</template>
