<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { publicContentApi, type SitePage } from '@/modules/content/api'
import ContentDocument from '@/modules/content/components/ContentDocument.vue'
import { setPageSEO } from '@/lib/seo'
import { useI18n } from 'vue-i18n'

const route = useRoute(); const { t } = useI18n(); const page = ref<Pick<SitePage, 'title' | 'excerpt' | 'seo_title' | 'seo_description'> & { content: SitePage['content_json'] }>(); const loading = ref(true); const error = ref(false)
onMounted(async () => { try { const value = await publicContentApi.show(String(route.params.slug)); page.value = value; setPageSEO({ title: value.seo_title || value.title, description: value.seo_description || value.excerpt || value.title, path: `/page/${route.params.slug}` }) } catch { error.value = true } finally { loading.value = false } })
</script>

<template><article class="mx-auto max-w-3xl"><div v-if="loading" class="space-y-4"><Skeleton class="h-10 w-2/3" /><Skeleton class="h-64 w-full" /></div><Alert v-else-if="error" variant="destructive"><AlertTitle>{{ t('content.publicLoadFailed') }}</AlertTitle><AlertDescription>{{ t('content.publicLoadDescription') }}</AlertDescription></Alert><template v-else-if="page"><header class="mb-8 space-y-3"><h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ page.title }}</h1><p v-if="page.excerpt" class="text-muted-foreground">{{ page.excerpt }}</p></header><ContentDocument :document="page.content" /></template></article></template>
