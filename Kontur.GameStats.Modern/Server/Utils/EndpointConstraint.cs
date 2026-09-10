using Microsoft.AspNetCore.Routing;
using Microsoft.AspNetCore.Http;
using System.Collections.Generic;
using System.Net.Http;
using System.Text.RegularExpressions;

namespace Kontur.GameStats.Server.Utils {
	public class EndpointConstraint : IRouteConstraint {
		// http://stackoverflow.com/questions/106179/regular-expression-to-match-dns-hostname-or-ip-address
		private const string HostnameRegexExpr =
			"^(([a-zA-Z]|[a-zA-Z][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z]|[A-Za-z][A-Za-z0-9\\-]*[A-Za-z0-9])$";
		private const string IpAddressRegexExpr =
			"^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$";
		// http://stackoverflow.com/questions/25450318/how-to-limit-regex-validation-for-port-number-validation-1024-to-65535
		private const string PortNumberRegexExpr =
			"^(102[4-9]|10[3-9]\\d|1[1-9]\\d{2}|[2-9]\\d{3}|[1-5]\\d{4}|6[0-4]\\d{3}|65[0-4]\\d{2}|655[0-2]\\d|6553[0-5])$";

		/// <summary>
		/// Проверяет формат аргумента на соответствие {IPv4-адрес}-{порт} или {имя хоста}-{порт}
		/// </summary>

		public bool Match(HttpContext? httpContext, IRouter? route, string routeKey, RouteValueDictionary values, RouteDirection routeDirection) {
			var endpoint = values["endpoint"].ToString();

			if (!endpoint.Contains("-")) {
				return false;
			}

			var host = endpoint.Split('-')[0];
			var port = endpoint.Split('-')[1];

			return Regex.IsMatch(host, $"{HostnameRegexExpr}|{IpAddressRegexExpr}")
				&& Regex.IsMatch(port, PortNumberRegexExpr);
		}

	}
}
