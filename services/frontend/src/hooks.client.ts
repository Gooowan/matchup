import { handleErrorWithSentry } from '@sentry/sveltekit';

// Catches all unhandled SvelteKit errors and sends them to Sentry.
// When VITE_SENTRY_DSN is not set, handleErrorWithSentry is a no-op pass-through.
export const handleError = handleErrorWithSentry();
