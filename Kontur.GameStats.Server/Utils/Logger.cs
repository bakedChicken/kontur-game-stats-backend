using System;

namespace Kontur.GameStats.Server.Utils {
	public static class Logger {
		public static void Trace(string message, string module, bool time = true) {
			System.Diagnostics.Trace.WriteLine($"{(time ? DateTimeOffset.Now + ". " : "")}ERROR in {module}: {message}");
		}

		public static void Debug(string message, string module, bool time = true) {
			System.Diagnostics.Debug.WriteLine($"{(time ? DateTimeOffset.Now + ". " : "")}DEBUG in {module}: {message}");
		}

		public static void Info(string message, string module, bool time = true) {
			System.Diagnostics.Trace.WriteLine($"{(time ? DateTimeOffset.Now + ". " : "")}Info in {module}: {message}");
		}
	}
}
