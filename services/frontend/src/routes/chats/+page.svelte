<script lang="ts">
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { unreadStore } from '$stores/unread.svelte';

	type ChatTab = 'chats' | 'marketplace';
	let activeTab = $state<ChatTab>('chats');

	interface ChatItem {
		id: string;
		name: string;
		avatarUrl?: string;
		lastMessage: string;
		timestamp: string;
		unread: boolean;
		online?: boolean;
		productThumb?: string;
	}

	let regularChats = $state<ChatItem[]>([
		{
			id: 'c1',
			name: 'Maria',
			avatarUrl: 'https://images.unsplash.com/photo-1518611012118-696072aa579a?w=80',
			lastMessage: 'Hey! When are you free to practice?',
			timestamp: '18:24',
			unread: true,
			online: true
		},
		{
			id: 'c2',
			name: 'Alex',
			avatarUrl: 'https://images.unsplash.com/photo-1547380236-48c58a1cd64f?w=80',
			lastMessage: 'Great session today!',
			timestamp: '02 Jun',
			unread: false
		},
		{
			id: 'c3',
			name: 'Sofia',
			lastMessage: 'See you at the competition 👋',
			timestamp: '14.06.25',
			unread: false
		}
	]);

	const marketplaceChats: ChatItem[] = [
		{
			id: 'm1',
			name: 'Golden dancing shoes',
			productThumb: 'https://images.unsplash.com/photo-1543163521-1bf539c55dd2?w=80',
			lastMessage: 'Is this still available?',
			timestamp: '18:24',
			unread: true
		},
		{
			id: 'm2',
			name: 'Dancewear for training',
			productThumb: 'https://images.unsplash.com/photo-1518611012118-696072aa579a?w=80',
			lastMessage: 'Sent you an offer',
			timestamp: '02 Jun',
			unread: false
		}
	];

	let chats = $derived(activeTab === 'chats' ? regularChats : marketplaceChats);

	onMount(async () => {
		unreadStore.reset();

		// Load chats from API
		try {
			const resp = await authFetch('/chats');
			if (resp.ok) {
				const response = await resp.json();
				if (response.data && Array.isArray(response.data)) {
					regularChats = response.data.map((c: any) => ({
						id: c.id,
						name: c.other_user?.profile_data?.first_name ?? c.other_user_id ?? 'Match',
						avatarUrl: c.other_user?.profile_data?.avatar,
						lastMessage: c.last_message?.content ?? '',
						timestamp: c.last_message?.created_at
							? new Date(c.last_message.created_at).toLocaleTimeString('en', {
									hour: '2-digit',
									minute: '2-digit'
								})
							: '',
						unread: c.unread_count > 0,
						online: false
					}));
				}
			}
		} catch {
			// Keep mock data on error
		}
	});
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden" style="background: #151517;">
	<!-- Status bar spacer -->
	<div class="pt-safe"></div>

	<!-- Search bar -->
	<div class="glass-pill mx-4 mt-2 flex items-center gap-3 px-4" style="height: 38px;">
		<i class="fi fi-rr-search" style="font-size: 20px; line-height: 1; color: #9c9c9c; flex-shrink: 0;"></i>
		<span class="text-[14px] font-semibold" style="color: #9c9c9c;">Search</span>
	</div>

	<!-- Tab switcher -->
	<div class="glass-pill mx-4 mt-3 flex items-center" style="height: 36px; padding: 4px;">
		<button
			class="flex flex-1 items-center justify-center rounded-[20px] text-[14px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'chats' ? 'white' : 'transparent'}; color: {activeTab === 'chats' ? '#3a3a3a' : 'white'};"
			onclick={() => (activeTab = 'chats')}
		>
			Chats
		</button>
		<button
			class="flex flex-1 items-center justify-center rounded-[20px] text-[14px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'marketplace' ? 'white' : 'transparent'}; color: {activeTab === 'marketplace' ? '#3a3a3a' : 'white'};"
			onclick={() => (activeTab = 'marketplace')}
		>
			Marketplace
		</button>
	</div>

	<!-- Chat list -->
	<div class="mt-6 flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 24px;">
		{#each chats as chat}
			<button class="flex items-center gap-4 text-left" onclick={() => goto(`/chats/${chat.id}`)}>
				<!-- Avatar / product thumb -->
				<div class="relative flex-shrink-0">
					{#if chat.productThumb}
						<div class="relative overflow-hidden rounded-[10px]" style="width: 52px; height: 56px;">
							<img src={chat.productThumb} alt={chat.name} class="h-full w-full object-cover" />
						</div>
					{:else}
						<div
							class="relative overflow-hidden rounded-full"
							style="width: 56px; height: 56px; border: 1px solid #313131;"
						>
							{#if chat.avatarUrl}
								<img src={chat.avatarUrl} alt={chat.name} class="h-full w-full object-cover" />
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
					{/if}
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
				<i class="fi fi-rr-comment-heart" style="font-size: 48px; color: #313131;"></i>
				<p class="mt-4 text-[16px] font-semibold" style="color: #696969;">No chats yet</p>
				<p class="mt-1 text-[13px] font-medium" style="color: #484848;">Match with someone to start chatting</p>
			</div>
		{/if}
	</div>

	<BottomNav active="chats" />
</div>
