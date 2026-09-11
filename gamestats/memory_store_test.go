package main

import "testing"

func TestMemoryStore(t *testing.T) {
	testGameStatsStore(t, func(t *testing.T) GameStatsStore {
		return NewMemoryStore()
	})
}
