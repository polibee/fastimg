<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ExternalLink, Link2, Send } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ApiError, errorMessageKey } from '@/lib/api'
import { friendApi, type FriendLink } from '@/modules/friend-links/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const links = ref<FriendLink[]>([])
const loading = ref(true)
const sending = ref(false)
const error = ref('')
const sent = ref(false)
const form = reactive({ site_name: '', url: '', logo_url: '', description: '', contact_email: '' })
function errorText(value: unknown) { return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('friendLinks.loadFailed') }
async function load() { try { links.value = await friendApi.list() } catch (value) { error.value = errorText(value) } finally { loading.value = false } }
async function submit() { sending.value = true; error.value = ''; sent.value = false; try { await friendApi.submit(form); sent.value = true; Object.assign(form, { site_name: '', url: '', logo_url: '', description: '', contact_email: '' }) } catch (value) { error.value = errorText(value) } finally { sending.value = false } }
onMounted(load)
</script>

<template>
  <section class="space-y-8"><header class="max-w-2xl space-y-3"><p class="text-sm font-medium text-primary">{{ t('friendLinks.eyebrow') }}</p><h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('friendLinks.title') }}</h1><p class="text-muted-foreground">{{ t('friendLinks.description') }}</p></header>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert><Alert v-if="sent"><AlertTitle>{{ t('friendLinks.pendingReview') }}</AlertTitle><AlertDescription>{{ t('friendLinks.submitted') }}</AlertDescription></Alert>
    <div v-if="loading" class="text-sm text-muted-foreground">{{ t('resource.loading') }}</div><div v-else-if="!links.length" class="text-sm text-muted-foreground">{{ t('friendLinks.empty') }}</div><div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"><a v-for="link in links" :key="link.id" :href="link.url" target="_blank" rel="noopener noreferrer" class="group rounded-xl border bg-card p-5 transition-colors hover:border-primary/50 hover:bg-muted/30"><div class="flex items-center gap-3"><img v-if="link.logo_url" :src="link.logo_url" :alt="link.site_name" class="size-10 rounded-lg object-cover" /><span v-else class="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary"><Link2 class="size-5" /></span><div class="min-w-0"><h2 class="truncate font-medium group-hover:text-primary">{{ link.site_name }}</h2><p class="truncate text-xs text-muted-foreground">{{ link.url }}</p></div><ExternalLink class="ml-auto size-4 shrink-0 text-muted-foreground" /></div><p v-if="link.description" class="mt-4 line-clamp-2 text-sm text-muted-foreground">{{ link.description }}</p></a></div>
    <Card><CardHeader><CardTitle>{{ t('friendLinks.submitTitle') }}</CardTitle><CardDescription>{{ t('friendLinks.submitDescription') }}</CardDescription></CardHeader><CardContent><form class="grid gap-4 md:grid-cols-2" @submit.prevent="submit"><Field><FieldLabel for="friend-site-name">{{ t('friendLinks.siteName') }}</FieldLabel><Input id="friend-site-name" v-model="form.site_name" required /></Field><Field><FieldLabel for="friend-site-url">{{ t('friendLinks.url') }}</FieldLabel><Input id="friend-site-url" v-model="form.url" type="url" required /></Field><Field><FieldLabel for="friend-logo-url">{{ t('friendLinks.logoUrl') }}</FieldLabel><Input id="friend-logo-url" v-model="form.logo_url" type="url" /></Field><Field><FieldLabel for="friend-contact">{{ t('friendLinks.contactEmail') }}</FieldLabel><Input id="friend-contact" v-model="form.contact_email" type="email" /></Field><Field class="md:col-span-2"><FieldLabel for="friend-description">{{ t('friendLinks.siteDescription') }}</FieldLabel><Textarea id="friend-description" v-model="form.description" rows="3" /></Field><div><Button type="submit" :disabled="sending"><Send data-icon="inline-start" />{{ sending ? t('friendLinks.submitting') : t('friendLinks.submit') }}</Button></div></form></CardContent></Card>
  </section>
</template>
