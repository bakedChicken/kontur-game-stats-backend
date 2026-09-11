using System;
using System.Collections.Generic;
using System.Linq;
using Bogus;
using Kontur.GameStats.Server.Models;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

namespace DataGenerator {
	internal static class Serializer {
		private static readonly JsonSerializerSettings SerializeSettings =
			new JsonSerializerSettings { ContractResolver = new CamelCasePropertyNamesContractResolver(), Formatting = Formatting.Indented };

		public static string Serialize(object obj) {
			return JsonConvert.SerializeObject(obj, Formatting.None, SerializeSettings);
		}
	}

	public sealed class Submission {
		public string Url { get; }
		public string Body { get; }
		public string Description { get; }

		public Submission(string url, string body, string description) {
			Url = url;
			Body = body;
			Description = description;
		}
	}

	public class Generator {
		private const double NewServerProbability = 0.05;
		private const int MinPlayersPerMatch = 1;
		private const int MaxPlayersPerMatch = 32;

		// Shared across every match for the life of a run, not regenerated
		// per match: player statistics (in particular best-players, which
		// requires at least ten matches for one player) only accumulate if
		// the same player name recurs across many different matches. A pool
		// comfortably larger than MaxPlayersPerMatch, so PickRandom always
		// has enough distinct names for one scoreboard, but small enough
		// that names repeat often across a run's worth of matches.
		private const int PlayerPoolSize = 300;

		private static readonly DateTime WindowStart = new DateTime(2018, 1, 1, 0, 0, 0, DateTimeKind.Utc);
		private static readonly DateTime WindowEnd = new DateTime(2018, 1, 14, 0, 0, 0, DateTimeKind.Utc);

		private readonly List<string> _gameModes = new List<string> { "DM", "TDM", "RUSH", "ZOMBIE", "ZOMBIE TDM" };

		private string GenerateEndpoint(Faker faker) {
			return $"{(faker.Random.Bool() ? faker.Internet.Ip() : faker.Internet.DomainName())}-{faker.Random.Number(1024, 65535)}";
		}

		private ServerInformation GenerateServerInformation() {
			return new Faker<ServerInformation>()
				.RuleFor(s => s.Name, f => f.Hacker.Adjective())
				.RuleFor(s => s.GameModes, f => new List<string>(f.PickRandom(_gameModes, f.Random.Number(1, 5))).ToArray())
				.Generate();
		}

		// The pool player names are drawn from for a whole run. First+last
		// name combinations give a space large enough that PlayerPoolSize
		// distinct draws don't take long to find.
		private List<string> GeneratePlayerNamePool(Faker faker, int size) {
			var names = new HashSet<string>();
			while (names.Count < size) {
				names.Add(faker.Name.FullName());
			}
			return new List<string>(names);
		}

		// A match's scoreboard can't contain the same player twice, so names
		// are drawn without replacement from the shared pool.
		private MatchInformation GenerateMatchInformation(Faker faker, List<string> playerNamePool, int playerCount) {
			var players = new Faker<PlayerResult>()
				.RuleFor(p => p.Frags, f => f.Random.Number(0, 50))
				.RuleFor(p => p.Deaths, f => f.Random.Number(0, 50))
				.RuleFor(p => p.Kills, f => f.Random.Number(0, 50));

			var results = faker.PickRandom(playerNamePool, playerCount)
				.Select(name => {
					var result = players.Generate();
					result.Name = name;
					return result;
				})
				.ToList();

			return new Faker<MatchInformation>()
				.RuleFor(m => m.GameMode, f => f.PickRandom(_gameModes))
				.RuleFor(m => m.TimeLimit, f => f.Random.Number(5, 60))
				.RuleFor(m => m.FragLimit, f => f.Random.Number(10, 100))
				.RuleFor(m => m.TimeElapsed, f => f.Random.Double(1, 20))
				.RuleFor(m => m.PlayerResults, _ => results)
				.RuleFor(m => m.Map, f => f.Address.City())
				.Generate();
		}

		private DateTime GenerateTimestamp(Faker faker, HashSet<DateTime> usedForEndpoint) {
			DateTime candidate;
			do {
				var raw = faker.Date.Between(WindowStart, WindowEnd);
				candidate = new DateTime(raw.Year, raw.Month, raw.Day, raw.Hour, raw.Minute, raw.Second, DateTimeKind.Utc);
			} while (!usedForEndpoint.Add(candidate));
			return candidate;
		}

		public IEnumerable<Submission> GenerateForever() {
			var faker = new Faker();
			var endpoints = new List<string>();
			var usedTimestamps = new Dictionary<string, HashSet<DateTime>>();
			var playerNamePool = GeneratePlayerNamePool(faker, PlayerPoolSize);

			while (true) {
				if (endpoints.Count == 0 || faker.Random.Double() < NewServerProbability) {
					var endpoint = GenerateEndpoint(faker);
					endpoints.Add(endpoint);
					usedTimestamps[endpoint] = new HashSet<DateTime>();

					var info = GenerateServerInformation();
					yield return new Submission($"/servers/{endpoint}/info", Serializer.Serialize(info), $"Server {endpoint}");
					continue;
				}

				var target = faker.PickRandom(endpoints);
				var playerCount = faker.Random.Number(MinPlayersPerMatch, MaxPlayersPerMatch);
				var timestamp = GenerateTimestamp(faker, usedTimestamps[target]);
				var formattedTimestamp = timestamp.ToString("yyyy'-'MM'-'dd'T'HH':'mm':'ss'Z'");
				var matchInfo = GenerateMatchInformation(faker, playerNamePool, playerCount);

				yield return new Submission(
					$"/servers/{target}/matches/{formattedTimestamp}",
					Serializer.Serialize(matchInfo),
					$"Match {target} {formattedTimestamp} ({playerCount} players)");
			}
		}
	}
}
