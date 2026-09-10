using System;

namespace Kontur.GameStats.Server.Models {
	public class RecentMatch {
		public string Endpoint { get; set; }
		public DateTimeOffset Timestamp { get; set; }
		public MatchInformation Results { get; set; }
	}
}