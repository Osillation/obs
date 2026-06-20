// Next.js server-side OTel instrumentation with PostHog session ID correlation.
// Next.js auto-loads this file before any request is handled.

export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { NodeSDK } = await import('@opentelemetry/sdk-node');
    const { getNodeAutoInstrumentations } = await import('@opentelemetry/auto-instrumentations-node');
    const { OTLPTraceExporter } = await import('@opentelemetry/exporter-trace-otlp-http');
    const { Resource } = await import('@opentelemetry/resources');
    const { SEMRESATTRS_SERVICE_NAME, SEMRESATTRS_SERVICE_NAMESPACE } =
      await import('@opentelemetry/semantic-conventions');
    const { trace } = await import('@opentelemetry/api');

    const endpoint = process.env.OBS_LOCAL === 'true'
      ? 'http://localhost:4318'
      : process.env.OTEL_EXPORTER_OTLP_ENDPOINT ?? 'http://localhost:4318';

    const headers: Record<string, string> = {};
    if (process.env.OBS_LOCAL !== 'true' && process.env.OBS_INGEST_TOKEN) {
      headers['x-obs-token'] = process.env.OBS_INGEST_TOKEN;
    }

    const sdk = new NodeSDK({
      resource: new Resource({
        [SEMRESATTRS_SERVICE_NAME]: process.env.OTEL_SERVICE_NAME ?? 'nextjs-frontend',
        [SEMRESATTRS_SERVICE_NAMESPACE]: process.env.OTEL_SERVICE_NAMESPACE ?? 'default',
      }),
      traceExporter: new OTLPTraceExporter({ url: `${endpoint}/v1/traces`, headers }),
      instrumentations: [
        getNodeAutoInstrumentations({
          '@opentelemetry/instrumentation-fs': { enabled: false },
        }),
      ],
    });

    sdk.start();
  }
}

// Middleware hook: reads obs_ph_session cookie and adds posthog.session_id
// attribute to the active span. Import this in middleware.ts:
//   import { addSessionIdToSpan } from './instrumentation';
export function addSessionIdToSpan(cookieHeader: string | null) {
  if (!cookieHeader) return;
  const match = cookieHeader.match(/obs_ph_session=([^;]+)/);
  if (!match) return;
  const sessionId = match[1];
  // Import at call site to avoid circular deps at module load
  const { trace } = require('@opentelemetry/api');
  const span = trace.getActiveSpan();
  if (span) {
    span.setAttribute('posthog.session_id', sessionId);
  }
}
