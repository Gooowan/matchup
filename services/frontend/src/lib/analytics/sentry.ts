import { browser } from '$app/environment';

// Dynamically imported to avoid bundling when env var is absent.
// The @sentry/capacitor init pattern requires @sentry/sveltekit as the second arg
// so that SvelteKit route tracking and error boundary hooks are wired up correctly.
// On iOS (Capacitor WebView) @sentry/capacitor also captures native crashes.

let sentryLoaded = false;

export async function initSentry(): Promise<void> {
	if (!browser) return;

	const dsn = import.meta.env.VITE_SENTRY_DSN;
	if (!dsn) return; // graceful no-op in dev when DSN is unset

	try {
		const [Sentry, SentrySvelteKit] = await Promise.all([
			import('@sentry/capacitor'),
			import('@sentry/sveltekit')
		]);

		Sentry.init(
			{
				dsn,
				environment: import.meta.env.MODE,
				release: `matchup@${import.meta.env.VITE_APP_VERSION ?? 'dev'}`,
				tracesSampleRate: 0.1,
				beforeSend(event) {
					// Scrub PII: remove auth tokens from request context
					if (event.request) {
						event.request.cookies = '';
						if (event.request.headers) {
							delete event.request.headers['Authorization'];
							delete event.request.headers['Cookie'];
						}
					}
					return event;
				}
			},
			SentrySvelteKit.init
		);

		sentryLoaded = true;
	} catch (e) {
		// Non-fatal: Sentry init failure should never crash the app
		console.warn('[Sentry] Failed to initialise:', e);
	}
}

/** Call after login to correlate errors to the authenticated user (ID only — no PII). */
export async function setSentryUser(userId: string): Promise<void> {
	if (!browser || !sentryLoaded) return;
	const Sentry = await import('@sentry/capacitor');
	Sentry.setUser({ id: userId });
}

/** Call on logout to clear user context from future error events. */
export async function clearSentryUser(): Promise<void> {
	if (!browser || !sentryLoaded) return;
	const Sentry = await import('@sentry/capacitor');
	Sentry.setUser(null);
}
