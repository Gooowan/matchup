<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { unreadStore } from '$stores/unread.svelte';
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	type ChatTab = 'chats' | 'business' | 'marketplace';
	let activeTab = $state<ChatTab>('chats');

	// Club picker state
	interface ClubItem { id: string; slug: string; name: string; city: string; logo?: string; }
	let showClubPicker = $state(false);
	let myClubs = $state<ClubItem[]>([]);
	let clubSearchQuery = $state('');
	let clubSearchResults = $state<ClubItem[]>([]);
	let isSearchingClubs = $state(false);
	let isStartingChat = $state<string | null>(null); // slug being opened
	let clubSearchTimer: ReturnType<typeof setTimeout> | null = null;
	let clubPickerList = $derived(clubSearchQuery.trim() ? clubSearchResults : myClubs);

	async function openClubPicker() {
		showClubPicker = true;
		clubSearchQuery = '';
		clubSearchResults = [];
		if (myClubs.length === 0) {
			try {
				const resp = await authFetch('/me/clubs');
				if (resp.ok) {
					const body = await resp.json();
					myClubs = (body.data ?? []).map((c: any) => ({
						id: c.id, slug: c.slug, name: c.name, city: c.city, logo: c.logo ?? c.metadata?.logo
					}));
				}
			} catch { /* keep empty */ }
		}
	}

	function handleClubSearchInput() {
		if (clubSearchTimer) clearTimeout(clubSearchTimer);
		if (!clubSearchQuery.trim()) { clubSearchResults = []; return; }
		clubSearchTimer = setTimeout(async () => {
			isSearchingClubs = true;
			try {
				const resp = await authFetch(`/clubs?q=${encodeURIComponent(clubSearchQuery)}&limit=15`);
				if (resp.ok) {
					const body = await resp.json();
					clubSearchResults = (body.data ?? []).map((c: any) => ({
						id: c.id, slug: c.slug, name: c.name, city: c.city, logo: c.logo ?? c.metadata?.logo
					}));
				}
			} catch { clubSearchResults = []; }
			finally { isSearchingClubs = false; }
		}, 300);
	}

	async function startClubChat(slug: string) {
		if (isStartingChat) return;
		isStartingChat = slug;
		try {
			const resp = await authFetch(`/clubs/${slug}/chat`, { method: 'POST' });
			if (resp.ok) {
				const body = await resp.json();
				const chatId = body.data?.chat_id ?? body.chat_id;
				if (chatId) {
					showClubPicker = false;
					goto(`/chats/${chatId}`);
					return;
				}
			}
			toast.error($t('chats.club_chat_error'));
		} catch {
			toast.error($t('chats.club_chat_error'));
		} finally {
			isStartingChat = null;
		}
	}

	interface ChatItem {
		id: string;
		name: string;
		avatarUrl?: string;
		lastMessage: string;
		timestamp: string;
		unread: boolean;
		online?: boolean;
		productThumb?: string;
		isBusiness?: boolean;
	}

	let regularChats = $state<ChatItem[]>([]);
	let businessChats = $state<ChatItem[]>([]);

	let chats = $derived(activeTab === 'business' ? businessChats : regularChats);

	onMount(async () => {
		try {
			const resp = await authFetch('/chats');
			if (resp.ok) {
				const response = await resp.json();
				if (response.data && Array.isArray(response.data)) {
					const raw: any[] = response.data;
					// Sync unread badge from server data
					unreadStore.syncFromChats(raw);

					const all: ChatItem[] = raw.map((c: any) => ({
						id: c.id,
						// Club chats are branded with the club; DMs use the peer's name.
						name:
							c.other_user?.profile_data?.first_name ??
							c.club?.name ??
							c.other_user_id ??
							'Match',
						avatarUrl: c.other_user?.profile_data?.avatar ?? c.club?.logo,
						lastMessage: c.last_message?.content ?? '',
						timestamp: c.last_message?.created_at
							? new Date(c.last_message.created_at).toLocaleTimeString('uk', {
									hour: '2-digit',
									minute: '2-digit',
									hour12: false
								})
							: '',
						unread: c.unread_count > 0,
						online: false,
						// Accept both is_club_chat (current) and is_club_owner (legacy alias).
						isBusiness: !!(c.is_club_chat ?? c.is_club_owner)
					}));
					businessChats = all.filter((c: any) => c.isBusiness);
					regularChats = all.filter((c: any) => !c.isBusiness);
				}
			}
		} catch {
			// keep empty state on error
		}
	});
</script>

<div style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch; background: #151517;">
	<!-- Status bar spacer -->
	<div class="pt-safe"></div>

	<!-- Search bar + compose button -->
	<div class="mx-4 mt-2 flex items-center gap-2">
		<div class="glass-pill flex flex-1 items-center gap-3 px-4" style="height: 38px;">
			<i class="fi fi-rr-search" style="font-size: 20px; line-height: 1; color: #9c9c9c; flex-shrink: 0;"></i>
			<span class="text-[14px] font-semibold" style="color: #9c9c9c;">{$t('chats.search_placeholder')}</span>
		</div>
		<button
			class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
			onclick={openClubPicker}
			aria-label="Написати клубу"
		>
			<i class="fi fi-rr-pencil" style="font-size: 16px; line-height: 1; color: white;"></i>
		</button>
	</div>

	<!-- Tab switcher -->
	<div class="glass-pill mx-4 mt-3 flex items-center" style="height: 36px; padding: 4px;">
		<button
			class="flex flex-1 items-center justify-center rounded-[20px] text-[13px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'chats' ? 'white' : 'transparent'}; color: {activeTab === 'chats' ? '#3a3a3a' : 'white'};"
			onclick={() => (activeTab = 'chats')}
		>
			{$t('chats.tab_chats')}
		</button>
		<button
			class="flex flex-1 items-center justify-center rounded-[20px] text-[13px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'business' ? 'white' : 'transparent'}; color: {activeTab === 'business' ? '#3a3a3a' : 'white'};"
			onclick={() => (activeTab = 'business')}
		>
			{$t('chats.tab_business')}
		</button>
		<button
			class="flex flex-1 items-center justify-center rounded-[20px] text-[13px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'marketplace' ? 'white' : 'transparent'}; color: {activeTab === 'marketplace' ? '#3a3a3a' : 'white'};"
			onclick={() => (activeTab = 'marketplace')}
		>
			{$t('chats.tab_market')}
		</button>
	</div>

	<!-- Chat list / Business / Marketplace -->
	{#if activeTab === 'marketplace'}
		<div class="flex flex-1 flex-col items-center justify-center gap-4">
			<i class="fi fi-rr-shopping-bag" style="font-size: 48px; color: #313131;"></i>
			<p class="text-[16px] font-semibold" style="color: #696969;">{$t('marketplace.title')}</p>
			<p class="text-[13px] font-medium" style="color: #484848;">{$t('chats.coming_soon')}</p>
		</div>
	{:else}
		<div class="mt-6 flex flex-col px-4 pb-[100px]" style="gap: 24px;">
			{#each chats as chat}
				<button class="flex items-center gap-4 text-left" onclick={() => goto(`/chats/${chat.id}`)}>
					<!-- Avatar -->
					<div class="relative flex-shrink-0">
						<div
							class="relative overflow-hidden rounded-full"
							style="width: 56px; height: 56px; border: 1px solid #313131;"
						>
							{#if chat.avatarUrl}
								<img src={chat.avatarUrl} alt={chat.name} loading="lazy" decoding="async" class="h-full w-full object-cover" />
							{:else}
								<div
									class="flex h-full w-full items-center justify-center"
									style="background: #2c2b30;"
								>
									<i class="fi fi-rr-user text-white" style="font-size: 24px;"></i>
								</div>
							{/if}
							{#if chat.online}
								<div
									class="absolute right-0 top-0 h-2.5 w-2.5 rounded-full border-2"
									style="background: #22c55e; border-color: #151517;"
								></div>
							{/if}
						</div>
					</div>

					<!-- Content -->
					<div class="flex min-w-0 flex-1 flex-col" style="gap: 4px;">
						<div class="flex items-center justify-between">
							<span class="text-[14px] font-semibold" style="color: #e1e1e1;">{chat.name}</span>
							<span class="text-[12px] font-normal" style="color: #e1e1e1;">{chat.timestamp}</span>
						</div>
						<div class="flex items-center gap-1">
							<i class="fi fi-rr-check" style="font-size: 12px; color: #7c7c7c; flex-shrink: 0;"></i>
							<span
								class="truncate text-[14px] font-normal"
								style="color: {chat.unread ? '#e1e1e1' : '#898484'};"
							>{chat.lastMessage}</span>
						</div>
					</div>
				</button>
			{/each}

		{#if chats.length === 0}
			<div class="flex flex-col items-center justify-center py-16">
				{#if activeTab === 'business'}
					<i class="fi fi-rr-store-alt" style="font-size: 48px; color: #313131;"></i>
					<p class="mt-4 text-[16px] font-semibold" style="color: #696969;">{$t('chats.empty')}</p>
					<button
						onclick={openClubPicker}
						class="mt-4 rounded-[50px] px-5 py-2.5 text-[13px] font-semibold text-white"
						style="background: #8984da;"
					>
						<i class="fi fi-rr-store-alt mr-1.5" style="font-size: 13px; line-height: 1; vertical-align: middle;"></i>
						{$t('chats.write_to_club')}
					</button>
				{:else}
					<i class="fi fi-rr-comment-heart" style="font-size: 48px; color: #313131;"></i>
					<p class="mt-4 text-[16px] font-semibold" style="color: #696969;">{$t('chats.empty')}</p>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<!-- Club picker bottom sheet -->
{#if showClubPicker}
	<div
		class="fixed inset-0 z-50 flex flex-col justify-end"
		style="background: rgba(0,0,0,0.5);"
		role="dialog"
		aria-modal="true"
	>
		<button class="absolute inset-0" onclick={() => (showClubPicker = false)} aria-label="Закрити"></button>
		<div
			class="relative flex flex-col rounded-t-[24px] pb-safe"
			style="background: #1e1e20; max-height: 80dvh;"
		>
			<!-- Handle -->
			<div class="flex justify-center pt-3 pb-1">
				<div class="h-1 w-10 rounded-full" style="background: #3a3a3a;"></div>
			</div>
			<div class="px-4 pb-2 pt-1">
				<h2 class="text-[16px] font-black" style="color: #e1e1e1;">{$t('chats.write_to_club')}</h2>
			</div>

			<!-- Search input -->
			<div class="px-4 pb-3">
				<div class="flex items-center gap-3 rounded-[12px] px-3" style="background: #2c2b30; height: 40px;">
					<i class="fi fi-rr-search" style="font-size: 15px; line-height: 1; color: #696969; flex-shrink: 0;"></i>
					<input
						type="search"
						placeholder={$t('chats.club_search_placeholder')}
						bind:value={clubSearchQuery}
						oninput={handleClubSearchInput}
						class="flex-1 bg-transparent text-[14px] font-medium outline-none"
						style="color: #e1e1e1;"
					/>
					{#if isSearchingClubs}
						<div class="h-4 w-4 animate-spin rounded-full border-2 border-white/20" style="border-top-color: #8984da; flex-shrink: 0;"></div>
					{/if}
				</div>
			</div>

			<!-- Club list -->
			<div class="flex-1 overflow-y-auto px-4 pb-6" style="-webkit-overflow-scrolling: touch;">
				{#if !clubSearchQuery.trim() && myClubs.length > 0}
					<p class="mb-2 text-[11px] font-semibold uppercase tracking-wider" style="color: #696969;">{$t('chats.my_clubs')}</p>
				{/if}

				{#each clubPickerList as club}
					<button
						class="flex w-full items-center gap-3 rounded-[14px] p-3 text-left transition-opacity"
						style="background: #2c2b30; margin-bottom: 8px; opacity: {isStartingChat === club.slug ? 0.6 : 1};"
						disabled={!!isStartingChat}
						onclick={() => startClubChat(club.slug)}
					>
						<div
							class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-full"
							style="background: #3a3a3a;"
						>
							{#if club.logo}
								<img src={club.logo} alt={club.name} class="h-full w-full object-cover" />
							{:else}
								<i class="fi fi-rr-store-alt" style="font-size: 18px; color: #696969; line-height: 1;"></i>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<p class="truncate text-[14px] font-semibold" style="color: #e1e1e1;">{club.name}</p>
							{#if club.city}
								<p class="text-[12px] font-medium" style="color: #696969;">{club.city}</p>
							{/if}
						</div>
						{#if isStartingChat === club.slug}
							<div class="h-4 w-4 animate-spin rounded-full border-2 border-white/20" style="border-top-color: #8984da;"></div>
						{:else}
							<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #3a3a3a; line-height: 1;"></i>
						{/if}
					</button>
				{/each}

				{#if clubPickerList.length === 0 && !isSearchingClubs}
					<div class="flex flex-col items-center justify-center py-10">
						<i class="fi fi-rr-store-alt" style="font-size: 32px; color: #313131;"></i>
						<p class="mt-3 text-[14px] font-medium" style="color: #696969;">
							{clubSearchQuery.trim() ? $t('chats.club_not_found') : $t('chats.no_clubs_joined')}
						</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
</div>
