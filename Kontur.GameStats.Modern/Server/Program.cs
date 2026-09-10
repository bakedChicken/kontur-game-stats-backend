using System;
using System.IO;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Routing;
using Microsoft.AspNetCore.HttpOverrides;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

var builder = WebApplication.CreateBuilder(args);

builder.WebHost.UseUrls(new[] {"http://0.0.0.0:8080"});

builder.Services
    .Configure<RouteOptions>(routeOptions =>{
	routeOptions.ConstraintMap.Add("utcdate", typeof(UtcDateConstraint));
	routeOptions.ConstraintMap.Add("endpoint", typeof(EndpointConstraint));
    })
    .AddControllers()
    .AddNewtonsoftJson(options => {
	options.SerializerSettings.Formatting = Formatting.Indented;
	options.SerializerSettings.ContractResolver = new CamelCasePropertyNamesContractResolver();
	options.SerializerSettings.DateFormatString = "yyyy'-'MM'-'dd'T'HH':'mm':'ss'Z'";
    });

builder.Services.AddScoped<PlayerService>();
builder.Services.AddScoped<MatchService>();
builder.Services.AddScoped<ServerService>();
builder.Services.AddScoped<ReportService>();

var app = builder.Build();
app.MapControllers();

Directory.CreateDirectory(Path.Combine(Environment.CurrentDirectory, "db"));
Directory.CreateDirectory(Path.Combine(Environment.CurrentDirectory, "log"));

app.Run();
