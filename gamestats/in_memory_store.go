package main

import (
	"sort"
	"strings"
	"time"
)

// MemoryStore is used only for Unit Tests and as a stub for development
type MemoryStore struct {
	servers map[string]ServerInformation
	matches map[matchKey]MatchInformation
}

type matchKey struct {
	endpoint  string
	timestamp time.Time
}

type storedMatch struct {
	endpoint  string
	timestamp time.Time
	match     MatchInformation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		servers: map[string]ServerInformation{},
		matches: map[matchKey]MatchInformation{},
	}
}

func (s *MemoryStore) ListServers() ([]GameServer, error) {
	servers := make([]GameServer, 0, len(s.servers))
	for endpoint, info := range s.servers {
		servers = append(servers, GameServer{Endpoint: endpoint, Info: info})
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Endpoint < servers[j].Endpoint })
	return servers, nil
}

func (s *MemoryStore) GetServerInfo(endpoint string) (ServerInformation, bool, error) {
	info, ok := s.servers[endpoint]
	return info, ok, nil
}

func (s *MemoryStore) PutServerInfo(endpoint string, info ServerInformation) error {
	s.servers[endpoint] = info
	return nil
}

func (s *MemoryStore) GetMatch(endpoint string, timestamp time.Time) (MatchInformation, bool, error) {
	match, ok := s.matches[matchKey{endpoint, timestamp.UTC()}]
	return match, ok, nil
}

func (s *MemoryStore) PutMatch(endpoint string, timestamp time.Time, match MatchInformation) (bool, error) {
	if _, ok := s.servers[endpoint]; !ok {
		return false, nil
	}

	key := matchKey{endpoint, timestamp.UTC()}
	if _, exists := s.matches[key]; !exists {
		s.matches[key] = match
	}
	return true, nil
}

func (s *MemoryStore) allMatches() []storedMatch {
	matches := make([]storedMatch, 0, len(s.matches))
	for key, match := range s.matches {
		matches = append(matches, storedMatch{endpoint: key.endpoint, timestamp: key.timestamp, match: match})
	}
	return matches
}

func (s *MemoryStore) GetServerStats(endpoint string) (ServerStatistic, bool, error) {
	if _, ok := s.servers[endpoint]; !ok {
		return ServerStatistic{}, false, nil
	}

	var forServer []storedMatch
	for _, m := range s.allMatches() {
		if m.endpoint == endpoint {
			forServer = append(forServer, m)
		}
	}

	return buildServerStatistic(forServer), true, nil
}

func buildServerStatistic(matches []storedMatch) ServerStatistic {
	if len(matches) == 0 {
		return ServerStatistic{Top5GameModes: []string{}, Top5Maps: []string{}}
	}

	perDay := map[string]int{}
	gameModeCounts := map[string]int{}
	mapCounts := map[string]int{}
	totalPopulation, maxPopulation := 0, 0

	for _, m := range matches {
		perDay[m.timestamp.UTC().Format("2006-01-02")]++
		gameModeCounts[m.match.GameMode]++
		mapCounts[m.match.Map]++

		population := len(m.match.Scoreboard)
		totalPopulation += population
		if population > maxPopulation {
			maxPopulation = population
		}
	}

	maxPerDay := 0
	for _, count := range perDay {
		if count > maxPerDay {
			maxPerDay = count
		}
	}

	return ServerStatistic{
		TotalMatchesPlayed:   len(matches),
		MaximumMatchesPerDay: maxPerDay,
		AverageMatchesPerDay: float64(len(matches)) / float64(len(perDay)),
		MaximumPopulation:    maxPopulation,
		AveragePopulation:    float64(totalPopulation) / float64(len(matches)),
		Top5GameModes:        topNByCount(gameModeCounts, 5),
		Top5Maps:             topNByCount(mapCounts, 5),
	}
}

func topNByCount(counts map[string]int, n int) []string {
	type entry struct {
		name  string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, entry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})
	if len(entries) > n {
		entries = entries[:n]
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.name
	}
	return names
}

type playerAppearance struct {
	endpoint  string
	timestamp time.Time
	gameMode  string
	won       bool
	percent   float64
	kills     int
	deaths    int
}

func (s *MemoryStore) GetPlayerStats(name string) (PlayerStatistic, bool, error) {
	var appearances []playerAppearance
	for _, m := range s.allMatches() {
		size := len(m.match.Scoreboard)
		for i, result := range m.match.Scoreboard {
			if !strings.EqualFold(result.Name, name) {
				continue
			}
			appearances = append(appearances, playerAppearance{
				endpoint:  m.endpoint,
				timestamp: m.timestamp,
				gameMode:  m.match.GameMode,
				won:       i == 0,
				percent:   scoreboardPercent(i, size),
				kills:     result.Kills,
				deaths:    result.Deaths,
			})
		}
	}

	if len(appearances) == 0 {
		return PlayerStatistic{}, false, nil
	}

	return buildPlayerStatistic(appearances), true, nil
}

func scoreboardPercent(position, scoreboardSize int) float64 {
	if scoreboardSize <= 1 {
		return 100
	}
	return float64(scoreboardSize-position-1) / float64(scoreboardSize-1) * 100
}

func buildPlayerStatistic(appearances []playerAppearance) PlayerStatistic {
	perDay := map[string]int{}
	servers := map[string]int{}
	gameModes := map[string]int{}
	totalKills, totalDeaths, totalWon := 0, 0, 0
	totalPercent := 0.0
	var lastPlayed time.Time

	for _, a := range appearances {
		if a.won {
			totalWon++
		}
		perDay[a.timestamp.UTC().Format("2006-01-02")]++
		servers[a.endpoint]++
		gameModes[a.gameMode]++
		totalKills += a.kills
		totalDeaths += a.deaths
		totalPercent += a.percent
		if a.timestamp.After(lastPlayed) {
			lastPlayed = a.timestamp
		}
	}

	maxPerDay := 0
	for _, count := range perDay {
		if count > maxPerDay {
			maxPerDay = count
		}
	}

	var killToDeathRatio float64
	if totalDeaths > 0 {
		killToDeathRatio = float64(totalKills) / float64(totalDeaths)
	}

	return PlayerStatistic{
		TotalMatchesPlayed:       len(appearances),
		TotalMatchesWon:          totalWon,
		UniqueServers:            len(servers),
		MaximumMatchesPerDay:     maxPerDay,
		AverageMatchesPerDay:     float64(len(appearances)) / float64(len(perDay)),
		LastMatchPlayed:          lastPlayed.UTC(),
		AverageScoreboardPercent: totalPercent / float64(len(appearances)),
		FavoriteGameMode:         topNByCount(gameModes, 1)[0],
		FavoriteServer:           topNByCount(servers, 1)[0],
		KillToDeathRatio:         killToDeathRatio,
	}
}

func (s *MemoryStore) RecentMatches(count int) ([]RecentMatch, error) {
	matches := s.allMatches()
	sort.Slice(matches, func(i, j int) bool { return matches[i].timestamp.After(matches[j].timestamp) })

	count = boundedCount(count, len(matches))

	recent := make([]RecentMatch, count)
	for i := 0; i < count; i++ {
		recent[i] = RecentMatch{
			Endpoint:  matches[i].endpoint,
			Timestamp: matches[i].timestamp.UTC(),
			Results:   matches[i].match,
		}
	}
	return recent, nil
}

func (s *MemoryStore) BestPlayers(count int) ([]BestPlayer, error) {
	type totals struct {
		matches int
		kills   int
		deaths  int
	}
	perPlayer := map[string]*totals{}

	for _, m := range s.allMatches() {
		for _, result := range m.match.Scoreboard {
			t, ok := perPlayer[result.Name]
			if !ok {
				t = &totals{}
				perPlayer[result.Name] = t
			}
			t.matches++
			t.kills += result.Kills
			t.deaths += result.Deaths
		}
	}

	candidates := make([]BestPlayer, 0, len(perPlayer))
	for name, t := range perPlayer {
		if t.matches < 10 || t.deaths < 1 {
			continue
		}
		candidates = append(candidates, BestPlayer{
			Name:             name,
			KillToDeathRatio: float64(t.kills) / float64(t.deaths),
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].KillToDeathRatio != candidates[j].KillToDeathRatio {
			return candidates[i].KillToDeathRatio > candidates[j].KillToDeathRatio
		}
		return candidates[i].Name < candidates[j].Name
	})

	return candidates[:boundedCount(count, len(candidates))], nil
}

func (s *MemoryStore) PopularServers(count int) ([]PopularServer, error) {
	byServer := map[string][]storedMatch{}
	for _, m := range s.allMatches() {
		byServer[m.endpoint] = append(byServer[m.endpoint], m)
	}

	candidates := make([]PopularServer, 0, len(byServer))
	for endpoint, matches := range byServer {
		stat := buildServerStatistic(matches)
		candidates = append(candidates, PopularServer{
			Endpoint:             endpoint,
			Name:                 s.servers[endpoint].Name,
			AverageMatchesPerDay: stat.AverageMatchesPerDay,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].AverageMatchesPerDay != candidates[j].AverageMatchesPerDay {
			return candidates[i].AverageMatchesPerDay > candidates[j].AverageMatchesPerDay
		}
		return candidates[i].Endpoint < candidates[j].Endpoint
	})

	return candidates[:boundedCount(count, len(candidates))], nil
}

func boundedCount(count, available int) int {
	count = clampReportCount(count)
	if count > available {
		return available
	}
	return count
}
