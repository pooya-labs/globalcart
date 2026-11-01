# Release & Handoff Process

## Versioning
- Patch: fixes and small refinements.
- Minor: a roadmap milestone or coherent feature slice.
- Major: stable v1.0 release or breaking architecture transition.

## Required updates
For every release update package.json, CHANGELOG.md, ROADMAP.md, and PROJECT_CONTEXT.md together.

## Verification
Run npm ci, npm run typecheck, npm run build, and relevant tests. Verify the GitHub Pages base path. Do not claim deployment success until Actions passes and the published URL is checked.

## AI handoff
A new assistant must read PROJECT_CONTEXT.md, ROADMAP.md, CHANGELOG.md, docs/ARCHITECTURE.md, and docs/DECISIONS.md, then inspect current files and workflow status. The roadmap is the source of truth, not chat history.
