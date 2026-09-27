<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Images, LoaderCircle, RotateCcw, Search, Trash2, Upload } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { apiFetch, apiFetchBlob, apiFetchEnvelope, ApiError } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

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
  variants: Record<string, MediaVariant>
}

interface UploadResponse {
  id: number
  upload_session_id: number
  status: string
  status_url: string
  error_code?: string
}

const auth = useAuthStore()
const { t } = useI18n()
const items = ref<MediaItem[]>([])
const previewURLs = ref<Record<number, string>>({})
const searchText = ref('')
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const uploading = ref(false)
const dragging = ref(false)
const showingTrash = ref(false)
const errorText = ref('')
const total = ref(0)
const pendingUploadIDs = ref<number[]>([])
const checkingUploadStatus = ref(false)
function releasePreviewURLs() {
  for (const url of Object.values(previewURLs.value)) URL.revokeObjectURL(url)
  previewURLs.value = {}
}

async function loadMedia() {
  if (!auth.token) return
  loading.value = true
  errorText.value = ''
  releasePreviewURLs()
  try {
    const params = new URLSearchParams({ page: '1', per_page: '48' })
    if (showingTrash.value) params.set('state', 'trash')
    if (searchText.value.trim()) params.set('search', searchText.value.trim())
    const response = await apiFetchEnvelope<MediaItem[]>(`/api/v1/media?${params}`, {}, auth.token)
    items.value = response.data
    total.value = Number(response.meta?.total ?? response.data.length)
    await Promise.allSettled(response.data.map(async (item) => {
      const original = item.variants?.original?.url
      if (!original || !auth.token) return
      const blob = await apiFetchBlob(original, auth.token)
      previewURLs.value[item.id] = URL.createObjectURL(blob)
    }))
  } catch {
    errorText.value = t('media.errors.loadFailed')
  } finally {
    loading.value = false
  }
}

async function uploadFiles(fileList: FileList | File[]) {
  if (!auth.token || uploading.value) return
  const selected = Array.from(fileList).filter((file) => /\.(jpe?g|png|gif)$/i.test(file.name))
  if (!selected.length) {
    errorText.value = t('media.errors.chooseFormat')
    return
  }
  uploading.value = true
  errorText.value = ''
  try {
    for (const file of selected) {
      const form = new FormData()
      form.append('file', file)
      const result = await apiFetch<UploadResponse>('/api/v1/uploads', {
        method: 'POST',
        headers: { 'Idempotency-Key': crypto.randomUUID() },
        body: form,
      }, auth.token)
      if (result.status === 'processing') pendingUploadIDs.value.push(result.upload_session_id)
    }
    await loadMedia()
  } catch (error) {
    await loadMedia()
    if (error instanceof ApiError && error.code === 'STORAGE_QUOTA_EXCEEDED') {
      errorText.value = t('media.errors.quotaExceeded')
    } else if (error instanceof ApiError && error.code === 'UPLOAD_FILE_TOO_LARGE') {
      errorText.value = t('media.errors.fileTooLarge')
    } else {
      errorText.value = t('media.errors.uploadFailed')
    }
  } finally {
    uploading.value = false
  }
}

async function checkUploadStatuses() {
  if (!auth.token || checkingUploadStatus.value || !pendingUploadIDs.value.length) return
  checkingUploadStatus.value = true
  const stillProcessing: number[] = []
  let hadFailedUpload = false
  try {
    for (const sessionID of pendingUploadIDs.value) {
      const status = await apiFetch<UploadResponse>(`/api/v1/uploads/${sessionID}`, {}, auth.token)
      if (status.status === 'processing') stillProcessing.push(sessionID)
      if (status.status === 'failed') hadFailedUpload = true
    }
    pendingUploadIDs.value = stillProcessing
    await loadMedia()
    if (hadFailedUpload) errorText.value = t('media.errors.processingFailed')
  } catch {
    errorText.value = t('media.errors.statusUnavailable')
  } finally {
    checkingUploadStatus.value = false
  }
}

async function retryPendingUploads() {
  if (!auth.token || checkingUploadStatus.value || !pendingUploadIDs.value.length) return
  checkingUploadStatus.value = true
  const stillProcessing: number[] = []
  let hadFailedUpload = false
  let leaseActive = false
  try {
    for (const sessionID of pendingUploadIDs.value) {
      try {
        const result = await apiFetch<UploadResponse>(`/api/v1/uploads/${sessionID}/retry`, { method: 'POST' }, auth.token)
        if (result.status === 'processing') stillProcessing.push(sessionID)
        if (result.status === 'failed') hadFailedUpload = true
      } catch (error) {
        if (error instanceof ApiError && error.code === 'UPLOAD_IN_PROGRESS') {
          stillProcessing.push(sessionID)
          leaseActive = true
          continue
        }
        throw error
      }
    }
    pendingUploadIDs.value = stillProcessing
    await loadMedia()
    if (hadFailedUpload) errorText.value = t('media.errors.objectVerificationFailed')
    else if (leaseActive) errorText.value = t('media.errors.recoveryNotReady')
  } catch {
    errorText.value = t('media.errors.recoveryUnavailable')
  } finally {
    checkingUploadStatus.value = false
  }
}

function handleFileInput(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files?.length) void uploadFiles(input.files)
  input.value = ''
}

function handleDrop(event: DragEvent) {
  dragging.value = false
  if (event.dataTransfer?.files.length) void uploadFiles(event.dataTransfer.files)
}

async function moveToTrash(item: MediaItem) {
  if (!auth.token) return
  try {
    await apiFetch(`/api/v1/media/${item.id}`, { method: 'DELETE' }, auth.token)
    await loadMedia()
  } catch {
    errorText.value = t('media.errors.moveToTrashFailed')
  }
}

async function restore(item: MediaItem) {
  if (!auth.token) return
  try {
    await apiFetch(`/api/v1/media/${item.id}/restore`, { method: 'POST' }, auth.token)
    await loadMedia()
  } catch {
    errorText.value = t('media.errors.restoreFailed')
  }
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

onMounted(() => void loadMedia())
onBeforeUnmount(releasePreviewURLs)
</script>

<template>
  <div class="mx-auto flex w-full max-w-7xl flex-col gap-6 p-5 md:p-8">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <div class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Images class="size-4" />{{ t('media.workspace') }}</div>
        <h1 class="text-3xl font-semibold tracking-tight">{{ t('media.title') }}</h1>
        <p class="mt-2 max-w-2xl text-sm text-muted-foreground">{{ t('media.description') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <Button :variant="showingTrash ? 'secondary' : 'outline'" @click="showingTrash = !showingTrash; void loadMedia()">
          <Trash2 data-icon="inline-start" />{{ showingTrash ? t('media.backToLibrary') : t('media.trash') }}
        </Button>
      <Button :disabled="uploading" @click="fileInput?.click()">
          <LoaderCircle v-if="uploading" class="size-4 animate-spin" />
          <Upload v-else data-icon="inline-start" />{{ uploading ? t('media.uploading') : t('media.uploadImages') }}
        </Button>
        <input ref="fileInput" class="hidden" type="file" accept="image/jpeg,image/png,image/gif,.jpg,.jpeg,.png,.gif" multiple @change="handleFileInput" />
      </div>
    </header>

    <Card :class="dragging ? 'border-primary bg-primary/5' : ''" @dragenter.prevent="dragging = true" @dragover.prevent="dragging = true" @dragleave.prevent="dragging = false" @drop.prevent="handleDrop">
      <CardContent class="flex min-h-32 flex-col items-center justify-center gap-2 text-center">
        <Upload class="size-6 text-muted-foreground" />
        <p class="text-sm font-medium">{{ t('media.dropImages') }}</p>
        <p class="text-xs text-muted-foreground">{{ t('media.formatsHint') }}</p>
      </CardContent>
    </Card>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="relative w-full sm:max-w-sm">
        <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input v-model="searchText" class="pl-9" :placeholder="t('media.searchPlaceholder')" @keydown.enter="loadMedia" />
      </div>
      <div class="flex items-center gap-2 text-sm text-muted-foreground">
        <span>{{ t('media.imageCount', { count: total }) }}</span>
        <Button variant="ghost" size="sm" :disabled="loading" @click="loadMedia">{{ t('media.refresh') }}</Button>
      </div>
    </div>

    <p v-if="errorText" role="alert" class="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">{{ errorText }}</p>

    <div v-if="pendingUploadIDs.length" role="status" class="flex flex-col gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
      <p class="text-sm">{{ t('media.pendingUploads', { count: pendingUploadIDs.length }) }}</p>
      <div class="flex items-center gap-2">
        <Button variant="ghost" size="sm" :disabled="checkingUploadStatus" @click="checkUploadStatuses">
          {{ t('media.checkStatus') }}
        </Button>
        <Button variant="outline" size="sm" :disabled="checkingUploadStatus" @click="retryPendingUploads">
          <LoaderCircle v-if="checkingUploadStatus" class="size-4 animate-spin" />
          {{ t('media.tryRecovery') }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Card v-for="n in 8" :key="n" class="h-64 animate-pulse bg-muted/40" />
    </div>
    <div v-else-if="items.length" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Card v-for="item in items" :key="item.id" class="overflow-hidden">
        <div class="flex aspect-[4/3] items-center justify-center bg-muted/40">
          <img v-if="previewURLs[item.id]" :src="previewURLs[item.id]" :alt="item.original_name" class="h-full w-full object-contain" />
          <Images v-else class="size-8 text-muted-foreground" />
        </div>
        <CardHeader class="gap-2 p-4">
          <div class="flex min-w-0 items-center justify-between gap-2">
            <CardTitle class="truncate text-sm" :title="item.original_name">{{ item.original_name }}</CardTitle>
            <Badge variant="outline">{{ item.content_type.replace('image/', '').toUpperCase() }}</Badge>
          </div>
          <CardDescription>{{ item.width }} × {{ item.height }} · {{ formatBytes(item.size_bytes) }}</CardDescription>
        </CardHeader>
        <CardContent class="flex justify-end gap-2 px-4 pb-4">
          <Button v-if="showingTrash" variant="outline" size="sm" @click="restore(item)"><RotateCcw data-icon="inline-start" />{{ t('media.restore') }}</Button>
          <Button v-else variant="ghost" size="sm" @click="moveToTrash(item)"><Trash2 data-icon="inline-start" />{{ t('media.moveToTrash') }}</Button>
        </CardContent>
      </Card>
    </div>
    <Card v-else>
      <CardContent class="flex min-h-64 flex-col items-center justify-center gap-3 text-center">
        <Images class="size-9 text-muted-foreground" />
        <div>
          <p class="font-medium">{{ showingTrash ? t('media.trashEmpty') : t('media.empty') }}</p>
          <p class="mt-1 text-sm text-muted-foreground">{{ showingTrash ? t('media.trashDescription') : t('media.emptyDescription') }}</p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
