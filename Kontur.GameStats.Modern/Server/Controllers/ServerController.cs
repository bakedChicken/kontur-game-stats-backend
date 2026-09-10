using System;
using System.Collections.Generic;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Http;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
        [ApiController]
        [Route("servers")]
	public class ServerController : ControllerBase {
		private readonly ServerService _servers;

		public ServerController(ServerService servers) {
			_servers = servers;
		}

		[HttpGet]
		[Route("info")]
		public ActionResult<IEnumerable<GameServer>> Get() {
			try {
				var servers = _servers.GetAllServers();

				return Ok(servers);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::Get");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif

			}
		}

		[HttpGet]
		[Route("{endpoint:endpoint}/info")]
		public ActionResult<ServerInformation> GetInfo(string endpoint) {
			try {
				var server = _servers.GetServer(endpoint);

				if (server == null) {
					return NotFound();
				}

				return Ok(server.ServerInformation);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::GetInfo");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}

		[HttpPut]
		[Route("{endpoint:endpoint}/info")]
		public ActionResult PutInfo(string endpoint, [FromBody] ServerInformation info) {
			try {
				if (!ModelState.IsValid)
					return BadRequest();

				_servers.Insert(new GameServer {
				    Endpoint = endpoint,
                    ServerInformation = info
				});

				return Ok();
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::PutInfo");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}

		[HttpGet]
		[Route("{endpoint:endpoint}/stats")]
		public ActionResult<ServerStatistic> GetStatistic(string endpoint) {
			try {
				var serverStatistic = _servers.GetServerStatistic(endpoint);

				if (serverStatistic == null) {
					return NotFound();
				}

				return Ok(serverStatistic);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::get_statistic()");
#if DEBUG
				return StatusCode(StatusCodes.Status500InternalServerError, e);
#else
				return StatusCode(StatusCodes.Status500InternalServerError);
#endif
			}
		}
	}

}
