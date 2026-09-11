package main

import (
	"database/sql"
	"encoding/json"
	"time"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) ListServers() ([]GameServer, error) {
	rows, err := s.db.Query(`SELECT endpoint, info FROM game_servers ORDER BY endpoint`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []GameServer{}
	for rows.Next() {
		var endpoint, infoJSON string
		if err := rows.Scan(&endpoint, &infoJSON); err != nil {
			return nil, err
		}

		var info ServerInformation
		if err := json.Unmarshal([]byte(infoJSON), &info); err != nil {
			return nil, err
		}
		servers = append(servers, GameServer{Endpoint: endpoint, Info: info})
	}
	return servers, rows.Err()
}

func (s *SQLiteStore) GetServerInfo(endpoint string) (ServerInformation, bool, error) {
	var infoJSON string
	err := s.db.QueryRow(`SELECT info FROM game_servers WHERE endpoint = ?`, endpoint).Scan(&infoJSON)
	if err == sql.ErrNoRows {
		return ServerInformation{}, false, nil
	}
	if err != nil {
		return ServerInformation{}, false, err
	}

	var info ServerInformation
	if err := json.Unmarshal([]byte(infoJSON), &info); err != nil {
		return ServerInformation{}, false, err
	}
	return info, true, nil
}

func (s *SQLiteStore) PutServerInfo(endpoint string, info ServerInformation) error {
	if info.GameModes == nil {
		info.GameModes = []string{}
	}

	metadata, err := json.Marshal(info)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO servers (endpoint, metadata) VALUES (?, ?)
		ON CONFLICT (endpoint) DO UPDATE SET metadata = excluded.metadata
	`, endpoint, string(metadata))
	return err
}

func (s *SQLiteStore) GetServerStats(endpoint string) (ServerStatistic, bool, error) {
	var stat ServerStatistic
	var top5GameModesJSON, top5MapsJSON string
	err := s.db.QueryRow(`
		SELECT total_matches_played, maximum_matches_per_day, average_matches_per_day,
		       maximum_population, average_population, top_5_game_modes, top_5_maps
		FROM server_statistics WHERE endpoint = ?
	`, endpoint).Scan(
		&stat.TotalMatchesPlayed, &stat.MaximumMatchesPerDay, &stat.AverageMatchesPerDay,
		&stat.MaximumPopulation, &stat.AveragePopulation, &top5GameModesJSON, &top5MapsJSON,
	)
	if err == sql.ErrNoRows {
		return ServerStatistic{}, false, nil
	}
	if err != nil {
		return ServerStatistic{}, false, err
	}

	if err := json.Unmarshal([]byte(top5GameModesJSON), &stat.Top5GameModes); err != nil {
		return ServerStatistic{}, false, err
	}
	if err := json.Unmarshal([]byte(top5MapsJSON), &stat.Top5Maps); err != nil {
		return ServerStatistic{}, false, err
	}
	return stat, true, nil
}

func (s *SQLiteStore) GetMatch(endpoint string, timestamp time.Time) (MatchInformation, bool, error) {
	var resultsJSON string
	err := s.db.QueryRow(`
		SELECT results FROM match_details WHERE endpoint = ? AND timestamp = ?
	`, endpoint, formatTimestamp(timestamp)).Scan(&resultsJSON)
	if err == sql.ErrNoRows {
		return MatchInformation{}, false, nil
	}
	if err != nil {
		return MatchInformation{}, false, err
	}

	var match MatchInformation
	if err := json.Unmarshal([]byte(resultsJSON), &match); err != nil {
		return MatchInformation{}, false, err
	}
	return match, true, nil
}

func (s *SQLiteStore) PutMatch(endpoint string, timestamp time.Time, match MatchInformation) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var serverID int64
	err = tx.QueryRow(`SELECT server_id FROM servers WHERE endpoint = ?`, endpoint).Scan(&serverID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	res, err := tx.Exec(`
		INSERT INTO matches (server_id, timestamp, frag_limit, game_mode, map, time_elapsed, time_limit)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (server_id, timestamp) DO NOTHING
	`, serverID, formatTimestamp(timestamp), match.FragLimit, match.GameMode, match.Map, match.TimeElapsed, match.TimeLimit)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return true, tx.Commit()
	}

	matchID, err := res.LastInsertId()
	if err != nil {
		return false, err
	}

	for position, result := range match.Scoreboard {
		var playerID int64
		err := tx.QueryRow(`
			INSERT INTO players (name) VALUES (?)
			ON CONFLICT (name) DO UPDATE SET name = players.name
			RETURNING player_id
		`, result.Name).Scan(&playerID)
		if err != nil {
			return false, err
		}

		if _, err := tx.Exec(`
			INSERT INTO player_performances (match_id, player_id, position, deaths, frags, kills)
			VALUES (?, ?, ?, ?, ?, ?)
		`, matchID, playerID, position, result.Deaths, result.Frags, result.Kills); err != nil {
			return false, err
		}
	}

	return true, tx.Commit()
}

func (s *SQLiteStore) GetPlayerStats(name string) (PlayerStatistic, bool, error) {
	var stat PlayerStatistic
	var lastMatchPlayed string
	err := s.db.QueryRow(`
		SELECT total_matches_played, total_matches_won, unique_servers,
		       maximum_matches_per_day, average_matches_per_day, last_match_played,
		       average_scoreboard_percent, favorite_game_mode, favorite_server, kill_to_death_ratio
		FROM player_statistics WHERE name = ?
	`, name).Scan(
		&stat.TotalMatchesPlayed, &stat.TotalMatchesWon, &stat.UniqueServers,
		&stat.MaximumMatchesPerDay, &stat.AverageMatchesPerDay, &lastMatchPlayed,
		&stat.AverageScoreboardPercent, &stat.FavoriteGameMode, &stat.FavoriteServer, &stat.KillToDeathRatio,
	)
	if err == sql.ErrNoRows {
		return PlayerStatistic{}, false, nil
	}
	if err != nil {
		return PlayerStatistic{}, false, err
	}

	stat.LastMatchPlayed, err = time.Parse(time.RFC3339, lastMatchPlayed)
	if err != nil {
		return PlayerStatistic{}, false, err
	}
	return stat, true, nil
}

func (s *SQLiteStore) RecentMatches(count int) ([]RecentMatch, error) {
	rows, err := s.db.Query(`
		SELECT endpoint, timestamp, results FROM match_details
		ORDER BY timestamp DESC LIMIT ?
	`, clampReportCount(count))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := []RecentMatch{}
	for rows.Next() {
		var endpoint, timestampText, resultsJSON string
		if err := rows.Scan(&endpoint, &timestampText, &resultsJSON); err != nil {
			return nil, err
		}

		timestamp, err := time.Parse(time.RFC3339, timestampText)
		if err != nil {
			return nil, err
		}
		var results MatchInformation
		if err := json.Unmarshal([]byte(resultsJSON), &results); err != nil {
			return nil, err
		}
		matches = append(matches, RecentMatch{Endpoint: endpoint, Timestamp: timestamp, Results: results})
	}
	return matches, rows.Err()
}

func (s *SQLiteStore) BestPlayers(count int) ([]BestPlayer, error) {
	rows, err := s.db.Query(`SELECT name, kill_to_death_ratio FROM best_players LIMIT ?`, clampReportCount(count))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := []BestPlayer{}
	for rows.Next() {
		var p BestPlayer
		if err := rows.Scan(&p.Name, &p.KillToDeathRatio); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (s *SQLiteStore) PopularServers(count int) ([]PopularServer, error) {
	rows, err := s.db.Query(`SELECT endpoint, name, average_matches_per_day FROM popular_servers LIMIT ?`, clampReportCount(count))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []PopularServer{}
	for rows.Next() {
		var p PopularServer
		if err := rows.Scan(&p.Endpoint, &p.Name, &p.AverageMatchesPerDay); err != nil {
			return nil, err
		}
		servers = append(servers, p)
	}
	return servers, rows.Err()
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
