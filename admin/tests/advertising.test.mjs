import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('advertising resource exposes header, footer, left and right placements', () => {
  const resource = read('src/modules/advertising/resource.js')
  for (const placement of ['header', 'footer', 'left', 'right']) {
    assert.match(resource, new RegExp(`value: "${placement}"`))
  }
  for (const legacyPlacement of ['upload', 'dashboard', 'discovery']) {
    assert.doesNotMatch(resource, new RegExp(`value: "${legacyPlacement}"`))
  }
})

test('advertising placement labels stay synchronized in both locales', () => {
  const zh = JSON.parse(read('src/locales/zh-CN/resource.json'))
  const en = JSON.parse(read('src/locales/en-US/resource.json'))
  for (const placement of ['header', 'footer', 'left', 'right']) {
    assert.equal(typeof zh.options.placement[placement], 'string')
    assert.equal(typeof en.options.placement[placement], 'string')
  }
})

test('advertising content is an admin-managed text, image, or script field', () => {
  const resource = read('src/modules/advertising/resource.js')
  for (const creativeType of ['text', 'image', 'script']) {
    assert.match(resource, new RegExp(`value: "${creativeType}"`))
  }
  assert.match(resource, /name: "creative_content".*type: "textarea"/)
  assert.doesNotMatch(resource, /name: "creative_url"/)
})

test('advertising content uses the shared textarea control with visible guidance', () => {
  const form = read('src/components/resource/ResourceFormView.vue')
  assert.match(form, /isTextarea\(field\)/)
  assert.match(form, /<Textarea/)
  assert.match(form, /fieldHint\(field\)/)
})
