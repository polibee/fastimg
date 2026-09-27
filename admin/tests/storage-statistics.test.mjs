import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const [page, api, storageApi, settings] = await Promise.all([
  readFile(new URL('../src/modules/settings/pages/AdminStoragePage.vue', import.meta.url), 'utf8'),
  readFile(new URL('../src/generated/api.ts', import.meta.url), 'utf8'),
  readFile(new URL('../src/modules/settings/storage-api.ts', import.meta.url), 'utf8'),
  readFile(new URL('../src/modules/settings/pages/AdminSettingsPage.vue', import.meta.url), 'utf8'),
])

test('enabled object storage connections expose real per-provider statistics', () => {
  assert.match(storageApi, /\/api\/v1\/admin\/storage\/statistics/)
  assert.match(api, /storage\/statistics/)
  assert.match(page, /storageStatistics\(/)
  assert.match(page, /object_count/)
  assert.match(page, /estimated_bandwidth_bytes/)
  assert.match(storageApi, /object_status_counts/)
  assert.match(storageApi, /variant_counts/)
  assert.match(storageApi, /daily/)
  assert.match(storageApi, /health/)
  assert.match(storageApi, /orphan_object_count/)
  assert.match(page, /\.daily/)
  assert.match(page, /\.health/)
  assert.match(page, /orphan_object_count/)
})

test('object storage settings are a separate section before general settings', () => {
  assert.ok(settings.indexOf('<StorageSettingsSection v-if="!loading"') < settings.indexOf('<form v-if="!loading"'))
})
