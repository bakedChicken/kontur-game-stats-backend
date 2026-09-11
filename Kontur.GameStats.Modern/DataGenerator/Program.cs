using System;
using System.Threading;
using System.Threading.Tasks;

namespace DataGenerator {
	internal class Program {
		private const int DefaultParallelism = 8;

		public static async Task Main(string[] args) {
			if (args.Length == 1 && (args[0] == "--help" || args[0] == "-h")) {
				Console.WriteLine("Usage: DataGenerator [parallelism]");
				Console.WriteLine("Continuously registers random servers and submits random matches");
				Console.WriteLine("against http://localhost:8080 until stopped (Ctrl+C).");
				Console.WriteLine($"parallelism: number of concurrent workers (default {DefaultParallelism})");
				return;
			}

			var parallelism = args.Length == 1 ? int.Parse(args[0]) : DefaultParallelism;

			using var cts = new CancellationTokenSource();
			Console.CancelKeyPress += (_, e) => {
				e.Cancel = true;
				cts.Cancel();
			};

			Console.WriteLine($"Shooting at the server with {parallelism} concurrent worker(s). Press Ctrl+C to stop.");

			var network = new Network();
			var workers = new Task[parallelism];
			for (var i = 0; i < parallelism; i++) {
				workers[i] = RunWorker(network, cts.Token);
			}

			await Task.WhenAll(workers);

			Console.WriteLine();
			Console.WriteLine($"Stopped. {network.TotalSent} sent, {network.TotalFailed} failed.");
		}

		// Each worker owns its own Generator: Faker wraps a plain Random,
		// which isn't safe to drive from multiple threads at once, so sharing
		// one generator across workers would corrupt its internal state
		// rather than just produce correlated randomness.
		private static async Task RunWorker(Network network, CancellationToken token) {
			var generator = new Generator();

			foreach (var submission in generator.GenerateForever()) {
				if (token.IsCancellationRequested) {
					break;
				}

				await network.Send(submission, token);
			}
		}
	}
}
