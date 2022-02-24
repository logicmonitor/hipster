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
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/stdout"

	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	port            = "8081"
	defaultCurrency = "USD"
	cookieMaxAge    = 60 * 60 * 48

	cookiePrefix    = "shop_"
	cookieSessionID = cookiePrefix + "session-id"
	cookieCurrency  = cookiePrefix + "currency"
)

var (
	catalogMutex          *sync.Mutex
	log                   *logrus.Logger
	whitelistedCurrencies = map[string]bool{
		"USD": true,
		"EUR": true,
		"CAD": true,
		"JPY": true,
		"GBP": true,
		"TRY": true}

    serviceName string
    serviceNameSpace string
)

type ctxKeySessionID struct{}

type frontendServer struct {
	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn

	currencySvcAddr string
	currencySvcConn *grpc.ClientConn

	cartSvcAddr string
	cartSvcConn *grpc.ClientConn

	recommendationSvcAddr string
	recommendationSvcConn *grpc.ClientConn

	checkoutSvcAddr string
	checkoutSvcConn *grpc.ClientConn

	shippingSvcAddr string
	shippingSvcConn *grpc.ClientConn

	adSvcAddr string
	adSvcConn *grpc.ClientConn
}

func frontendserverConstructor(productCatalogSvcAddr string, currencySvcAddr string, cartSvcAddr string, recommendationSvcAddr string, checkoutSvcAddr string, shippingSvcAddr string, adSvcAddr string) *frontendServer {
	obj := new(frontendServer)

	obj.productCatalogSvcAddr = productCatalogSvcAddr
	obj.currencySvcAddr = currencySvcAddr
	obj.cartSvcAddr = cartSvcAddr
	obj.recommendationSvcAddr = recommendationSvcAddr
	obj.checkoutSvcAddr = checkoutSvcAddr
	obj.shippingSvcAddr = shippingSvcAddr
	obj.adSvcAddr = adSvcAddr

	return obj
}

func main() {

    if serviceName=os.Getenv("SERVICE_NAME");serviceName==""{
        serviceName = "Frontend-service"
    }
    if serviceNameSpace=os.Getenv("SERVICE_NAMESPACE");serviceNameSpace==""{
            serviceNameSpace = "hipster"
    }

	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	initTracing()

	ctx := context.Background()
	srvPort := port
	addr := os.Getenv("LISTEN_ADDR")
	var PRODUCT_CATALOG_SERVICE_ADDR = os.Getenv("PRODUCT_CATALOG_SERVICE_ADDR")
	if PRODUCT_CATALOG_SERVICE_ADDR=="" {
		PRODUCT_CATALOG_SERVICE_ADDR="localhost:4000"
	}
	var CURRENCY_SERVICE_ADDR = os.Getenv("CURRENCY_SERVICE_ADDR")
		if CURRENCY_SERVICE_ADDR=="" {
		CURRENCY_SERVICE_ADDR="localhost:9001"
	}
	var CART_SERVICE_ADDR = os.Getenv("CART_SERVICE_ADDR")
		if CART_SERVICE_ADDR=="" {
		CART_SERVICE_ADDR="localhost:8100"
	}
	var RECOMMENDATION_SERVICE_ADDR = os.Getenv("RECOMMENDATION_SERVICE_ADDR")
		if RECOMMENDATION_SERVICE_ADDR=="" {
		RECOMMENDATION_SERVICE_ADDR="localhost:8082"
	}
	var CHECKOUT_SERVICE_ADDR = os.Getenv("CHECKOUT_SERVICE_ADDR")
		if CHECKOUT_SERVICE_ADDR=="" {
		CHECKOUT_SERVICE_ADDR="localhost:5050"
	}
	var SHIPPING_SERVICE_ADDR = os.Getenv("SHIPPING_SERVICE_ADDR")
		if SHIPPING_SERVICE_ADDR=="" {
		SHIPPING_SERVICE_ADDR="localhost:5551"
	}
	var AD_SERVICE_ADDR = os.Getenv("AD_SERVICE_ADDR")
		if AD_SERVICE_ADDR=="" {
		AD_SERVICE_ADDR="localhost:9555"
	}

	svc := frontendserverConstructor(PRODUCT_CATALOG_SERVICE_ADDR, CURRENCY_SERVICE_ADDR, CART_SERVICE_ADDR, RECOMMENDATION_SERVICE_ADDR, CHECKOUT_SERVICE_ADDR, SHIPPING_SERVICE_ADDR, AD_SERVICE_ADDR)

	mustConnGRPC(ctx, &svc.currencySvcConn, svc.currencySvcAddr)
	mustConnGRPC(ctx, &svc.productCatalogSvcConn, svc.productCatalogSvcAddr)
	mustConnGRPC(ctx, &svc.cartSvcConn, svc.cartSvcAddr)
	mustConnGRPC(ctx, &svc.recommendationSvcConn, svc.recommendationSvcAddr)
	mustConnGRPC(ctx, &svc.shippingSvcConn, svc.shippingSvcAddr)
	mustConnGRPC(ctx, &svc.checkoutSvcConn, svc.checkoutSvcAddr)
	mustConnGRPC(ctx, &svc.adSvcConn, svc.adSvcAddr)

	r := mux.NewRouter()
	r.Use(MuxMiddleware(), otelmux.Middleware(serviceName))
	r.HandleFunc("/", svc.homeHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/product/{id}", svc.productHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/cart", svc.viewCartHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/cart", svc.addToCartHandler).Methods(http.MethodPost)
	r.HandleFunc("/cart/empty", svc.emptyCartHandler).Methods(http.MethodPost)
	r.HandleFunc("/setCurrency", svc.setCurrencyHandler).Methods(http.MethodPost)
	r.HandleFunc("/logout", svc.logoutHandler).Methods(http.MethodGet)
	r.HandleFunc("/cart/checkout", svc.placeOrderHandler).Methods(http.MethodPost)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") })
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	var handler http.Handler = r
	handler = &logHandler{log: log, next: handler} // add logging
	handler = ensureSessionID(handler)             // add session ID
	log.Infof("starting server on " + addr + ":" + srvPort)
	log.Fatal(http.ListenAndServe(addr+":"+srvPort, handler))
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
		if os.Getenv("OTLP_ENDPOINT") != "" {
			addr := os.Getenv("OTLP_ENDPOINT")
			if addr == "" {
				addr="localhost:4317"
			}
			return otlp.NewExporter(
				context.Background(),
				otlp.WithInsecure(),
				otlp.WithAddress(addr),
			)
		}
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

type errorHandler struct {
	log *logrus.Logger
}

func (eh errorHandler) Handle(err error) {
	eh.log.Error(err)
}

func mustConnGRPC(ctx context.Context, conn **grpc.ClientConn, addr string) {
	var err error

	*conn, err = grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", addr))
	}
}
