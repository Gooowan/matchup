<script lang="ts">
	import { page } from '$app/state';
	import { unreadStore } from '$stores/unread.svelte';
	import { authStore } from '$stores/auth.svelte';
	import { isRestrictedAccountType } from '$lib/types/accountType';
	import { t } from '$lib/locale';

	type Tab = 'map' | 'marketplace' | 'feed' | 'chats' | 'settings';

	interface Props {
		active?: Tab;
	}

	let { active }: Props = $props();

	let currentPath = $derived(page.url.pathname);
	let accountType = $derived(
		(authStore.user?.profile_data?.account_type as string | undefined) ?? null
	);
	let restricted = $derived(isRestrictedAccountType(accountType));

	function isActive(tab: Tab): boolean {
		if (active !== undefined) return active === tab;
		if (tab === 'map') return currentPath.startsWith('/map');
		if (tab === 'marketplace') return currentPath.startsWith('/marketplace');
		if (tab === 'feed') return currentPath === '/' || currentPath.startsWith('/feed');
		if (tab === 'chats') return currentPath.startsWith('/chats');
		if (tab === 'settings') return currentPath.startsWith('/settings');
		return false;
	}
</script>

<nav
	class="fixed right-0 left-0 z-50 px-4"
	style="bottom: max(env(safe-area-inset-bottom), 8px);"
>
	<!-- 60px container: pill is 52px centered, home button fills 60px -->
	<div class="relative mb-2 h-[60px]">
		<!-- Glassmorphic pill background, 52px centered -->
		<div
			class="glass-pill absolute top-1/2 right-0 left-0 h-[52px] -translate-y-1/2 rounded-[50px]"
		></div>

		{#if restricted}
			<!-- Restricted nav: 3 icons evenly spaced (Marketplace, Chats, Settings) -->
			<div class="absolute inset-0 flex items-center justify-around px-6">
				<a href="/marketplace" class="flex items-center justify-center" aria-label={$t('nav.marketplace')}>
					<i
						class="fi {isActive('marketplace') ? 'fi-sr-shopping-bag' : 'fi-rr-shopping-bag'} leading-none"
						style="font-size: 22px; color: {isActive('marketplace') ? '#8984da' : 'white'};"
					></i>
				</a>

				<a href="/chats" class="relative flex items-center justify-center" aria-label={$t('nav.chats')}>
					<i
						class="fi {isActive('chats') ? 'fi-sr-comment-heart' : 'fi-rr-comment-heart'} leading-none"
						style="font-size: 22px; color: {isActive('chats') ? '#8984da' : 'white'};"
					></i>
					{#if unreadStore.count > 0}
						<div
							class="absolute -top-1 -right-1.5 flex h-[14px] min-w-[14px] items-center justify-center rounded-full bg-red-500 px-[3px] text-[9px] font-bold text-white"
						>
							{unreadStore.count > 9 ? '9+' : unreadStore.count}
						</div>
					{/if}
				</a>

				<a href="/settings" class="flex items-center justify-center" aria-label={$t('nav.settings')}>
					<i
						class="fi {isActive('settings') ? 'fi-sr-settings' : 'fi-rr-settings'} leading-none"
						style="font-size: 22px; color: {isActive('settings') ? '#8984da' : 'white'};"
					></i>
				</a>
			</div>
		{:else}
			<!-- Full nav: Map, Marketplace, center Feed, Chats, Settings -->
			<div class="absolute inset-0 flex items-center justify-between px-4">
				<a href="/map" class="flex items-center justify-center" aria-label={$t('nav.map')}>
					<i
						class="fi {isActive('map') ? 'fi-sr-marker' : 'fi-rr-marker'} leading-none"
						style="font-size: 20px; color: {isActive('map') ? '#8984da' : 'white'};"
					></i>
				</a>

				<a href="/marketplace" class="flex items-center justify-center" aria-label={$t('nav.marketplace')}>
					<i
						class="fi {isActive('marketplace') ? 'fi-sr-shopping-bag' : 'fi-rr-shopping-bag'} leading-none"
						style="font-size: 20px; color: {isActive('marketplace') ? '#8984da' : 'white'};"
					></i>
				</a>

				<a
					href="/feed"
					class="relative flex h-[60px] w-[60px] flex-shrink-0 items-center justify-center rounded-full"
					style="background: linear-gradient(135deg, #8984da 0%, #b4b0e8 100%); box-shadow: 0 2px 16px rgba(137,132,218,0.5);"
					aria-label={$t('nav.matchup')}
				>
					<img src="/match_icon.svg" alt="MatchUp" class="h-7 w-7" />
				</a>

				<a href="/chats" class="relative flex items-center justify-center" aria-label={$t('nav.chats')}>
					<i
						class="fi {isActive('chats') ? 'fi-sr-comment-heart' : 'fi-rr-comment-heart'} leading-none"
						style="font-size: 20px; color: {isActive('chats') ? '#8984da' : 'white'};"
					></i>
					{#if unreadStore.count > 0}
						<div
							class="absolute -top-1 -right-1.5 flex h-[14px] min-w-[14px] items-center justify-center rounded-full bg-red-500 px-[3px] text-[9px] font-bold text-white"
						>
							{unreadStore.count > 9 ? '9+' : unreadStore.count}
						</div>
					{/if}
				</a>

				<a href="/settings" class="flex items-center justify-center" aria-label={$t('nav.settings')}>
					<i
						class="fi {isActive('settings') ? 'fi-sr-settings' : 'fi-rr-settings'} leading-none"
						style="font-size: 20px; color: {isActive('settings') ? '#8984da' : 'white'};"
					></i>
				</a>
			</div>
		{/if}
	</div>
</nav>
