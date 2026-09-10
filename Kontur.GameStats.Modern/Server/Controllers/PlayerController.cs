using System;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Http;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
        [ApiController]
	[Route("players")]
	public class PlayerController : ControllerBase {
		private readonly PlayerService _players;

		public PlayerController(PlayerService players) {
			_players = players;
		}

		[HttpGet]
		[Route("{name}/stats")]
		public ActionResult<PlayerStatistic> GetPlayerStatistic(string name) {
			try {
				var playerStatistic = _players.GetPlayerStatistic(name);

				if (playerStatistic == null) {
					return NotFound();
				}

				return Ok(playerStatistic);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "player_controller::get_statistic()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}
	}
}
