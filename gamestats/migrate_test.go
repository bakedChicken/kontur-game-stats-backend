package main

import "testing"

func TestMigrate(t *testing.T) {
	t.Run("creates every table and view", func(t *testing.T) {
		db := newTestSQLiteDB(t)

		expected := []string{
			"servers", "matches", "players", "player_performances", // tables
			"game_servers", "server_statistics", "popular_servers", // views
			"match_details", "player_statistics", "best_players",
		}

		for _, name := range expected {
			var exists bool
			err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE name = ?)`, name).Scan(&exists)
			assert(t, err, nil)
			if !exists {
				t.Errorf("expected sqlite_master to contain %q, it doesn't", name)
			}
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		db := newTestSQLiteDB(t)

		if err := Migrate(db); err != nil {
			t.Fatalf("second Migrate call failed: %v", err)
		}

		var applied int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
			t.Fatalf("unable to count schema_migrations: %v", err)
		}
		assert(t, applied, 2)
	})
}
