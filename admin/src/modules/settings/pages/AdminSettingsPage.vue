<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { Save, Settings2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { generatedApi, type SystemSetting } from '@/generated/api'
import { ApiError, errorMessageKey } from '@/lib/api'
import { buildSettingUpdates } from '@/modules/settings/save'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const initialValues = reactive<Record<string, string>>({})
const groups = [
  { key: 'seo', fields: ['site_title', 'site_description', 'site_keywords', 'robots', 'sitemap.enabled', 'sitemap.extra_paths'] },
  { key: 'gateway', fields: ['payment.default_gateway', 'payment.fake.enabled', 'payment.paypal.enabled', 'payment.paypal.environment', 'payment.paypal.client_id', 'payment.paypal.client_secret', 'payment.paypal.webhook_id', 'payment.paypal.base_url', 'payment.paypal.webhook_url', 'payment.paypal.return_url', 'payment.paypal.cancel_url', 'payment.xcash.enabled', 'payment.xcash.app_id', 'payment.xcash.hmac_key', 'payment.xcash.base_url', 'payment.xcash.callback_url', 'payment.xcash.return_url', 'payment.nowpayments.enabled', 'payment.nowpayments.api_key', 'payment.nowpayments.ipn_secret', 'payment.nowpayments.base_url', 'payment.nowpayments.callback_url', 'payment.nowpayments.success_url', 'payment.nowpayments.cancel_url'] },
  { key: 'statistics', fields: ['stats.enabled', 'stats.retention_days'] },
  { key: 'code', fields: ['custom.site_verification', 'custom.ad_verification', 'custom.head', 'custom.body'] },
  { key: 'other', fields: ['site_url', 'upload.max_file_mb', 'discover.enabled', 'watermark.text', 'watermark.domain', 'watermark.fallback_image_url'] },
]
const values = reactive<Record<string, string>>({})
const gatewaySections = [
  { key: 'paypal', fields: ['payment.paypal.enabled', 'payment.paypal.environment', 'payment.paypal.client_id', 'payment.paypal.client_secret', 'payment.paypal.webhook_id'] },
  { key: 'xcash', fields: ['payment.xcash.enabled', 'payment.xcash.app_id', 'payment.xcash.hmac_key'] },
  { key: 'nowpayments', fields: ['payment.nowpayments.enabled', 'payment.nowpayments.api_key', 'payment.nowpayments.ipn_secret'] },
]
const gatewayRegistrationLinks: Record<string, string> = {
  xcash: 'https://dash.xca.sh/register?ref=2GWV5MKT',
  nowpayments: 'https://account.nowpayments.io/create-account?link_id=3940543227',
}
const definitions: Record<string, { type: string; group: string; multiline?: boolean }> = {}
for (const group of groups) for (const key of group.fields) definitions[key] = { type: key.endsWith('.enabled') || key === 'stats.enabled' || key === 'discover.enabled' ? 'boolean' : key.endsWith('retention_days') || key.endsWith('max_file_mb') ? 'integer' : ['app_id', 'hmac_key', 'api_key', 'ipn_secret', 'client_id', 'client_secret', 'webhook_id'].some((part) => key.endsWith(part)) ? 'secret' : 'string', group: group.key, multiline: key.includes('description') || key === 'custom.head' || key === 'custom.body' || key === 'sitemap.extra_paths' }

function fieldLabel(key: string) { return t(`settings.fields.${key.replaceAll('.', '_')}`, key) }
function fieldHint(key: string) { return t(`settings.hints.${key.replaceAll('.', '_')}`, '') }
function isBoolean(key: string) { return definitions[key]?.type === 'boolean' }
function isSecret(key: string) { return definitions[key]?.type === 'secret' }
function fieldValue(key: string) { return values[key] === '__configured__' ? '' : values[key] || '' }
function setValue(key: string, value: string | number) { values[key] = String(value ?? '') }
function setBoolean(key: string, event: Event) { values[key] = (event.target as HTMLInputElement).checked ? 'true' : 'false' }
function setSelectValue(key: string, event: Event) { values[key] = (event.target as HTMLSelectElement).value }
function siteOrigin() {
  return String(values.site_url || window.location.origin).replace(/\/+$/, '')
}
type GatewayURLKind = 'api' | 'webhook' | 'return' | 'cancel'
function gatewayURL(provider: string, kind: GatewayURLKind) {
  const site = siteOrigin()
  if (kind === 'api') {
    if (provider === 'paypal') return ['production', 'live'].includes(values['payment.paypal.environment']) ? 'https://api-m.paypal.com' : 'https://api-m.sandbox.paypal.com'
    if (provider === 'xcash') return 'https://pay.xca.sh'
    return 'https://api.nowpayments.io'
  }
  if (kind === 'webhook') return `${site}/api/v1/payment-gateways/${provider}/webhook`
  if (kind === 'cancel') return `${site}/orders/{order_id}?payment=cancelled`
  return `${site}/orders/{order_id}?payment=success`
}
function gatewayURLFields(provider: string) {
  if (provider === 'paypal') return [
    { key: 'payment.paypal.base_url', kind: 'api' as const, label: 'settings.fields.payment_paypal_base_url', hint: 'settings.hints.payment_paypal_base_url' },
    { key: 'payment.paypal.webhook_url', kind: 'webhook' as const, label: 'settings.fields.payment_paypal_webhook_url', hint: 'settings.hints.payment_paypal_webhook_url' },
    { key: 'payment.paypal.return_url', kind: 'return' as const, label: 'settings.fields.payment_paypal_return_url', hint: 'settings.hints.payment_paypal_return_url' },
    { key: 'payment.paypal.cancel_url', kind: 'cancel' as const, label: 'settings.fields.payment_paypal_cancel_url', hint: 'settings.hints.payment_paypal_cancel_url' },
  ]
  if (provider === 'xcash') return [
    { key: 'payment.xcash.base_url', kind: 'api' as const, label: 'settings.fields.payment_xcash_base_url', hint: 'settings.hints.payment_xcash_base_url' },
    { key: 'payment.xcash.callback_url', kind: 'webhook' as const, label: 'settings.fields.payment_xcash_callback_url', hint: 'settings.hints.payment_xcash_callback_url' },
    { key: 'payment.xcash.return_url', kind: 'return' as const, label: 'settings.fields.payment_xcash_return_url', hint: 'settings.hints.payment_xcash_return_url' },
  ]
  return [
    { key: 'payment.nowpayments.base_url', kind: 'api' as const, label: 'settings.fields.payment_nowpayments_base_url', hint: 'settings.hints.payment_nowpayments_base_url' },
    { key: 'payment.nowpayments.callback_url', kind: 'webhook' as const, label: 'settings.fields.payment_nowpayments_callback_url', hint: 'settings.hints.payment_nowpayments_callback_url' },
    { key: 'payment.nowpayments.success_url', kind: 'return' as const, label: 'settings.fields.payment_nowpayments_success_url', hint: 'settings.hints.payment_nowpayments_success_url' },
    { key: 'payment.nowpayments.cancel_url', kind: 'cancel' as const, label: 'settings.fields.payment_nowpayments_cancel_url', hint: 'settings.hints.payment_nowpayments_cancel_url' },
  ]
}
const gatewayAutoValues = reactive<Record<string, string>>({})
function applyGatewayDefaults() {
  for (const section of gatewaySections) {
    for (const item of gatewayURLFields(section.key)) {
      const next = gatewayURL(section.key, item.kind)
      if (!values[item.key] || values[item.key] === gatewayAutoValues[item.key]) values[item.key] = next
      gatewayAutoValues[item.key] = next
    }
  }
}

function snapshotInitialValues() {
  for (const key of Object.keys(definitions)) initialValues[key] = values[key] || (definitions[key].type === 'boolean' ? 'false' : '')
}

onMounted(async () => {
  if (!auth.token) return
  try {
    const settings = await generatedApi.settings(auth.token)
    for (const item of settings as SystemSetting[]) values[item.key] = item.value
    applyGatewayDefaults()
    snapshotInitialValues()
  } catch { error.value = t('settings.loadFailed') } finally { loading.value = false }
})
watch(() => [values.site_url, values['payment.paypal.environment']], () => applyGatewayDefaults())

async function save() {
  if (!auth.token) return
  saving.value = true; error.value = ''; saved.value = false
  let currentKey = ''
  try {
    const updates = buildSettingUpdates(definitions, values, initialValues, fieldLabel)
    for (const update of updates) {
      currentKey = update.key
      await generatedApi.updateSetting(update.key, update, auth.token)
      initialValues[update.key] = update.value
    }
    saved.value = true
  } catch (cause) {
    const detail = cause instanceof ApiError && cause.code ? t(errorMessageKey(cause.code)) : t('settings.saveFailed')
    error.value = currentKey ? `${t('settings.saveFailed')}: ${fieldLabel(currentKey)}（${detail}）` : detail
  } finally { saving.value = false }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center gap-2"><Settings2 class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('settings.title') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('settings.description') }}</p></div></div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Alert v-if="saved"><AlertDescription>{{ t('settings.saved') }}</AlertDescription></Alert>
    <div v-if="loading" class="text-sm text-muted-foreground">{{ t('resource.loading') }}</div>
    <form v-else class="grid gap-5" @submit.prevent="save">
      <Card v-for="group in groups" :key="group.key"><CardHeader><CardTitle>{{ t(`settings.groups.${group.key}`) }}</CardTitle><CardDescription>{{ t(`settings.groupDescriptions.${group.key}`) }}</CardDescription></CardHeader><CardContent>
        <div v-if="group.key === 'gateway'" class="grid gap-5 xl:grid-cols-3">
          <section v-for="section in gatewaySections" :key="section.key" class="rounded-lg border bg-muted/10 p-4">
            <div class="mb-4 flex items-start justify-between gap-3"><div><h3 class="font-medium">{{ t(`settings.gatewayProviders.${section.key}.title`) }}</h3><p class="mt-1 text-xs text-muted-foreground">{{ t(`settings.gatewayProviders.${section.key}.description`) }}</p></div><a v-if="gatewayRegistrationLinks[section.key]" :href="gatewayRegistrationLinks[section.key]" target="_blank" rel="noopener noreferrer" class="shrink-0 text-xs text-primary underline-offset-4 hover:underline">{{ t('settings.openProviderRegistration') }}</a></div>
            <FieldGroup class="grid gap-4">
              <Field v-for="key in section.fields" :key="key"><FieldLabel :for="`setting-${key}`">{{ fieldLabel(key) }}</FieldLabel>
                <input v-if="isBoolean(key)" :id="`setting-${key}`" type="checkbox" class="size-4" :checked="values[key] === 'true'" @change="setBoolean(key, $event)" />
                <select v-else-if="key === 'payment.paypal.environment'" :id="`setting-${key}`" class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm" :value="values[key] || 'sandbox'" @change="setSelectValue(key, $event)">
                  <option value="sandbox">{{ t('settings.paypalEnvironments.sandbox') }}</option>
                  <option value="production">{{ t('settings.paypalEnvironments.production') }}</option>
                </select>
                <Input v-else :id="`setting-${key}`" :model-value="fieldValue(key)" :type="isSecret(key) ? 'password' : definitions[key].type === 'integer' ? 'number' : 'text'" :placeholder="isSecret(key) && values[key] === '__configured__' ? t('settings.secretConfigured') : ''" @update:model-value="setValue(key, $event)" />
                <FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription>
              </Field>
            </FieldGroup>
            <div class="mt-5 grid gap-4 border-t pt-4">
              <p class="text-xs font-medium text-muted-foreground">{{ t('settings.gatewayURLsTitle') }}</p>
              <Field v-for="item in gatewayURLFields(section.key)" :key="item.key">
                <FieldLabel :for="`generated-${section.key}-${item.kind}`">{{ t(item.label) }}</FieldLabel>
                <Input :id="`generated-${section.key}-${item.kind}`" v-model="values[item.key]" class="font-mono text-xs" />
                <FieldDescription>{{ t(item.hint) }}</FieldDescription>
              </Field>
            </div>
          </section>
        </div>
        <FieldGroup v-else class="grid gap-4 md:grid-cols-2">
          <Field v-for="key in group.fields" :key="key" :class="definitions[key].multiline ? 'md:col-span-2' : ''"><FieldLabel :for="`setting-${key}`">{{ fieldLabel(key) }}</FieldLabel>
            <input v-if="isBoolean(key)" :id="`setting-${key}`" type="checkbox" class="size-4" :checked="values[key] === 'true'" @change="setBoolean(key, $event)" />
            <Textarea v-else-if="definitions[key].multiline" :id="`setting-${key}`" v-model="values[key]" class="min-h-24" />
            <Input v-else :id="`setting-${key}`" v-model="values[key]" :type="definitions[key].type === 'integer' ? 'number' : 'text'" />
            <FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription>
          </Field>
        </FieldGroup>
      </CardContent></Card>
      <div class="flex justify-end"><Button type="submit" :disabled="saving"><Save data-icon="inline-start" />{{ saving ? t('settings.saving') : t('settings.save') }}</Button></div>
    </form>
  </div>
</template>
