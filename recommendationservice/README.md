# Recommendation Service

The Recommendation service recommends other products based on what's given in the cart.

## Building locally

    pip3 install -r requirements.txt
    opentelemetry-instrument python3 recommendation_server.py


## Building docker image

From `recommendationservice/`, run:

```
docker build ./
```

## Instrumentation

Fist install the following opentelemetry packages:
  - for Tracing
    - opentelemetry-api v1.9.1
    - opentelemetry-sdk v1.9.1
    - opentelemetry-instrumentation v0.28b1
  - Instrumentation specific
    - opentelemetry.instrumentation.grpc v0.28b1
    - opentelemetry--instrumentation-flask v0.28b1
    - opentelemetry--instrumentation-jinja2 v0.28b1
    - opentelemetry--instrumentation-requests v0.28b1
    - opentelemetry--instrumentation-sqlite3 v0.28b1
    - opentelemetry--instrumentation-urllib v0.28b1
  - Exporter packages
    - opentelemetry-exporter-otlp-proto-http v1.9.1
    - opentelemetry-exporter-otlp-proto-grpc v1.9.1

All the opentelemetry packages are available on the Python Package Index (PyPI). You can install them via pip.


STEP 1 : Import the packages in python application as follows:
```
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter as OTLP_HTTP_EXPORTER
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter as OTLP_GRPC_EXPORTER

```

STEP 2: Create Resource.
```
    attributes={}
    attributes["service.name"]=serviceName
    attributes["service.namespace"]=servicenamespace
    resource = Resource(attributes)

```

STEP 3: Initialize Exporter.
```
        # FOR OTLP HTTP EXPORTER
        otlp_exporter = OTLP_HTTP_EXPORTER(endpoint=otlp_endpoint , headers=json.loads(otlp_headers))
        span_processor = BatchSpanProcessor(otlp_exporter)


        # FOR OTLP GRPC EXPORTER
        otlp_exporter = OTLP_GRPC_EXPORTER(endpoint=otlp_endpoint)
        span_processor = BatchSpanProcessor(otlp_exporter)

```
For information about using OTLP/HTTP or OTLP/gRPC refer [this.](../README.md#When-to-use-OTLP/HTTP-or-OTLP/gRPC)

STEP 4: Initialize the tracer.
```
        trace.set_tracer_provider(TracerProvider(resource=resource))
        trace.get_tracer_provider().add_span_processor(span_processor)

```

Now to run the instrumented application, use the following command to generate traces:
```
    opentelemetry-instrument python3 application.py

```

For more information, visit [Opentelemetry Python Documentation.](https://opentelemetry-python.readthedocs.io/en/latest/)