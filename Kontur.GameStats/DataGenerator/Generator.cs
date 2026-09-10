using System;
using System.Collections.Generic;
using Bogus;
using Kontur.GameStats.Server.Models;

namespace DataGenerator {
	public class Generator {
		private readonly List<string> _gameModes = new List<string> { "DM", "TDM", "RUSH", "ZOMBIE", "ZOMBIE TDM" };

		private GameServer GenerateServers() {
			var serverInfos = new Faker<ServerInformation>()
				.RuleFor(s => s.Name, f => f.Hacker.Adjective())
				.RuleFor(s => s.GameModes, f => new List<string>(f.PickRandom(_gameModes, f.Random.Number(1, 5))).ToArray());

			var servers = new Faker<GameServer>()
				.RuleFor(s => s.Endpoint, f => $"{(f.Random.Bool() ? f.Internet.Ip() : f.Internet.DomainName())}-{f.Random.Number(1024, 65535)}")
				.RuleFor(s => s.ServerInformation, f => serverInfos.Generate());

			return servers.Generate();
		}

		public KeyValuePair<GameServer, IEnumerable<Match>> Generate(int matchCount, int playercount) {
			var players = new Faker<PlayerResult>()
				.RuleFor(p => p.Name, f => f.Name.LastName())
				.RuleFor(p => p.Frags, f => f.Random.Number(0, 50))
				.RuleFor(p => p.Deaths, f => f.Random.Number(0, 50))
				.RuleFor(p => p.Kills, f => f.Random.Number(0, 50));

			var matchInfos = new Faker<MatchInformation>()
				.RuleFor(m => m.GameMode, f => f.PickRandom(_gameModes))
				.RuleFor(m => m.TimeLimit, f => f.Random.Number(5, 60))
				.RuleFor(m => m.FragLimit, f => f.Random.Number(10, 100))
				.RuleFor(m => m.TimeElapsed, f => f.Random.Double(1, 20))
				.RuleFor(m => m.PlayerResults, f => new List<PlayerResult>(players.Generate(playercount)))
				.RuleFor(m => m.Map, f => f.Address.City());

			var server = GenerateServers();

			var matches = new Faker<Match>()
				.RuleFor(m => m.MatchId, f => new MatchId {Endpoint = server.Endpoint, Timestamp = f.Date.Between(new DateTime(2018, 1, 1), new DateTime(2018, 1, 14))})
				.RuleFor(m => m.MatchInformation, _ => matchInfos.Generate());

			return new KeyValuePair<GameServer, IEnumerable<Match>>(server, matches.Generate(matchCount));
		}
	}
}