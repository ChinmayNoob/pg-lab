# PostgreSQL Internals Resources

Trusted sources for the pg-lab teaching workspace. Lesson content must be grounded here, not guessed.

## Knowledge

- [PostgreSQL Docs: Concurrency Control / MVCC (Ch. 13)](https://www.postgresql.org/docs/17/mvcc.html)
  The primary source for MVCC, transaction isolation, and visibility. Use for: xmin/xmax semantics, snapshots, serialization — every claim in lessons should trace here first.
- [PostgreSQL Docs: Routine Vacuuming (Ch. 25)](https://www.postgresql.org/docs/17/routine-vacuuming.html)
  VACUUM, autovacuum, the visibility map, and the threshold math. Use for: dead-tuple reclamation, autovacuum trigger points, vacuum cost limits.
- [PostgreSQL Docs: The Statistics Collector / pg_stat_* views](https://www.postgresql.org/docs/17/monitoring-stats.html)
  Every column `pglab` will ever read: `pg_stat_user_tables`, `pg_stat_activity`, `pg_stat_database`. Use for: collector query design.
- [Article: "Dealing With Large Tables in PostgreSQL" (PlanetScale blog)](https://planetscale.com/blog/dealing-with-large-tables-in-postgres)
  The article that started this project: large tables → vacuum lag → long transactions → WAL → partitioning → sharding. Use for: the experimental narrative of the series, motivation.
- [Book (free, online): _The Internals of PostgreSQL_ by InterDB](http://www.interdb.jp/pg/)
  Chapter-level deep dives with heap page and tuple header layouts. Use for: what a tuple header actually stores, WAL internals (Phase 3), vacuum internals. Heavier — go here when the official docs feel too abstract.

## Wisdom (Communities)

- [r/PostgreSQL](https://www.reddit.com/r/PostgreSQL/)
  Active, practitioner-heavy. Use for: reality-checking explanations, seeing real production failures others hit.
- [PostgreSQL Discord](https://discord.com/invite/postgresql)
  Fast answers from experienced users. Use for: unblocking experiments when catalog behavior confuses.
- [pgsql-general mailing list](https://www.postgresql.org/list/pgsql-general/)
  Where core-adjacent experts answer. Use for: subtle MVCC/vacuum questions the docs don't settle.

## Gaps

- No trusted, current resource yet for **bloat estimation techniques** (needed for Phase 2 `pglab bloat`) — search for pgstattuple-based approaches and the classic `pg_bloat_check` scripts when Phase 2 nears.
- No resource yet for **replication setup in Docker** (Phase 3 will need a working compose topology).
