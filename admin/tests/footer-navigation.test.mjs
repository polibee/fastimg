import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('footer navigation has a public endpoint and permission-gated admin route', () => {
  const web = read('../backend/routes/web.go')
  const routes = read('src/apps/admin/routes.ts')
  assert.match(web, /site\/footer-navigation/)
  assert.match(web, /admin\/footer-navigation/)
  assert.match(routes, /path:\s*'footer-navigation',\s*name:\s*'admin-footer-navigation'[\s\S]*?admin\.footer_navigation\.view/)
  assert.match(read('src/modules/footer-navigation/api.ts'), /footer-navigation/)
  assert.match(read('src/modules/footer-navigation/pages/FooterNavigationPage.vue'), /footerApi/)
})

test('footer navigation supports group, item, and parent item editing', () => {
  const page = read('src/modules/footer-navigation/pages/FooterNavigationPage.vue')
  assert.match(page, /group_id/)
  assert.match(page, /parent_id/)
  assert.match(page, /target_type/)
  assert.match(page, /is_enabled/)
})
