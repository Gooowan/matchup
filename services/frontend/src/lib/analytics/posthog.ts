import { browser } from '$app/environment';
import posthog from 'posthog-js';

let initialised = false;

export function initPostHog(): void {
	if (!browser) return;

	const key = import.meta.env.VITE_POSTHOG_KEY;
	if (!key) return; // graceful no-op in dev when key is unset

	const host = import.meta.env.VITE_POSTHOG_HOST ?? 'https://app.posthog.com';

	posthog.init(key, {
		api_host: host,
		capture_pageview: false, // We track page views manually on SvelteKit navigation
		capture_pageleave: true,
		// IMPORTANT for Capacitor iOS: WKWebView has unreliable cookie behaviour.
		// localStorage is stable across app sessions on iOS.
		persistence: 'localStorage'
	});

	initialised = true;
}

/** Call on SvelteKit page navigation to track page views in SPAs. */
export function capturePageView(url: string): void {
	if (!browser || !initialised) return;
	posthog.capture('$pageview', { $current_url: url });
}

export function captureSwipe(action: 'LIKE' | 'PASS', source: string): void {
	if (!browser || !initialised) return;
	posthog.capture('swipe', { action, source });
}

export function captureMatch(): void {
	if (!browser || !initialised) return;
	posthog.capture('match_created');
}

export function captureAuthEvent(event: 'login' | 'register' | 'logout'): void {
	if (!browser || !initialised) return;
	posthog.capture(`auth_${event}`);
}

/** Identify the authenticated user (opaque ID only — no PII). */
export function identifyUser(userId: string): void {
	if (!browser || !initialised) return;
	posthog.identify(userId);
}

/** Reset PostHog identity on logout. */
export function resetUser(): void {
	if (!browser || !initialised) return;
	posthog.reset();
}
