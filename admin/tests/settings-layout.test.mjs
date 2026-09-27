import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('settings schema exposes the fixed business group order and development sender default', () => {
  const schema = read('src/modules/settings/settings-schema.ts')
  assert.match(schema, /site|basic/i)
  assert.match(schema, /auth|registration/i)
  assert.match(schema, /email/)
  assert.match(schema, /upload|media/i)
  assert.match(schema, /seo|discovery/i)
  assert.match(schema, /payment|billing/i)
  assert.match(schema, /statistics|stats/i)
  assert.match(schema, /code|custom/i)
  assert.match(schema, /other/)
  assert.match(schema, /noreply@example\.com/)
})

test('settings page uses shadcn selects and aligned boolean controls', () => {
  const page = read('src/modules/settings/pages/AdminSettingsPage.vue')
  assert.match(page, /SelectTrigger/)
  assert.match(page, /SelectValue/)
  assert.match(page, /SelectContent/)
  assert.match(page, /SelectItem/)
  assert.match(page, /Switch|Checkbox/)
  assert.match(page, /@\/components\/ui\/checkbox/)
  assert.doesNotMatch(page, /@\/components\/ui\/switch/)
  assert.doesNotMatch(page, /<select\b/)
  assert.doesNotMatch(page, /<input[^>]+type="checkbox"/)
  assert.match(page, /min-h-9[^\n]*border-input[^\n]*bg-background/)
})

test('settings page exposes backend sitemap and robots links', () => {
  const page = read('src/modules/settings/pages/AdminSettingsPage.vue')
  assert.match(page, /publicEndpointURL\('\/sitemap\.xml'\)/)
  assert.match(page, /publicEndpointURL\('\/robots\.txt'\)/)
  assert.match(page, /settings\.openSitemap/)
  assert.match(page, /settings\.openRobots/)
})

test('development email defaults remain saveable instead of being snapshotted as persisted values', () => {
  const page = read('src/modules/settings/pages/AdminSettingsPage.vue')
  const mounted = page.slice(page.indexOf('onMounted(async'))
  assert.match(mounted, /snapshotInitialValues\(\)[\s\S]*applyEmailDefaults\(\)/)
})

test('settings locales contain the layout copy in both languages', () => {
  const zh = JSON.parse(read('src/locales/zh-CN/settings.json'))
  const en = JSON.parse(read('src/locales/en-US/settings.json'))
  for (const key of ['site', 'auth', 'email', 'media', 'seo', 'gateway', 'statistics', 'code', 'other']) {
    assert.equal(typeof zh.groups?.[key], 'string', `zh-CN missing group ${key}`)
    assert.equal(typeof en.groups?.[key], 'string', `en-US missing group ${key}`)
  }
  assert.notEqual(zh.fields.email_from_address, en.fields.email_from_address)
  assert.match(zh.groupDescriptions.gateway, /多个支付渠道.*同时启用/)
  assert.match(en.groupDescriptions.gateway, /Multiple payment providers/)
})
