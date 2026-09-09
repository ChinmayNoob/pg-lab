# pg-lab Lab Manual

Every terminal command used in the lab, what it does, how to read its output, and the information worth keeping. New commands get added here as the series grows. (Predictions & transcripts live in `blog/experiments/`; this file is the permanent reference.)

---

## 1. Lab control (Docker)

### `docker compose up -d`
Starts the lab: one PostgreSQL 17 container (`pglab`) on port 5432 with a healthcheck.
- **Read it as**: "the database is starting in the background." Run it after `down -v` or on a fresh clone.
- **Info**: `-d` = detached (doesn't tie up your terminal).

### `docker compose ps`
Shows container status.
- **Healthy** = ready. **health: starting** = wait a few seconds.

### `docker compose down -v`
Stops the container **and deletes the data volume**.
- **Read it as**: "wipe the lab clean." This is our freshness rule — every experiment starts from zero state, so results are reproducible.
- **Info**: without `-v`, data survives restarts; with it, everything (tables, transactions, stats history) is gone.

---

## 2. Talking to Postgres (psql inside the container)

### `docker exec -it pglab psql -U pglab -d pglab -c "<SQL>"`
Runs one SQL statement inside the container.
- `-it` interactive terminal, `-U pglab` user, `-d pglab` database, `-c` command.
- **Read the output tags**: `DROP TABLE` / `UPDATE 3` (rows affected) / `SELECT n` (rows returned). `NOTICE: table "t" does not exist, skipping` is informational, not an error.

### `SELECT xmin, xmax, * FROM <table> LIMIT 5;`
Shows each row version's header fields: `xmin` = transaction that created this version, `xmax` = transaction that replaced/deleted it (0 = nobody yet).
- **Info**: all rows inserted by one statement share the same xmin — one statement = one transaction = one transaction ID.

### `CREATE TABLE t AS SELECT generate_series(1,1000000) AS id, md5(random()::text) AS payload;`
Creates and fills a table in one statement. `generate_series(n,m)` produces m−n+1 integers; `md5(random()::text)` fakes a 32-char text payload per row.
- **Info**: `CREATE TABLE AS SELECT` (CTAS) is the fastest way to fabricate test data; it skips per-row triggers/foreign keys.

### `UPDATE <table> SET col = col;`
Rewrites every row — even though no value actually changes.
- **Info**: in MVCC, an UPDATE that changes nothing still writes a full new row version per row. This is the cheapest way to manufacture dead tuples at scale.

### `ALTER TABLE <table> SET (autovacuum_enabled = false);`
Turns off autovacuum **for that table only**.
- **Why we use it**: autovacuum is a background janitor that would clean our dead tuples mid-experiment and ruin the measurement. Disabling it per-table gives us controlled conditions — *we* decide when vacuum runs.

### `VACUUM <table>;`
Manually removes dead tuples. Their space is **recycled for future rows** but normally **not returned to the OS** — the table file typically keeps its size.
- **Info**: `VACUUM` works in the background without blocking reads/writes (`VACUUM FULL` rewrites the whole table and shrinks it, but takes an exclusive lock — a later episode).

---

## 3. Reading `pglab` output

### `go run ./cmd/pglab ping`
Connectivity + server identity + `pg_stat_database` row: backends (open connections), committed/rolled-back transactions, cache hit %, deadlocks.

### `go run ./cmd/pglab tables`
One row per user table, from **`pg_stat_user_tables`**:

| Column in output | Underlying stat | Meaning |
|---|---|---|
| `LIVE` | `n_live_tup` | Estimated row versions visible to current snapshots |
| `DEAD` | `n_dead_tup` | Estimated dead tuples waiting for vacuum |
| `DEAD %` | computed | dead / (live + dead) — bloat pressure in one number |
| `LAST VACUUM` | `last_vacuum` | Last **manual** vacuum |
| `LAST AUTOVACUUM` | `last_autovacuum` | Last background autovacuum |
| `SIZE` | `pg_total_relation_size` | Table + indexes + toast, on disk |

- **Info**: these are *estimates* maintained by the stats collector — exact after bulk operations, approximate during steady churn. Good enough to reason with, exactly like production monitoring does.

---

## 4. Lab conventions

- **Prediction before running.** Write expected numbers in `blog/experiments/NNNN-*.txt` before executing anything.
- **Transcripts, not screenshots.** Paste real terminal output under a `TRANSCRIPT` heading.
- **Fresh DB per experiment**: `docker compose down -v && docker compose up -d`.
- **Observation after running**: what matched, what surprised you — written before consulting docs.
