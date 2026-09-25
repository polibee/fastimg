import assert from 'node:assert/strict'
import test from 'node:test'

const moduleUnderTest = await import('../src/lib/login-redirect.ts').catch(() => undefined)

test('accepts same-origin application routes with query and hash', () => {
  assert.ok(moduleUnderTest, 'login redirect helper module exists')
  const { resolveLoginRedirect } = moduleUnderTest
  assert.equal(resolveLoginRedirect('/'), '/')
  assert.equal(resolveLoginRedirect('/media'), '/media')
  assert.equal(resolveLoginRedirect('/admin/plans?tab=active#top'), '/admin/plans?tab=active#top')
})

test('rejects external, protocol-relative, backslash, and non-string redirects', () => {
  assert.ok(moduleUnderTest, 'login redirect helper module exists')
  const { resolveLoginRedirect } = moduleUnderTest
  assert.equal(resolveLoginRedirect('https://evil.example'), '/')
  assert.equal(resolveLoginRedirect('//evil.example'), '/')
  assert.equal(resolveLoginRedirect('\\\\evil.example'), '/')
  assert.equal(resolveLoginRedirect(42), '/')
})
