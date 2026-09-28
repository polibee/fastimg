import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('admin content page routes are permission gated and provide list/create/edit surfaces', () => {
  const routes = read('src/apps/admin/routes.ts')
  assert.match(routes, /path:\s*'content-pages',\s*name:\s*'admin-content-pages'[\s\S]*?admin\.content_pages\.view/)
  assert.match(routes, /path:\s*'content-pages\/new',\s*name:\s*'admin-content-page-new'[\s\S]*?admin\.content_pages\.manage/)
  assert.match(routes, /path:\s*'content-pages\/:id\/edit',\s*name:\s*'admin-content-page-edit'[\s\S]*?admin\.content_pages\.manage/)
  assert.match(read('src/modules/content/api.ts'), /content-pages/)
  assert.match(read('src/modules/content/pages/ContentPagesListPage.vue'), /contentApi\.list/)
  assert.match(read('src/modules/content/pages/ContentPageFormPage.vue'), /contentApi\.(create|update)/)
})

test('content form sends Tiptap JSON and never exposes raw HTML editing', () => {
  const form = read('src/modules/content/pages/ContentPageFormPage.vue')
  const editor = read('src/modules/content/components/RichTextEditor.vue')
  assert.match(form, /content_json/)
  assert.match(form, /RichTextEditor/)
  assert.match(editor, /@tiptap\/vue-3/)
  assert.match(editor, /EditorContent/)
  assert.doesNotMatch(editor, /<textarea[\s\S]*html/i)
  assert.doesNotMatch(form, /v-html|content_html|rawHtml/i)
})
