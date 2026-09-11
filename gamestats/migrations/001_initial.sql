PRAGMA journal_mode = WAL;
-- PRAGMA locking_mode = EXCLUSIVE;
-- PRAGMA foreign keys = ON;

CREATE TABLE servers (
  server_id INTEGER PRIMARY KEY,
  endpoint TEXT NOT NULL UNIQUE,
  metadata BLOB NOT NULL DEFAULT (json_object()),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE matches (
  match_id INTEGER PRIMARY KEY,
  server_id INTEGER NOT NULL REFERENCES servers,
  timestamp TEXT NOT NULL,
  frag_limit INTEGER NOT NULL,
  game_mode TEXT NOT NULL,
  map TEXT NOT NULL,
  time_elapsed NUMERIC NOT NULL,
  time_limit NUMERIC NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE (server_id, timestamp) ON CONFLICT IGNORE
);

CREATE INDEX matches_server_id_idx ON matches (server_id);
CREATE INDEX matches_timestamp_idx ON matches (timestamp);

CREATE TABLE players (
  player_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL COLLATE NOCASE UNIQUE ON CONFLICT IGNORE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE player_performances (
  match_id INTEGER NOT NULL REFERENCES matches,
  player_id INTEGER NOT NULL REFERENCES players,
  position INTEGER NOT NULL,
  deaths INTEGER NOT NULL,
  frags INTEGER NOT NULL,
  kills INTEGER NOT NULL,
  PRIMARY KEY (match_id, player_id)
);

CREATE INDEX player_performances_player_id_idx ON player_performances (player_id);
