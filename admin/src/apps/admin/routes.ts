import type { RouteRecordRaw } from 'vue-router'
import { generatedResourceDefinitions, generatedResourceRoutes } from '@/core/resource/generated'
import { adminResourcePath } from '@/lib/dashboard-resources'
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

export const adminLegacyResourceRedirects: RouteRecordRaw[] = generatedResourceRoutes
  .filter((route) => {
    const resourceName = String(route.name || '').replace(/-resource-(list|create|edit|detail)$/, '')
    // Member-owned resources must remain on their flat member URLs.
    return adminResourceNames.has(resourceName) && !['media', 'folders', 'albums', 'plans'].includes(resourceName)
  })
  .map((route) => {
    const resourceName = String(route.name || '').replace(/-resource-(list|create|edit|detail)$/, '')
    return {
      path: String(route.path),
      redirect: (to) => ({
        path: adminResourcePath(to.path, resourceName),
        query: to.query,
        hash: to.hash,
      }),
    }
  })

/**
 * Administrator application boundary.
 *
 * This route tree is permission-gated at the root. The pages and generated
 * resource routes here are never part of the member application surface.
 */
export const adminRoutes: RouteRecordRaw = {
  path: '/admin',
  component: () => import('@/core/layouts/AdminShell.vue'),
  meta: { requiresAuth: true, requiresAdminAccess: true },
  children: [
    { path: '', name: 'admin-home', component: () => import('@/core/pages/AdminHomePage.vue') },
    { path: 'rbac', name: 'rbac', meta: { anyPermissions: ['admin.users.view', 'admin.roles.manage', 'admin.permissions.manage'] }, component: () => import('@/modules/rbac/pages/RBACPage.vue') },
    { path: 'audit-logs', name: 'audit-logs', meta: { permission: 'admin.users.view' }, component: () => import('@/modules/audit/pages/AuditLogPage.vue') },
    { path: 'media-access-logs', name: 'media-access-logs', meta: { permission: 'admin.media_access_logs.view' }, component: () => import('@/modules/access/pages/MediaAccessLogPage.vue') },
    { path: 'tasks', name: 'admin-tasks', meta: { permission: 'admin.tasks.view' }, component: () => import('@/modules/tasks/pages/AdminTasksPage.vue') },
    { path: 'settings', name: 'admin-settings', meta: { permission: 'admin.settings.manage' }, component: () => import('@/modules/settings/pages/AdminSettingsPage.vue') },
    { path: 'storage', name: 'admin-storage', meta: { permission: 'admin.storage.view' }, component: () => import('@/modules/settings/pages/AdminStoragePage.vue') },
    { path: 'statistics', name: 'admin-statistics', meta: { permission: 'admin.users.view' }, component: () => import('@/modules/statistics/pages/AdminStatisticsPage.vue') },
    { path: 'albums/:id/media', name: 'admin-album-media', meta: { permission: 'admin.albums.view' }, component: () => import('@/modules/albums/pages/AdminAlbumMediaPage.vue') },
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
}
