using System.Collections.Generic;
using System.Linq;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Services {
	public class ReportService {
		public IEnumerable<RecentMatch> GetRecentMatches(int count) => 
			GetReport<RecentMatch>(count);

		public IEnumerable<BestPlayer> GetBestPlayers(int count) => 
			GetReport<BestPlayer>(count);

		public IEnumerable<PopularServer> GetPopularServers(int count) => 
			GetReport<PopularServer>(count);

		private static IEnumerable<T> GetReport<T>(int count) {
			return count == 0 ? Enumerable.Empty<T>() : DefaultDatabase.Instance.GetCollection<T>().Find(_ => true, 0, count);
		}
	}
}