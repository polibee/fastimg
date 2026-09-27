import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const [routes, api, page, openapi] = await Promise.all([
  readFile(new URL('../src/apps/admin/routes.ts', import.meta.url), 'utf8'),
  readFile(new URL('../src/generated/api.ts', import.meta.url), 'utf8'),
  readFile(new URL('../src/modules/tasks/pages/AdminTasksPage.vue', import.meta.url), 'utf8').catch(() => ''),
  readFile(new URL('../../backend/app/openapi/spec.go', import.meta.url), 'utf8'),
])

test('admin task center has a permission-gated route and retry API', () => {
  assert.match(routes, /path: 'tasks'/)
  assert.match(routes, /permission: 'admin\.tasks\.view'/)
  assert.match(api, /failedTasks\(/)
  assert.match(api, /retryFailedTask\(/)
  assert.match(api, /apiFetch<\{ uuid: string; status: 'queued' \}>\('/)
  assert.match(page, /admin\.tasks\.retry/)
  assert.match(openapi, /admin\/tasks/)
})
