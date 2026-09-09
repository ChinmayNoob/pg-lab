# Mission: PostgreSQL internals (via the pg-lab project)

## Why
Understand what actually happens inside a production PostgreSQL database — MVCC, vacuum, dead tuples, WAL — well enough to reason about large-table problems and outages from first principles, by building a Go CLI lab (`pglab`) and publishing every experiment to the public **pg-lab** blog series.

## Success looks like
- Whiteboard MVCC **from memory** (Phase 1 exit test): row versions, `xmin`/`xmax`, snapshots, dead tuples, why an idle-in-transaction session stalls vacuum — demonstrating every claim live with `pglab`
- Six published posts (`pg-lab #0`–`#5`) grounded in real experiment transcripts, not docs-quotes
- A repo a stranger can clone and run in under five minutes

## Constraints
- 3–6 hrs/week; an episode is capped at 2 weeks
- CLI only in Phase 1 — no UI, no web server, no TUI
- Blog is published on the existing personal site; zero site-building work
- One branch per phase; `main` merges only when a phase's exit test passes
- Lessons come **before** code: predict first, build second

## Out of scope (for now)
- Locks/bloat deep-dives (Phase 2), WAL/replication (Phase 3), benchmark harness (Phase 4), partitioning/sharding (Phase 5)
- TUI/web dashboards, ORMs, production-hardening of `pglab`
