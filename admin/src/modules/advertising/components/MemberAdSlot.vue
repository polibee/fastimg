<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiFetch } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import { buildSandboxedAdDocument } from '@/modules/advertising/sandbox'

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
  <div v-if="ads.length" class="flex min-w-0 max-w-full flex-col gap-2 overflow-hidden py-2 text-xs text-muted-foreground">
    <span class="sr-only">{{ t('member.ads.label') }}</span>
    <article v-for="ad in ads" :key="ad.id" class="min-w-0 max-w-full overflow-hidden rounded-xl border border-border/60 bg-card/70 p-2 shadow-sm">
      <span v-if="ad.creative_type === 'text'" class="block min-w-0 break-words whitespace-pre-wrap">{{ ad.creative_content }}</span>
      <a v-else-if="ad.creative_type === 'image' && safeTarget(ad.target_url)" :href="safeTarget(ad.target_url)" target="_blank" rel="noopener noreferrer" class="block max-w-full overflow-hidden rounded-lg">
        <img :src="ad.creative_content" :alt="t('member.ads.imageAlt')" class="block max-h-32 max-w-full object-contain" loading="lazy">
      </a>
      <img v-else-if="ad.creative_type === 'image'" :src="ad.creative_content" :alt="t('member.ads.imageAlt')" class="block max-h-32 max-w-full rounded-lg object-contain" loading="lazy">
      <iframe
        v-else-if="ad.creative_type === 'script'"
        :srcdoc="buildSandboxedAdDocument(ad.creative_content)"
        sandbox="allow-scripts"
        scrolling="no"
        referrerpolicy="no-referrer"
        :title="ad.name || t('member.ads.label')"
        class="block h-24 w-full max-w-full overflow-hidden border-0"
      />
    </article>
  </div>
</template>
