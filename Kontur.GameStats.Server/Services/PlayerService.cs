using System;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Services {
    public class PlayerService {
		public PlayerStatistic GetPlayerStatistic(string name) => 
			DefaultDatabase.Instance.GetCollection<PlayerStatistic>()
			.FindOne(p => p.Name.Equals(name, StringComparison.OrdinalIgnoreCase));
    }
}