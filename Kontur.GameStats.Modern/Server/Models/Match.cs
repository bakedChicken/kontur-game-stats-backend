using System;
using LiteDB;
using Newtonsoft.Json;

namespace Kontur.GameStats.Server.Models {
	public struct MatchId {
		public DateTimeOffset Timestamp { get; set; }

		public string Endpoint { get; set; }
	}

	public class Match {
		[BsonId]
		[JsonIgnore]
		public MatchId MatchId { get; set; }

		public MatchInformation MatchInformation { get; set; }
	}
}