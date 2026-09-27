import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => {
  const target = path.join(adminRoot, relativePath)
  return existsSync(target) ? readFileSync(target, 'utf8') : ''
}

test('member home is the default route and owns the flat member links', () => {
  const router = read('src/router/index.ts')
  const memberRoutes = read('src/apps/member/route-public.ts')
  const home = read('src/modules/member/pages/MemberHomePage.vue')
  assert.match(router, /memberRoutes/)
  assert.match(memberRoutes, /name:\s*['"]member-home['"][\s\S]*?MemberHomePage/)
  assert.match(home, /MemberUploadPanel/)
  assert.match(home, /to="\/media"/)
  assert.match(home, /to="\/plans"/)
  assert.match(home, /emptyRecent/)
  assert.match(home, /MemberUploadPanel/)
})

test('member home loads recent own-scope media and server-reported plan usage', () => {
  const home = read('src/modules/member/pages/MemberHomePage.vue')
  assert.match(home, /\/api\/v1\/media/)
  assert.match(home, /\/api\/v1\/subscription/)
  assert.match(home, /\/api\/v1\/me\/usage/)
  assert.match(home, /apiFetchBlob/)
  assert.doesNotMatch(home, /user_id\s*[:=]/)
  assert.match(home, /member\.home\.errors\.recentMedia/)
})

test('shared upload panel supports keyboard file selection, drop, processing feedback, and retry', () => {
  const panel = read('src/modules/member/components/MemberUploadPanel.vue')
  const composable = read('src/modules/member/composables/useMemberUpload.ts')
  assert.match(panel, /useMemberUpload/)
  assert.match(panel, /type="file"/)
  assert.match(panel, /@drop/)
  assert.match(panel, /member\.upload\.processing/)
  assert.match(panel, /member\.upload\.ready/)
  assert.match(panel, /member\.upload\.retry/)
  assert.match(composable, /'uploading'\s*\|\s*'processing'\s*\|\s*'ready'\s*\|\s*'failed'/)
  assert.match(composable, /\/api\/v1\/uploads/)
  assert.match(composable, /\/api\/v1\/uploads\/\$\{/)
  assert.doesNotMatch(composable, /user_id\s*[:=]/)
})

test('guest home keeps upload behind authentication and exposes a login CTA', () => {
  const panel = read('src/modules/member/components/MemberUploadPanel.vue')
  const home = read('src/modules/member/pages/MemberHomePage.vue')
  assert.match(panel, /auth\.isAuthenticated/)
  assert.match(panel, /member\.upload\.loginRequired/)
  assert.match(panel, /member\.upload\.loginToUpload/)
  assert.match(home, /auth\.isAuthenticated/)
})

test('member media page reuses the shared upload panel instead of owning a second uploader', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  assert.match(page, /MemberUploadPanel/)
  assert.doesNotMatch(page, /async function uploadFiles/)
  assert.doesNotMatch(page, /pendingUploadIDs/)
})
