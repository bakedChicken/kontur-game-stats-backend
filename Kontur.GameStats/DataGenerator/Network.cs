using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
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


	class Network {
		private async Task<HttpResponseMessage> SendRequest(string url, string data) {
			using (var client = new HttpClient { BaseAddress = new Uri("http://localhost:8080") }) {
				client.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));

				return await client.PutAsync(url, new StringContent(data, Encoding.UTF8, "application/json")).ConfigureAwait(false);
			}
		}

		public async Task Put(IEnumerable<KeyValuePair<GameServer, IEnumerable<Match>>> servers) {
			var putServersResponseTimes = new List<long>();
			var putMatchesResponseTimes = new List<long>();

			var i = 0;

			foreach (var server in servers) {
				var data = Serializer.Serialize(server.Key.ServerInformation);
				var watch = Stopwatch.StartNew();
				var result = await SendRequest($"/servers/{server.Key.Endpoint}/info", data);
				watch.Stop();
				putServersResponseTimes.Add(watch.ElapsedMilliseconds);
				Console.WriteLine($"{++i}. Server {server.Key.Endpoint}: {result.StatusCode} {watch.ElapsedMilliseconds}");

				foreach (var match in server.Value) {
					data = Serializer.Serialize(match.MatchInformation);
					watch = Stopwatch.StartNew();
					
					result = await SendRequest($"/servers/{match.MatchId.Endpoint}/matches/{match.MatchId.Timestamp:yyyy'-'MM'-'dd'T'HH':'mm':'ss'Z'}",
						data);
					watch.Stop();
					putMatchesResponseTimes.Add(watch.ElapsedMilliseconds);
					Console.WriteLine($"{++i}. Match {match.MatchId.Timestamp:yyyy'-'MM'-'dd'T'HH':'mm':'ss'Z'}: {result.StatusCode} {watch.ElapsedMilliseconds}");
				}
			}

			Console.WriteLine($"Servers time for graph: {string.Join(";", putServersResponseTimes)}");
			Console.WriteLine($"Matches time for graph: {string.Join(";", putMatchesResponseTimes)}");
		}
	}
}