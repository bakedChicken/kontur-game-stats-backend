package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func assert[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertDeepEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func newJSONRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()

	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			t.Fatalf("unable to encode request body: %v", err)
		}
	}

	request, err := http.NewRequest(method, url, buf)
	if err != nil {
		t.Fatalf("unable to create request: %v", err)
	}
	return request
}

func decodeJSON[T any](t *testing.T, body *bytes.Buffer) T {
	t.Helper()

	var got T
	if err := json.NewDecoder(body).Decode(&got); err != nil {
		t.Fatalf("unable to decode JSON response: %v", err)
	}
	return got
}

func TestHealthz(t *testing.T) {
	server := NewGameStatsServer(NewMemoryStore())

	request, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assert(t, response.Code, http.StatusOK)
	assert(t, response.Body.String(), "OK")
}

func TestUnknownRoute(t *testing.T) {
	server := NewGameStatsServer(NewMemoryStore())

	request, _ := http.NewRequest(http.MethodGet, "/non-existent", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assert(t, response.Code, http.StatusNotFound)
}

func TestServerRegistrationAndDiscovery(t *testing.T) {
	t.Run("no registered servers returns an empty array", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		request, _ := http.NewRequest(http.MethodGet, "/servers/info", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusOK)
		assertDeepEqual(t, decodeJSON[[]GameServer](t, response.Body), []GameServer{})
	})

	t.Run("an unregistered endpoint's info is not found", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		request, _ := http.NewRequest(http.MethodGet, "/servers/cs.net-1337/info", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusNotFound)
	})

	t.Run("registering a server makes it discoverable", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())
		info := ServerInformation{Name: "K1ll Zone", GameModes: []string{"TDM", "RUSH"}}

		putRequest := newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", info)
		putResponse := httptest.NewRecorder()
		server.ServeHTTP(putResponse, putRequest)
		assert(t, putResponse.Code, http.StatusOK)

		getInfoRequest, _ := http.NewRequest(http.MethodGet, "/servers/cs.net-1337/info", nil)
		getInfoResponse := httptest.NewRecorder()
		server.ServeHTTP(getInfoResponse, getInfoRequest)
		assert(t, getInfoResponse.Code, http.StatusOK)
		assertDeepEqual(t, decodeJSON[ServerInformation](t, getInfoResponse.Body), info)

		listRequest, _ := http.NewRequest(http.MethodGet, "/servers/info", nil)
		listResponse := httptest.NewRecorder()
		server.ServeHTTP(listResponse, listRequest)
		assert(t, listResponse.Code, http.StatusOK)
		assertDeepEqual(t, decodeJSON[[]GameServer](t, listResponse.Body), []GameServer{
			{Endpoint: "cs.net-1337", Info: info},
		})
	})

	t.Run("repeating the PUT replaces the stored metadata", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())
		original := ServerInformation{Name: "K1ll Zone", GameModes: []string{"TDM"}}
		updated := ServerInformation{Name: "K1ll Zone 2", GameModes: []string{"TDM", "DM"}}

		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", original))
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", updated))

		getRequest, _ := http.NewRequest(http.MethodGet, "/servers/cs.net-1337/info", nil)
		getResponse := httptest.NewRecorder()
		server.ServeHTTP(getResponse, getRequest)

		assertDeepEqual(t, decodeJSON[ServerInformation](t, getResponse.Body), updated)
	})
}

func exampleMatch() MatchInformation {
	return MatchInformation{
		FragLimit:   20,
		TimeLimit:   20,
		TimeElapsed: 56.321,
		GameMode:    "TDM",
		Map:         "DM-MAP11",
		Scoreboard: []PlayerResult{
			{Name: "Player1", Frags: 20, Kills: 21, Deaths: 3},
			{Name: "Player2", Frags: 15, Kills: 16, Deaths: 8},
		},
	}
}

func TestMatchIngestion(t *testing.T) {
	const endpoint = "cs.net-1337"
	const timestamp = "2017-01-22T15:17:00Z"

	registeredServer := func() *GameStatsServer {
		server := NewGameStatsServer(NewMemoryStore())
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/"+endpoint+"/info", ServerInformation{Name: "K1ll Zone"}))
		return server
	}

	t.Run("submitting a match for an unregistered endpoint is rejected", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		url := fmt.Sprintf("/servers/%s/matches/%s", endpoint, timestamp)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, newJSONRequest(t, http.MethodPut, url, exampleMatch()))

		assert(t, response.Code, http.StatusBadRequest)
	})

	t.Run("a submitted match can be fetched back by endpoint and timestamp", func(t *testing.T) {
		server := registeredServer()
		match := exampleMatch()
		url := fmt.Sprintf("/servers/%s/matches/%s", endpoint, timestamp)

		putResponse := httptest.NewRecorder()
		server.ServeHTTP(putResponse, newJSONRequest(t, http.MethodPut, url, match))
		assert(t, putResponse.Code, http.StatusOK)

		getRequest, _ := http.NewRequest(http.MethodGet, url, nil)
		getResponse := httptest.NewRecorder()
		server.ServeHTTP(getResponse, getRequest)

		assert(t, getResponse.Code, http.StatusOK)
		assertDeepEqual(t, decodeJSON[MatchInformation](t, getResponse.Body), match)
	})

	t.Run("an unsubmitted match is not found", func(t *testing.T) {
		server := registeredServer()

		request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/servers/%s/matches/%s", endpoint, timestamp), nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusNotFound)
	})

	t.Run("resubmitting a match does not replace the stored one", func(t *testing.T) {
		server := registeredServer()
		url := fmt.Sprintf("/servers/%s/matches/%s", endpoint, timestamp)

		first := exampleMatch()
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, first))

		second := exampleMatch()
		second.Map = "DM-MAP22"
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, second))

		getRequest, _ := http.NewRequest(http.MethodGet, url, nil)
		getResponse := httptest.NewRecorder()
		server.ServeHTTP(getResponse, getRequest)

		assertDeepEqual(t, decodeJSON[MatchInformation](t, getResponse.Body), first)
	})

	t.Run("a registered server with no matches has zero-value statistics", func(t *testing.T) {
		server := registeredServer()

		request, _ := http.NewRequest(http.MethodGet, "/servers/"+endpoint+"/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusOK)
		assertDeepEqual(t, decodeJSON[ServerStatistic](t, response.Body), ServerStatistic{
			Top5GameModes: []string{},
			Top5Maps:      []string{},
		})
	})

	t.Run("server statistics aggregate submitted matches", func(t *testing.T) {
		server := registeredServer()
		put := func(ts string, m MatchInformation) {
			url := fmt.Sprintf("/servers/%s/matches/%s", endpoint, ts)
			server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, m))
		}

		put("2017-01-22T15:17:00Z", exampleMatch())
		second := exampleMatch()
		second.Scoreboard = append(second.Scoreboard, PlayerResult{Name: "Player3", Frags: 1, Kills: 1, Deaths: 5})
		put("2017-01-22T16:00:00Z", second)
		put("2017-01-23T10:00:00Z", exampleMatch())

		request, _ := http.NewRequest(http.MethodGet, "/servers/"+endpoint+"/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		got := decodeJSON[ServerStatistic](t, response.Body)
		assert(t, got.TotalMatchesPlayed, 3)
		assert(t, got.MaximumMatchesPerDay, 2)
		assert(t, got.AverageMatchesPerDay, 1.5)
		assert(t, got.MaximumPopulation, 3)
	})

	t.Run("stats for an unregistered endpoint are not found", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		request, _ := http.NewRequest(http.MethodGet, "/servers/"+endpoint+"/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusNotFound)
	})
}

func TestPlayerStatistics(t *testing.T) {
	server := NewGameStatsServer(NewMemoryStore())
	server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", ServerInformation{Name: "K1ll Zone"}))

	match := exampleMatch() // Player1 wins, Player2 loses
	url := "/servers/cs.net-1337/matches/2017-01-22T15:17:00Z"
	server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, match))

	t.Run("lookup is case-insensitive and aggregates appearances", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/players/player1/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusOK)
		got := decodeJSON[PlayerStatistic](t, response.Body)
		assert(t, got.TotalMatchesPlayed, 1)
		assert(t, got.TotalMatchesWon, 1)
		assert(t, got.KillToDeathRatio, 21.0/3.0)
		assert(t, got.FavoriteServer, "cs.net-1337")
		assert(t, got.AverageScoreboardPercent, 100.0) // sole winner of a two-player scoreboard
	})

	t.Run("last place on a scoreboard scores zero percent", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/players/player2/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		got := decodeJSON[PlayerStatistic](t, response.Body)
		assert(t, got.TotalMatchesWon, 0)
		assert(t, got.AverageScoreboardPercent, 0.0)
	})

	t.Run("an unknown player is not found", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/players/nobody/stats", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Code, http.StatusNotFound)
	})
}

func TestReports(t *testing.T) {
	t.Run("reports are empty arrays before any data exists", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		for _, route := range []string{"/reports/recent-matches", "/reports/best-players", "/reports/popular-servers"} {
			request, _ := http.NewRequest(http.MethodGet, route, nil)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			assert(t, response.Code, http.StatusOK)
			if body := response.Body.String(); body != "[]\n" {
				t.Errorf("%s: got body %q, want an empty JSON array", route, body)
			}
		}
	})

	t.Run("recent matches are newest first and respect the count cap", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", ServerInformation{Name: "K1ll Zone"}))

		put := func(ts string) {
			url := "/servers/cs.net-1337/matches/" + ts
			server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, exampleMatch()))
		}
		put("2017-01-22T15:17:00Z")
		put("2017-01-23T15:17:00Z")
		put("2017-01-24T15:17:00Z")

		request, _ := http.NewRequest(http.MethodGet, "/reports/recent-matches/2", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		got := decodeJSON[[]RecentMatch](t, response.Body)
		if len(got) != 2 {
			t.Fatalf("got %d matches, want 2", len(got))
		}
		assert(t, got[0].Timestamp.Format("2006-01-02"), "2017-01-24")
		assert(t, got[1].Timestamp.Format("2006-01-02"), "2017-01-23")
	})

	t.Run("a zero or negative count returns an empty list", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())

		request, _ := http.NewRequest(http.MethodGet, "/reports/best-players/0", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Body.String(), "[]\n")
	})

	t.Run("best players requires at least ten matches and one death", func(t *testing.T) {
		server := NewGameStatsServer(NewMemoryStore())
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/info", ServerInformation{Name: "K1ll Zone"}))

		for day := 1; day <= 9; day++ {
			ts := fmt.Sprintf("2017-01-%02dT15:17:00Z", day)
			url := "/servers/cs.net-1337/matches/" + ts
			server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, url, exampleMatch()))
		}

		request, _ := http.NewRequest(http.MethodGet, "/reports/best-players", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assert(t, response.Body.String(), "[]\n")

		ts := "2017-01-10T15:17:00Z"
		server.ServeHTTP(httptest.NewRecorder(), newJSONRequest(t, http.MethodPut, "/servers/cs.net-1337/matches/"+ts, exampleMatch()))

		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)
		got := decodeJSON[[]BestPlayer](t, response.Body)
		if len(got) == 0 {
			t.Fatal("got no ranked players after ten matches, want Player1 and Player2 ranked")
		}
		assert(t, got[0].Name, "Player1")
	})
}
