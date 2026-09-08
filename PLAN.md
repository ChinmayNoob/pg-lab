# pg-lab — Project Plan & Blog Series Roadmap

> **pg-lab**: Learning PostgreSQL internals by building a lab, one experiment at a time.
> A public blog series where every concept is proven with a small Go CLI tool against a real Postgres.

---

## 1. What this is

`pg-lab` is a learning laboratory, not a monitoring product. Every phase follows the same loop: learn one PostgreSQL internals concept, build the smallest CLI feature that makes the concept observable, run a controlled experiment, and publish what happened.

- **Language/binary**: Go, single binary `pglab` (subcommands grow over time)
- **Database**: PostgreSQL in Docker, disposable between experiments
- **Interface (Phase 1)**: CLI only — no UI, no web server, no TUI
- **Output**: public blog series published on your existing site (drafts live in `blog/` in this repo)
- **Pace**: 3–6 hrs/week, honest budget

### Series naming (locked — never rename)

- **Series name**: `pg-lab`
- **Series title/subtitle**: *pg-lab — Learning PostgreSQL internals by building a lab, one experiment at a time*
- **Episode titles**: `pg-lab #N: <concept title>` — e.g. *pg-lab #3: Watching a table rot in real time*
- **Numbering is continuous across all phases** (#0–#5 in Phase 1, then #6, #7… never restart per phase). The phase is metadata stated inside the post ("Part of pg-lab, Phase 1: MVCC & Vacuum"), not in the title — this keeps titles phase-agnostic and the series feeling like one long project.
- On your site, keep one landing page/tag for the series: title + tagline + index of episodes. New phases just add episodes to the same series.

**The distinction that governs everything** (from Chat 2): a monitoring tool asks *"what is happening?"* — this lab asks *"why is it happening, can I reproduce it, and what changes when I modify X?"*

---

## 2. The goal, stated precisely

Phase 1 is done when you can do this, without opening a browser:

> Whiteboard MVCC from memory: how `xmin`/`xmax` work, what a dead tuple is, how snapshots decide visibility, why an `idle in transaction` session blocks vacuum, and why `VACUUM` sometimes "does nothing" — and demonstrate every claim live with `pglab`.

Secondary outcomes of Phase 1:

1. A working `pglab` binary with `tables`, `transactions`, `vacuum --watch`, `experiment vacuum`
2. ~6 published posts (Episode 0 through the capstone)
3. A repo another developer could clone and run

---

## 3. Non-goals for Phase 1

- No web server, dashboard, React, or TUI (bubbletea can wait for a later phase)
- No `pgload` workload generator (Episode 4 generates load inline with raw SQL)
- No bloat estimation, WAL, locks, replication, partitioning, sharding (Phases 2–5)
- No blog theme customization — stock theme, shipped
- No renaming. The repo is `pg-lab`, the binary is `pglab`. Done.

---

## 4. The learning method: Predict → Observe → Explain

Every episode uses this loop. The **prediction step is mandatory** — it is what turns tool-building into learning:

```text
1. LEARN     /teach builds a lesson for the concept (from trusted resources in RESOURCES.md)
2. PREDICT   write down, before running anything: "when I do X, I expect Y because Z"
3. BUILD     the smallest pglab feature that makes Y visible
4. OBSERVE   run the experiment, capture real output
5. EXPLAIN   was the prediction right? If not — that gap is the actual lesson
6. PUBLISH   the post leads with the prediction vs. reality
7. RECORD    /teach writes a learning-record: what stuck, what didn't → shapes the next lesson
```

**How `/teach` fits in** (the private learning layer — `blog/` stays the public layer):

- **`MISSION.md`** = PLAN.md §2 distilled (the Phase 1 exit test). Create it in the first `/teach` session; it grounds every lesson.
- **READ becomes LEARN**: instead of raw doc-dumping, each episode's concept gets an interactive `/teach` lesson (`lessons/NNNN-*.html`) built from the trusted resources tracked in `RESOURCES.md` (Postgres docs, the PlanetScale article, etc.). Lessons cite sources; we never trust parametric knowledge alone.
- **Retrieval practice = the prediction step**: `/teach`'s philosophy (fluency vs. storage strength, desirable difficulty) is exactly our Predict → Observe loop. Lesson quizzes happen *before* experiments; the experiment is the real quiz.
- **`learning-records/`** capture what actually stuck after each episode — they let `/teach` compute your zone of proximal development and decide whether the next episode proceeds or needs a reinforcement lesson first.
- **`reference/`** accumulates cheat sheets (catalog queries, MVCC glossary) — these become source material for the capstone explainer and future posts.
- **Wisdom**: `/teach` pushes toward communities (r/PostgreSQL, PostgreSQL Discord, pgsql-general) — commenting on others' large-table problems is free Phase 5 training.

**From-memory test protocol** (how you know a concept stuck):

> Before writing the blog post for a capstone/closing episode: open an empty editor and write the full explanation from memory. Gaps and wrong statements = go back to the lab and re-experiment. Only then consult the docs to fill remaining holes. That draft becomes the spine of the post.

This is the Phase 1 exit test: the MVCC explainer written from memory. `/teach` learning-records are the evidence trail for it.

---

## 5. Ground rules

| Rule                                                                                                                                                                             | Why                                                                        |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Time split per week: ~50% build, ~30% experiment/observe, ~20% write                                                                                                             | Writing eats coding time if unbounded; experiments are the point           |
| An episode takes**max 2 weeks**. If it's not done, cut scope, not the experiment                                                                                           | 3–6 h/week means episodes must stay small                                 |
| Publish before perfect; every post gets a*"What I got wrong"* section                                                                                                          | Public series ≠ polished series; being wrong on record is the format      |
| Blog work: drafts in `blog/` (source of truth), published to your existing site. **Never build site features for the blog** | You already have a site; the lab's job is content, not web dev |
| Every experiment runs against a**fresh Docker database** (`docker compose down -v && up`)                                                                                | Reproducibility; no state pollution between posts                          |
| Commit the raw experiment output (text files) into`blog/experiments/`                                                                                                          | Screenshots lie; transcripts don't                                         |
| **One branch per phase**, merged into `main` only when the phase's exit test passes (see below)                                                                     | `main` always = working, published state; phases stay isolated and movable |

### Git workflow

```text
main                          ── stable: every merge = a completed phase ──
  └── phase-1-mvcc-vacuum     ── all Episode 1–5 work happens here
```

- Branch per phase: `phase-1-mvcc-vacuum`, `phase-2-concurrency-bloat`, `phase-3-wal-replication`, …
- Episodes are ordinary commits on the phase branch (code + experiment transcript + post draft together)
- Merge to `main` **only when the phase's exit test passes** (e.g. Phase 1: the from-memory MVCC explainer). Merge commit message = the phase's one-line lesson.
- Phase 0 already lives on `main` (it is the lab itself — infrastructure, not a learning phase).
- If a phase derails: abandon the branch, re-plan, start a fresh branch. `main` never rots.

---

## 6. The tool

Single Go binary. Phase 1 command set (ruthlessly cut from the 7+ in Chat 2):

```text
pglab ping                  # Phase 0 — prove connectivity
pglab tables                # Phase 1 — dead tuples per table
pglab transactions          # Phase 1 — who is holding back vacuum
pglab vacuum --watch        # Phase 1 — watch dead tuples live
pglab experiment vacuum     # Phase 1 — the long-transaction MVCC demo
```

Later phases add: `locks`, `bloat`, `wal`, `load`, `experiment *` (see §9).

### Repo layout (target by end of Phase 1)

```text
pg-lab/
├── PLAN.md                  # this file
├── docker-compose.yml       # postgres:17, disposable lab DB
├── cmd/pglab/main.go
├── internal/
│   ├── postgres/            # pgx pool, connection config
│   ├── collector/           # queries against pg_stat_* / pg_catalog
│   ├── render/              # table formatting for CLI output
│   └── experiment/          # scripted scenarios (start with vacuum)
├── blog/                    # markdown drafts of posts (source of truth for your site)
│   └── experiments/         # raw transcripts of experiment runs
├── MISSION.md               # /teach workspace — why you're learning this (grounding for all lessons)
├── RESOURCES.md             # /teach workspace — trusted sources (Postgres docs, PlanetScale article, blogs)
├── NOTES.md                 # /teach workspace — how you like to be taught
├── lessons/*.html           # /teach workspace — private interactive lessons (one per concept)
├── reference/*.html         # /teach workspace — cheat sheets, catalog query reference, glossary
├── learning-records/*.md    # /teach workspace — what stuck, what didn't; drives what's next
└── assets/                  # /teach workspace — shared lesson components (stylesheet, quizzes)
```

All `/teach` artifacts are **private learning layer**; `blog/` is the **public layer**. Both get committed.

Use `pgx` (pgxpool) — no ORM, raw SQL everywhere (you are here to learn the catalogs).

---

## 7. Roadmap overview

| Phase       | Theme                   | Commands                                                                | Core concepts                                                                  | Posts |
| ----------- | ----------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------ | ----- |
| **0** | The lab itself          | `ping`                                                                | Docker, pgx, connection basics                                                 | 1     |
| **1** | MVCC & Vacuum           | `tables`, `transactions`, `vacuum --watch`, `experiment vacuum` | MVCC, xmin/xmax, dead tuples, snapshots, VACUUM, autovacuum, long transactions | 5–6  |
| **2** | Concurrency & bloat     | `locks`, `bloat`                                                    | Lock modes, blocking chains, page bloat, indexes, autovacuum tuning            | 3–4  |
| **3** | WAL & replication       | `wal`                                                                 | WAL, checkpoints, LSNs, streaming replication, lag                             | 3–4  |
| **4** | Workload & experiments  | `load`, `experiment *`                                              | Controlled benchmarks, autovacuum tuning A/B, batch deletes                    | 3–4  |
| **5** | Big data & distribution | partitioning + mini-shard router                                        | Partition pruning, sharding, resharding pain                                   | 3+    |

Phases 2–5 are sketches (§9). Phase 1 is fully specified (§8) — do not pre-plan it harder, the later phases will be re-grilled with what you learned.

---

## 8. Phase 1 — episode by episode

Pace assumption: 3–6 h/week → each episode ≈ 1–2 weeks.

---

### Episode 0 — "Building the lab" (~1 week)

- **Concept**: none yet — this is infrastructure, kept deliberately boring.
- **Build**:
  - `docker-compose.yml`: single `postgres:17`, volume for data, port 5432
  - Go module, `pgxpool`, `pglab ping` → prints server version + `pg_stat_database` row for the connected DB
  - Draft the post in `blog/` and publish it on your site
- **Post**: "pg-lab #0: I'm going to learn PostgreSQL internals by building a lab" — the mission, the rules (§4, §5), the roadmap table. ends with `pglab ping` output.
- **Done when**: fresh clone + `docker compose up -d` + `go run ./cmd/pglab ping` works on your machine, post is live.

---

### Episode 1 — "Dead tuples: the ghost rows" (~1–2 weeks)

- **Concept**: MVCC basics. `UPDATE` in Postgres doesn't modify a row — it writes a new version and marks the old one dead. Read: MVCC chapter of Postgres docs + the PlanetScale article's vacuum section.
- **Build**: `pglab tables` — one query against `pg_stat_user_tables`: `relname, n_live_tup, n_dead_tup, last_vacuum, last_autovacuum, n_tup_ins/upd/del`, plus table size from `pg_total_relation_size`. Render as a table.
- **Predict**: "If I insert 1M rows, update all of them once, and never vacuum: `n_live_tup` ≈ 1M, `n_dead_tup` ≈ 1M, table size ≈ 2×."
- **Experiment**: raw psql or `pglab` inline SQL: create `events`, insert 1M rows, update every row, check `pglab tables` before/after. Record sizes.
- **Post**: prediction vs. reality; explain `xmin`/`xmax` for the first time using the actual row versions (query `xmin` directly with `SELECT xmin, * FROM events LIMIT 3` — this always lands well in a post).
- **Done when**: `pglab tables` works; post live; you can say from memory why an UPDATE doubles table size.

---

### Episode 2 — "Transactions: the reason vacuum can't help you" (~1–2 weeks)

- **Concept**: snapshots and the horizon. A transaction that started before your UPDATE still needs the old row versions → dead tuples are not reclaimable. Read: snapshot/visibility docs, `pg_stat_activity`.
- **Build**: `pglab transactions` — query `pg_stat_activity` for `pid, usename, state, xact_start, query_start, backend_xmin`; flag anything `idle in transaction` or older than N minutes (`--longer-than 5m` flag).
- **Predict**: "With an open transaction holding `backend_xmin = X`, updating 1M rows and running `VACUUM` leaves `n_dead_tup` ≈ 1M; after committing, `VACUUM` reclaims ~all."
- **Experiment**: two sessions (two `pglab` runs or psql): session A `BEGIN; SELECT 1;` (stays open), session B updates 1M rows, then `VACUUM (VERBOSE events)` → capture dead-tuples-removed = 0-ish. Commit A, vacuum again → all reclaimed.
- **Post**: the "lazy but harmless-looking transaction" story — this is the PlanetScale article's core lesson, now personally witnessed.
- **Done when**: `pglab transactions` flags the open session live during the experiment; post live.

---

### Episode 3 — "Watching a table rot in real time" (~2 weeks)

- **Concept**: autovacuum. Threshold math: `autovacuum_vacuum_threshold + autovacuum_vacuum_scale_factor × n_live_tup`. Why big tables vacuum "late".
- **Build**: `pglab vacuum --watch` — polls `pg_stat_user_tables` every second, redraws one table's `n_dead_tup` vs. computed autovacuum threshold, shows last autovacuum time. Also `pglab vacuum` (single shot) with status column: `OK / ⚠ OVERDUE`.
- **Predict**: "With default settings on a 1M-row table, autovacuum triggers only after ~205K dead tuples (scale factor 0.2). A hot loop of updates makes dead tuples oscillate — never zero."
- **Experiment**: run `pglab vacuum --watch events` in one terminal, a brutal `UPDATE ... WHERE id % 10 = 0` loop in another, and just watch for 15 minutes. Capture the transcript. Then `ALTER TABLE events SET (autovacuum_vacuum_scale_factor = 0.02)` and repeat. Compare.
- **Post**: the watching-a-graph-form post. Show the threshold math that explains everything you saw.
- **Done when**: you can compute the autovacuum trigger point for any table from memory; post live.

---

### Episode 4 — "The long-transaction experiment" (capstone build, ~2 weeks)

- **Concept**: everything from Episodes 1–3, composed into one scripted, reproducible scenario.
- **Build**: `pglab experiment vacuum` — orchestrates the full Chat-2 scenario:
  1. create fresh `events` table, insert 1M rows
  2. open a connection, `BEGIN; SELECT 1;` (the villain)
  3. update 200K rows
  4. run `VACUUM (VERBOSE)` in a second pool
  5. print before/after dead tuples + vacuum's own output
  6. prompt: `Commit the old transaction? [y/N]`
  7. commit, vacuum again, print the delta
- **This is the deliverable of Phase 1** — one command that teaches MVCC to anyone who runs it.
- **Post**: "I built a machine that demonstrates why idle transactions are dangerous" — mostly the transcript with narration.
- **Done when**: the experiment runs green from a fresh database on a fresh clone.

---

### Episode 5 — Phase 1 capstone: "MVCC, explained from memory" (~1 week)

- **Build**: nothing. Blog only.
- **Protocol**: apply §4's from-memory test — write the complete MVCC explainer with zero lookups first. Gaps → go run `pglab experiment vacuum` again. Then fill holes from docs, and cite your own experiment transcripts instead of the docs.
- **Post**: the reference post of the series. Link every claim back to an Episode 1–4 transcript.
- **Done when**: Phase 1 definition (§2) passes. This is the exit gate for Phase 2.

---

**Phase 1 total: ~8–10 weeks at your stated pace.** That is the honest number. Six posts, four commands, one deep concept.

---

## 9. Later phases (sketches — will be re-planned after Phase 1)

- **Phase 2 — Concurrency & bloat**: `pglab locks` (blocking chains via `pg_blocking_pids`), `pglab bloat` (start naive: dead-ratio heuristics, later pgstattuple). Concepts: lock modes, bloat vs. dead tuples, why `VACUUM FULL` locks everything.
- **Phase 3 — WAL & replication**: `pglab wal` (LSN, `pg_stat_wal`, `pg_stat_replication` with a real replica in docker-compose). Concepts: WAL volume under load, checkpoints, replica lag.
- **Phase 4 — The experiment framework**: `pglab load` (insert/update/delete/long-transaction generator), A/B harness for autovacuum settings and batch-delete sizes. This is where the lab becomes genuinely yours.
- **Phase 5 — Boss level**: partitioning benchmarks (unpartitioned vs. partitioned vacuum times), then a tiny Go sharding router over 3 Postgres containers. Only after this does a UI (dashboard over the collector) earn its place.

---

## 10. Episode post template

```markdown
# pg-lab #N: <concept title>

Prediction (written before running anything):
- I expect ... because ...

What I built:
- <command>, <the one catalog query that powers it>

The experiment:
- <setup, transcript, annotated>

What actually happened vs. prediction:
- ...

What I got wrong / what surprised me:
- ...

Try it yourself:
- <docker compose up; pglab ...>

Next: <teaser>
```

---

## 11. Next actions (this week)

1. First `/teach` session: create `MISSION.md` (from §2) and `RESOURCES.md` (Postgres MVCC docs, PlanetScale article, pg_stat_* docs)
2. `docker-compose.yml` + `go mod init` + `pglab ping` working (Episode 0 build)
3. Create the series landing page/tag on your site: *pg-lab — Learning PostgreSQL internals by building a lab, one experiment at a time*
4. Write and publish Episode 0 post from the template
5. When Episode 0 is live: `/teach` lesson on MVCC basics, then write Episode 1's prediction *before* writing `pglab tables`
