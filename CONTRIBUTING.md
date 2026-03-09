# Contributing Guide

## Branch Naming

Use short, purpose-driven names:

- `feature/<topic>`
- `fix/<topic>`
- `chore/<topic>`
- `release/<version>`

Examples:

- `feature/notice-center`
- `fix/server-state-404`

## Commit Message Convention

This repository uses Conventional Commits:

`<type>(<scope>): <subject>`

Examples:

- `feat(workplace): add notice card`
- `fix(state): handle missing metrics`
- `docs(deploy): add k3s instructions`

Allowed `type`:

- `feat`
- `fix`
- `docs`
- `style`
- `refactor`
- `perf`
- `test`
- `build`
- `ci`
- `chore`
- `revert`

Rules:

- Subject should be imperative and concise.
- Keep first line within 100 chars.
- Use scope when possible (`backend`, `frontend`, `deploy`, `ci`, etc).

## Enable Local Git Hooks

Run once in repo root:

```bash
git config core.hooksPath .githooks
git config commit.template .gitmessage.txt
```

What hooks enforce:

- `commit-msg`: validates Conventional Commit format.
- `pre-commit`: blocks accidental staging of runtime/generated directories.

## Pull Request Checklist

- Keep PR scope small and focused.
- Include test evidence for behavior changes.
- Update docs/config examples when behavior changes.
- Use a Conventional Commit style PR title.

