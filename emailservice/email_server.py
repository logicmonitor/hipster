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

from concurrent import futures
import argparse
import os
import sys
import time
import grpc
from jinja2 import Environment, FileSystemLoader, select_autoescape, TemplateError
from google.api_core.exceptions import GoogleAPICallError
from google.auth.exceptions import DefaultCredentialsError

import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc


from opentelemetry import trace
from opentelemetry.exporter import jaeger
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter as HTTPEXPORTER
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter as GRPCEXPORTER
import json
#gRPC OTEL
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer, server_interceptor

from logger import getJSONLogger
logger = getJSONLogger('emailservice-server')

export_type = os.environ["EXPORT_TYPE"]
serviceName="DEFAULT_NAME"
try:
    serviceName=os.environ['SERVICE_NAME']
except:
    serviceName = "email-service"

if export_type == "JAEGER" :
    logger.info("JAEGER SET")
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

if export_type == 'OTLP':
    otlp_endpoint= os.environ['OTLP_ENDPOINT']
    otlp_format=os.environ['OTLP_FORMAT']
    attributes={}
    try:
        resource_attributes = os.environ['OTEL_RESOURCE_ATTRIBUTES']
        resource_attributes = resource_attributes.split(",")
        for attribute in resource_attributes:
            keyvalue= attribute.split("=")
            attributes[keyvalue[0]]=keyvalue[1]
    except:
        logger.debug("OTEL RESOURCE ATTRIBURTES NOT SET") 
    try:
        attributes["service.namespace"]=os.environ['SERVICE_NAMESPACE']
    except:
        attributes["service.namespace"]="DEFAULT_NAMESPACE"
    attributes["service.name"]=serviceName
    resource = Resource(attributes)
    if otlp_format == "HTTP":
        otlp_headers="{}"
        try :
            otlp_header = os.environ['OTEL_EXPORTER_OTLP_HEADERS']
            headers= otlp_header.split(",")
            if len(headers) > 0 :
                otlp_headers=""
                otlp_headers+="{"
                for header in headers :
                    keyvalue= header.split("=")
                    if len(keyvalue)==2:
                        otlp_headers+="\""+keyvalue[0]+"\""
                        otlp_headers+=":"
                        otlp_headers+="\""+keyvalue[1]+"\""
                    otlp_headers+=","
                otlp_headers=otlp_headers[:-1]
                otlp_headers+="}"
                logger.info(otlp_headers)
        except Exception as e:
            logger.info("OTLP HEADERS NOT SET"+ str(e)) 
        otlp_exporter = HTTPEXPORTER(endpoint=otlp_endpoint , headers=json.loads(otlp_headers))
        span_processor = BatchSpanProcessor(otlp_exporter)
        trace.set_tracer_provider(TracerProvider(resource=resource))
        trace.get_tracer_provider().add_span_processor(span_processor)
    else :
        otlp_exporter = GRPCEXPORTER(endpoint=otlp_endpoint)
        span_processor = BatchSpanProcessor(otlp_exporter)
        trace.set_tracer_provider(TracerProvider(resource=resource))
        trace.get_tracer_provider().add_span_processor(span_processor)

grpc_server_instrumentor = GrpcInstrumentorServer()
grpc_server_instrumentor.instrument()
# Create a BatchExportSpanProcessor and add the exporter to it

# Loads confirmation email template from file
env = Environment(
    loader=FileSystemLoader('templates'),
    autoescape=select_autoescape(['html', 'xml'])
)
template = env.get_template('confirmation.html')

class BaseEmailService(demo_pb2_grpc.EmailServiceServicer):
  def Check(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.SERVING)
  def Watch(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)

class EmailService(BaseEmailService):
  def __init__(self):
    super().__init__()

  @staticmethod
  def send_email(client, email_address, content):
    response = client.send_message(
      sender = client.sender_path(project_id, region, sender_id),
      envelope_from_authority = '',
      header_from_authority = '',
      envelope_from_address = from_address,
      simple_message = {
        "from": {
          "address_spec": from_address,
        },
        "to": [{
          "address_spec": email_address
        }],
        "subject": "Your Confirmation Email",
        "html_body": content
      }
    )
    logger.info("Message sent: {}".format(response.rfc822_message_id))

  def SendOrderConfirmation(self, request, context):
    email = request.email
    order = request.order

    try:
      confirmation = template.render(order = order)
    except TemplateError as err:
      context.set_details("An error occurred when preparing the confirmation mail.")
      logger.error(err.message)
      context.set_code(grpc.StatusCode.INTERNAL)
      return demo_pb2.Empty()

    try:
      EmailService.send_email(self.client, email, confirmation)
    except GoogleAPICallError as err:
      context.set_details("An error occurred when sending the email.")
      print(err.message)
      context.set_code(grpc.StatusCode.INTERNAL)
      return demo_pb2.Empty()

    return demo_pb2.Empty()

class DummyEmailService(BaseEmailService):
  def SendOrderConfirmation(self, request, context):
    logger.info('A request to send order confirmation email to {} has been received.'.format(request.email))
    return demo_pb2.Empty()

class HealthCheck():
  def Check(self, request, context):
    return health_pb2.HealthCheckResponse(
      status=health_pb2.HealthCheckResponse.SERVING)

def start(dummy_mode):
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  service = None
  if dummy_mode:
    service = DummyEmailService()
  else:
    raise Exception('non-dummy mode not implemented yet')

  demo_pb2_grpc.add_EmailServiceServicer_to_server(service, server)
  health_pb2_grpc.add_HealthServicer_to_server(service, server)

  port = os.environ['PORT']
  logger.info("listening on port: "+port)
  server.add_insecure_port('[::]:'+port)
  server.start()
  try:
    while True:
      time.sleep(3600)
  except KeyboardInterrupt:
    server.stop(0)

if __name__ == '__main__':
  logger.info('starting the email service in dummy mode.')

  # Profiler
  try:
    if "DISABLE_PROFILER" in os.environ:
      raise KeyError()
    else:
      logger.info("Profiler enabled.")
      #initStackdriverProfiling()
  except KeyError:
      logger.info("Profiler disabled.")

  # Tracing
 
  logger.info("Tracing enabled.")
  #sampler = always_on.AlwaysOnSampler()
  #tracer_interceptor = server_interceptor.OpenCensusServerInterceptor(sampler, exporter)
  start(dummy_mode = True)
