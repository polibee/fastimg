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
  assert.match(page, /footerNavigation\.locales/)
  assert.match(page, /footerNavigation\.valueSeparator/)
  assert.doesNotMatch(page, /v-model="groupForm\.locale" \/>/)
  assert.doesNotMatch(page, /item\.target_type \+?\s*:/)
})

test('footer navigation stores separate Chinese and English labels and uses reliable edit actions', () => {
  const page = read('src/modules/footer-navigation/pages/FooterNavigationPage.vue')
  const api = read('src/modules/footer-navigation/api.ts')
  const memberFooter = read('src/modules/member/components/MemberFooterNavigation.vue')
  assert.match(api, /title_zh_cn/)
  assert.match(api, /title_en_us/)
  assert.match(api, /label_zh_cn/)
  assert.match(api, /label_en_us/)
  assert.match(page, /footerNavigation\.groupTitleZh/)
  assert.match(page, /footerNavigation\.groupTitleEn/)
  assert.match(page, /footerNavigation\.itemLabelZh/)
  assert.match(page, /footerNavigation\.itemLabelEn/)
  assert.match(page, /@click\.stop="editGroup/)
  assert.match(page, /@click\.stop="editItem/)
  assert.match(memberFooter, /watch\(locale/)
})

test('footer navigation backend persists and exposes bilingual labels', () => {
  const service = read('../backend/app/modules/footer_navigation/services/service.go')
  const group = read('../backend/app/modules/footer_navigation/models/navigation_group.go')
  const item = read('../backend/app/modules/footer_navigation/models/navigation_item.go')
  const migration = read('../backend/database/migrations/20260928000002_add_footer_navigation_bilingual_labels.go')
  assert.match(service, /TitleZhCN/)
  assert.match(service, /TitleEnUS/)
  assert.match(service, /LabelZhCN/)
  assert.match(service, /LabelEnUS/)
  assert.match(service, /Where\("id = \?", id\)\.First/)
  assert.match(group, /TitleZhCN/)
  assert.match(item, /LabelEnUS/)
  assert.match(migration, /title_zh_cn/)
  assert.match(migration, /label_en_us/)
})

test('Chinese locale does not expose English language labels in footer navigation', () => {
  const locale = JSON.parse(read('src/locales/zh-CN/footer-navigation.json'))
  assert.equal(locale.locales['en-US'], '英文')
})
