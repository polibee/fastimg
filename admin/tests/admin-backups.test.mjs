import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const root = path.resolve(import.meta.dirname, '..')
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8')

test('admin backup route and menu are permission gated', () => {
  const routes = read('src/apps/admin/routes.ts')
  const shell = read('src/core/layouts/AdminShell.vue')
  assert.match(routes, /path: 'backups'/)
  assert.match(routes, /admin\.backups\.manage/)
  assert.match(shell, /admin-backups/)
})

test('backup UI exposes validation confirmation and bilingual locales', () => {
  const page = read('src/modules/backups/pages/AdminBackupsPage.vue')
  const zh = JSON.parse(read('src/locales/zh-CN/backups.json'))
  const en = JSON.parse(read('src/locales/en-US/backups.json'))
  assert.match(page, /RESTORE_FASTIMG_BACKUP/)
  assert.match(page, /FormData/)
  assert.deepEqual(Object.keys(zh), Object.keys(en))
})

test('backup UI polls persisted jobs and displays classified failures', () => {
  const page = read('src/modules/backups/pages/AdminBackupsPage.vue')
  assert.match(page, /schedulePoll/)
  assert.match(page, /job\.status === 'failed'/)
  assert.match(page, /jobError\(job\)/)
})
