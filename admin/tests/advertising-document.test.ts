import assert from 'node:assert/strict'
import test from 'node:test'
import { buildSandboxedAdDocument } from '../src/modules/advertising/sandbox.ts'

test('preserves arbitrary third-party HTML ad embeds', () => {
  const embed = '<script src="https://ads.example.test/widget.js"></script><object data="https://ads.example.test/slot"></object>'

  assert.equal(buildSandboxedAdDocument(embed), embed)
})

test('wraps plain JavaScript in an isolated HTML document', () => {
  const document = buildSandboxedAdDocument("console.log('ad')")

  assert.match(document, /<script>console\.log\('ad'\)<\/script>/)
})
