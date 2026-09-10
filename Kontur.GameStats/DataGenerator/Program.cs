using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Kontur.GameStats.Server.Models;

namespace DataGenerator {
	internal class Program {
		public static void Main(string[] args) {
			if (args.Length == 4) {
				var serverCount = Convert.ToInt32(args[1]);
				var matchCount = Convert.ToInt32(args[2]);
				var playerCount = Convert.ToInt32(args[3]);

				var pairs = new List<KeyValuePair<GameServer, IEnumerable<Match>>>();
				var generator = new Generator();

				Console.WriteLine("Generating...");

				for (var i = 0; i < serverCount; i++) {
					pairs.Add(generator.Generate(matchCount, playerCount));
				}

				switch (args[0]) {
					case "--gen":
						Console.WriteLine(Serializer.Serialize(pairs));
						break;
					case "--put":
						Task.Run(() => new Network().Put(pairs)).GetAwaiter().GetResult();
						break;
					default:
						Console.WriteLine("args: --gen | --put [count of servers] [count of matches on server] [count of players in match]");
						break;
				}
			} else {
				Console.WriteLine("Provide 3 args: --gen | --put [count of servers] [count of matches on server] [count of players in match]");
			}
		}
	}
}