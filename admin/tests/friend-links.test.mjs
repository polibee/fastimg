import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('friend links expose a public page, guest submission, and moderated admin route', () => {
  const web = read('../backend/routes/web.go')
  const publicRoutes = read('src/apps/member/route-public.ts')
  const adminRoutes = read('src/apps/admin/routes.ts')
  assert.match(web, /api\/v1\/friend-links/)
  assert.match(publicRoutes, /friends/)
  assert.match(adminRoutes, /path:\s*'friend-links',\s*name:\s*'admin-friend-links'[\s\S]*?admin\.friend_links\.view/)
  assert.match(read('src/modules/friend-links/api.ts'), /friend-links/)
})

test('public friend links are rendered as cards and submissions stay pending', () => {
  const page = read('src/modules/member/pages/FriendLinksPage.vue')
  assert.match(page, /friendApi\.list/)
  assert.match(page, /friendApi\.presentation/)
  assert.match(read('src/modules/friend-links/api.ts'), /site\/friend-links\/presentation/)
  assert.match(page, /friendApi\.submit/)
  assert.match(page, /grid/)
  assert.match(page, /accepted|pending/i)
})

test('friend-link copy is editable through bilingual public-content settings', () => {
  const schema = read('src/modules/settings/settings-schema.ts')
  const backend = read('../backend/routes/web.go')
  const settingsPage = read('src/modules/settings/pages/AdminSettingsPage.vue')
  const adminPage = read('src/modules/friend-links/pages/AdminFriendLinksPage.vue')
  assert.match(schema, /friend_links\.zh_cn\.title/)
  assert.match(schema, /friend_links\.en_us\.title/)
  assert.match(backend, /site\/friend-links\/presentation/)
  assert.match(settingsPage, /friend-links-copy/)
  assert.match(settingsPage, /settings\.contentLanguages\.zhCN/)
  assert.match(settingsPage, /settings\.contentLanguages\.enUS/)
  assert.match(adminPage, /admin\/settings#friend-links-copy/)
})

test('admin friend-link review exposes mapped fields, filters, notes, and moderation actions', () => {
  const page = read('src/modules/friend-links/pages/AdminFriendLinksPage.vue')
  assert.match(page, /friendApi\.adminList\(statusFilter\.value/)
  assert.match(page, /friendApi\.review\(link\.id/)
  assert.match(page, /reviewNotes/)
  assert.match(page, /value="hidden"/)
  assert.match(page, /admin\.friend_links\.moderate/)
  assert.match(page, /link\.contact_email/)
  assert.match(page, /link\.review_note/)
  assert.match(page, /friendLinks\.valueSeparator/)
})
