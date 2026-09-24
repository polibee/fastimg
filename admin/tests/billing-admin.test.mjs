import assert from 'node:assert/strict'
import fs from 'node:fs'

const router = fs.readFileSync(new URL('../src/router/index.ts', import.meta.url), 'utf8')
const shell = fs.readFileSync(new URL('../src/core/layouts/AdminShell.vue', import.meta.url), 'utf8')
const zh = JSON.parse(fs.readFileSync(new URL('../src/locales/zh-CN/billing.json', import.meta.url), 'utf8'))
const en = JSON.parse(fs.readFileSync(new URL('../src/locales/en-US/billing.json', import.meta.url), 'utf8'))

for (const route of ['/admin/orders', '/admin/payment-transactions', '/admin/payment-events', '/admin/refunds']) {
  assert.ok(router.includes(`path: '${route.slice('/admin/'.length)}'`), `missing admin route ${route}`)
}
assert.ok(router.includes("permission: 'admin.orders.view'"))
assert.ok(router.includes("permission: 'admin.payment_transactions.view'"))
assert.ok(router.includes("permission: 'admin.payment_events.view'"))
assert.ok(router.includes("permission: 'admin.refunds.view'"))
assert.ok(shell.includes("billing.admin.orders"), 'admin navigation is missing billing entry')

function keys(value, prefix = '') {
  return Object.entries(value).flatMap(([key, child]) => {
    const path = prefix ? `${prefix}.${key}` : key
    return child && typeof child === 'object' && !Array.isArray(child) ? keys(child, path) : [path]
  })
}
assert.deepEqual(keys(zh).sort(), keys(en).sort(), 'billing locale key sets must match')
console.log('billing-admin tests: PASS')
