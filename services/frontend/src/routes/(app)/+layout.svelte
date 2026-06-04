<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { authStore } from '$stores/auth.svelte';
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';
	import { isRestrictedAccountType } from '$lib/types/accountType';
	import { registerPushNotifications } from '$lib/utils/pushNotifications';

	let { children } = $props();

	let hideNav = $derived(/^\/chats\/[^/]+/.test(page.url.pathname));

	// Routes that restricted (trainer/club) accounts are allowed to visit.
	const RESTRICTED_ALLOWED_PREFIXES = ['/marketplace', '/chats', '/settings', '/business'];

	onMount(async () => {
		if (!authStore.isAuthenticated) {
			const ok = await authStore.checkAuth();
			if (!ok) {
				await goto('/login');
				return;
			}
		}
		const user = authStore.user as Record<string, unknown> | null;
		const pd = user?.profile_data as Record<string, string> | undefined;
		if (!pd?.account_type && !page.url.pathname.startsWith('/onboarding')) {
			await goto('/onboarding');
			return;
		}
		// Trainer/Club accounts can only view marketplace, chats, settings, and business.
		if (
			isRestrictedAccountType(pd?.account_type) &&
			!RESTRICTED_ALLOWED_PREFIXES.some((p) => page.url.pathname.startsWith(p))
		) {
			await goto('/settings');
			return;
		}
		registerPushNotifications();
	});

	// Reactive guard: enforced on every client-side navigation, not just initial mount.
	$effect(() => {
		const path = page.url.pathname;
		const pd = (authStore.user as Record<string, unknown> | null)
			?.profile_data as Record<string, string> | undefined;
		if (
			authStore.isAuthenticated &&
			pd?.account_type &&
			isRestrictedAccountType(pd?.account_type) &&
			!RESTRICTED_ALLOWED_PREFIXES.some((p) => path.startsWith(p))
		) {
			goto('/settings');
		}
	});
</script>

{@render children?.()}
{#if !hideNav}
	<div class="nav-fade"></div>
	<BottomNav />
{/if}
