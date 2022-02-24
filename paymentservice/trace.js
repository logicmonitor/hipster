const { logger } = require('@opencensus/core');
const { JaegerTraceExporter } = require('@opencensus/exporter-jaeger');
const tracing = require('@opencensus/nodejs');

// Add service name and jaeger options
const jaegerOptions = {
    serviceName: 'Payments',
    host: 'jaeger',
    port: 6832,
    tags: [{ key: 'opencensus-exporter-jeager', value: '0.0.9' }],
    bufferTimeout: 10, // time in milliseconds
    logger: logger.logger('debug')
};

const exporter = new JaegerTraceExporter(jaegerOptions);
tracing.registerExporter(exporter).start();