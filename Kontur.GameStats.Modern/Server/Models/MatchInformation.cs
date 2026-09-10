using System.Collections.Generic;
using Newtonsoft.Json;

namespace Kontur.GameStats.Server.Models {
	public class MatchInformation {
		public int FragLimit { get; set; }

		public int TimeLimit { get; set; }

		public double TimeElapsed { get; set; }

		public string GameMode { get; set; }

		public string Map { get; set; }

		[JsonProperty("scoreboard")]
		public IList<PlayerResult> PlayerResults { get; set; }
	}
}
