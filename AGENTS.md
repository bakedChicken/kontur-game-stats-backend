# Repository Guidelines

## Project Structure & Module Organization

- `revisited/` is the active Go rewrite. It contains the `revisited` module (`go.mod`) and the application entry point in `main.go`.
- `original/` preserves the legacy C#/.NET application for reference. Its server project lives in `original/Kontur.GameStats.Server/`; controllers, services, models, workers, and utilities are organized in matching subdirectories. `original/DataGenerator/` is a separate supporting executable.
- `flake.nix` and `flake.lock` define reproducible Nix/devenv development shells.

## Build, Test, and Development Commands

Use the Go shell when working on the rewrite:

```bash
nix develop .#revisited
```

Run `nix fmt` after changing Nix files. For the legacy code, enter `nix develop .#original` and build `original/Kontur.GameStats.Server.sln` with the available .NET tooling. Do not make legacy changes unless the task requires them.

## Testing Guidelines

There is currently no committed test suite or coverage threshold. Add Go tests with every behavior change, place them beside the implementation, and run `go test ./...` before submitting. Test externally visible behavior and error paths; avoid tests that depend on real network services or local machine state.

## Development stages

1. The most important part is to restore the original requirements based on the original .NET project. The project has some background workers and probably some low-level technical details, like using the LiteDB, which are not relevant to the requirements. We just need to collect the original high-level requirements for the service in form of OpenAPI document with model and endpoint definitions and descriptions. The OpenAPI file should pass the `swagger validate openapi.yaml` correctness validation. The project also has postman files which should also give a hint of how the project was used and tested.  When we have a full description of the project in this document, in `Original Requirements` section, only then we can proceed to other stages. If Original Requirements section is empty, it means we're still on the stage 1 and we should figure the requirements out, so please do that.
2. We would need to recreate the data generator and run it against the original .NET backend service. It should be implemented as a cmd, and should be ran as a subcommand of the root Go project.
3. Having the data generator recreated and having all the specification available, would be nice for me to figure out what to do next.

## Original Requirements

Kontur GameStats is an unauthenticated HTTP JSON service for collecting final
scoreboards from game servers and exposing the resulting statistics. Its public
contract is documented in [`openapi.yaml`](openapi.yaml); JSON uses camelCase
property names and timestamps are UTC.

### Feature: server registration and discovery

Servers are identified by `{host-or-ipv4}-{port}`, where the port is in the
range 1024–65535. A client registers the endpoint and its display metadata via
`PUT /servers/{endpoint}/info`. The operation creates a previously unknown
server and replaces the metadata of an existing one. The registration must
happen before the server can accept match submissions.

`GET /servers/info` lists all registrations, `GET /servers/{endpoint}/info`
returns one server's metadata, and `GET /servers/{endpoint}/stats` returns its
derived aggregate statistics. On an initial empty service, the list is empty
and endpoint-specific reads return 404.

### Feature: match ingestion and statistics

A registered server submits each final scoreboard through
`PUT /servers/{endpoint}/matches/{timestamp}`. The timestamp is a UTC match
identifier scoped to the server. The payload records the map, game mode,
configured frag/time limits, elapsed time, and non-empty ordered scoreboard.
The first result is the winner; each result carries the player's name, frags,
kills, and deaths. `GET` on that same path retrieves the submitted payload.

The service derives a player's appearances, wins, distinct servers, daily
activity, latest match, average placement, favourite server and game mode, and
aggregate kill/death ratio. Server statistics include match/day counts,
population, and up to five most frequent maps and game modes. Server statistics
and reports are asynchronously refreshed on a 20-second cadence, so a client
may need to wait after submitting a match before observing it through a
statistics or report read.

### Feature: leaderboards and recent activity

`GET /reports/recent-matches[/{count}]` exposes matches newest first.
`GET /reports/best-players[/{count}]` ranks players with at least ten matches
and one death by aggregate kill/death ratio. `GET
/reports/popular-servers[/{count}]` ranks servers by average matches per active
UTC day. Each report defaults to five entries, never returns more than 50, and
returns an empty array for zero or negative requested counts or where no
matching data has yet been derived.

### Canonical client flow

1. Read `GET /servers/info`; initially it is `[]`.
2. Register one or more servers with `PUT /servers/{endpoint}/info`, then read
   their info or the server list to confirm the metadata.
3. Submit completed match scoreboards to each registered server using their
   endpoint and UTC timestamp.
4. Retrieve an individual submitted match when needed, then read
   `GET /players/{name}/stats`, `GET /servers/{endpoint}/stats`, and the report
   endpoints after the derived views have caught up.

The OpenAPI document is the authoritative endpoint, payload, response-model,
field-description, and error-response specification for these features.
