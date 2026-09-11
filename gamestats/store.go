package main

import "time"

// ServerStore registers servers and reports their metadata and statistics.
type ServerStore interface {
	ListServers() ([]GameServer, error)
	GetServerInfo(endpoint string) (info ServerInformation, found bool, err error)
	PutServerInfo(endpoint string, info ServerInformation) error
	GetServerStats(endpoint string) (stats ServerStatistic, found bool, err error)
}

// MatchStore stores completed matches submitted by registered servers.
type MatchStore interface {
	GetMatch(endpoint string, timestamp time.Time) (match MatchInformation, found bool, err error)
	PutMatch(endpoint string, timestamp time.Time, match MatchInformation) (registered bool, err error)
}

// PlayerStore reports statistics derived from submitted scoreboard entries.
type PlayerStore interface {
	GetPlayerStats(name string) (stats PlayerStatistic, found bool, err error)
}

// ReportStore serves the precomputed leaderboard and activity reports.
type ReportStore interface {
	RecentMatches(count int) ([]RecentMatch, error)
	BestPlayers(count int) ([]BestPlayer, error)
	PopularServers(count int) ([]PopularServer, error)
}

// GameStatsStore is the full persistence contract GameStatsServer depends on.
type GameStatsStore interface {
	ServerStore
	MatchStore
	PlayerStore
	ReportStore
}

// clampReportCount applies the ReportCount contract shared by every report
// endpoint: non-positive values report nothing, and values above 50 are
// capped at 50.
func clampReportCount(count int) int {
	if count <= 0 {
		return 0
	}
	if count > 50 {
		return 50
	}
	return count
}
