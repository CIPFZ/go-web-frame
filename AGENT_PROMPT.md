# Agent Prompt

Copy the content below into a new Antigravity agent as the first message.

---

Continue work in:

- `C:\Users\ytq\work\ai\web-cms`

Current branch:

- `codex/plugin-release-platform`

Read these files first:

1. `C:\Users\ytq\work\ai\web-cms\HANDOFF.md`
2. `C:\Users\ytq\work\ai\web-cms\NEXT_STEPS.md`

Then continue the frontend task with these rules:

## Current priority

The current top priority is the **public plugin release center frontend**:

- `/plugins`
- `/plugins/:id`

Main files:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\index.tsx`
- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\market\detail.tsx`

## User requirements

The user thinks the current public plugin pages are still:

- not enterprise enough
- too rounded
- not aligned enough with the existing `web-cms` style

The desired direction is:

- more enterprise
- more Ant Design Pro / CMS-like
- less decorative
- less consumer marketplace feeling
- more restrained and structured

## Collaboration rule

For frontend changes, do this in order:

1. ask Claude Code for the implementation draft first
2. use Gemini only as supplement or fallback
3. review the generated code carefully
4. adapt it to the real repo and API
5. integrate, rebuild, and verify

Important:

- do not use Claude/Gemini only for suggestions
- the user wants their generated code to be the starting point

## Product boundary

Keep these areas separate:

- `/plugins` = public plugin browse/download center
- `/plugin/center` = CMS internal plugin management

Do not move new core UX into the legacy page:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\pages\plugin\center\index.tsx`

Use the active pages and routes already in the repo.

## Visual guardrails

Bias toward:

- neutral background like `#f5f7fa`
- small radius `2px` to `4px`
- standard Ant Design borders
- subtle shadows
- denser layout
- clearer information hierarchy

Avoid:

- large gradients
- oversized rounded cards
- decorative marketplace visuals
- playful or consumer-style UI

## Reality checks

Before trusting generated frontend code, verify:

- field names match our actual plugin API
- route assumptions are correct
- no fake rating/store concepts are introduced
- downloads still use our real data model

Relevant API file:

- `C:\Users\ytq\work\ai\web-cms\front-end\src\services\api\plugin.ts`

Relevant route file:

- `C:\Users\ytq\work\ai\web-cms\front-end\config\routes.ts`

## Validation workflow

After changes:

1. run targeted TS check in `front-end`
2. rebuild frontend container
3. verify these pages manually:
   - [http://localhost/#/plugins](http://localhost/#/plugins)
   - [http://localhost/#/plugins/1](http://localhost/#/plugins/1)

Commands:

```powershell
cd C:\Users\ytq\work\ai\web-cms\front-end
npm run tsc -- --pretty false
```

```powershell
cd C:\Users\ytq\work\ai\web-cms
docker compose build frontend
docker compose up -d frontend
```

Start by summarizing the current public plugin list page and detail page structure, then propose the next UI refinement step before editing.

