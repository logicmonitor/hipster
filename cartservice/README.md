# Cart Service

Stores the items in the user's shopping cart in Redis and retrieves it.                 

## Building locally
To build the cartservice dll , use **dotnet build** command.

## Building docker image

From `cartservice/`, run:

```
docker build ./
```

## Instrumentation
For manually instrumenting dotnet application, first add the below nuget packages.
  - OpenTelemetry
  - OpenTelemetry.Api
currently for cart service we have used opentelemetry version 1.2.0-beta2.1

Now add the instrumentation specific packages.Below is a list of few packages which can be used:
  - OpenTelemetry.Instrumentation.Http
  - OpenTelemetry.Instrumentation.AspNetCore
  - OpenTelemetry.Instrumentation.StackExchangeRedis

Now add the exporter nuget packages:
  - OpenTelemetry.Exporter.OpenTelemetryProtocol
  - OpenTelemetry.Exporter.Jaeger

To instrument a webapp application, the following configurations are to be added in Startup.cs file
In ConfigureServices method :

```

 public void ConfigureServices(IServiceCollection services) {
            services.AddOpenTelemetryTracing(builder => ConfigureOpenTelemetry(builder));

    }

```
Now create a ConfigureOpentelemetry method as shown below :
```

    private static void ConfigureOpenTelemetry(TracerProviderBuilder builder){
            // Add instrumentations
            builder.AddAspNetCoreInstrumentation().AddHttpClientInstrumentation();         

            // Set resource
            builder.SetResourceBuilder(ResourceBuilder.CreateDefault().AddService(string serviceName,[string serviceNamespace=null], [string serviceVersion=null], [bool autoGenerateServiceInstanceId=true],[string serviceInstanceId=null]").AddAttributes(attributeList));

            // Configure exporter
            builder.AddOtlpExporter(otlpOptions =>
                {
                    otlpOptions.Endpoint = new Uri(otlpEndpoint);
                    if (otlp_format == "HTTP")
                    {
                        otlpOptions.Protocol = OpenTelemetry.Exporter.OtlpExportProtocol.HttpProtobuf;
                    }
                    else if(otlp_format == "GRPC")
                    {
                        otlpOptions.Protocol = OpenTelemetry.Exporter.OtlpExportProtocol.Grpc;
                    }
                });

    }
```

For information about using OTLP/HTTP or OTLP/gRPC refer [this.](../README.md#When-to-use-OTLP/HTTP-or-OTLP/gRPC)

For Semantic conventions, visit [Opentelemetry Semantic Conventions](https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/resource/semantic_conventions#service)

If your application is targeting .NET Core 3.1, and you are using an insecure (HTTP) endpoint, the following switch must be set before adding OtlpExporter.
```
    AppContext.SetSwitch("System.Net.Http.SocketsHttpHandler.Http2UnencryptedSupport",
    true);
```


Now to configure the OTLP exporter, the following attributes can be set:
  - OTEL_EXPORTER_OTLP_ENDPOINT
  - OTEL_EXPORTER_OTLP_HEADERS
  - OTEL_EXPORTER_OTLP_PROTOCOL	 // grpc or http/protobuf
  - OTEL_RESOURCE_ATTRIBUTES
  - OTEL_SERVICE_NAME

To learn more about .Net Opentelemetry manual instrumentation visit [Opentelemetry Manual Instrumentation](https://github.com/open-telemetry/opentelemetry-dotnet/tree/main/src/OpenTelemetry#readme)

