import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

const read = (file) => fs.readFileSync(new URL(`../${file}`, import.meta.url), 'utf8')

test('admin surface exposes finance, settings and all-user media management', () => {
  const router = read('src/router/index.ts')
  const shell = read('src/core/layouts/AdminShell.vue')
  const generated = read('src/core/resource/generated.ts')
  const media = read('src/modules/media/resource.ts')
  assert.match(router, /path: 'settings'/)
  assert.match(router, /path: 'statistics'/)
  assert.match(shell, /admin\.settings\.manage/)
  assert.match(generated, /resourceDefinition as mediaResource/)
  assert.match(media, /admin\.media\.view/)
})
