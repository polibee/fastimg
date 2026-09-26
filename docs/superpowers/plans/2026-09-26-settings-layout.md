# FastImg Admin Settings Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reorganize `/admin/settings` into understandable business groups and make every input, select, switch, and multiline field aligned, responsive, bilingual, and safe for provider-specific secrets.

**Architecture:** Keep the existing settings API and differential-save helper as the single persistence path. Refactor only the settings page presentation and field metadata: a declarative ordered group model drives rendering, a shared field renderer composes existing shadcn-vue `Field`, `Input`, `Textarea`, `Select`, `Checkbox`, and `Switch`, and provider visibility remains a UI concern without clearing hidden values. Backups stay on `/admin/backups` and are not mixed into this form.

**Tech Stack:** Vue 3, TypeScript, vue-i18n, existing `buildSettingUpdates`, shadcn-vue/Radix-style components in `admin/src/components/ui`, Vite, and existing frontend contract tests.

**Spec:** `docs/superpowers/specs/2026-09-26-settings-layout-design.md`

## Global Constraints

- Preserve the existing `/api/v1/admin/settings` and `/api/v1/admin/settings/{key}` contract and differential save semantics.
- Use only the existing shadcn-vue primitives; remove raw native `<select>` elements from the settings page.
- All settings fields use `Field`, `FieldLabel`, `FieldDescription`, and `FieldGroup`; controls have consistent width/height and do not overflow on narrow screens.
- Secret inputs never display plaintext; blank secret values preserve the stored value through the existing `__configured__` convention.
- Switching an email provider changes visibility only and never clears another provider’s saved values.
- The fixed group order is: site basics, registration/login, email, upload/media, SEO/discovery, payments/orders, statistics, custom code, other.
- All new or changed text has identical key trees in `admin/src/locales/zh-CN/` and `admin/src/locales/en-US/`.

## Review Focus

- A provider switch must not submit or clear hidden credentials; covered by the settings component contract test in Task 3.
- A select loaded from the API must show the current value and save the changed value; covered by the select model-binding test in Task 2.
- A boolean must be keyboard accessible and aligned with its label/description rather than rendered as a stray native checkbox; covered by the control contract test in Task 2.
- Long code/text fields must span the form width without breaking the two-column grid; covered by responsive class assertions in Task 2 and the build in Task 4.
- A failed save must identify the group/field while preserving unsaved values; covered by the save regression test in Task 3.

---

### Task 1: Ordered settings metadata and locale contract

**Files:**
- Create: `admin/src/modules/settings/settings-schema.ts`
- Create: `admin/tests/settings-layout.test.mjs`
- Modify: `admin/src/modules/settings/pages/AdminSettingsPage.vue`
- Modify: `admin/src/locales/zh-CN/settings.json`
- Modify: `admin/src/locales/en-US/settings.json`
- Modify: `admin/src/i18n/index.ts`
- Modify: `admin/src/i18n/index.js`

**Interfaces:**
- Produces `SettingGroupDefinition[]` in the required nine-group order and a typed field metadata lookup consumed by the page.
- Consumes the existing `SystemSetting` keys and `buildSettingUpdates`; it must not introduce a second settings persistence model.

- [ ] **Step 1: Write failing tests** asserting the exact group order, every declared field has a locale label/description in both languages, and the page no longer uses native `<select>`.
- [ ] **Step 2: Run the settings contract test to verify failure**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: FAIL because the ordered schema and locale keys are not yet present.

- [ ] **Step 3: Implement `settings-schema.ts`** with the nine ordered groups, field types (`text`, `number`, `secret`, `boolean`, `select`, `textarea`), grid span, provider visibility metadata, and stable translation keys.
- [ ] **Step 4: Move existing field lists into the schema and add missing group labels/descriptions/hints** in zh-CN and en-US without changing backend setting keys.
- [ ] **Step 5: Run the contract test**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: PASS for order and locale parity.

- [ ] **Step 6: Commit metadata and copy**

```bash
git add admin/src/modules/settings/settings-schema.ts admin/src/modules/settings/pages/AdminSettingsPage.vue admin/src/locales admin/src/i18n admin/tests/settings-layout.test.mjs
git commit -m "refactor: define ordered admin settings schema"
```

### Task 2: Shared aligned control rendering with shadcn-vue

**Files:**
- Modify: `admin/src/modules/settings/pages/AdminSettingsPage.vue`
- Modify: `admin/src/components/ui/field/Field.vue` only if an existing primitive lacks the required layout slot
- Modify: `admin/src/components/ui/select/*` only if required existing Select behavior is missing
- Modify: `admin/src/components/ui/switch/*` only if required existing Switch behavior is missing
- Modify: `admin/tests/settings-layout.test.mjs`

**Interfaces:**
- Produces a single rendering path for setting fields: `Field` + label + control + description, with `Select` for choices and `Switch`/`Checkbox` for booleans.
- Preserves existing helpers `isSecret`, `fieldValue`, `setValue`, `setBoolean`, `emailFieldVisible`, `gatewayURLFields`, and `buildSettingUpdates`.

- [ ] **Step 1: Extend failing tests** to assert `SelectTrigger`, `SelectValue`, `SelectContent`, `SelectItem`, full-width control classes, aligned boolean rows, and no raw `<select>`/unlabelled checkbox branches.
- [ ] **Step 2: Run the focused test to verify failure**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: FAIL against the current native selects and duplicated field branches.

- [ ] **Step 3: Replace native selects** for email provider, SMTP encryption, and PayPal environment with existing shadcn-vue `Select` components using `model-value`/`update:model-value` and translated option labels.
- [ ] **Step 4: Replace boolean branches** with an aligned horizontal control row using the existing `Switch` or `Checkbox`, with a stable `id`, label association, and description column.
- [ ] **Step 5: Render text/number/secret/textarea fields from the schema** with consistent `w-full`, `min-w-0`, height, and `md:col-span-2` behavior; keep provider-specific gateway fields in their existing provider sections.
- [ ] **Step 6: Run the focused test**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: PASS for shadcn controls, alignment classes, and no native select.

- [ ] **Step 7: Commit the shared form renderer**

```bash
git add admin/src/modules/settings/pages/AdminSettingsPage.vue admin/src/components/ui/field admin/src/components/ui/select admin/src/components/ui/switch admin/tests/settings-layout.test.mjs
git commit -m "refactor: align admin settings controls"
```

### Task 3: Provider visibility, differential save, and error regression

**Files:**
- Modify: `admin/src/modules/settings/pages/AdminSettingsPage.vue`
- Modify: `admin/src/modules/settings/save.ts` only if the existing helper needs a typed preservation fix
- Modify: `admin/tests/settings-layout.test.mjs`

**Interfaces:**
- Produces tested provider-specific visibility for SMTP, Aliyun, and Resend; payment sections remain independent.
- Consumes `initialValues` and `buildSettingUpdates`; blank secrets continue to produce `__configured__` rather than an empty overwrite.

- [ ] **Step 1: Write failing save/visibility tests** for switching providers, leaving a configured secret blank, changing a select, and rendering a field-specific save error.
- [ ] **Step 2: Run the regression test to verify failure**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: FAIL if the refactor submits hidden values, loses configured secrets, or loses field context.

- [ ] **Step 3: Implement provider visibility from the schema**. Do not mutate hidden provider values and do not include hidden-field clearing in the update list.
- [ ] **Step 4: Keep differential-save behavior** and ensure selects update `values` through the same model path as text inputs; preserve the current field label in save errors.
- [ ] **Step 5: Run the regression test**

Run: `node --test admin/tests/settings-layout.test.mjs`

Expected: PASS with hidden credentials preserved and changed fields submitted exactly once.

- [ ] **Step 6: Commit provider/save regression fixes**

```bash
git add admin/src/modules/settings/pages/AdminSettingsPage.vue admin/src/modules/settings/save.ts admin/tests/settings-layout.test.mjs
git commit -m "fix: preserve provider settings during differential save"
```

### Task 4: Responsive visual verification, i18n parity, and documentation

**Files:**
- Modify: `admin/tests/settings-layout.test.mjs`
- Modify: `docs/fastimg-frontend-design.md`
- Modify: `docs/fastimg-stage-development-plan.md`
- Modify: `docs/README.md`

- [ ] **Step 1: Run the frontend contract and i18n checks**

Run: `node --test admin/tests/settings-layout.test.mjs admin/tests/fastimg-i18n.test.mjs`

Expected: PASS with identical zh-CN/en-US settings key trees.

- [ ] **Step 2: Run TypeScript/build verification**

Run: `npm run build`

Expected: PASS without Vue template, Select, or locale-loader errors.

- [ ] **Step 3: Visually verify `/admin/settings`** at desktop and narrow viewport**. Confirm all controls share left/right boundaries, code textareas span both columns, provider sections do not overflow, and save/error alerts remain readable.
- [ ] **Step 4: Document the fixed group order, shadcn control rule, and backup page separation** in the frontend design and stage plan.
- [ ] **Step 5: Commit the verification/documentation slice**

```bash
git add admin/tests/settings-layout.test.mjs docs/fastimg-frontend-design.md docs/fastimg-stage-development-plan.md docs/README.md
git commit -m "docs: record admin settings layout contract"
```
