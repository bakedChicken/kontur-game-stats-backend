using System;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;
using Kontur.GameStats.Server.Workers;

namespace Kontur.GameStats.Server.Services {
	public class MatchService {
		public MatchInformation GetMatchInformation(string endpoint, DateTimeOffset timestamp) => 
			DefaultDatabase.Instance.GetCollection<Match>()
			.FindOne(m => m.MatchId.Endpoint == endpoint && m.MatchId.Timestamp == timestamp)?
			.MatchInformation;

		public bool Insert(string endpoint, DateTimeOffset timestamp, MatchInformation info) {
			var db = DefaultDatabase.Instance;
		    var dirtyServers = db.GetCollection<DirtyServer>();

            if (!db.GetCollection<GameServer>().Exists(s => s.Endpoint == endpoint))
				return false;

			using (var transaction = db.MatchDatabase.BeginTrans()) {
				try {
					ScheduleJob.BackgroundTask<UpdateStatisticWorker>(w => w.UpdateUserStatistic(info, endpoint, timestamp)).ConfigureAwait(false);

                    if (!db.GetCollection<Match>().Exists(m => m.MatchId.Timestamp == timestamp && m.MatchId.Endpoint == endpoint))
						db.GetCollection<Match>().Insert(new Match {
							MatchId = new MatchId {
								Timestamp = timestamp,
								Endpoint = endpoint
							},
							MatchInformation = info
						});

					transaction.Commit();
				} catch {
					transaction.Rollback();
					throw;
				}
			}

            if (!dirtyServers.Exists(s => s.Endpoint == endpoint))
                dirtyServers.Insert(new DirtyServer {Endpoint = endpoint});

            return true;
		}
	}
}