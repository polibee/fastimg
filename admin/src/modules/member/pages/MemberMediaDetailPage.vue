<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Check, Copy, Image, Link2, LoaderCircle } from '@lucide/vue'
import { RouterLink, useRoute } from 'vue-router'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { apiFetch, apiFetchBlob, apiFetchEnvelope } from '@/lib/api'
import { useSitePresentation } from '@/lib/site-presentation'
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
  format: string
  size_bytes: number
  width: number
  height: number
  status: string
  visibility: 'private' | 'link' | 'public'
  moderation_status: string
  created_at: string
  links: Record<string, string>
  variants: Record<string, MediaVariant>
}

interface ShareLink {
  url: string
  expires_at: string | null
}

interface HotlinkPolicy {
  media_id: number
  mode: 'off' | 'referer' | 'signed' | 'hybrid'
  allow_no_referer: boolean
}

interface HotlinkDomain {
  id: number
  host: string
  status: string
}

interface SignedURL {
  url: string
  variant: string
  expires_at: string
}

const { t, locale } = useI18n()
const auth = useAuthStore()
const { fallbackImageURL, loadSitePresentation } = useSitePresentation()
const route = useRoute()
const item = ref<MediaItem>()
const previews = ref<Record<string, string>>({})
const loading = ref(true)
const failed = ref(false)
const copiedKey = ref('')
const shareExpiry = ref('')
const sharePassword = ref('')
const shareLink = ref<ShareLink>()
const shareError = ref(false)
const creatingShare = ref(false)
const signedVariant = ref('original')
const signedExpiresIn = ref('600')
const signedURL = ref<SignedURL>()
const creatingSignedURL = ref(false)
const signedURLError = ref(false)
const visibility = ref<'private' | 'link' | 'public'>('public')
const savingVisibility = ref(false)
const visibilityError = ref(false)
const visibilitySaved = ref(false)
const hotlinkPolicy = ref<HotlinkPolicy>({ media_id: 0, mode: 'off', allow_no_referer: false })
const savingHotlinkPolicy = ref(false)
const hotlinkPolicyError = ref(false)
const hotlinkPolicySaved = ref(false)
const hotlinkDomains = ref<HotlinkDomain[]>([])
const newHotlinkDomain = ref('')
const hotlinkDomainError = ref(false)
const addingHotlinkDomain = ref(false)
const variantNames = ['original'] as const
const linkKeys = ['url', 'markdown', 'html', 'bbcode'] as const
const availableVariants = computed(() => variantNames.filter((name) => item.value?.variants?.[name]))
const shareURL = computed(() => shareLink.value?.url ? new URL(shareLink.value.url, globalThis.location?.origin ?? 'http://localhost').toString() : '')
const signedURLAbsolute = computed(() => signedURL.value?.url ? new URL(signedURL.value.url, globalThis.location?.origin ?? 'http://localhost').toString() : '')

function releasePreviews() {
  Object.values(previews.value).forEach((url) => URL.revokeObjectURL(url))
  previews.value = {}
}

function handleImageError(event: Event) {
  const image = event.target as HTMLImageElement
  if (image.src !== fallbackImageURL.value) image.src = fallbackImageURL.value
}

async function loadDetails() {
  loading.value = true
  failed.value = false
  releasePreviews()
  try {
    if (!auth.token) throw new Error('unauthenticated')
    const response = await apiFetchEnvelope<MediaItem>(`/api/v1/media/${encodeURIComponent(String(route.params.id))}`, {}, auth.token)
    item.value = response.data
    visibility.value = response.data.visibility || 'public'
    await loadLinkSecurity()
    await Promise.allSettled(availableVariants.value.map(async (name) => {
      const blob = await apiFetchBlob(item.value!.variants[name].url, auth.token!)
      previews.value[name] = URL.createObjectURL(blob)
    }))
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

async function saveVisibility() {
  if (!auth.token || !item.value) return
  savingVisibility.value = true
  visibilityError.value = false
  visibilitySaved.value = false
  try {
    await apiFetch(`/api/v1/media/${item.value.id}/visibility`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ visibility: visibility.value }),
    }, auth.token)
    item.value.visibility = visibility.value
    visibilitySaved.value = true
  } catch {
    visibilityError.value = true
  } finally {
    savingVisibility.value = false
  }
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

function formatDate(value: string) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

async function copyLink(key: typeof linkKeys[number]) {
  if (!item.value?.links?.[key] || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(item.value.links[key])
    copiedKey.value = key
  } catch {
    copiedKey.value = ''
  }
}

async function createShareLink() {
  if (!auth.token || !item.value) return
  creatingShare.value = true
  shareError.value = false
  try {
    const expiresAt = shareExpiry.value ? new Date(`${shareExpiry.value}T23:59:59`).toISOString() : null
    shareLink.value = await apiFetch<ShareLink>(`/api/v1/media/${item.value.id}/share-links`, {
      method: 'POST',
      body: JSON.stringify({ expires_at: expiresAt, password: sharePassword.value.trim() || null }),
    }, auth.token)
  } catch {
    shareError.value = true
  } finally {
    creatingShare.value = false
  }
}

async function loadLinkSecurity() {
  if (!auth.token || !item.value) return
  try {
    const [policy, domains] = await Promise.all([
      apiFetchEnvelope<HotlinkPolicy>(`/api/v1/media/${item.value.id}/hotlink-policy`, {}, auth.token),
      apiFetchEnvelope<HotlinkDomain[]>('/api/v1/hotlink-domains', {}, auth.token),
    ])
    hotlinkPolicy.value = policy.data
    hotlinkDomains.value = domains.data
  } catch {
    hotlinkPolicyError.value = true
  }
}

async function createSignedURL() {
  if (!auth.token || !item.value) return
  creatingSignedURL.value = true
  signedURLError.value = false
  try {
    signedURL.value = await apiFetch<SignedURL>(`/api/v1/media/${item.value.id}/signed-url`, {
      method: 'POST',
      body: JSON.stringify({ variant: signedVariant.value, expires_in: Number(signedExpiresIn.value) }),
    }, auth.token)
  } catch {
    signedURLError.value = true
  } finally {
    creatingSignedURL.value = false
  }
}

async function copySignedURL() {
  if (!signedURL.value?.url || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(new URL(signedURL.value.url, window.location.origin).toString())
    copiedKey.value = 'signed'
  } catch {
    copiedKey.value = ''
  }
}

async function saveHotlinkPolicy() {
  if (!auth.token || !item.value) return
  savingHotlinkPolicy.value = true
  hotlinkPolicyError.value = false
  hotlinkPolicySaved.value = false
  try {
    hotlinkPolicy.value = await apiFetch<HotlinkPolicy>(`/api/v1/media/${item.value.id}/hotlink-policy`, {
      method: 'PUT',
      body: JSON.stringify({ mode: hotlinkPolicy.value.mode, allow_no_referer: hotlinkPolicy.value.allow_no_referer }),
    }, auth.token)
    hotlinkPolicySaved.value = true
  } catch {
    hotlinkPolicyError.value = true
  } finally {
    savingHotlinkPolicy.value = false
  }
}

async function addHotlinkDomain() {
  if (!auth.token || !newHotlinkDomain.value.trim()) return
  addingHotlinkDomain.value = true
  hotlinkDomainError.value = false
  try {
    const domain = await apiFetch<HotlinkDomain>('/api/v1/hotlink-domains', {
      method: 'POST',
      body: JSON.stringify({ domain: newHotlinkDomain.value.trim() }),
    }, auth.token)
    hotlinkDomains.value = [...hotlinkDomains.value, domain]
    newHotlinkDomain.value = ''
  } catch {
    hotlinkDomainError.value = true
  } finally {
    addingHotlinkDomain.value = false
  }
}

async function removeHotlinkDomain(domain: HotlinkDomain) {
  if (!auth.token) return
  try {
    await apiFetch(`/api/v1/hotlink-domains/${domain.id}`, { method: 'DELETE' }, auth.token)
    hotlinkDomains.value = hotlinkDomains.value.filter((item) => item.id !== domain.id)
  } catch {
    hotlinkDomainError.value = true
  }
}

async function copyShareURL() {
  if (!shareLink.value?.url || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(new URL(shareLink.value.url, window.location.origin).toString())
    copiedKey.value = 'share'
  } catch {
    copiedKey.value = ''
  }
}

onMounted(() => void Promise.all([loadDetails(), loadSitePresentation()]))
onBeforeUnmount(releasePreviews)
</script>

<template>
  <div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <Button variant="ghost" size="sm" as-child class="mb-3 -ml-3"><RouterLink to="/media"><ArrowLeft data-icon="inline-start" />{{ t('member.media.backToLibrary') }}</RouterLink></Button>
        <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Image />{{ t('member.media.workspace') }}</p>
        <h1 class="text-3xl font-semibold tracking-tight">{{ t('member.media.detailTitle') }}</h1>
      </div>
      <Badge v-if="item" variant="outline">{{ item.content_type.replace('image/', '').toUpperCase() }}</Badge>
    </header>

    <Alert v-if="failed" variant="destructive">
      <AlertTitle>{{ t('member.media.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t('member.media.errors.detailLoadFailed') }}</AlertDescription>
    </Alert>

    <div v-if="loading" class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
      <Card class="h-[28rem] animate-pulse bg-muted/40" />
      <Card class="h-72 animate-pulse bg-muted/40" />
    </div>
    <template v-else-if="item">
      <section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Card class="overflow-hidden">
          <CardHeader>
            <CardTitle>{{ t('member.media.variants') }}</CardTitle>
            <CardDescription>{{ item.original_name }}</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 sm:grid-cols-2">
            <div v-for="name in availableVariants" :key="name" class="overflow-hidden rounded-lg border">
              <div class="flex aspect-[4/3] items-center justify-center bg-muted/40">
                <img :src="previews[name] || fallbackImageURL" :alt="`${item.original_name} ${t(`member.media.${name}`)}`" class="h-full w-full object-contain" @error="handleImageError" />
              </div>
              <div class="flex items-center justify-between gap-3 p-3 text-sm">
                <span class="font-medium">{{ t(`member.media.${name}`) }}</span>
                <span class="text-muted-foreground">{{ formatBytes(item.variants[name].size_bytes) }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card class="h-fit">
          <CardHeader>
            <CardTitle>{{ t('member.media.metadata') }}</CardTitle>
          </CardHeader>
          <CardContent class="flex flex-col gap-4 text-sm">
            <div><p class="text-muted-foreground">{{ t('member.media.fileName') }}</p><p class="mt-1 break-all font-medium">{{ item.original_name }}</p></div>
            <div><p class="text-muted-foreground">{{ t('member.media.format') }}</p><p class="mt-1 font-medium">{{ item.format.toUpperCase() }}</p></div>
            <div><p class="text-muted-foreground">{{ t('member.media.dimensions') }}</p><p class="mt-1 font-medium">{{ item.width }} × {{ item.height }}</p></div>
            <div><p class="text-muted-foreground">{{ t('member.media.fileSize') }}</p><p class="mt-1 font-medium">{{ formatBytes(item.size_bytes) }}</p></div>
            <div><p class="text-muted-foreground">{{ t('member.media.uploadedAt') }}</p><p class="mt-1 font-medium">{{ formatDate(item.created_at) }}</p></div>
          </CardContent>
        </Card>
      </section>
      <Card>
        <CardHeader>
          <CardTitle>{{ t('member.media.visibilityTitle') }}</CardTitle>
          <CardDescription>{{ t('member.media.visibilityDescription') }}</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-col gap-3 sm:flex-row sm:items-end">
          <label class="flex max-w-sm flex-1 flex-col gap-1 text-sm">
            <span class="text-muted-foreground">{{ t('member.media.visibility') }}</span>
            <select v-model="visibility" class="h-9 rounded-md border border-input bg-background px-3 text-sm">
              <option value="public">{{ t('member.media.visibilityOptions.public') }}</option>
              <option value="link">{{ t('member.media.visibilityOptions.link') }}</option>
              <option value="private">{{ t('member.media.visibilityOptions.private') }}</option>
            </select>
          </label>
          <Button :disabled="savingVisibility" @click="saveVisibility">
            <LoaderCircle v-if="savingVisibility" class="animate-spin" data-icon="inline-start" />
            {{ visibilitySaved ? t('member.media.saved') : t('member.media.saveVisibility') }}
          </Button>
        </CardContent>
        <CardContent v-if="visibilityError" class="pt-0 text-sm text-destructive">{{ t('member.media.errors.visibilityFailed') }}</CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>{{ t('member.media.linkFormats') }}</CardTitle>
          <CardDescription>{{ item.original_name }}</CardDescription>
        </CardHeader>
        <CardContent class="grid gap-3">
          <div v-for="key in linkKeys" :key="key" class="grid gap-2 sm:grid-cols-[7rem_minmax(0,1fr)_auto] sm:items-center">
            <span class="text-sm font-medium">{{ t(`member.media.links.${key}`) }}</span>
            <Input readonly :model-value="item.links[key]" :aria-label="t(`member.media.links.${key}`)" />
            <Button variant="outline" size="sm" @click="copyLink(key)">
              <Check v-if="copiedKey === key" data-icon="inline-start" />
              <Copy v-else data-icon="inline-start" />
              {{ copiedKey === key ? t('member.media.copied') : t('member.media.copyLink') }}
            </Button>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2"><Link2 />{{ t('member.media.signedURLTitle') }}</CardTitle>
          <CardDescription>{{ t('member.media.signedURLDescription') }}</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-col gap-4">
          <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] sm:items-end">
            <label class="flex flex-col gap-1 text-sm"><span class="text-muted-foreground">{{ t('member.media.signedVariant') }}</span><select v-model="signedVariant" class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm"><option v-for="name in availableVariants" :key="name" :value="name">{{ t(`member.media.${name}`) }}</option></select></label>
            <label class="flex flex-col gap-1 text-sm"><span class="text-muted-foreground">{{ t('member.media.signedExpiry') }}</span><select v-model="signedExpiresIn" class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm"><option value="300">{{ t('member.media.signedExpiryOptions.fiveMinutes') }}</option><option value="600">{{ t('member.media.signedExpiryOptions.tenMinutes') }}</option><option value="3600">{{ t('member.media.signedExpiryOptions.oneHour') }}</option><option value="86400">{{ t('member.media.signedExpiryOptions.oneDay') }}</option></select></label>
            <Button :disabled="creatingSignedURL" @click="createSignedURL"><LoaderCircle v-if="creatingSignedURL" class="animate-spin" data-icon="inline-start" /><Link2 v-else data-icon="inline-start" />{{ t('member.media.createSignedURL') }}</Button>
          </div>
          <Alert v-if="signedURLError" variant="destructive"><AlertTitle>{{ t('member.media.errorTitle') }}</AlertTitle><AlertDescription>{{ t('member.media.errors.signedURLFailed') }}</AlertDescription></Alert>
          <div v-if="signedURL" class="flex flex-col gap-2 rounded-lg border bg-muted/30 p-3"><span class="text-sm font-medium">{{ t('member.media.signedURLCreated', { expires: formatDate(signedURL.expires_at) }) }}</span><div class="flex flex-col gap-2 sm:flex-row"><Input readonly :model-value="signedURLAbsolute" /><Button variant="outline" size="sm" @click="copySignedURL"><Check v-if="copiedKey === 'signed'" data-icon="inline-start" /><Copy v-else data-icon="inline-start" />{{ copiedKey === 'signed' ? t('member.media.copied') : t('member.media.copyLink') }}</Button></div></div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>{{ t('member.media.hotlinkPolicyTitle') }}</CardTitle><CardDescription>{{ t('member.media.hotlinkPolicyDescription') }}</CardDescription></CardHeader>
        <CardContent class="flex flex-col gap-4">
          <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:items-end"><label class="flex flex-col gap-1 text-sm"><span class="text-muted-foreground">{{ t('member.media.hotlinkMode') }}</span><select v-model="hotlinkPolicy.mode" class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm"><option value="off">{{ t('member.media.hotlinkModes.off') }}</option><option value="referer">{{ t('member.media.hotlinkModes.referer') }}</option><option value="signed">{{ t('member.media.hotlinkModes.signed') }}</option><option value="hybrid">{{ t('member.media.hotlinkModes.hybrid') }}</option></select></label><label class="flex items-center gap-2 pb-1 text-sm"><input v-model="hotlinkPolicy.allow_no_referer" type="checkbox" />{{ t('member.media.allowNoReferer') }}</label><Button :disabled="savingHotlinkPolicy" @click="saveHotlinkPolicy">{{ hotlinkPolicySaved ? t('member.media.saved') : t('member.media.savePolicy') }}</Button></div>
          <Alert v-if="hotlinkPolicyError" variant="destructive"><AlertTitle>{{ t('member.media.errorTitle') }}</AlertTitle><AlertDescription>{{ t('member.media.errors.policyFailed') }}</AlertDescription></Alert>
          <div class="border-t pt-4"><p class="mb-2 text-sm font-medium">{{ t('member.media.hotlinkDomainsTitle') }}</p><div class="flex gap-2"><Input v-model="newHotlinkDomain" :placeholder="t('member.media.hotlinkDomainPlaceholder')" @keyup.enter="addHotlinkDomain" /><Button variant="outline" :disabled="addingHotlinkDomain" @click="addHotlinkDomain">{{ t('member.media.addDomain') }}</Button></div><Alert v-if="hotlinkDomainError" class="mt-3" variant="destructive"><AlertDescription>{{ t('member.media.errors.domainFailed') }}</AlertDescription></Alert><ul v-if="hotlinkDomains.length" class="mt-3 divide-y rounded-lg border text-sm"><li v-for="domain in hotlinkDomains" :key="domain.id" class="flex items-center justify-between gap-3 px-3 py-2"><span class="break-all">{{ domain.host }}</span><Button variant="ghost" size="sm" @click="removeHotlinkDomain(domain)">{{ t('member.media.removeDomain') }}</Button></li></ul><p v-else class="mt-3 text-sm text-muted-foreground">{{ t('member.media.noHotlinkDomains') }}</p></div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2"><Link2 />{{ t('member.media.shareTitle') }}</CardTitle>
          <CardDescription>{{ t('member.media.shareDescription') }}</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-col gap-4">
          <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] sm:items-end">
            <label class="flex flex-1 flex-col gap-1 text-sm">
              <span class="text-muted-foreground">{{ t('member.media.shareExpiry') }}</span>
              <input v-model="shareExpiry" type="date" class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm" />
            </label>
            <label class="flex flex-1 flex-col gap-1 text-sm">
              <span class="text-muted-foreground">{{ t('member.media.sharePassword') }}</span>
              <input v-model="sharePassword" type="password" maxlength="72" autocomplete="new-password" :placeholder="t('member.media.sharePasswordHint')" class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm" />
            </label>
            <Button :disabled="creatingShare" @click="createShareLink">
              <LoaderCircle v-if="creatingShare" class="animate-spin" data-icon="inline-start" />
              <Link2 v-else data-icon="inline-start" />
              {{ t('member.media.createShare') }}
            </Button>
          </div>
          <Alert v-if="shareError" variant="destructive">
            <AlertTitle>{{ t('member.shareLinks.errorTitle') }}</AlertTitle>
            <AlertDescription>{{ t('member.shareLinks.errors.createFailed') }}</AlertDescription>
          </Alert>
          <div v-if="shareLink" class="flex flex-col gap-2 rounded-lg border bg-muted/30 p-3">
            <span class="text-sm font-medium">{{ t('member.media.shareCreated') }}</span>
            <div class="flex flex-col gap-2 sm:flex-row">
              <input readonly :value="shareURL" class="h-8 min-w-0 flex-1 rounded-lg border border-input bg-background px-2.5 text-sm" />
              <Button variant="outline" size="sm" @click="copyShareURL">
                <Check v-if="copiedKey === 'share'" data-icon="inline-start" />
                <Copy v-else data-icon="inline-start" />
                {{ copiedKey === 'share' ? t('member.media.copied') : t('member.media.copyLink') }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>
