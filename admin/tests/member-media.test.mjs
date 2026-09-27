import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const adminRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = (relativePath) => readFileSync(path.join(adminRoot, relativePath), 'utf8')

test('member media route is nested under the dedicated MemberShell at a flat URL', () => {
  const router = read('src/router/index.ts')
  const memberRoutes = read('src/apps/member/routes.ts')
  const memberRouteParts = read('src/apps/member/route-parts.ts')
  const memberRoutePublic = read('src/apps/member/route-public.ts')
  const memberRoutePrivate = read('src/apps/member/route-private.ts')
  assert.match(router, /memberRoutes/)
  assert.match(memberRoutes, /path:\s*'\/'[\s\S]*?MemberShell\.vue/)
  assert.match(`${memberRoutePublic}\n${memberRoutePrivate}`, /path:\s*['"]media['"],\s*name:\s*['"]member-media['"][\s\S]*?MemberMediaPage/)
  assert.match(`${memberRoutePublic}\n${memberRoutePrivate}`, /path:\s*['"]media\/:id['"],\s*name:\s*['"]member-media-detail['"][\s\S]*?MemberMediaDetailPage/)
})

test('member media details use the authenticated owner-scoped API and private content previews', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  const library = read('src/modules/member/pages/MemberMediaPage.vue')
  assert.match(page, /\/api\/v1\/media\/\$\{encodeURIComponent/)
  assert.match(page, /apiFetchBlob\(item\.value!\.variants\[name\]\.url, auth\.token!\)/)
  assert.match(library, /name:\s*'member-media-detail'/)
  assert.doesNotMatch(page, /publicUrl|shareUrl|developer/)
  assert.match(page, /member\.media\.linkFormats/)
  assert.match(page, /navigator\.clipboard\.writeText/)
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    assert.equal(typeof messages.media.details, 'string')
    assert.equal(typeof messages.media.errors.detailLoadFailed, 'string')
  }
})

test('member media link formats stay explicit and localized', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /item\.links\[key\]/)
  for (const locale of [chinese, english]) {
    assert.equal(typeof locale.media?.linkFormats, 'string')
    assert.equal(typeof locale.media?.copyLink, 'string')
    assert.equal(typeof locale.media?.copied, 'string')
    assert.equal(typeof locale.media?.links?.url, 'string')
    assert.equal(typeof locale.media?.links?.markdown, 'string')
    assert.equal(typeof locale.media?.links?.html, 'string')
    assert.equal(typeof locale.media?.links?.bbcode, 'string')
  }
})

test('admin does not expose the own-scope media library as an all-user manager', () => {
  const shell = read('src/core/layouts/MemberShell.vue')
  const router = read('src/router/index.ts')
  assert.match(shell, /to="\/media"[\s\S]*?member\.media\.title/)
  assert.doesNotMatch(router, /name:\s*'media-library'/)
})

test('member media operations use authenticated own-scope endpoints without accepting user_id', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  const uploader = read('src/modules/member/composables/useMemberUpload.ts')
  assert.match(page, /\/api\/v1\/media/)
  assert.match(uploader, /\/api\/v1\/uploads/)
  assert.match(page, /\/restore/)
  assert.match(page, /method:\s*'DELETE'/)
  assert.match(page, /useAuthStore\(\)/)
  assert.match(uploader, /useAuthStore\(\)/)
  assert.doesNotMatch(page, /user_id\s*[:=]/)
  assert.doesNotMatch(uploader, /user_id\s*[:=]/)
  assert.match(page, /\/api\/v1\/media\/\$\{item\.id\}\/folder/)
  assert.match(page, /folder_id: folderID/)
})

test('member trash exposes a confirmed empty action and keeps the trash surface styled', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /AlertDialog/)
  assert.match(page, /\/api\/v1\/media\/trash/)
  assert.match(page, /empty-trash/)
  assert.match(page, /member\.media\.emptyTrash/)
  assert.match(page, /class="flex flex-col gap-7"/)
  for (const locale of [chinese, english]) {
    assert.equal(typeof locale.media?.emptyTrash, 'string')
    assert.equal(typeof locale.media?.emptyTrashTitle, 'string')
    assert.equal(typeof locale.media?.emptyTrashDescription, 'string')
    assert.equal(typeof locale.media?.emptyTrashSuccess, 'string')
  }
})

test('member media library uses compact single-image cards', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  assert.match(page, /h-32 rounded-lg sm:h-36/)
  assert.match(page, /class="gap-0 overflow-hidden py-0"/)
  assert.match(page, /CardHeader class="gap-1 p-2"/)
  assert.match(page, /CardContent class="flex flex-col gap-1 px-2 pb-2"/)
  assert.doesNotMatch(page, /aspect-\[4\/3\]/)
})

test('member media user-facing copy comes from synchronized member locale keys', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /useI18n\(\)/)
  assert.match(page, /t\('member\.media\./)
  assert.equal(typeof chinese.media?.title, 'string')
  assert.equal(typeof english.media?.title, 'string')
  assert.doesNotMatch(page, /function\s+t\(zh:/)
})

test('member media library exposes server pagination and localized navigation labels', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /const currentPage = ref\(1\)/)
  assert.match(page, /page: String\(currentPage\.value\)/)
  assert.match(page, /const pageCount = computed\(/)
  assert.match(page, /currentPage\.value\s*[<>=]+\s*pageCount\.value/)
  assert.match(page, /member\.media\.previousPage/)
  assert.match(page, /member\.media\.nextPage/)
  assert.equal(typeof chinese.media?.previousPage, 'string')
  assert.equal(typeof english.media?.previousPage, 'string')
  assert.equal(typeof chinese.media?.nextPage, 'string')
  assert.equal(typeof english.media?.nextPage, 'string')
})

test('member media folder controls use synchronized locale keys', () => {
  const page = read('src/modules/member/pages/MemberMediaPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /member\.media\.folder/)
  assert.match(page, /member\.media\.unfiled/)
  assert.equal(typeof chinese.media?.folder, 'string')
  assert.equal(typeof english.media?.folder, 'string')
  assert.equal(typeof chinese.media?.errors?.moveToFolderFailed, 'string')
  assert.equal(typeof english.media?.errors?.moveToFolderFailed, 'string')
})

test('member media share controls use owner-scoped share-link APIs', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  const sharePage = read('src/modules/member/pages/MemberShareLinksPage.vue')
  assert.match(page, /\/api\/v1\/media\/\$\{item\.value\.id\}\/share-links/)
  assert.match(page, /expires_at: expiresAt/)
  assert.match(sharePage, /\/api\/v1\/share-links/)
  assert.match(sharePage, /method:\s*'DELETE'/)
})

test('member media share creation supports an optional password', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  assert.match(page, /sharePassword/)
  assert.match(page, /password:\s*sharePassword/)
  assert.match(page, /type="password"/)
  for (const locale of ['zh-CN', 'en-US']) {
    const messages = JSON.parse(read(`src/locales/${locale}/member.json`))
    assert.equal(typeof messages.media.sharePassword, 'string')
    assert.equal(typeof messages.media.sharePasswordHint, 'string')
  }
})

test('member media details expose signed URL and hotlink policy controls', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  const chinese = JSON.parse(read('src/locales/zh-CN/member.json'))
  const english = JSON.parse(read('src/locales/en-US/member.json'))
  assert.match(page, /\/api\/v1\/media\/\$\{item\.value\.id\}\/signed-url/)
  assert.match(page, /\/hotlink-policy/)
  assert.match(page, /\/api\/v1\/hotlink-domains/)
  assert.match(page, /signedVariant/)
  assert.match(page, /allow_no_referer/)
  for (const locale of [chinese, english]) {
    assert.equal(typeof locale.media?.signedURLTitle, 'string')
    assert.equal(typeof locale.media?.hotlinkPolicyTitle, 'string')
    assert.equal(typeof locale.media?.hotlinkDomainsTitle, 'string')
    assert.equal(typeof locale.media?.hotlinkModes?.off, 'string')
    assert.equal(typeof locale.media?.hotlinkModes?.hybrid, 'string')
  }
})

test('member media visibility is owner-controlled and admin media details use an all-user preview endpoint', () => {
  const page = read('src/modules/member/pages/MemberMediaDetailPage.vue')
  const detail = read('src/core/resource/pages/ResourceDetailPage.vue')
  assert.match(page, /\/api\/v1\/media\/\$\{item\.value\.id\}\/visibility/)
  assert.match(page, /v-model="visibility"/)
  assert.match(detail, /\/api\/v1\/admin\/media\/\$\{String\(route\.params\.id\)\}\/content/)
})
