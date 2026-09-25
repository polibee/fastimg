import assert from 'node:assert/strict'
import test from 'node:test'
import { adminResourcePath, dashboardResourceRoute, visibleDashboardResources, type DashboardResource } from '../src/lib/dashboard-resources.ts'

const resources: DashboardResource[] = [
  { name: 'users', label: 'Users', route: '/admin/users', permissions: ['admin.users.view'] },
  { name: 'orders', label: 'Orders', route: '/admin/orders', permissions: ['admin.orders.view'] },
]

test('dashboard only exposes resources covered by the current permissions', () => {
  assert.deepEqual(visibleDashboardResources(resources, ['admin.orders.view']).map((resource) => resource.name), ['orders'])
})

test('own-scope resources never appear as global admin resources', () => {
  const resources = [
    { name: 'albums', label: 'Albums', route: '/albums', permissions: ['admin.albums.view'], data_scope: 'own' as const },
    { name: 'advertising', label: 'Advertising', route: '/advertising', permissions: ['admin.advertising.view'], data_scope: 'all' as const },
  ]
  assert.deepEqual(visibleDashboardResources(resources, ['admin.albums.view', 'admin.advertising.view']).map((item) => item.name), ['advertising'])
})

test('dashboard maps resource routes to admin frontend routes', () => {
  assert.equal(dashboardResourceRoute(resources[0]), '/admin/users')
  assert.equal(dashboardResourceRoute({ name: 'roles', route: '/admin/roles' }), '/admin/roles')
})

test('admin resource paths normalize old prefixes and preserve detail suffixes', () => {
  assert.equal(adminResourcePath('/plans', 'plans'), '/admin/plans')
  assert.equal(adminResourcePath('/admin/plans', 'plans'), '/admin/plans')
  assert.equal(adminResourcePath('plans', 'plans'), '/admin/plans')
  assert.equal(adminResourcePath('/', 'plans'), '/admin/plans')
  assert.equal(adminResourcePath('/plans/12/edit', 'plans'), '/admin/plans/12/edit')
})
