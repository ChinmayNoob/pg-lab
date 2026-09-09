# pg-lab #1: Dead tuples — the ghost rows

> Part of **pg-lab** — *Learning PostgreSQL internals by building a lab, one experiment at a time.*
> Episode 1 of Phase 1: MVCC & Vacuum.

Last episode I built the lab: a disposable Postgres 17 in Docker and a Go CLI (`pglab`) that could only say "ping". This episode it got its first real instrument — `pglab tables`, which reads `pg_stat_user_tables` — and I used it to catch Postgres in the act of doing the thing that surprised me most when I first read about it:

**Postgres never edits a row in place.** Every `UPDATE` writes a brand-new copy of the row and leaves the old one behind. The old copy is a *dead tuple*: invisible to every reader, but still sitting on your disk, eating space, waiting for `VACUUM`.

That's the theory. Here's how I proved it to myself — and where my predictions fell flat on their face.

## Experiment 1 — seeing row versions with my own eyes

Every row version in Postgres carries two transaction IDs in its header: `xmin` (the transaction that created this version) and `xmax` (the transaction that replaced/deleted it, `0` while nobody has). You can just select them:

```sql
SELECT xmin, xmax, i FROM t;
```

I created a 3-row table and looked. Three rows, one insert statement — and all three rows share **the same `xmin` (751)**, because one statement is one transaction:

```text
 xmin | xmax | i
------+------+---
  751 |    0 | 1
  751 |    0 | 2
  751 |    0 | 3
(3 rows)
```

Then the crime:

```text
PS> docker exec -it pglab psql -U pglab -d pglab -c "update t set i = i * 10;"
UPDATE 3
```

And the line-up after:

```text
 xmin | xmax | i
------+------+---
  752 |    0 | 10
  752 |    0 | 20
  752 |    0 | 30
(3 rows)
```

Every row got a **new `xmin` (752)** — the updating transaction. New row versions, written by one transaction, same ID on all three. And `xmax` is still 0 on *these* rows — because these are the *new* versions, and nobody has replaced *them* yet. The `xmax = 752` stamp lives on the **old** versions, which my query can no longer see.

One table, two generations of rows, and each `SELECT` only ever sees its own generation. That's the visibility rule doing its job.

## Experiment 2 — scaling it up: 1,000,000 ghost rows

Three rows prove the mechanism. But the PlanetScale article that started this whole series is about what this behavior does to *large* tables. So: does the same trick produce measurable bloat at 1M rows?

Before running anything, I wrote my predictions (the rules of this lab say I have to commit before I see data):

> 1. After inserting 1M rows: `n_live_tup` = 1,000,003, table size = 2600 KB
> 2. After updating ALL 1M rows (no vacuum): live stays same, `n_dead_tup` = 1,000,003, size = 5200 KB
> 3. Bonus: after VACUUM, table size will... 2600 KB (yes, I dodged the shrink-or-stay question — more on that below)

Fresh database, 1M rows via `generate_series` + a fake md5 payload, autovacuum **disabled for this table** so the background janitor can't clean up my crime scene mid-experiment:

```text
CREATE TABLE events AS SELECT generate_series(1,1000000) AS id, md5(random()::text) AS payload;
ALTER TABLE events SET (autovacuum_enabled = false);

TABLE   SIZE     LIVE       DEAD  DEAD %  LAST VACUUM  LAST AUTOVACUUM
events  65.5 MB  1,000,000  0     0.0%    never        never
```

Now update *every* row — note the payload doesn't even change (`SET payload = payload`); in MVCC you don't need to change anything to rewrite everything:

```text
UPDATE events SET payload = payload;
=> UPDATE 1000000

TABLE   SIZE      LIVE       DEAD       DEAD %  LAST VACUUM  LAST AUTOVACUUM
events  130.7 MB  1,000,000  1,000,000  50.0%   never        never
```

**One million dead tuples. 50% of the table is ghosts.** Every live row has a corpse of almost exactly the same size lying next to it. Nothing was deleted — the table just has two generations now.

Then vacuum, and the moment I'd been waiting for:

```text
VACUUM events;

TABLE   SIZE      LIVE       DEAD  DEAD %  LAST VACUUM  LAST AUTOVACUUM
events  130.7 MB  1,000,000  0     0.0%    0s ago       never
```

## Scoring my predictions

| Prediction | Reality | Verdict |
|---|---|---|
| 1M rows → LIVE = 1,000,003 | 1,000,000 | ~right (the +3 was me accidentally counting a leftover table from an earlier experiment 😅) |
| size 2600 KB | **65.5 MB** | **wrong — by ~25×** |
| live stays same after update | 1,000,000 | right |
| DEAD = ~1M after update | 1,000,000 | right |
| size doubles | 65.5 → 130.7 MB, almost exactly 2× | right — and honestly, the doubling model felt *great* |
| after VACUUM size will be "2600 KB" | **130.7 MB, completely unchanged** | dodged the question and got humbled by the answer |

## The part that actually surprised me

My intuition for sizes came from my 3-row experiment: "3 rows took 8 KB, so 1M rows will take ~2.6 GB-ish... wait, let me just scale it." Terrible math, two separate problems:

1. **Small tables lie about per-row cost.** Those 3 rows fit in one 8 KB page — a *minimum allocation*, not a per-row price. At 1M rows the amortized truth shows up: ~68 bytes per row (a ~24-byte row header + 32-byte payload + id), ≈ 65 MB. The fixed overhead that dominated at n=3 becomes invisible at n=1M.
2. **The VACUUM reveal:** `DEAD` went from 1,000,000 → 0, but `SIZE` didn't move a byte. **VACUUM recycles dead tuples' space for *future* rows of the same table — it does not hand disk back to the operating system.** A table that was ever 50% bloated *stays* 130 MB on disk until something explicitly rewrites it (`VACUUM FULL`, which takes a heavy lock, or tools like pg_repack).

And that's the exact mechanism behind the "our tables grew and never shrank" production stories — the size on disk records your *worst* churn history, not your current data volume.

## What I got wrong

- Estimated 1M-row table size off by ~25× by extrapolating from a 3-row table.
- Thought VACUUM would shrink the file. It recycles; it doesn't return.
- My first xmin transcript was missing the post-UPDATE query entirely — I had captured the crime but not the evidence. (Lesson: run the experiment *completely* before writing anything down.)

## Try it yourself

```bash
git clone https://github.com/ChinmayNoob/pg-lab
cd pg-lab
docker compose up -d
docker exec -it pglab psql -U pglab -d pglab -c "create table events as select generate_series(1,1000000) as id, md5(random()::text) as payload;"
docker exec -it pglab psql -U pglab -d pglab -c "alter table events set (autovacuum_enabled = false);"
go run ./cmd/pglab tables                        # baseline
docker exec -it pglab psql -U pglab -d pglab -c "update events set payload = payload;"
go run ./cmd/pglab tables                        # 1M dead tuples, size 2x
docker exec -it pglab psql -U pglab -d pglab -c "vacuum events;"
go run ./cmd/pglab tables                        # dead: 0, size: unchanged
```

Raw transcripts: [`0001-xmin.txt`](experiments/0001-xmin.txt) · [`0002-1m-update.txt`](experiments/0002-1m-update.txt)

## Next

**pg-lab #2: Transactions — the reason vacuum can't help you.** If vacuum removes dead tuples, why do real tables drown in them? Because a transaction that opened *before* your update — even one sitting perfectly idle — still holds the door open for readers who need the old versions. I'll build `pglab transactions` to catch the culprit in the act, then script the whole murder with `pglab experiment vacuum`.
