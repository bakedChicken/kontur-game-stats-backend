using System;
using System.Collections.Generic;
using System.Web.Http;
using System.Web.Http.Description;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
	[RoutePrefix("reports")]
	public class ReportController : ApiController {
		private readonly ReportService _service;

		public ReportController(ReportService service) {
			_service = service;
		}

		private static int Normalize(int count) => count > 50 ? 50 : (count <= 0 ? 0 : count);

		[HttpGet]
		[Route("recent-matches/{count:int?}")]
		[ResponseType(typeof(IEnumerable<RecentMatch>))]
		public IHttpActionResult RecentMatches(int count = 5) {
			try {
				var matches = _service.GetRecentMatches(Normalize(count));

				return Ok(matches);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::recent_matches()");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}

		[HttpGet]
		[Route("best-players/{count:int?}")]
		[ResponseType(typeof(IEnumerable<BestPlayer>))]
		public IHttpActionResult BestPlayers(int count = 5) {
			try {
				var players = _service.GetBestPlayers(Normalize(count));

				return Ok(players);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::best_players()");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}

		[HttpGet]
		[Route("popular-servers/{count:int?}")]
		[ResponseType(typeof(IEnumerable<PopularServer>))]
		public IHttpActionResult PopularServers(int count = 5) {
			try {
				var servers = _service.GetPopularServers(Normalize(count));

				return Ok(servers);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::popular_servers()");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}
	}
}
