<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute } from 'vue-router'
import { ArrowDown, ArrowLeft, ArrowUp, Library, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { apiFetch, apiFetchBlob, apiFetchEnvelope } from '@/lib/api'
import { useSitePresentation } from '@/lib/site-presentation'
import { useAuthStore } from '@/stores/auth'

interface AlbumItem {
  id: number
  name: string
  visibility?: string
}

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
  variants: Record<string, MediaVariant>
}

const route = useRoute()
const { t } = useI18n()
const auth = useAuthStore()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const album = ref<AlbumItem>()
const albums = ref<AlbumItem[]>([])
const items = ref<MediaItem[]>([])
const selected = ref<number[]>([])
const destinationAlbumID = ref('')
const previewURLs = ref<Record<number, string>>({})
const loading = ref(true)
const errorKey = ref('')
const orderMessageKey = ref('')
const orderDirty = ref(false)
const savingOrder = ref(false)
const draggedID = ref<number>()
const albumID = computed(() => Number(route.params.id))
const allSelected = computed(() => items.value.length > 0 && selected.value.length === items.value.length)

function releasePreviewURLs() {
  for (const url of Object.values(previewURLs.value)) URL.revokeObjectURL(url)
  previewURLs.value = {}
}

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

function toggleSelection(id: number) {
  selected.value = selected.value.includes(id)
    ? selected.value.filter((itemID) => itemID !== id)
    : [...selected.value, id]
}

function toggleAll() {
  selected.value = allSelected.value ? [] : items.value.map((item) => item.id)
}

function moveItem(sourceID: number, targetID: number) {
  const sourceIndex = items.value.findIndex((item) => item.id === sourceID)
  const targetIndex = items.value.findIndex((item) => item.id === targetID)
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return
  const nextItems = [...items.value]
  const [item] = nextItems.splice(sourceIndex, 1)
  nextItems.splice(targetIndex, 0, item)
  items.value = nextItems
  orderDirty.value = true
  orderMessageKey.value = ''
}

function moveBy(id: number, offset: number) {
  const index = items.value.findIndex((item) => item.id === id)
  const target = items.value[index + offset]
  if (target) moveItem(id, target.id)
}

function startDrag(id: number, event: DragEvent) {
  draggedID.value = id
  event.dataTransfer?.setData('text/plain', String(id))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function dropItem(targetID: number) {
  if (draggedID.value) moveItem(draggedID.value, targetID)
  draggedID.value = undefined
}

async function saveOrder() {
  if (!auth.token || !orderDirty.value || !items.value.length) return
  savingOrder.value = true
  errorKey.value = ''
  orderMessageKey.value = ''
  try {
    await apiFetch(`/api/v1/me/albums/${albumID.value}/media/order`, {
      method: 'PATCH',
      body: JSON.stringify({ media_ids: items.value.map((item) => item.id) }),
    }, auth.token)
    orderDirty.value = false
    orderMessageKey.value = 'member.albums.orderSaved'
  } catch {
    errorKey.value = 'member.albums.errors.orderFailed'
  } finally {
    savingOrder.value = false
  }
}

async function load() {
  if (!auth.token || !albumID.value) {
    loading.value = false
    errorKey.value = 'member.albums.errors.loadMediaFailed'
    return
  }
  loading.value = true
  errorKey.value = ''
  orderMessageKey.value = ''
  orderDirty.value = false
  selected.value = []
  releasePreviewURLs()
  try {
    const [albumsResponse, idsResponse] = await Promise.all([
      apiFetchEnvelope<AlbumItem[]>('/api/v1/me/albums', {}, auth.token),
      apiFetchEnvelope<number[]>(`/api/v1/me/albums/${albumID.value}/media`, {}, auth.token),
    ])
    albums.value = albumsResponse.data
    album.value = albumsResponse.data.find((item) => item.id === albumID.value)
    if (!album.value) throw new Error('album not found')
    const details = await Promise.allSettled(idsResponse.data.map((id) => apiFetch<MediaItem>(`/api/v1/media/${id}`, {}, auth.token)))
    items.value = details.flatMap((result) => result.status === 'fulfilled' ? [result.value] : [])
    await Promise.allSettled(items.value.map(async (item) => {
      const original = item.variants?.original?.url
      if (!original || !auth.token) return
      const blob = await apiFetchBlob(original, auth.token)
      previewURLs.value[item.id] = URL.createObjectURL(blob)
    }))
  } catch {
    errorKey.value = 'member.albums.errors.loadMediaFailed'
  } finally {
    loading.value = false
  }
}

async function removeSelected() {
  if (!auth.token || !selected.value.length) return
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/me/albums/${albumID.value}/media`, {
      method: 'DELETE',
      body: JSON.stringify({ media_ids: selected.value }),
    }, auth.token)
    await load()
  } catch {
    errorKey.value = 'member.albums.errors.removeMediaFailed'
  }
}

async function moveSelected() {
  if (!auth.token || !selected.value.length || !destinationAlbumID.value) return
  errorKey.value = ''
  try {
    await apiFetch(`/api/v1/me/albums/${destinationAlbumID.value}/media/move`, {
      method: 'POST',
      body: JSON.stringify({ source_album_id: albumID.value, media_ids: selected.value }),
    }, auth.token)
    destinationAlbumID.value = ''
    selected.value = []
    await load()
  } catch {
    errorKey.value = 'member.albums.errors.moveMediaFailed'
  }
}

onMounted(() => void Promise.all([load(), loadSitePresentation()]))
onBeforeUnmount(releasePreviewURLs)
</script>

<template>
  <div class="flex flex-col gap-7">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <Button variant="ghost" size="sm" as-child class="mb-3 -ml-3"><RouterLink to="/albums"><ArrowLeft data-icon="inline-start" />{{ t('member.albums.backToAlbums') }}</RouterLink></Button>
        <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Library />{{ t('member.albums.workspace') }}</p>
        <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ album?.name || t('member.albums.title') }}</h1>
        <p class="mt-3 text-muted-foreground">{{ t('member.albums.mediaDescription') }}</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button variant="outline" :disabled="!selected.length" @click="removeSelected"><Trash2 data-icon="inline-start" />{{ t('member.albums.removeSelected', { count: selected.length }) }}</Button>
        <Button :disabled="!orderDirty || savingOrder" @click="saveOrder">{{ savingOrder ? t('member.albums.savingOrder') : t('member.albums.saveOrder') }}</Button>
      </div>
    </header>

    <Alert v-if="errorKey" variant="destructive">
      <AlertTitle>{{ t('member.albums.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t(errorKey) }}</AlertDescription>
    </Alert>

    <div class="flex items-center gap-3 text-sm text-muted-foreground">
      <label class="flex items-center gap-2"><input type="checkbox" :checked="allSelected" @change="toggleAll" />{{ t('member.albums.selectAll') }}</label>
      <span>{{ t('member.albums.selected', { count: selected.length }) }}</span>
      <span v-if="orderDirty">{{ t('member.albums.orderPending') }}</span>
      <span v-if="orderMessageKey" class="text-emerald-600">{{ t(orderMessageKey) }}</span>
    </div>

    <div class="flex flex-wrap items-end gap-3 rounded-lg border border-dashed border-border p-3">
      <div class="min-w-56 flex-1">
        <label class="mb-2 block text-sm font-medium" for="album-move-target">{{ t('member.albums.moveTarget') }}</label>
        <select id="album-move-target" v-model="destinationAlbumID" class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
          <option value="">{{ t('member.albums.chooseMoveTarget') }}</option>
          <option v-for="entry in albums.filter((item) => item.id !== albumID)" :key="entry.id" :value="String(entry.id)">{{ entry.name }}</option>
        </select>
      </div>
      <Button variant="outline" :disabled="!selected.length || !destinationAlbumID" @click="moveSelected">{{ t('member.albums.moveSelected', { count: selected.length }) }}</Button>
    </div>

    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Skeleton v-for="index in 8" :key="index" class="h-64 rounded-lg" />
    </div>
    <div v-else-if="items.length" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Card v-for="(item, index) in items" :key="item.id" draggable="true" class="overflow-hidden" @dragstart="startDrag(item.id, $event)" @dragover.prevent @drop="dropItem(item.id)">
        <div class="relative flex aspect-[4/3] items-center justify-center bg-muted/40">
          <img :src="previewURLs[item.id] || fallbackImageURL" :alt="item.original_name" class="h-full w-full object-contain" @error="handleImageError" />
          <label class="absolute left-3 top-3 rounded-md bg-background/90 p-1.5 shadow-sm"><input type="checkbox" :checked="selected.includes(item.id)" :aria-label="item.original_name" @change="toggleSelection(item.id)" /></label>
        </div>
        <CardHeader class="gap-2 p-4">
          <div class="flex min-w-0 items-center justify-between gap-2"><CardTitle class="truncate text-sm" :title="item.original_name">{{ item.original_name }}</CardTitle><Badge variant="outline">{{ item.content_type.replace('image/', '').toUpperCase() }}</Badge></div>
          <CardDescription>{{ item.width }} × {{ item.height }}</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-between gap-2 px-4 pb-4">
          <div class="flex gap-1">
            <Button variant="ghost" size="icon" :disabled="index === 0" :aria-label="t('member.albums.moveUp', { name: item.original_name })" @click="moveBy(item.id, -1)"><ArrowUp /></Button>
            <Button variant="ghost" size="icon" :disabled="index === items.length - 1" :aria-label="t('member.albums.moveDown', { name: item.original_name })" @click="moveBy(item.id, 1)"><ArrowDown /></Button>
          </div>
          <Button variant="outline" size="sm" as-child><RouterLink :to="{ name: 'member-media-detail', params: { id: item.id } }">{{ t('member.media.details') }}</RouterLink></Button>
        </CardContent>
      </Card>
    </div>
    <Card v-else>
      <Empty>
        <EmptyHeader><EmptyMedia variant="icon"><Library /></EmptyMedia><EmptyTitle>{{ t('member.albums.emptyMedia') }}</EmptyTitle><EmptyDescription>{{ t('member.albums.emptyMediaDescription') }}</EmptyDescription></EmptyHeader>
      </Empty>
    </Card>
  </div>
</template>
