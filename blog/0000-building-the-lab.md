# pg-lab #0: I'm going to learn PostgreSQL internals by building a lab

> Part of **pg-lab** — *Learning PostgreSQL internals by building a lab, one experiment at a time.*
> Episode 0 of Phase 1: MVCC & Vacuum.

I've read articles about PostgreSQL internals — MVCC, vacuum, dead tuples, why huge tables hurt. I understood maybe 40% of what I read. That number is about to be tested in public.

The plan for this series is not "build a monitoring tool." It's **build a laboratory**: every episode learns one concept, builds the smallest possible CLI feature that makes the concept *visible*, runs a controlled experiment, and reports prediction vs. reality.

## The rules I've committed to

1. **Predict before you run.** Every experiment starts with a written prediction. If the prediction is wrong, that gap is the actual lesson.
2. **The exit test is from-memory.** Phase 1 ends when I can whiteboard MVCC — `xmin`/`xmax`, snapshots, dead tuples, why vacuum stalls — without opening a browser, demonstrating each claim live with my own tool.
3. **Small episodes, hard cap.** 3–6 hours a week. An episode takes max two weeks or gets cut.
4. **Publish before perfect.** Every post ends with "What I got wrong."

## The lab itself (this episode)

Phase 0 is deliberately boring infrastructure:

- one disposable `postgres:17` container (docker compose, fresh volume between experiments)
- one Go binary, `pglab`, that talks to it via `pgx`
- one command so far: `pglab ping`

The compose file is thirty lines with a healthcheck so `docker compose up -d` is deterministic:

```yaml
services:
  lab:
    image: postgres:17
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pglab -d pglab"]
      interval: 2s
      retries: 15
```

And the first command reads `pg_stat_database` — the first catalog table of many you'll see in this series:

```sql
select datname, numbackends, xact_commit, xact_rollback,
       blks_hit, blks_read, deadlocks
from pg_stat_database
where datname = current_database();
```

## First light

Fresh clone → `docker compose up -d` → `go run ./cmd/pglab ping`:

```text
pg-lab ping

server     PostgreSQL 17.11 (Debian 17.11-1.pgdg13+2) on x86_64-pc-linux-gnu, compiled by gcc (Debian 14.2.0-19) 14.2.0, 64-bit
database   pglab
backends   1
xacts      6 committed / 0 rolled back
blocks     95.5% cache hit (1,740 hit / 82 read)
deadlocks  0

OK in 39ms
```

Nothing here is impressive yet. That's the point: this is the control room. Everything interesting in this series happens by pointing more commands at this same database and watching what changes.

## What I got wrong

- (draft note: first `go get` pulled `pgx/v5` but a bare `go build` failed until `go mod tidy` — the pool package needs `puddle` transitively. Small, but worth recording: let the toolchain resolve the module graph, don't hand-pick.)

## Try it yourself

```bash
git clone https://github.com/ChinmayNoob/pg-lab
cd pg-lab
docker compose up -d
go run ./cmd/pglab ping
```

## Next

**pg-lab #1: Dead tuples — the ghost rows.** Prediction already written: if I insert 1M rows, update all of them once and never vacuum, the table should be ~2× its original size and `pg_stat_user_tables.n_dead_tup` should read ~1M. Let's find out.
