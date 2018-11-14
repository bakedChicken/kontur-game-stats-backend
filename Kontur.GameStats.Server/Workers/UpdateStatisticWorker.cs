using System;
using System.Collections.Generic;
using System.Linq;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Workers {
	public class UpdateStatisticWorker {
		private const int TopSize = 5;
		
		public void UpdateUserStatistic(MatchInformation info, string endpoint, DateTimeOffset timestamp) {
			var count = info.PlayerResults.Count;
			var playerStatistics = DefaultDatabase.Instance.GetCollection<PlayerStatistic>();
		    var statistics = new HashSet<PlayerStatistic>();

			for (var i = 0; i < count; i++) {
				var result = info.PlayerResults[i];
				var statistic = playerStatistics.FindById(result.Name) ?? new PlayerStatistic { Name = result.Name };

				#region Update Player Statistic

				statistic.Servers.IncrementValue(endpoint);
				statistic.GameModes.IncrementValue(info.GameMode);
				statistic.Timestamps.IncrementValue(timestamp.Date.Ticks);

				statistic.TotalMatchesPlayed += 1;
				statistic.TotalMatchesWon += i == 0 ? 1 : 0;
				statistic.TotalKills += result.Kills;
				statistic.TotalDeaths += result.Deaths;

				if (timestamp > statistic.LastMatchPlayed)
					statistic.LastMatchPlayed = timestamp;

				statistic.KillToDeathRatio = (double) statistic.TotalKills / statistic.TotalDeaths;
			    statistic.ScoreboardPercents.Add(count == 1 ? 100 : (double) (count - i - 1) / (count - 1) * 100);
			    statistic.AverageScoreboardPercent = statistic.ScoreboardPercents.Average();
				statistic.MaximumMatchesPerDay = statistic.Timestamps.Max(x => x.Value);
				statistic.AverageMatchesPerDay = statistic.Timestamps.Average(x => x.Value);
				statistic.UniqueServers = statistic.Servers.Count;
				statistic.FavoriteServer = statistic.Servers.Aggregate((l, r) => l.Value > r.Value ? l : r).Key;
				statistic.FavoriteGameMode = statistic.GameModes.Aggregate((l, r) => l.Value > r.Value ? l : r).Key;

				#endregion

			    statistics.Add(statistic);
			}

            playerStatistics.Upsert(statistics);
        }

	    private void UpdateServerStatistic(string endpoint) {
			var db = DefaultDatabase.Instance;
			var servers = db.GetCollection<GameServer>();
			var server = servers.FindById(endpoint);
			var statistic = server.ServerStatistic;
			var matches = db.GetCollection<Match>().Find(m => m.MatchId.Endpoint == server.Endpoint).ToList();

		    statistic.TopFiveGameModes = matches
                .Select(m => m.MatchInformation.GameMode)
                .GroupBy(g => g)
                .OrderByDescending(g => g.Count())
                .Select(g => g.Key)
                .Take(TopSize);
		    statistic.TopFiveMaps = matches
                .Select(m => m.MatchInformation.Map)
		        .GroupBy(g => g)
		        .OrderByDescending(g => g.Count())
		        .Select(g => g.Key)
		        .Take(TopSize);

			statistic.TotalMatchesPlayed = matches.Count;
			statistic.MaximumMatchesPerDay = matches.GroupBy(m => m.MatchId.Timestamp.Date).Max(g => g.Count());
			statistic.AverageMatchesPerDay = matches.GroupBy(m => m.MatchId.Timestamp.Date).Average(g => g.Count());
			statistic.MaximumPopulation = matches.Select(m => m.MatchInformation.PlayerResults.Count).Max();
			statistic.AveragePopulation = matches.Select(m => m.MatchInformation.PlayerResults.Count).Average();

			servers.Update(server);
		}

	    public void UpdateDirtyServers() {
	        var db = DefaultDatabase.Instance;
            var dirtyServers = db.GetCollection<DirtyServer>();
	        var servers = dirtyServers.FindAll();

	        using (var transaction = db.GeneralDatabase.BeginTrans()) {
                foreach (var server in servers) {
                    UpdateServerStatistic(server.Endpoint);
                }

	            db.DropCollection<DirtyServer>();

                transaction.Commit();
            }
	    }
	}
}