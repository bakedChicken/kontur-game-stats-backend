using System;
using System.IO;
using System.Net.Http.Headers;
using System.Reflection;
using System.Web.Http;
using System.Web.Http.Routing;
using Kontur.GameStats.Server.Services;
using Kontur.GameStats.Server.Utils;
using Microsoft.Owin.Hosting;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;
using Owin;
using SimpleInjector;
using SimpleInjector.Integration.WebApi;

namespace Kontur.GameStats.Server {
	internal class Startup {
		public void Configuration(IAppBuilder app) {
			var config = new HttpConfiguration();

			var constraints = new DefaultInlineConstraintResolver();

			// https://blog.zwezdin.com/2014/webapi-attribute-routing/
			constraints.ConstraintMap.Add("endpoint", typeof(EndpointConstraint));
			constraints.ConstraintMap.Add("utcdate", typeof(UtcDateContstraint));
			config.MapHttpAttributeRoutes(constraints);

			config.Formatters.Remove(config.Formatters.XmlFormatter);
            config.Formatters.JsonFormatter.SupportedMediaTypes.Add(new MediaTypeHeaderValue("text/plain"));
			config.Formatters.JsonFormatter.SerializerSettings.Formatting = Formatting.Indented;
			config.Formatters.JsonFormatter.SerializerSettings.ContractResolver = new CamelCasePropertyNamesContractResolver();
			config.Formatters.JsonFormatter.SerializerSettings.DateFormatString = "yyyy'-'MM'-'dd'T'HH':'mm':'ss'Z'";
			
			var container = new Container();
			container.Options.DefaultScopedLifestyle = new WebApiRequestLifestyle();
			
			container.Register<PlayerService>(Lifestyle.Scoped);
			container.Register<MatchService>(Lifestyle.Scoped);
			container.Register<ServerService>(Lifestyle.Scoped);
			container.Register<ReportService>(Lifestyle.Scoped);
			container.Verify();

			config.DependencyResolver = new SimpleInjectorWebApiDependencyResolver(container);

            app.UseWebApi(config);
		}
    }

    internal static class Program {
		public static void Main(string[] args) {
			string prefix;
			if (args.Length != 2 || args[0] != "--prefix") {
				prefix = "http://+:8080";
			} else {
				prefix = args[1];
			}

			Directory.CreateDirectory(Environment.CurrentDirectory + "/db");
			Directory.CreateDirectory(Environment.CurrentDirectory + "/log");

		    try {
		        using (WebApp.Start<Startup>(prefix)) {
		            Logger.Info($"Running with prefix {prefix}", "Main");
                    Console.ReadKey(true);
                }
		    } catch (TargetInvocationException) {
		        Logger.Info($"You should run 'netsh http add urlacl url={prefix} user=%username%' before continue", "Main");
                Console.ReadKey(true);
            }
		}
	}
}
