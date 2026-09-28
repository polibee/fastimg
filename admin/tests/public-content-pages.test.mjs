import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('public content pages, friends, and footer navigation are guest routes', () => {
  const routes = read('src/apps/member/route-public.ts')
  assert.match(routes, /path:'friends',name:'member-friends'/)
  assert.match(routes, /path:'page\/:slug',name:'member-content-page'/)
  assert.match(read('src/modules/member/pages/SiteContentPage.vue'), /publicContentApi\.show/)
  assert.match(read('src/modules/member/components/MemberFooterNavigation.vue'), /footerPublicApi\.list/)
})

test('SSG manifest includes default policy pages and friend links', () => {
  const pages = JSON.parse(read('src/apps/member/public-pages.json'))
  const paths = pages.map((page) => page.path)
  for (const expected of ['/friends', '/page/privacy', '/page/terms', '/page/about']) assert.ok(paths.includes(expected), expected)
})
