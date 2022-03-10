# Shipping Service

The Shipping service provides price quote, tracking IDs, and the impression of order fulfillment & shipping processes.

## Local
To run the shipping , use **go run main.go** command.

## Building docker image

From `checkoutservice/`, run:

```
docker build ./

```
## Instrumentation

For instrumentating go applications, first install the following golang instrumentation libraries:
  - go.opentelemetry.io/otel v1.4.1
  -	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.4.1
  -	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.4.1
  -	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.4.1
  -	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.4.1

STEP 1: Create Resource

```

import (
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)
func detectResource() (*resource.Resource, error) {

    return resource.New(
      context.Background(),
      resource.WithAttributes(
         semconv.ServiceNameKey.String(service_name),
         attribute.String("service.namespace",service_nameSpace),
      ),
    )
    
}
```
STEP 2: Intialise the span exporter.

```
import(
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"


)

func spanExporter() (sdktrace.SpanExporter, error) {
    // for exporting traces in OTLP HTTP/PROTOBUF format
    return otlptracehttp.New(
        context.Background(),
        otlptracehttp.WithInsecure(),
        otlptracehttp.WithEndpoint(OTLPENDPOINTHOST),
        otlptracehttp.WithURLPath(OTLPENDPOINTPATH),
        otlptracehttp.WithHeaders(OTLPHEADERS),
        )
    //OR
    // for exporting traces in OTLP GRPC format
	 return otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(OTLPENDPOINTHOST),
		)
}

```

For information about using OTLP/HTTP or OTLP/gRPC refer [this.](../README.md#When-to-use-OTLP/HTTP-or-OTLP/gRPC)
For more information about OTLP Configurations, visit [Opentelemetry OTLP Exporter GO](https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/otlp/otlptrace/README.md)

STEP 3: Configure the TraceProvider. 
```
import(
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel"
)

    func initTracing() {
    res, err := detectResource()
	if err != nil {
		log.WithError(err).Fatal("failed to detect environment resource")
	}

	exp, err := spanExporter()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize Span exporter")
		return
	}

	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	),
	)
	otel.SetTextMapPropagator(propagation.TraceContext{})
    }

```
To configure the exporter , the following environment variables can aslo be set:
  - OTEL_EXPORTER_OTLP_ENDPOINT
  - OTEL_EXPORTER_OTLP_HEADERS
  - OTEL_RESOURCE_ATTRIBUTES
  - OTEL_SERVICE_NAME

Now run the application to generate traces using **go run** command. 