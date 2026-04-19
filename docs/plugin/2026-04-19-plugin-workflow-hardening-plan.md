# Plugin Workflow Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the plugin provider/reviewer workflow so release editing, compatibility modeling, work-order management, and system-side master data all follow enforceable business rules with Docker-backed real validation after each module.

**Architecture:** Keep the current plugin module and repository layout, but split responsibilities into four bounded slices: backend workflow/domain rules, provider-facing project detail UI, reviewer-facing work-order UI, and system-side master data management. Every slice must land with unit tests first, then module-level Docker verification, and finally an end-to-end Docker workflow covering create, submit, claim, approve/reject, release, and offline actions.

**Tech Stack:** Go + Gin + Gorm, React + Umi Max + Ant Design Pro, Jest/RTL, Go test, Docker Compose (PostgreSQL stack), Playwright CLI / HTTP verification.

---

## File Map

### Backend
- Modify: `backend/internal/modules/plugin/dto/dto.go`
- Modify: `backend/internal/modules/plugin/model/model.go`
- Modify: `backend/internal/modules/plugin/repository/plugin_repo.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service_test.go`
- Modify: `backend/internal/modules/plugin/router/router_test.go`
- Modify: `backend/cmd/server/seed_admin.go`
- Modify: `backend/internal/core/server/router.go` if new plugin master-data endpoints need wiring
- Create or modify error definitions in `backend/pkg/errcode/plugin_code.go`

### Frontend
- Modify: `front-end/src/services/api/plugin.ts`
- Modify: `front-end/src/pages/plugin/project-detail/index.tsx`
- Modify: `front-end/src/pages/plugin/project-detail/index.test.tsx`
- Modify: `front-end/src/pages/plugin/work-order-pool/index.tsx`
- Modify: `front-end/src/pages/plugin/work-order-pool/index.test.tsx`
- Create: `front-end/src/pages/plugin/work-order-detail/index.tsx`
- Create: `front-end/src/pages/plugin/work-order-detail/index.test.tsx`
- Modify: `front-end/src/pages/plugin/project-management/index.tsx`
- Modify: `front-end/src/pages/plugin/project-management/index.test.tsx`
- Modify: `front-end/src/pages/plugin/components/status.tsx`
- Modify: `front-end/src/utils/componentMap.tsx`
- Modify: `front-end/src/utils/componentMap.test.tsx`
- Modify: `front-end/config/routes.ts` only if static route fallback or public route glue changes are needed
- Modify: `front-end/src/locales/zh-CN/menu.ts`
- Modify: `front-end/src/locales/en-US/menu.ts`
- Create system-management pages if needed:
  - `front-end/src/pages/sys/plugin-master/index.tsx`
  - `front-end/src/pages/sys/plugin-master/index.test.tsx`

### Docs / Validation
- Modify: `docs/plugin/README.md`
- Append test evidence to `docs/plugin/2026-04-18-plugin-real-env-test-plan.md` or create a new report file if the flow changes materially

---

### Task 1: Harden Release Domain Rules In Backend

**Files:**
- Modify: `backend/internal/modules/plugin/dto/dto.go`
- Modify: `backend/internal/modules/plugin/model/model.go`
- Modify: `backend/internal/modules/plugin/repository/plugin_repo.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service_test.go`
- Modify: `backend/pkg/errcode/plugin_code.go`

- [ ] **Step 1: Write failing backend tests for release editing, versioning, and compatibility rules**

Add tests that cover:
- only draft/ready and rejected releases are editable by provider
- first release version must be `1.0.0`
- duplicate version is rejected
- non-increasing version is rejected
- compatibility is required
- compatibility is split into `product`, `acli`, and `universal`
- reviewer/admin cannot mutate release content through update API

Run: `go test ./internal/modules/plugin/service -run "Test(CreateRelease|UpdateRelease|TransitionRelease|GetWorkOrderPool)" -count=1`
Expected: FAIL on missing rules / missing fields

- [ ] **Step 2: Extend DTO/model shape for new compatibility structure and claimer display fields**

Implement:
- request/response DTOs for:
  - product compatibility
  - aCLI compatibility
  - universal support flag or compatibility type
- work-order response fields:
  - `claimerName`
  - `claimerUsername`
- department response fields:
  - `nameZh`
  - `nameEn`

- [ ] **Step 3: Add minimal repository support**

Implement:
- version lookup helpers by plugin
- work-order list preloading for claimer user
- optional full-status work-order query
- compatibility persistence that distinguishes `product`, `acli`, and `universal`

- [ ] **Step 4: Implement service-layer validation and permission rules**

Implement:
- release content editable only when provider owns release and release state is `ready` or `rejected`
- first version forced to `1.0.0`
- later versions must be semver and greater than max existing version
- compatibility required
- `request_offline` remains separate from release content editing
- work-order pool returns all statuses and supports `all`/`mine` filtering
- claimer display information returned to frontend

- [ ] **Step 5: Re-run backend tests**

Run: `go test ./internal/modules/plugin/... -count=1`
Expected: PASS

- [ ] **Step 6: Docker verify backend module before moving on**

Run:
- `docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build backend`
- login request for seed admin
- create/update/list release API checks
- work-order list API check for all statuses / mine tab semantics

Expected:
- container healthy
- invalid version / missing compatibility rejected
- work-order response contains claimer display fields

---

### Task 2: Rebuild Provider-Facing Project Detail UI

**Files:**
- Modify: `front-end/src/services/api/plugin.ts`
- Modify: `front-end/src/pages/plugin/project-detail/index.tsx`
- Modify: `front-end/src/pages/plugin/project-detail/index.test.tsx`
- Modify: `front-end/src/pages/plugin/project-management/index.tsx`
- Modify: `front-end/src/pages/plugin/project-management/index.test.tsx`

- [ ] **Step 1: Write failing frontend tests for provider workflow**

Add tests that cover:
- edit button hidden for non-editable release states
- release list search/filter works
- compatibility form requires at least one entry
- first version placeholder/help indicates `1.0.0`
- upload controls render for report/x86/arm instead of URL inputs
- department option renders according to locale

Run: `npm test -- --runInBand src/pages/plugin/project-detail/index.test.tsx src/pages/plugin/project-management/index.test.tsx`
Expected: FAIL on missing UI behavior

- [ ] **Step 2: Update API typings**

Implement:
- release compatibility groups
- upload-return payload fields
- localized department fields
- release editability flags if backend exposes them

- [ ] **Step 3: Implement provider detail UI changes**

Implement:
- searchable release list
- edit button only for editable states
- upload controls for test report and package files
- compatibility editor split into:
  - product compatibility
  - aCLI compatibility
  - universal support toggle
- inline version rule hints

- [ ] **Step 4: Re-run targeted frontend tests**

Run: `npm test -- --runInBand src/pages/plugin/project-detail/index.test.tsx src/pages/plugin/project-management/index.test.tsx`
Expected: PASS

- [ ] **Step 5: Docker verify provider UI before moving on**

Run:
- `docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build frontend backend`
- real provider flow in Docker:
  - open project detail
  - create first release
  - verify invalid version fails
  - verify upload-driven form works

Expected:
- provider can edit only draft/rejected releases
- uploaded file URLs are persisted and visible

---

### Task 3: Split Reviewer Work-Order Detail From Provider Detail

**Files:**
- Create: `front-end/src/pages/plugin/work-order-detail/index.tsx`
- Create: `front-end/src/pages/plugin/work-order-detail/index.test.tsx`
- Modify: `front-end/src/utils/componentMap.tsx`
- Modify: `front-end/src/utils/componentMap.test.tsx`
- Modify: `front-end/src/locales/zh-CN/menu.ts`
- Modify: `front-end/src/locales/en-US/menu.ts`

- [ ] **Step 1: Write failing tests for read-only reviewer detail**

Add tests that cover:
- reviewer detail page renders release content and timeline
- no editable form controls exist
- approve/reject/release/offline buttons only appear when status allows
- back navigation returns to work-order pool

Run: `npm test -- --runInBand src/pages/plugin/work-order-detail/index.test.tsx src/utils/componentMap.test.tsx`
Expected: FAIL because page/route do not exist yet

- [ ] **Step 2: Implement new reviewer detail page**

Implement:
- read-only release summary
- compatibility groups display
- file download section
- timeline
- action bar for review transitions only

- [ ] **Step 3: Wire route/component map**

Implement:
- component map entry
- hidden menu route if needed for dynamic routing

- [ ] **Step 4: Re-run targeted tests**

Run: `npm test -- --runInBand src/pages/plugin/work-order-detail/index.test.tsx src/utils/componentMap.test.tsx`
Expected: PASS

- [ ] **Step 5: Docker verify reviewer detail flow**

Run:
- reviewer login
- open work-order detail from pool
- verify fields are read-only
- verify transition buttons work only in valid states

Expected:
- reviewer page no longer shares provider editor

---

### Task 4: Upgrade Work-Order Pool Into Full Management Console

**Files:**
- Modify: `backend/internal/modules/plugin/dto/dto.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service_test.go`
- Modify: `front-end/src/pages/plugin/work-order-pool/index.tsx`
- Modify: `front-end/src/pages/plugin/work-order-pool/index.test.tsx`
- Modify: `front-end/src/services/api/plugin.ts`

- [ ] **Step 1: Write failing tests for tabs and search**

Add tests that cover:
- `all` and `mine` tabs
- keyword search
- status/request-type/process-status filters
- claimer display shows human-readable name
- all-status list rendered, not only pending subset

Run:
- backend: `go test ./internal/modules/plugin/service -run TestGetWorkOrderPool -count=1`
- frontend: `npm test -- --runInBand src/pages/plugin/work-order-pool/index.test.tsx`

Expected: FAIL on current filtered implementation

- [ ] **Step 2: Implement backend query/filter changes**

Implement:
- tab mode parameter
- all-status query default
- mine filter on current reviewer
- keyword filter for plugin code/name/version
- preload claimer user display name

- [ ] **Step 3: Implement frontend console UI**

Implement:
- tabbed console
- search form
- status chips / table filters
- claimer column with user display
- detail link to reviewer-only detail page

- [ ] **Step 4: Re-run targeted tests**

Run:
- `go test ./internal/modules/plugin/service -run TestGetWorkOrderPool -count=1`
- `npm test -- --runInBand src/pages/plugin/work-order-pool/index.test.tsx`

Expected: PASS

- [ ] **Step 5: Docker verify work-order console**

Run reviewer flow in Docker:
- create test releases
- submit for review
- verify `all` tab shows historical + current work orders
- verify `mine` tab changes after claim
- verify claimer shows actual person name

Expected:
- reviewer sees complete work-order management console

---

### Task 5: Add System-Side Plugin Master Data Management

**Files:**
- Modify: `backend/internal/modules/plugin/dto/dto.go`
- Modify: `backend/internal/modules/plugin/repository/plugin_repo.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service.go`
- Modify: `backend/internal/modules/plugin/service/plugin_service_test.go`
- Modify: `backend/cmd/server/seed_admin.go`
- Create or modify UI:
  - `front-end/src/pages/sys/plugin-master/index.tsx`
  - `front-end/src/pages/sys/plugin-master/index.test.tsx`
- Modify: `front-end/src/utils/componentMap.tsx`
- Modify: `front-end/src/locales/zh-CN/menu.ts`
- Modify: `front-end/src/locales/en-US/menu.ts`

- [ ] **Step 1: Write failing tests for localized departments / compatibility master data**

Add tests that cover:
- department list returns zh/en names
- sys page can manage departments and compatibility entries
- plugin project modal uses localized department labels

Run:
- `go test ./internal/modules/plugin/service -run "Test(GetDepartmentList|CreateProduct|UpdateProduct)" -count=1`
- `npm test -- --runInBand src/pages/sys/plugin-master/index.test.tsx src/pages/plugin/project-management/index.test.tsx`

Expected: FAIL

- [ ] **Step 2: Implement backend master-data changes**

Implement:
- localized department fields
- compatibility entry model semantics suitable for plugin release form
- admin-only CRUD where needed

- [ ] **Step 3: Implement system management page**

Implement:
- menu under system management
- tabs or sections for:
  - departments
  - compatibility master data

- [ ] **Step 4: Re-run targeted tests**

Run:
- `go test ./internal/modules/plugin/service -run "Test(GetDepartmentList|CreateProduct|UpdateProduct)" -count=1`
- `npm test -- --runInBand src/pages/sys/plugin-master/index.test.tsx src/pages/plugin/project-management/index.test.tsx`

Expected: PASS

- [ ] **Step 5: Docker verify admin master-data flow**

Run admin flow in Docker:
- manage departments in both locales
- verify provider modal reflects latest localized values

Expected:
- system management owns plugin base data

---

### Task 6: End-To-End Docker Regression And Documentation

**Files:**
- Modify: `docs/plugin/README.md`
- Modify or create: `docs/plugin/2026-04-19-plugin-workflow-hardening-test-report.md`

- [ ] **Step 1: Execute full Docker rebuild**

Run: `docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build backend frontend`
Expected: healthy backend/frontend containers

- [ ] **Step 2: Run full backend and frontend verification**

Run:
- `go test ./internal/modules/plugin/... -count=1`
- `npm run tsc`
- `npm test -- --runInBand src/pages/plugin/project-management/index.test.tsx src/pages/plugin/project-detail/index.test.tsx src/pages/plugin/work-order-pool/index.test.tsx src/pages/plugin/work-order-detail/index.test.tsx src/pages/sys/plugin-master/index.test.tsx src/utils/componentMap.test.tsx`

Expected: PASS

- [ ] **Step 3: Run real Docker end-to-end workflow**

Verify in Docker:
- provider creates plugin project
- first release must be `1.0.0`
- provider uploads files
- provider submits for review
- reviewer sees work order in `all`
- reviewer claims work order, then sees it in `mine`
- reviewer detail page is read-only
- reviewer approves or rejects
- provider edits only rejected release
- reviewer releases approved version
- provider requests offline for released version
- reviewer offlines request

- [ ] **Step 4: Record evidence**

Save:
- commands run
- API samples
- screenshots or concrete route checks
- any residual risk

- [ ] **Step 5: Commit**

Run:
- `git add ...`
- `git commit -m "feat: harden plugin workflow and review console"`

Expected: clean working tree except intentionally deferred items

---

## Self-Review

- Spec coverage:
  - release editability restrictions: Task 1 + Task 2
  - upload-based release files: Task 2
  - compatibility split/universal: Task 1 + Task 2
  - version sequencing: Task 1
  - release search: Task 2
  - localized department + system management: Task 5
  - work-order all/mine tabs + search + claimer display: Task 4
  - separate reviewer detail page: Task 3
  - Docker validation after each module: Tasks 1-5 each end with Docker verification
- Placeholder scan: no `TODO`/`TBD` steps left.
- Type consistency: plan uses `product compatibility`, `aCLI compatibility`, `universal support`, `claimerName`, `claimerUsername`, and distinct `work-order-detail` route consistently.

## Execution Handoff

Execution choice is fixed by user instruction for this thread:

- Use **Subagent-Driven** execution.
- After each module, stop only long enough to run Docker validation and capture evidence, then continue automatically.
