FROM adoptopenjdk/openjdk11 as builder

WORKDIR /usr/app
COPY . .
RUN ./gradlew build

RUN curl -LJO https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v1.7.0/opentelemetry-javaagent-all.jar

WORKDIR /usr/app
EXPOSE 9555

ENTRYPOINT java -javaagent:opentelemetry-javaagent-all.jar -Dotel.exporter=$EXPORT_TYPE -Dotel.resource.attributes=service.name=$SERVICE_NAME,host.name=$HOSTNAME,service.namespace=$SERVICE_NAMESPACE,$OTEL_RESOURCE_ATTRIBUTES -Dotel.exporter.$EXPORT_TYPE.endpoint=$EXPORTER_ENDPOINT  -Dotel.javaagent.debug=true -Dotel.exporter.otlp.protocol=$OTLP_PROTOCOL -jar ./build/libs/hipstershop-0.1.0-SNAPSHOT-fat.jar
