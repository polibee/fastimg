# FastImg Backup and Restore Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a permissioned administrator backup and migration workflow that exports FastImg business data, non-sensitive settings, and media into a validated archive and restores it only to a confirmed new-server instance.

**Architecture:** A dedicated `backups` module owns HTTP controllers and view contracts; `backend/app/services/backups` owns archive, database-dump, validation, and restore orchestration. The feature is not a Resource CRUD. Existing audit, RBAC, database, storage, maintenance, and configuration services remain the authority for their domains. The first release exposes a staged `upload -> validate -> preview -> confirm -> restore` flow and explicitly does not overwrite an existing production instance from a browser.

**Tech Stack:** Go 1.25, Goravel 1.18, PostgreSQL `pg_dump`/`pg_restore` through fixed executable paths and structured arguments, `github.com/klauspost/compress/zstd` for in-process `.tar.zst` writing/reading, existing FastImg `StorageProvider`, PostgreSQL migrations, Vue 3, Vue Router, shadcn-vue, generated API client, vue-i18n, and existing audit/RBAC middleware.

**Spec:** `docs/superpowers/specs/2026-09-26-backup-restore-design.md`

## Global Constraints

- Only administrators with `admin.backups.manage` may create, validate, restore, or delete backups; `admin.backups.download` may be granted separately for downloads.
- The archive format is `fastimg-backup-<timestamp>-<id>.tar.zst` with `manifest.json`, `database.dump`, `settings.json`, `storage/media/...`, `storage/variants/...`, and `checksums.sha256`.
- APP_KEY, JWT/DB/Redis credentials, email credentials, Turnstile secrets, payment secrets, and other secret settings are never exported; secret fields remain blank/configured placeholders after restore.
- Archive input is untrusted: enforce upload size, entry count, expansion size, per-file size, allowed names, canonical paths, duplicate rejection, and symlink/hardlink rejection before extraction.
- Database tools use fixed executable identity, structured arguments, bounded context cancellation, and a minimal inherited environment; never build shell command strings from request data.
- Compression is performed in-process with one pinned zstd library dependency; the HTTP path does not invoke `tar`, `zstd`, or a shell command for archive input.
- Restore mode is fixed to `new_server`; no browser-triggered replacement of an existing site is implemented in this phase.
- Every create, download, delete, validate, restore start, restore success, and restore failure action writes an audit record without archive contents or secrets.
- All user-visible text is maintained in `admin/src/locales/zh-CN/` and `admin/src/locales/en-US/` with identical key trees.
- Do not touch `backend/storage/fastimg`, database data, or Laragon PostgreSQL/Redis configuration outside the explicit feature path and authorized local migration.

## Review Focus

- A crafted archive path such as `../`, an absolute path, a Windows drive path, duplicate names, or a symlink must be rejected before extraction; covered by the archive validator tests in Task 2.
- A valid archive containing a secret setting or unexpected executable must not restore it; covered by manifest/settings filtering tests in Task 2 and Task 4.
- A missing or incompatible `pg_dump`/`pg_restore` binary must produce a classified failed job and leave the instance writable; covered by tool-runner tests in Task 3 and restore failure tests in Task 4.
- An interrupted restore must not report `restored` or leave maintenance mode permanently enabled; covered by restore orchestration tests in Task 4.
- A user without backup permissions must receive authorization failure and no archive metadata; covered by controller contract tests in Task 1 and Task 5.

---

### Task 1: Backup job contract, migration, RBAC, and admin route boundary

**Files:**
- Create: `backend/app/models/backup_job.go`
- Create: `backend/database/migrations/20260926000004_create_backup_jobs_table.go`
- Modify: `backend/bootstrap/migrations.go`
- Modify: `backend/database/seeders/admin_user.go`
- Modify: `backend/database/seeders/admin_user_test.go`
- Create: `backend/app/modules/backups/controllers/backup_controller.go`
- Modify: `backend/routes/web.go`
- Create: `backend/app/modules/backups/controllers/backup_controller_test.go`
- Modify: `backend/app/openapi/spec.go`

**Interfaces:**
- Produces `BackupJob` persistence with `id`, `kind`, `status`, `file_name`, `storage_path`, `manifest_json`, `size_bytes`, `error_code`, `created_by`, timestamps, and a unique active-job constraint.
- Produces controller contracts for `GET/POST /api/v1/admin/backups`, `GET /{id}`, `GET /{id}/download`, `DELETE /{id}`, `POST /validate`, and `POST /restore`.
- Produces permissions `admin.backups.manage` and `admin.backups.download` and an `/admin/backups` route protected by the manage permission.

- [ ] **Step 1: Write failing migration, permission, and route tests** asserting the backup table fields, both permission names, method/path registration, and authorization boundary.
- [ ] **Step 2: Run the focused backend tests**

Run: `go test ./database/migrations ./database/seeders ./app/modules/backups/controllers`

Expected: FAIL because the model, migration, controller, and routes do not exist.

- [ ] **Step 3: Implement the model, migration registration, idempotent permission seeding, controller stubs, and protected routes**. Keep the controller dependent on a service interface; do not put archive or database logic in the controller.
- [ ] **Step 4: Run the focused tests again**

Run: `go test ./database/migrations ./database/seeders ./app/modules/backups/controllers`

Expected: PASS, with unauthorized requests rejected before any backup metadata is returned.

- [ ] **Step 5: Commit the contract boundary**

```bash
git add backend/app/models/backup_job.go backend/database/migrations/20260926000004_create_backup_jobs_table.go backend/bootstrap/migrations.go backend/database/seeders/admin_user.go backend/database/seeders/admin_user_test.go backend/app/modules/backups backend/routes/web.go backend/app/openapi/spec.go
git commit -m "feat: add backup restore admin contract"
```

### Task 2: Secure archive manifest, settings filtering, and media snapshot

**Files:**
- Create: `backend/app/services/backups/manifest.go`
- Create: `backend/app/services/backups/archive.go`
- Create: `backend/app/services/backups/archive_security_test.go`
- Create: `backend/app/services/backups/settings_export.go`
- Create: `backend/app/services/backups/settings_export_test.go`
- Modify: `backend/app/services/settings/secret.go`
- Modify: `backend/app/services/storage/provider.go` only if a read-only object enumeration boundary is required

**Interfaces:**
- Produces `Manifest` and `BuildArchive(ctx context.Context, jobID uint64, destination string) (Manifest, error)`.
- Produces `ValidateArchive(ctx context.Context, source io.ReaderAt, size int64) (ValidationPreview, error)`.
- Produces an allowlisted settings export that returns non-sensitive values and records excluded secret keys without values.
- Consumes the existing `StorageProvider` and media/variant records; it must not infer file paths from user input.

- [ ] **Step 1: Write failing tests** for path traversal, absolute/drive paths, symlinks, hardlinks, duplicate entries, unexpected executables, archive expansion limits, checksum mismatch, and secret-setting exclusion.
- [ ] **Step 2: Run the security tests to verify failure**

Run: `go test ./app/services/backups -run 'Test(Archive|Settings)' -v`

Expected: FAIL because the validator and export functions are absent.

- [ ] **Step 3: Implement canonical archive validation and bounded extraction**. Use `filepath.Clean` plus a trusted destination root check, reject links before extraction, allow only the exact manifest/database/settings/checksum names and `storage/media`/`storage/variants` prefixes, and enforce configured byte/entry limits before writing.
- [ ] **Step 4: Implement manifest/checksum creation and settings filtering**. Reuse the existing secret-key classification; never serialize secret values, tokens, raw credentials, or archive contents into errors or audits. Snapshot media through database-owned object keys and the storage boundary.
- [ ] **Step 5: Run the security tests**

Run: `go test ./app/services/backups -run 'Test(Archive|Settings)' -v`

Expected: PASS, including malformed archive rejection without writing outside the staging directory.

- [ ] **Step 6: Commit the archive boundary**

```bash
git add backend/app/services/backups backend/app/services/settings/secret.go backend/app/services/storage/provider.go
git commit -m "feat: add secure FastImg backup archive boundary"
```

### Task 3: Database dump, asynchronous backup jobs, listing, download, and delete

**Files:**
- Create: `backend/app/services/backups/database_tool.go`
- Create: `backend/app/services/backups/backup_service.go`
- Create: `backend/app/services/backups/backup_service_test.go`
- Modify: `backend/app/modules/backups/controllers/backup_controller.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/core/admin/controllers/audit_controller.go` only if the existing audit helper needs a typed backup category
- Modify: `backend/.env.example`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

**Interfaces:**
- Produces `BackupService.Create(ctx, operatorID uint64) (BackupJob, error)`, `List`, `Show`, `Download`, `Delete`, and `ValidateUpload` methods.
- Produces a fixed `DatabaseTool` interface so tests can use a fake and production can resolve configured PostgreSQL client binaries without shell interpretation.
- Produces job states `queued -> running -> ready|failed` and no download before `ready`.

- [ ] **Step 1: Write failing service tests** for idempotent active-job rejection, status transitions, fixed tool arguments, secret-free logs, download authorization, and delete cleanup.
- [ ] **Step 2: Run the focused service tests to verify failure**

Run: `go test ./app/services/backups -run 'TestBackupService|TestDatabaseTool' -v`

Expected: FAIL because the service and tool boundary are absent.

- [ ] **Step 3: Add and verify the pinned zstd dependency, then implement the fixed database-tool runner** using `exec.CommandContext`, an allowlisted executable path, structured arguments (`--no-owner`, `--no-privileges`, fixed database connection fields), bounded timeout, and sanitized environment. Reuse the existing local backup script’s tool discovery rules only as operational documentation; the HTTP service must remain native Go orchestration.
- [ ] **Step 4: Implement asynchronous job creation and archive assembly** with a temporary staging directory, atomic final rename, SHA-256 checksum generation, manifest persistence, and audit events. Use the existing queue/runtime boundary when available; otherwise provide a bounded in-process worker with one active backup job and a restart-visible `failed` state rather than pretending work survived a process restart.
- [ ] **Step 5: Implement authenticated streaming download and safe deletion**. Keep files outside public storage, prevent path selection from request data, and delete only records owned by the backup service.
- [ ] **Step 6: Run service and controller tests**

Run: `go test ./app/services/backups ./app/modules/backups/controllers -v`

Expected: PASS with no secret values in responses, audits, or error strings.

- [ ] **Step 7: Commit the backup job slice**

```bash
git add backend/app/services/backups backend/app/modules/backups/controllers backend/routes/web.go backend/.env.example
git commit -m "feat: add FastImg backup jobs and downloads"
```

### Task 4: Upload validation, preview, new-server restore, maintenance, and recovery

**Files:**
- Create: `backend/app/services/backups/restore_service.go`
- Create: `backend/app/services/backups/restore_service_test.go`
- Modify: `backend/app/modules/backups/controllers/backup_controller.go`
- Create: `backend/app/services/production/maintenance.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/services/audit/audit_service.go` only for typed backup action labels if required

**Interfaces:**
- Produces `RestoreService.Validate(ctx, upload io.Reader, size int64) (RestorePreview, error)` and `Restore(ctx, RestoreRequest) (BackupJob, error)`.
- `RestoreRequest` requires `jobID`, `confirmation == "RESTORE_FASTIMG_BACKUP"`, and `mode == "new_server"`.
- Consumes Task 2 validation and Task 3 database/storage/tool boundaries; produces `uploaded -> validating -> validated|invalid -> restoring -> restored|restore_failed`.

- [ ] **Step 1: Write failing restore tests** for confirmation phrase, fixed mode, incompatible schema/app version, non-empty target refusal, invalid checksum, secret placeholder clearing, maintenance release on tool failure, and success/failure state persistence.
- [ ] **Step 2: Run restore tests to verify failure**

Run: `go test ./app/services/backups -run 'TestRestore' -v`

Expected: FAIL because restore orchestration is absent.

- [ ] **Step 3: Implement upload staging and preview**. Store uploads in a private temporary root, validate without executing or copying archive entries into live storage, return only manifest counts/bytes/compatibility/secrets-excluded summary, and require the explicit confirmation phrase.
- [ ] **Step 4: Implement new-server preflight**. Require a clean/approved target state, acquire the maintenance lock, reject new uploads/payments/queue writes while restoring, and create a recoverable target snapshot or isolated restore target before changing live state.
- [ ] **Step 5: Implement structured restore and post-restore checks**. Restore database with fixed tool arguments, copy media through the storage boundary, run migrations/consistency checks, clear secret placeholders, persist final state, and always release maintenance in a deferred cleanup path.
- [ ] **Step 6: Run restore tests and a local dry-run against the authorized `fastimg_dev` database only**

Run: `go test ./app/services/backups -run 'TestRestore' -v`

Expected: PASS; local dry-run must use a temporary staging directory and must not overwrite the current development database without an explicit migration/restore confirmation.

- [ ] **Step 7: Commit the restore slice**

```bash
git add backend/app/services/backups backend/app/modules/backups/controllers backend/app/services/production backend/app/services/audit/audit_service.go backend/routes/web.go
git commit -m "feat: add validated new-server restore workflow"
```

### Task 5: Admin backup page, API client, menu, and bilingual UI

**Files:**
- Create: `admin/src/modules/backups/pages/AdminBackupsPage.vue`
- Create: `admin/src/modules/backups/backup-api.ts`
- Create: `admin/tests/admin-backups.test.mjs`
- Modify: `admin/src/apps/admin/routes.ts`
- Modify: `admin/src/core/layouts/AdminShell.vue`
- Modify: `admin/src/generated/api.ts`
- Create: `admin/src/locales/zh-CN/backups.json`
- Create: `admin/src/locales/en-US/backups.json`
- Modify: `admin/src/i18n/index.ts`
- Modify: `admin/src/i18n/index.js`

**Interfaces:**
- Produces typed client methods for list/create/status/download/delete/validate/restore and maps backend error codes to human-readable messages.
- Produces `/admin/backups` with two explicit areas: “创建备份/下载” and “迁移恢复”; no backup form is embedded in `/admin/settings`.

- [ ] **Step 1: Write failing route/i18n/UI contract tests** for the `/admin/backups` route, permission-gated menu item, download response handling, file upload field, confirmation phrase, no secret rendering, and zh/en key parity.
- [ ] **Step 2: Run the frontend contract test to verify failure**

Run: `node --test admin/tests/admin-backups.test.mjs`

Expected: FAIL because the route, client, page, and locale namespace are absent.

- [ ] **Step 3: Implement the typed API client and route/menu**. Use the existing `apiFetch`, `apiDownload`, auth store, and admin permission checks; do not add a generic CRUD manifest for backups.
- [ ] **Step 4: Implement the page** with status table, create/download/delete actions, multipart upload, validation preview, explicit confirmation dialog, progress/error states, and accessible responsive layout using existing shadcn-vue components.
- [ ] **Step 5: Add synchronized zh-CN/en-US text** for status, warnings, backup contents, secret exclusions, restore failure, and maintenance mode.
- [ ] **Step 6: Run frontend tests and build**

Run: `node --test admin/tests/admin-backups.test.mjs`; `npm run build`

Expected: PASS and a production build without new TypeScript or route errors.

- [ ] **Step 7: Commit the admin UI slice**

```bash
git add admin/src/modules/backups admin/src/apps/admin/routes.ts admin/src/core/layouts/AdminShell.vue admin/src/generated/api.ts admin/src/i18n admin/src/locales admin/tests/admin-backups.test.mjs
git commit -m "feat: add admin backup restore workspace"
```

### Task 6: Documentation, migration review, and end-to-end acceptance

**Files:**
- Create: `docs/fastimg-backup-restore.md`
- Modify: `docs/README.md`
- Modify: `docs/fastimg-release-runbook.md`
- Modify: `docs/fastimg-stage-development-plan.md`
- Modify: `docs/superpowers/sdd/` progress ledger selected by the repository plan
- Test: `backend/app/services/backups/*_test.go`, `admin/tests/admin-backups.test.mjs`

- [ ] **Step 1: Document operator workflow** for source backup, local download, target upload/validate/restore, secret re-entry, failure recovery, retention, and the fact that browser restore is `new_server` only.
- [ ] **Step 2: Review and apply only the authorized local migration** to `fastimg_dev`, verify migration status and table shape, and record exact command/result.
- [ ] **Step 3: Run backend and frontend verification**

Run: `go test ./...`; `npm run build`; `node --test admin/tests/admin-backups.test.mjs`; `scripts/fastimg-disk-audit.ps1`

Expected: all focused and full tests pass; disk audit remains within the repository budget; no media directory is deleted.

- [ ] **Step 4: Perform an end-to-end local archive/validate/restore rehearsal** using a temporary target database or isolated target instance, not the active development database, and verify restored user/plan/media/non-sensitive settings plus empty secrets.
- [ ] **Step 5: Update the stage ledger with implemented/verified/blocked status** and commit the documentation and acceptance evidence.

```bash
git add docs
git commit -m "docs: document FastImg backup restore operations"
```
