import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

test('member reports route and navigation expose own report status', () => {
  const routes = readFileSync(new URL('../src/apps/member/route-private.ts', import.meta.url), 'utf8')
  const shell = readFileSync(new URL('../src/core/layouts/MemberShell.vue', import.meta.url), 'utf8')
  const page = readFileSync(new URL('../src/modules/member/pages/MemberReportsPage.vue', import.meta.url), 'utf8')

  assert.match(routes, /path:\s*'reports'/)
  assert.match(shell, /to="\/reports"/)
  assert.match(page, /\/api\/v1\/me\/reports/)
  assert.match(page, /member\.reports\.status/)
})

test('admin report resource requires the moderation resolve action', () => {
  const manifest = readFileSync(new URL('../../backend/app/modules/moderation/resource/manifest.go', import.meta.url), 'utf8')
  const handlerRegistry = readFileSync(new URL('../../backend/app/modules/admin/actions/registry.go', import.meta.url), 'utf8')

  assert.match(manifest, /report-resolve/)
  assert.match(manifest, /hide_media/)
  assert.match(handlerRegistry, /NewResolveHandler/)
})
