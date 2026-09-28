import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member navigation uses compact localized labels while preserving full accessible names', () => {
  const shell = read('src/core/layouts/MemberShell.vue')
  assert.match(shell, /member\.nav\.media/)
  assert.match(shell, /:aria-label="t\('member\.media\.title'\)"/)
  assert.match(shell, /:title="t\('member\.media\.title'\)"/)
  assert.match(shell, /overflow-x-auto/)
  assert.match(shell, /order-last basis-full/)
})

test('compact member navigation labels exist in both locales', () => {
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    for (const key of ['home', 'plans', 'media', 'folders', 'albums', 'shareLinks', 'tokens']) {
      assert.equal(typeof messages.nav?.[key], 'string', `${locale} nav.${key} must be localized`)
    }
  }
})

test('member and admin welcome surfaces do not use an email as a visible fallback name', () => {
  const memberShell = read('src/core/layouts/MemberShell.vue')
  const adminHome = read('src/core/pages/AdminHomePage.vue')
  const adminShell = read('src/core/layouts/AdminShell.vue')
  assert.doesNotMatch(memberShell, /auth\.user\?\.email|auth\.user\.email/)
  assert.doesNotMatch(adminHome, /auth\.user\?\.email|auth\.user\.email/)
  assert.match(adminShell, /accountLabel/)
  assert.doesNotMatch(adminShell, /\{\{\s*auth\.user\?\.name\s*\}\}/)
})
