'use client';

// Wrap your root layout with this provider.
// Required env vars:
//   NEXT_PUBLIC_POSTHOG_KEY=phc_xxxx
//   NEXT_PUBLIC_POSTHOG_HOST=/ingest

import posthog from 'posthog-js';
import { PostHogProvider as PHProvider, usePostHog } from 'posthog-js/react';
import { useEffect, useRef } from 'react';
import { usePathname, useSearchParams } from 'next/navigation';

// Monotonic UTC epoch milliseconds — immune to DST transitions mid-session.
// performance.timeOrigin is set once at page load (UTC ms).
// performance.now() is monotonic and relative to timeOrigin.
const epochMs = (): number => Math.floor(performance.timeOrigin + performance.now());

function PostHogPageView() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const ph = usePostHog();

  useEffect(() => {
    if (pathname && ph) {
      const url = `${window.origin}${pathname}${searchParams.toString() ? `?${searchParams}` : ''}`;
      ph.capture('$pageview', {
        $current_url: url,
        obs_timestamp_ms: epochMs(),
      });
    }
  }, [pathname, searchParams, ph]);

  return null;
}

// Inject PostHog session ID into the OTel resource so every backend span
// carries posthog_session_id — enabling SigNoz → PostHog deep links.
function injectSessionIdIntoOTel(sessionId: string) {
  if (typeof window === 'undefined') return;
  // Store session ID in a cookie so the Next.js server instrumentation
  // can read it and set it on OTel spans.
  document.cookie = `obs_ph_session=${sessionId}; path=/; SameSite=Strict`;
}

// Capture traceparent from network requests and attach to PostHog events.
// This creates PostHog → SigNoz deep links via trace_id.
function instrumentNetworkRequests(ph: ReturnType<typeof usePostHog>) {
  if (typeof window === 'undefined' || !ph) return;

  const origFetch = window.fetch;
  window.fetch = async (...args) => {
    const response = await origFetch(...args);
    const traceparent = response.headers.get('traceparent');
    if (traceparent) {
      // traceparent format: 00-<trace_id_128hex>-<span_id_hex>-<flags>
      const parts = traceparent.split('-');
      if (parts.length === 4) {
        ph.capture('$network_request', {
          trace_id: parts[1],
          span_id: parts[2],
          obs_timestamp_ms: epochMs(),
          $current_url: typeof args[0] === 'string' ? args[0] : String(args[0]),
        });
      }
    }
    return response;
  };
}

export function PostHogProvider({ children }: { children: React.ReactNode }) {
  const initialized = useRef(false);
  const ph = usePostHog();

  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;

    const key = process.env.NEXT_PUBLIC_POSTHOG_KEY;
    const host = process.env.NEXT_PUBLIC_POSTHOG_HOST ?? '/ingest';

    if (!key) {
      console.warn('[obs] NEXT_PUBLIC_POSTHOG_KEY is not set — PostHog disabled');
      return;
    }

    posthog.init(key, {
      api_host: host,
      capture_pageview: false,
      capture_pageleave: true,
      session_recording: {
        maskAllInputs: false,
        maskInputOptions: { password: true },
      },
      persistence: 'localStorage+cookie',
      loaded: (ph) => {
        const sessionId = ph.get_session_id();
        if (sessionId) {
          injectSessionIdIntoOTel(sessionId);
          instrumentNetworkRequests(ph);
        }
      },
    });
  }, []);

  return (
    <PHProvider client={posthog}>
      <PostHogPageView />
      {children}
    </PHProvider>
  );
}
