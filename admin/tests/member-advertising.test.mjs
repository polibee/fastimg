import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member shell mounts admin-configured advertising slots in compact card shells', () => {
  const shell = read('src/core/layouts/MemberShell.vue')
  for (const placement of ['header', 'footer', 'left', 'right']) {
    assert.match(shell, new RegExp(`<MemberAdSlot placement="${placement}"`))
  }
  assert.match(read('src/modules/advertising/components/MemberAdSlot.vue'), /api\/v1\/ads\?placement=/)
  assert.doesNotMatch(read('src/modules/advertising/components/MemberAdSlot.vue'), /Card(Content)?/)
})

test('member advertising supports text, image and sandboxed javascript only', () => {
  const component = read('src/modules/advertising/components/MemberAdSlot.vue')
  for (const creativeType of ['text', 'image', 'script']) {
    assert.match(component, new RegExp(`creative_type.*${creativeType}|${creativeType}.*creative_type`, 's'))
  }
  assert.match(component, /sandbox="allow-scripts"/)
  assert.match(component, /:srcdoc=/)
  assert.match(component, /buildSandboxedAdDocument/)
  assert.match(component, /rounded-xl/)
  assert.match(component, /scrolling="no"/)
  assert.match(component, /max-w-full/)
  assert.doesNotMatch(component, /v-html/)
})

test('member advertising locale keys are synchronized', () => {
  const zh = JSON.parse(read('src/locales/zh-CN/member.json'))
  const en = JSON.parse(read('src/locales/en-US/member.json'))
  for (const key of ['label', 'imageAlt', 'loadFailed']) {
    assert.equal(typeof zh.ads[key], 'string')
    assert.equal(typeof en.ads[key], 'string')
  }
})
