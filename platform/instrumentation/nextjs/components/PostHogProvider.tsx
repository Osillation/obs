'use client';

// Wrap your root layout with this provider:
//   import { PostHogProvider } from '@/components/PostHogProvider';
//   <PostHogProvider><html>...</html></PostHogProvider>
//
// Required env vars (NEXT_PUBLIC_ prefix - exposed to browser):
//   NEXT_PUBLIC_POSTHOG_KEY=phc_xxxx   (get from PostHog UI after first deploy)
//   NEXT_PUBLIC_POSTHOG_HOST=/ingest   (proxied via next.config.js rewrite)

import posthog from 'posthog-js';
import { PostHogProvider as PHProvider, usePostHog } from 'posthog-js/react';
import { useEffect } from 'react';
import { usePathname, useSearchParams } from 'next/navigation';

function PostHogPageView() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const ph = usePostHog();

  useEffect(() => {
    if (pathname && ph) {
      const url = `${window.origin}${pathname}${searchParams.toString() ? `?${searchParams}` : ''}`;
      ph.capture('$pageview', { $current_url: url });
    }
  }, [pathname, searchParams, ph]);

  return null;
}

export function PostHogProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    const key = process.env.NEXT_PUBLIC_POSTHOG_KEY;
    const host = process.env.NEXT_PUBLIC_POSTHOG_HOST ?? '/ingest';

    if (!key) {
      console.warn('[obs] NEXT_PUBLIC_POSTHOG_KEY is not set - PostHog disabled');
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
    });
  }, []);

  return (
    <PHProvider client={posthog}>
      <PostHogPageView />
      {children}
    </PHProvider>
  );
}
