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
			if (!ok) await goto('/login');
		}
	});
</script>

{@render children?.()}
{#if !hideNav}
	<BottomNav />
{/if}
