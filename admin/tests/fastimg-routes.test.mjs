import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member home, media, and plans are mounted under MemberShell at flat URLs', () => {
  const router = read('src/router/index.ts')
  const memberRoutes = read('src/apps/member/routes.ts')
  const memberRouteParts = read('src/apps/member/route-public.ts') + read('src/apps/member/route-private.ts')
  assert.match(router, /memberRoutes/)
  assert.match(memberRoutes, /path:\s*'\/'[\s\S]*?MemberShell\.vue/)
  assert.match(memberRouteParts, /path:\s*''\s*,\s*name:\s*'member-home'/)
  assert.match(memberRouteParts, /path:\s*'media',\s*name:\s*'member-media'/)
  assert.match(memberRouteParts, /path:\s*'share-links',\s*name:\s*'member-share-links'/)
  assert.match(memberRouteParts, /path:\s*'tokens',\s*name:\s*'member-tokens'/)
  assert.match(memberRouteParts, /path:\s*'plans',\s*name:\s*'member-plans'/)
  assert.match(memberRouteParts, /path:\s*'discover',\s*name:\s*'member-discover'/)
  assert.doesNotMatch(memberRoutes, /path:\s*'\/app',\s*component:[\s\S]*?MemberShell/)
})

test('guest can open the member home and public plans while private member pages stay guarded', () => {
  const memberRoutes = read('src/apps/member/route-public.ts') + read('src/apps/member/route-private.ts')
  const memberHeader = memberRoutes.match(/path:\s*'\/',[\s\S]*?children:\s*\[/u)?.[0] ?? ''
  assert.doesNotMatch(memberHeader, /meta:\s*\{\s*requiresAuth:\s*true\s*\}/)
  assert.match(memberRoutes, /path:\s*'media',[\s\S]*?meta:\s*(?:\{\s*requiresAuth:\s*true\s*\}|a)/)
  assert.match(memberRoutes, /path:\s*'folders',[\s\S]*?meta:\s*(?:\{\s*requiresAuth:\s*true\s*\}|a)/)
  assert.match(memberRoutes, /path:\s*'share-links',[\s\S]*?meta:\s*(?:\{\s*requiresAuth:\s*true\s*\}|a)/)
  assert.match(memberRoutes, /path:\s*'tokens',[\s\S]*?meta:\s*(?:\{\s*requiresAuth:\s*true\s*\}|a)/)
  assert.match(memberRoutes, /path:\s*'plans',\s*name:\s*'member-plans'/)
  assert.match(memberRoutes, /path:\s*''\s*,\s*name:\s*'member-home'/)
  assert.match(memberRoutes, /path:\s*'plans',\s*name:\s*'member-plans'/)
})

test('SSG pages come from the registered public member page manifest', () => {
  const pages = JSON.parse(read('src/apps/member/public-pages.json'))
  const prerender = read('scripts/prerender-seo.mjs')
  assert.deepEqual(pages.map((page) => page.path), ['/', '/plans', '/discover', '/friends', '/page/privacy', '/page/terms', '/page/about'])
  assert.match(prerender, /public-pages\.json/)
})

test('management pages are nested under the guarded admin namespace only', () => {
  const router = read('src/router/index.ts')
  const adminRoutes = read('src/apps/admin/routes.ts')
  assert.match(router, /adminRoutes/)
  assert.match(adminRoutes, /path:\s*'\/admin'[\s\S]*?AdminShell\.vue/)
  assert.match(adminRoutes, /requiresAdminAccess:\s*true/)
  assert.match(router, /hasAdminAccess\(/)
  assert.doesNotMatch(router, /name:\s*'media-library'/)
  assert.doesNotMatch(router, /\/admin\/media/)
})

test('legacy member routes redirect to flat routes and keep query and hash', () => {
  const router = read('src/router/index.ts')
  assert.match(router, /path:\s*'\/app'[\s\S]*?redirect/)
  assert.match(router, /\/app\/media/)
  assert.match(router, /\/app\/plans/)
  assert.match(router, /query:\s*to\.query/)
  assert.match(router, /hash:\s*to\.hash/)
})

test('conflicting member URLs are never redirected by account role', () => {
  const router = read('src/router/index.ts')
  const routerMirror = read('src/router/index.js')
  const memberRoutes = read('src/apps/member/route-public.ts') + read('src/apps/member/route-private.ts')
  assert.doesNotMatch(router, /path:\s*'\/media',\s*redirect:\s*[^\n]*admin/i)
  assert.doesNotMatch(router, /path:\s*'\/plans',\s*redirect:\s*[^\n]*admin/i)
  assert.doesNotMatch(router, /path:\s*'\/folders',\s*redirect:\s*[^\n]*admin/i)
  assert.doesNotMatch(router, /path:\s*'\/albums',\s*redirect:\s*[^\n]*admin/i)
  assert.match(memberRoutes, /path:\s*'folders',\s*name:\s*'member-folders'/)
  assert.match(memberRoutes, /path:\s*'albums',\s*name:\s*'member-albums'/)
  assert.match(routerMirror, /memberRoutes/)
  assert.match(read('src/core/layouts/MemberShell.vue'), /to="\/share-links"[\s\S]*?member\.shareLinks\.title/)
  assert.match(read('src/core/layouts/MemberShell.vue'), /to="\/tokens"[\s\S]*?member\.tokens\.title/)
})

test('admin collection resources remain all-scope while member pages use own-scope APIs', () => {
  assert.match(read('src/modules/folders/resource.ts'), /dataScope:\s*"all"/)
  assert.match(read('src/modules/albums/resource.ts'), /dataScope:\s*"all"/)
  assert.doesNotMatch(read('src/apps/admin/routes.ts'), /adminResourceDefinitions[\s\S]*dataScope === 'own'/)
})

test('login page validates explicit redirect and defaults to member home', () => {
  const login = read('src/modules/auth/pages/LoginPage.vue')
  assert.match(login, /resolveLoginRedirect/)
  assert.match(login, /route\.query\.redirect/)
})

test('authentication has public registration and logout returns to the member home', () => {
  const router = read('src/router/index.ts')
  const publicRoutes = read('src/apps/member/route-public.ts')
  const login = read('src/modules/auth/pages/LoginPage.vue')
  const register = read('src/modules/auth/pages/RegisterPage.vue')
  const memberShell = read('src/core/layouts/MemberShell.vue')
  const adminShell = read('src/core/layouts/AdminShell.vue')
  assert.match(publicRoutes, /RegisterPage\.vue/)
  assert.match(publicRoutes, /VerifyEmailPage\.vue/)
  assert.match(login, /register/)
  assert.match(register, /name="password"[\s\S]*autocomplete="new-password"/)
  assert.match(register, /name="password_confirmation"[\s\S]*autocomplete="new-password"/)
  assert.match(memberShell, /router\.replace\(\{ path: '\/' \}\)/)
  assert.match(adminShell, /router\.replace\(\{ path: '\/' \}\)/)
})
