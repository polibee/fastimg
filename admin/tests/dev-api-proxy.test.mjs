import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const viteConfig = readFileSync(path.join(adminRoot, 'vite.config.ts'), 'utf8')
const isolatedViteConfig = readFileSync(path.join(adminRoot, 'vite.fastimg-isolated.config.mjs'), 'utf8')
const apiClient = readFileSync(path.join(adminRoot, 'src', 'lib', 'api.ts'), 'utf8')

test('Vite forwards API requests to the separately configured backend', () => {
  assert.match(viteConfig, /proxy:\s*\{\s*['"]\/api['"]:\s*\{/s)
  assert.match(viteConfig, /const backendTarget = process\.env\.FASTIMG_BACKEND_URL\s*\?\?\s*['"]http:\/\/127\.0\.0\.1:53085['"]/)
  assert.match(viteConfig, /changeOrigin:\s*true/)
})

test('Vite forwards machine-readable SEO files to the backend', () => {
  assert.match(viteConfig, /['"]\/sitemap\.xml['"]:\s*\{\s*target:\s*backendTarget/)
  assert.match(viteConfig, /['"]\/robots\.txt['"]:\s*\{\s*target:\s*backendTarget/)
})

test('Vite refuses to silently share its configured frontend port', () => {
  assert.match(viteConfig, /port:\s*5180,\s*strictPort:\s*true/)
})

test('FastImg isolated Vite config keeps Tailwind utility generation enabled', () => {
  assert.match(isolatedViteConfig, /@tailwindcss\/vite/)
  assert.match(isolatedViteConfig, /tailwindcss\(\)/)
})

test('browser API calls stay same-origin by default and can be explicitly overridden', () => {
  assert.match(apiClient, /VITE_API_BASE_URL/)
  assert.match(apiClient, /previewPorts\s*=\s*new Set\(\['53083',\s*'53084'\]\)/)
  assert.match(apiClient, /previewPorts\.has\(globalThis\.location\?\.port\s*\?\?\s*''\)/)
  assert.match(apiClient, /VITE_API_BASE_URL\s*\?\?/)
})
