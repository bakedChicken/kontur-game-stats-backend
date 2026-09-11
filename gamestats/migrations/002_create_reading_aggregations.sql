CREATE VIEW game_servers AS
  SELECT
    server_id,
    endpoint,
    json_object(
      'name', metadata ->> '$.name',
      'gameModes', json(COALESCE(metadata -> '$.gameModes', json_array()))
    ) AS info
  FROM servers;

CREATE VIEW server_statistics AS
  SELECT
    s.server_id,
    s.endpoint,
    (SELECT COUNT(*) FROM matches m WHERE m.server_id = s.server_id)
      AS total_matches_played,
    (SELECT COALESCE(MAX(day_count), 0) FROM (
       SELECT COUNT(*) AS day_count FROM matches m
       WHERE m.server_id = s.server_id
       GROUP BY date(m.timestamp)
     )) AS maximum_matches_per_day,
    (SELECT COALESCE(AVG(day_count), 0) FROM (
       SELECT COUNT(*) AS day_count FROM matches m
       WHERE m.server_id = s.server_id
       GROUP BY date(m.timestamp)
     )) AS average_matches_per_day,
    (SELECT COALESCE(MAX(population), 0) FROM (
       SELECT COUNT(*) AS population FROM matches m
       JOIN player_performances pp USING (match_id)
       WHERE m.server_id = s.server_id
       GROUP BY m.match_id
     )) AS maximum_population,
    (SELECT COALESCE(AVG(population), 0) FROM (
       SELECT COUNT(*) AS population FROM matches m
       JOIN player_performances pp USING (match_id)
       WHERE m.server_id = s.server_id
       GROUP BY m.match_id
     )) AS average_population,
    (SELECT COALESCE(json_group_array(game_mode), json_array()) FROM (
       SELECT game_mode FROM matches m WHERE m.server_id = s.server_id
       GROUP BY game_mode ORDER BY COUNT(*) DESC, game_mode LIMIT 5
     )) AS top_5_game_modes,
    (SELECT COALESCE(json_group_array(map), json_array()) FROM (
       SELECT map FROM matches m WHERE m.server_id = s.server_id
       GROUP BY map ORDER BY COUNT(*) DESC, map LIMIT 5
     )) AS top_5_maps
  FROM servers s;

CREATE VIEW popular_servers AS
  SELECT
    s.endpoint,
    s.metadata ->> '$.name' AS name,
    ss.average_matches_per_day
  FROM servers s
  JOIN server_statistics ss USING (server_id)
  WHERE ss.total_matches_played > 0
  ORDER BY ss.average_matches_per_day DESC, s.endpoint
  LIMIT 50;

CREATE VIEW match_details AS
  SELECT
    s.endpoint,
    m.timestamp,
    json_object(
      'fragLimit', m.frag_limit,
      'timeLimit', m.time_limit,
      'timeElapsed', m.time_elapsed,
      'gameMode', m.game_mode,
      'map', m.map,
      'scoreboard', (
        SELECT json_group_array(
          json_object('name', pl.name, 'frags', pp.frags, 'kills', pp.kills, 'deaths', pp.deaths)
        )
        FROM player_performances pp
        JOIN players pl USING (player_id)
        WHERE pp.match_id = m.match_id
        ORDER BY pp.position
      )
    ) AS results
  FROM matches m
  JOIN servers s USING (server_id);

CREATE VIEW player_statistics AS
  SELECT
    p.player_id,
    p.name,
    (SELECT COUNT(*) FROM player_performances pp WHERE pp.player_id = p.player_id)
      AS total_matches_played,
    (SELECT COUNT(*) FROM player_performances pp WHERE pp.player_id = p.player_id AND pp.position = 0)
      AS total_matches_won,
    (SELECT COUNT(DISTINCT m.server_id) FROM player_performances pp
       JOIN matches m USING (match_id) WHERE pp.player_id = p.player_id)
      AS unique_servers,
    (SELECT COALESCE(MAX(day_count), 0) FROM (
       SELECT COUNT(*) AS day_count FROM player_performances pp
       JOIN matches m USING (match_id) WHERE pp.player_id = p.player_id
       GROUP BY date(m.timestamp)
     )) AS maximum_matches_per_day,
    (SELECT COALESCE(AVG(day_count), 0) FROM (
       SELECT COUNT(*) AS day_count FROM player_performances pp
       JOIN matches m USING (match_id) WHERE pp.player_id = p.player_id
       GROUP BY date(m.timestamp)
     )) AS average_matches_per_day,
    (SELECT MAX(m.timestamp) FROM player_performances pp
       JOIN matches m USING (match_id) WHERE pp.player_id = p.player_id)
      AS last_match_played,
    (SELECT COALESCE(AVG(
        CASE WHEN pop.population <= 1 THEN 100.0
             ELSE (pop.population - pp2.position - 1) * 100.0 / (pop.population - 1)
        END
      ), 0)
     FROM player_performances pp2
     JOIN (SELECT match_id, COUNT(*) AS population FROM player_performances GROUP BY match_id) pop
       USING (match_id)
     WHERE pp2.player_id = p.player_id
    ) AS average_scoreboard_percent,
    (SELECT m.game_mode FROM player_performances pp JOIN matches m USING (match_id)
     WHERE pp.player_id = p.player_id
     GROUP BY m.game_mode ORDER BY COUNT(*) DESC, m.game_mode LIMIT 1
    ) AS favorite_game_mode,
    (SELECT s.endpoint FROM player_performances pp
       JOIN matches m USING (match_id) JOIN servers s USING (server_id)
     WHERE pp.player_id = p.player_id
     GROUP BY s.endpoint ORDER BY COUNT(*) DESC, s.endpoint LIMIT 1
    ) AS favorite_server,
    (SELECT COALESCE(SUM(pp.deaths), 0) FROM player_performances pp WHERE pp.player_id = p.player_id)
      AS total_deaths,
    (SELECT CASE WHEN SUM(pp.deaths) > 0 THEN CAST(SUM(pp.kills) AS REAL) / SUM(pp.deaths) ELSE 0 END
     FROM player_performances pp WHERE pp.player_id = p.player_id
    ) AS kill_to_death_ratio
  FROM players p;

CREATE VIEW best_players AS
  SELECT name, kill_to_death_ratio
  FROM player_statistics
  WHERE total_matches_played >= 10 AND total_deaths >= 1
  ORDER BY kill_to_death_ratio DESC, name
  LIMIT 50;
