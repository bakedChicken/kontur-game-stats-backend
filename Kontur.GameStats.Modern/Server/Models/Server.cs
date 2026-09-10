using LiteDB;
using Newtonsoft.Json;

namespace Kontur.GameStats.Server.Models {
	public class GameServer {
		[BsonId]
		public string Endpoint { get; set; }

		[JsonProperty("info")]
		public ServerInformation ServerInformation { get; set; } = new ServerInformation();

		[JsonIgnore]
		public ServerStatistic ServerStatistic { get; set; } = new ServerStatistic();
	}
}