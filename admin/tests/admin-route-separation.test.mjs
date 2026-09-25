import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (file) => readFileSync(path.join(root, file), 'utf8')

test('admin navigation stays under the admin shell and excludes member-owned media', () => {
  const shell = read('src/core/layouts/AdminShell.vue')
  assert.match(shell, /to="\/admin"/)
  assert.match(shell, /to="\/admin\/rbac"/)
  assert.match(shell, /to="\/admin\/audit-logs"/)
  assert.doesNotMatch(shell, /to="\/(media|folders|albums)"/)
  assert.match(shell, /data_scope\s*!==\s*'own'/)
})

test('admin shell provides a localized shortcut to the member homepage', () => {
  const shell = read('src/core/layouts/AdminShell.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/core.json'))
  const english = JSON.parse(read('src/locales/en-US/core.json'))
  assert.match(shell, /RouterLink to="\/"[\s\S]*?core\.memberFrontend/)
  assert.equal(typeof chinese.memberFrontend, 'string')
  assert.equal(typeof english.memberFrontend, 'string')
  assert.notEqual(chinese.memberFrontend, english.memberFrontend)
})

test('admin global search normalizes only registered all-scope resources', () => {
  const shell = read('src/core/layouts/AdminShell.vue')
  assert.match(shell, /adminResourcePath\(result\.route, result\.resource\)/)
  assert.match(shell, /data_scope\s*!==\s*'own'/)
})

test('admin resource registry filtering excludes own-scope routes in both naming forms', () => {
  const router = read('src/router/index.ts')
  assert.match(router, /dataScope/)
  assert.match(router, /data_scope === 'own'/)
})
