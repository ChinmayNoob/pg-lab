package collector

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TableStat struct {
	Name           string
	LiveTuples     int64
	DeadTuples     int64
	LastVacuum     *time.Time
	LastAutovacuum *time.Time
	TotalSize      int64
}

func TableStats(ctx context.Context, pool *pgxpool.Pool) ([]TableStat, error) {
	rows, err := pool.Query(ctx, `
		select relname,
		       n_live_tup,
		       n_dead_tup,
		       last_vacuum,
		       last_autovacuum,
		       pg_total_relation_size(relid)
		from pg_stat_user_tables
		order by pg_total_relation_size(relid) desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []TableStat
	for rows.Next() {
		var s TableStat
		if err := rows.Scan(&s.Name, &s.LiveTuples, &s.DeadTuples,
			&s.LastVacuum, &s.LastAutovacuum, &s.TotalSize); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
