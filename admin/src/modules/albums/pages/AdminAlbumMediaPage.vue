<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ArrowLeft, Library, Trash2 } from '@lucide/vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { apiFetch, apiFetchEnvelope } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface AdminAlbumMediaItem { id: number; original_name: string; user_id: number; status: string; visibility: string; sort_order: number }
const route = useRoute(); const router = useRouter(); const { t } = useI18n(); const auth = useAuthStore()
const albumID = Number(route.params.id)
interface AlbumMediaMeta { album_id: number; page: number; per_page: number; total: number; last_page: number }
const items = ref<AdminAlbumMediaItem[]>([]); const mediaIDs = ref(''); const loading = ref(true); const error = ref(''); const success = ref(false); const pageSize = ref('20'); const meta = ref<AlbumMediaMeta>({ album_id: albumID, page: 1, per_page: 20, total: 0, last_page: 0 })
async function load(page = meta.value.page) { if (!auth.token || !albumID) { loading.value = false; error.value = 'albums.loadFailed'; return }; loading.value = true; error.value = ''; try { const response = await apiFetchEnvelope<AdminAlbumMediaItem[]>(`/api/v1/admin/albums/${albumID}/media?page=${page}&per_page=${pageSize.value}`, {}, auth.token); items.value = response.data; if (response.meta) meta.value = response.meta as unknown as AlbumMediaMeta } catch { error.value = 'albums.loadFailed' } finally { loading.value = false } }
async function changePageSize(value: unknown) { pageSize.value = String(value); await load(1) }
function parseIDs() { return [...new Set(mediaIDs.value.split(',').map((value) => Number(value.trim())).filter((value) => Number.isInteger(value) && value > 0))] }
async function add() { const ids = parseIDs(); if (!auth.token || !ids.length) return; error.value = ''; success.value = false; try { await apiFetch(`/api/v1/admin/albums/${albumID}/media`, { method: 'POST', body: JSON.stringify({ media_ids: ids }) }, auth.token); mediaIDs.value = ''; success.value = true; await load() } catch { error.value = 'albums.saveFailed' } }
async function remove(ids: number[]) { if (!auth.token || !ids.length) return; error.value = ''; success.value = false; try { await apiFetch(`/api/v1/admin/albums/${albumID}/media`, { method: 'DELETE', body: JSON.stringify({ media_ids: ids }) }, auth.token); success.value = true; await load() } catch { error.value = 'albums.saveFailed' } }
onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <header class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><Button variant="ghost" size="sm" @click="router.push(`/admin/albums/${albumID}`)"><ArrowLeft data-icon="inline-start" />{{ t('resource.back') }}</Button><div class="mt-2 flex items-center gap-2"><Library class="size-5 text-primary" /><div><h1 class="text-2xl font-semibold tracking-tight">{{ t('albums.title') }}</h1><p class="text-sm text-muted-foreground">#{{ albumID }} · {{ t('albums.description') }}</p></div></div></div></header>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ t(error) }}</AlertDescription></Alert><Alert v-if="success"><AlertTitle>{{ t('albums.success') }}</AlertTitle></Alert>
    <Card><CardHeader><CardTitle>{{ t('albums.add') }}</CardTitle><CardDescription>{{ t('albums.mediaIdHint') }}</CardDescription></CardHeader><CardContent class="flex flex-col gap-3 sm:flex-row sm:items-end"><label class="flex-1 text-sm font-medium">{{ t('albums.mediaIdList') }}<Input v-model="mediaIDs" class="mt-2" placeholder="12, 13, 14" /></label><Button :disabled="!parseIDs().length" @click="add">{{ t('albums.add') }}</Button></CardContent></Card>
    <Card><CardHeader><CardTitle>{{ t('albums.title') }}</CardTitle></CardHeader><CardContent><div v-if="loading" class="text-sm text-muted-foreground">{{ t('core.loading') }}</div><Empty v-else-if="!items.length"><EmptyHeader><EmptyTitle>{{ t('albums.empty') }}</EmptyTitle><EmptyDescription>{{ t('albums.description') }}</EmptyDescription></EmptyHeader></Empty><div v-else class="overflow-x-auto"><table class="w-full min-w-[700px] text-sm"><thead><tr class="border-b text-left text-muted-foreground"><th class="px-3 py-2">ID</th><th class="px-3 py-2">{{ t('albums.mediaIdList') }}</th><th class="px-3 py-2">{{ t('albums.owner') }}</th><th class="px-3 py-2">{{ t('albums.status') }}</th><th class="px-3 py-2">{{ t('albums.visibility') }}</th><th class="px-3 py-2"></th></tr></thead><tbody><tr v-for="item in items" :key="item.id" class="border-b last:border-0"><td class="px-3 py-2">{{ item.id }}</td><td class="max-w-64 truncate px-3 py-2" :title="item.original_name">{{ item.original_name }}</td><td class="px-3 py-2">{{ item.user_id }}</td><td class="px-3 py-2">{{ item.status }}</td><td class="px-3 py-2">{{ item.visibility }}</td><td class="px-3 py-2 text-right"><Button variant="ghost" size="sm" @click="remove([item.id])"><Trash2 data-icon="inline-start" />{{ t('albums.remove') }}</Button></td></tr></tbody></table></div><div v-if="!loading && meta.total > 0" class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div></CardContent></Card>
  </div>
</template>
