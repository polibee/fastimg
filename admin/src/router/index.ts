import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { generatedResourceDefinitions, generatedResourceRoutes } from '@/core/resource/generated'
import { adminResourcePath } from '@/lib/dashboard-resources'
import { hasAdminAccess } from '@/lib/admin-access'
import ResourceListPage from '@/core/resource/pages/ResourceListPage.vue'
import ResourceFormPage from '@/core/resource/pages/ResourceFormPage.vue'
import ResourceDetailPage from '@/core/resource/pages/ResourceDetailPage.vue'

const isOwnScope = (resource: unknown) => {
  const scoped = resource as { dataScope?: string; data_scope?: string }
  return scoped.dataScope === 'own' || scoped.data_scope === 'own'
}

const adminResourceNames = new Set<string>(
  generatedResourceDefinitions
    .filter((resource) => !isOwnScope(resource))
    .map((resource) => resource.name),
)

const adminGeneratedResourceRoutes = generatedResourceRoutes
  .filter((route) => {
    const resourceName = String(route.name || '').replace(/-resource-(list|create|edit|detail)$/, '')
    return adminResourceNames.has(resourceName)
  })
  .map((route) => ({ ...route, path: String(route.path).replace(/^\/+/, '') }))

const legacyResourceRedirects: RouteRecordRaw[] = generatedResourceRoutes
  .filter((route) => {
    const resourceName = String(route.name || '').replace(/-resource-(list|create|edit|detail)$/, '')
    // `/folders` and `/albums` are member-owned pages. Their admin resources
    // remain available under `/admin/**`, so do not create a legacy redirect
    // that would shadow the member routes below.
    return adminResourceNames.has(resourceName) && !['folders', 'albums'].includes(resourceName)
  })
  .map((route) => {
    const resourceName = String(route.name || '').replace(/-resource-(list|create|edit|detail)$/, '')
    return {
      path: String(route.path),
      redirect: (to) => ({ path: adminResourcePath(to.path, resourceName), query: to.query, hash: to.hash }),
    }
  })

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/modules/auth/pages/LoginPage.vue'),
    },
    {
      path: '/app/media',
      redirect: (to) => ({ path: '/media', query: to.query, hash: to.hash }),
    },
    {
      path: '/app/plans',
      redirect: (to) => ({ path: '/plans', query: to.query, hash: to.hash }),
    },
    {
      path: '/app',
      redirect: (to) => ({ path: '/', query: to.query, hash: to.hash }),
    },
    { path: '/rbac', redirect: '/admin/rbac' },
    { path: '/audit-logs', redirect: '/admin/audit-logs' },
    { path: '/users', redirect: '/admin/users' },
    { path: '/users/new', redirect: '/admin/users/new' },
    { path: '/users/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
    { path: '/roles', redirect: '/admin/roles' },
    { path: '/roles/new', redirect: '/admin/roles/new' },
    { path: '/roles/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
    { path: '/permissions', redirect: '/admin/permissions' },
    { path: '/plans/new', redirect: '/admin/plans/new' },
    { path: '/plans/:id', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
    { path: '/plans/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
    ...legacyResourceRedirects,
    {
      path: '/',
      component: () => import('@/core/layouts/MemberShell.vue'),
      children: [
        { path: '', name: 'member-home', component: () => import('@/modules/member/pages/MemberHomePage.vue') },
        { path: 'media', name: 'member-media', component: () => import('@/modules/member/pages/MemberMediaPage.vue'), meta: { requiresAuth: true } },
        { path: 'media/:id', name: 'member-media-detail', component: () => import('@/modules/member/pages/MemberMediaDetailPage.vue'), meta: { requiresAuth: true } },
        { path: 'folders', name: 'member-folders', component: () => import('@/modules/member/pages/MemberFoldersPage.vue'), meta: { requiresAuth: true } },
        { path: 'albums', name: 'member-albums', component: () => import('@/modules/member/pages/MemberAlbumsPage.vue'), meta: { requiresAuth: true } },
        { path: 'share-links', name: 'member-share-links', component: () => import('@/modules/member/pages/MemberShareLinksPage.vue'), meta: { requiresAuth: true } },
        { path: 'tokens', name: 'member-tokens', component: () => import('@/modules/member/pages/MemberTokensPage.vue'), meta: { requiresAuth: true } },
        { path: 'plans', name: 'member-plans', component: () => import('@/modules/member/pages/MemberPlansPage.vue') },
        { path: 'checkout/:orderId', name: 'member-checkout', component: () => import('@/modules/member/pages/MemberCheckoutPage.vue'), meta: { requiresAuth: true } },
        { path: 'orders', name: 'member-orders', component: () => import('@/modules/member/pages/MemberOrdersPage.vue'), meta: { requiresAuth: true } },
        { path: 'orders/:id', name: 'member-order-detail', component: () => import('@/modules/member/pages/MemberOrderDetailPage.vue'), meta: { requiresAuth: true } },
      ],
    },
    {
      path: '/admin',
      component: () => import('@/core/layouts/AdminShell.vue'),
      meta: { requiresAuth: true, requiresAdminAccess: true },
      children: [
        { path: '', name: 'admin-home', component: () => import('@/core/pages/AdminHomePage.vue') },
        { path: 'rbac', name: 'rbac', meta: { anyPermissions: ['admin.users.view', 'admin.roles.manage', 'admin.permissions.manage'] }, component: () => import('@/modules/rbac/pages/RBACPage.vue') },
        { path: 'audit-logs', name: 'audit-logs', meta: { permission: 'admin.users.view' }, component: () => import('@/modules/audit/pages/AuditLogPage.vue') },
        { path: 'media-access-logs', name: 'media-access-logs', meta: { permission: 'admin.media_access_logs.view' }, component: () => import('@/modules/access/pages/MediaAccessLogPage.vue') },
        { path: 'orders', name: 'admin-orders', meta: { permission: 'admin.orders.view' }, component: () => import('@/modules/billing/pages/AdminOrdersPage.vue') },
        { path: 'payment-transactions', name: 'admin-payment-transactions', meta: { permission: 'admin.payment_transactions.view' }, component: () => import('@/modules/billing/pages/AdminPaymentTransactionsPage.vue') },
        { path: 'payment-events', name: 'admin-payment-events', meta: { permission: 'admin.payment_events.view' }, component: () => import('@/modules/billing/pages/AdminPaymentEventsPage.vue') },
        { path: 'refunds', name: 'admin-refunds', meta: { permission: 'admin.refunds.view' }, component: () => import('@/modules/billing/pages/AdminRefundsPage.vue') },
        ...adminGeneratedResourceRoutes,
        { path: ':resource(users|roles|permissions|plans)', name: 'resource-list', component: ResourceListPage },
        { path: 'users/new', name: 'user-create', meta: { permission: 'admin.users.manage' }, component: () => import('@/modules/users/pages/UserFormPage.vue') },
        { path: 'users/:id/edit', name: 'user-edit', meta: { permission: 'admin.users.manage' }, component: () => import('@/modules/users/pages/UserFormPage.vue') },
        { path: 'roles/new', name: 'role-create', meta: { permission: 'admin.roles.manage' }, component: () => import('@/modules/roles/pages/RoleFormPage.vue') },
        { path: 'roles/:id/edit', name: 'role-edit', meta: { permission: 'admin.roles.manage' }, component: () => import('@/modules/roles/pages/RoleFormPage.vue') },
        { path: 'plans/new', name: 'plans-resource-create', meta: { permission: 'admin.plans.create' }, props: { resource: 'plans' }, component: ResourceFormPage },
        { path: 'plans/:id/edit', name: 'plans-resource-edit', meta: { permission: 'admin.plans.update' }, props: { resource: 'plans' }, component: ResourceFormPage },
        { path: ':resource(users|roles|permissions|plans)/:id', name: 'resource-detail', component: ResourceDetailPage },
        { path: 'loading', name: 'loading', component: () => import('@/core/pages/LoadingPage.vue') },
        { path: 'empty', name: 'empty', component: () => import('@/core/pages/EmptyPage.vue') },
        { path: 'error', name: 'error', component: () => import('@/core/pages/ErrorPage.vue') },
      ],
    },
    { path: '/forbidden', name: 'forbidden', component: () => import('@/core/pages/ForbiddenPage.vue') },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.restore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { path: '/' }
  }
  if (to.meta.requiresAdminAccess && !hasAdminAccess(auth.user?.permissions || [])) {
    return { name: 'forbidden' }
  }
  if (to.meta.permission && !auth.can(String(to.meta.permission))) {
    return { name: 'forbidden' }
  }
  if (to.meta.anyPermissions && !auth.canAny(to.meta.anyPermissions as string[])) {
    return { name: 'forbidden' }
  }
  const resource = typeof to.params.resource === 'string' ? to.params.resource : ''
  const resourcePermission: Record<string, string> = { users: 'admin.users.view', roles: 'admin.roles.manage', permissions: 'admin.permissions.manage', plans: 'admin.plans.view' }
  if (resource && resourcePermission[resource] && !auth.can(resourcePermission[resource])) {
    return { name: 'forbidden' }
  }
})

export default router
