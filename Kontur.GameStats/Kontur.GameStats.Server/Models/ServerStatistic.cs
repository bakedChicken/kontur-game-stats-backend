using System.Collections.Generic;
using LiteDB;
using Newtonsoft.Json;

namespace Kontur.GameStats.Server.Models {
	public class ServerStatistic {
		public int TotalMatchesPlayed { get; set; }

		public int MaximumMatchesPerDay { get; set; }

		[BsonIndex]
		public double AverageMatchesPerDay { get; set; }

		public int MaximumPopulation { get; set; }

		public double AveragePopulation { get; set; }

		[JsonProperty("top5GameModes")]
		public IEnumerable<string> TopFiveGameModes { get; set; }

		[JsonProperty("top5Maps")]
		public IEnumerable<string> TopFiveMaps { get; set; }
	}
}
