<script setup lang="ts">
import { reactive, watch } from 'vue'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ modelValue: Record<string, unknown> }>()
const emit = defineEmits<{ 'update:modelValue': [value: Record<string, unknown>] }>()
const { t } = useI18n()
const numericFields = ['storage_bytes', 'max_file_bytes', 'daily_uploads', 'monthly_api_uploads', 'monthly_bandwidth_bytes', 'transform_count', 'api_rate_per_minute', 'token_limit']
const megabyteFields = ['storage_bytes', 'max_file_bytes', 'monthly_bandwidth_bytes']
const bytesPerMegabyte = 1024 * 1024
const draft = reactive<Record<string, number | boolean>>({})

function sync(value: Record<string, unknown>) {
  for (const field of numericFields) draft[field] = Number(value?.[field] ?? 0)
	draft.ads_enabled = Boolean(value?.ads_enabled)
	draft.watermark_enabled = Boolean(value?.watermark_enabled)
}
watch(() => props.modelValue, sync, { immediate: true, deep: true })
function emitChange() { emit('update:modelValue', { ...draft }) }
function setToggle(field: 'ads_enabled' | 'watermark_enabled', value: boolean) {
  draft[field] = Boolean(value)
  emitChange()
}
function numericValue(field: string) {
  const value = Number(draft[field] ?? 0)
  return megabyteFields.includes(field) ? Math.round((value / bytesPerMegabyte) * 100) / 100 : value
}
function unit(field: string) { return megabyteFields.includes(field) ? 'MB' : '' }
function setNumeric(field: string, value: string | number) {
  const parsed = Number(value || 0)
  draft[field] = megabyteFields.includes(field) ? Math.round(parsed * bytesPerMegabyte) : parsed
  emitChange()
}
</script>

<template>
  <div class="grid gap-3 rounded-md border bg-muted/20 p-3 sm:grid-cols-2">
    <label v-for="field in numericFields" :key="field" class="grid gap-1 text-sm"><span>{{ t(`resource.entitlements.${field}`) }}<span v-if="unit(field)" class="ml-1 text-muted-foreground">({{ unit(field) }})</span></span><Input :model-value="numericValue(field)" type="number" min="0" step="0.01" @update:model-value="setNumeric(field, $event)" /></label>
    <div class="flex items-center justify-between rounded-md border bg-background px-3 py-2 text-sm sm:col-span-2">
      <div class="flex min-w-0 flex-col gap-0.5">
        <span>{{ t('resource.entitlements.ads_enabled') }}</span>
        <span class="text-xs text-muted-foreground">{{ draft.ads_enabled ? t('resource.entitlements.enabled') : t('resource.entitlements.disabled') }}</span>
      </div>
      <Switch :aria-label="t('resource.entitlements.ads_enabled')" :model-value="Boolean(draft.ads_enabled)" @update:model-value="setToggle('ads_enabled', $event)" />
    </div>
    <div class="flex items-center justify-between rounded-md border bg-background px-3 py-2 text-sm sm:col-span-2">
      <div class="flex min-w-0 flex-col gap-0.5">
        <span>{{ t('resource.entitlements.watermark_enabled') }}</span>
        <span class="text-xs text-muted-foreground">{{ draft.watermark_enabled ? t('resource.entitlements.enabled') : t('resource.entitlements.disabled') }}</span>
      </div>
      <Switch :aria-label="t('resource.entitlements.watermark_enabled')" :model-value="Boolean(draft.watermark_enabled)" @update:model-value="setToggle('watermark_enabled', $event)" />
    </div>
  </div>
</template>
