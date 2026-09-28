<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Check, ExternalLink, Link2, RefreshCw, X } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, errorMessageKey } from '@/lib/api'
import { friendApi, type FriendLink } from '@/modules/friend-links/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n(); const auth = useAuthStore(); const links = ref<FriendLink[]>([]); const loading = ref(true); const error = ref('')
function label(status?: string) { return t(`friendLinks.status.${status || 'pending'}`) }
async function load() { if (!auth.token) return; loading.value = true; try { links.value = await friendApi.adminList('', auth.token) } catch (value) { error.value = value instanceof ApiError ? t(errorMessageKey(value.code)) : t('friendLinks.loadFailed') } finally { loading.value = false } }
async function review(link: FriendLink, status: 'approved' | 'rejected' | 'hidden') { if (!auth.token) return; try { await friendApi.review(link.id, status, '', auth.token); await load() } catch (value) { error.value = value instanceof ApiError ? t(errorMessageKey(value.code)) : t('friendLinks.reviewFailed') } }
onMounted(load)
</script>

<template><div class="flex flex-col gap-6"><div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div class="flex items-center gap-3"><Link2 class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('friendLinks.adminTitle') }}</h1><p class="mt-1 text-sm text-muted-foreground">{{ t('friendLinks.adminDescription') }}</p></div></div><Button variant="outline" size="sm" :disabled="loading" @click="load"><RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}</Button></div><Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert><Card><CardHeader><CardTitle>{{ t('friendLinks.adminTitle') }}</CardTitle><CardDescription>{{ t('friendLinks.adminDescription') }}</CardDescription></CardHeader><CardContent><div v-if="loading" class="py-8 text-sm text-muted-foreground">{{ t('resource.loading') }}</div><div v-else-if="!links.length" class="py-8 text-sm text-muted-foreground">{{ t('friendLinks.adminEmpty') }}</div><div v-else class="divide-y rounded-md border"><div v-for="link in links" :key="link.id" class="flex flex-col gap-3 p-4 lg:flex-row lg:items-center lg:justify-between"><div class="min-w-0"><div class="flex items-center gap-2"><span class="font-medium">{{ link.site_name }}</span><span class="rounded-full bg-muted px-2 py-0.5 text-xs">{{ label(link.status) }}</span></div><a :href="link.url" target="_blank" rel="noopener noreferrer" class="mt-1 inline-flex items-center gap-1 truncate text-xs text-primary hover:underline">{{ link.url }}<ExternalLink class="size-3" /></a><p v-if="link.description" class="mt-1 text-sm text-muted-foreground">{{ link.description }}</p></div><div class="flex shrink-0 gap-2"><Button size="sm" :disabled="link.status === 'approved'" @click="review(link, 'approved')"><Check class="mr-1 size-4" />{{ t('friendLinks.approve') }}</Button><Button variant="outline" size="sm" :disabled="link.status === 'rejected'" @click="review(link, 'rejected')"><X class="mr-1 size-4" />{{ t('friendLinks.reject') }}</Button></div></div></div></CardContent></Card></div></template>
