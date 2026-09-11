package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const windowFunctionServerStatisticsView = `
CREATE VIEW server_statistics AS
  WITH match_days AS (
    SELECT server_id, date(timestamp) AS match_day, COUNT(*) AS day_count
    FROM matches
    GROUP BY server_id, match_day
  ),
  day_aggregates AS (
    SELECT server_id, MAX(day_count) AS maximum_matches_per_day, AVG(day_count) AS average_matches_per_day
    FROM match_days
    GROUP BY server_id
  ),
  match_populations AS (
    SELECT m.match_id, m.server_id, COUNT(*) AS population
    FROM matches m
    JOIN player_performances pp USING (match_id)
    GROUP BY m.match_id
  ),
  population_aggregates AS (
    SELECT server_id, MAX(population) AS maximum_population, AVG(population) AS average_population
    FROM match_populations
    GROUP BY server_id
  ),
  match_totals AS (
    SELECT server_id, COUNT(*) AS total_matches_played
    FROM matches
    GROUP BY server_id
  ),
  ranked_game_modes AS (
    SELECT server_id, game_mode,
           ROW_NUMBER() OVER (PARTITION BY server_id ORDER BY COUNT(*) DESC, game_mode) AS rank
    FROM matches
    GROUP BY server_id, game_mode
  ),
  top_game_modes AS (
    SELECT server_id, json_group_array(game_mode) AS game_modes
    FROM ranked_game_modes WHERE rank <= 5
    GROUP BY server_id
  ),
  ranked_maps AS (
    SELECT server_id, map,
           ROW_NUMBER() OVER (PARTITION BY server_id ORDER BY COUNT(*) DESC, map) AS rank
    FROM matches
    GROUP BY server_id, map
  ),
  top_maps AS (
    SELECT server_id, json_group_array(map) AS maps
    FROM ranked_maps WHERE rank <= 5
    GROUP BY server_id
  )
  SELECT
    s.server_id,
    s.endpoint,
    COALESCE(mt.total_matches_played, 0) AS total_matches_played,
    COALESCE(da.maximum_matches_per_day, 0) AS maximum_matches_per_day,
    COALESCE(da.average_matches_per_day, 0) AS average_matches_per_day,
    COALESCE(pa.maximum_population, 0) AS maximum_population,
    COALESCE(pa.average_population, 0) AS average_population,
    COALESCE(tgm.game_modes, json_array()) AS top_5_game_modes,
    COALESCE(tmp.maps, json_array()) AS top_5_maps
  FROM servers s
  LEFT JOIN match_totals mt USING (server_id)
  LEFT JOIN day_aggregates da USING (server_id)
  LEFT JOIN population_aggregates pa USING (server_id)
  LEFT JOIN top_game_modes tgm USING (server_id)
  LEFT JOIN top_maps tmp USING (server_id);
`

const windowFunctionPlayerStatisticsView = `
CREATE VIEW player_statistics AS
  WITH match_populations AS (
    SELECT match_id, COUNT(*) AS population
    FROM player_performances
    GROUP BY match_id
  ),
  appearances AS (
    SELECT pp.player_id, pp.position, pp.deaths, pp.kills, m.match_id, m.server_id, m.game_mode, m.timestamp,
           mp.population
    FROM player_performances pp
    JOIN matches m USING (match_id)
    JOIN match_populations mp USING (match_id)
  ),
  player_days AS (
    SELECT player_id, date(timestamp) AS play_day, COUNT(*) AS day_count
    FROM appearances
    GROUP BY player_id, play_day
  ),
  day_aggregates AS (
    SELECT player_id, MAX(day_count) AS maximum_matches_per_day, AVG(day_count) AS average_matches_per_day
    FROM player_days
    GROUP BY player_id
  ),
  totals AS (
    SELECT
      player_id,
      COUNT(*) AS total_matches_played,
      SUM(CASE WHEN position = 0 THEN 1 ELSE 0 END) AS total_matches_won,
      COUNT(DISTINCT server_id) AS unique_servers,
      MAX(timestamp) AS last_match_played,
      AVG(CASE WHEN population <= 1 THEN 100.0
               ELSE (population - position - 1) * 100.0 / (population - 1) END) AS average_scoreboard_percent,
      SUM(deaths) AS total_deaths,
      CASE WHEN SUM(deaths) > 0 THEN CAST(SUM(kills) AS REAL) / SUM(deaths) ELSE 0 END AS kill_to_death_ratio
    FROM appearances
    GROUP BY player_id
  ),
  ranked_game_modes AS (
    SELECT player_id, game_mode,
           ROW_NUMBER() OVER (PARTITION BY player_id ORDER BY COUNT(*) DESC, game_mode) AS rank
    FROM appearances
    GROUP BY player_id, game_mode
  ),
  ranked_servers AS (
    SELECT a.player_id, s.endpoint,
           ROW_NUMBER() OVER (PARTITION BY a.player_id ORDER BY COUNT(*) DESC, s.endpoint) AS rank
    FROM appearances a
    JOIN servers s USING (server_id)
    GROUP BY a.player_id, s.endpoint
  )
  SELECT
    p.player_id,
    p.name,
    t.total_matches_played,
    t.total_matches_won,
    t.unique_servers,
    da.maximum_matches_per_day,
    da.average_matches_per_day,
    t.last_match_played,
    t.average_scoreboard_percent,
    rgm.game_mode AS favorite_game_mode,
    rs.endpoint AS favorite_server,
    t.total_deaths,
    t.kill_to_death_ratio
  FROM players p
  JOIN totals t USING (player_id)
  JOIN day_aggregates da USING (player_id)
  JOIN ranked_game_modes rgm ON rgm.player_id = p.player_id AND rgm.rank = 1
  JOIN ranked_servers rs ON rs.player_id = p.player_id AND rs.rank = 1;
`

func newBenchDB(b *testing.B, useWindowFunctions bool) *sql.DB {
	b.Helper()

	db, err := OpenSQLite(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		b.Fatal(err)
	}

	if useWindowFunctions {
		for _, stmt := range []string{
			`DROP VIEW server_statistics`,
			windowFunctionServerStatisticsView,
			`DROP VIEW player_statistics`,
			windowFunctionPlayerStatisticsView,
		} {
			if _, err := db.Exec(stmt); err != nil {
				b.Fatal(err)
			}
		}
	}
	return db
}

func seedBenchData(b *testing.B, db *sql.DB, servers, matchesPerServer, playersPerMatch, playerPoolSize int) {
	b.Helper()
	rng := rand.New(rand.NewSource(42))
	gameModes := []string{"TDM", "DM", "RUSH", "CTF", "KOTH"}
	maps := []string{"DM-MAP1", "DM-MAP2", "DM-MAP3", "DM-MAP4", "DM-MAP5"}

	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}

	playerStmt, err := tx.Prepare(`
		INSERT INTO players (name) VALUES (?)
		ON CONFLICT (name) DO UPDATE SET name = players.name
		RETURNING player_id
	`)
	if err != nil {
		b.Fatal(err)
	}
	playerIDs := make(map[int]int64, playerPoolSize)
	playerID := func(n int) int64 {
		if id, ok := playerIDs[n]; ok {
			return id
		}
		var id int64
		if err := playerStmt.QueryRow(fmt.Sprintf("player-%d", n)).Scan(&id); err != nil {
			b.Fatal(err)
		}
		playerIDs[n] = id
		return id
	}

	epoch := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := 0

	for s := range servers {
		endpoint := fmt.Sprintf("server-%d.example.com-%d", s, 1024+s)
		var serverID int64
		err := tx.QueryRow(`
			INSERT INTO servers (endpoint, metadata) VALUES (?, json_object('name', ?, 'gameModes', json_array()))
			RETURNING server_id
		`, endpoint, fmt.Sprintf("Server %d", s)).Scan(&serverID)
		if err != nil {
			b.Fatal(err)
		}

		for range matchesPerServer {
			tick++
			timestamp := epoch.Add(time.Duration(tick) * time.Hour).Format(time.RFC3339)

			var matchID int64
			err := tx.QueryRow(`
				INSERT INTO matches (server_id, timestamp, frag_limit, game_mode, map, time_elapsed, time_limit)
				VALUES (?, ?, 20, ?, ?, 60, 20)
				RETURNING match_id
			`, serverID, timestamp, gameModes[rng.Intn(len(gameModes))], maps[rng.Intn(len(maps))]).Scan(&matchID)
			if err != nil {
				b.Fatal(err)
			}

			for p, n := range rng.Perm(playerPoolSize)[:playersPerMatch] {
				_, err := tx.Exec(`
					INSERT INTO player_performances (match_id, player_id, position, deaths, frags, kills)
					VALUES (?, ?, ?, ?, ?, ?)
				`, matchID, playerID(n), p, rng.Intn(15)+1, rng.Intn(20), rng.Intn(20))
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
}

func benchmarkBothViews(b *testing.B, query string) {
	for _, variant := range []struct {
		name               string
		useWindowFunctions bool
	}{
		{"Subqueries", false},
		{"CTEAndWindowFunctions", true},
	} {
		b.Run(variant.name, func(b *testing.B) {
			db := newBenchDB(b, variant.useWindowFunctions)
			defer db.Close()
			seedBenchData(b, db, 30, 150, 6, 300) // 30 servers, 4500 matches, 27000 performances

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rows, err := db.Query(query)
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

// BenchmarkPopularServers scans server_statistics for every server, the
// full-table-scan shape the /reports/popular-servers endpoint exercises.
func BenchmarkPopularServers(b *testing.B) {
	benchmarkBothViews(b, `SELECT * FROM popular_servers`)
}

// BenchmarkBestPlayers scans player_statistics for every player, the
// full-table-scan shape the /reports/best-players endpoint exercises.
func BenchmarkBestPlayers(b *testing.B) {
	benchmarkBothViews(b, `SELECT * FROM best_players`)
}

// BenchmarkSingleServerStats looks up one server's stats, the shape
// GET /servers/{endpoint}/stats exercises, to check whether the
// window-function cost shows up there too or only in full scans.
func BenchmarkSingleServerStats(b *testing.B) {
	benchmarkBothViews(b, `SELECT * FROM server_statistics WHERE endpoint = 'server-0.example.com-1024'`)
}
