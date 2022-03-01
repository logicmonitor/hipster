// 'use strict';
const opentelemetry = require('@opentelemetry/api');

const { WebTracerProvider } = require('@opentelemetry/sdk-trace-web');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { Resource } = require('@opentelemetry/resources');
const os = require('os');
const { BatchSpanProcessor } = require("@opentelemetry/sdk-trace-base");
//const { BasicTracerProvider, BatchSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-proto');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const OTLPTraceExporterGRPC = require('@opentelemetry/exporter-trace-otlp-grpc');

var OTLP_FORMAT = process.env.OTLP_FORMAT || "HTTP"
console.log("Exporter type  Set To: OTLP" + OTLP_FORMAT)

opentelemetry.diag.setLogger(
    new opentelemetry.DiagConsoleLogger(),
    opentelemetry.DiagLogLevel.DEBUG,
);
var subs = [];

var otlpAttributes = process.env.OTEL_RESOURCE_ATTRIBUTES
try {
    if (otlpAttributes != null) {
        subs = otlpAttributes.split(/[=,\,]/)
    }
    var keys = []
    var value = []
    for (var i = 0; i < subs.length; i++) {
        keys[i] = subs[i];
        value[i] = subs[++i];
    }
    var attributeslist = {};
    for (var i = 0; i < keys.length; i++) {
        attributeslist[keys[i]] = value[i];
    }
} catch (error) {
    console.error(error);
}
const collectorOptions = {
    url: process.env.OTLP_ENDPOINT, // url is optional and can be omitted - default is http://localhost:55681/v1/traces
    serviceName: process.env.SERVICE_NAME || "payment",
    // an optional object containing custom headers to be sent with each request will only work with http
    attributes: attributeslist
};

const identifier = process.env.HOSTNAME || os.hostname()
const provider = new WebTracerProvider({
    resource: new Resource({
        [SemanticResourceAttributes.INSTANCE_ID]: identifier,
        [SemanticResourceAttributes.SERVICE_NAME]: process.env.SERVICE_NAME || 'docker-payment-service',
        [SemanticResourceAttributes.SERVICE_NAMESPACE]: process.env.SERVICE_NAMESPACE || 'LM_DOCKER_HIPSTERSHOP',
    }),
});
if (OTLP_FORMAT == "HTTP") {
    const exporter = new OTLPTraceExporter(collectorOptions)
    provider.addSpanProcessor(new BatchSpanProcessor(exporter, {
        // The maximum queue size. After the size is reached spans are dropped.
        maxQueueSize: 1000,
        // The interval between two consecutive exports
    }));
} else {
    const exporter = new OTLPTraceExporterGRPC.OTLPTraceExporter(collectorOptions)
    provider.addSpanProcessor(new BatchSpanProcessor(exporter, {
        // The maximum queue size. After the size is reached spans are dropped.
        maxQueueSize: 1000,
        // The interval between two consecutive exports
    }));
}

provider.register();

// //provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
registerInstrumentations({
    instrumentations: [
        getNodeAutoInstrumentations({
            // load custom configuration for http instrumentation
            '@opentelemetry/instrumentation-http': {},
            '@opentelemetry/instrumentation-grpc': {},
            '@opentelemetry/instrumentation-dns': {},
            '@opentelemetry/instrumentation-express': {}

        }),
    ],
})
console.log("tracing initialized")