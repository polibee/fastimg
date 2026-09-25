import assert from 'node:assert/strict'
import { readFileSync, readdirSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const localesRoot = path.join(adminRoot, 'src', 'locales')

function localeFiles(locale, directory = path.join(localesRoot, locale), prefix = '') {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const relativePath = path.join(prefix, entry.name)
    const absolutePath = path.join(directory, entry.name)
    return entry.isDirectory()
      ? localeFiles(locale, absolutePath, relativePath)
      : entry.name.endsWith('.json') ? [relativePath] : []
  })
}

function keyPaths(value, prefix = '') {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return [prefix]
  return Object.entries(value).flatMap(([key, child]) => keyPaths(child, prefix ? `${prefix}.${key}` : key))
}

test('zh-CN and en-US locale JSON files and nested keys stay in sync', () => {
  const chineseFiles = localeFiles('zh-CN').sort()
  const englishFiles = localeFiles('en-US').sort()
  assert.deepEqual(chineseFiles, englishFiles, 'locale JSON file paths must exist for both supported languages')

  for (const file of chineseFiles) {
    const chinese = JSON.parse(readFileSync(path.join(localesRoot, 'zh-CN', file), 'utf8'))
    const english = JSON.parse(readFileSync(path.join(localesRoot, 'en-US', file), 'utf8'))
    assert.deepEqual(keyPaths(chinese).sort(), keyPaths(english).sort(), `${file} must have identical nested keys in both locales`)
  }
})

test('media library uses the framework locale namespace instead of inline bilingual copy', () => {
  const page = readFileSync(path.join(adminRoot, 'src', 'modules', 'media', 'pages', 'MediaLibraryPage.vue'), 'utf8')
  assert.match(page, /const\s+\{\s*t\s*\}\s*=\s*useI18n\(\)/)
  assert.doesNotMatch(page, /function\s+t\(zh:\s*string,\s*en:\s*string\)/)
  assert.doesNotMatch(page, /\bt\(\s*['"][^'"]*['"]\s*,\s*['"]/)
})

test('admin and member navigation labels use registered locale keys', () => {
  const shell = readFileSync(path.join(adminRoot, 'src', 'core', 'layouts', 'AdminShell.vue'), 'utf8')
  const memberShell = readFileSync(path.join(adminRoot, 'src', 'core', 'layouts', 'MemberShell.vue'), 'utf8')
  assert.match(shell, /:tooltip="t\('auth\.dashboard'\)"/)
  assert.match(shell, /<span>\{\{ t\('auth\.dashboard'\) \}\}<\/span>/)
  assert.match(memberShell, /member\.home\.title/)
  assert.match(memberShell, /member\.media\.title/)
  assert.doesNotMatch(shell + memberShell, /locale(?:\.value)? === 'zh-CN' \? '(?:媒体库|首页)' : '(?:Media library|Home)'/)
})

test('FastImg generic resources have Chinese and English labels for fields and options', () => {
  const chinese = JSON.parse(readFileSync(path.join(localesRoot, 'zh-CN', 'resource.json'), 'utf8'))
  const english = JSON.parse(readFileSync(path.join(localesRoot, 'en-US', 'resource.json'), 'utf8'))
  const paths = [
    ['labels.folders', 'folders'],
    ['labels.albums', 'albums'],
    ['labels.advertising', 'advertising'],
    ['labels.api_tokens', 'API token resource'],
    ['fields.folders.name', 'folder name'],
    ['fields.folders.parent_id', 'parent folder'],
    ['fields.albums.visibility', 'album visibility'],
    ['fields.advertising.placement', 'ad placement'],
    ['fields.advertising.creative_url', 'creative URL'],
    ['fields.api_tokens.token_prefix', 'token prefix'],
    ['options.visibility.private', 'private visibility'],
    ['options.placement.header', 'header ad placement'],
    ['options.status.revoked', 'revoked status'],
  ]

  for (const [path, description] of paths) {
    const lookup = (message) => path.split('.').reduce((value, key) => value?.[key], message)
    assert.equal(typeof lookup(chinese), 'string', `zh-CN ${description} translation must exist`)
    assert.equal(typeof lookup(english), 'string', `en-US ${description} translation must exist`)
    assert.notEqual(lookup(chinese), lookup(english), `${description} must be localized rather than shared fallback text`)
  }
})
