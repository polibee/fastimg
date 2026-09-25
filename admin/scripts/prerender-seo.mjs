import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import publicPages from '../src/apps/member/public-pages.json' with { type: 'json' }

const root = dirname(fileURLToPath(import.meta.url))
const dist = join(root, '..', 'dist')
const template = await readFile(join(dist, 'index.html'), 'utf8')
const origin = process.env.SSG_PUBLIC_ORIGIN || 'https://img.example.com'

const pages = publicPages

for (const page of pages) {
  const canonical = `${origin.replace(/\/$/, '')}${page.path}`
  const head = `<meta name="description" content="${page.description}"><meta property="og:title" content="${page.title}"><meta property="og:description" content="${page.description}"><meta property="og:type" content="website"><meta property="og:url" content="${canonical}"><link rel="canonical" href="${canonical}"><script type="application/ld+json">${JSON.stringify({ '@context': 'https://schema.org', '@type': 'WebSite', name: 'FastImg', url: origin })}</script>`
  const content = `<main data-fastimg-ssg="true"><h1>${page.heading}</h1><p>${page.body}</p><p><a href="/login">Sign in to upload</a> · <a href="/plans">View plans</a></p></main>`
  const output = template.replace('</head>', `${head}</head>`).replace('<div id="app"></div>', content)
  const target = page.path === '/' ? join(dist, 'index.html') : join(dist, page.path.slice(1), 'index.html')
  await mkdir(dirname(target), { recursive: true })
  await writeFile(target, output)
}

console.log(`Generated SEO entry pages: ${pages.map((page) => page.path).join(', ')}`)
