<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Languages, LogOut } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useAuthStore } from '@/stores/auth'
import { hasAdminAccess } from '@/lib/admin-access'
import MemberAdSlot from '@/modules/advertising/components/MemberAdSlot.vue'
import MemberFooterNavigation from '@/modules/member/components/MemberFooterNavigation.vue'

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()

function toggleLocale() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  localStorage.setItem('locale', locale.value)
}

async function logout() {
  try {
    await auth.logout()
  } finally {
    await router.replace({ path: '/' })
  }
}
</script>

<template>
  <div class="min-h-svh bg-background text-foreground">
    <header class="border-b bg-card">
      <div class="mx-auto flex min-h-16 max-w-6xl flex-wrap items-center gap-3 px-4 md:flex-nowrap md:gap-5 md:px-6">
        <RouterLink to="/" :aria-label="t('member.brand')" :title="t('member.brand')" class="flex shrink-0 items-center gap-3 rounded-sm font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <span class="flex size-9 items-center justify-center rounded-md bg-primary text-sm text-primary-foreground">F</span>
          <span class="hidden sm:inline">{{ t('member.brand') }}</span>
        </RouterLink>
        <Separator orientation="vertical" class="hidden h-7 sm:block" />
        <nav class="order-last basis-full min-w-0 flex-1 overflow-x-auto whitespace-nowrap py-1 md:order-none md:basis-auto md:py-0" :aria-label="t('member.navigation')">
          <RouterLink
            to="/"
            :aria-label="t('member.home.title')"
            :title="t('member.home.title')"
            class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
            active-class="bg-muted font-medium"
            exact-active-class="bg-muted font-medium"
          >
            {{ t('member.nav.home') }}
          </RouterLink>
          <RouterLink
            to="/plans"
            :aria-label="t('member.plans.title')"
            :title="t('member.plans.title')"
            class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
            active-class="bg-muted font-medium"
          >
            {{ t('member.nav.plans') }}
          </RouterLink>
          <RouterLink
            to="/discover"
            :aria-label="t('member.discover.title')"
            :title="t('member.discover.title')"
            class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
            active-class="bg-muted font-medium"
          >
            {{ t('member.nav.discover') }}
          </RouterLink>
          <RouterLink to="/friends" :aria-label="t('friendLinks.title')" :title="t('friendLinks.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
            {{ t('member.nav.friends') }}
          </RouterLink>
          <template v-if="auth.isAuthenticated">
            <RouterLink
              to="/orders"
              :aria-label="t('member.billing.ordersTitle')"
              :title="t('member.billing.ordersTitle')"
              class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
              active-class="bg-muted font-medium"
            >
              {{ t('member.nav.orders') }}
            </RouterLink>
            <RouterLink
              to="/media"
              :aria-label="t('member.media.title')"
              :title="t('member.media.title')"
              class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
              active-class="bg-muted font-medium"
            >
              {{ t('member.nav.media') }}
            </RouterLink>
            <RouterLink to="/folders" :aria-label="t('member.folders.title')" :title="t('member.folders.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.folders') }}
            </RouterLink>
            <RouterLink to="/albums" :aria-label="t('member.albums.title')" :title="t('member.albums.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.albums') }}
            </RouterLink>
            <RouterLink to="/share-links" :aria-label="t('member.shareLinks.title')" :title="t('member.shareLinks.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.shareLinks') }}
            </RouterLink>
            <RouterLink to="/tokens" :aria-label="t('member.tokens.title')" :title="t('member.tokens.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.tokens') }}
            </RouterLink>
            <RouterLink to="/exports" :aria-label="t('member.exports.title')" :title="t('member.exports.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.exports') }}
            </RouterLink>
            <RouterLink to="/reports" :aria-label="t('member.reports.title')" :title="t('member.reports.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3" active-class="bg-muted font-medium">
              {{ t('member.nav.reports') }}
            </RouterLink>
          </template>
          <RouterLink
            v-else
            :to="{ name: 'login', query: { redirect: '/' } }"
            class="ml-auto inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm text-primary transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
          >
            {{ t('member.actions.login') }}
          </RouterLink>
          <RouterLink to="/status" :aria-label="t('statusPage.title')" :title="t('statusPage.title')" class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3">
            {{ t('statusPage.shortTitle') }}
          </RouterLink>
          <RouterLink
            v-if="!auth.isAuthenticated"
            to="/register"
            class="inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm text-primary transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
          >
            {{ t('auth.register') }}
          </RouterLink>
          <RouterLink
            v-if="hasAdminAccess(auth.user?.permissions ?? [])"
            to="/admin"
            :aria-label="t('member.adminLink')"
            :title="t('member.adminLink')"
            class="ml-auto inline-flex shrink-0 rounded-md px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3"
          >
            {{ t('member.adminLink') }}
          </RouterLink>
        </nav>
        <Button variant="ghost" size="sm" :aria-label="t('member.actions.changeLanguage')" @click="toggleLocale">
          <Languages />
          {{ locale === 'zh-CN' ? 'EN' : '中文' }}
        </Button>
        <Button v-if="auth.isAuthenticated" variant="ghost" size="icon" :aria-label="t('member.actions.logout')" @click="logout">
          <LogOut />
        </Button>
      </div>
    </header>
    <MemberAdSlot placement="header" class="mx-auto w-full max-w-6xl px-4 sm:px-6" />
    <div class="mx-auto flex w-full max-w-[80rem] items-start gap-4 px-4 sm:px-6">
      <MemberAdSlot placement="left" class="hidden w-32 shrink-0 xl:block" />
      <main class="min-w-0 flex-1 py-8 sm:py-10">
        <RouterView />
      </main>
      <MemberAdSlot placement="right" class="hidden w-32 shrink-0 xl:block" />
    </div>
    <MemberAdSlot placement="footer" class="mx-auto w-full max-w-6xl px-4 pb-6 sm:px-6" />
    <MemberFooterNavigation />
  </div>
</template>
