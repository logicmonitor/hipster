FROM adoptopenjdk/openjdk11 as builder
#ENV EXPORTER=jaeger
#ENV EXPORTER_ENDPOINT=jaeger:14250

WORKDIR /usr/app
COPY ./build/libs/hipstershop-0.1.0-SNAPSHOT-fat.jar /usr/app/
#COPY ./tracinglib/opentelemetry-javaagent-all.jar /usr/app/

RUN curl -LJO https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v0.8.0/opentelemetry-javaagent-all.jar

WORKDIR /usr/app
EXPOSE 9555

#ENTRYPOINT ["java","-javaagent:opentelemetry-javaagent-all.jar","-Dotel.exporter=jaeger","-Dotel.exporter.jaeger.service.name=AdService","-Dotel.exporter.jaeger.endpoint=jaeger:14250","-jar", "hipstershop-0.1.0-SNAPSHOT-fat.jar"]
ARG EXPORTER
ENV EXPORTER={EXPORT_TYPE}
ARG EXPORTER_ENDPOINT
ENV EXPORT_TYPE={EXPORT_TYPE}
ARG SERVICE_NAME
ENV SERVICE_NAME={SERVICE_NAME}
ARG HOST_NAME
ENV HOST_NAME={HOST_NAME}
ARG OTEL_RESOURCE_ATTRIBUTES
ENV OTEL_RESOURCE_ATTRIBUTES={OTEL_RESOURCE_ATTRIBUTES}

ENTRYPOINT java -javaagent:opentelemetry-javaagent-all.jar -Dotel.exporter=$EXPORTER -Dotel.resource.attributes=$OTEL_RESOURCE_ATTRIBUTES,service.name=$SERVICE_NAME,host.name=$HOST_NAME, -Dotel.exporter.$EXPORTER.endpoint=$EXPORTER_ENDPOINT -Dotel.exporter.otlp.insecure=true -jar hipstershop-0.1.0-SNAPSHOT-fat.jar