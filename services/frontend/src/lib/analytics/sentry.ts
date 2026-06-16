import { browser } from '$app/environment';

let sentryLoaded = false;

export async function initSentry(): Promise<void> {
	if (!browser) return;

	const dsn = import.meta.env.VITE_SENTRY_DSN;
	if (!dsn) return;

	try {
		const Sentry = await import('@sentry/sveltekit');

		Sentry.init({
			dsn,
			environment: import.meta.env.MODE,
			release: `matchup@${import.meta.env.VITE_APP_VERSION ?? 'dev'}`,
			tracesSampleRate: 0.1,
			beforeSend(event) {
				if (event.request?.headers) {
					delete event.request.headers['Authorization'];
					delete event.request.headers['Cookie'];
				}
				if (event.request) {
					delete event.request.cookies;
				}
				return event;
			}
		});

		sentryLoaded = true;
	} catch (e) {
		console.warn('[Sentry] Failed to initialise:', e);
	}
}

export async function setSentryUser(userId: string): Promise<void> {
	if (!browser || !sentryLoaded) return;
	const Sentry = await import('@sentry/sveltekit');
	Sentry.setUser({ id: userId });
}

export async function clearSentryUser(): Promise<void> {
	if (!browser || !sentryLoaded) return;
	const Sentry = await import('@sentry/sveltekit');
	Sentry.setUser(null);
}
