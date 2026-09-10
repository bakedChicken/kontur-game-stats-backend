using System.Collections.Generic;

namespace Kontur.GameStats.Server.Utils {
	public static class DictionaryExtension {
		public static void IncrementValue<T>(this IDictionary<T, int> dictionary, T key) {
			int count;
			dictionary.TryGetValue(key, out count);
			dictionary[key] = count + 1;
		}

		public static void DecrementValue<T>(this IDictionary<T, int> dictionary, T key) {
			int count;
			dictionary.TryGetValue(key, out count);
			dictionary[key] = count - 1;
		}
	}
}