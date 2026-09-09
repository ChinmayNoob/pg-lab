package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ChinmayNoob/pg-lab/internal/collector"
	"github.com/ChinmayNoob/pg-lab/internal/postgres"
	"github.com/ChinmayNoob/pg-lab/internal/render"
	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx := context.Background()

	var err error
	switch os.Args[1] {
	case "ping":
		err = ping(ctx, os.Args[2:])
	case "tables":
		err = tables(ctx, os.Args[2:])
	case "transactions":
		err = transactions(ctx, os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "pglab:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `pg-lab: a PostgreSQL internals laboratory

usage: pglab <command> [flags]

commands:
  ping        check connectivity to the lab database
  tables      dead/live tuple stats per table (pg_stat_user_tables)
  transactions  sessions holding/awaiting transactions (pg_stat_activity)`)
}

func transactions(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("transactions", flag.ExitOnError)
	dsn := fs.String("dsn", postgres.DefaultDSN, "postgres connection string")
	longer := fs.Duration("longer-than", 5*time.Minute, "flag idle-in-transaction sessions older than this")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, *dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	stats, err := collector.Transactions(ctx, pool)
	if err != nil {
		return fmt.Errorf("collect transactions: %w", err)
	}
	if len(stats) == 0 {
		fmt.Println("no other sessions")
		return nil
	}

	t := render.NewTable("PID", "USER", "STATE", "XACT AGE", "BACKEND_XMIN")
	for _, s := range stats {
		pid := strconv.Itoa(int(s.PID))
		if s.XactStart == nil {
			t.Row(pid, s.User, s.State, "-", s.BackendXmin)
			continue
		}
		age := time.Since(*s.XactStart)
		marker := ""
		if s.State == "idle in transaction" && age > *longer {
			marker = " ⚠"
		}
		t.Row(pid, s.User, s.State+marker, render.Duration(age), s.BackendXmin)
	}
	t.Flush()
	return nil
}

func tables(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("tables", flag.ContinueOnError)
	dsn := fs.String("dsn", postgres.DefaultDSN, "postgres connection string")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, *dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	stats, err := collector.TableStats(ctx, pool)
	if err != nil {
		return fmt.Errorf("collect table stats: %w", err)
	}
	if len(stats) == 0 {
		fmt.Println("no user tables found")
		return nil
	}
	render.Tables(stats)
	return nil
}

func ping(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("ping", flag.ContinueOnError)
	dsn := fs.String("dsn", postgres.DefaultDSN, "postgres connection string")
	if err := fs.Parse(args); err != nil {
		return err
	}

	start := time.Now()

	pool, err := postgres.NewPool(ctx, *dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	elapsed := time.Since(start)

	var serverVersion string
	if err := pool.QueryRow(ctx, "select version()").Scan(&serverVersion); err != nil {
		return fmt.Errorf("query version: %w", err)
	}

	var db struct {
		Name         string
		Backends     int
		XactCommit   int64
		XactRollback int64
		BlksHit      int64
		BlksRead     int64
		Deadlocks    int64
	}
	err = pool.QueryRow(ctx, `
		select datname, numbackends, xact_commit, xact_rollback,
		       blks_hit, blks_read, deadlocks
		from pg_stat_database
		where datname = current_database()`).Scan(
		&db.Name, &db.Backends, &db.XactCommit, &db.XactRollback,
		&db.BlksHit, &db.BlksRead, &db.Deadlocks)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("query pg_stat_database: %w", err)
	}

	fmt.Println("pg-lab ping")
	fmt.Println()
	fmt.Printf("server     %s\n", serverVersion)
	fmt.Printf("database   %s\n", db.Name)
	fmt.Printf("backends   %d\n", db.Backends)
	fmt.Printf("xacts      %s committed / %s rolled back\n",
		render.Comma(db.XactCommit), render.Comma(db.XactRollback))
	if hit, total := db.BlksHit, db.BlksHit+db.BlksRead; total > 0 {
		fmt.Printf("blocks     %.1f%% cache hit (%s hit / %s read)\n",
			100*float64(hit)/float64(total), render.Comma(hit), render.Comma(db.BlksRead))
	}
	fmt.Printf("deadlocks  %s\n", render.Comma(db.Deadlocks))
	fmt.Println()
	fmt.Printf("OK in %s\n", elapsed.Round(time.Millisecond))

	return nil
}
