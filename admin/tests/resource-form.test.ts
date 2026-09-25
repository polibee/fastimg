import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { createResourceForm, serializeResourceForm } from '../src/lib/resource-form.ts'

test('creates typed form values from resource field metadata', () => {
  const form = createResourceForm([
    { name: 'name', label: 'Name', type: 'text' },
    { name: 'age', label: 'Age', type: 'number' },
    { name: 'is_active', label: 'Active', type: 'boolean' },
    { name: 'birthday', label: 'Birthday', type: 'date' },
  ], { name: 'Ada', age: 37, is_active: 1, birthday: null })

  assert.deepEqual(form, { name: 'Ada', age: 37, is_active: true, birthday: '' })
})

test('serializes empty optional values without changing boolean semantics', () => {
  const payload = serializeResourceForm([
    { name: 'name', label: 'Name', type: 'text' },
    { name: 'age', label: 'Age', type: 'number' },
    { name: 'is_active', label: 'Active', type: 'boolean' },
  ], { name: 'Ada', age: '', is_active: false })

  assert.deepEqual(payload, { name: 'Ada', age: null, is_active: false })
})

test('serializes role fields through the same resource form contract', () => {
  const payload = serializeResourceForm([
    { name: 'name', label: 'Name', type: 'text' },
    { name: 'display_name', label: 'Display name', type: 'text' },
  ], { name: 'editor', display_name: 'Editor' })

  assert.deepEqual(payload, { name: 'editor', display_name: 'Editor' })
})

test('defaults a new user resource to active status', () => {
  const form = createResourceForm([
    { name: 'status', label: 'Status', type: 'select', options: [] },
  ])

  assert.equal(form.status, 'active')
})

test('keeps datetime-local values editable and serializes entitlements as JSON', () => {
  const form = createResourceForm([
    { name: 'starts_at', label: 'Starts at', type: 'datetime-local' },
    { name: 'entitlements_json', label: 'Entitlements', type: 'entitlements' },
  ], {
    starts_at: '2026-09-24T08:30:00Z',
    entitlements_json: '{"storage_bytes":1000,"ads_enabled":false}',
  })

  assert.equal(form.starts_at, '2026-09-24T16:30')
  assert.deepEqual(form.entitlements_json, { storage_bytes: 1000, ads_enabled: false })
  assert.deepEqual(serializeResourceForm([
    { name: 'starts_at', label: 'Starts at', type: 'datetime-local' },
    { name: 'entitlements_json', label: 'Entitlements', type: 'entitlements' },
  ], form), {
    starts_at: '2026-09-24T16:30',
    entitlements_json: '{"storage_bytes":1000,"ads_enabled":false}',
  })
})

test('entitlement switches explicitly update the resource form model', () => {
  const source = readFileSync(new URL('../src/components/resource/ResourceEntitlementsEditor.vue', import.meta.url), 'utf8')
  assert.match(source, /function setToggle\(/)
  assert.match(source, /:model-value="Boolean\(draft\.ads_enabled\)"/)
  assert.match(source, /:model-value="Boolean\(draft\.watermark_enabled\)"/)
  assert.match(source, /:aria-label="t\('resource\.entitlements\.ads_enabled'\)"/)
  assert.match(source, /resource\.entitlements\.enabled/)
  assert.match(source, /resource\.entitlements\.disabled/)
})
