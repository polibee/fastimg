<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle, Check, Copy, ExternalLink, LoaderCircle, RotateCcw, Upload } from '@lucide/vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { useAuthStore } from '@/stores/auth'
import { useSitePresentation } from '@/lib/site-presentation'
import { useMemberUpload } from '../composables/useMemberUpload'

const props = defineProps<{ compact?: boolean }>()
const emit = defineEmits<{ updated: [] }>()
const { t } = useI18n()
const auth = useAuthStore()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const fileInput = ref<HTMLInputElement>()
const copiedKey = ref<string>()
const uploadLinkKeys = ['url', 'markdown', 'html', 'bbcode'] as const
const { items, uploading, uploadFiles, retry } = useMemberUpload(() => emit('updated'))

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

function linkStateKey(itemID: string, key: string) {
  return `${itemID}:${key}`
}

async function copyLink(itemID: string, key: typeof uploadLinkKeys[number], value?: string) {
  if (!value || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(value)
    copiedKey.value = linkStateKey(itemID, key)
    window.setTimeout(() => {
      if (copiedKey.value === linkStateKey(itemID, key)) copiedKey.value = undefined
    }, 1500)
  } catch {
    copiedKey.value = undefined
  }
}

function handleFileInput(event: Event) {
  if (!auth.isAuthenticated) return
  const input = event.target as HTMLInputElement
  if (input.files?.length) void uploadFiles(input.files)
  input.value = ''
}

function handleDrop(event: DragEvent) {
  if (!auth.isAuthenticated) return
  if (event.dataTransfer?.files.length) void uploadFiles(event.dataTransfer.files)
}

function clipboardImageFiles(event: ClipboardEvent) {
  const clipboard = event.clipboardData
  if (!clipboard) return []

  const files = Array.from(event.clipboardData?.files ?? []).filter((file) => file.type.startsWith('image/'))
  if (files.length) return files

  return Array.from(clipboard.items)
    .filter((item) => item.kind === 'file' && item.type.startsWith('image/'))
    .map((item) => item.getAsFile())
    .filter((file): file is File => Boolean(file))
}

function handlePaste(event: ClipboardEvent) {
  if (!auth.isAuthenticated || uploading.value) return
  const files = clipboardImageFiles(event)
  if (!files.length) return

  event.preventDefault()
  void uploadFiles(files)
}

onMounted(() => {
  void loadSitePresentation()
  window.addEventListener('paste', handlePaste)
})

onBeforeUnmount(() => window.removeEventListener('paste', handlePaste))
</script>

<template>
  <Card class="border-dashed" @dragover.prevent @drop.prevent="handleDrop">
    <CardContent class="flex flex-col gap-4 px-5 py-7" :class="compact ? 'sm:flex-row sm:items-center sm:justify-between' : 'min-h-40 items-center justify-center text-center'">
      <div class="flex items-center gap-4" :class="compact ? '' : 'flex-col'">
        <div class="flex size-12 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary"><Upload /></div>
        <div>
          <p class="font-medium">{{ t('member.upload.dropFiles') }}</p>
          <p class="mt-1 text-sm text-muted-foreground">{{ t('member.media.formatsHint') }}</p>
        </div>
      </div>
      <div v-if="auth.isAuthenticated" class="flex items-center gap-3" :class="compact ? '' : 'flex-col'">
        <Button :disabled="uploading" @click="fileInput?.click()">
          <LoaderCircle v-if="uploading" class="size-4 animate-spin" />
          <Upload v-else data-icon="inline-start" />{{ uploading ? t('member.media.uploading') : t('member.upload.choose') }}
        </Button>
        <span class="text-xs text-muted-foreground">{{ t('member.upload.privateHint') }}</span>
        <span class="text-xs text-muted-foreground">{{ t('member.upload.pasteHint') }}</span>
      </div>
      <div v-else class="flex flex-col items-center gap-2 text-center sm:items-end sm:text-right">
        <p class="text-sm text-muted-foreground">{{ t('member.upload.loginRequired') }}</p>
        <Button as-child><RouterLink :to="{ name: 'login', query: { redirect: '/' } }">{{ t('member.upload.loginToUpload') }}</RouterLink></Button>
      </div>
      <input ref="fileInput" class="hidden" type="file" accept="image/jpeg,image/png,image/gif" multiple @change="handleFileInput" />
    </CardContent>
    <CardContent v-if="items.length" class="flex flex-col gap-2 border-t pt-4">
      <div v-for="item in items" :key="item.id" class="flex flex-col gap-3 rounded-md bg-muted/40 px-3 py-2 text-sm">
        <div class="flex items-center justify-between gap-3">
          <span class="min-w-0 truncate">{{ item.file.name || t('member.upload.invalidFile') }}</span>
          <span v-if="item.state === 'uploading'" class="flex shrink-0 items-center gap-1 text-muted-foreground"><LoaderCircle class="size-4 animate-spin" />{{ t('member.upload.uploading') }}</span>
          <span v-else-if="item.state === 'processing'" class="shrink-0 text-muted-foreground">{{ t('member.upload.processing') }}</span>
          <span v-else-if="item.state === 'ready'" class="flex shrink-0 items-center gap-1 text-primary"><Check class="size-4" />{{ t('member.upload.ready') }}</span>
          <div v-else class="flex shrink-0 items-center gap-2">
            <span class="text-destructive">{{ t(item.errorKey || 'member.media.errors.uploadFailed') }}</span>
            <Button v-if="item.errorKey !== 'member.media.errors.chooseFormat'" variant="ghost" size="sm" @click="retry(item)"><RotateCcw data-icon="inline-start" />{{ t('member.upload.retry') }}</Button>
          </div>
        </div>
        <div v-if="item.state === 'ready' && item.links" class="grid gap-2 border-t pt-3">
          <div v-if="item.links.url" class="flex flex-col gap-3 rounded-lg border bg-background p-3 sm:flex-row sm:items-center">
            <div class="flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-md bg-muted/40">
              <img :src="item.links.url" :alt="t('member.upload.previewAlt', { name: item.file.name })" class="size-full object-contain" loading="lazy" @error="handleImageError" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="font-medium">{{ t('member.upload.resultTitle') }}</p>
              <p class="mt-1 truncate text-sm text-muted-foreground" :title="item.file.name">{{ item.file.name }}</p>
              <p class="mt-1 text-xs text-muted-foreground">{{ t('member.upload.resultDescription') }}</p>
              <Button class="mt-2" variant="outline" size="sm" as-child>
                <a :href="item.links.url" target="_blank" rel="noreferrer"><ExternalLink data-icon="inline-start" />{{ t('member.upload.openImage') }}</a>
              </Button>
            </div>
          </div>
          <p class="font-medium">{{ t('member.upload.linksTitle') }}</p>
          <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
            <Button v-for="key in uploadLinkKeys" :key="key" type="button" variant="outline" size="sm" class="min-w-0 justify-start gap-2 text-left" :title="item.links[key]" :aria-label="`${t(`member.upload.links.${key}`)}: ${item.links[key]}`" @click="copyLink(item.id, key, item.links[key])">
              <span class="shrink-0 font-medium">{{ t(`member.upload.links.${key}`) }}</span>
              <span class="min-w-0 truncate text-xs text-muted-foreground">{{ item.links[key] }}</span>
              <Check v-if="copiedKey === linkStateKey(item.id, key)" data-icon="inline-start" />
              <Copy v-else data-icon="inline-start" />
              <span class="sr-only">{{ copiedKey === linkStateKey(item.id, key) ? t('member.upload.copied') : t('member.upload.copyLink') }}</span>
            </Button>
          </div>
        </div>
      </div>
      <Alert v-if="items.some((item) => item.errorKey === 'member.media.errors.statusUnavailable')" variant="destructive">
        <AlertCircle /><AlertDescription>{{ t('member.media.errors.statusUnavailable') }}</AlertDescription>
      </Alert>
    </CardContent>
  </Card>
</template>
