// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	//opentelemetry
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"

	"go.opentelemetry.io/otel/propagation"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"

	
	"go.opentelemetry.io/otel/sdk/trace"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/shippingservice/genproto"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)


var log *logrus.Logger
var serviceName string
var serviceNameSpace string

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

func detectResource() (*resource.Resource, error) {
	var instID label.KeyValue
	if host, ok := os.LookupEnv("HOSTNAME"); ok && host != "" {
		instID = semconv.ServiceInstanceIDKey.String(host)
	} else {
		instID = semconv.ServiceInstanceIDKey.String(uuid.New().String())
	}
    hostName:=os.Getenv("MY_POD_NAME")
   hostIp:=os.Getenv("MY_POD_IP")
   resourceType:=os.Getenv("RESOURCE_TYPE")
   return resource.New(
      context.Background(),
      resource.WithAttributes(
         instID,
         semconv.ServiceNameKey.String(serviceName),
         semconv.HostNameKey.String(hostName),
         label.String("service.namespace",serviceNameSpace),
         label.String("ip",hostIp),
         label.String("resource.type",resourceType),
      ),
   )
}
func spanExporter() (exporttrace.SpanExporter, error) {

    var user = os.Getenv("JAEGER_USER")
    var password = os.Getenv("JAEGER_PASSWORD")
	export_type := os.Getenv("EXPORT_TYPE")
	if export_type==""{
		export_type="OTLP"
	}
	if export_type == "JAEGER" {
		log.Info("exporting with JAEGER logger")
		addr1 := os.Getenv("JAEGER_ENDPOINT")
		return jaeger.NewRawExporter(
			jaeger.WithCollectorEndpoint(addr1,jaeger.WithUsername(user),jaeger.WithPassword(password)),
			jaeger.WithProcess(jaeger.Process{
				ServiceName: serviceName,
			}),
		)
	}
	if export_type == "OTLP" {
		log.Info("exporting with OTLP logger")
		addr := os.Getenv("OTLP_ENDPOINT")
		if addr == "" {
			addr = "localhost:4317"
		}
			return otlp.NewExporter(
				context.Background(),
				otlp.WithInsecure(),
				otlp.WithAddress(addr),
			)
		
	}

	log.Info("exporting with STDOUT logger")
	return stdout.NewExporter(
		stdout.WithPrettyPrint(),
		stdout.WithWriter(log.Writer()),
	)
}
func initTracing() {
	if os.Getenv("DISABLE_TRACING") != "" {
		log.Info("tracing disabled")
		return
	}

	res, err := detectResource()
	if err != nil {
		log.WithError(err).Fatal("failed to detect environment resource")
	}

	exp, err := spanExporter()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize Span exporter")
		return
	}

	log.Info("tracing enabled")
	otel.SetTracerProvider(
		trace.NewTracerProvider(
			trace.WithConfig(
				trace.Config{
					DefaultSampler: trace.AlwaysSample(),
					Resource:       res,
				},
			),
			trace.WithSpanProcessor(
				trace.NewBatchSpanProcessor(exp),
			),
		),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func main() {

    if serviceName=os.Getenv("SERVICE_NAME");serviceName==""{
        serviceName = "Shipping-service"
    }
    if serviceNameSpace=os.Getenv("SERVICE_NAMESPACE");serviceNameSpace==""{
        serviceNameSpace = "hipster"
    }
	initTracing()

	var port = os.Getenv("PORT")
	if port == "" {
		port="5551"
	}

	port = fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var srv *grpc.Server

	srv = grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	svc := &server{}
	pb.RegisterShippingServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)
	log.Infof("Shipping Service listening on port %s", port)

	// Register reflection service on gRPC server.
	reflection.Register(srv)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server controls RPC service responses.
type server struct{}

// Check is for health checking.
func (s *server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *server) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

// GetQuote produces a shipping quote (cost) in USD.
func (s *server) GetQuote(ctx context.Context, in *pb.GetQuoteRequest) (*pb.GetQuoteResponse, error) {
	log.Info("[GetQuote] received request")
	defer log.Info("[GetQuote] completed request")

	// 1. Our quote system requires the total number of items to be shipped.
	count := 0
	for _, item := range in.Items {
		count += int(item.Quantity)
	}

	// 2. Generate a quote based on the total number of items to be shipped.
	quote := CreateQuoteFromCount(count)

	// 3. Generate a response.
	return &pb.GetQuoteResponse{
		CostUsd: &pb.Money{
			CurrencyCode: "USD",
			Units:        int64(quote.Dollars),
			Nanos:        int32(quote.Cents * 10000000)},
	}, nil

}

// ShipOrder mocks that the requested items will be shipped.
// It supplies a tracking ID for notional lookup of shipment delivery status.
func (s *server) ShipOrder(ctx context.Context, in *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	log.Info("[ShipOrder] received request")
	defer log.Info("[ShipOrder] completed request")
	// 1. Create a Tracking ID
	baseAddress := fmt.Sprintf("%s, %s, %s", in.Address.StreetAddress, in.Address.City, in.Address.State)
	id := CreateTrackingId(baseAddress)

	// 2. Generate a response.
	return &pb.ShipOrderResponse{
		TrackingId: id,
	}, nil
}
