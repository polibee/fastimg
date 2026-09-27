<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{ siteKey: string }>()
const emit = defineEmits<{ 'update:token': [token: string] }>()
const container = ref<HTMLElement | null>(null)
let widgetId: string | number | undefined

type TurnstileAPI = {
  render: (element: HTMLElement, options: Record<string, unknown>) => string | number
  remove?: (id: string | number) => void
  reset?: (id?: string | number) => void
}

function getTurnstile() {
  return (window as unknown as { turnstile?: TurnstileAPI }).turnstile
}

function loadScript() {
  if (getTurnstile()) return Promise.resolve()
  const existing = document.querySelector<HTMLScriptElement>('script[data-fastimg-turnstile]')
  if (existing) return new Promise<void>((resolve) => existing.addEventListener('load', () => resolve(), { once: true }))
  return new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.dataset.fastimgTurnstile = 'true'
    script.addEventListener('load', () => resolve(), { once: true })
    script.addEventListener('error', () => reject(new Error('Turnstile script could not load')), { once: true })
    document.head.appendChild(script)
  })
}

async function renderWidget() {
  if (!props.siteKey || !container.value) return
  try {
    await loadScript()
    const api = getTurnstile()
    if (!api || !container.value) return
    widgetId = api.render(container.value, {
      sitekey: props.siteKey,
      callback: (token: string) => emit('update:token', token),
      'expired-callback': () => emit('update:token', ''),
      'error-callback': () => emit('update:token', ''),
    })
  } catch {
    emit('update:token', '')
  }
}

onMounted(renderWidget)
watch(() => props.siteKey, () => {
  if (container.value) container.value.innerHTML = ''
  widgetId = undefined
  void renderWidget()
})
onBeforeUnmount(() => {
  const api = getTurnstile()
  if (api && widgetId !== undefined) api.remove?.(widgetId)
})
</script>

<template>
  <div ref="container" class="min-h-16" aria-label="Cloudflare Turnstile" />
</template>
