# pg-lab #2: Transactions — the reason vacuum can't help you

> Part of **pg-lab** — *Learning PostgreSQL internals by building a lab, one experiment at a time.*
> Episode 2 of Phase 1: MVCC & Vacuum.

Episode 1 ended on a cliffhanger: vacuum removes dead tuples, but in production huge tables drown in them. Why doesn't the janitor just do its job?

The answer is a single word: **snapshots**. And my own tool ended up catching the culprit red-handed — matching vacuum's own confession by the exact same number.

## The rule that runs everything

When any statement runs, Postgres takes a **snapshot**: essentially the list of transactions that were *still in progress* at that instant. Visibility of every row version is then decided by comparing its header IDs against that list:

> A row version is visible if its `xmin` transaction had **committed** before the snapshot was taken, and its `xmax` transaction had **not**.

*Committed* — not "started". An open transaction that hasn't written anything yet still counts, because Postgres can't know its future. And since every query needs a snapshot, the database must keep every row version that *any* open snapshot might still want to see.

Follow that to its conclusion and you get the production story:

> A transaction that is open but idle — doing nothing, sending nothing — still pins the horizon. Vacuum cannot reclaim anything its snapshot might need. The janitor is not on strike; it's being legally prevented from entering.

## Experiment 1 — a session that refuses to move on

Two terminals. Terminal A opens a transaction at `REPEATABLE READ` isolation (snapshot frozen at the first statement) and reads the table. Terminal B updates every row underneath it. What does A see now?

I had to earn this one. My first attempt failed twice, informatively:

- I pasted a multi-line block (`BEGIN ...; SELECT ...`) through PowerShell into psql — PowerShell ate the lines, psql never saw the `BEGIN`. psql later told me directly: `WARNING: there is no transaction in progress` when I tried to COMMIT. (Corollary worth keeping: in psql's default autocommit mode, every bare statement is its own transaction — which is exactly why Episode 1's `docker exec -c` runs each got a fresh xmin.)
- I also ran Terminal B's update *before* opening A — so when A finally looked, it saw `xmin 755, values 101,102,103`: the post-update world. A can only hold the past if its snapshot exists *before* the crime.

Third attempt, correct order:

```text
-- Terminal A
pglab=# BEGIN ISOLATION LEVEL REPEATABLE READ;
BEGIN
pglab=*# SELECT xmin, xmax, i FROM t;      -- prompt shows * = transaction open
 xmin | xmax |  i
------+------+-----
  755 |    0 | 101
  755 |    0 | 102
  755 |    0 | 103
(3 rows)
```

```text
-- Terminal B
PS> docker exec -it pglab psql -U pglab -d pglab -c "update t set i = i + 100;"
UPDATE 3
```

Terminal A re-runs the same SELECT — *without committing*:

```text
pglab=*# SELECT xmin, xmax, i FROM t;
 xmin | xmax |  i
------+------+-----
  755 |  756 | 101     ← old values, and xmax is now stamped!
  755 |  756 | 102
  755 |  756 | 103
(3 rows)
```

B committed. A still sees the **old generation** — because A's snapshot needs it. But look at the header: `xmax = 756`. Those old rows are now officially dead-for-everyone-except-A, and the killing transaction's ID is written right on them. This is also the answer to the mystery I filed in Episode 1: I kept seeing `xmax = 0` after updates because I was looking at the *new* generation. The xmax stamp lands on the *old* one — which each snapshot can only see if it still needs it.

Then A commits, looks again, and the world moves on:

```text
pglab=# COMMIT;
COMMIT
pglab=# SELECT xmin, xmax, i FROM t;
 xmin | xmax |  i
------+------+-----
  756 |    0 | 201
  756 |    0 | 202
  756 |    0 | 203
(3 rows)
```

## Experiment 2 — vacuum, blocked: the same number from two instruments

Episode 1 left the lab with a mess I created on purpose. Time for the real question: **what exactly does vacuum do when a live session is holding the archive door open?**

My predictions (written first, per lab rules):

> 1. While A sits idle-in-transaction, after an UPDATE, VACUUM will **not** remove the dead tuples — an older snapshot may still be able to see those versions.
> 2. After A exits, vacuum removes them — the snapshot is gone.
> 3. `pglab transactions` will show A's `BACKEND_XMIN` as the xid of A's still-open transaction.

I opened A again (`REPEATABLE READ`, snapshot frozen at xid horizon 757), updated all rows from B, then asked vacuum to clean up — verbosely, so it would have to explain itself:

```text
PS> docker exec -it pglab psql -U pglab -d pglab -c "vacuum (verbose) t;"
INFO:  vacuuming "pglab.public.t"
INFO:  finished vacuuming "pglab.public.t": index scans: 0
pages: 0 removed, 1 remain, 1 scanned (100.00% of total)
tuples: 6 removed, 6 remain, 3 are dead but not yet removable
removable cutoff: 757, which was 1 XIDs old when operation ended
```

Read that third line again: **"3 are dead but not yet removable."** Vacuum did its job, walked the table, cleaned up what it legally could — and reported, in writing, that it was forbidden from removing the 3 tuples my idle session was pinning.

Now the part that made my week. While the hostage situation was in progress, I pointed my own tool at the database:

```text
PS> go run ./cmd/pglab transactions
PID    USER   STATE                XACT AGE  BACKEND_XMIN
12244  pglab  idle in transaction  2m        757
```

**757 = 757.** My two-minute-old, perfectly idle session — a session that has done nothing but *exist* — was carrying `BACKEND_XMIN 757`, and vacuum's own `removable cutoff` was 757. The exact same number, reported by two independent instruments: mine naming the hostage-taker, Postgres's vacuum confirming the negotiation failed.

Release the hostage, clean again:

```text
-- Terminal A: COMMIT;

PS> docker exec -it pglab psql -U pglab -d pglab -c "vacuum (verbose) t;"
INFO:  finished vacuuming "pglab.public.t": index scans: 0
tuples: 3 removed, 3 remain, 0 are dead but not yet removable
removable cutoff: 758, which was 0 XIDs old when operation ended

PS> go run ./cmd/pglab tables
TABLE   SIZE      LIVE       DEAD  DEAD %  LAST VACUUM  LAST AUTOVACUUM
t       40.0 KB   3          0     0.0%    8s ago       never
```

Cutoff moved 757 → 758 the instant A committed. Dead tuples: gone.

## Scoring my predictions

| Prediction | Reality | Verdict |
|---|---|---|
| Vacuum won't remove the tuples while A is open | "3 are dead but not yet removable" | right |
| After A exits, they go | "3 removed", `DEAD 0` in `pglab tables` | right |
| `BACKEND_XMIN` = A's transaction's xid | 757, and it equals vacuum's cutoff | right number, slightly wrong *idea* — see below |

That last one deserves precision, because it's the difference between passing and understanding: `BACKEND_XMIN` is not "the session's ID". It's the **oldest xid the session might still need** — its snapshot floor. Vacuum must treat everything older than that floor as potentially-wanted. That's how a session that has been idle for two minutes keeps strangling cleanup: it did its damage by *existing*, not by *doing*.

## The question this left me with

If every UPDATE doubles a table's size (Episode 1), doesn't a busy 10 GB table need 20 GB of disk forever?

No — and the distinction matters: vacuum's reclaimed space goes into the table's **free-space map** and the *next* round of updates reuses those pages. Steady churn (update 10%, vacuum, repeat) circulates the same disk forever. The 2× spike is *transient* — disk usage records your worst churn, not your total history.

**Unless** something pins the horizon — like the idle session above. Horizon-blocked dead tuples can't be recycled, so growth becomes permanent and compounding. That's precisely the large-table outage story that started this series, and it's why monitoring tools scream about `idle in transaction` sessions.

## What I got wrong

- Ran the update before opening the observing session — the demo needs the snapshot to exist *first*. Order matters.
- Pasted multi-line SQL through PowerShell into psql and lost the `BEGIN`. Type one line at a time inside psql; confirm the `*` in the prompt.
- Said `BACKEND_XMIN` is "the xid of the open transaction". It's the snapshot floor — the oldest xid the session might still need.
- Skipped re-querying A after B's update in the final experiment (the "old values visible" money shot) — I'd already captured that evidence in experiment 1, but a complete transcript beats a assembled story.

## Try it yourself

```bash
git clone https://github.com/ChinmayNoob/pg-lab && cd pg-lab
docker compose up -d

# make the practice table (fresh lab has none):
docker exec -it pglab psql -U pglab -d pglab -c "create table t as select generate_series(1,3) as i;"

# Terminal A — interactive, keep open:
docker exec -it pglab psql -U pglab -d pglab
  BEGIN ISOLATION LEVEL REPEATABLE READ;
  SELECT xmin, xmax, i FROM t;

# Terminal B:
docker exec -it pglab psql -U pglab -d pglab -c "update t set i = i + 1;"
docker exec -it pglab psql -U pglab -d pglab -c "vacuum (verbose) t;"   # watch 'not yet removable'
go run ./cmd/pglab transactions                                          # find the hostage-taker
go run ./cmd/pglab tables                                                # dead tuples stuck

# back in Terminal A: COMMIT;
docker exec -it pglab psql -U pglab -d pglab -c "vacuum (verbose) t;"   # now it's removable
go run ./cmd/pglab tables                                                # DEAD 0
```

Raw transcripts: [`0003-snapshot-horizon.txt`](experiments/0003-snapshot-horizon.txt) · [`0004-idle-vacuum.txt`](experiments/0004-idle-vacuum.txt)

## Next

**pg-lab #3: Autovacuum's trigger math — why big tables clean up late.** Vacuum *can* save you, but autovacuum only wakes up when `dead_tuples > 50 + 0.2 × live_tuples`. On a 1M-row table that's ~200K tolerated ghosts. I'll derive the threshold, then build `pglab vacuum --watch` — a live oscilloscope for dead tuples — and watch autovacuum's actual trigger point against the math.
