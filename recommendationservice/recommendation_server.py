import os
import random
import time
import traceback
from concurrent import futures
# Google Cloud Debugger not supported for Python>3.8
#import googleclouddebugger
#import googlecloudprofiler
#from google.auth.exceptions import DefaultCredentialsError
import grpc
import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc
from opentelemetry import trace
#from opentelemetry.exporter.otlp.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter import jaeger
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
import json


#gRPC OTEL
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer, server_interceptor

#from opentelemetry.instrumentation.grpc.grpcext import intercept_server
# create a JaegerSpanExporter


export_type = os.environ["EXPORT_TYPE"]
serviceName = os.environ['SERVICE_NAME']
if serviceName=="":
    serviceName = "recommendation-service"

if export_type == "JAEGER" :
    jaeger_user = os.environ['USER']
    jaeger_password = os.environ['PASSWORD']
    jaeger_exporter = jaeger.JaegerSpanExporter(
        service_name=serviceName,
        # configure agent
        agent_host_name='jaeger',
        agent_port=6831,
        # optional: configure also collector
        #collector_host_name='jaeger',
        #collector_port=14268,
        #collector_endpoint='/api/traces?format=jaeger.thrift',
        # collector_protocol='http',
        username=jaeger_user, # optional
        password=jaeger_password, # optional
    )
    span_processor = BatchSpanProcessor(jaeger_exporter)
else:
    otlp_endpoint= os.environ['OTLP_ENDPOINT']
    otlp_header = os.environ['OTLP_HEADERS']
    if otlp_header != "" :
        otlp_exporter = OTLPSpanExporter(endpoint=otlp_endpoint , headers=json.loads(otlp_header))
    else :
        otlp_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
    span_processor = BatchSpanProcessor(otlp_exporter)


resource = Resource(attributes={
    "service.name": serviceName,
    "service.namespace": os.environ['SERVICE_NAMESPACE']
})
trace.set_tracer_provider(TracerProvider(resource=resource))
trace.get_tracer_provider().add_span_processor(span_processor)
grpc_server_instrumentor = GrpcInstrumentorServer()
grpc_server_instrumentor.instrument()
#from logger import getJSONLogger
#logger = getJSONLogger(‘recommendationservice-server’)


class RecommendationService(demo_pb2_grpc.RecommendationServiceServicer):
    def ListRecommendations(self, request, context):
        max_responses = 5
        # fetch list of products from product catalog stub
        cat_response = product_catalog_stub.ListProducts(demo_pb2.Empty())
        product_ids = [x.id for x in cat_response.products]
        filtered_products = list(set(product_ids)-set(request.product_ids))
        num_products = len(filtered_products)
        num_return = min(max_responses, num_products)
        # sample list of indicies to return
        indices = random.sample(range(num_products), num_return)
        # fetch product ids from indices
        prod_list = [filtered_products[i] for i in indices]
        #logger.info(“[Recv ListRecommendations] product_ids={}“.format(prod_list))
        # build and return response
        response = demo_pb2.ListRecommendationsResponse()
        response.product_ids.extend(prod_list)
        return response
    def Check(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.SERVING)
    def Watch(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)
if __name__ == "__main__":
    #logger.info("initializing recommendationservice")
    port = os.environ['PORT']
    catalog_addr = os.environ['PRODUCT_CATALOG_SERVICE_ADDR']
    #logger.info("product catalog address: " + catalog_addr +" PORT :" +port)
    channel = grpc.insecure_channel(catalog_addr)
    product_catalog_stub = demo_pb2_grpc.ProductCatalogServiceStub(channel)
    # create gRPC server
    server = grpc.server(futures.ThreadPoolExecutor())
    # add class to gRPC server
    service = RecommendationService()
    demo_pb2_grpc.add_RecommendationServiceServicer_to_server(service, server)
    health_pb2_grpc.add_HealthServicer_to_server(service, server)
    # start server
    #logger.info("listening on port: " + port)
    server.add_insecure_port('[::]:'+port)
    server.start()
    # keep alive
    try:
        while True:
            time.sleep(10000)
    except KeyboardInterrupt:
        server.stop(0)