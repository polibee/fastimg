import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (file) => readFileSync(path.join(root, file), 'utf8')

test('admin media access logs stay in the admin surface with explicit permission', () => {
  const router = read('src/apps/admin/routes.ts')
  const shell = read('src/core/layouts/AdminShell.vue')
  const page = read('src/modules/access/pages/MediaAccessLogPage.vue')
  assert.match(router, /path: 'media-access-logs'/)
  assert.match(router, /permission: 'admin\.media_access_logs\.view'/)
  assert.match(shell, /to="\/admin\/media-access-logs"/)
  assert.match(shell, /admin\.media_access_logs\.view/)
  assert.match(page, /mediaAccessLogs\(/)
  assert.match(page, /auth\.mediaAccessLogs/)
})

test('access log UI uses the redacted contract and exposes bilingual labels', () => {
  const page = read('src/modules/access/pages/MediaAccessLogPage.vue')
  const api = read('src/generated/api.ts')
  const zh = JSON.parse(read('src/locales/zh-CN/auth.json'))
  const en = JSON.parse(read('src/locales/en-US/auth.json'))
  assert.match(api, /\/api\/v1\/admin\/media-access-logs/)
  assert.doesNotMatch(page, /entry\.(signature|password|token)/i)
  for (const key of ['mediaAccessLogs', 'mediaAccessDescription', 'mediaAccessResult', 'mediaAccessMode']) {
    assert.equal(typeof zh[key], 'string')
    assert.equal(typeof en[key], 'string')
    assert.notEqual(zh[key], en[key])
  }
})
