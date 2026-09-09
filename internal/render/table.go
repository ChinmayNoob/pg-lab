package render

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ChinmayNoob/pg-lab/internal/collector"
)

func Comma(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}

func HumanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func RelTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "never"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func Tables(stats []collector.TableStat) {
	t := NewTable("TABLE", "SIZE", "LIVE", "DEAD", "DEAD %", "LAST VACUUM", "LAST AUTOVACUUM")
	for _, s := range stats {
		total := s.LiveTuples + s.DeadTuples
		deadPct := 0.0
		if total > 0 {
			deadPct = 100 * float64(s.DeadTuples) / float64(total)
		}
		t.Row(s.Name, HumanBytes(s.TotalSize), Comma(s.LiveTuples), Comma(s.DeadTuples),
			fmt.Sprintf("%.1f%%", deadPct), RelTime(s.LastVacuum), RelTime(s.LastAutovacuum))
	}
	t.Flush()
}

type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

func (t *Table) Row(cells ...string) {
	t.rows = append(t.rows, cells)
}

func (t *Table) Flush() {
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, r := range t.rows {
		for i, c := range r {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	printRow := func(cells []string, sep string) {
		for i, c := range cells {
			fmt.Printf("%-*s%s", widths[i], c, sep)
		}
		fmt.Println()
	}
	printRow(t.headers, "  ")
	for _, r := range t.rows {
		printRow(r, "  ")
	}
}
