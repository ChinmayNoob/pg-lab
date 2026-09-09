package collector

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionStat struct {
	PID         int32
	User        string
	State       string
	XactStart   *time.Time
	QueryStart  *time.Time
	BackendXmin string
}

func Transactions(ctx context.Context, pool *pgxpool.Pool) ([]TransactionStat, error) {
	rows, err := pool.Query(ctx, `
		select pid,
		       coalesce(usename, '-'),
		       coalesce(state, '-'),
		       xact_start,
		       query_start,
		       coalesce(backend_xmin::text, '-')
		from pg_stat_activity
		where pid <> pg_backend_pid()
		  and datname = current_database()
		order by xact_start asc nulls last`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []TransactionStat
	for rows.Next() {
		var s TransactionStat
		if err := rows.Scan(&s.PID, &s.User, &s.State,
			&s.XactStart, &s.QueryStart, &s.BackendXmin); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
