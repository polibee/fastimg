# 内容页、页脚导航与友情链接实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 FastImg 建立可由管理员维护的公开内容页、页脚导航和友情链接审核闭环，并让游客可以安全访问和提交友链申请。

**Architecture:** 复用现有 Goravel Service/Controller、迁移、RBAC、审计、Resource Engine、Vue Router、MemberShell/AdminShell、API Client 和双语 locale。页面内容以服务端清洗后的 Tiptap JSON 为唯一来源；公共端只读取已发布内容和已审核友链，管理端通过显式权限和专用动作管理状态。新功能放在 `modules/content`、`modules/footer_navigation`、`modules/friend_links` 边界内，不把业务规则复制进 Admin Core。

**Tech Stack:** Go 1.27、Goravel、PostgreSQL、Vue 3、Vue Router、TypeScript、shadcn-vue、Tiptap (`@tiptap/vue-3`、`@tiptap/starter-kit`、`@tiptap/extension-link`、`@tiptap/extension-image`)、vue-i18n、Node test runner。

**Spec:** `docs/superpowers/specs/2026-09-28-content-pages-footer-friend-links-design.md`

## Global Constraints

- 公共路径固定为 `/page/privacy`、`/page/terms`、`/page/about` 和 `/friends`，不要求登录。
- 管理路径固定在 `/admin/**`，管理写操作需要显式 RBAC 权限和审计日志。
- 内容页只保存 Tiptap JSON；不提供 HTML 源码编辑，不渲染未清洗的 `v-html`。
- 允许的内容节点和 URL 协议必须由服务端白名单再次校验；拒绝脚本、事件属性、危险 URL、任意 iframe 和内网目标。
- 友链默认 `pending`，只有 `approved` 才能出现在公共 API；审核状态不能通过通用 CRUD 直接修改。
- 所有用户可见文本同步维护 `admin/src/locales/zh-CN` 和 `admin/src/locales/en-US`。
- 不覆盖工作区已有的 `backend/app/services/auth/registration_service.go`、`backend/app/services/media/repository.go`、`backend/app/services/plans/lifecycle.go`、`backend/app/services/users/user_service.go` 未提交改动。
- 迁移文件只创建和测试，不执行任何数据库迁移，除非用户明确授权目标本地开发库。

## Review Focus

- 危险 Tiptap 节点、属性和 URL 被服务端拒绝；由 Task 1 的 sanitizer 测试覆盖。
- 未发布页面、未审核友链和被隐藏友链不出现在公开 API；由 Task 2/5 的公开接口测试覆盖。
- 游客重复提交和危险/内网 URL 被拒绝或限流；由 Task 5 的请求校验测试覆盖。
- 导航循环、超过两级子菜单和被引用页面删除被阻止；由 Task 4 的服务测试覆盖。
- 未授权管理员不能读取或执行管理写操作，且审核动作记录审计；由 Task 2/5 的授权测试覆盖。

---

### Task 1: 数据模型、迁移、默认页面和 Tiptap 清洗边界

**Files:**
- Create: `backend/app/modules/content/models/site_page.go`
- Create: `backend/app/modules/content/services/content_service.go`
- Create: `backend/app/modules/content/services/content_sanitizer.go`
- Create: `backend/app/modules/footer_navigation/models/navigation_group.go`
- Create: `backend/app/modules/footer_navigation/models/navigation_item.go`
- Create: `backend/app/modules/friend_links/models/submission.go`
- Create: `backend/database/migrations/20260928000001_create_content_navigation_friend_links_tables.go`
- Create: `backend/database/seeders/site_content.go`
- Create: `backend/app/modules/content/services/content_sanitizer_test.go`
- Create: `backend/database/migrations/content_navigation_friend_links_migration_test.go`
- Modify: `backend/database/seeders/database_seeder.go` (register the new seeder using the existing seeder order)

**Interfaces:**
- Produces `content.Service.GetPublished(ctx, slug)`, `content.Service.SaveDraft(ctx, input)`, `content.Service.Publish(ctx, id)`, `content.Service.Archive(ctx, id)`.
- Produces `content.SanitizeDocument(input map[string]any) (map[string]any, error)` and `content.ValidateURL(raw string, allowMailto bool) error`.
- Produces models and migration tables named `site_pages`, `footer_navigation_groups`, `footer_navigation_items`, and `friend_link_submissions` with the fields in the design spec.

- [ ] **Step 1: Write failing sanitizer tests**

  Add tests for: a valid paragraph/heading/link document is retained; `script`, `iframe`, event attributes, `javascript:` and `data:` URLs are rejected; external links retain only safe attributes.

- [ ] **Step 2: Run the sanitizer tests and verify RED**

  Run: `go test ./app/modules/content/services -run 'TestSanitize|TestValidateURL' -count=1`
  Expected: FAIL because the sanitizer package and functions do not exist.

- [ ] **Step 3: Implement the sanitizer and content service**

  Parse only the permitted Tiptap node/mark shapes, normalize text, validate links/images, and return a new document without unknown fields. Do not use a browser HTML sanitizer as the primary boundary; the persisted format is JSON.

- [ ] **Step 4: Add migration and default seeder**

  Add foreign keys, unique slug/indexes, status checks, navigation parent/group indexes, and friend-link review indexes. Seed `privacy`, `terms`, and `about` with plain-language content and no fictional company contact details, plus enabled “法律与隐私” and “关于” footer groups linking the published pages and `/friends`. Make the seeder idempotent and do not overwrite administrator edits.

- [ ] **Step 5: Run backend migration/seeder tests**

  Run: `go test ./database/migrations ./database/seeders ./app/modules/content/... -count=1`
  Expected: PASS; migration tests inspect schema definitions without connecting to a production database.

- [ ] **Step 6: Commit the data/security slice**

  ```bash
  git add backend/app/modules/content backend/app/modules/footer_navigation backend/app/modules/friend_links backend/database/migrations/20260928000001_create_content_navigation_friend_links_tables.go backend/database/seeders
  git commit -m "feat: add content navigation and friend link schema"
  ```

### Task 2: Public and administrator content-page APIs with RBAC and audit

**Files:**
- Create: `backend/app/modules/content/controllers/public_controller.go`
- Create: `backend/app/modules/content/controllers/admin_controller.go`
- Create: `backend/app/modules/content/request.go`
- Create: `backend/app/modules/content/service_test.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/modules/admin/registry/discovery.go` or the generated content discovery file, following the existing resource registration marker
- Modify: `backend/app/console/admin_bootstrap_command.go` and permission seeding in the existing RBAC seeder
- Modify: `backend/app/openapi/spec.go`
- Modify: `backend/app/openapi/spec_test.go`

**Interfaces:**
- Public: `GET /api/v1/site/pages/{slug}` returns only a published page.
- Admin: `GET/POST/GET/{id}/PUT/{id}` for content-page management plus `POST /actions/publish` and `POST /actions/archive`.
- Permissions: `admin.content_pages.view` and `admin.content_pages.manage`.

- [ ] **Step 1: Write failing service/controller tests**

  Cover published-only public reads, draft reads for authorized admins, slug validation, sanitizer rejection, publish/archive transitions, and 403 for missing permissions.

- [ ] **Step 2: Run the focused tests and verify RED**

  Run: `go test ./app/modules/content/... ./app/core/admin/... -run 'Test(Content|Published|Publish|Archive|Permission)' -count=1`
  Expected: FAIL because routes/controllers and service behavior are absent.

- [ ] **Step 3: Implement content service/controller and routes**

  Keep public controller response fields separate from admin fields. Use the existing permission middleware, validation/error envelope and audit service. Publish only sanitized JSON; set `published_at` on transition and clear it when archived.

- [ ] **Step 4: Register permissions and OpenAPI**

  Add permission labels in both languages, expose paths and schemas in the generated OpenAPI definition, and ensure the default administrator role receives the new management permission through existing bootstrap/seed behavior.

- [ ] **Step 5: Run focused and OpenAPI tests**

  Run: `go test ./app/modules/content/... ./app/openapi/... ./app/core/admin/... -count=1`
  Expected: PASS.

- [ ] **Step 6: Commit the content API slice**

  ```bash
  git add backend/app/modules/content backend/routes/web.go backend/app/openapi backend/app/console/admin_bootstrap_command.go
  git commit -m "feat: add published content page APIs"
  ```

### Task 3: Tiptap administrator editor and content-page management UI

**Files:**
- Modify: `admin/package.json` and the package lock used by this checkout
- Create: `admin/src/modules/content/api.ts`
- Create: `admin/src/modules/content/components/RichTextEditor.vue`
- Create: `admin/src/modules/content/pages/ContentPagesListPage.vue`
- Create: `admin/src/modules/content/pages/ContentPageFormPage.vue`
- Create: `admin/tests/content-pages.test.mjs`
- Create: `admin/tests/content-editor-security.test.mjs`
- Modify: `admin/src/apps/admin/routes.ts`
- Modify: `admin/src/locales/zh-CN/content.json`
- Modify: `admin/src/locales/en-US/content.json`
- Modify: `admin/src/i18n/index.ts` only if locale discovery requires registration

**Interfaces:**
- `contentApi.list(token)`, `contentApi.show(token, id)`, `contentApi.create(token, payload)`, `contentApi.update(token, id, payload)`, `contentApi.publish(token, id)`, `contentApi.archive(token, id)`.
- `RichTextEditor` accepts `modelValue: JSONContent`, emits `update:modelValue`, and never accepts raw HTML.

- [ ] **Step 1: Add failing route/editor contract tests**

  Assert `/admin/content-pages` and create/edit routes are permission-gated, the editor imports Tiptap Vue/StarterKit extensions, there is no HTML source textarea, and the payload uses `content_json` rather than `content_html`.

- [ ] **Step 2: Run the UI tests and verify RED**

  Run: `node --experimental-strip-types --test admin/tests/content-pages.test.mjs admin/tests/content-editor-security.test.mjs`
  Expected: FAIL because routes, API module and editor do not exist.

- [ ] **Step 3: Install the same-version Tiptap packages**

  Add `@tiptap/vue-3`, `@tiptap/starter-kit`, `@tiptap/extension-link`, and `@tiptap/extension-image` together. Keep the lockfile consistent and follow the repository WSL network procedure before downloading dependencies.

- [ ] **Step 4: Implement the editor and pages**

  Reuse shadcn-vue buttons, inputs, cards and alerts. The form edits slug/title/SEO fields, shows a bounded editor, saves draft, publishes, and archives. Display server validation errors in plain language and keep all strings in both locale files.

- [ ] **Step 5: Run UI tests and type checks**

  Run: `node --experimental-strip-types --test admin/tests/content-pages.test.mjs admin/tests/content-editor-security.test.mjs`
  Then run: `pnpm exec vue-tsc -b`
  Expected: tests PASS; type check has no new errors.

- [ ] **Step 6: Commit the editor slice**

  ```bash
  git add admin/package.json admin/pnpm-lock.yaml admin/src/modules/content admin/src/apps/admin/routes.ts admin/src/locales admin/tests/content-pages.test.mjs admin/tests/content-editor-security.test.mjs
  git commit -m "feat: add admin content page editor"
  ```

### Task 4: Footer navigation service, API and administrator tree UI

**Files:**
- Create: `backend/app/modules/footer_navigation/controllers/public_controller.go`
- Create: `backend/app/modules/footer_navigation/controllers/admin_controller.go`
- Create: `backend/app/modules/footer_navigation/service.go`
- Create: `backend/app/modules/footer_navigation/service_test.go`
- Create: `admin/src/modules/footer-navigation/api.ts`
- Create: `admin/src/modules/footer-navigation/pages/FooterNavigationPage.vue`
- Create: `admin/tests/footer-navigation.test.mjs`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/openapi/spec.go`
- Modify: `admin/src/apps/admin/routes.ts`
- Modify: `admin/src/locales/zh-CN/footer-navigation.json`
- Modify: `admin/src/locales/en-US/footer-navigation.json`

**Interfaces:**
- Public: `GET /api/v1/site/footer-navigation?locale=zh-CN` returns groups with ordered enabled items and children.
- Admin: group/item CRUD endpoints under `/api/v1/admin/footer-navigation` plus reorder endpoint.
- Permissions: `admin.footer_navigation.view` and `admin.footer_navigation.manage`.

- [ ] **Step 1: Write failing navigation tests**

  Assert enabled/locale-filtered tree output, stable sort order, cycle rejection, parent depth limit of two levels, target validation, and blocked deletion of a page/group still referenced by an enabled item.

- [ ] **Step 2: Run tests and verify RED**

  Run: `go test ./app/modules/footer_navigation/... -count=1`
  Expected: FAIL because the service is absent.

- [ ] **Step 3: Implement service/controllers/permissions**

  Keep navigation target resolution server-side. Page targets must resolve to published pages, `friends` maps to `/friends`, and URL targets must pass the shared safe URL validator. Record create/update/reorder/enable/disable/delete in audit logs.

- [ ] **Step 4: Implement admin tree editor**

  Use a compact grouped list with explicit “分类 / 菜单 / 子菜单” labels, add-child actions, drag-free numeric ordering first, enable switches, target type selector and server error display. Do not introduce a second generic tree primitive.

- [ ] **Step 5: Run backend/UI tests**

  Run: `go test ./app/modules/footer_navigation/... ./app/openapi/... -count=1` and `node --experimental-strip-types --test admin/tests/footer-navigation.test.mjs`.
  Expected: PASS.

- [ ] **Step 6: Commit the navigation slice**

  ```bash
  git add backend/app/modules/footer_navigation backend/routes/web.go backend/app/openapi admin/src/modules/footer-navigation admin/src/apps/admin/routes.ts admin/src/locales admin/tests/footer-navigation.test.mjs
  git commit -m "feat: add managed footer navigation"
  ```

### Task 5: Friend-link public submission and administrator moderation

**Files:**
- Create: `backend/app/modules/friend_links/controllers/public_controller.go`
- Create: `backend/app/modules/friend_links/controllers/admin_controller.go`
- Create: `backend/app/modules/friend_links/service.go`
- Create: `backend/app/modules/friend_links/service_test.go`
- Create: `admin/src/modules/friend-links/api.ts`
- Create: `admin/src/modules/friend-links/pages/FriendLinksAdminPage.vue`
- Create: `admin/src/modules/member/pages/FriendLinksPage.vue`
- Create: `admin/src/modules/member/components/FriendLinkSubmitForm.vue`
- Create: `admin/tests/friend-links.test.mjs`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/openapi/spec.go`
- Modify: `admin/src/apps/member/route-parts.ts`
- Modify: `admin/src/apps/admin/routes.ts`
- Modify: `admin/src/core/layouts/MemberShell.vue`
- Modify: `admin/src/locales/zh-CN/member.json`
- Modify: `admin/src/locales/en-US/member.json`
- Modify: `admin/src/locales/zh-CN/moderation.json`
- Modify: `admin/src/locales/en-US/moderation.json`

**Interfaces:**
- Public: `GET /api/v1/friend-links`, `POST /api/v1/friend-links`.
- Admin: `GET /api/v1/admin/friend-links` and action endpoints `approve`, `reject`, `hide`.
- Permissions: `admin.friend_links.view` and `admin.friend_links.moderate`.

- [ ] **Step 1: Write failing moderation and abuse tests**

  Cover pending-by-default, approved-only public output, reject/hide transitions, duplicate URL, invalid public URL, loopback/private host rejection, IP/email cooldown, optional Turnstile invocation when enabled, and audit event creation.

- [ ] **Step 2: Run tests and verify RED**

  Run: `go test ./app/modules/friend_links/... -count=1`
  Expected: FAIL because the service and endpoints are absent.

- [ ] **Step 3: Implement service/controllers and routes**

  Accept nullable authenticated user identity, never fetch the submitted URL server-side, use existing rate-limit/settings services, and expose only public-safe fields. State transitions must be explicit and idempotent where appropriate.

- [ ] **Step 4: Implement public page and admin moderation UI**

  Add `/friends` to the member public route tree. Render approved links as restrained responsive cards. Put submission below the list with a clear pending notice. Add the admin table with approve/reject/hide actions and a review-note field; keep actions separate from generic editing.

- [ ] **Step 5: Run backend/UI tests**

  Run: `go test ./app/modules/friend_links/... ./app/openapi/... -count=1` and `node --experimental-strip-types --test admin/tests/friend-links.test.mjs`.
  Expected: PASS.

- [ ] **Step 6: Commit the friend-link slice**

  ```bash
  git add backend/app/modules/friend_links backend/routes/web.go backend/app/openapi admin/src/modules/friend-links admin/src/modules/member admin/src/apps admin/src/core/layouts/MemberShell.vue admin/src/locales admin/tests/friend-links.test.mjs
  git commit -m "feat: add moderated friend links"
  ```

### Task 6: Public footer rendering, SEO/SSG manifest and full integration tests

**Files:**
- Create: `admin/src/modules/member/components/MemberFooterNavigation.vue`
- Create: `admin/src/modules/member/api/public-content-api.ts`
- Modify: `admin/src/core/layouts/MemberShell.vue`
- Modify: `admin/src/apps/member/public-pages.json`
- Modify: `admin/src/apps/member/route-public.ts`
- Modify: `admin/scripts/prerender-seo.mjs`
- Create: `admin/tests/public-content-pages.test.mjs`
- Modify: `admin/tests/fastimg-routes.test.mjs`
- Modify: `admin/tests/member-home.test.mjs`
- Modify: `docs/fastimg-stage-development-plan.md`
- Modify: `docs/fastimg-product-design.md`

**Interfaces:**
- `publicContentApi.page(slug)`, `publicContentApi.footerNavigation(locale)`, and `publicContentApi.friendLinks()` use `apiFetch` and never require a token.
- `MemberFooterNavigation` renders the server-provided groups and children and exposes visible keyboard focus.

- [ ] **Step 1: Write failing public route/footer/SSG tests**

  Assert the four public paths are registered, public API calls do not require auth, the MemberShell renders a footer component, and `public-pages.json` includes `/friends` and the three `/page/*` entries.

- [ ] **Step 2: Run tests and verify RED**

  Run: `node --experimental-strip-types --test admin/tests/public-content-pages.test.mjs admin/tests/fastimg-routes.test.mjs`
  Expected: FAIL because public routes/footer/manifest entries are absent.

- [ ] **Step 3: Implement public API/page/footer integration**

  Fetch published content on route entry, set document title/description from API data, render sanitized content, add the footer after existing footer ad content, and preserve current member authentication/nav behavior. Make empty navigation graceful.

- [ ] **Step 4: Extend SSG manifest/prerender handling**

  Add public paths and static SEO metadata. Keep the runtime API fallback for content changes after build; do not claim dynamic content is embedded until a real build output is inspected.

- [ ] **Step 5: Run full frontend contract tests and build checks**

  Run: `node --experimental-strip-types --test admin/tests/*.mjs admin/tests/*.ts` (using the repository’s existing test selection if the shell cannot expand TypeScript tests), then `pnpm exec vue-tsc -b`, and finally `pnpm run build:ssg` when dependencies are available.
  Expected: all applicable tests pass; if the known Windows Tailwind optional dependency limitation blocks build, record the exact error without changing dependency versions.

- [ ] **Step 6: Update progress documentation and commit**

  Record implemented/verified/not-yet-verified status in the stage plan and product design. Commit only feature files and docs; do not stage the four pre-existing user modifications.

### Task 7: Local migration and acceptance handoff

**Files:**
- No new production files; use the migration from Task 1 and existing local acceptance scripts.
- Modify only if needed: `backend/scripts/contract-smoke.ps1` and the feature test fixtures.

- [ ] **Step 1: Review migration target and obtain explicit local database authorization**

  Confirm the target is only the local `fastimg_dev` database and that no production database or external service is in scope.

- [ ] **Step 2: Run the migration and seed the three default pages**

  Execute the existing migration command against `fastimg_dev`; verify the four new tables, default pages, permissions and idempotent rerun behavior.

- [ ] **Step 3: Restart only FastImg-owned backend/frontend processes**

  Preserve unrelated projects, use separate ports, retain PID/log ownership, and validate public/member/admin URLs with the existing local network procedure before reporting links.

- [ ] **Step 4: Run end-to-end acceptance**

  Verify anonymous page/footer/friend access, guest submission, admin approval/rejection/hide, content editing/publishing, XSS rejection, audit records and permission failures.

- [ ] **Step 5: Commit only acceptance/documentation updates**

  ```bash
  git add docs/fastimg-stage-development-plan.md docs/fastimg-product-design.md backend/scripts/contract-smoke.ps1
  git commit -m "docs: record content and friend link acceptance"
  ```

