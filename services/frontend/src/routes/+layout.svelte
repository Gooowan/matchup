<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import { initSentry } from '$lib/analytics/sentry';
	import { initPostHog, capturePageView } from '$lib/analytics/posthog';
	import { setLocale, loadTranslations, defaultLocale } from '$lib/locale';

	let { children } = $props();

	onMount(async () => {
		if (!browser) return;

		if (localStorage.getItem('mu-theme') === 'dark') {
			document.documentElement.classList.add('dark');
		}

		// Restore user locale preference from localStorage (client-side SPA source of truth).
		// Falls back to the cookie-based default set by the server at build time.
		const saved = localStorage.getItem('mu.locale') ?? defaultLocale;
		if (saved !== defaultLocale) {
			// Override the prerendered server locale with the user's saved preference.
			await setLocale(saved as 'uk' | 'en');
			await loadTranslations(saved, page.url.pathname);
			// Keep cookie in sync so it matches localStorage.
			document.cookie = `locale=${saved}; max-age=${60 * 60 * 24 * 365}; path=/; SameSite=Lax`;
		}

		// Initialise observability — all calls are no-ops when env vars are absent.
		initSentry();
		initPostHog();
	});

	// Track SvelteKit page navigations for PostHog analytics.
	$effect(() => {
		if (browser && page.url) {
			capturePageView(page.url.pathname);
		}
	});
</script>

<svelte:head>
	<title>MatchUp</title>
</svelte:head>

{@render children?.()}
