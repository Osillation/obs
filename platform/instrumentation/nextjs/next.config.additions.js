// Add these rewrites to your next.config.js / next.config.ts
// They proxy PostHog event calls through your own domain to avoid ad-blockers.

// async rewrites() {
//   return [
//     {
//       source: '/ingest/static/:path*',
//       destination: `${process.env.NEXT_PUBLIC_POSTHOG_HOST}/static/:path*`,
//     },
//     {
//       source: '/ingest/:path*',
//       destination: `${process.env.NEXT_PUBLIC_POSTHOG_HOST}/:path*`,
//     },
//   ];
// },
//
// Required env vars:
//   Server-side (Vercel or container):
//     OTEL_EXPORTER_OTLP_ENDPOINT=https://ingest.<CLIENT_DOMAIN>
//     OBS_INGEST_TOKEN=<secret from config.env>
//     OTEL_SERVICE_NAME=nextjs-frontend
//     OTEL_SERVICE_NAMESPACE=<client-name>
//
//   Public (browser):
//     NEXT_PUBLIC_POSTHOG_KEY=phc_<get from PostHog UI>
//     NEXT_PUBLIC_POSTHOG_HOST=/ingest
