<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Image, LoaderCircle, RefreshCw } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { setPageSEO } from '@/lib/seo'
import { apiFetch, ApiError } from '@/lib/api'
import { useSitePresentation } from '@/lib/site-presentation'
import { useRoute, RouterLink } from 'vue-router'

interface PublicAlbumMedia {
  id: number
  original_name: string
  content_type: string
  width: number
  height: number
  size_bytes: number
  created_at?: string
  thumbnail_url: string
  original_url: string
}

interface PublicAlbum {
  id: number
  name: string
  visibility: 'public'
  media: PublicAlbumMedia[]
}

const { t, locale } = useI18n()
const route = useRoute()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const album = ref<PublicAlbum>()
const loading = ref(true)
const failed = ref(false)
const notFound = ref(false)
const albumID = computed(() => Number(route.params.id))

function formatSize(value: number) {
  if (!value) return ''
  const units = ['B', 'KB', 'MB', 'GB']
  let amount = value
  let unit = 0
  while (amount >= 1024 && unit < units.length - 1) { amount /= 1024; unit += 1 }
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(amount)} ${units[unit]}`
}

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

async function load() {
  loading.value = true
  failed.value = false
  notFound.value = false
  try {
    if (!albumID.value) throw new ApiError('public album not found', 404, 'PUBLIC_ALBUM_NOT_FOUND')
    album.value = await apiFetch<PublicAlbum>(`/api/v1/public/albums/${albumID.value}`)
    setPageSEO({
      title: `${album.value.name} · ${t('member.publicAlbum.siteName')}`,
      description: t('member.publicAlbum.description', { name: album.value.name }),
      path: `/a/${album.value.id}`,
      type: 'article',
    })
  } catch (error) {
    notFound.value = error instanceof ApiError && error.status === 404
    failed.value = !notFound.value
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  void loadSitePresentation()
  await load()
})
</script>

<template>
  <section class="space-y-8">
    <header class="max-w-3xl space-y-3">
      <Button variant="ghost" size="sm" as-child class="-ml-3"><RouterLink to="/discover"><ArrowLeft data-icon="inline-start" />{{ t('member.publicAlbum.backToDiscover') }}</RouterLink></Button>
      <p class="text-sm font-medium text-primary">{{ t('member.publicAlbum.eyebrow') }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ album?.name || t('member.publicAlbum.title') }}</h1>
      <p class="text-muted-foreground">{{ t('member.publicAlbum.description', { name: album?.name || t('member.publicAlbum.title') }) }}</p>
    </header>

    <Alert v-if="notFound" variant="destructive" class="max-w-2xl">
      <AlertTitle>{{ t('member.publicAlbum.notFoundTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.publicAlbum.notFoundDescription') }}</AlertDescription>
    </Alert>
    <Alert v-else-if="failed" variant="destructive" class="max-w-2xl">
      <AlertTitle>{{ t('member.publicAlbum.errorTitle') }}</AlertTitle>
      <AlertDescription class="flex items-center justify-between gap-4"><span>{{ t('member.publicAlbum.errorDescription') }}</span><Button variant="outline" size="sm" @click="load"><RefreshCw class="size-4" />{{ t('member.publicAlbum.retry') }}</Button></AlertDescription>
    </Alert>
    <div v-else-if="loading" class="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle class="size-4 animate-spin" />{{ t('member.publicAlbum.loading') }}</div>
    <div v-else-if="album?.media.length" class="columns-1 gap-4 sm:columns-2 lg:columns-3 xl:columns-4">
      <article v-for="item in album.media" :key="item.id" class="group mb-4 break-inside-avoid overflow-hidden rounded-lg border bg-card shadow-sm">
        <a :href="item.original_url" target="_blank" rel="noreferrer" class="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset">
          <img :src="item.thumbnail_url" :alt="item.original_name" loading="lazy" class="block h-auto w-full bg-muted object-cover transition-opacity group-hover:opacity-90" @error="handleImageError" />
        </a>
        <div class="space-y-1 px-3 py-3"><p class="truncate text-sm font-medium" :title="item.original_name">{{ item.original_name }}</p><p class="text-xs text-muted-foreground">{{ item.width }} × {{ item.height }} · {{ formatSize(item.size_bytes) }}</p></div>
      </article>
    </div>
    <div v-else class="rounded-lg border border-dashed p-10 text-center text-muted-foreground"><Image class="mx-auto mb-3 size-8" /><p class="font-medium text-foreground">{{ t('member.publicAlbum.emptyTitle') }}</p><p class="mt-1 text-sm">{{ t('member.publicAlbum.emptyDescription') }}</p></div>
  </section>
</template>
