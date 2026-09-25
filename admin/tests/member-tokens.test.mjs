import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member Token page uses session-scoped Token management APIs and never stores a token in a URL', () => {
  const page = read('src/modules/member/pages/MemberTokensPage.vue')
  assert.match(page, /\/api\/v1\/tokens/)
  assert.match(page, /\/rotate/)
  assert.match(page, /method:\s*'DELETE'/)
  assert.match(page, /navigator\.clipboard\.writeText\(createdToken\.value\.token\)/)
  assert.doesNotMatch(page, /localStorage\.(setItem|getItem)\([^)]*token/i)
  assert.doesNotMatch(page, /window\.location.*token/i)
})

test('member Token page makes permanent and custom expiry explicit', () => {
  const page = read('src/modules/member/pages/MemberTokensPage.vue')
  assert.match(page, /expiryMode/)
  assert.match(page, /permanent/)
  assert.match(page, /custom/)
  assert.match(page, /type="datetime-local"/)
  assert.match(page, /expiryMode\.value === 'permanent'/)
})

test('member Token copy uses synchronized localized labels', () => {
  const page = read('src/modules/member/pages/MemberTokensPage.vue')
  assert.match(page, /useI18n\(\)/)
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    assert.equal(typeof messages.tokens?.title, 'string')
    assert.equal(typeof messages.tokens?.createdDescription, 'string')
    assert.equal(typeof messages.tokens?.fixedScopes, 'string')
  }
})
