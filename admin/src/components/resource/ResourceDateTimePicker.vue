<script setup lang="ts">
import { computed } from 'vue'
import { CalendarDate, getLocalTimeZone, parseDate, today } from '@internationalized/date'
import type { DateValue } from 'reka-ui'
import { Calendar } from '@/components/ui/calendar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const { t, locale } = useI18n()

const dateValue = computed<CalendarDate | undefined>(() => {
  const value = props.modelValue?.slice(0, 10)
  if (!value) return undefined
  try { return parseDate(value) }
  catch { return undefined }
})
const timeValue = computed(() => props.modelValue?.slice(11, 16) || '00:00')
const label = computed(() => {
  if (!dateValue.value) return t('resource.pickDateTime')
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(`${props.modelValue}:00`))
})

function update(date: DateValue | undefined, time = timeValue.value) {
  if (!date) return
  emit('update:modelValue', `${date.toString().slice(0, 10)}T${time || '00:00'}`)
}
function updateTime(event: Event) {
  update(dateValue.value || today(getLocalTimeZone()), (event.target as HTMLInputElement).value)
}
function clear() { emit('update:modelValue', '') }
</script>

<template>
  <div class="flex flex-wrap gap-2">
    <Popover>
      <PopoverTrigger as-child><Button type="button" variant="outline" class="min-w-48 justify-start font-normal">{{ label }}</Button></PopoverTrigger>
      <PopoverContent class="w-auto p-0" align="start"><Calendar :model-value="dateValue" @update:model-value="update" /></PopoverContent>
    </Popover>
    <Input :value="timeValue" type="time" class="w-28" :aria-label="t('resource.pickTime')" @input="updateTime" />
    <Button v-if="props.modelValue" type="button" variant="ghost" @click="clear">{{ t('resource.clearDateTime') }}</Button>
  </div>
</template>
