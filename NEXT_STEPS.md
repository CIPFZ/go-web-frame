# Next Steps

## Goal

This file is a concrete execution checklist for the next agent continuing frontend work in:

- `C:\Users\ytq\work\ai\web-cms`

Main branch at handoff:

- `codex/plugin-release-platform`

This file is intentionally shorter and more action-oriented than `HANDOFF.md`.

---

## Current Priority

The current highest-priority task is:

- continue refining the **public plugin release center frontend**

Main routes:

- `/plugins`
- `/plugins/:id`

Main files:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\detail.tsx`

The user is still not fully satisfied with the visual style.

The target direction is:

- more enterprise
- more Ant Design Pro
- less rounded
- less decorative
- more aligned with the rest of `web-cms`

---

## Non-Negotiable Rules

### 1. Frontend collaboration rule

For frontend work, do this in order:

1. ask Claude Code for implementation draft first
2. use Gemini as design supplement or fallback
3. review and adapt generated code locally
4. integrate into the real repo
5. rebuild frontend and verify in browser

Do not use Claude/Gemini only as “advice”.

The user explicitly wants their generated code to be the starting point.

### 2. Do not invent a separate permission system

Plugin menu visibility and operation permissions are controlled by CMS role management.

Do not implement ad hoc visibility rules that bypass:

- `系统管理 -> 角色管理`

### 3. Public center and CMS center are different

Do not mix these together:

- `/plugins` is public browse/download
- `/plugin/center` is admin-side management

### 4. Avoid building new core interaction in legacy page

Treat this file as legacy unless absolutely necessary:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\center\index.tsx`

Use these active pages instead:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\detail.tsx`

---

## Suggested Immediate Work Order

### Phase 1. Re-read active public pages

Open these first:

1. `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\index.tsx`
2. `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\detail.tsx`
3. `C:\Users\ytq\work\ai\web-cms\front-end\config\routes.ts`

Then compare with:

4. `C:\Users\ytq\work\ai\web-cms\front-end\src\app.tsx`
5. `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\dashboard\workplace\index.tsx`

Reason:

- the plugin public pages should visually feel like they belong to this CMS

### Phase 2. Ask Claude Code for the next UI revision

Prompt focus:

- enterprise public plugin directory
- Ant Design Pro-aligned
- low radius, restrained hierarchy
- no consumer-marketplace look
- keep current real API structure

If Claude output is unstable or unavailable:

- ask Gemini for backup implementation

### Phase 3. Review generated code before integrating

Check especially for:

- fake fields not in our API
- wrong route assumptions
- wrong button actions
- incorrect imports
- components too decorative for `web-cms`
- mismatched typography or spacing scale

### Phase 4. Integrate and validate

Run:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false
```

Note:

- repo has historical TS errors
- focus on whether changed files add new errors

Then rebuild:

```powershell
cd C:\Users\ytq\work\ai\web-cms
docker compose build frontend
docker compose up -d frontend
```

Then verify:

- [http://localhost/#/plugins](http://localhost/#/plugins)
- [http://localhost/#/plugins/1](http://localhost/#/plugins/1)

---

## Most Likely Next UI Tasks

These are the most probable changes the user will ask for next.

### A. Public list page refinement

Potential adjustments:

- make list rows denser
- reduce visual weight of top summary block
- tighten padding and spacing
- make the left filter panel feel more like CMS filters
- reduce remaining marketplace feel

File:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\index.tsx`

### B. Public detail page refinement

Potential adjustments:

- improve header information hierarchy
- make selected version more prominent
- simplify right column cards
- refine tabs content spacing and grouping
- make timeline and history feel more official / enterprise

File:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\detail.tsx`

### C. CMS-side project pages

Only continue here if user asks to return to admin-side management UX.

Files:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project-center\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\project\index.tsx`

---

## Visual Guardrails

When making UI choices, bias toward:

- `#f5f7fa` style neutral background
- 2px to 4px radius
- standard Ant Design borders
- subtle shadows only
- flatter cards
- denser content
- clearer title / meta / action hierarchy

Avoid:

- large gradients
- oversized rounded cards
- floating consumer marketplace look
- exaggerated hero sections
- decorative empty space

---

## API Reality Check

Public pages already use real backend APIs.

Relevant file:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\services\api\plugin.ts`

Important client methods:

- `getPublishedPluginList`
- `getPublishedPluginDetail`

Before merging Claude/Gemini generated code, verify field usage against this file.

---

## Quick Resume Prompt

If a new agent needs an even shorter kick-off prompt, use this:

> Continue frontend work in `C:\Users\ytq\work\ai\web-cms` on branch `codex/plugin-release-platform`. Focus on public plugin pages `/plugins` and `/plugins/:id`. User wants a stronger enterprise / Ant Design Pro style and thinks current pages are still too rounded and not aligned enough with `web-cms`. Use Claude Code first for the next implementation draft, then review and integrate locally. Main files are `front-end/src/pages/plugin/market/index.tsx` and `front-end/src/pages/plugin/market/detail.tsx`. Rebuild with `docker compose build frontend` and `docker compose up -d frontend`.

---

## Final Reminder

The best default assumption is:

- the user will keep iterating on the public plugin UI until it feels like a serious enterprise plugin directory inside a CMS

So optimize for:

- structure
- clarity
- consistency
- restrained styling

not for:

- novelty
- visual flair
- consumer marketplace aesthetics

