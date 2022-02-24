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

import grpc

import demo_pb2
import demo_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('emailservice-client')

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
# Create a BatchExportSpanProcessor and add the exporter to it
span_processor = BatchExportSpanProcessor(jaeger_exporter)

def send_confirmation_email(email, order):
  channel = grpc.insecure_channel('0.0.0.0:8081')
  #channel = grpc.intercept_channel(channel, tracer_interceptor)
  stub = demo_pb2_grpc.EmailServiceStub(channel)
  try:
    response = stub.SendOrderConfirmation(demo_pb2.SendOrderConfirmationRequest(
      email = email,
      order = order
    ))
    logger.info('Request sent.')
  except grpc.RpcError as err:
    logger.error(err.details())
    logger.error('{}, {}'.format(err.code().name, err.code().value))

if __name__ == '__main__':
  logger.info('Client for email service.')
