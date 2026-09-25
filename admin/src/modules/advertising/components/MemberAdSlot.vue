<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiFetch } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

type Placement = 'header' | 'footer' | 'left' | 'right'
type CreativeType = 'text' | 'image' | 'script'

interface AdSlot {
  id: number
  name: string
  placement: Placement
  creative_type: CreativeType
  creative_content: string
  target_url?: string
}

const props = defineProps<{ placement: Placement }>()
const { t } = useI18n()
const auth = useAuthStore()
const ads = ref<AdSlot[]>([])

function safeTarget(value?: string) {
  if (!value) return undefined
  try {
    const target = new URL(value, window.location.origin)
    return target.protocol === 'http:' || target.protocol === 'https:' ? target.href : undefined
  } catch {
    return undefined
  }
}

function scriptDocument(source: string) {
  const escaped = source.replace(/<\/script/gi, '<\\/script')
  return `<!doctype html><html><body><script>${escaped}<\/script></body></html>`
}

onMounted(async () => {
  if (!auth.token) return
  try {
    ads.value = await apiFetch<AdSlot[]>(`/api/v1/ads?placement=${encodeURIComponent(props.placement)}`, {}, auth.token)
  } catch {
    ads.value = []
  }
})
</script>

<template>
  <div v-if="ads.length" class="flex min-w-0 flex-col gap-2 py-2 text-xs text-muted-foreground">
    <span class="sr-only">{{ t('member.ads.label') }}</span>
    <div v-for="ad in ads" :key="ad.id" class="min-w-0 overflow-hidden">
      <span v-if="ad.creative_type === 'text'" class="break-words">{{ ad.creative_content }}</span>
      <a v-else-if="ad.creative_type === 'image' && safeTarget(ad.target_url)" :href="safeTarget(ad.target_url)" target="_blank" rel="noopener noreferrer">
        <img :src="ad.creative_content" :alt="t('member.ads.imageAlt')" class="max-h-28 w-full object-contain" loading="lazy">
      </a>
      <img v-else-if="ad.creative_type === 'image'" :src="ad.creative_content" :alt="t('member.ads.imageAlt')" class="max-h-28 w-full object-contain" loading="lazy">
      <iframe
        v-else-if="ad.creative_type === 'script'"
        :srcdoc="scriptDocument(ad.creative_content)"
        sandbox="allow-scripts"
        referrerpolicy="no-referrer"
        :title="ad.name || t('member.ads.label')"
        class="h-24 w-full border-0"
      />
    </div>
  </div>
</template>
