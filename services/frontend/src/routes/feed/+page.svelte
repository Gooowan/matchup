<script lang="ts">
	import SwipeCard, { type DancerProfile } from '$lib/components/matchup/SwipeCard.svelte';
	import MatchPopup from '$lib/components/matchup/MatchPopup.svelte';
	import FilterSheet from '$lib/components/matchup/FilterSheet.svelte';
	import CardActionMenu from '$lib/components/matchup/CardActionMenu.svelte';
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';
	import { unreadStore } from '$stores/unread.svelte';

	let profiles = $state<DancerProfile[]>([
		{
			id: '1',
			name: 'Maria',
			age: 24,
			photoUrl: 'https://images.unsplash.com/photo-1518611012118-696072aa579a?w=800&fit=crop',
			tags: ['Ballroom', 'Pro', '175 cm'],
			location: 'New York, NY',
			school: 'Ballroom Dance Academy',
			goals: 'Competition-ready, Open to relocation'
		},
		{
			id: '2',
			name: 'Alex',
			age: 27,
			photoUrl: 'https://images.unsplash.com/photo-1547380236-48c58a1cd64f?w=800&fit=crop',
			tags: ['Latin', 'Leader', '182 cm'],
			location: 'New York, NY',
			school: 'Latin Dance Studio',
			goals: 'Social dancing, Competitions'
		},
		{
			id: '3',
			name: 'Sofia',
			age: 22,
			photoUrl: 'https://images.unsplash.com/photo-1508700929628-666bc8bd84ea?w=800&fit=crop',
			tags: ['Salsa', 'Follower', '168 cm'],
			location: 'Brooklyn, NY',
			goals: 'Social dancing'
		},
		{
			id: '4',
			name: 'Dmitri',
			age: 29,
			photoUrl: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=800&fit=crop',
			tags: ['Bachata', 'Leader', '185 cm'],
			location: 'Manhattan, NY',
			school: 'NYC Dance Academy',
			goals: 'Competitions, Teaching'
		},
		{
			id: '5',
			name: 'Elena',
			age: 25,
			photoUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&fit=crop',
			tags: ['Contemporary', 'Follower', '170 cm'],
			location: 'Queens, NY',
			goals: 'Stage performances'
		},
		{
			id: '6',
			name: 'Carlos',
			age: 31,
			photoUrl: 'https://images.unsplash.com/photo-1531746020798-e6953c6e8e04?w=800&fit=crop',
			tags: ['Tango', 'Leader', '178 cm'],
			location: 'Bronx, NY',
			school: 'Tango Buenos Aires Studio',
			goals: 'Milonga, Social dancing'
		},
		{
			id: '7',
			name: 'Anastasia',
			age: 23,
			photoUrl: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=800&fit=crop',
			tags: ['Jazz', 'Follower', '165 cm'],
			location: 'New York, NY',
			goals: 'Broadway-style, Competitions'
		},
		{
			id: '8',
			name: 'Miguel',
			age: 26,
			photoUrl: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=800&fit=crop',
			tags: ['Salsa', 'Leader', '180 cm'],
			location: 'Staten Island, NY',
			school: 'Salsa Caliente Studio',
			goals: 'Social dancing, Performances'
		},
		{
			id: '9',
			name: 'Natalia',
			age: 28,
			photoUrl: 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=800&fit=crop',
			tags: ['Swing', 'Follower', '162 cm'],
			location: 'New York, NY',
			goals: 'Lindy Hop, Competitions'
		},
		{
			id: '10',
			name: 'Lucas',
			age: 24,
			photoUrl: 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=800&fit=crop',
			tags: ['Ballroom', 'Leader', '183 cm'],
			location: 'New York, NY',
			school: 'Imperial Dance Academy',
			goals: 'Amateur competitions, Latin'
		},
		{
			id: '11',
			name: 'Valentina',
			age: 21,
			photoUrl: 'https://images.unsplash.com/photo-1502323703975-b32a7b4be31f?w=800&fit=crop',
			tags: ['Latin', 'Follower', '167 cm'],
			location: 'New York, NY',
			goals: 'Professional path, Competitions'
		}
	]);

	let showMatch = $state(false);
	let matchedProfile = $state<DancerProfile | null>(null);
	let showFilter = $state(false);
	let menuProfileId = $state<string | null>(null);
	let hasNotification = $state(true);

	function handleLike(id: string) {
		const matched = Math.random() > 0.5;
		if (matched) {
			matchedProfile = profiles.find((p) => p.id === id) ?? null;
			showMatch = true;
			unreadStore.increment();
		}
		removeTopCard();
	}

	function handlePass(_id: string) {
		removeTopCard();
	}

	function removeTopCard() {
		profiles = profiles.slice(1);
	}

	function handleChatNow() {
		showMatch = false;
	}
</script>

<!-- Full-screen wrapper -->
<div class="relative h-[100dvh] overflow-hidden bg-black">
	{#if profiles.length > 0}
		<!-- Card stack: background cards + top interactive card -->
		{#each profiles.slice(0, 3).reverse() as profile, i}
			{@const stackLen = Math.min(profiles.length, 3)}
			{@const stackPos = stackLen - 1 - i}
			{@const isTop = stackPos === 0}
			<div
				style="
					position: absolute;
					inset: 0;
					transform: scale({1 - stackPos * 0.025}) translateY({stackPos * 10}px);
					z-index: {10 + stackLen - stackPos};
					pointer-events: {isTop ? 'auto' : 'none'};
				"
			>
				{#if isTop}
					<SwipeCard
						{profile}
						{isTop}
						onlike={handleLike}
						onpass={handlePass}
						onviewprofile={(id) => console.log('view profile', id)}
						onmenu={(id) => (menuProfileId = id)}
						zIndex={0}
						fullScreen={true}
					/>
				{:else}
					<img
						src={profile.photoUrl}
						alt={profile.name}
						class="h-full w-full object-cover"
						draggable="false"
						style="filter: blur(1px) brightness(0.85);"
					/>
				{/if}
			</div>
		{/each}

		<!-- Top gradient overlay -->
		<div
			class="pointer-events-none absolute right-0 left-0 top-0 z-30 h-[130px]"
			style="background: linear-gradient(180deg, rgba(0,0,0,0.45) 0%, rgba(0,0,0,0) 100%);"
		></div>

		<!-- Bottom gradient overlay -->
		<div
			class="pointer-events-none absolute right-0 bottom-0 left-0 z-30 h-[295px]"
			style="background: linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,0.88) 100%);"
		></div>
	{:else}
		<!-- Empty state: centered on #dae1eb bg -->
		<div
			class="absolute inset-0 flex flex-col items-center justify-center"
			style="background: #dae1eb;"
		>
			<i class="fi fi-rr-heart" style="font-size: 64px; color: #aeb4bc;"></i>
			<p class="mt-4 text-[18px] font-bold" style="color: #7d7d7d;">No more profiles</p>
			<p class="mt-1 text-[14px] font-medium" style="color: #b1b1b1;">Check back later</p>
		</div>
	{/if}

	<!-- Glassmorphic header (z-40, over all gradients) -->
	<div
		class="absolute right-0 left-0 z-40 flex items-center justify-between px-4"
		style="top: max(env(safe-area-inset-top), 8px); height: 54px;"
	>
		<!-- Filter pill button -->
		<button
			class="glass-pill flex items-center gap-2 px-3"
			style="height: 38px;"
			onclick={() => (showFilter = true)}
			aria-label="Filter"
		>
			<i class="fi fi-rr-settings-sliders" style="font-size: 16px; line-height: 1; color: white;"></i>
			<span class="text-[13px] font-semibold text-white">Filter</span>
		</button>

		<div class="flex items-center gap-2">
			<!-- Bell button with notification dot -->
			<button
				class="glass-pill relative flex h-[38px] w-[38px] items-center justify-center"
				aria-label="Notifications"
			>
				<i class="fi fi-rr-bell" style="font-size: 16px; line-height: 1; color: white;"></i>
				{#if hasNotification}
					<div
						class="absolute top-[8px] right-[8px] h-[8px] w-[8px] rounded-full"
						style="background: #e74c3c;"
					></div>
				{/if}
			</button>
			<!-- Three-dot menu button -->
			{#if profiles.length > 0}
				<button
					class="glass-pill flex h-[38px] w-[38px] items-center justify-center"
					aria-label="More options"
					onclick={() => (menuProfileId = profiles[0].id)}
				>
					<i class="fi fi-rr-menu-dots" style="font-size: 16px; line-height: 1; color: white;"></i>
				</button>
			{/if}
		</div>
	</div>

	<!-- Bottom nav -->
	<BottomNav active="feed" />
</div>

<!-- Match popup (outside main div, no z-index conflict) -->
<MatchPopup
	open={showMatch}
	theirProfile={matchedProfile}
	onchat={handleChatNow}
	onclose={() => (showMatch = false)}
/>

<!-- Filter sheet -->
<FilterSheet
	open={showFilter}
	onclose={() => (showFilter = false)}
	onclear={() => console.log('clear filters')}
/>

<!-- 3-dot action menu -->
<CardActionMenu
	open={!!menuProfileId}
	onclose={() => (menuProfileId = null)}
	onclear={() => (profiles = [])}
	onhide={() => removeTopCard()}
	onreport={() => console.log('report', menuProfileId)}
/>
