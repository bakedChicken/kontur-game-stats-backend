using System;
using System.Linq;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Workers {
	public class UpdateReportsWorker {
		private const int TableSize = 50;
		
		public void UpdateBestPlayers() {
			var db = DefaultDatabase.Instance;
			var bestPlayers = db.GetCollection<BestPlayer>();

			using (var transaction = db.GeneralDatabase.BeginTrans()) {
				try {
				    var min = bestPlayers.Min(s => s.KillToDeathRatio);

                    var players = db.GetCollection<PlayerStatistic>().Find(p =>
                        p.TotalDeaths > 0 && p.TotalMatchesPlayed >= 10 &&
				        p.KillToDeathRatio >= min
                    );

                    db.DropCollection<BestPlayer>();

                    bestPlayers.Insert(players
                        .OrderByDescending(p => p.KillToDeathRatio)
                        .Take(TableSize)
                        .Select(x => new BestPlayer { Name = x.Name, KillToDeathRatio = x.KillToDeathRatio }));
                    
                    transaction.Commit();
				} catch (Exception e) {
					Logger.Trace(e.Message, "UpdateReportsWorker::BestPlayers");
					transaction.Rollback();
					throw;
				}
			}
		}
		
		public void UpdatePopularServers() {
			var db = DefaultDatabase.Instance;
			var popularServers = db.GetCollection<PopularServer>();

			using (var transaction = db.GeneralDatabase.BeginTrans()) {
				try {
                    db.DropCollection<PopularServer>();
                    popularServers.Insert(db.GetCollection<GameServer>()
                        .Find(p => p.ServerStatistic.AverageMatchesPerDay > 0)
                        .OrderByDescending(p => p.ServerStatistic.AverageMatchesPerDay)
                        .Take(TableSize)
                        .Select(x => new PopularServer {
                            Endpoint = x.Endpoint,
                            Name = x.ServerInformation.Name,
                            AverageMatchesPerDay = x.ServerStatistic.AverageMatchesPerDay
                        })
                    );

					transaction.Commit();
				} catch (Exception ex) {
					Logger.Trace(ex.Message, "UpdateReportsWorker::PopularServers");
					transaction.Rollback();
					throw;
				}
			}
		}
		
		public void UpdateRecentMatches() {
			var db = DefaultDatabase.Instance;
			var popularServers = db.GetCollection<RecentMatch>();

			using (var transaction = db.GeneralDatabase.BeginTrans()) {
				try {
                    db.DropCollection<RecentMatch>();
                    popularServers.Insert(db.GetCollection<Match>().FindAll()
                        .OrderByDescending(m => m.MatchId.Timestamp).Take(TableSize)
						.Select(m => new RecentMatch {
							Endpoint = m.MatchId.Endpoint,
							Timestamp = m.MatchId.Timestamp,
							Results = m.MatchInformation
						})
                    );

					transaction.Commit();
				} catch (Exception ex) {
					Logger.Trace(ex.Message, "UpdateReportsWorker::RecentMatches");
					transaction.Rollback();
					throw;
				}
			}
		}
	}
}