# Payment Service

The Payment service charges the given credit card info (mock) with the given amount and returns a transaction.

## Building locally

    npm install
    node -r ./tracing index.js


## Building docker image

From `paymentservice/`, run:

```
docker build ./
```

## Instrumentation

To enable instrumentation, create a tracing.js file which will be executed before the main application begins execution.

The following npm packages need to be installed for instrumentation:

 - "@opentelemetry/api": v"^1.0.4"
 - "@opentelemetry/sdk-trace-base": v"^1.0.1",
 - "@opentelemetry/sdk-trace-node": v"^1.0.1",
 - "@opentelemetry/sdk-trace-web": "v^1.0.1",
 - "@opentelemetry/semantic-conventions": v"^1.0.1",
 - "@opentelemetry/auto-instrumentations-node": v"^0.27.3",

 - "@opentelemetry/exporter-trace-otlp-grpc": v"^0.27.0",
 - "@opentelemetry/exporter-trace-otlp-http": v"^0.27.0",
 - "@opentelemetry/exporter-trace-otlp-proto": v"^0.27.0",

 - "@opentelemetry/instrumentation": v"^0.27.0",
 - "@opentelemetry/instrumentation-http": v"^0.27.0",

STEP 1: Import opentelemetry packages.
```
    const { WebTracerProvider } = require('@opentelemetry/sdk-trace-web');
    const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
    const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
    const { Resource } = require('@opentelemetry/resources');
    const { BatchSpanProcessor } = require("@opentelemetry/sdk-trace-base");
    const OTLP_HTTP_EXPORTER = require('@opentelemetry/exporter-trace-otlp-proto');
    const { registerInstrumentations } = require('@opentelemetry/instrumentation');
    const OTLP_GRPC_EXPORTER = require('@opentelemetry/exporter-trace-otlp-grpc');
 
```

STEP 2: Initialize Resource.
```
    const otel_resource = new Resource({
        [SemanticResourceAttributes.INSTANCE_ID]: 'INSTANCE_ID',
        [SemanticResourceAttributes.SERVICE_NAME]:'SERVICE_NAME' ,
        [SemanticResourceAttributes.SERVICE_NAMESPACE]:'SERVICE_NAMESPACE',
    }),
```

For semantic conventions, refer [Opentelemetry Node Semantic Conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md#service)

STEP 3: Configure Trace Exporter.
```
    const collectorOptions = {
        url: OTEL_EXPORTER_ENDPOINT, // url is optional and can be omitted - default is http://localhost:55681/v1/traces
        serviceName: "SERVICE_NAME",
        headers{
            key : 'value',
        },// an optional object containing custom headers to be sent with each request will only work with http
        attributes: attributeslist,
    };
    // FOR OTLP HTTP
    const exporter = new OTLP_HTTP_EXPORTER.OTLPTraceExporter(collectorOptions)

    // FOR OTLP GRPC
    const exporter = new OTLP_GRPC_EXPORTER.OTLPTraceExporter(collectorOptions)

```
For information about using OTLP/HTTP or OTLP/gRPC refer [this.](../README.md#When-to-use-OTLP/HTTP-or-OTLP/gRPC)

STEP 4: Register Instrumentations.
```
registerInstrumentations({
    instrumentations: [
        getNodeAutoInstrumentations({
            // load custom configuration for http instrumentation
            '@opentelemetry/instrumentation-http': {},
            '@opentelemetry/instrumentation-grpc': {},
            '@opentelemetry/instrumentation-dns': {},
            '@opentelemetry/instrumentation-express': {}

        }),
    ],
})
```

STEP 5: Configure Trace Provider.
```
    const provider = new WebTracerProvider({
    resource:otel_resource ,
    });

    provider.addSpanProcessor(new BatchSpanProcessor(exporter, {
        // The maximum queue size. After the size is reached spans are dropped.
        maxQueueSize: 1000,
    }));

    provider.register();

```

Now to instrument the application and generate traces, use the following command:
```
    node -r ./tracing index.js

```

For more information on configuring otlp exporter , visit [OTLP PROTOCOL SPECIFICATIONS.](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options)