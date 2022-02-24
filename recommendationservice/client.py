#!/usr/bin/python
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys
import grpc
import demo_pb2
import demo_pb2_grpc

#from opencensus.trace.tracer import Tracer
#from opencensus.trace.exporters import stackdriver_exporter
#from opencensus.trace.ext.grpc import client_interceptor
# server.py

from opentelemetry import trace
from opentelemetry.exporter import jaeger
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchExportSpanProcessor
from opentelemetry import trace
#gRPC Opentelemetry Instrumentation
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient, client_interceptor
from opentelemetry.sdk.trace.export import (
    ConsoleSpanExporter,
    SimpleExportSpanProcessor,
)
from opentelemetry import metrics
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import ConsoleMetricsExporter

# create a JaegerSpanExporter
print("client")
jaeger_exporter = jaeger.JaegerSpanExporter(
    service_name='client',
    # configure agent
    agent_host_name='jaeger',
    agent_port=6831,
    # optional: configure also collector
    #collector_host_name='jaeger',
    #collector_port=14268,
    #collector_endpoint='/api/traces?format=jaeger.thrift',
    # collector_protocol='http',
    # username=xxxx, # optional
    # password=xxxx, # optional
)
print("batch")
# Create a BatchExportSpanProcessor and add the exporter to it
span_processor = BatchExportSpanProcessor(jaeger_exporter)

print("tracer")
# add to the tracer

# Set meter provider to opentelemetry-sdk's MeterProvider
metrics.set_meter_provider(MeterProvider())

from logger import getJSONLogger
logger = getJSONLogger('recommendationservice-server')

if __name__ == "__main__":
    # get port
    port = "8080"
    exporter = jaeger_exporter
        #tracer = Tracer(exporter=exporter)
        #tracer_interceptor = client_interceptor.OpenCensusClientInterceptor(tracer, host_port='localhost:'+port)
    
        #tracer_interceptor = client_interceptor.OpenCensusClientInterceptor()
    print("grpc channel init")
    channel = grpc.insecure_channel("localhost:8080")
    print("stub init")
    stub = demo_pb2_grpc.RecommendationServiceStub(channel)
    print("request int")
    request = demo_pb2.ListRecommendationsRequest(user_id="test", product_ids=["test"])
    print("response init")
    response = stub.ListRecommendations(request)
    print("logging init")
    logger.info(response)