import { ref } from 'vue'
import { apiFetch } from './api.ts'

const builtInSVG = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 360"><rect width="640" height="360" fill="#f3f4f6"/><path d="M0 300 170 150l100 90 85-110 285 170v60H0z" fill="#d1d5db"/><text x="320" y="190" fill="#6b7280" font-family="Arial,sans-serif" font-size="42" text-anchor="middle">FastImg</text><text x="320" y="235" fill="#9ca3af" font-family="Arial,sans-serif" font-size="18" text-anchor="middle">image unavailable</text></svg>'
export const builtInFallbackImageURL = `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(builtInSVG)}`

export function resolveSafeFallbackImageURL(value: string | undefined | null) {
  const candidate = String(value ?? '').trim()
  if (candidate.startsWith('/') && !candidate.startsWith('//')) return candidate
  try {
    const parsed = new URL(candidate)
    if (parsed.protocol === 'http:' || parsed.protocol === 'https:') return parsed.toString()
  } catch {
    // Use the built-in fallback for malformed or unsafe settings.
  }
  return builtInFallbackImageURL
}

const fallbackImageURL = ref(builtInFallbackImageURL)
let loaded = false
let loading: Promise<void> | undefined

export function useSitePresentation() {
  async function loadSitePresentation() {
    if (loaded) return
    if (!loading) {
      loading = apiFetch<{ watermark_fallback_image_url?: string }>('/api/v1/site/presentation')
        .then((presentation) => {
          fallbackImageURL.value = resolveSafeFallbackImageURL(presentation.watermark_fallback_image_url)
          loaded = true
        })
        .catch(() => {
          loaded = true
        })
    }
    await loading
  }

  return { fallbackImageURL, loadSitePresentation }
}
