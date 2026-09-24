import assert from 'node:assert/strict'
import fs from 'node:fs'

const router = fs.readFileSync(new URL('../src/router/index.ts', import.meta.url), 'utf8')
const plans = fs.readFileSync(new URL('../src/modules/member/pages/MemberPlansPage.vue', import.meta.url), 'utf8')
const billingApi = fs.readFileSync(new URL('../src/modules/billing/api.ts', import.meta.url), 'utf8')
const memberShell = fs.readFileSync(new URL('../src/core/layouts/MemberShell.vue', import.meta.url), 'utf8')
const checkout = fs.readFileSync(new URL('../src/modules/member/pages/MemberCheckoutPage.vue', import.meta.url), 'utf8')

assert.ok(router.includes("path: 'checkout/:orderId'"), 'member checkout route is missing')
assert.ok(router.includes("path: 'orders'"), 'member orders route is missing')
assert.ok(router.includes("path: 'orders/:id'"), 'member order detail route is missing')
assert.ok(router.includes("name: 'member-checkout'"))
assert.ok(router.includes("name: 'member-orders'"))
assert.ok(router.includes("name: 'member-order-detail'"))
assert.ok(billingApi.includes("/api/v1/orders"), 'billing API must start server-priced checkout')
assert.ok(billingApi.includes('/cancel'), 'billing API must support canceling unpaid orders')
assert.ok(memberShell.includes("member.nav.orders"), 'member navigation must expose own orders')
assert.ok(checkout.includes('member.billing.payWithFake'), 'checkout must expose the development gateway action')
assert.ok(checkout.includes('cancelMemberOrder'), 'checkout must allow canceling unpaid orders')
assert.ok(plans.includes('free'), 'plans page must retain a Free branch')
console.log('billing-ui tests: PASS')
