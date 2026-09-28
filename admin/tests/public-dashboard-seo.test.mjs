import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('admin dashboard renders every real overview metric returned by the API', () => {
  const page = read('src/core/pages/AdminHomePage.vue')
  for (const metric of ['users', 'media', 'albums', 'folders', 'orders', 'payment_transactions']) {
    assert.match(page, new RegExp(`['\"]${metric}['\"]`), `dashboard should expose ${metric}`)
  }
  assert.match(page, /generatedApi\.overview\(auth\.token\)/)
})

test('SSG keeps a Vue mount point while injecting crawlable public content', () => {
  const prerender = read('scripts/prerender-seo.mjs')
  assert.match(prerender, /id="app"/)
  assert.match(prerender, /data-fastimg-ssg/)
  assert.match(prerender, /<title>\$\{escapeHTML\(page\.title\)\}<\/title>/)
  assert.match(prerender, /api\/v1\/site\/pages/)
  assert.match(prerender, /renderNode/)
  assert.doesNotMatch(prerender, /replace\('\<div id="app"\>\<\/div\>', content\)/)
})

test('public member routes remain outside the authentication guard', () => {
  const routes = read('src/apps/member/route-public.ts')
  assert.match(routes, /name:\s*'member-home'/)
  assert.match(routes, /name:\s*'member-plans'/)
  assert.match(routes, /name:\s*'member-discover'/)
  assert.doesNotMatch(routes, /requiresAuth/)
})
