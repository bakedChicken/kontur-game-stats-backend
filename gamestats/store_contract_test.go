package main

import (
	"testing"
	"time"
)

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("unable to parse timestamp %q: %v", value, err)
	}
	return parsed
}

// testGameStatsStore runs the same behavioral contract against any
// GameStatsStore implementation, so MemoryStore and SQLiteStore are held
// to identical semantics instead of duplicating assertions per store.
func testGameStatsStore(t *testing.T, newStore func(t *testing.T) GameStatsStore) {
	t.Helper()

	t.Run("server registration is an upsert", func(t *testing.T) {
		store := newStore(t)

		_, found, err := store.GetServerInfo("cs.net-1337")
		assert(t, err, nil)
		assert(t, found, false)

		original := ServerInformation{Name: "K1ll Zone", GameModes: []string{"TDM"}}
		assert(t, store.PutServerInfo("cs.net-1337", original), nil)

		info, found, err := store.GetServerInfo("cs.net-1337")
		assert(t, err, nil)
		assert(t, found, true)
		assertDeepEqual(t, info, original)

		updated := ServerInformation{Name: "K1ll Zone 2", GameModes: []string{"TDM", "DM"}}
		assert(t, store.PutServerInfo("cs.net-1337", updated), nil)

		info, _, err = store.GetServerInfo("cs.net-1337")
		assert(t, err, nil)
		assertDeepEqual(t, info, updated)

		servers, err := store.ListServers()
		assert(t, err, nil)
		assertDeepEqual(t, servers, []GameServer{{Endpoint: "cs.net-1337", Info: updated}})
	})

	t.Run("matches require registration and resubmission is a no-op", func(t *testing.T) {
		store := newStore(t)
		timestamp := mustParseTime(t, "2017-01-22T15:17:00Z")
		match := exampleMatch()

		registered, err := store.PutMatch("cs.net-1337", timestamp, match)
		assert(t, err, nil)
		assert(t, registered, false)

		assert(t, store.PutServerInfo("cs.net-1337", ServerInformation{Name: "K1ll Zone"}), nil)

		registered, err = store.PutMatch("cs.net-1337", timestamp, match)
		assert(t, err, nil)
		assert(t, registered, true)

		got, found, err := store.GetMatch("cs.net-1337", timestamp)
		assert(t, err, nil)
		assert(t, found, true)
		assertDeepEqual(t, got, match)

		resubmission := exampleMatch()
		resubmission.Map = "DM-MAP99"
		registered, err = store.PutMatch("cs.net-1337", timestamp, resubmission)
		assert(t, err, nil)
		assert(t, registered, true)

		got, _, err = store.GetMatch("cs.net-1337", timestamp)
		assert(t, err, nil)
		assertDeepEqual(t, got, match) // unchanged: resubmission does not replace

		_, found, err = store.GetMatch("cs.net-1337", mustParseTime(t, "2099-01-01T00:00:00Z"))
		assert(t, err, nil)
		assert(t, found, false)
	})

	t.Run("server statistics aggregate submitted matches", func(t *testing.T) {
		store := newStore(t)
		assert(t, store.PutServerInfo("cs.net-1337", ServerInformation{Name: "K1ll Zone"}), nil)

		put := func(ts string, m MatchInformation) {
			t.Helper()
			registered, err := store.PutMatch("cs.net-1337", mustParseTime(t, ts), m)
			assert(t, err, nil)
			assert(t, registered, true)
		}

		put("2017-01-22T15:17:00Z", exampleMatch())
		second := exampleMatch()
		second.Scoreboard = append(second.Scoreboard, PlayerResult{Name: "Player3", Frags: 1, Kills: 1, Deaths: 5})
		put("2017-01-22T16:00:00Z", second)
		put("2017-01-23T10:00:00Z", exampleMatch())

		stats, found, err := store.GetServerStats("cs.net-1337")
		assert(t, err, nil)
		assert(t, found, true)
		assert(t, stats.TotalMatchesPlayed, 3)
		assert(t, stats.MaximumMatchesPerDay, 2)
		assert(t, stats.AverageMatchesPerDay, 1.5)
		assert(t, stats.MaximumPopulation, 3)
		assertDeepEqual(t, stats.Top5GameModes, []string{"TDM"})

		assert(t, store.PutServerInfo("empty.net-1337", ServerInformation{Name: "Ghost Town"}), nil)
		emptyStats, found, err := store.GetServerStats("empty.net-1337")
		assert(t, err, nil)
		assert(t, found, true)
		assertDeepEqual(t, emptyStats, ServerStatistic{Top5GameModes: []string{}, Top5Maps: []string{}})

		_, found, err = store.GetServerStats("nowhere-1")
		assert(t, err, nil)
		assert(t, found, false)
	})

	t.Run("player statistics are case-insensitive and score placements correctly", func(t *testing.T) {
		store := newStore(t)
		assert(t, store.PutServerInfo("cs.net-1337", ServerInformation{Name: "K1ll Zone"}), nil)

		registered, err := store.PutMatch("cs.net-1337", mustParseTime(t, "2017-01-22T15:17:00Z"), exampleMatch())
		assert(t, err, nil)
		assert(t, registered, true)

		winner, found, err := store.GetPlayerStats("player1") // different case than submitted "Player1"
		assert(t, err, nil)
		assert(t, found, true)
		assert(t, winner.TotalMatchesPlayed, 1)
		assert(t, winner.TotalMatchesWon, 1)
		assert(t, winner.AverageScoreboardPercent, 100.0)
		assert(t, winner.KillToDeathRatio, 21.0/3.0)
		assert(t, winner.FavoriteServer, "cs.net-1337")

		loser, found, err := store.GetPlayerStats("Player2")
		assert(t, err, nil)
		assert(t, found, true)
		assert(t, loser.TotalMatchesWon, 0)
		assert(t, loser.AverageScoreboardPercent, 0.0)

		_, found, err = store.GetPlayerStats("nobody")
		assert(t, err, nil)
		assert(t, found, false)
	})

	t.Run("reports respect ordering, thresholds, and count clamping", func(t *testing.T) {
		store := newStore(t)
		assert(t, store.PutServerInfo("cs.net-1337", ServerInformation{Name: "K1ll Zone"}), nil)

		empty, err := store.RecentMatches(5)
		assert(t, err, nil)
		assertDeepEqual(t, empty, []RecentMatch{})

		for day := 1; day <= 3; day++ {
			ts := mustParseTime(t, timeOnDay(day))
			registered, err := store.PutMatch("cs.net-1337", ts, exampleMatch())
			assert(t, err, nil)
			assert(t, registered, true)
		}

		recent, err := store.RecentMatches(2)
		assert(t, err, nil)
		if len(recent) != 2 {
			t.Fatalf("got %d recent matches, want 2", len(recent))
		}
		if !recent[0].Timestamp.After(recent[1].Timestamp) {
			t.Errorf("recent matches are not newest-first: %v then %v", recent[0].Timestamp, recent[1].Timestamp)
		}

		zero, err := store.RecentMatches(0)
		assert(t, err, nil)
		assertDeepEqual(t, zero, []RecentMatch{})

		capped, err := store.RecentMatches(1000)
		assert(t, err, nil)
		assert(t, len(capped), 3) // capped at what's available, well under the 50 ceiling

		bestBefore, err := store.BestPlayers(5)
		assert(t, err, nil)
		assertDeepEqual(t, bestBefore, []BestPlayer{})

		for day := 4; day <= 10; day++ {
			registered, err := store.PutMatch("cs.net-1337", mustParseTime(t, timeOnDay(day)), exampleMatch())
			assert(t, err, nil)
			assert(t, registered, true)
		}

		bestAfter, err := store.BestPlayers(5)
		assert(t, err, nil)
		if len(bestAfter) == 0 {
			t.Fatal("got no ranked players after ten matches, want Player1 ranked")
		}
		assert(t, bestAfter[0].Name, "Player1")

		popular, err := store.PopularServers(5)
		assert(t, err, nil)
		if len(popular) != 1 || popular[0].Endpoint != "cs.net-1337" {
			t.Errorf("got %+v, want cs.net-1337 as the sole popular server", popular)
		}
	})
}

func timeOnDay(day int) string {
	return time.Date(2017, 1, day, 15, 17, 0, 0, time.UTC).Format(time.RFC3339)
}
