# Vacuum horizon demonstrated: idle transaction blocks cleanup, backend_xmin == removable cutoff

The user's two-terminal experiment (blog/experiments/0004-idle-in-transaction.txt) demonstrated the full vacuum-horizon story: with a REPEATABLE READ session idle-in-transaction, `VACUUM (VERBOSE)` reported "3 are dead but not yet removable" with removable cutoff 757, while their own `pglab transactions` showed the blocker's `BACKEND_XMIN = 757` — the same number from two independent instruments. After the session committed, the next vacuum removed the tuples (cutoff 758, DEAD back to 0 in `pglab tables`). All three of their predictions were correct.

## Evidence
- Transcript: vacuum #1 (3 unremovable, cutoff 757), pglab transactions (idle in transaction, 2m, xmin 757), vacuum #2 (3 removed, cutoff 758), pglab tables (DEAD 0).
- Prediction precision gap noted: user called backend_xmin "the xid of the open transaction" — refined: it is the snapshot floor (oldest xid the session may still need).

## Implications
- MVCC row versions, snapshots, AND the vacuum horizon are now all user-verified end-to-end — the conceptual core of Phase 1 is in place.
- Glossary promotions: `backend_xmin`, `idle in transaction`.
- Episode 2's remaining work is packaging: blog post 0002 from 0003+0004 transcripts, then Episode 3 (autovacuum threshold math + `pglab vacuum --watch`), which directly answers their earlier 10GB/2× disk question (recycling vs. horizon-blocked growth).
- The user skipped A's post-update re-select (the "old values visible" moment) — acceptable: 0003 already evidenced it; keep in mind for the capstone from-memory test.
