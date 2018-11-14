using System;
using System.Collections.Generic;
using LiteDB;
using Newtonsoft.Json;

namespace Kontur.GameStats.Server.Models {
	public class PlayerStatistic {
		[JsonIgnore]
		[BsonId]
		public string Name { get; set; }

        [BsonIndex]
		public int TotalMatchesPlayed { get; set; }

		public int TotalMatchesWon { get; set; }

		public int UniqueServers { get; set; }

		public int MaximumMatchesPerDay { get; set; }

		public double AverageMatchesPerDay { get; set; }

		public DateTimeOffset LastMatchPlayed { get; set; }

		public double AverageScoreboardPercent { get; set; }

		public string FavoriteGameMode { get; set; }

		public string FavoriteServer { get; set; }

		[BsonIndex]
		public double KillToDeathRatio { get; set; }

		[JsonIgnore]
		public int TotalKills { get; set; }

		[JsonIgnore]
		public int TotalDeaths { get; set; }

		[JsonIgnore]
		public IDictionary<string, int> GameModes { get; set; } = new Dictionary<string, int>();

		[JsonIgnore]
		public IDictionary<string, int> Servers { get; set; } = new Dictionary<string, int>();

		[JsonIgnore]
		public IDictionary<long, int> Timestamps { get; set; } = new Dictionary<long, int>();

		[JsonIgnore]
		public IList<double> ScoreboardPercents { get; set; } = new List<double>();
	}
}
