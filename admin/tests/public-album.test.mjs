import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const root = new URL('../', import.meta.url)

async function text(path) {
  return readFile(new URL(path, root), 'utf8')
}

test('public album has a guest route and SEO page', async () => {
  const routes = await text('src/apps/member/route-public.ts')
  const page = await text('src/modules/member/pages/PublicAlbumPage.vue')
  assert.match(routes, /path:'a\/:id'/)
  assert.match(routes, /PublicAlbumPage/)
  assert.match(page, /setPageSEO/)
  assert.match(page, /api\/v1\/public\/albums/)
  assert.match(page, /thumbnail_url/)
})

test('public album API is documented and does not reuse member album API', async () => {
  const routes = await text('../backend/routes/web.go')
  const openapi = await text('../backend/app/openapi/spec.go')
  assert.match(routes, /facades\.Route\(\)\.Get\("\/api\/v1\/public\/albums\/\{id\}"/)
  assert.match(openapi, /public\/albums\/{id}/)
  assert.match(routes, /RequireMemberAuthentication\(\), adminmiddleware\.RequireMemberSession\(\)\)\.Get\("\/api\/v1\/me\/albums\/{id}\/media"/)
})

test('SSG supports explicitly selected public albums', async () => {
  const script = await text('scripts/prerender-seo.mjs')
  assert.match(script, /SSG_PUBLIC_ALBUM_IDS/)
  assert.match(script, /\/a\//)
})
