# Ad Service

The Ad service provides advertisement based on context keys. If no context keys are provided then it returns random ads.

## Building locally

The Ad service uses gradlew to compile/install/distribute. Gradle wrapper is already part of the source code. To build Ad Service, run:

```
./gradlew build
```
It will create executable script src/adservice/build/install/hipstershop/bin/AdService

### Upgrading gradle version
If you need to upgrade the version of gradle then run

```
./gradlew wrapper --gradle-version <new-version>
```

## Building docker image

From `adservice/`, run:

```
docker build ./
```

## Instrumentation

For instrumenting java application with opentelemetry, first download the opentelemetry-javaagent-all.jar

> Note : Download opentelemetry-javaagent-all.jar : https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v1.7.0/opentelemetry-javaagent-all.jar

To enable the instrumentation agent use the -javaagent flag to the JVM.

Configuration parameters are passed as Java system properties (-D flags) or as environment variables.

To configure the java agent, the following attributes can be set:
  - otel.traces.exporter   // default is otlp
  - otel.exporter.otlp.endpoint	//
  - otel.exporter.otlp.headers	//Key-value pairs separated by commas to pass as request headers
  - otel.exporter.otlp.protocol	//values can be grpc or http/protobuf. Default is grpc
  - otel.resource.attributes	//Specify resource attributes in the following format: key1=val1,key2=val2,key3=val3
  - otel.service.name	

For information about using OTLP/HTTP or OTLP/gRPC refer [this.](../README.md#When-to-use-OTLP/HTTP-or-OTLP/gRPC)

For more information about configurations,visit [Opentelemetry Configuration java.](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters)


Finally, the command to auto instrument java application will look something similar to below command.

```
        java -javaagent:opentelemetry-javaagent-all.jar \
        -Dotel.exporter=otlp \
        -Dotel.resource.attributes=$OTEL_RESOURCE_ATTRIBUTES ,service.name=$SERVICE_NAME,host.name=$HOST_NAME \
        -Dotel.exporter.otlp.endpoint=http://localhost:4317 \
        -Dotel.exporter.otlp.insecure=true \
        -Dotel.exporter.otlp.protocol=$OTLP_PROTOCOL \
        -jar hipstershop-0.1.0-SNAPSHOT-fat.jar
```


