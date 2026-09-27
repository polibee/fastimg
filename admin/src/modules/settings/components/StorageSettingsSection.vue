<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ExternalLink, LoaderCircle, Save } from '@lucide/vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { ApiError, errorMessageKey } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { storageConnections, testStorageConnection, updateStorageConnection, type StorageConnection, type StorageOverview } from '../storage-api'

type Draft = {
  id: number
  name: string
  enabled: boolean
  is_primary: boolean
  public_base_url: string
  path_prefix: string
  default_visibility: string
  signed_url_ttl_seconds: number
  config: Record<string, string>
}

const { t } = useI18n()
const auth = useAuthStore()
const overview = ref<StorageOverview>()
const drafts = reactive<Record<string, Draft>>({})
const loading = ref(true)
const saving = ref('')
const testing = ref('')
const message = ref('')
const error = ref('')
const codes = ['local', 'cloudflare_r2', 'aliyun_oss', 'tencent_cos']

const orderedConnections = computed(() => codes.map((code) => overview.value?.connections.find((item) => item.provider_code === code)).filter(Boolean) as StorageConnection[])

function providerTitle(code: string) { return t(`storage.providers.${code}.title`, overview.value?.providers[code]?.label || code) }
function providerDescription(code: string) { return t(`storage.providers.${code}.description`, overview.value?.providers[code]?.description || '') }
function fieldLabel(name: string, fallback: string) { return t(`storage.fields.${name}`, fallback) }
function statusLabel(status: string) { return t(`storage.${status}`, status) }
function statusClass(status: string) {
  if (status === 'healthy') return 'text-emerald-600'
  if (status === 'degraded') return 'text-amber-600'
  if (status === 'error' || status === 'incomplete') return 'text-destructive'
  return 'text-muted-foreground'
}

function createDraft(connection: StorageConnection): Draft {
  return {
    id: connection.id,
    name: connection.name,
    enabled: connection.enabled,
    is_primary: connection.is_primary,
    public_base_url: connection.public_base_url || connection.config.public_base_url || '',
    path_prefix: connection.path_prefix || '',
    default_visibility: connection.default_visibility || 'public',
    signed_url_ttl_seconds: connection.signed_url_ttl_seconds || 3600,
    config: { ...connection.config },
  }
}

function setDraftValue(code: string, key: string, value: string | number) {
  const draft = drafts[code]
  if (draft) draft.config[key] = String(value ?? '')
}

function draft(code: string): Draft { return drafts[code] as Draft }

async function load() {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    overview.value = await storageConnections(auth.token)
    for (const connection of overview.value.connections) drafts[connection.provider_code] = createDraft(connection)
  } catch (cause) {
    error.value = cause instanceof ApiError && cause.code ? t(errorMessageKey(cause.code)) : t('storage.loadFailed')
  } finally { loading.value = false }
}

async function save(code: string) {
  if (!auth.token || !drafts[code]) return
  saving.value = code
  message.value = ''
  error.value = ''
  try {
    const updated = await updateStorageConnection(code, drafts[code], auth.token)
    drafts[code] = createDraft(updated)
    const index = overview.value?.connections.findIndex((item) => item.provider_code === code) ?? -1
    if (overview.value && index >= 0) overview.value.connections[index] = updated
    message.value = t('storage.saved')
  } catch (cause) {
    error.value = cause instanceof ApiError && cause.code ? t(errorMessageKey(cause.code)) : t('storage.saveFailed')
  } finally { saving.value = '' }
}

async function test(code: string) {
  if (!auth.token) return
  testing.value = code
  message.value = ''
  error.value = ''
  try {
    const updated = await testStorageConnection(code, auth.token)
    drafts[code] = createDraft(updated)
    const index = overview.value?.connections.findIndex((item) => item.provider_code === code) ?? -1
    if (overview.value && index >= 0) overview.value.connections[index] = updated
    message.value = updated.status === 'degraded' ? t('storage.adapterPending') : t('storage.tested')
  } catch (cause) {
    error.value = cause instanceof ApiError && cause.code ? t(errorMessageKey(cause.code)) : t('storage.testFailed')
  } finally { testing.value = '' }
}

onMounted(load)
</script>

<template>
  <Card>
    <CardHeader><CardTitle>{{ t('storage.settingsTitle') }}</CardTitle><CardDescription>{{ t('storage.settingsDescription') }}</CardDescription></CardHeader>
    <CardContent class="space-y-5">
      <Alert v-if="message"><AlertDescription>{{ message }}</AlertDescription></Alert>
      <Alert v-if="error" variant="destructive"><AlertDescription>{{ error }}</AlertDescription></Alert>
      <div v-if="loading" class="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle class="size-4 animate-spin" />{{ t('resource.loading') }}</div>
      <div v-else class="grid gap-5 xl:grid-cols-2">
        <section v-for="connection in orderedConnections" :key="connection.provider_code" class="min-w-0 rounded-lg border p-4">
          <div class="mb-4 flex items-start justify-between gap-3">
            <div class="min-w-0"><h3 class="font-medium">{{ providerTitle(connection.provider_code) }}</h3><p class="mt-1 text-xs text-muted-foreground">{{ providerDescription(connection.provider_code) }}</p></div>
            <a v-if="overview?.providers[connection.provider_code]?.registration_url" :href="overview.providers[connection.provider_code].registration_url" target="_blank" rel="noopener noreferrer" class="shrink-0 text-xs text-primary underline-offset-4 hover:underline"><ExternalLink class="mr-1 inline size-3" />{{ t('storage.registration') }}</a>
          </div>
          <template v-if="draft(connection.provider_code)">
            <div class="mb-4 flex flex-wrap items-center gap-4 text-sm">
              <label class="flex items-center gap-2"><Checkbox :model-value="draft(connection.provider_code).enabled" @update:model-value="draft(connection.provider_code).enabled = Boolean($event)" />{{ t('storage.enabled') }}</label>
              <label class="flex items-center gap-2"><Checkbox :model-value="draft(connection.provider_code).is_primary" @update:model-value="draft(connection.provider_code).is_primary = Boolean($event)" />{{ t('storage.primary') }}</label>
              <span :class="['font-medium', statusClass(connection.status)]">{{ statusLabel(connection.status) }}</span>
            </div>
            <FieldGroup class="grid min-w-0 gap-4 md:grid-cols-2">
              <Field v-for="field in overview?.providers[connection.provider_code]?.fields || []" :key="field.name" class="min-w-0">
                <FieldLabel :for="`storage-${connection.provider_code}-${field.name}`">{{ fieldLabel(field.name, field.label) }}</FieldLabel>
                <Input :id="`storage-${connection.provider_code}-${field.name}`" :model-value="draft(connection.provider_code).config[field.name] === '__configured__' ? '' : draft(connection.provider_code).config[field.name] || ''" :type="field.secret ? 'password' : 'text'" :placeholder="draft(connection.provider_code).config[field.name] === '__configured__' ? t('settings.secretConfigured') : field.placeholder" @update:model-value="setDraftValue(connection.provider_code, field.name, String($event))" />
              </Field>
              <Field><FieldLabel :for="`storage-${connection.provider_code}-public-base-url`">{{ t('storage.publicBaseURL') }}</FieldLabel><Input :id="`storage-${connection.provider_code}-public-base-url`" v-model="drafts[connection.provider_code].public_base_url" placeholder="https://img.example.com" /></Field>
              <Field><FieldLabel :for="`storage-${connection.provider_code}-path-prefix`">{{ t('storage.pathPrefix') }}</FieldLabel><Input :id="`storage-${connection.provider_code}-path-prefix`" v-model="drafts[connection.provider_code].path_prefix" placeholder="media" /></Field>
              <Field><FieldLabel :for="`storage-${connection.provider_code}-signed-url-ttl`">{{ t('storage.signedURLTTL') }}</FieldLabel><Input :id="`storage-${connection.provider_code}-signed-url-ttl`" v-model.number="drafts[connection.provider_code].signed_url_ttl_seconds" type="number" min="60" max="86400" /></Field>
            </FieldGroup>
            <div class="mt-5 flex flex-wrap justify-end gap-2 border-t pt-4"><Button type="button" variant="outline" size="sm" :disabled="Boolean(testing || saving)" @click="test(connection.provider_code)"><LoaderCircle v-if="testing === connection.provider_code" class="mr-2 size-4 animate-spin" /><span v-else>{{ t('storage.test') }}</span></Button><Button type="button" size="sm" :disabled="Boolean(testing || saving)" @click="save(connection.provider_code)"><LoaderCircle v-if="saving === connection.provider_code" class="mr-2 size-4 animate-spin" /><Save v-else class="mr-2 size-4" />{{ t('storage.save') }}</Button></div>
          </template>
        </section>
      </div>
    </CardContent>
  </Card>
</template>
