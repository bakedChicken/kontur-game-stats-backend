using System;
using System.Diagnostics;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace DataGenerator {
	class Network {
		private static readonly HttpClient Client = CreateClient();

		private static HttpClient CreateClient() {
			var client = new HttpClient { BaseAddress = new Uri("http://localhost:8080") };
			client.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
			return client;
		}

		private long _totalSent;
		private long _totalFailed;

		public long TotalSent => Interlocked.Read(ref _totalSent);
		public long TotalFailed => Interlocked.Read(ref _totalFailed);

		public async Task Send(Submission submission, CancellationToken token) {
			var watch = Stopwatch.StartNew();
			try {
				var content = new StringContent(submission.Body, Encoding.UTF8, "application/json");
				var response = await Client.PutAsync(submission.Url, content, token).ConfigureAwait(false);
				watch.Stop();
				Interlocked.Increment(ref _totalSent);
				Console.WriteLine($"{submission.Description}: {(int) response.StatusCode} {watch.ElapsedMilliseconds}ms");
			} catch (OperationCanceledException) when (token.IsCancellationRequested) {
				// Shutting down; the in-flight request was cancelled on purpose.
			} catch (Exception ex) {
				watch.Stop();
				Interlocked.Increment(ref _totalFailed);
				Console.WriteLine($"{submission.Description}: FAILED {ex.GetType().Name}: {ex.Message}");
			}
		}
	}
}
