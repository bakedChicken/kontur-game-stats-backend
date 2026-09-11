package main

import (
	"database/sql"
	"testing"
)

func dropPerformanceIndexes(b *testing.B, db *sql.DB) {
	b.Helper()
	for _, stmt := range []string{
		`DROP INDEX matches_server_id_idx`,
		`DROP INDEX matches_timestamp_idx`,
		`DROP INDEX player_performances_player_id_idx`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkWithIndexes(b *testing.B, setup func(db *sql.DB) func() (*sql.Rows, error)) {
	for _, variant := range []struct {
		name        string
		withIndexes bool
	}{
		{"WithIndexes", true},
		{"WithoutIndexes", false},
	} {
		b.Run(variant.name, func(b *testing.B) {
			db := newBenchDB(b, false)
			defer db.Close()
			if !variant.withIndexes {
				dropPerformanceIndexes(b, db)
			}
			seedBenchData(b, db, 30, 150, 6, 300) // 30 servers, 4500 matches, 27000 performances

			query := setup(db)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rows, err := query()
				if err != nil {
					b.Fatal(err)
				}
				for rows.Next() {
				}
				if err := rows.Err(); err != nil {
					b.Fatal(err)
				}
				rows.Close()
			}
		})
	}
}

// BenchmarkServerStatsIndexes exercises matches_server_id_idx: every
// correlated subquery in server_statistics filters on server_id.
func BenchmarkServerStatsIndexes(b *testing.B) {
	benchmarkWithIndexes(b, func(db *sql.DB) func() (*sql.Rows, error) {
		return func() (*sql.Rows, error) {
			return db.Query(`SELECT * FROM server_statistics WHERE endpoint = 'server-0.example.com-1024'`)
		}
	})
}

// BenchmarkPlayerStatsIndexes exercises player_performances_player_id_idx:
// every correlated subquery in player_statistics filters on player_id.
func BenchmarkPlayerStatsIndexes(b *testing.B) {
	benchmarkWithIndexes(b, func(db *sql.DB) func() (*sql.Rows, error) {
		var samplePlayer string
		if err := db.QueryRow(`SELECT name FROM players LIMIT 1`).Scan(&samplePlayer); err != nil {
			b.Fatal(err)
		}
		return func() (*sql.Rows, error) {
			return db.Query(`SELECT * FROM player_statistics WHERE name = ?`, samplePlayer)
		}
	})
}

// BenchmarkRecentMatchesIndexes exercises matches_timestamp_idx: the
// recent-matches report is match_details ordered by timestamp descending.
func BenchmarkRecentMatchesIndexes(b *testing.B) {
	benchmarkWithIndexes(b, func(db *sql.DB) func() (*sql.Rows, error) {
		return func() (*sql.Rows, error) {
			return db.Query(`SELECT * FROM match_details ORDER BY timestamp DESC LIMIT 5`)
		}
	})
}
