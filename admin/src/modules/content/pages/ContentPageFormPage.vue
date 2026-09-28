<script setup lang="ts">
import type { JSONContent } from '@tiptap/core'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Save } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ApiError, errorMessageKey } from '@/lib/api'
import { contentApi } from '@/modules/content/api'
import RichTextEditor from '@/modules/content/components/RichTextEditor.vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const emptyDocument: JSONContent = { type: 'doc', content: [{ type: 'paragraph' }] }
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const editing = computed(() => Boolean(route.params.id))
const loading = ref(editing.value)
const saving = ref(false)
const error = ref('')
const form = reactive({ slug: '', title: '', content_json: emptyDocument, excerpt: '', seo_title: '', seo_description: '' })

function localError(value: unknown) { return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('content.saveFailed') }

onMounted(async () => {
  if (!editing.value || !auth.token) { loading.value = false; return }
  try {
    const page = await contentApi.show(String(route.params.id), auth.token)
    Object.assign(form, page, { content_json: page.content_json || emptyDocument })
  } catch (value) { error.value = localError(value) } finally { loading.value = false }
})

async function submit() {
  if (!auth.token) return
  saving.value = true
  error.value = ''
  try {
    if (editing.value) await contentApi.update(String(route.params.id), form, auth.token)
    else await contentApi.create(form, auth.token)
    await router.push('/admin/content-pages')
  } catch (value) { error.value = localError(value) } finally { saving.value = false }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center gap-3"><Button variant="ghost" size="icon" :aria-label="t('resource.back')" @click="router.push('/admin/content-pages')"><ArrowLeft /></Button><div><h1 class="text-2xl font-semibold tracking-tight">{{ editing ? t('content.editPage') : t('content.newPage') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('content.editorDescription') }}</p></div></div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card v-if="loading"><CardContent class="py-8">{{ t('resource.loading') }}</CardContent></Card>
    <form v-else class="grid gap-5" @submit.prevent="submit">
      <Card><CardHeader><CardTitle>{{ t('content.pageDetails') }}</CardTitle><CardDescription>{{ t('content.pageDetailsDescription') }}</CardDescription></CardHeader><CardContent><FieldGroup class="grid gap-4 md:grid-cols-2"><Field><FieldLabel for="page-slug">{{ t('content.slug') }}</FieldLabel><Input id="page-slug" v-model="form.slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" required /><p class="text-xs text-muted-foreground">{{ t('content.slugHint') }}</p></Field><Field><FieldLabel for="page-title">{{ t('content.title') }}</FieldLabel><Input id="page-title" v-model="form.title" required /></Field><Field class="md:col-span-2"><FieldLabel for="page-excerpt">{{ t('content.excerpt') }}</FieldLabel><Textarea id="page-excerpt" v-model="form.excerpt" rows="2" /></Field></FieldGroup></CardContent></Card>
      <Card><CardHeader><CardTitle>{{ t('content.body') }}</CardTitle><CardDescription>{{ t('content.richTextOnly') }}</CardDescription></CardHeader><CardContent><RichTextEditor v-model="form.content_json" /></CardContent></Card>
      <Card><CardHeader><CardTitle>{{ t('content.seoTitle') }}</CardTitle><CardDescription>{{ t('content.seoDescription') }}</CardDescription></CardHeader><CardContent><FieldGroup class="grid gap-4 md:grid-cols-2"><Field><FieldLabel for="seo-title">{{ t('content.seoTitle') }}</FieldLabel><Input id="seo-title" v-model="form.seo_title" /></Field><Field><FieldLabel for="seo-description">{{ t('content.seoDescription') }}</FieldLabel><Textarea id="seo-description" v-model="form.seo_description" rows="2" /></Field></FieldGroup></CardContent></Card>
      <div class="flex gap-2"><Button type="submit" :disabled="saving"><Save data-icon="inline-start" />{{ saving ? t('resource.saving') : t('resource.save') }}</Button><Button type="button" variant="outline" @click="router.push('/admin/content-pages')">{{ t('resource.cancel') }}</Button></div>
    </form>
  </div>
</template>
