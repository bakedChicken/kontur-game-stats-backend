using System;
using Microsoft.AspNetCore.Routing;
using Microsoft.AspNetCore.Http;
using System.Collections.Generic;
using System.Globalization;
using System.Net.Http;

namespace Kontur.GameStats.Server.Utils {
	public class UtcDateConstraint : IRouteConstraint {
	    public bool Match(HttpContext? httpContext, IRouter? route, string routeKey, RouteValueDictionary values, RouteDirection routeDirection) {
			var timestamp = values["timestamp"].ToString();

			DateTime date;
			if (!DateTime.TryParse(timestamp, CultureInfo.InvariantCulture, DateTimeStyles.AdjustToUniversal, out date))
				return false;

			return date.Kind == DateTimeKind.Utc;
		}
	}
}
