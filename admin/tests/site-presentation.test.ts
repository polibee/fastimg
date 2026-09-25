import assert from 'node:assert/strict'
import test from 'node:test'
import { builtInFallbackImageURL, resolveSafeFallbackImageURL } from '../src/lib/site-presentation.ts'

test('member image fallback accepts only HTTP(S) or site-absolute URLs', () => {
  assert.equal(resolveSafeFallbackImageURL('https://img.example.com/fallback.svg'), 'https://img.example.com/fallback.svg')
  assert.equal(resolveSafeFallbackImageURL('/public/fallback.svg'), '/public/fallback.svg')
  assert.equal(resolveSafeFallbackImageURL('javascript:alert(1)'), builtInFallbackImageURL)
  assert.equal(resolveSafeFallbackImageURL(''), builtInFallbackImageURL)
})
