using System;
using System.IO;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Workers;
using LiteDB;

namespace Kontur.GameStats.Server.Utils {
    public class DefaultDatabase {
        private static DefaultDatabase _instance;
        private const int ScheludeUpdateTime = 20;

		private DefaultDatabase()  {
            BsonMapper.Global.TrimWhitespace = false;

            GeneralDatabase.Log.Level = 4;
            GeneralDatabase.Log.Logging += s => Logger.Debug(s, "DB", false);

            ScheduleJob.ScheduleTask<UpdateReportsWorker>(w => w.UpdateBestPlayers(), TimeSpan.FromSeconds(ScheludeUpdateTime)).ConfigureAwait(false);
            ScheduleJob.ScheduleTask<UpdateReportsWorker>(w => w.UpdatePopularServers(), TimeSpan.FromSeconds(ScheludeUpdateTime)).ConfigureAwait(false);
            ScheduleJob.ScheduleTask<UpdateReportsWorker>(w => w.UpdateRecentMatches(), TimeSpan.FromSeconds(ScheludeUpdateTime)).ConfigureAwait(false);
            ScheduleJob.ScheduleTask<UpdateStatisticWorker>(w => w.UpdateDirtyServers(), TimeSpan.FromSeconds(ScheludeUpdateTime)).ConfigureAwait(false);
        }

        public static DefaultDatabase Instance => _instance ?? (_instance = new DefaultDatabase());

        public LiteDatabase GeneralDatabase { get; } = new LiteDatabase(new MemoryStream());

        public LiteDatabase MatchDatabase { get; } = new LiteDatabase("Filename=db/match.db");

        public LiteDatabase ServerDatabase { get; } = new LiteDatabase("Filename=db/server.db");

        public LiteDatabase PlayerDatabase { get; } = new LiteDatabase("Filename=db/player.db");

	    public LiteCollection<T> GetCollection<T>(string collection = "") {
		    LiteDatabase db;

		    if (typeof(T) == typeof(Match)) {
			    db = MatchDatabase;
		    } else if (typeof(T) == typeof(GameServer)) {
			    db = ServerDatabase;
		    } else if (typeof(T) == typeof(PlayerStatistic)) {
			    db = PlayerDatabase;
		    } else {
			    db = GeneralDatabase;
		    }

		    return db.GetCollection<T>(collection.Length > 0 ? collection : typeof(T).Name);
	    }

	    public bool DropCollection<T>(string collection = "") => 
			GeneralDatabase.DropCollection(collection.Length > 0 ? collection : typeof(T).Name);

        public void Dispose() {
            GeneralDatabase?.Dispose();
            MatchDatabase?.Dispose();
            ServerDatabase?.Dispose();
            PlayerDatabase?.Dispose();
        }
    }
}