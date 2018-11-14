using System.Collections.Generic;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Services {
	public class ServerService {
		public IEnumerable<GameServer> GetAllServers() => 
			DefaultDatabase.Instance.GetCollection<GameServer>().FindAll();

		public GameServer GetServer(string endpoint) => 
			DefaultDatabase.Instance.GetCollection<GameServer>().FindOne(s => s.Endpoint == endpoint);
		
		public ServerStatistic GetServerStatistic(string endpoint) =>
			DefaultDatabase.Instance.GetCollection<GameServer>().FindOne(s => s.Endpoint == endpoint)?.ServerStatistic;

		public void Insert(GameServer server) {
            var db = DefaultDatabase.Instance;
			var servers = db.GetCollection<GameServer>();
			using (var transaction = db.GeneralDatabase.BeginTrans()) {
		        try {
			        servers.Upsert(server.Endpoint, server);

		            transaction.Commit();
		        } catch {
		            transaction.Rollback();
		            throw;
		        }
		    }
		}
	}
}
