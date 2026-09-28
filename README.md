# FastImg

FastImg is a membership-based media hosting platform for developers, site owners, and content creators. It provides stable image links, a member media library, image processing, Personal API Tokens, plan-controlled quotas, payment channels, moderation, hotlink protection, and an administrator operations panel.

It is built on the Go Vue Admin foundation: Goravel 1.18, Go, PostgreSQL, Redis, Vue 3, TypeScript, Vite, and shadcn-vue. Existing framework capabilities such as authentication, RBAC, Resource pages, migrations, queues, audit logging, i18n, and the API client are reused rather than duplicated.

[![RackNerd VPS](https://img.shields.io/badge/RackNerd-VPS-2563eb?style=for-the-badge)](https://my.racknerd.com/aff.php?aff=7572)
[![Vast.ai GPU Cloud](https://img.shields.io/badge/Vast.ai-GPU%20Cloud-7c3aed?style=for-the-badge)](https://cloud.vast.ai/?ref_id=91181)

## Documentation

- [中文文档](./README.zh-CN.md)
- [FastImg deployment guide](./docs/fastimg-deployment.md)
- [FastImg production gates](./docs/fastimg-production-gates.md)
- [FastImg product design](./docs/fastimg-product-design.md)
- [FastImg stage development plan](./docs/fastimg-stage-development-plan.md)
- [FastImg developer API](./docs/fastimg-developer-api.md)
- [FastImg enhancement notes](./docs/fastimg-enhancements.md)
- [FastImg authentication](./docs/authentication.md)
- [FastImg payment gateway design](./docs/fastimg-payment-gateway.md)
- [FastImg frontend design](./docs/fastimg-frontend-design.md)
- [AI Quickstart](./docs/ai-quickstart.md)
- [Platform architecture and roadmap](./docs/README.md)

## Product capabilities

### Member experience

- Guest-accessible home, plans, and discovery pages; upload and personal operations require sign-in.
- Free service is available. Paid plans add configurable storage, file size, upload, API, bandwidth, image-processing, token, advertising, and watermark entitlements.
- Upload works from the web picker, drag and drop, Ctrl+V, Personal API Token, PicGo, ShareX, curl, and CLI. All entry points use the same quota and authorization rules.
- Image upload returns absolute Original, Markdown, HTML, BBCode, and direct-link formats immediately when processing is ready; the current storage policy keeps one normalized original instead of creating three fixed copies.
- `Image processing operations per month` counts one successful image-processing pipeline per upload. It covers format/magic validation, decoding and metadata cleanup, plus plan-controlled watermark processing; failed uploads are not counted and `0` means unlimited.
- Image upload returns absolute Original, Markdown, HTML, BBCode, and direct-link formats immediately when processing is ready.
- Personal media library with public, unlisted, and private visibility, folders, albums, recycle bin, restore, permanent deletion, and access controls.
- Media expiry supports user-selected retention, reminders, recycle-bin recovery, and permanent deletion. Member export archives include original files, metadata, link lists, and collection relations.
- Personal API Tokens with fixed minimal permissions: `upload:write`, `media:read`, and `media:delete`. Token expiration can be custom or permanent.
- API upload, image list/detail/link retrieval, and owner-only deletion. Member API tokens cannot call administrator APIs.
- Plan checkout supports multiple enabled payment providers; each order chooses one provider, while the default gateway is only a recommendation.

### Administrator experience

- Separate `/admin/**` routes and AdminShell. The member frontend uses `/`, `/media`, `/plans`, `/discover`, `/folders`, and `/albums`.
- Member, plan, advertising, media, album, folder, order, payment event, refund, report, access-log, statistics, settings, token, and audit management according to RBAC permissions.
- Administrator media actions are explicit: hide, restore, approve, reject, and permanently delete. They are not exposed through generic media CRUD and are audited.
- Post-publication moderation: newly uploaded media is available by default; reports and administrator review can later hide or reject content and remove it from discovery.
- Provider-specific payment configuration with encrypted credentials, configurable enablement, callbacks, order snapshots, payment transactions, webhook events, fulfillment, and refunds.
- Site settings for SEO, sitemap, robots, email verification, custom code, watermarks, storage limits, and payment providers.
- Public status and maintenance pages expose service health without leaking credentials or internal addresses; discovery uses post-publication moderation and report handling.

### Developer and operations foundation

- JWT authentication, refresh, logout, PostgreSQL-authoritative refresh-token storage, and Redis acceleration/queue support.
- Resource Manifest-driven list, create, edit, detail, permissions, menus, search, filters, relations, and batch actions.
- High-cardinality member and administrator lists use server pagination with total counts; low-cardinality configuration lists remain bounded full reads.
- Request/response audit logging with sensitive-field redaction and retention cleanup.
- Redis-backed recovery and fulfillment jobs with retry boundaries; PostgreSQL remains the source of business facts.
- Access Tokens are kept in memory; Access Tokens default to 60 minutes and are automatically refreshed before expiry. Durable refresh sessions default to 30 days and are controlled by `JWT_REFRESH_TTL`.
- SSG entry pages for the public home, plans, and discovery pages. Full SSR is not currently enabled.
- Local storage is suitable for development and controlled validation. Production still requires object storage/CDN integration and recovery verification.

## Local development

Start PostgreSQL and Redis through Laragon. The backend must use PostgreSQL and Redis; it must not silently fall back to an in-memory replacement.

~~~powershell
cd backend
go run .
~~~

In another terminal, start the admin panel:

~~~powershell
cd admin
pnpm install
pnpm dev
~~~

Development port examples:

- Go API: http://127.0.0.1:3000
- Go/Vue member and admin frontend: http://127.0.0.1:5180
- Admin routes: http://127.0.0.1:5180/admin

When multiple projects are running, assign each project a unique frontend and backend port, use Vite `strictPort`, and verify the actual Windows-side URL before sharing it. Do not let a dev server silently increment into another project's port.

Review generated migrations before applying them. The generator does not silently execute migrations.

## Resource generation

~~~powershell
cd backend
go run . admin:make-resource announcements --fields="title:text:required,status:select:required:draft=Draft|published=Published"
~~~

Resource generation is the main business-development workflow: generated resources are discovered by the admin panel automatically. The generator avoids overwriting manual files. Complex business flows belong in `backend/app/modules/<module>` and `admin/src/modules/<module>` rather than being forced into a generic Resource.

## Verification

~~~powershell
cd backend
go test ./...

cd ..\admin
pnpm exec vue-tsc --noEmit
pnpm run build
~~~

For a public SEO build:

~~~powershell
cd admin
pnpm run build:ssg
~~~

## Deployment

Docker deployment and Linux source deployment are provided under [`deploy/`](./deploy/) and documented in the [deployment guide](./docs/fastimg-deployment.md). Both deployment modes run only the FastImg application and require externally managed PostgreSQL and Redis.

~~~bash
cp deploy/docker/fastimg.env.example deploy/docker/fastimg.env
# Fill production APP_KEY, JWT_SECRET, APP_URL, CORS, PostgreSQL and Redis values.
deploy/docker/deploy.sh --check
deploy/docker/deploy.sh
~~~

The deployment scripts do not claim that real payment callbacks, TLS, isolated backup recovery, dependency scanning, object storage/CDN, or production load testing have passed. Review the [production gates](./docs/fastimg-production-gates.md) before public launch.

## Directory conventions

- Backend foundation: `backend/app/core`
- Backend business modules: `backend/app/modules/<module>`
- Backend services: `backend/app/services/<domain>`
- Backend console commands: `backend/app/console/<domain>`
- Frontend shared components: `admin/src/components`
- Frontend infrastructure: `admin/src/core`
- Frontend business pages and components: `admin/src/modules/<module>`

## Project status

FastImg is suitable for continued development and controlled integration testing. It is not declared formally production-ready until every target-environment gate is verified, including real payment callbacks, TLS/reverse proxy, backup and restore, dependency scanning, object storage/CDN, pressure testing, monitoring, and rollback.

Before production deployment, review secrets, database, Redis, reverse proxy, retention, backups, monitoring, migrations, provider credentials, and administrator bootstrap procedures for the target environment. This README is an entry point, not a substitute for environment acceptance.
