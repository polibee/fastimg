import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member upload state keeps share links from immediate and polled ready responses', () => {
  const composable = read('src/modules/member/composables/useMemberUpload.ts')

  assert.match(composable, /links\?:\s*Record<string, string>/)
  assert.match(composable, /links\?:\s*Record<string, string>\s*\|\s*null/)
  const readyLinkUpdates = composable.match(/setItem\(item\.id, \{ state: 'ready', links: await resolveLinks\(result\) \}\)/g) ?? []
  assert.equal(readyLinkUpdates.length, 2, 'immediate and polled ready responses must both retain links')
  assert.match(composable, /apiFetch<\{ links\?: Record<string, string> \| null \}>\(`\/api\/v1\/media\/\$\{result\.id\}`/)
  assert.match(composable, /requiredLinkKeys\.some/)
})

test('member upload panel renders localized copyable link formats after success', () => {
  const panel = read('src/modules/member/components/MemberUploadPanel.vue')

  assert.match(panel, /const uploadLinkKeys\s*=\s*\['url',\s*'markdown',\s*'html',\s*'bbcode'\]/)
  assert.match(panel, /item\.links\.url/)
  assert.match(panel, /member\.upload\.previewAlt/)
  assert.match(panel, /member\.upload\.openImage/)
  assert.match(panel, /navigator\.clipboard\.writeText/)
  assert.match(panel, /item\.links\[key\]/)
  assert.match(panel, /member\.upload\.linksTitle/)
  assert.match(panel, /member\.upload\.copyLink/)
  assert.match(panel, /<Button[^>]*type="button"[^>]*@click="copyLink/)
  assert.doesNotMatch(panel, /<Input[^>]*:model-value="item\.links\[key\]"/)
})

test('upload link labels and copy feedback are synchronized in both locales', () => {
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    assert.equal(typeof messages.upload?.linksTitle, 'string')
    assert.equal(typeof messages.upload?.previewAlt, 'string')
    assert.equal(typeof messages.upload?.openImage, 'string')
    assert.equal(typeof messages.upload?.copyLink, 'string')
    assert.equal(typeof messages.upload?.copied, 'string')
    for (const key of ['url', 'markdown', 'html', 'bbcode']) {
      assert.equal(typeof messages.upload?.links?.[key], 'string')
    }
  }
})
