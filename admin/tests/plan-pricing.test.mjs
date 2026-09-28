import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

const read = (path) => fs.readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')

test('plan editor embeds price version management instead of a duplicate price menu', () => {
  const form = read('src/components/resource/ResourceFormView.vue')
  const pricing = read('src/modules/plans/components/PlanPricingEditor.vue')
  assert.match(form, /PlanPricingEditor/)
  assert.match(form, /editing && isPlan/)
  assert.match(pricing, /\/api\/v1\/admin\/plans\/.*\/prices/)
  assert.match(pricing, /amount_minor: Math\.round\(amount \* 100\)/)
  assert.match(pricing, /price\.status === 'active'/)
})

test('plan pricing and image-processing labels stay synchronized in both locales', () => {
  const zh = JSON.parse(read('src/locales/zh-CN/plans.json'))
  const en = JSON.parse(read('src/locales/en-US/plans.json'))
  const zhResource = JSON.parse(read('src/locales/zh-CN/resource.json'))
  const enResource = JSON.parse(read('src/locales/en-US/resource.json'))
  assert.deepEqual(Object.keys(zh.pricing).sort(), Object.keys(en.pricing).sort())
  assert.equal(zhResource.entitlements.transform_count, '每月图片处理次数')
  assert.equal(enResource.entitlements.transform_count, 'Monthly image processing operations')
})
