# Plugin Center Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the CMS-side plugin release module into a role-driven structure with three sub-views (`项目管理`, `审核工作台`, `发布工作台`) and one unified project detail page centered on project/version/workflow.

**Architecture:** Keep `web-cms` global navigation and role/menu control unchanged, and refactor only the plugin module’s internal information architecture. Split the module into role-oriented list pages while unifying all deep work inside one shared project detail page that can default to different versions/tabs based on entry context.

**Tech Stack:** React, Umi Max, Ant Design, Ant Design Pro, TypeScript, existing plugin APIs in `front-end/src/services/api/plugin.ts`

---

## File Structure

### Existing files to modify

- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\config\routes.ts`
  - Ensure plugin module routes align with the new menu/page structure.
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\locales\zh-CN\menu.ts`
  - Add/update menu text for `项目管理`, `审核工作台`, `发布工作台`.
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\locales\en-US\menu.ts`
  - Keep English menu labels aligned.
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\services\api\plugin.ts`
  - Normalize plugin project, version, workflow list/detail request helpers used by all pages.
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`
  - Refactor into the final `项目管理` page focused on light project cards/list and one-line workflow summary.
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`
  - Rebuild into the unified project detail page with left project info, middle version list, right workflow + tabs.

### Existing files to repurpose or replace

- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\center\index.tsx`
  - Keep as compatibility export or redirect to the project management implementation.

### New files to create

- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\review-workbench\index.tsx`
  - New review workbench list page.
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\publish-workbench\index.tsx`
  - New publish workbench list page.
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\ProjectSummaryCard.tsx`
  - Reusable lightweight project card for project management.
- Create: `C:\Users\\ytq\\work\\ai\\web-cms\\front-end\\src\\pages\\plugin\\components\\VersionListPanel.tsx`
  - Reusable version list panel with filters and current-selection state.
- Create: `C:\Users\\ytq\\work\\ai\\web-cms\\front-end\\src\\pages\\plugin\\components\\WorkflowPanel.tsx`
  - Reusable workflow step header and action area.
- Create: `C:\Users\\ytq\\work\\ai\\web-cms\\front-end\\src\\pages\\plugin\\components\\VersionInfoTabs.tsx`
  - Reusable tabs for overview/files/review/timeline.
- Create: `C:\Users\\ytq\\work\\ai\\web-cms\\front-end\\src\\pages\\plugin\\components\\ProjectFilterBar.tsx`
  - Shared filter/search bar for project management.
- Create: `C:\Users\\ytq\\work\\ai\\web-cms\\front-end\\src\\pages\\plugin\\components\\WorkbenchFilterBar.tsx`
  - Shared filter/search bar for review/publish workbenches.

### Files to test

- Test: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`
- Test: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\review-workbench\index.tsx`
- Test: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\publish-workbench\index.tsx`
- Test: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`

## Task 1: Establish Plugin Module Route and Menu Skeleton

**Files:**
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\config\routes.ts`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\locales\zh-CN\menu.ts`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\locales\en-US\menu.ts`

- [ ] **Step 1: Add the target route structure in the routing config**

Update the plugin route block so it supports one compatibility entry plus three role-driven sub-pages and one unified project detail page.

```ts
{
  path: '/plugin',
  routes: [
    {
      path: '/plugin/center',
      component: './plugin/project-center',
    },
    {
      path: '/plugin/project-management',
      component: './plugin/project-center',
    },
    {
      path: '/plugin/review-workbench',
      component: './plugin/review-workbench',
    },
    {
      path: '/plugin/publish-workbench',
      component: './plugin/publish-workbench',
    },
    {
      path: '/plugin/project/:id',
      component: './plugin/project',
      hideInMenu: true,
    },
  ],
}
```

- [ ] **Step 2: Add Chinese menu labels**

Insert the following keys into `zh-CN/menu.ts`:

```ts
'menu.plugin': '插件发布',
'menu.plugin.center': '插件中心',
'menu.plugin.project-management': '项目管理',
'menu.plugin.review-workbench': '审核工作台',
'menu.plugin.publish-workbench': '发布工作台',
```

- [ ] **Step 3: Add English menu labels**

Insert the following keys into `en-US/menu.ts`:

```ts
'menu.plugin': 'Plugin Release',
'menu.plugin.center': 'Plugin Center',
'menu.plugin.project-management': 'Project Management',
'menu.plugin.review-workbench': 'Review Workbench',
'menu.plugin.publish-workbench': 'Publish Workbench',
```

- [ ] **Step 4: Run a targeted route/menu type check**

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false 2>&1 | Select-String 'config\\routes.ts|locales\\zh-CN\\menu.ts|locales\\en-US\\menu.ts'
```

Expected: no new output for these files

- [ ] **Step 5: Commit the route/menu skeleton**

```bash
git add front-end/config/routes.ts front-end/src/locales/zh-CN/menu.ts front-end/src/locales/en-US/menu.ts
git commit -m "feat: add plugin module route skeleton"
```

## Task 2: Refactor Project Management Into a Lightweight Project Entry Page

**Files:**
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\center\index.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\ProjectSummaryCard.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\ProjectFilterBar.tsx`

- [ ] **Step 1: Extract the lightweight project card component**

Create a reusable summary card that only shows project identity, latest version, a one-line workflow summary, and project status.

```tsx
type ProjectSummaryCardProps = {
  record: ProjectRecord;
  onOpen: () => void;
};

const ProjectSummaryCard: React.FC<ProjectSummaryCardProps> = ({ record, onOpen }) => (
  <Card hoverable onClick={onOpen} bodyStyle={{ padding: 14 }}>
    <Space direction="vertical" size={10} style={{ width: '100%' }}>
      <Space align="start" style={{ justifyContent: 'space-between', width: '100%' }}>
        <Avatar shape="square" size={44}>
          {(record.code || 'P').slice(0, 1).toUpperCase()}
        </Avatar>
        <Tag>{record.phaseLabel}</Tag>
      </Space>
      <div>
        <Typography.Title level={5} style={{ margin: 0 }} ellipsis>
          {record.nameZh || '-'}
        </Typography.Title>
        <Typography.Text type="secondary">{record.nameEn || '-'}</Typography.Text>
      </div>
      <Typography.Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>
        {record.descriptionZh || '-'}
      </Typography.Paragraph>
      <Space size={8} wrap>
        <Tag>{record.latestVersion || '-'}</Tag>
        <Tag color="processing">{record.workflowSummary || '暂无流程'}</Tag>
      </Space>
    </Space>
  </Card>
);

export default ProjectSummaryCard;
```

- [ ] **Step 2: Extract the project filter bar**

Create a shared filter bar that only contains project management search and filters.

```tsx
type ProjectFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  status: string;
  onStatusChange: (value: string) => void;
  hci: string;
  onHciChange: (value: string) => void;
  acli: string;
  onAcliChange: (value: string) => void;
  viewMode: 'card' | 'list';
  onViewModeChange: (value: 'card' | 'list') => void;
  canCreate: boolean;
  onCreate: () => void;
};
```

- [ ] **Step 3: Simplify `project-center` to project-centric content**

Refactor `project-center/index.tsx` so the page is responsible only for:

- loading projects/releases
- computing one latest workflow summary string per project
- rendering card/list view
- opening create/edit project modal
- navigating to project detail

Use this workflow summary helper:

```ts
const buildWorkflowSummary = (record: ProjectRecord) => {
  const version = record.activeRelease?.version || record.latestReleased?.version || record.latestVersion || '-';
  const status = workflowLabel(record.activeRelease?.status || record.latestReleased?.status);
  return `${version} ${status}`;
};
```

- [ ] **Step 4: Keep compatibility by re-exporting `center`**

Ensure `front-end/src/pages/plugin/center/index.tsx` remains:

```ts
export { default } from '../project-center';
```

- [ ] **Step 5: Run a targeted type check for project management**

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false 2>&1 | Select-String 'src\\pages\\plugin\\project-center\\index.tsx|src\\pages\\plugin\\center\\index.tsx|src\\pages\\plugin\\components\\ProjectSummaryCard.tsx|src\\pages\\plugin\\components\\ProjectFilterBar.tsx'
```

Expected: no new output for these files

- [ ] **Step 6: Commit project management refactor**

```bash
git add front-end/src/pages/plugin/project-center/index.tsx front-end/src/pages/plugin/center/index.tsx front-end/src/pages/plugin/components/ProjectSummaryCard.tsx front-end/src/pages/plugin/components/ProjectFilterBar.tsx
git commit -m "feat: refactor plugin project management page"
```

## Task 3: Build Review and Publish Workbench List Pages

**Files:**
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\review-workbench\index.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\publish-workbench\index.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\WorkbenchFilterBar.tsx`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\services\api\plugin.ts`

- [ ] **Step 1: Normalize list query helpers in plugin service**

Add or normalize helpers so review/publish pages can fetch filtered version queues cleanly.

```ts
export async function getReleaseList(params?: API.ReleaseQuery, options?: Record<string, any>) {
  return request<API.BasicResponse<API.PageResult<API.ReleaseItem>>>('/api/v1/plugin/release/getReleaseList', {
    method: 'POST',
    data: params,
    ...(options || {}),
  });
}

export async function getProjectDetail(params: { pluginId: number }, options?: Record<string, any>) {
  return request<API.BasicResponse<API.ProjectDetail>>('/api/v1/plugin/plugin/getProjectDetail', {
    method: 'POST',
    data: params,
    ...(options || {}),
  });
}
```

- [ ] **Step 2: Create the shared workbench filter bar**

Implement one shared filter bar for review/publish pages.

```tsx
type WorkbenchFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  type: string;
  onTypeChange: (value: string) => void;
  status: string;
  onStatusChange: (value: string) => void;
  mineOnly: boolean;
  onMineOnlyChange: (value: boolean) => void;
};
```

- [ ] **Step 3: Build the review workbench page**

Create a table-first page with summary cards and pending/reviewed filters. Route table row clicks into the unified project detail page using query params.

```tsx
history.push({
  pathname: `/plugin/project/${record.pluginId}`,
  query: {
    releaseId: String(record.ID),
    from: 'review',
    tab: 'review',
  },
});
```

- [ ] **Step 4: Build the publish workbench page**

Mirror the review page structure, but with publish-oriented columns and filters.

```tsx
history.push({
  pathname: `/plugin/project/${record.pluginId}`,
  query: {
    releaseId: String(record.ID),
    from: 'publish',
    tab: 'review',
  },
});
```

- [ ] **Step 5: Run a targeted type check for workbench pages**

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false 2>&1 | Select-String 'src\\pages\\plugin\\review-workbench\\index.tsx|src\\pages\\plugin\\publish-workbench\\index.tsx|src\\pages\\plugin\\components\\WorkbenchFilterBar.tsx|src\\services\\api\\plugin.ts'
```

Expected: no new output for these files

- [ ] **Step 6: Commit workbench pages**

```bash
git add front-end/src/pages/plugin/review-workbench/index.tsx front-end/src/pages/plugin/publish-workbench/index.tsx front-end/src/pages/plugin/components/WorkbenchFilterBar.tsx front-end/src/services/api/plugin.ts
git commit -m "feat: add plugin review and publish workbenches"
```

## Task 4: Rebuild the Unified Project Detail Page Around Project + Version + Workflow

**Files:**
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\VersionListPanel.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\WorkflowPanel.tsx`
- Create: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\components\VersionInfoTabs.tsx`

- [ ] **Step 1: Extract the version list panel**

Build a focused version list component that supports:

- filters (`全部`, `进行中`, `已发布`, `已下架/已归档`)
- current selection
- current-to-me indicator
- fixed-height scrolling

```tsx
type VersionListPanelProps = {
  releases: ReleaseItem[];
  activeReleaseId?: number;
  filter: string;
  onFilterChange: (value: string) => void;
  onSelect: (release: ReleaseItem) => void;
  currentUserId?: number;
};
```

- [ ] **Step 2: Extract the workflow panel**

Implement the fixed top-right workflow area with conditional three-step flows.

```tsx
const buildSteps = (requestType: RequestType) =>
  requestType === 'offline'
    ? ['下架申请', '下架审核', '执行下架']
    : ['提交资料', '审核', '发布'];
```

```tsx
type WorkflowPanelProps = {
  release?: ReleaseItem;
  canManageProject: boolean;
  canReview: boolean;
  canPublish: boolean;
  onSubmit: () => void;
  onApprove: () => void;
  onReject: () => void;
  onPublish: () => void;
  onOffline: () => void;
  onArchive: () => void;
};
```

- [ ] **Step 3: Extract the version info tabs**

Implement tabs for:

- `概览`
- `文件资料`
- `审核与发布`
- `时间轴`

```tsx
type VersionInfoTabsProps = {
  release?: ReleaseItem;
  activeTab: string;
  onTabChange: (key: string) => void;
  editable: boolean;
  onUploadFile: (field: 'testReportUrl' | 'packageX86Url' | 'packageArmUrl', file: File) => Promise<void>;
};
```

- [ ] **Step 4: Rebuild `project/index.tsx` around the shared shell**

Refactor `project/index.tsx` so it:

- reads `releaseId`, `from`, `tab` from `history.location.query`
- loads the project detail
- picks the initial selected version from route context
- defaults tabs by source:
  - `from=review` -> `review`
  - `from=publish` -> `review`
  - otherwise -> `overview`

Use this initial tab helper:

```ts
const resolveInitialTab = (from?: string, tab?: string) => {
  if (tab) return tab;
  if (from === 'review' || from === 'publish') return 'review';
  return 'overview';
};
```

- [ ] **Step 5: Run a targeted type check for project detail**

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false 2>&1 | Select-String 'src\\pages\\plugin\\project\\index.tsx|src\\pages\\plugin\\components\\VersionListPanel.tsx|src\\pages\\plugin\\components\\WorkflowPanel.tsx|src\\pages\\plugin\\components\\VersionInfoTabs.tsx'
```

Expected: no new output for these files

- [ ] **Step 6: Commit unified detail page refactor**

```bash
git add front-end/src/pages/plugin/project/index.tsx front-end/src/pages/plugin/components/VersionListPanel.tsx front-end/src/pages/plugin/components/WorkflowPanel.tsx front-end/src/pages/plugin/components/VersionInfoTabs.tsx
git commit -m "feat: rebuild unified plugin project detail page"
```

## Task 5: Wire Role Actions, Create-Version Flow, and Final Verification

**Files:**
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\services\api\plugin.ts`
- Modify: `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`

- [ ] **Step 1: Implement the “create version shell first” flow**

In the project detail page, add a light create-version action that only asks for:

- version
- requestType
- versionConstraint

Then reload the project and auto-select the new version in `筹备中`.

```ts
const payload = {
  pluginId: project.ID,
  requestType: values.requestType,
  version: values.version,
  versionConstraint: values.versionConstraint,
};
```

- [ ] **Step 2: Keep actions aligned to role boundaries**

Ensure the action buttons match the approved role model:

```ts
const canSubmit = canManageProject && ['draft', 'release_preparing', 'rejected'].includes(release.status);
const canApprove = canReview && release.status === 'pending_review';
const canReject = canReview && release.status === 'pending_review';
const canRelease = canPublish && release.status === 'approved';
const canRequestOffline = canManageProject && release.status === 'released' && !release.isOfflined;
```

- [ ] **Step 3: Verify route-to-detail behavior from all entry pages**

Check these flows manually:

- project management -> project detail -> default overview
- review workbench -> project detail -> target version + review tab
- publish workbench -> project detail -> target version + review tab

Use:

```powershell
cd C:\Users\ytq\work\ai\web-cms
docker compose build frontend
docker compose up -d frontend
```

Expected:

- `http://localhost/#/plugin/project-management`
- `http://localhost/#/plugin/review-workbench`
- `http://localhost/#/plugin/publish-workbench`
- `http://localhost/#/plugin/project/1`

all load successfully and preserve expected context.

- [ ] **Step 4: Run final plugin-page type check**

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false 2>&1 | Select-String 'src\\pages\\plugin'
```

Expected: no new plugin-page-specific errors introduced by the refactor

- [ ] **Step 5: Commit the final integration**

```bash
git add front-end/src/pages/plugin/project/index.tsx front-end/src/pages/plugin/project-center/index.tsx front-end/src/services/api/plugin.ts
git commit -m "feat: wire plugin role actions and version flow"
```

## Self-Review

### Spec coverage

This plan covers:

- role-driven menu structure
- separate list pages for project/review/publish
- unified project detail page
- lightweight project summaries
- version list + workflow + tabs detail shell
- create-version shell-first flow
- role-based operations

No spec section was intentionally omitted.

### Placeholder scan

No `TBD`, `TODO`, “implement later”, or unspecified “handle appropriately” placeholders remain.

### Type consistency

The plan consistently uses:

- `ProjectRecord` for project management lists
- `ReleaseItem` for version/workflow units
- `RequestType` and `ReleaseStatus` for flow branching
- route query keys `releaseId`, `from`, `tab`

