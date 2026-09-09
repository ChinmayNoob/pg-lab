# Snapshots confirmed: REPEATABLE READ pins the past until commit

The user ran the two-terminal experiment (blog/experiments/0003-snapshot-horizon.txt). After a false start (BEGIN swallowed by PowerShell multi-line paste; update ran before the snapshot existed), they corrected the ordering and demonstrated: an open REPEATABLE READ transaction keeps seeing the old generation (xmin 755) even after the updater commits, with xmax=756 now stamped on those old rows; after COMMIT, a new snapshot shows the new generation (xmin 756, xmax=0). Their Episode 1 xmax=0 confusion is thereby resolved with their own evidence. They then asked the production-grade question: "if my DB is 10GB do I need 20GB free?" — understanding that version churn doubles size, but not yet the recycle-vs-return distinction.

## Evidence
- Transcript shows all five stages: pre-update select, pg_stat_activity gap (not yet run), B's commit, A still seeing old rows with xmax=756, post-commit new generation.
- First failed run also produced real learning: autocommit means every bare statement is its own transaction; "WARNING: there is no transaction in progress" proves BEGIN never executed.

## Implications
- Snapshot/visibility and generation mechanics: established. Promote `snapshot` and `xmax` to the glossary.
- Remaining gap: the vacuum horizon (backend_xmin) — the pg_stat_activity check was skipped. Confirm via the pglab transactions experiment (Episode 2 build) before the capstone.
- Their disk-growth question is the hook for Episode 3's topic (autovacuum thresholds + why churn is normally recycled but blocked transactions break recycling).
- PowerShell multi-line paste into interactive psql eats lines — lab manual should note: paste one line at a time inside psql.
