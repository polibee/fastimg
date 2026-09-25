<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { AlertCircle, Image, Images, RotateCcw, Search, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { apiFetch, apiFetchBlob, apiFetchEnvelope } from '@/lib/api'
import { useSitePresentation } from '@/lib/site-presentation'
import { useAuthStore } from '@/stores/auth'
import MemberUploadPanel from '../components/MemberUploadPanel.vue'

interface MediaVariant {
  url: string
  content_type: string
  size_bytes: number
  width: number
  height: number
}

interface MediaItem {
  id: number
  original_name: string
  content_type: string
  size_bytes: number
  width: number
  height: number
  status: string
  folder_id?: number | null
  variants: Record<string, MediaVariant>
}

interface MemberFolder {
  id: number
  name: string
}

interface MemberAlbum {
  id: number
  name: string
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const items = ref<MediaItem[]>([])
const folders = ref<MemberFolder[]>([])
const albums = ref<MemberAlbum[]>([])
const previewURLs = ref<Record<number, string>>({})
const loading = ref(true)
const showingTrash = ref(false)
const searchText = ref('')
const errorKey = ref('')
const noticeKey = ref('')
const emptyTrashDialogOpen = ref(false)
const emptyingTrash = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = 48
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

function releasePreviewURLs() {
  for (const url of Object.values(previewURLs.value)) URL.revokeObjectURL(url)
  previewURLs.value = {}
}

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

async function loadMedia() {
  if (!auth.token) {
    loading.value = false
    errorKey.value = 'member.media.errors.loadFailed'
    return
  }
  loading.value = true
  errorKey.value = ''
  noticeKey.value = ''
  releasePreviewURLs()
  try {
    const params = new URLSearchParams({ page: String(currentPage.value), per_page: String(pageSize) })
    if (showingTrash.value) params.set('state', 'trash')
    if (searchText.value.trim()) params.set('search', searchText.value.trim())
    const response = await apiFetchEnvelope<MediaItem[]>(`/api/v1/media?${params}`, {}, auth.token)
    items.value = response.data
    total.value = Number(response.meta?.total ?? response.data.length)
    if (currentPage.value > pageCount.value) {
      currentPage.value = pageCount.value
      await loadMedia()
      return
    }
    await Promise.allSettled(response.data.map(async (item) => {
      const thumbnail = item.variants?.thumbnail?.url
      if (!thumbnail || !auth.token) return
      const blob = await apiFetchBlob(thumbnail, auth.token)
      previewURLs.value[item.id] = URL.createObjectURL(blob)
    }))
  } catch {
    errorKey.value = 'member.media.errors.loadFailed'
  } finally {
    loading.value = false
  }
}

async function loadFolders() {
  if (!auth.token) return
  try {
    const response = await apiFetchEnvelope<MemberFolder[]>('/api/v1/me/folders', {}, auth.token)
    folders.value = response.data
  } catch {
    folders.value = []
  }
}

async function loadAlbums() {
  if (!auth.token) return
  try {
    const response = await apiFetchEnvelope<MemberAlbum[]>('/api/v1/me/albums', {}, auth.token)
    albums.value = response.data
  } catch {
    albums.value = []
  }
}

async function moveToFolder(item: MediaItem, event: Event) {
  if (!auth.token) return
  const value = (event.target as HTMLSelectElement).value
  const folderID = value ? Number(value) : null
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/media/${item.id}/folder`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ folder_id: folderID }),
    }, auth.token)
    item.folder_id = folderID
  } catch {
    errorKey.value = 'member.media.errors.moveToFolderFailed'
  }
}

async function addToAlbum(item: MediaItem, event: Event) {
  if (!auth.token) return
  const albumID = Number((event.target as HTMLSelectElement).value)
  if (!albumID) return
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/me/albums/${albumID}/media`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ media_ids: [item.id] }),
    }, auth.token)
    ;(event.target as HTMLSelectElement).value = ''
  } catch {
    errorKey.value = 'member.media.errors.addToAlbumFailed'
  }
}

async function moveToTrash(item: MediaItem) {
  if (!auth.token) return
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/media/${item.id}`, { method: 'DELETE' }, auth.token)
    if (items.value.length === 1 && currentPage.value > 1) currentPage.value -= 1
    await loadMedia()
  } catch {
    errorKey.value = 'member.media.errors.moveToTrashFailed'
  }
}

async function restore(item: MediaItem) {
  if (!auth.token) return
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/media/${item.id}/restore`, { method: 'POST' }, auth.token)
    await loadMedia()
  } catch {
    errorKey.value = 'member.media.errors.restoreFailed'
  }
}

async function emptyTrash() {
  if (!auth.token) return
  emptyingTrash.value = true
  errorKey.value = ''
  try {
    await apiFetch('/api/v1/media/trash', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ confirm: 'empty-trash' }),
    }, auth.token)
    emptyTrashDialogOpen.value = false
    currentPage.value = 1
    noticeKey.value = 'member.media.emptyTrashSuccess'
    await loadMedia()
  } catch {
    errorKey.value = 'member.media.errors.emptyTrashFailed'
  } finally {
    emptyingTrash.value = false
  }
}

function searchMedia() {
  currentPage.value = 1
  void loadMedia()
}

function changeTrashState() {
  showingTrash.value = !showingTrash.value
  currentPage.value = 1
  void loadMedia()
}

function changePage(page: number) {
  currentPage.value = Math.min(pageCount.value, Math.max(1, page))
  void loadMedia()
}

function formatBytes(bytes: number) {
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit += 1
  }
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)} ${units[unit]}`
}

onMounted(() => void Promise.all([loadMedia(), loadFolders(), loadAlbums(), loadSitePresentation()]))
onBeforeUnmount(() => {
  releasePreviewURLs()
})
</script>

<template>
  <div class="flex flex-col gap-7">
    <section class="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
      <div class="max-w-2xl">
        <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Images />{{ t('member.media.workspace') }}</p>
        <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('member.media.heading') }}</h1>
        <p class="mt-3 text-muted-foreground">{{ t('member.media.description') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <Button :variant="showingTrash ? 'secondary' : 'outline'" @click="changeTrashState">
          <Trash2 data-icon="inline-start" />{{ showingTrash ? t('member.media.backToLibrary') : t('member.media.trash') }}
        </Button>
      </div>
    </section>

    <MemberUploadPanel v-if="!showingTrash" @updated="loadMedia" />

    <Alert v-if="errorKey" variant="destructive">
      <AlertCircle />
      <AlertTitle>{{ t('member.media.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t(errorKey) }}</AlertDescription>
    </Alert>

    <section class="flex flex-col gap-4">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-xl font-semibold">{{ showingTrash ? t('member.media.trash') : t('member.media.libraryTitle') }}</h2>
          <p class="mt-1 text-sm text-muted-foreground">{{ t('member.media.imageCount', { count: total }) }}</p>
        </div>
        <div class="flex items-center gap-2">
          <div class="relative w-full sm:w-64">
            <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="searchText" class="pl-9" :placeholder="t('member.media.searchPlaceholder')" @keydown.enter="searchMedia" />
          </div>
          <Button variant="ghost" :disabled="loading" @click="loadMedia">{{ t('member.media.refresh') }}</Button>
          <Button v-if="showingTrash && total > 0" variant="outline" :disabled="loading || emptyingTrash" @click="emptyTrashDialogOpen = true"><Trash2 data-icon="inline-start" />{{ t('member.media.emptyTrash') }}</Button>
        </div>
      </div>

      <Alert v-if="noticeKey"><AlertDescription>{{ t(noticeKey) }}</AlertDescription></Alert>

      <div v-if="loading" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <Skeleton v-for="index in 8" :key="index" class="h-32 rounded-lg sm:h-36" />
      </div>
      <div v-else-if="items.length" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <Card v-for="item in items" :key="item.id" class="gap-0 overflow-hidden py-0">
          <div class="flex h-32 min-h-0 items-center justify-center bg-muted/40 sm:h-36">
            <img :src="previewURLs[item.id] || fallbackImageURL" :alt="item.original_name" class="h-full w-full object-contain" @error="handleImageError" />
          </div>
          <CardHeader class="gap-1 p-2">
            <div class="flex min-w-0 items-center justify-between gap-2">
              <CardTitle class="truncate text-sm" :title="item.original_name">{{ item.original_name }}</CardTitle>
              <Badge variant="outline">{{ item.content_type.replace('image/', '').toUpperCase() }}</Badge>
            </div>
            <CardDescription>{{ item.width }} × {{ item.height }} · {{ formatBytes(item.size_bytes) }}</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-1 px-2 pb-2">
            <div v-if="!showingTrash" class="grid grid-cols-2 gap-1.5">
            <label class="min-w-0" :aria-label="t('member.media.folder')">
              <span class="sr-only">{{ t('member.media.folder') }}</span>
              <select class="h-7 w-full min-w-0 rounded-md border border-input bg-background px-1 text-[11px] text-foreground" :value="item.folder_id ?? ''" @change="moveToFolder(item, $event)">
                <option value="">{{ t('member.media.unfiled') }}</option>
                <option v-for="folder in folders" :key="folder.id" :value="folder.id">{{ folder.name }}</option>
              </select>
            </label>
            <label class="min-w-0" :aria-label="t('member.media.album')">
              <span class="sr-only">{{ t('member.media.album') }}</span>
              <select class="h-7 w-full min-w-0 rounded-md border border-input bg-background px-1 text-[11px] text-foreground" @change="addToAlbum(item, $event)">
                <option value="">{{ t('member.media.addToAlbum') }}</option>
                <option v-for="album in albums" :key="album.id" :value="album.id">{{ album.name }}</option>
              </select>
            </label>
            </div>
            <div class="flex justify-end gap-2">
            <Button v-if="!showingTrash" variant="outline" size="sm" as-child><RouterLink :to="{ name: 'member-media-detail', params: { id: item.id } }">{{ t('member.media.details') }}</RouterLink></Button>
            <Button v-if="showingTrash" variant="outline" size="sm" @click="restore(item)"><RotateCcw data-icon="inline-start" />{{ t('member.media.restore') }}</Button>
            <Button v-else variant="ghost" size="sm" @click="moveToTrash(item)"><Trash2 data-icon="inline-start" />{{ t('member.media.moveToTrash') }}</Button>
            </div>
          </CardContent>
        </Card>
      </div>
      <Card v-else class="min-h-64">
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon"><Images /></EmptyMedia>
            <EmptyTitle>{{ showingTrash ? t('member.media.trashEmpty') : t('member.media.empty') }}</EmptyTitle>
            <EmptyDescription>{{ showingTrash ? t('member.media.trashDescription') : t('member.media.emptyDescription') }}</EmptyDescription>
          </EmptyHeader>
        </Empty>
      </Card>
      <nav v-if="!loading && pageCount > 1" class="flex items-center justify-between gap-3" :aria-label="t('member.media.pagination')">
        <Button variant="outline" :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">
          {{ t('member.media.previousPage') }}
        </Button>
        <span class="text-sm text-muted-foreground" aria-live="polite">
          {{ t('member.media.pageStatus', { current: currentPage, total: pageCount }) }}
        </span>
        <Button variant="outline" :disabled="currentPage >= pageCount" @click="changePage(currentPage + 1)">
          {{ t('member.media.nextPage') }}
        </Button>
      </nav>
    </section>

    <AlertDialog v-model:open="emptyTrashDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t('member.media.emptyTrashTitle') }}</AlertDialogTitle>
          <AlertDialogDescription>{{ t('member.media.emptyTrashDescription') }}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t('member.media.cancel') }}</AlertDialogCancel>
          <AlertDialogAction :disabled="emptyingTrash" @click="emptyTrash">{{ emptyingTrash ? t('member.media.emptyingTrash') : t('member.media.emptyTrashConfirm') }}</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
