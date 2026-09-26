<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { Save, Settings2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { generatedApi, type SystemSetting } from '@/generated/api'
import { ApiError, errorMessageKey } from '@/lib/api'
import { buildSettingUpdates } from '@/modules/settings/save'
import { emailProviderOptions, paymentEnvironmentOptions, settingDefaults, settingDefinitions, settingGroups, smtpEncryptionOptions } from '@/modules/settings/settings-schema'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const initialValues = reactive<Record<string, string>>({})
const groups = settingGroups
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
const definitions = settingDefinitions()

function fieldLabel(key: string) { return t(`settings.fields.${key.replaceAll('.', '_')}`, key) }
function fieldHint(key: string) { return t(`settings.hints.${key.replaceAll('.', '_')}`, '') }
function isBoolean(key: string) { return definitions[key]?.type === 'boolean' }
function isSecret(key: string) { return definitions[key]?.type === 'secret' }
function fieldValue(key: string) { return values[key] === '__configured__' ? '' : values[key] || '' }
function setValue(key: string, value: string | number) { values[key] = String(value ?? '') }
function setBoolean(key: string, checked: boolean) { values[key] = checked ? 'true' : 'false' }
function setSelectValue(key: string, value: string) { values[key] = value }
function fieldVisible(key: string) { return definitions[key]?.visible?.(values) ?? true }
function selectOptions(key: string) {
  if (key === 'email.provider') return emailProviderOptions
  if (key === 'email.smtp.encryption') return smtpEncryptionOptions
  if (key === 'payment.paypal.environment') return paymentEnvironmentOptions
  if (key === 'payment.default_gateway') return [
    { value: 'paypal', labelKey: 'settings.gatewayOptions.paypal' },
    { value: 'xcash', labelKey: 'settings.gatewayOptions.xcash' },
    { value: 'nowpayments', labelKey: 'settings.gatewayOptions.nowpayments' },
    { value: 'fake', labelKey: 'settings.gatewayOptions.fake' },
  ]
  return []
}
const gatewayGeneralKeys = ['payment.default_gateway', 'payment.fake.enabled']
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

function applyEmailDefaults() {
  for (const [key, value] of Object.entries(settingDefaults)) if (!values[key]) values[key] = value
}

onMounted(async () => {
  if (!auth.token) return
  try {
    const settings = await generatedApi.settings(auth.token)
    for (const item of settings as SystemSetting[]) values[item.key] = item.value
    applyEmailDefaults()
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
      <template v-for="group in groups" :key="group.key">
      <Card v-if="group.fields.length">
        <CardHeader><CardTitle>{{ t(`settings.groups.${group.key}`) }}</CardTitle><CardDescription>{{ t(`settings.groupDescriptions.${group.key}`) }}</CardDescription></CardHeader>
        <CardContent>
          <FieldGroup v-if="group.key !== 'gateway'" class="grid min-w-0 gap-4 md:grid-cols-2">
            <Field v-for="definition in group.fields" v-show="fieldVisible(definition.key)" :key="definition.key" :class="definition.span === 'full' ? 'min-w-0 md:col-span-2' : 'min-w-0'">
              <template v-if="definition.type === 'boolean'">
                <div class="flex min-h-9 items-start gap-3 rounded-md border border-transparent py-1">
                  <Switch :id="`setting-${definition.key}`" :model-value="values[definition.key] === 'true'" @update:model-value="setBoolean(definition.key, Boolean($event))" />
                  <div class="min-w-0"><FieldLabel :for="`setting-${definition.key}`" class="cursor-pointer">{{ fieldLabel(definition.key) }}</FieldLabel><FieldDescription v-if="fieldHint(definition.key)">{{ fieldHint(definition.key) }}</FieldDescription></div>
                </div>
              </template>
              <template v-else>
                <FieldLabel :for="`setting-${definition.key}`">{{ fieldLabel(definition.key) }}</FieldLabel>
                <Select v-if="definition.type === 'select'" :model-value="values[definition.key] || undefined" @update:model-value="setSelectValue(definition.key, String($event))">
                  <SelectTrigger :id="`setting-${definition.key}`" class="w-full"><SelectValue :placeholder="fieldLabel(definition.key)" /></SelectTrigger>
                  <SelectContent><SelectItem v-for="option in selectOptions(definition.key)" :key="option.value" :value="option.value">{{ t(option.labelKey) }}</SelectItem></SelectContent>
                </Select>
                <Textarea v-else-if="definition.type === 'textarea'" :id="`setting-${definition.key}`" :model-value="values[definition.key] || ''" class="min-h-24 w-full" @update:model-value="setValue(definition.key, $event)" />
                <Input v-else :id="`setting-${definition.key}`" :model-value="fieldValue(definition.key)" class="w-full" :type="isSecret(definition.key) ? 'password' : definition.type === 'integer' ? 'number' : 'text'" :placeholder="isSecret(definition.key) && values[definition.key] === '__configured__' ? t('settings.secretConfigured') : ''" @update:model-value="setValue(definition.key, $event)" />
                <FieldDescription v-if="fieldHint(definition.key)">{{ fieldHint(definition.key) }}</FieldDescription>
              </template>
            </Field>
          </FieldGroup>

          <template v-else>
            <FieldGroup class="mb-6 grid min-w-0 gap-4 md:grid-cols-2">
              <Field v-for="key in gatewayGeneralKeys" :key="key" class="min-w-0">
                <template v-if="definitions[key].type === 'boolean'">
                  <div class="flex min-h-9 items-start gap-3 rounded-md border border-transparent py-1"><Switch :id="`setting-${key}`" :model-value="values[key] === 'true'" @update:model-value="setBoolean(key, Boolean($event))" /><div class="min-w-0"><FieldLabel :for="`setting-${key}`" class="cursor-pointer">{{ fieldLabel(key) }}</FieldLabel><FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription></div></div>
                </template>
                <template v-else>
                  <FieldLabel :for="`setting-${key}`">{{ fieldLabel(key) }}</FieldLabel>
                  <Select :model-value="values[key] || undefined" @update:model-value="setSelectValue(key, String($event))"><SelectTrigger :id="`setting-${key}`" class="w-full"><SelectValue :placeholder="fieldLabel(key)" /></SelectTrigger><SelectContent><SelectItem v-for="option in selectOptions(key)" :key="option.value" :value="option.value">{{ t(option.labelKey) }}</SelectItem></SelectContent></Select>
                  <FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription>
                </template>
              </Field>
            </FieldGroup>
            <div class="grid min-w-0 gap-5 xl:grid-cols-3">
              <section v-for="section in gatewaySections" :key="section.key" class="min-w-0 rounded-lg border bg-muted/10 p-4">
                <div class="mb-4 flex items-start justify-between gap-3"><div class="min-w-0"><h3 class="font-medium">{{ t(`settings.gatewayProviders.${section.key}.title`) }}</h3><p class="mt-1 text-xs text-muted-foreground">{{ t(`settings.gatewayProviders.${section.key}.description`) }}</p></div><a v-if="gatewayRegistrationLinks[section.key]" :href="gatewayRegistrationLinks[section.key]" target="_blank" rel="noopener noreferrer" class="shrink-0 text-xs text-primary underline-offset-4 hover:underline">{{ t('settings.openProviderRegistration') }}</a></div>
                <FieldGroup class="grid min-w-0 gap-4">
                  <Field v-for="key in section.fields" :key="key" class="min-w-0">
                    <template v-if="definitions[key].type === 'boolean'"><div class="flex min-h-9 items-start gap-3 rounded-md border border-transparent py-1"><Switch :id="`setting-${key}`" :model-value="values[key] === 'true'" @update:model-value="setBoolean(key, Boolean($event))" /><div class="min-w-0"><FieldLabel :for="`setting-${key}`" class="cursor-pointer">{{ fieldLabel(key) }}</FieldLabel><FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription></div></div></template>
                    <template v-else><FieldLabel :for="`setting-${key}`">{{ fieldLabel(key) }}</FieldLabel><Select v-if="definitions[key].type === 'select'" :model-value="values[key] || undefined" @update:model-value="setSelectValue(key, String($event))"><SelectTrigger :id="`setting-${key}`" class="w-full"><SelectValue :placeholder="fieldLabel(key)" /></SelectTrigger><SelectContent><SelectItem v-for="option in selectOptions(key)" :key="option.value" :value="option.value">{{ t(option.labelKey) }}</SelectItem></SelectContent></Select><Input v-else :id="`setting-${key}`" :model-value="fieldValue(key)" class="w-full" :type="isSecret(key) ? 'password' : 'text'" :placeholder="isSecret(key) && values[key] === '__configured__' ? t('settings.secretConfigured') : ''" @update:model-value="setValue(key, $event)" /><FieldDescription v-if="fieldHint(key)">{{ fieldHint(key) }}</FieldDescription></template>
                  </Field>
                </FieldGroup>
                <div class="mt-5 grid min-w-0 gap-4 border-t pt-4"><p class="text-xs font-medium text-muted-foreground">{{ t('settings.gatewayURLsTitle') }}</p><Field v-for="item in gatewayURLFields(section.key)" :key="item.key" class="min-w-0"><FieldLabel :for="`generated-${section.key}-${item.kind}`">{{ t(item.label) }}</FieldLabel><Input :id="`generated-${section.key}-${item.kind}`" v-model="values[item.key]" class="w-full font-mono text-xs" /><FieldDescription>{{ t(item.hint) }}</FieldDescription></Field></div>
              </section>
            </div>
          </template>
        </CardContent>
      </Card>
      </template>
      <div class="flex justify-end"><Button type="submit" :disabled="saving"><Save data-icon="inline-start" />{{ saving ? t('settings.saving') : t('settings.save') }}</Button></div>
    </form>
  </div>
</template>
