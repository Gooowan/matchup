<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { authStore } from '$stores/auth.svelte';
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';

	let { children } = $props();

	let hideNav = $derived(/^\/chats\/[^/]+/.test(page.url.pathname));

	onMount(async () => {
		if (!authStore.isAuthenticated) {
			const ok = await authStore.checkAuth();
			if (!ok) {
				await goto('/login');
				return;
			}
		}
		// Redirect to onboarding if no dance profile (no first_name set)
		const user = authStore.user as Record<string, unknown> | null;
		const pd = user?.profile_data as Record<string, string> | undefined;
		if (!pd?.first_name && !page.url.pathname.startsWith('/onboarding')) {
			await goto('/onboarding');
		}
	});
</script>

{@render children?.()}
{#if !hideNav}
	<BottomNav />
{/if}
