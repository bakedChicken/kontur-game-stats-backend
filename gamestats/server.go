package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

const defaultReportCount = 5

type GameStatsServer struct {
	store  GameStatsStore
	logger *slog.Logger
	http.Handler
}

func NewGameStatsServer(store GameStatsStore) *GameStatsServer {
	s := &GameStatsServer{store: store, logger: slog.Default()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)

	mux.HandleFunc("GET /servers/info", s.listServers)
	mux.HandleFunc("GET /servers/{endpoint}/info", s.getServerInfo)
	mux.HandleFunc("PUT /servers/{endpoint}/info", s.putServerInfo)
	mux.HandleFunc("GET /servers/{endpoint}/stats", s.getServerStats)
	mux.HandleFunc("GET /servers/{endpoint}/matches/{timestamp}", s.getMatch)
	mux.HandleFunc("PUT /servers/{endpoint}/matches/{timestamp}", s.putMatch)

	mux.HandleFunc("GET /players/{name}/stats", s.getPlayerStats)

	mux.HandleFunc("GET /reports/recent-matches", s.recentMatches)
	mux.HandleFunc("GET /reports/recent-matches/{count}", s.recentMatches)
	mux.HandleFunc("GET /reports/best-players", s.bestPlayers)
	mux.HandleFunc("GET /reports/best-players/{count}", s.bestPlayers)
	mux.HandleFunc("GET /reports/popular-servers", s.popularServers)
	mux.HandleFunc("GET /reports/popular-servers/{count}", s.popularServers)

	s.Handler = mux
	return s
}

func (s *GameStatsServer) healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func (s *GameStatsServer) listServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.store.ListServers()
	if s.internalError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func (s *GameStatsServer) getServerInfo(w http.ResponseWriter, r *http.Request) {
	info, ok, err := s.store.GetServerInfo(r.PathValue("endpoint"))
	if s.internalError(w, err) {
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *GameStatsServer) putServerInfo(w http.ResponseWriter, r *http.Request) {
	var info ServerInformation
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if s.internalError(w, s.store.PutServerInfo(r.PathValue("endpoint"), info)) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *GameStatsServer) getServerStats(w http.ResponseWriter, r *http.Request) {
	stats, ok, err := s.store.GetServerStats(r.PathValue("endpoint"))
	if s.internalError(w, err) {
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *GameStatsServer) getMatch(w http.ResponseWriter, r *http.Request) {
	timestamp, err := parseTimestamp(r.PathValue("timestamp"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	match, ok, err := s.store.GetMatch(r.PathValue("endpoint"), timestamp)
	if s.internalError(w, err) {
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (s *GameStatsServer) putMatch(w http.ResponseWriter, r *http.Request) {
	endpoint := r.PathValue("endpoint")

	timestamp, err := parseTimestamp(r.PathValue("timestamp"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var match MatchInformation
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	registered, err := s.store.PutMatch(endpoint, timestamp, match)
	if s.internalError(w, err) {
		return
	}
	if !registered {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *GameStatsServer) getPlayerStats(w http.ResponseWriter, r *http.Request) {
	stats, ok, err := s.store.GetPlayerStats(r.PathValue("name"))
	if s.internalError(w, err) {
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *GameStatsServer) recentMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := s.store.RecentMatches(reportCount(r))
	if s.internalError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, matches)
}

func (s *GameStatsServer) bestPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := s.store.BestPlayers(reportCount(r))
	if s.internalError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, players)
}

func (s *GameStatsServer) popularServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.store.PopularServers(reportCount(r))
	if s.internalError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func reportCount(r *http.Request) int {
	raw := r.PathValue("count")
	if raw == "" {
		return defaultReportCount
	}

	count, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return count
}

func parseTimestamp(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func (s *GameStatsServer) internalError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	s.logger.Error("store operation failed", "error", err)
	w.WriteHeader(http.StatusInternalServerError)
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
