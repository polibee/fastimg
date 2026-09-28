import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

test('production login does not expose demo credentials', () => {
  const loginPage = fs.readFileSync(new URL('../src/modules/auth/pages/LoginPage.vue', import.meta.url), 'utf8')
  assert.doesNotMatch(loginPage, /DEMO_CREDENTIALS|admin@example\.com|Admin123!|demoTitle|fillAndLogin/)
  assert.match(loginPage, /const email = ref\(''\)/)
  assert.match(loginPage, /const password = ref\(''\)/)
})
