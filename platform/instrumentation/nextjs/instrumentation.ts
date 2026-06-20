// Next.js auto-loads this file before any request is handled.
// Ref: https://nextjs.org/docs/app/building-your-application/optimizing/instrumentation

export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { NodeSDK } = await import('@opentelemetry/sdk-node');
    const { getNodeAutoInstrumentations } = await import('@opentelemetry/auto-instrumentations-node');
    const { OTLPTraceExporter } = await import('@opentelemetry/exporter-trace-otlp-http');
    const { Resource } = await import('@opentelemetry/resources');
    const { SEMRESATTRS_SERVICE_NAME, SEMRESATTRS_SERVICE_NAMESPACE } =
      await import('@opentelemetry/semantic-conventions');

    const sdk = new NodeSDK({
      resource: new Resource({
        [SEMRESATTRS_SERVICE_NAME]: process.env.OTEL_SERVICE_NAME ?? 'nextjs-frontend',
        [SEMRESATTRS_SERVICE_NAMESPACE]: process.env.OTEL_SERVICE_NAMESPACE ?? 'default',
      }),
      traceExporter: new OTLPTraceExporter({
        // Routes through the proxy route (app/api/o/[...path]/route.ts)
        // so the OBS_INGEST_TOKEN is added server-side - browser never sees the token.
        url: '/api/o/v1/traces',
      }),
      instrumentations: [
        getNodeAutoInstrumentations({
          '@opentelemetry/instrumentation-fs': { enabled: false },
        }),
      ],
    });

    sdk.start();
  }
}
