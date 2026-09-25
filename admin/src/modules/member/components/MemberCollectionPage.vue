<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { Folder, ImagePlus, Library, Pencil, Plus, Trash2, X } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { apiFetch, apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

type CollectionKind = 'folders' | 'albums'
interface CollectionItem {
  id: number
  name: string
  parent_id?: number | null
  cover_media_id?: number | null
  visibility?: string
  media_count?: number
  created_at?: string
}

const props = defineProps<{ kind: CollectionKind }>()
const { t } = useI18n()
const auth = useAuthStore()
const items = ref<CollectionItem[]>([])
const name = ref('')
const parentID = ref('')
const visibility = ref('private')
const editingID = ref<number>()
const loading = ref(true)
const saving = ref(false)
const errorKey = ref('')
const endpoint = computed(() => `/api/v1/me/${props.kind}`)
const copy = computed(() => props.kind === 'folders' ? 'member.folders' : 'member.albums')
const isFolder = computed(() => props.kind === 'folders')

function resetForm() {
  name.value = ''
  parentID.value = ''
  visibility.value = 'private'
  editingID.value = undefined
}

async function load() {
  if (!auth.token) return
  loading.value = true
  errorKey.value = ''
  try {
    const response = await apiFetchEnvelope<CollectionItem[]>(endpoint.value, {}, auth.token)
    items.value = response.data
  } catch {
    errorKey.value = 'member.collections.errors.loadFailed'
  } finally {
    loading.value = false
  }
}

function edit(item: CollectionItem) {
  editingID.value = item.id
  name.value = item.name
  parentID.value = item.parent_id ? String(item.parent_id) : ''
  visibility.value = item.visibility || 'private'
}

async function save() {
  if (!auth.token || !name.value.trim() || saving.value) return
  saving.value = true
  errorKey.value = ''
  try {
    const payload: Record<string, unknown> = { name: name.value.trim() }
    if (isFolder.value) payload.parent_id = parentID.value ? Number(parentID.value) : null
    else payload.visibility = visibility.value
    const path = editingID.value ? `${endpoint.value}/${editingID.value}` : endpoint.value
    await apiFetch(path, { method: editingID.value ? 'PATCH' : 'POST', body: JSON.stringify(payload) }, auth.token)
    resetForm()
    await load()
  } catch {
    errorKey.value = 'member.collections.errors.saveFailed'
  } finally {
    saving.value = false
  }
}

async function remove(item: CollectionItem) {
  if (!auth.token) return
  errorKey.value = ''
  try {
    await apiFetch(`${endpoint.value}/${item.id}`, { method: 'DELETE' }, auth.token)
    if (editingID.value === item.id) resetForm()
    await load()
  } catch {
    errorKey.value = 'member.collections.errors.deleteFailed'
  }
}

onMounted(() => void load())
</script>

<template>
  <div class="flex flex-col gap-7">
    <header>
      <p class="mb-2 flex items-center gap-2 text-sm text-muted-foreground"><Folder v-if="isFolder" /><Library v-else />{{ t(`${copy}.workspace`) }}</p>
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t(`${copy}.title`) }}</h1>
      <p class="mt-3 max-w-2xl text-muted-foreground">{{ t(`${copy}.description`) }}</p>
    </header>

    <Alert v-if="errorKey" variant="destructive">
      <AlertTitle>{{ t('member.collections.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ t(errorKey) }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2"><Plus />{{ editingID ? t('member.collections.edit') : t('member.collections.create') }}</CardTitle>
        <CardDescription>{{ t('member.collections.ownerHint') }}</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-4 sm:flex-row sm:items-end">
        <div class="flex-1">
          <label class="mb-2 block text-sm font-medium" :for="`${props.kind}-name`">{{ t('member.collections.name') }}</label>
          <Input :id="`${props.kind}-name`" v-model="name" maxlength="120" :placeholder="t('member.collections.namePlaceholder')" @keydown.enter="save" />
        </div>
        <div v-if="isFolder" class="sm:w-56">
          <label class="mb-2 block text-sm font-medium" :for="`${props.kind}-parent`">{{ t('member.folders.parent') }}</label>
          <select :id="`${props.kind}-parent`" v-model="parentID" class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
            <option value="">{{ t('member.folders.root') }}</option>
            <option v-for="item in items.filter((entry) => entry.id !== editingID)" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select>
        </div>
        <div v-else class="sm:w-40">
          <label class="mb-2 block text-sm font-medium" :for="`${props.kind}-visibility`">{{ t('member.albums.visibility') }}</label>
          <select :id="`${props.kind}-visibility`" v-model="visibility" class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
            <option value="private">{{ t('member.albums.private') }}</option>
            <option value="unlisted">{{ t('member.albums.unlisted') }}</option>
            <option value="public">{{ t('member.albums.public') }}</option>
          </select>
        </div>
        <div class="flex gap-2">
          <Button :disabled="saving || !name.trim()" @click="save"><ImagePlus data-icon="inline-start" />{{ saving ? t('member.collections.saving') : t('member.collections.save') }}</Button>
          <Button v-if="editingID" variant="ghost" @click="resetForm"><X data-icon="inline-start" />{{ t('member.collections.cancel') }}</Button>
        </div>
      </CardContent>
    </Card>

    <section>
      <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card v-for="index in 3" :key="index" class="h-28 animate-pulse bg-muted/40" />
      </div>
      <Card v-else-if="!items.length">
        <CardContent class="flex min-h-40 flex-col items-center justify-center gap-2 text-center text-muted-foreground">
          <Folder v-if="isFolder" class="size-8" /><Library v-else class="size-8" />
          <p>{{ t(`${copy}.empty`) }}</p>
        </CardContent>
      </Card>
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card v-for="item in items" :key="item.id">
          <CardHeader class="pb-3"><CardTitle class="truncate text-base" :title="item.name">{{ item.name }}</CardTitle><CardDescription v-if="isFolder">{{ item.parent_id ? t('member.folders.nested') : t('member.folders.root') }}</CardDescription><CardDescription v-else>{{ t(`member.albums.${item.visibility || 'private'}`) }} · {{ t('member.albums.mediaCount', { count: item.media_count || 0 }) }}</CardDescription></CardHeader>
          <CardContent class="flex justify-end gap-2">
            <Button v-if="!isFolder" variant="outline" size="sm" as-child><RouterLink :to="{ name: 'member-album', params: { id: item.id } }">{{ t('member.albums.viewMedia') }}</RouterLink></Button>
            <Button variant="outline" size="sm" @click="edit(item)"><Pencil data-icon="inline-start" />{{ t('member.collections.edit') }}</Button>
            <Button variant="ghost" size="sm" @click="remove(item)"><Trash2 data-icon="inline-start" />{{ t('member.collections.delete') }}</Button>
          </CardContent>
        </Card>
      </div>
    </section>
  </div>
</template>
