# FastImg 内容页、页脚导航与友情链接设计

日期：2026-09-28  
状态：待用户评审  
范围：会员端公开内容、管理员内容运营、游客友情链接申请

## 1. 目标与边界

FastImg 需要一组由管理员维护、对游客公开的站点内容能力：

- 默认提供隐私政策、使用条款、关于我们三个公开页面；
- 管理员可以使用富文本编辑器维护页面内容、SEO 信息和发布状态；
- 管理员可以维护页脚导航分类、菜单和子菜单；
- 游客可以提交友情链接申请，申请必须经过管理员审核后才展示；
- 公开页面、页脚和友链页面属于会员端/公共端，不跳转到 `/admin/**`；
- 后台能力使用现有认证、RBAC、审计、i18n、API Client、迁移和资源边界，不复制基础设施。

本切片不提供 HTML 源码编辑器，不允许管理员任意插入脚本、事件属性或 iframe。

## 2. 页面与路由

### 2.1 公共/会员端

| 路径 | 访问 | 说明 |
| --- | --- | --- |
| `/page/privacy` | 游客 | 隐私政策 |
| `/page/terms` | 游客 | 使用条款 |
| `/page/about` | 游客 | 关于我们 |
| `/friends` | 游客 | 已审核友情链接卡片 |

页面挂在 `MemberShell` 下，使用公共 API；不要求登录。页脚导航根据服务端返回的启用项渲染，当前语言按已有 vue-i18n 语言环境选择标题。

### 2.2 管理端

| 路径 | 权限 | 说明 |
| --- | --- | --- |
| `/admin/content-pages` | `admin.content_pages.view` | 内容页列表 |
| `/admin/content-pages/new` | `admin.content_pages.manage` | 新建内容页 |
| `/admin/content-pages/:id/edit` | `admin.content_pages.manage` | 编辑并发布内容页 |
| `/admin/footer-navigation` | `admin.footer_navigation.manage` | 分类、菜单、子菜单、排序和启用状态 |
| `/admin/friend-links` | `admin.friend_links.view` | 申请列表和审核 |

管理员页面挂在 `AdminShell` 下，所有写操作使用显式权限和审计日志。友链审核是专用动作，不允许通过通用 CRUD 直接修改审核状态。

## 3. 数据模型

### 3.1 `site_pages`

- `id`
- `slug`：小写 URL slug，唯一，只允许字母、数字、短横线；默认页面使用 `privacy`、`terms`、`about`
- `title`
- `content_json`：Tiptap JSON，唯一内容源
- `excerpt`
- `seo_title`
- `seo_description`
- `status`：`draft`、`published`、`archived`
- `published_at`
- `created_by`、`updated_by`
- 时间戳

服务端只允许白名单节点和属性：段落、标题、列表、引用、代码块、粗体、斜体、链接、图片（仅受控 URL）。链接协议仅允许 `https`、`http`、`mailto`；拒绝 `javascript:`、`data:`、事件属性、脚本节点和任意 iframe。

### 3.2 `footer_navigation_groups`

- `id`
- `title`
- `locale`：`zh-CN`、`en-US` 或 `all`
- `sort_order`
- `is_enabled`

### 3.3 `footer_navigation_items`

- `id`
- `group_id`
- `parent_id`：为空表示菜单，非空表示子菜单；服务端禁止循环和超过两级
- `label`
- `target_type`：`page`、`friends`、`url`
- `target_value`：页面 slug、固定 `friends` 或外部 URL
- `open_in_new_tab`
- `sort_order`
- `is_enabled`

站内页面目标只允许指向已发布 `site_pages`；外部 URL 进行协议和长度校验。删除页面或导航分类时，服务端阻止仍被引用的记录，避免产生死链。

### 3.4 `friend_link_submissions`

- `id`
- `site_name`
- `url`
- `logo_url`
- `description`
- `contact_email`
- `submitted_by`：游客为空，登录用户可记录用户 ID
- `status`：`pending`、`approved`、`rejected`、`hidden`
- `review_note`
- `reviewed_by`、`reviewed_at`
- `created_at`、`updated_at`

默认状态为 `pending`。只有 `approved` 且 URL 校验通过的记录进入公共 API。审核通过后仍允许管理员隐藏，不删除原申请审计记录。

## 4. API 契约

### 4.1 公共 API

- `GET /api/v1/site/pages/{slug}`：返回已发布页面、标题和 SEO 字段
- `GET /api/v1/site/footer-navigation`：按语言返回启用的导航树
- `GET /api/v1/friend-links`：分页返回已审核友链
- `POST /api/v1/friend-links`：游客提交申请

公开接口不返回审核备注、申请人邮箱、内部用户 ID 或审计字段。

### 4.2 管理 API

- `GET /api/v1/admin/content-pages`
- `POST /api/v1/admin/content-pages`
- `GET /api/v1/admin/content-pages/{id}`
- `PUT /api/v1/admin/content-pages/{id}`
- `POST /api/v1/admin/content-pages/{id}/actions/publish`
- `POST /api/v1/admin/content-pages/{id}/actions/archive`
- `GET/POST/PUT/DELETE /api/v1/admin/footer-navigation/groups...`
- `GET /api/v1/admin/friend-links`
- `POST /api/v1/admin/friend-links/{id}/actions/approve`
- `POST /api/v1/admin/friend-links/{id}/actions/reject`
- `POST /api/v1/admin/friend-links/{id}/actions/hide`

审核动作必须使用显式管理员权限、状态机校验和审计日志；不能通过通用 `PUT` 修改 `status` 绕过动作日志。

## 5. 富文本编辑器与安全

管理端使用 Tiptap StarterKit，初始扩展仅包含：

- 段落、标题、粗体、斜体、删除线；
- 有序/无序列表、引用、代码块；
- 链接和图片 URL；
- 撤销/重做、占位提示、字数提示。

不提供 HTML 源码编辑入口。保存前做前端 schema 校验，服务端再次解析并清洗 Tiptap JSON。读取时只渲染清洗后的内容；不使用 `v-html` 直接输出未清洗字符串。链接和图片 URL 必须经过协议白名单校验，外链默认 `rel="nofollow noopener noreferrer"`。

Tiptap 依赖固定在 `admin/package.json`，不复制一套编辑器实现；编辑器样式复用现有 shadcn-vue 输入、按钮和语义色彩。

## 6. 默认内容

迁移/Seeder 创建三个默认草稿或发布页面（由部署环境决定，开发环境默认发布）：

- 隐私政策：说明账户、上传媒体、访问记录、支付和 Cookie 的收集与用途；
- 使用条款：说明允许用途、版权责任、违规内容处理、账户和套餐规则；
- 关于我们：说明 FastImg 面向开发者、站长和内容创作者，提供媒体托管、稳定链接和 API。

默认内容使用简洁、可修改的普通语言，不包含虚构的公司地址、电话、监管资质或法律承诺。管理员首次进入内容页面时可以直接编辑和替换。

## 7. 前端体验

- `MemberShell` 页脚保持简洁，按分类横向/换行显示，移动端不溢出；
- `/friends` 使用轻量网格卡片，展示 Logo、站点名称、描述和访问链接；不展示申请邮箱或审核信息；
- 游客提交表单提供明确的“提交审核”状态反馈，成功后提示“提交成功，等待管理员审核”；
- 友链申请失败时显示 URL、邮箱、限流或验证原因的可读错误；
- 管理员内容页编辑器、导航树和审核列表均同步提供中文/英文文案；
- 不在会员端显示管理端链接，除非当前用户拥有既有管理员访问权限。

## 8. 防滥用与审计

- 游客提交按 IP、URL 和联系邮箱做冷却与每日上限；
- 若站点启用 Turnstile，则友链提交必须通过现有 Turnstile 校验；未启用时不强制外部服务；
- URL 禁止内网、回环和明显危险协议，避免把友链表单变成 SSRF 入口；服务端不主动抓取站点内容；
- 创建、编辑、发布、归档、审核通过、拒绝、隐藏和删除都写入审计日志；
- 审计日志不记录完整邮箱、Token 或正文密钥。

## 9. SEO/SSG

将 `/friends` 和三个 `/page/*` 页面加入公开页面清单与 SSG 路由契约。构建时使用默认内容生成静态壳；运行时从公开 API 获取最新发布内容。若发布内容在构建后更新，页面仍可通过 API 展示最新内容，站点地图只收录已发布页面和 `/friends`。

当前阶段先完成真实公开 API、页面和 SSG 页面契约；不宣称动态内容已经在构建产物中预渲染，直到实际执行构建并验收 HTML。

## 10. 验收标准

1. 未登录访问四个公共路径均不跳转登录。
2. 默认三个页面可打开，页脚显示对应链接。
3. 管理员可编辑 Tiptap 内容、保存草稿、发布和归档。
4. 脚本、事件属性、危险 URL 和 iframe 无法保存或输出。
5. 管理员可创建分类、菜单、子菜单、排序和启停，会员端实时按公开 API 展示。
6. 游客可提交友链，提交后不出现在 `/friends`。
7. 管理员审核通过后，友链出现在公共页面；拒绝/隐藏后立即消失。
8. 未授权管理员访问管理 API 返回 403，所有管理动作进入审计日志。
9. 中英文 locale、前端契约测试、后端服务测试和迁移测试通过。
10. 迁移只在明确授权的本地开发数据库执行；生产数据库、外部邮件和真实站点发布不在本切片内。

