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
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
    "net/url"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


var (
	cat          pb.ListProductsResponse
	catalogMutex *sync.Mutex
	log          *logrus.Logger
	extraLatency time.Duration

	port = os.Getenv("PORT")
	reloadCatalog bool
	serviceName string
  serviceNameSpace string
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	if port == "" {
		port="4000"
	}

	log.Out = os.Stdout
	catalogMutex = &sync.Mutex{}
	err := readCatalogFile(&cat)
	if err != nil {
		log.Warnf("could not parse product catalog")
	}
}

func detectResource() (*resource.Resource, error) {
	var instID attribute.KeyValue
	if host, ok := os.LookupEnv("HOSTNAME"); ok && host != "" {
		instID = semconv.ServiceInstanceIDKey.String(host)
	} else {
		instID = semconv.ServiceInstanceIDKey.String(uuid.New().String())
	}
   return resource.New(
      context.Background(),
      resource.WithAttributes(
         instID,
         semconv.ServiceNameKey.String(serviceName),
         attribute.String("service.namespace",serviceNameSpace),

      ),
   )
}

func spanExporter() (sdktrace.SpanExporter, error) {

	export_type := os.Getenv("EXPORT_TYPE")
	if export_type==""{
		export_type="OTLP"
	}

	if export_type == "JAEGER" {
	addr1 := os.Getenv("JAEGER_ENDPOINT");
	log.Info("exporting with JAEGER logger")
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(addr1)))
	 }
	if export_type == "OTLP" {
		otlp_protocol := os.Getenv("OTLP_FORMAT")
		if os.Getenv("OTLP_ENDPOINT") != "" {
			addr := os.Getenv("OTLP_ENDPOINT")
			u, err := url.Parse(addr)
				if err != nil {
					panic(err)
				}
			log.Info("Scheme: "+u.Scheme+" "+" Host: "+u.Host+" Path: "+u.Path)
			log.Info(u.User)
			if otlp_protocol == "HTTP"{
			log.Info("exporting with OTLP HTTP ")

				if u.Scheme == "https" {
					return otlptracehttp.New(
					context.Background(),
					otlptracehttp.WithEndpoint(u.Host),
					otlptracehttp.WithURLPath(u.Path),
					)
				} else {
					return otlptracehttp.New(
					context.Background(),
					otlptracehttp.WithInsecure(),
					otlptracehttp.WithEndpoint(u.Host),
					otlptracehttp.WithURLPath(u.Path),
					)
				}
			}else {
			log.Info("exporting with OTLP GRPC ")

				if u.Scheme == "https" {
					return otlptracegrpc.New(
					context.Background(),
					otlptracegrpc.WithEndpoint(u.Host),
					)
				} else {
					return otlptracegrpc.New(
					context.Background(),
					otlptracegrpc.WithInsecure(),
					otlptracegrpc.WithEndpoint(u.Host),
					)
				}

			}
			
		}
	}
	log.Info("exporting with STDOUT logger")
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(log.Writer()),
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
	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	),
	)
	otel.SetTextMapPropagator(propagation.TraceContext{})}

type errorHandler struct {
	log *logrus.Logger
}

func (eh errorHandler) Handle(err error) {
	eh.log.Error(err)
}

func main() {

    if serviceName=os.Getenv("SERVICE_NAME");serviceName==""{
        serviceName = "Productcatalog-service"
    }
    if serviceNameSpace=os.Getenv("SERVICE_NAMESPACE");serviceNameSpace==""{
         serviceNameSpace = "hipster"
    }
	initTracing()
	otel.SetErrorHandler(errorHandler{log: log})
	flag.Parse()

	// set injected latency
	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("failed to parse EXTRA_LATENCY (%s) as time.Duration: %+v", v, err)
		}
		extraLatency = v
		log.Infof("extra latency enabled (duration: %v)", extraLatency)
	} else {
		extraLatency = time.Duration(0)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-sigs
			log.Printf("Received signal: %s", sig)
			if sig == syscall.SIGUSR1 {
				reloadCatalog = true
				log.Infof("Enable catalog reloading")
			} else {
				reloadCatalog = false
				log.Infof("Disable catalog reloading")
			}
		}
	}()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Infof("starting grpc server at :%s", port)
	run(port)
	select {}
}

func run(port string) string {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}
	var srv *grpc.Server

	srv = grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	//srv = grpc.NewServer()

	svc := &productCatalog{}

	pb.RegisterProductCatalogServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)
	go srv.Serve(l)
	return l.Addr().String()
}

type productCatalog struct{}

func readCatalogFile(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()
	catalogJSON, err := ioutil.ReadFile("products.json")
	if err != nil {
		log.Fatalf("failed to open product catalog json file: %v", err)
		return err
	}
	if err := jsonpb.Unmarshal(bytes.NewReader(catalogJSON), catalog); err != nil {
		log.Warnf("failed to parse the catalog JSON: %v", err)
		return err
	}
	log.Info("successfully parsed product catalog json")
	return nil
}

func parseCatalog() []*pb.Product {
	if reloadCatalog || len(cat.Products) == 0 {
		err := readCatalogFile(&cat)
		if err != nil {
			return []*pb.Product{}
		}
	}
	return cat.Products
}

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (p *productCatalog) ListProducts(context.Context, *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)
	return &pb.ListProductsResponse{Products: parseCatalog()}, nil
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)
	var found *pb.Product
	for i := 0; i < len(parseCatalog()); i++ {
		if req.Id == parseCatalog()[i].Id {
			found = parseCatalog()[i]
		}
	}
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}
	return found, nil
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(extraLatency)
	// Intepret query as a substring match in name or description.
	var ps []*pb.Product
	for _, p := range parseCatalog() {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(p.Description), strings.ToLower(req.Query)) {
			ps = append(ps, p)
		}
	}
	return &pb.SearchProductsResponse{Results: ps}, nil
}
