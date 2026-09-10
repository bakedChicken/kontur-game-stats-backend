using System;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Http;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
	[ApiController]
	[Route("servers")]
	public class MatchController : ControllerBase {
		private readonly MatchService _matches;

		public MatchController(MatchService matches) {
			_matches = matches;
		}

		[HttpGet]
		[Route("{endpoint:endpoint}/matches/{timestamp:utcdate}")]
		public ActionResult<MatchInformation> GetMatchInfo(string endpoint, DateTimeOffset timestamp) {
			try {
				var match = _matches.GetMatchInformation(endpoint, timestamp);

				if (match == null) {
					return NotFound();
				}

				return Ok(match);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "match_controller::get_match_info()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}

		[HttpPut]
		[Route("{endpoint:endpoint}/matches/{timestamp:utcdate}")]
		public ActionResult PutMatchInfo(string endpoint, DateTimeOffset timestamp, [FromBody] MatchInformation info) {
			try {
				if (!ModelState.IsValid)
					return BadRequest();

				if (!_matches.Insert(endpoint, timestamp, info))
					return BadRequest();

				return Ok();
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "match_controller::put_match_info()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}
	}
}
