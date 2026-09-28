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
  assert.match(page, /friendApi\.submit/)
  assert.match(page, /grid/)
  assert.match(page, /accepted|pending/i)
})
