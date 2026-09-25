import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member navigation exposes owner-scoped folders and albums without admin URLs', () => {
  const router = read('src/router/index.ts')
  const shell = read('src/core/layouts/MemberShell.vue')
  assert.match(router, /path:\s*'folders',\s*name:\s*'member-folders'/)
  assert.match(router, /path:\s*'albums',\s*name:\s*'member-albums'/)
  assert.match(shell, /to="\/folders"/)
  assert.match(shell, /to="\/albums"/)
  assert.doesNotMatch(router, /path:\s*'\/folders',\s*redirect:\s*['"]\/admin/)
  assert.doesNotMatch(router, /path:\s*'\/albums',\s*redirect:\s*['"]\/admin/)
})

test('member collections use authenticated own-scope APIs and never submit user_id', () => {
  const source = read('src/modules/member/components/MemberCollectionPage.vue')
  assert.match(source, /useAuthStore\(\)/)
  assert.match(source, /\/api\/v1\/me\/\$\{props\.kind\}/)
  assert.doesNotMatch(source, /user_id\s*[:=]/)
  assert.match(source, /useI18n\(\)/)
  assert.match(read('src/modules/member/pages/MemberFoldersPage.vue'), /kind="folders"/)
  assert.match(read('src/modules/member/pages/MemberAlbumsPage.vue'), /kind="albums"/)
})

test('member collection messages are synchronized in both locales', () => {
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    assert.equal(typeof messages.folders?.title, 'string')
    assert.equal(typeof messages.albums?.title, 'string')
    assert.equal(typeof messages.collections?.create, 'string')
  }
})
