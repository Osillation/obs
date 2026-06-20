// Proxy route - forwards OTel HTTP telemetry from the browser (or SSR) to
// the obs ingest endpoint, adding the OBS_INGEST_TOKEN server-side.
// The browser posts to /api/o/v1/traces; this route forwards to:
//   OTEL_EXPORTER_OTLP_ENDPOINT/v1/traces with the token header.

import { type NextRequest, NextResponse } from 'next/server';

const OTEL_ENDPOINT = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
const OBS_TOKEN = process.env.OBS_INGEST_TOKEN;

if (!OTEL_ENDPOINT || !OBS_TOKEN) {
  console.error('[obs] OTEL_EXPORTER_OTLP_ENDPOINT and OBS_INGEST_TOKEN must be set');
}

async function proxy(
  request: NextRequest,
  { params }: { params: { path: string[] } },
) {
  if (!OTEL_ENDPOINT || !OBS_TOKEN) {
    return new NextResponse('Observability not configured', { status: 503 });
  }

  const targetUrl = `${OTEL_ENDPOINT}/v1/${params.path.join('/')}`;
  const body = await request.arrayBuffer();

  const response = await fetch(targetUrl, {
    method: 'POST',
    headers: {
      'Content-Type': request.headers.get('Content-Type') ?? 'application/x-protobuf',
      'x-obs-token': OBS_TOKEN,
    },
    body,
  });

  return new NextResponse(response.body, {
    status: response.status,
    headers: {
      'Content-Type': response.headers.get('Content-Type') ?? 'application/json',
    },
  });
}

export const POST = proxy;
