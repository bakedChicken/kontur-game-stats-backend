package main

import "time"

// ServerInformation is the metadata a server registers about itself.
type ServerInformation struct {
	Name      string   `json:"name"`
	GameModes []string `json:"gameModes"`
}

// GameServer is a registered server and its public metadata.
type GameServer struct {
	Endpoint string            `json:"endpoint"`
	Info     ServerInformation `json:"info"`
}

// PlayerResult is one player's result in final-scoreboard order.
type PlayerResult struct {
	Name   string `json:"name"`
	Frags  int    `json:"frags"`
	Kills  int    `json:"kills"`
	Deaths int    `json:"deaths"`
}

// MatchInformation is a match's settings and ordered final scoreboard.
// The first entry in Scoreboard is the winner.
type MatchInformation struct {
	FragLimit   int            `json:"fragLimit"`
	TimeLimit   int            `json:"timeLimit"`
	TimeElapsed float64        `json:"timeElapsed"`
	GameMode    string         `json:"gameMode"`
	Map         string         `json:"map"`
	Scoreboard  []PlayerResult `json:"scoreboard"`
}

// PlayerStatistic holds a player's aggregates across every submitted match.
type PlayerStatistic struct {
	TotalMatchesPlayed       int       `json:"totalMatchesPlayed"`
	TotalMatchesWon          int       `json:"totalMatchesWon"`
	UniqueServers            int       `json:"uniqueServers"`
	MaximumMatchesPerDay     int       `json:"maximumMatchesPerDay"`
	AverageMatchesPerDay     float64   `json:"averageMatchesPerDay"`
	LastMatchPlayed          time.Time `json:"lastMatchPlayed"`
	AverageScoreboardPercent float64   `json:"averageScoreboardPercent"`
	FavoriteGameMode         string    `json:"favoriteGameMode"`
	FavoriteServer           string    `json:"favoriteServer"`
	KillToDeathRatio         float64   `json:"killToDeathRatio"`
}

// ServerStatistic holds aggregates for matches submitted to one server.
type ServerStatistic struct {
	TotalMatchesPlayed   int      `json:"totalMatchesPlayed"`
	MaximumMatchesPerDay int      `json:"maximumMatchesPerDay"`
	AverageMatchesPerDay float64  `json:"averageMatchesPerDay"`
	MaximumPopulation    int      `json:"maximumPopulation"`
	AveragePopulation    float64  `json:"averagePopulation"`
	Top5GameModes        []string `json:"top5GameModes"`
	Top5Maps             []string `json:"top5Maps"`
}

// RecentMatch is a match entry in the recent-matches report.
type RecentMatch struct {
	Endpoint  string           `json:"endpoint"`
	Timestamp time.Time        `json:"timestamp"`
	Results   MatchInformation `json:"results"`
}

// BestPlayer is a player entry in the best-players report.
type BestPlayer struct {
	Name             string  `json:"name"`
	KillToDeathRatio float64 `json:"killToDeathRatio"`
}

// PopularServer is a server entry in the popular-servers report.
type PopularServer struct {
	Endpoint             string  `json:"endpoint"`
	Name                 string  `json:"name"`
	AverageMatchesPerDay float64 `json:"averageMatchesPerDay"`
}
