import assert from 'node:assert/strict'
import test from 'node:test'

const moduleUnderTest = await import('../src/lib/admin-access.ts').catch(() => undefined)

test('ordinary member and own-scope resource permissions do not grant admin shell access', () => {
  assert.ok(moduleUnderTest, 'admin access helper module exists')
  const { hasAdminAccess } = moduleUnderTest
  assert.equal(hasAdminAccess([]), false)
  assert.equal(hasAdminAccess(['media.upload']), false)
  assert.equal(hasAdminAccess(['admin.folders.view']), false)
  assert.equal(hasAdminAccess(['admin.albums.update']), false)
})

test('a permission for a registered admin resource grants admin shell access', () => {
  assert.ok(moduleUnderTest, 'admin access helper module exists')
  const { hasAdminAccess } = moduleUnderTest
  assert.equal(hasAdminAccess(['admin.plans.view']), true)
  assert.equal(hasAdminAccess(['admin.roles.manage']), true)
})
