import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member plans live under a dedicated MemberShell route, separate from AdminShell', () => {
  const router = read('src/router/index.ts')
  const memberRoutes = read('src/apps/member/routes.ts') + read('src/apps/member/route-public.ts')
  assert.match(router, /memberRoutes/)
  assert.match(memberRoutes, /path:\s*'\/'/)
  assert.match(memberRoutes, /path:\s*'plans',\s*name:\s*'member-plans'/)
  assert.match(memberRoutes, /MemberShell\.vue/)
})

test('member plan page reads public plans and the authenticated member subscription without user_id input', () => {
  const page = read('src/modules/member/pages/MemberPlansPage.vue')
  assert.match(page, /\/api\/v1\/plans/)
  assert.match(page, /\/api\/v1\/subscription/)
  assert.match(page, /useAuthStore\(\)/)
  assert.doesNotMatch(page, /user_id\s*[:=]/)
  assert.match(page, /createMemberOrder/)
  assert.match(page, /member-checkout/)
})

test('guest can load the public plan catalog and sees a login prompt for private usage', () => {
  const page = read('src/modules/member/pages/MemberPlansPage.vue')
  const chinese = JSON.parse(readFileSync(path.join(adminRoot, 'src/locales/zh-CN/member.json'), 'utf8'))
  const english = JSON.parse(readFileSync(path.join(adminRoot, 'src/locales/en-US/member.json'), 'utf8'))
  assert.match(page, /apiFetch<Plan\[\]>\('\/api\/v1\/plans'/)
  assert.match(page, /if\s*\(auth\.token\)/)
  assert.match(page, /member\.plans\.loginToViewUsage/)
  assert.equal(typeof chinese.plans.guestDescription, 'string')
  assert.equal(typeof english.plans.guestDescription, 'string')
})

test('member plan page uses localized member keys for user-facing content', () => {
  const page = read('src/modules/member/pages/MemberPlansPage.vue')
  assert.match(page, /useI18n\(\)/)
  assert.match(page, /t\('member\./)
  assert.doesNotMatch(page, /function\s+t\(zh:/)
})

test('member plan page displays server-reported measured bandwidth', () => {
	const page = read('src/modules/member/pages/MemberPlansPage.vue')
  const chinese = JSON.parse(readFileSync(path.join(adminRoot, 'src/locales/zh-CN/member.json'), 'utf8'))
  const english = JSON.parse(readFileSync(path.join(adminRoot, 'src/locales/en-US/member.json'), 'utf8'))
	assert.match(page, /bandwidth_metered:\s*boolean/)
	assert.match(page, /computed\(\(\)\s*=>\s*usage\.value\?\.bandwidth_metered/)
	assert.match(page, /member\.plans\.bandwidthNotMetered/)
  assert.equal(typeof chinese.plans.bandwidthNotMetered, 'string')
  assert.equal(typeof english.plans.bandwidthNotMetered, 'string')
})
