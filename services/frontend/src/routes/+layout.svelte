<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import { initSentry } from '$lib/analytics/sentry';
	import { initPostHog, capturePageView } from '$lib/analytics/posthog';

	let { children } = $props();

	onMount(() => {
		if (!browser) return;

		if (localStorage.getItem('mu-theme') === 'dark') {
			document.documentElement.classList.add('dark');
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
