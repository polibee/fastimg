import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'

const root = new URL('../', import.meta.url)

async function source(path) {
  return readFile(new URL(path, root), 'utf8')
}

test('member high-cardinality pages keep server pagination in the UI', async () => {
  const files = [
    'src/modules/member/pages/MemberReportsPage.vue',
    'src/modules/member/pages/MemberOrdersPage.vue',
    'src/modules/member/pages/MemberShareLinksPage.vue',
    'src/modules/member/pages/MemberTokensPage.vue',
  ]
  for (const file of files) {
    const text = await source(file)
    assert.match(text, /Pagination/)
    assert.match(text, /page=/)
    assert.match(text, /per_page/)
  }
})

test('admin backup and friend-link review lists expose pagination metadata', async () => {
  const backupController = await source('../backend/app/modules/backups/controllers/backup_controller.go')
  const backupService = await source('../backend/app/services/backups/backup_service.go')
  const friendController = await source('../backend/app/modules/friend_links/controllers/controller.go')
  const friendService = await source('../backend/app/modules/friend_links/services/service.go')
  assert.match(backupController, /Query\("page"/)
  assert.match(backupController, /"last_page"/)
  assert.match(backupService, /Paginate\(page, perPage/)
  assert.match(friendController, /ListPaginated/)
  assert.match(friendController, /"last_page"/)
  assert.match(friendService, /Paginate\(page, perPage/)
})

test('member exports and admin album media expose server pagination', async () => {
  const exportsPage = await source('src/modules/member/pages/MemberExportsPage.vue')
  const exportController = await source('../backend/app/modules/media/controllers/export_controller.go')
  const exportService = await source('../backend/app/services/media/export_service.go')
  const albumPage = await source('src/modules/albums/pages/AdminAlbumMediaPage.vue')
  const albumController = await source('../backend/app/modules/albums/controllers/admin_media_controller.go')
  const collectionService = await source('../backend/app/services/collections/collection_service.go')
  assert.match(exportsPage, /Pagination/)
  assert.match(exportsPage, /me\/exports\?page=/)
  assert.match(exportController, /QueryInt\("page"/)
  assert.match(exportController, /"last_page"/)
  assert.match(exportService, /Paginate\(page, perPage/)
  assert.match(albumPage, /Pagination/)
  assert.match(albumPage, /admin\/albums\/\$\{albumID\}\/media\?page=/)
  assert.match(albumController, /QueryInt\("page"/)
  assert.match(albumController, /"last_page"/)
  assert.match(collectionService, /ListAdminAlbumMediaPage/)
})
