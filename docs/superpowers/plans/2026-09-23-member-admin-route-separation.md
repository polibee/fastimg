# 会员端与管理端路由分离实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将会员端设为根路径下的上传优先体验，并将实际管理页面统一迁移到 `/admin/**`，保持会员 own-scope 与管理 RBAC 边界。

**Architecture:** 继续使用一个 Vue Router 和现有两个 Shell。根路由及 `/media`、`/plans` 交给 `MemberShell`；原管理首页、标准资源、RBAC 和审计页面挂到 `/admin` 下。会员首页和媒体页通过共享的会员上传逻辑复用上传、处理轮询和媒体 API；生成资源仅对确属管理范围的资源注册，所有资源链接使用统一的管理路由解析规则。

**Tech Stack:** Vue 3、TypeScript、Vue Router、Pinia、vue-i18n、shadcn-vue、Vite、Node.js `node:test`、`vue-tsc`。

**Spec:** `docs/superpowers/specs/2026-09-23-member-admin-route-separation-design.md`

## Global Constraints

- 保持单一 Vue Router；通过嵌套路由和独立 Shell 清晰区分体验，不建立第二套前端工程或认证状态。
- 管理路由前缀集中定义/组合。`admin/src/core/resource/generated.ts` 等生成产物遵循生成流程更新；不得手改生成产物绕过生成器。
- `dashboardResourceRoute`、管理员侧栏、管理首页快捷入口、面包屑和命令式导航共用一致的管理路由解析规则。
- 路由守卫复用现有认证与 `auth.canAny`/权限状态机制；后端仍是最终授权边界，前端隐藏导航不作为安全措施。
- 会员页复用当前 MemberMediaPage 的 API 模式与上传状态机制。不得为了换 URL 新增重复上传服务、媒体服务或配额计算。
- 此次仅做前端路由/壳/导航/文档和相应测试；不改数据库、迁移、Redis、后端数据范围或生产服务。
- 不因管理路由收拢而扩展 API 权限。跨用户媒体管理作为依赖明确列入后续开发；应另行设计和测试后端 API/RBAC/审计闭环。
- 所有新增用户可见文案同时维护 `zh-CN`/`en-US`；保留工作区中其他未提交修改，不将无关变更放进本阶段提交。
- Laragon PostgreSQL/Redis 是用户管理的服务，本计划不连接、不迁移、不修改其配置。

## Review Focus

1. **同一账号兼有会员和管理权限时**，`/`、`/media`、`/plans` 仍始终是会员页面；测试覆盖双身份账号，不允许按角色重解释 URL。
2. **无管理权限的已登录用户直接访问 `/admin` 深链时**，守卫拒绝访问且不发起管理页面数据请求；测试覆盖直接导航和刷新入口。
3. **生成资源 path 有无前导斜杠、是否已包含 `/admin` 时**，路由与链接都只出现一个管理前缀；测试覆盖 `/plans`、`/admin/plans` 和根资源路径。
4. **会员文件夹/相册仍是 own-scope 且尚无会员页面时**，它们不会因为生成器注册而出现在管理导航或 `/admin/**`；测试断言二者不在 Admin 路由集。
5. **上传返回 processing、失败、配额错误或页面切换时**，首页不得误报成功或遗失现有媒体页重试体验；测试覆盖处理状态和错误呈现，避免复制第二套上传流程。

---

## 文件与职责映射

| 文件 | 职责 |
| --- | --- |
| `admin/src/router/index.ts` | 根会员路由、`/admin` 管理路由、兼容重定向和认证/管理权限守卫 |
| `admin/src/lib/admin-access.ts` | 当前管理权限白名单及可复用的后台入口可见性判断 |
| `admin/src/lib/login-redirect.ts` | 只接受站内绝对路径形式的登录 redirect，拒绝外站/协议相对地址 |
| `admin/src/lib/dashboard-resources.ts` | 资源显示权限过滤及生成资源到 `/admin/**` 的统一链接解析 |
| `admin/src/core/resource/generated.ts` | 由现有生成流程产出的标准资源页面路由；实现时不可手工编辑 |
| `admin/src/core/layouts/MemberShell.vue` | 会员主导航、会员账号操作、可选管理入口 |
| `admin/src/core/layouts/AdminShell.vue` | 管理导航、面包屑、资源搜索和管理首页/资源链接 |
| `admin/src/modules/member/pages/MemberHomePage.vue` | 根路径上传优先的会员首页、最近媒体与套餐摘要 |
| `admin/src/modules/member/components/MemberUploadPanel.vue` | 会员上传交互和队列展示，供首页与媒体库共用 |
| `admin/src/modules/member/composables/useMemberUpload.ts` | 调用现有上传 API、验证文件类型、跟踪 processing 状态、提供错误/重试状态 |
| `admin/src/modules/member/pages/MemberMediaPage.vue` | 本人媒体列表、回收站；替换内嵌上传实现为共享上传组件 |
| `admin/src/modules/auth/pages/LoginPage.vue` | 保留受限的内部 redirect；缺省登录落点为会员首页 `/` |
| `admin/src/locales/{zh-CN,en-US}/member.json` | 会员首页、会员导航、上传组件新增双语文案 |
| `admin/src/locales/{zh-CN,en-US}/auth.json` | 如需新增管理入口/403/认证重定向文案则同步维护 |
| `admin/src/i18n/index.ts` | 仅当新增 namespace 时注册；优先复用现有 `member`、`auth` namespace |
| `admin/tests/fastimg-routes.test.mjs` | 路由挂载、兼容跳转、管理资源路径及守卫契约测试 |
| `admin/tests/member-media.test.mjs` | 更新会员路由和导航契约，继续验证 own-scope API 与会员媒体功能 |
| `admin/tests/member-home.test.mjs` | 上传优先首页、共享上传组件、最近媒体和双语文案契约测试 |
| `admin/tests/dashboard-resources.test.ts` | 更新资源管理 URL 前缀及路径规范化测试 |
| `docs/fastimg-frontend-design.md` | 页面职责、扁平 URL、导航与管理 API 边界的当前设计/实现状态 |
| `docs/fastimg-product-design.md` | 如产品入口说明涉及旧 `/app` 路径则同步 |
| `docs/fastimg-stage-development-plan.md` | 更新阶段进度和入口重构验收项 |
| `docs/superpowers/plans/2026-09-23-fastimg-repository-development.md` | 更正 Task 4 的 `/app` 状态，引用本实施里程碑的完成结果 |
| `.superpowers/sdd/2026-09-23-fastimg-repository-development/progress.md` | 记录实现、测试、构建和未完成的跨用户后台媒体能力 |

> 若实现中确认某个建议的新文件边界与现有模块代码冲突，应优先复用已存在模块并在进度台账说明；不得因此复制上传业务逻辑。

## Task 1: 建立路由与管理资源路径测试

**Files:**
- Modify: `admin/tests/dashboard-resources.test.ts`
- Modify: `admin/src/lib/dashboard-resources.ts`

**Interfaces:**
- Consumes: 当前生成 manifest 的 `{ name, route, permissions }` 结构以及 `auth.user.permissions`。
- Produces: `adminResourcePath(route: string, fallbackName: string): string` 将根路径或旧 `/admin` 路径规范化为管理 URL（含记录详情后缀）；`dashboardResourceRoute({ name, route })` 委托给它并提供资源名 fallback。

- [x] **Step 1: 添加资源链接失败测试**，覆盖 `route: '/plans'`、`route: '/admin/plans'`、`route: 'plans'`、`route: '/'` 和记录路径 `/plans/12/edit`，期望统一得到一个且仅一个 `/admin` 前缀且保留详情后缀；把现有期望 `/plans` 的断言改为 `/admin/plans`。

```ts
assert.equal(adminResourcePath('/plans', 'plans'), '/admin/plans')
assert.equal(adminResourcePath('/admin/plans', 'plans'), '/admin/plans')
assert.equal(adminResourcePath('/plans/12/edit', 'plans'), '/admin/plans/12/edit')
assert.equal(adminResourcePath('/', 'plans'), '/admin/plans')
```
- [x] **Step 2: 运行资源 helper 专项测试确认失败**。

运行：

```powershell
cd admin
node --experimental-strip-types --test tests/dashboard-resources.test.ts
```

预期：资源链接断言因当前 helper 去掉 `/admin` 前缀而失败。

- [x] **Step 3: 实现唯一的资源路径规范化 helper**，测试先通过；不改 `generated.ts`。

```ts
export function adminResourcePath(route: string, fallbackName: string): string {
  const normalized = route.replace(/^\/+/, '').replace(/^admin\/+/, '')
  return `/admin/${normalized || fallbackName}`
}

export function dashboardResourceRoute(resource: Pick<DashboardResource, 'name' | 'route'>): string {
  return adminResourcePath(resource.route, resource.name)
}
```

- [x] **Step 4: 重跑 helper 专项测试**，确认规范化通过。

## Task 2: 迁移 Vue Router 与访问守卫

**Files:**
- Create: `admin/tests/fastimg-routes.test.mjs`
- Create: `admin/src/lib/admin-access.ts`
- Create: `admin/tests/admin-access.test.ts`
- Create: `admin/src/lib/login-redirect.ts`
- Create: `admin/tests/login-redirect.test.ts`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/tests/member-media.test.mjs`
- Modify: `admin/src/modules/auth/pages/LoginPage.vue`

**Interfaces:**
- Consumes: Task 1 的 `/admin/**` 路径规范化规则、当前 auth store `isAuthenticated`、`can`、`canAny`。
- Produces: `/`、`/media`、`/plans` 的会员页面；`/admin` 下的管理页面；静态旧会员路径重定向；拒绝无管理权限访问 AdminShell。

- [x] **Step 1: 扩充失败测试**，覆盖 `/app`、`/app/media`、`/app/plans` 的静态重定向和查询保留；验证管理路径在 `/admin/**`，旧根管理 `/media` 不再映射为 AdminShell 页面。
- [x] **Step 2: 为守卫和权限 helper 添加契约测试**，断言未登录 `/admin/plans` 保留 redirect，已登录但无后台权限的身份被拒绝，拥有当前已注册管理资源权限者可进入；`hasAdminAccess(permissions: string[]): boolean` 仅接受以下当前管理入口权限：`admin.users.view`、`admin.roles.manage`、`admin.permissions.manage`、`admin.plans.view`、`admin.advertising.view`、`admin.announcements.view`、`admin.departments.view`；不把 own-scope 的 `admin.folders.*`/`admin.albums.*` 单独视为全站管理权限；根 `/` 对会员和管理员双身份均为会员页。

```ts
assert.equal(hasAdminAccess([]), false)
assert.equal(hasAdminAccess(['media.upload']), false)
assert.equal(hasAdminAccess(['admin.folders.view']), false)
assert.equal(hasAdminAccess(['admin.plans.view']), true)
assert.equal(hasAdminAccess(['admin.roles.manage']), true)
```
- [x] **Step 3: 为登录 redirect helper 添加单元测试**，接受 `/`、`/media`、`/admin/plans?tab=active`，拒绝 `https://evil.example`、`//evil.example`、反斜线起始和非字符串值；拒绝项回落 `/`。实现使用固定基准 `new URL(value, 'https://app.invalid')` 检查同源，并要求输入以单个 `/` 开始且不含反斜线，避免协议相对/浏览器路径解析歧义。
- [x] **Step 4: 运行路由和 redirect 专项测试确认失败**，记录失败原因。

```powershell
cd admin
node --test tests/fastimg-routes.test.mjs tests/member-media.test.mjs
node --experimental-strip-types --test tests/admin-access.test.ts
node --experimental-strip-types --test tests/login-redirect.test.ts
```
- [x] **Step 5: 调整 router 路由树**：把 MemberShell 挂在 `/`，在 Task 3 首页完成前让默认 child 暂时静态重定向到已存在的会员媒体页，保持每一步都可编译；会员媒体、套餐 path 改为绝对扁平 URL；创建 `/admin` parent 并将管理首页、RBAC、审计、plans CRUD、标准后台资源挂入 children。

```ts
{
  path: '/',
  component: MemberShell,
  meta: { requiresAuth: true },
  children: [
    { path: '', redirect: { name: 'member-media' } }, // Task 3 替换为 member-home
    { path: 'media', name: 'member-media', component: MemberMediaPage },
    { path: 'plans', name: 'member-plans', component: MemberPlansPage },
  ],
},
{
  path: '/admin',
  component: AdminShell,
  meta: { requiresAuth: true, requiresAdminAccess: true },
  children: [
    { path: '', name: 'home', component: AdminHomePage },
    { path: 'rbac', name: 'rbac', component: RBACPage },
    { path: 'audit-logs', name: 'audit-logs', component: AuditLogPage },
    { path: 'plans', name: 'plans-resource-list', component: ResourceListPage, props: { resource: 'plans' } },
  ],
}
```
- [x] **Step 6: 将生成资源路由装配到 Admin children**，只纳入确属管理数据范围的定义；排除 `albums`、`folders` own-scope 路由；不注册 `/admin/media`，因为没有全站媒体 API/RBAC/审计。
- [x] **Step 7: 增加父级管理入口访问检查**，路由守卫调用 `hasAdminAccess(auth.user?.permissions ?? [])`；无有效后台权限时访问 `/admin/**` 返回既有 Forbidden 页面。`MemberShell` 的后台入口使用同一 helper。保留所有管理子页面已有 `permission`/`anyPermissions` 检查。
- [x] **Step 8: 处理认证落点**：登录组件调用 `resolveLoginRedirect(route.query.redirect)`，登录成功且有显式有效站内 redirect 时保留目标，否则到 `/`；已登录用户访问 `/login` 无 redirect 也回 `/`。不得按管理角色自动切换到 `/admin`。
- [x] **Step 9: 对无路径冲突的旧管理静态 URL 增加精确重定向**（例如 `/rbac`、`/audit-logs`、`/advertising` 到新管理路径）；不得给 `/media`、`/plans` 加按身份判断的重定向。`/folders`、`/albums` 不得重定向到伪管理页面。
- [x] **Step 10: 运行路由专项测试**，确保所有路由、身份组合、查询保留和旧链接断言通过；根路径此时仍重定向到会员媒体，Task 3 将其切换为上传首页并补充最终断言。

## Task 3: 共享会员上传流程并建立上传首页

**Files:**
- Create: `admin/src/modules/member/pages/MemberHomePage.vue`
- Create: `admin/src/modules/member/components/MemberUploadPanel.vue`
- Create: `admin/src/modules/member/composables/useMemberUpload.ts`
- Create: `admin/tests/member-home.test.mjs`
- Modify: `admin/src/modules/member/pages/MemberMediaPage.vue`
- Modify: `admin/src/locales/zh-CN/member.json`
- Modify: `admin/src/locales/en-US/member.json`
- Modify: `admin/tests/member-media.test.mjs`
- Modify: `admin/src/router/index.ts`

**Interfaces:**
- Consumes: 现有 `/api/v1/uploads` 上传 API、`GET /api/v1/uploads/:id` 状态查询、`GET /api/v1/media` 本人媒体列表、auth token 和当前图片格式约束。
- Produces: `MemberUploadPanel` 提供统一拖放/选择上传入口；`useMemberUpload` 暴露上传中状态、队列结果、错误 key、处理轮询和重试所需动作；首页显示最近媒体及 `/media`、`/plans` 入口。

```ts
type UploadStatus = 'uploading' | 'processing' | 'ready' | 'failed'
interface UploadQueueItem { id: string; file: File; status: UploadStatus; mediaId?: string; errorKey?: string }
function useMemberUpload(): {
  queue: Ref<UploadQueueItem[]>
  uploading: ComputedRef<boolean>
  uploadFiles(files: FileList | File[]): Promise<void>
  retry(id: string): Promise<void>
}
```

- [x] **Step 1: 写 MemberUploadPanel 合同测试**，断言文件选择、拖放、处理中/失败提示来自 i18n key；上传逻辑无 `user_id` 输入。

```js
assert.match(uploadPanel, /type="file"/)
assert.match(uploadPanel, /@drop/)
assert.match(uploadPanel, /member\.upload\.(processing|failed)/)
assert.doesNotMatch(uploadComposable, /user_id\s*[:=]/)
```
- [x] **Step 2: 写首页测试**，验证默认首页、最近媒体/套餐入口、own-scope API 和简单会员首页职责。
- [x] **Step 3: 运行新测试确认失败**；新建测试在实现前按预期失败。
- [x] **Step 4: 从 MemberMediaPage 当前上传实现提取 composable**：保留 JPEG/PNG/GIF 白名单、认证上传请求、processing 状态轮询、配额/大小/处理失败错误 key 和轮询清理；成功只在 ready 后显示完成。
- [x] **Step 5: 实现共享上传组件**，包含可键盘操作的 file input 和按钮、拖拽区域、队列状态及错误重试反馈。
- [x] **Step 6: 更新 MemberMediaPage 使用共享上传组件**，保留本人列表、搜索、软删除和恢复行为，避免双重上传逻辑。
- [x] **Step 7: 实现 MemberHomePage**，复用同一上传组件；通过现有会员 API 拉取最近图片和套餐/用量摘要，不提交 `user_id`、不另算配额。
- [x] **Step 8: 将 MemberHomePage 注册为 `/` 默认页面并更新 MemberShell 导航**；管理员入口只在统一管理权限通过时显示。
- [x] **Step 9: 同步中英文文案和测试**，新增会员首页/上传组件文案均在 zh-CN/en-US 中对齐并通过 i18n namespace 使用。
- [x] **Step 10: 运行首页和会员媒体专项测试**，专项用例通过。

## Task 4: 收拢管理壳、导航和生成资源入口

**Files:**
- Modify: `admin/src/core/layouts/AdminShell.vue`
- Modify: `admin/src/core/pages/AdminHomePage.vue`
- Modify: `admin/src/lib/dashboard-resources.ts`
- Modify: `admin/tests/dashboard-resources.test.ts`
- Modify: `admin/tests/resource-navigation.test.ts`
- Modify: `admin/tests/fastimg-routes.test.mjs`

**Interfaces:**
- Consumes: Task 1 path helper 与 Task 2 的 `/admin` 路由 parent。
- Produces: 管理 Shell 内所有 RouterLink、搜索跳转、仪表盘卡片、breadcrumb 可解析地址均处于 `/admin/**`；用户无权资源仍被隐藏。

- [x] **Step 1: 扩充失败测试**，覆盖 AdminShell `/admin` 链接、资源路由 scope 过滤和全局搜索规范化。
- [x] **Step 2: 明确可注册资源列表**：依据 `dataScope` 与 API manifest 的 `data_scope` 两种命名形式排除 own-scope 资源，不编辑生成产物。
- [x] **Step 3: 运行管理导航测试确认旧链接失败**；新增用例先红后绿。
- [x] **Step 4: 更新 AdminShell**：首页、RBAC、审计均使用 `/admin/**`；移除 own-scope media 入口；菜单、资源详情搜索统一规范化，并排除 own-scope 搜索结果。
- [x] **Step 5: 更新 AdminHomePage**：继续经统一的 `dashboardResourceRoute` helper 生成管理入口，并通过共用资源可见性过滤排除 own-scope。
- [x] **Step 6: 更新 breadcrumb 当前页解析**，管理首页命名与标签匹配；会员页面使用独立 MemberShell，不承载管理 breadcrumb。
- [x] **Step 7: 运行管理导航和资源路径专项测试**；路径及 scope 用例通过。

## Task 5: 文档同步、回归验证与阶段交付

**Files:**
- Modify: `docs/fastimg-frontend-design.md`
- Modify: `docs/fastimg-product-design.md`（仅更新与根入口/旧路径直接相关的产品说明）
- Modify: `docs/fastimg-stage-development-plan.md`
- Modify: `docs/superpowers/plans/2026-09-23-fastimg-repository-development.md`
- Modify: `.superpowers/sdd/2026-09-23-fastimg-repository-development/progress.md`
- Test: `admin/tests/fastimg-i18n.test.mjs`
- Test: all relevant `admin/tests/*` and frontend build/type checks

**Interfaces:**
- Consumes: Tasks 1-4 shipped route and UI behavior; current accurate test results.
- Produces: 文档按 designed/implemented/verified 分别描述新入口；保留全站媒体管理、文件夹/相册会员页等真实未完成功能状态。

- [x] **Step 1: 更新前端设计文档**，同步扁平会员路由、`/admin/**`、上传首页、历史兼容策略、own-scope 和 `/admin/media` 边界。
- [x] **Step 2: 更新产品文档中直接受影响的入口说明**，避免继续把 `/app` 写作会员首页；未来能力仍标记为未实现。
- [x] **Step 3: 更新阶段计划与主实施计划状态**，记录已实现路由/UI切片并保留跨用户媒体及会员文件夹/相册未完成项。
- [x] **Step 4: 更新 SDD progress**，记录实现、专项测试、构建/浏览器限制；完整测试扫描待本轮最终复跑后补录。
- [x] **Step 5: 执行会员与路由专项测试**：

```powershell
cd admin
node --test tests/fastimg-routes.test.mjs tests/member-home.test.mjs tests/member-media.test.mjs tests/member-plans.test.mjs tests/fastimg-i18n.test.mjs
node --experimental-strip-types --test tests/admin-access.test.ts tests/login-redirect.test.ts tests/dashboard-resources.test.ts tests/resource-navigation.test.ts
node --experimental-strip-types --test tests/dashboard-resources.test.ts tests/resource-navigation.test.ts
```

预期：所有专项测试通过；如测试运行器不支持 `--experimental-strip-types`，按仓库现有 Node/TypeScript 测试运行惯例调整命令，不修改依赖锁文件。

- [ ] **Step 6: 执行完整前端验证**：

```powershell
cd admin
pnpm exec vue-tsc --noEmit
pnpm run build
```

预期：类型检查和生产构建通过；记录每项命令实际输出，不以 Vite 启动成功替代 build 验证。

- [ ] **Step 7: 浏览器验收两个体验**：普通会员登录检查 `/` 上传页、`/media` 和 `/plans`；有管理权限账号检查 `/admin`、`/admin/plans`、RBAC、审计和生成资源；无管理权限账号直接访问 `/admin` 应被拒绝。若当前预览或账号不可用，记为未验证，不更改数据库、Redis、端口服务配置。
- [x] **Step 8: 检查 git diff 与安全边界**：本切片未修改后端/迁移/服务配置，会员 API 不接收 `user_id`，own-scope 资源已从后台路由/菜单/搜索过滤；保留工作树既有改动。
- [ ] **Step 9: 在整个入口重构切片验证通过后形成一个阶段里程碑本地提交**，仅暂存本计划范围内文件，不提交本工作树已有的无关改动，不推送远程仓库。

## Completion Criteria

- `/` 是会员上传首页；`/media` 和 `/plans` 是会员页面；会员导航没有 `/app/` 前缀。
- 管理壳与管理功能只从 `/admin/**` 进入，并同时通过父路由入口权限和原有具体页面权限检查。
- 管理侧栏、首页、资源搜索和 breadcrumb 不再链接根路径下的旧管理页面。
- `/app`、`/app/media`、`/app/plans` 兼容跳转正确；冲突路径没有角色相关隐式切换。
- Admin 自有媒体页不再冒充全站媒体管理；`/admin/media` 未注册；own-scope 文件夹和相册不显示为全站管理资源。
- 会员上传流程只有一套实现，媒体 ready 状态、失败、配额和重试信息正确；所有新增文案中英文齐备。
- 专项测试、i18n 检查、Vue 类型检查和构建实际结果已记录；未完成的后端跨用户管理能力如实保留。
