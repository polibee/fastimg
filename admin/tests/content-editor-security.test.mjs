import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('the content editor only uses the structured Tiptap document contract', () => {
  const editor = read('src/modules/content/components/RichTextEditor.vue')
  assert.match(editor, /JSONContent/)
  assert.match(editor, /StarterKit/)
  assert.match(editor, /Link/)
  assert.match(editor, /Image/)
  assert.doesNotMatch(editor, /contenteditable=["']false/i)
})
