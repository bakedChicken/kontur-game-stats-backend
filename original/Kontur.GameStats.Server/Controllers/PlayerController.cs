using System;
using System.Web.Http;
using System.Web.Http.Description;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
	[RoutePrefix("players")]
	public class PlayerController : ApiController {
		private readonly PlayerService _players;

		public PlayerController(PlayerService players) {
			_players = players;
		}

		[HttpGet]
		[Route("{name}/stats")]
		[ResponseType(typeof(PlayerStatistic))]
		public IHttpActionResult GetPlayerStatistic(string name) {
			try {
				var playerStatistic = _players.GetPlayerStatistic(name);

				if (playerStatistic == null) {
					return NotFound();
				}

				return Ok(playerStatistic);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "player_controller::get_statistic()");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}
	}
}
