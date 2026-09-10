using System;
using System.Collections.Generic;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Http;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
        [ApiController]
	[Route("reports")]
	public class ReportController : ControllerBase {
		private readonly ReportService _service;

		public ReportController(ReportService service) {
			_service = service;
		}

		private static int Normalize(int count) => count > 50 ? 50 : (count <= 0 ? 0 : count);

		[HttpGet]
		[Route("recent-matches/{count:int?}")]
		public ActionResult<RecentMatch> RecentMatches(int count = 5) {
			try {
				var matches = _service.GetRecentMatches(Normalize(count));

				return Ok(matches);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::recent_matches()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}

		[HttpGet]
		[Route("best-players/{count:int?}")]
		public ActionResult<IEnumerable<BestPlayer>> BestPlayers(int count = 5) {
			try {
				var players = _service.GetBestPlayers(Normalize(count));

				return Ok(players);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::best_players()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}

		[HttpGet]
		[Route("popular-servers/{count:int?}")]
		public ActionResult<IEnumerable<PopularServer>> PopularServers(int count = 5) {
			try {
				var servers = _service.GetPopularServers(Normalize(count));

				return Ok(servers);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "report_controller::popular_servers()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}
	}
}
