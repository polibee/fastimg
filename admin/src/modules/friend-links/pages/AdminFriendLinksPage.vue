<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Check, ExternalLink, EyeOff, Link2, RefreshCw, X } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { ApiError, errorMessageKey } from '@/lib/api'
import { friendApi, type FriendLink } from '@/modules/friend-links/api'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const links = ref<FriendLink[]>([])
const loading = ref(true)
const error = ref('')
const statusFilter = ref('')
const reviewNotes = reactive<Record<number, string>>({})
const canModerate = computed(() => auth.can('admin.friend_links.moderate'))
const meta = ref({ page: 1, per_page: 20, total: 0 })
const pageSize = ref('20')

function label(status?: string) {
  return t(`friendLinks.status.${status || 'pending'}`)
}

function statusClass(status?: string) {
  if (status === 'approved') return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200'
  if (status === 'rejected') return 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-200'
  if (status === 'hidden') return 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-200'
  return 'bg-muted text-muted-foreground'
}

async function load(page = 1) {
  if (!auth.token) return
  loading.value = true
  error.value = ''
  try {
    const response = await friendApi.adminList(statusFilter.value, auth.token, page, Number(pageSize.value))
    links.value = response.data
    meta.value = { page: Number(response.meta?.page || page), per_page: Number(response.meta?.per_page || pageSize.value), total: Number(response.meta?.total || 0) }
  } catch (value) {
    error.value = value instanceof ApiError ? t(errorMessageKey(value.code)) : t('friendLinks.loadFailed')
  } finally {
    loading.value = false
  }
}

function changePageSize(value: unknown) {
  pageSize.value = String(value)
  void load(1)
}

async function review(link: FriendLink, status: 'approved' | 'rejected' | 'hidden') {
  if (!auth.token || !canModerate.value) return
  try {
    await friendApi.review(link.id, status, reviewNotes[link.id] || '', auth.token)
    delete reviewNotes[link.id]
    await load(meta.value.page)
  } catch (value) {
    error.value = value instanceof ApiError ? t(errorMessageKey(value.code)) : t('friendLinks.reviewFailed')
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-3">
        <Link2 class="size-5 text-primary" />
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">{{ t('friendLinks.adminTitle') }}</h1>
          <p class="mt-1 text-sm text-muted-foreground">{{ t('friendLinks.adminDescription') }}</p>
        </div>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" as-child>
          <RouterLink to="/admin/settings#friend-links-copy">{{ t('friendLinks.editPublicCopy') }}</RouterLink>
        </Button>
        <Button variant="outline" size="sm" :disabled="loading" @click="load">
          <RefreshCw class="mr-2 size-4" />{{ t('resource.refresh') }}
        </Button>
      </div>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertTitle>{{ t('states.errorTitle') }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Alert v-if="!canModerate">
      <AlertTitle>{{ t('friendLinks.viewOnlyTitle') }}</AlertTitle>
      <AlertDescription>{{ t('friendLinks.viewOnlyDescription') }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <CardTitle>{{ t('friendLinks.adminTitle') }}</CardTitle>
          <CardDescription>{{ t('friendLinks.adminDescription') }}</CardDescription>
        </div>
        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          <span>{{ t('friendLinks.statusFilter') }}</span>
          <select v-model="statusFilter" class="h-9 rounded-md border bg-background px-3 text-sm text-foreground" @change="load(1)">
            <option value="">{{ t('friendLinks.allStatuses') }}</option>
            <option value="pending">{{ t('friendLinks.status.pending') }}</option>
            <option value="approved">{{ t('friendLinks.status.approved') }}</option>
            <option value="rejected">{{ t('friendLinks.status.rejected') }}</option>
            <option value="hidden">{{ t('friendLinks.status.hidden') }}</option>
          </select>
        </label>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="py-8 text-sm text-muted-foreground">{{ t('resource.loading') }}</div>
        <div v-else-if="!links.length" class="py-8 text-sm text-muted-foreground">{{ t('friendLinks.adminEmpty') }}</div>
        <div v-else class="divide-y rounded-md border">
          <div v-for="link in links" :key="link.id" class="flex flex-col gap-4 p-4 xl:flex-row xl:items-start xl:justify-between">
            <div class="flex min-w-0 gap-3">
              <img v-if="link.logo_url" :src="link.logo_url" :alt="link.site_name" class="size-12 shrink-0 rounded-md border object-cover" />
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium">{{ link.site_name }}</span>
                  <span class="rounded-full px-2 py-0.5 text-xs" :class="statusClass(link.status)">{{ label(link.status) }}</span>
                </div>
                <a :href="link.url" target="_blank" rel="noopener noreferrer" class="mt-1 inline-flex max-w-full items-center gap-1 truncate text-xs text-primary hover:underline">
                  {{ link.url }}<ExternalLink class="size-3 shrink-0" />
                </a>
                <p v-if="link.contact_email" class="mt-1 text-xs text-muted-foreground">{{ t('friendLinks.contactEmail') }}{{ t('friendLinks.valueSeparator') }}{{ link.contact_email }}</p>
                <p v-if="link.description" class="mt-2 text-sm text-muted-foreground">{{ link.description }}</p>
                <p v-if="link.review_note" class="mt-2 rounded-md bg-muted/60 px-3 py-2 text-xs text-muted-foreground"><span class="font-medium text-foreground">{{ t('friendLinks.reviewNote') }}{{ t('friendLinks.valueSeparator') }}</span>{{ link.review_note }}</p>
              </div>
            </div>

            <div v-if="canModerate" class="flex w-full shrink-0 flex-col gap-2 xl:w-72">
              <Textarea v-model="reviewNotes[link.id]" rows="2" :placeholder="t('friendLinks.reviewNotePlaceholder')" :aria-label="t('friendLinks.reviewNote')" />
              <div class="flex flex-wrap gap-2 xl:justify-end">
                <Button size="sm" :disabled="link.status === 'approved'" @click="review(link, 'approved')"><Check class="mr-1 size-4" />{{ t('friendLinks.approve') }}</Button>
                <Button variant="outline" size="sm" :disabled="link.status === 'rejected'" @click="review(link, 'rejected')"><X class="mr-1 size-4" />{{ t('friendLinks.reject') }}</Button>
                <Button variant="outline" size="sm" :disabled="link.status === 'hidden'" @click="review(link, 'hidden')"><EyeOff class="mr-1 size-4" />{{ t('friendLinks.hide') }}</Button>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!loading && meta.total > 0" class="mt-4 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"><p class="text-sm text-muted-foreground">{{ t('resource.page', { page: meta.page }) }}</p><Select :model-value="pageSize" @update:model-value="changePageSize"><SelectTrigger class="w-24"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="10">{{ t('resource.perPage', { count: 10 }) }}</SelectItem><SelectItem value="20">{{ t('resource.perPage', { count: 20 }) }}</SelectItem><SelectItem value="50">{{ t('resource.perPage', { count: 50 }) }}</SelectItem></SelectContent></Select><Pagination v-model:page="meta.page" :items-per-page="meta.per_page" :total="meta.total" @update:page="load"><PaginationContent v-slot="{ items }"><PaginationPrevious /><template v-for="(item, index) in items" :key="index"><PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === meta.page">{{ item.value }}</PaginationItem></template><PaginationNext /></PaginationContent></Pagination></div>
      </CardContent>
    </Card>
  </div>
</template>
