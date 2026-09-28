<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Activity, ArrowRight, BarChart3, ClipboardList, FileText, HardDrive, Languages, LayoutDashboard, ListTodo, LogOut, Search, Settings2, ShieldCheck, Unplug, WalletCards } from '@lucide/vue'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarHeader, SidebarInset, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { useAuthStore } from '@/stores/auth'
import { generatedApi, type GlobalSearchResult, type ResourceManifest } from '@/generated/api'
import { adminResourcePath, dashboardResourceRoute, visibleDashboardResources } from '@/lib/dashboard-resources'
import { groupResourceNavigation } from '@/lib/resource-navigation'
import NotificationMenu from '@/core/notifications/NotificationMenu.vue'
import { localizedResourceLabel } from '@/core/resource/resource-i18n'
import { storageConnections, type StorageOverview } from '@/modules/settings/storage-api'

const { t, te, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const resourceManifests = ref<ResourceManifest[]>([])
const storageOverview = ref<StorageOverview>()
const hasEnabledStorage = computed(() => Boolean(storageOverview.value?.connections.some((connection) => connection.enabled)))
const adminResourceManifests = computed(() => resourceManifests.value.filter((item) => item.data_scope !== 'own'))
const searchOpen = ref(false)
const searchResults = computed(() => visibleDashboardResources(adminResourceManifests.value, auth.user?.permissions || []))
const resourceNavigationGroups = computed(() => groupResourceNavigation(adminResourceManifests.value, auth.user?.permissions || []))
const globalSearchResults = ref<GlobalSearchResult[]>([])
const globalSearchLoading = ref(false)
let globalSearchTimer: ReturnType<typeof setTimeout> | undefined
let globalSearchRequest = 0
const breadcrumbResource = computed(() => {
  const routeResource = typeof route.params.resource === 'string' ? route.params.resource : ''
  if (routeResource) return routeResource
  const routeName = String(route.name || '')
  if (routeName.includes('user')) return 'users'
  if (routeName.includes('role')) return 'roles'
  if (routeName.includes('permission')) return 'permissions'
  const generatedResource = routeName.match(/^(.+)-resource-/)?.[1]
  return generatedResource || ''
})
const breadcrumbLabel = computed(() => {
  if (route.name === 'admin-home') return t('auth.dashboard')
  if (route.name === 'rbac') return t('rbac.title')
  if (route.name === 'audit-logs') return t('auth.auditLogs')
  if (route.name === 'media-access-logs') return t('auth.mediaAccessLogs')
  if (route.name === 'admin-tasks') return t('tasks.title')
  if (route.name === 'admin-orders') return t('billing.admin.orders')
  if (route.name === 'admin-payment-transactions') return t('billing.admin.transactions')
  if (route.name === 'admin-payment-events') return t('billing.admin.events')
  if (route.name === 'admin-refunds') return t('billing.admin.refunds')
  if (route.name === 'admin-settings') return t('settings.title')
  if (String(route.name).startsWith('admin-content-page')) return t('content.pagesTitle')
  if (route.name === 'admin-storage') return t('storage.title')
  if (route.name === 'admin-statistics') return t('statistics.title')
  const resource = resourceManifests.value.find((item) => item.name === breadcrumbResource.value)
  return resource ? localizedResourceLabel(t, te, resource.name, resource.label) : breadcrumbResource.value || t('auth.dashboard')
})

function displayResourceLabel(resource: { name: string; label: string }) {
  return localizedResourceLabel(t, te, resource.name, resource.label)
}

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

async function logoutAll() {
  try {
    await auth.logoutAll()
  } finally {
    await router.replace({ path: '/' })
  }
}

function openResource(resource: { name: string; route: string }) {
  searchOpen.value = false
  void router.push(dashboardResourceRoute(resource))
}

function openSearchResult(result: GlobalSearchResult) {
  if (!adminResourceManifests.value.some((resource) => resource.name === result.resource)) return
  searchOpen.value = false
  void router.push(adminResourcePath(result.route, result.resource))
}

function resourceGroupLabel(name: string) {
  const key = `core.resourceGroups.${name}`
  return te(key) ? t(key) : name
}

function handleSearchInput(value: string) {
  const query = value.trim()
  globalSearchRequest += 1
  const request = globalSearchRequest
  if (globalSearchTimer) clearTimeout(globalSearchTimer)
  if (query.length < 2 || !auth.token) {
    globalSearchResults.value = []
    globalSearchLoading.value = false
    return
  }
  globalSearchLoading.value = true
  globalSearchTimer = setTimeout(async () => {
    try {
      const results = await generatedApi.globalSearch(query, auth.token as string)
      if (request === globalSearchRequest) globalSearchResults.value = results
    } catch {
      if (request === globalSearchRequest) globalSearchResults.value = []
    } finally {
      if (request === globalSearchRequest) globalSearchLoading.value = false
    }
  }, 250)
}

function handleSearchShortcut(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    searchOpen.value = true
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleSearchShortcut)
  if (!auth.token) return
  try { resourceManifests.value = await generatedApi.resourceRegistry(auth.token) } catch { resourceManifests.value = [] }
  if (auth.can('admin.storage.view')) {
    try { storageOverview.value = await storageConnections(auth.token) } catch { storageOverview.value = undefined }
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleSearchShortcut)
  if (globalSearchTimer) clearTimeout(globalSearchTimer)
})
</script>

<template>
  <SidebarProvider>
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" :tooltip="t('core.appName')">
              <span class="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <LayoutDashboard />
              </span>
              <span class="truncate font-semibold">{{ t('core.appName') }}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>{{ t('core.navigation') }}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton as-child :is-active="$route.name === 'admin-home'" :tooltip="t('auth.dashboard')">
                  <RouterLink to="/admin">
                    <LayoutDashboard />
                    <span>{{ t('auth.dashboard') }}</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem v-if="auth.canAny(['admin.users.view', 'admin.roles.manage', 'admin.permissions.manage'])">
                <SidebarMenuButton as-child :is-active="$route.name === 'rbac'" :tooltip="t('rbac.title')">
                  <RouterLink to="/admin/rbac"><ShieldCheck /><span>{{ t('rbac.title') }}</span></RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.users.view')">
                <SidebarMenuButton as-child :is-active="$route.name === 'audit-logs'" :tooltip="t('auth.auditLogs')">
                  <RouterLink to="/admin/audit-logs"><ClipboardList /><span>{{ t('auth.auditLogs') }}</span></RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.media_access_logs.view')">
                <SidebarMenuButton as-child :is-active="$route.name === 'media-access-logs'" :tooltip="t('auth.mediaAccessLogs')">
                  <RouterLink to="/admin/media-access-logs"><Activity /><span>{{ t('auth.mediaAccessLogs') }}</span></RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.tasks.view')">
                <SidebarMenuButton as-child :is-active="$route.name === 'admin-tasks'" :tooltip="t('tasks.title')">
                  <RouterLink to="/admin/tasks"><ListTodo /><span>{{ t('tasks.title') }}</span></RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.orders.view')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-orders'" :tooltip="t('billing.admin.orders')"><RouterLink to="/admin/orders"><WalletCards /><span>{{ t('billing.admin.orders') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.payment_transactions.view')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-payment-transactions'" :tooltip="t('billing.admin.transactions')"><RouterLink to="/admin/payment-transactions"><WalletCards /><span>{{ t('billing.admin.transactions') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.payment_events.view')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-payment-events'" :tooltip="t('billing.admin.events')"><RouterLink to="/admin/payment-events"><WalletCards /><span>{{ t('billing.admin.events') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.refunds.view')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-refunds'" :tooltip="t('billing.admin.refunds')"><RouterLink to="/admin/refunds"><WalletCards /><span>{{ t('billing.admin.refunds') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.users.view')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-statistics'" :tooltip="t('statistics.title')"><RouterLink to="/admin/statistics"><BarChart3 /><span>{{ t('statistics.title') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.settings.manage')"><SidebarMenuButton as-child :is-active="$route.name === 'admin-settings'" :tooltip="t('settings.title')"><RouterLink to="/admin/settings"><Settings2 /><span>{{ t('settings.title') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.content_pages.view')"><SidebarMenuButton as-child :is-active="String($route.name).startsWith('admin-content-page')" :tooltip="t('content.pagesTitle')"><RouterLink to="/admin/content-pages"><FileText /><span>{{ t('content.pagesTitle') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
              <SidebarMenuItem v-if="auth.can('admin.storage.view') && hasEnabledStorage"><SidebarMenuButton as-child :is-active="$route.name === 'admin-storage'" :tooltip="t('storage.title')"><RouterLink to="/admin/storage"><HardDrive /><span>{{ t('storage.title') }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup v-for="group in resourceNavigationGroups" :key="group.name">
          <SidebarGroupLabel>{{ resourceGroupLabel(group.name) }}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem v-for="item in group.items" :key="item.name">
                <SidebarMenuButton as-child :is-active="$route.path.startsWith(dashboardResourceRoute(item))" :tooltip="displayResourceLabel(item)">
                  <RouterLink :to="dashboardResourceRoute(item)"><LayoutDashboard /><span>{{ displayResourceLabel(item) }}</span></RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <SidebarMenuButton size="lg" :tooltip="auth.user?.name">
                  <Avatar class="size-8 rounded-lg">
                    <AvatarFallback class="rounded-lg">{{ auth.user?.name?.slice(0, 1).toUpperCase() }}</AvatarFallback>
                  </Avatar>
                  <span class="truncate">{{ auth.user?.name }}</span>
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="right" align="end" class="w-56">
                <DropdownMenuLabel>{{ auth.user?.email }}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="logout">
                  <LogOut />
                  {{ t('auth.logout') }}
                </DropdownMenuItem>
                <DropdownMenuItem @click="logoutAll">
                  <Unplug />
                  {{ t('auth.logoutAll') }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>

    <SidebarInset>
      <header class="flex h-16 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger class="-ml-1" />
        <Separator orientation="vertical" class="mr-2 h-4" />
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem class="hidden md:block">
              <BreadcrumbLink href="#">{{ t('core.appName') }}</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator class="hidden md:block" />
            <BreadcrumbItem>
              <BreadcrumbPage>{{ breadcrumbLabel }}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
        <Button variant="outline" size="sm" as-child class="shrink-0" :aria-label="t('core.memberFrontend')">
          <RouterLink to="/">
            <ArrowRight data-icon="inline-start" />
            <span class="hidden sm:inline">{{ t('core.memberFrontend') }}</span>
            <span class="sr-only sm:hidden">{{ t('core.memberFrontend') }}</span>
          </RouterLink>
        </Button>
        <Button variant="outline" class="ml-auto hidden h-9 w-56 justify-start gap-2 font-normal text-muted-foreground sm:flex" @click="searchOpen = true">
          <Search data-icon="inline-start" />
          <span>{{ t('core.searchResources') }}</span>
          <kbd class="ml-auto rounded border bg-muted px-1.5 py-0.5 text-[10px]">⌘K</kbd>
        </Button>
        <Button variant="ghost" size="icon" class="ml-auto sm:hidden" :aria-label="t('core.searchResources')" @click="searchOpen = true">
          <Search />
        </Button>
        <div class="ml-auto md:hidden">
          <Button variant="ghost" size="icon" @click="logout">
            <LogOut />
            <span class="sr-only">{{ t('auth.logout') }}</span>
          </Button>
        </div>
        <NotificationMenu />
        <Button variant="ghost" size="sm" @click="toggleLocale"><Languages />{{ locale === 'zh-CN' ? 'EN' : '中文' }}</Button>
      </header>
      <CommandDialog v-model:open="searchOpen" :title="t('core.searchResources')" :description="t('core.searchResources')">
        <CommandInput :placeholder="t('core.searchResources')" @update:model-value="handleSearchInput" />
        <CommandList>
          <CommandEmpty v-if="!globalSearchLoading">{{ t('core.noSearchResults') }}</CommandEmpty>
          <CommandGroup v-if="globalSearchResults.length" :heading="t('core.searchData')">
            <CommandItem v-for="result in globalSearchResults" :key="result.resource + ':' + result.id" :value="result.title + ' ' + (result.subtitle || '')" @select="openSearchResult(result)">
              <span>{{ result.title }}</span>
              <span v-if="result.subtitle" class="truncate text-xs text-muted-foreground">{{ result.subtitle }}</span>
              <span class="ml-auto text-xs text-muted-foreground">{{ result.label }}</span>
            </CommandItem>
          </CommandGroup>
          <CommandGroup :heading="t('core.searchResources')">
            <CommandItem v-for="resource in searchResults" :key="resource.name" :value="resource.name + ' ' + displayResourceLabel(resource)" @select="openResource(resource)">
              <span>{{ displayResourceLabel(resource) }}</span>
              <span class="ml-auto text-xs text-muted-foreground">{{ resource.name }}</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
      <div class="flex flex-1 flex-col gap-4 p-4 pt-6">
        <RouterView />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
