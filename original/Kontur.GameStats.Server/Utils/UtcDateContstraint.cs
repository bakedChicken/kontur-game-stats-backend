using System;
using System.Collections.Generic;
using System.Globalization;
using System.Net.Http;
using System.Web.Http.Routing;

namespace Kontur.GameStats.Server.Utils {
	public class UtcDateContstraint : IHttpRouteConstraint {
		public bool Match(HttpRequestMessage request, IHttpRoute route, string parameterName, IDictionary<string, object> values,
			HttpRouteDirection routeDirection) {
			var timestamp = values["timestamp"].ToString();

			DateTime date;
			if (!DateTime.TryParse(timestamp, CultureInfo.InvariantCulture, DateTimeStyles.AdjustToUniversal, out date))
				return false;

			return date.Kind == DateTimeKind.Utc;
		}
	}
}
