using System;
using System.Collections.Generic;
using System.Web.Http;
using System.Web.Http.Description;
using Kontur.GameStats.Server.Models;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;

namespace Kontur.GameStats.Server.Controllers {
    [RoutePrefix("servers")]
	public class ServerController : ApiController {
		private readonly ServerService _servers;

		public ServerController(ServerService servers) {
			_servers = servers;
		}

		[HttpGet]
		[Route("info")]
		[ResponseType(typeof(List<GameServer>))]
		public IHttpActionResult Get() {
			try {
				var servers = _servers.GetAllServers();

				return Ok(servers);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::Get");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif

			}
		}

		[HttpGet]
		[Route("{endpoint:endpoint}/info")]
		[ResponseType(typeof(ServerInformation))]
		public IHttpActionResult GetInfo(string endpoint) {
			try {
				var server = _servers.GetServer(endpoint);

				if (server == null) {
					return NotFound();
				}

				return Ok(server.ServerInformation);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::GetInfo");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}

		[HttpPut]
		[Route("{endpoint:endpoint}/info")]
		public IHttpActionResult PutInfo(string endpoint, [FromBody] ServerInformation info) {
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
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}

		[HttpGet]
		[Route("{endpoint:endpoint}/stats")]
		[ResponseType(typeof(ServerStatistic))]
		public IHttpActionResult GetStatistic(string endpoint) {
			try {
				var serverStatistic = _servers.GetServerStatistic(endpoint);

				if (serverStatistic == null) {
					return NotFound();
				}

				return Ok(serverStatistic);
			} catch (Exception e) {
				Logger.Trace(e.ToString(), "server_controller::get_statistic()");
#if DEBUG
				return InternalServerError(e);
#else
				return InternalServerError();
#endif
			}
		}
	}

}
