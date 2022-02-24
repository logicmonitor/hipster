module github.com/GoogleCloudPlatform/microservices-demo/src/frontend

go 1.15

require (
	cloud.google.com/go v0.74.0
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/stackdriver v0.5.0
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/newrelic/newrelic-telemetry-sdk-go v0.5.1
	github.com/newrelic/opentelemetry-exporter-go v0.14.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/uber/jaeger-client-go v2.21.1+incompatible // indirect
	go.opencensus.io v0.22.5
	go.opentelemetry.io/contrib/detectors/gcp v0.15.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.15.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/stdout v0.15.0
	go.opentelemetry.io/otel/exporters/otlp v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11
	google.golang.org/grpc v1.34.0
)
