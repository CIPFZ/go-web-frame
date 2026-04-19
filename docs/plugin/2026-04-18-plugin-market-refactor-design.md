# Plugin Market / CMS Refactor Design

Date: 2026-04-18  
Branch: `codex/plugin-market-refactor`

## 1. Goal

This refactor has two goals:

1. Fix the current plugin module defects in the CMS.
2. Rebuild the public plugin market so it can later be deployed independently from the CMS, while still staying in the same repository for now.

The public market is not treated as a temporary CMS sub-page anymore. It is treated as a lightweight public-facing application boundary with its own routes, layout rules, and backend read APIs.

## 2. Scope

### In Scope

1. Fix anonymous access for `/plugins` and `/plugins/:id`.
2. Rebuild the public market UI using the old plugin market interaction model as the primary reference.
3. Refactor CMS plugin admin pages using the old project center / project detail interaction model as the reference.
4. Fix plugin menu i18n for both Chinese and English.
5. Make plugin name / description / changelog display follow current UI language in CMS.
6. Add a reliable back affordance in plugin detail pages.
7. Improve plugin detail layout so long release lists do not push the detail area and timeline to the bottom.
8. Fix garbled seed/demo data and ensure plugin-related base data no longer renders as `???`.
9. Keep public APIs and admin APIs clearly separated for future extraction.
10. Add or update tests and rerun Docker-based real-environment validation.

### Out of Scope

1. Creating a new repository.
2. Fully splitting deployment in this round.
3. Changing the current frontend/backend stack.
4. Reworking plugin workflow semantics already validated in the previous round.

## 3. Architecture Direction

### 3.1 Repository Strategy

The repository stays unified.

The separation is done through feature boundaries:

1. CMS admin side remains inside the current admin application.
2. Plugin market side is rebuilt as a logically independent public app area.
3. Backend public plugin APIs remain in the current backend, but are treated as a future-extractable lightweight service boundary.

### 3.2 Future Deployment Direction

After this refactor, the code should support a later split into:

1. `cms-admin-web`
2. `plugin-market-web`
3. `cms-admin-api`
4. `plugin-market-api`

This round does not physically split those services, but the code must stop coupling market pages to CMS login and layout assumptions.

## 4. Problem Summary

### 4.1 Public Market Access

Current public routes exist, but anonymous access is blocked by global login guards.

Result:

1. `/plugins` redirects to login.
2. `/plugins/:id` redirects to login.

### 4.2 Menu i18n

The CMS menu shows raw locale keys such as `menu.plugin` and `menu.plugin.projectManagement` because locale entries are missing.

### 4.3 Garbled Data

Some seed/demo/plugin texts are already garbled in source and persisted data, which causes `???` in the UI.

### 4.4 CMS Language Display

Plugin project tables and detail pages do not consistently switch between Chinese and English fields based on current UI language.

### 4.5 Detail Navigation

Plugin detail pages do not provide a clear and safe back affordance.

### 4.6 Detail Layout

Current project detail layout is vertically stacked, so a long release list pushes current release information and timeline too far down.

## 5. Information Architecture

### 5.1 CMS Admin Area

Routes:

1. `/plugin/project-management`
2. `/plugin/project/:id`
3. `/plugin/work-order-pool`

Purpose:

1. Internal plugin project management
2. Release workflow management
3. Review / release workbench flow

### 5.2 Public Market Area

Routes:

1. `/plugins`
2. `/plugins/:id`

Purpose:

1. Public plugin discovery
2. Public plugin detail and version browsing
3. Download entry exposure
4. Bilingual presentation

Public market routes must be anonymous and must not depend on CMS shell state.

## 6. Frontend Refactor Plan

### 6.1 Public Market List

The new public list page will follow the old `market/index.tsx` direction:

1. Hero section
2. Search box
3. Category and architecture filters
4. Language toggle
5. Sort selector
6. Card grid

Each card shows:

1. Display name in current language
2. Version badge
3. Short description in current language
4. Architecture tags
5. Entry to detail page

### 6.2 Public Market Detail

The new public detail page will follow the old `market/detail.tsx` direction:

1. Top breadcrumb / back affordance
2. Header summary card
3. Main content tabs
4. Right-side download panel
5. Version switcher

Tabs should cover:

1. Overview
2. Changelog
3. Test / report / performance
4. History

### 6.3 CMS Project Management

Current table-first implementation will be upgraded using the old `project-center` ideas:

1. Stronger filter bar
2. Card/list dual view
3. Current-language display
4. More compact status summary
5. Faster entry into project detail

The existing CRUD and API contracts remain unless a specific mismatch must be corrected.

### 6.4 CMS Project Detail

Current detail page will be restructured into a stable split layout:

1. Top header with explicit back action
2. Left column: release list / release selector
3. Right column: active release detail and workflow
4. Timeline stays visible without being pushed below large lists

This keeps the workflow-oriented nature from the old design while preserving the validated current backend flow.

## 7. Language Strategy

### 7.1 CMS

CMS follows current app locale:

1. Chinese locale uses `nameZh`, `descriptionZh`, `changelogZh`
2. English locale uses `nameEn`, `descriptionEn`, `changelogEn`
3. If English is empty, fallback to Chinese

This rule applies to:

1. Project management cards/tables
2. Project detail header
3. Release detail blocks
4. Public market content when market locale is English

### 7.2 Public Market

Public market keeps an explicit Chinese / English switch, but must also work sensibly with default locale.

Priority:

1. User-selected market language
2. Fallback to current app locale or browser language
3. English fields fallback to Chinese when missing

## 8. Public Route and Layout Rules

### 8.1 Anonymous Allowlist

Global auth guards must explicitly exempt:

1. `/plugins`
2. `/plugins/:id`

### 8.2 Layout Isolation

Public market pages must not rely on:

1. Logged-in current user
2. Dynamic menu fetch
3. Admin menu shell
4. Admin-only redirects

They should render even when there is no token and no initial CMS state.

## 9. Backend Plan

### 9.1 Public/Admin Boundary

The backend keeps current plugin module structure, but the separation must become clearer:

1. admin-oriented logic remains under plugin private routes
2. public-oriented logic remains under `/api/v1/plugin/public/*`

This public surface should remain lightweight:

1. published plugin list
2. published plugin detail
3. current version / history view data
4. download-facing fields
5. bilingual display fields

### 9.2 Data Cleanup

Need to fix:

1. garbled plugin seed data in source
2. plugin-related Chinese copy in seed setup
3. existing local test/demo data in the running database as needed for verification

### 9.3 Error Semantics

Where possible, public interfaces should return cleaner public-facing results instead of generic internal errors for removed/offlined data.

## 10. Testing Plan

### 10.1 Frontend Unit Tests

Must cover:

1. plugin menu locale keys
2. language field selection / fallback
3. public routes rendering without login
4. project detail back affordance

### 10.2 Frontend E2E

Must cover:

1. anonymous `/plugins`
2. anonymous `/plugins/:id`
3. CMS language switch affecting plugin text display
4. CMS project detail still usable with long release lists

### 10.3 Backend Tests

Must cover at least:

1. public plugin query behavior
2. plugin admin workflow tests remain green
3. any new helper for public/detail normalization or error mapping

### 10.4 Real Environment Validation

Run again in Docker with PostgreSQL:

1. anonymous market access
2. authenticated CMS plugin pages
3. language display check
4. seeded text no longer showing `???`

## 11. Implementation Order

1. Add the missing menu locales and language-display helpers.
2. Fix global route/auth behavior for public market routes.
3. Rebuild public market list/detail pages using old design direction.
4. Refactor CMS project management and project detail layout.
5. Fix seed/plugin text corruption and any required demo data.
6. Add/adjust tests.
7. Run frontend tests, backend tests, and Docker real-environment verification.

## 12. Acceptance Criteria

This refactor is complete only if all of the following are true:

1. `/plugins` and `/plugins/:id` are accessible without login.
2. Public market UI follows the old plugin market interaction model closely enough to feel like a dedicated ecosystem center.
3. CMS plugin menu shows proper Chinese and English labels instead of raw locale keys.
4. Plugin list/detail content in CMS follows current language selection.
5. Plugin detail page has a visible and reliable back path.
6. Long release lists no longer make the detail/timeline area unusable.
7. Plugin-related seeded/demo text no longer appears as `???`.
8. Relevant frontend/backend tests pass.
9. Docker-based real-environment verification confirms the public market and CMS plugin pages both work.
