<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import SwipeCard, { type DancerProfile } from '$lib/components/matchup/SwipeCard.svelte';
	import MatchPopup from '$lib/components/matchup/MatchPopup.svelte';
	import FilterSheet, { type FilterState } from '$lib/components/matchup/FilterSheet.svelte';
	import CardActionMenu from '$lib/components/matchup/CardActionMenu.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { unreadStore } from '$stores/unread.svelte';
	import { filterStore } from '$stores/filters.svelte';
	import { captureSwipe, captureMatch } from '$lib/analytics/posthog';
	import toast from 'svelte-french-toast';
	import { t } from '$lib/locale';

	interface FeedCandidate {
		user_id: string;
		dance_styles: string[];
		gender: string;
		birth_date?: string;
		height_cm?: number;
		goal: string;
		program: string;
		categories: string[];
		country?: string;
		city?: string;
		metadata: Record<string, unknown>;
		profile_data: Record<string, unknown>;
		distance_km: number;
		source?: string;
	}

	let profiles = $state<DancerProfile[]>([]);
	let isLoading = $state(true);
	let isLoadingMore = $state(false);
	let exhausted = $state(false);

	let showMatch = $state(false);
	let matchedProfile = $state<DancerProfile | null>(null);
	let matchedChatId = $state<string | null>(null);
	let showFilter = $state(false);
	let menuProfileId = $state<string | null>(null);
	let hasNotification = $state(false);
	let noProfile = $state(false);

	function calcAge(birthDate: string): number {
		const dob = new Date(birthDate);
		const diff = Date.now() - dob.getTime();
		return Math.floor(diff / (365.25 * 24 * 3600 * 1000));
	}

	function mapCandidate(c: FeedCandidate): DancerProfile {
		const pd = c.profile_data as Record<string, string> ?? {};
		const tags: string[] = [];
		const style = (c.categories?.[0] ?? c.dance_styles?.[0] ?? '').trim();
		if (style) tags.push(style);
		if (c.program && c.program !== 'standard') tags.push(c.program);
		if (c.height_cm) tags.push(`${c.height_cm} cm`);
		return {
			id: c.user_id,
			name: pd.first_name ?? $t('feed.dancer_fallback'),
			age: c.birth_date ? calcAge(c.birth_date) : 0,
			photoUrl: (pd.avatar as string) ?? '',
			tags,
			location: [c.city, c.country].filter(Boolean).join(', '),
			school: undefined,
			goals: c.goal
		};
	}

	async function loadFeed(replace = false) {
		if (exhausted && !replace) return;
		if (isLoadingMore) return;
		isLoadingMore = !replace;
		isLoading = replace;

		try {
			const resp = await authFetch('/matchup/feed?limit=20');
			if (resp.ok) {
				const body = await resp.json();
				const candidates: FeedCandidate[] = body.data?.candidates ?? [];
				const mapped = candidates.map(mapCandidate);
				if (replace) {
					profiles = mapped;
				} else {
					profiles = [...profiles, ...mapped];
				}
				if (mapped.length === 0) exhausted = true;
			}
		} catch {
			// keep existing cards
		} finally {
			isLoading = false;
			isLoadingMore = false;
		}
	}

	onMount(async () => {
		await filterStore.load();
		try {
			const resp = await authFetch('/me/profile');
			if (resp.status === 404) {
				noProfile = true;
			}
		} catch {
			// leave noProfile as false
		}
		loadFeed(true);
	});

	async function handleLike(id: string) {
		captureSwipe('LIKE', 'feed');
		// Capture before removing the card so we have the profile for the match popup
		const liked = profiles.find((p) => p.id === id) ?? null;
		removeTopCard();
		try {
			const resp = await authFetch('/matchup/swipe', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ target_user_id: id, action: 'LIKE', source: 'feed' })
			});
			if (resp.ok) {
				const body = await resp.json();
				if (body.data?.is_mutual_match) {
					matchedProfile = liked;
					matchedChatId = body.data?.chat_id ?? null;
					showMatch = true;
					captureMatch();
					unreadStore.increment();
				}
			}
		} catch {
			// silently ignore
		}
		if (profiles.length < 3 && !exhausted) loadFeed();
	}

	async function handlePass(id: string) {
		captureSwipe('PASS', 'feed');
		removeTopCard();
		authFetch('/matchup/swipe', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ target_user_id: id, action: 'PASS', source: 'feed' })
		}).catch(() => {});
		if (profiles.length < 3 && !exhausted) loadFeed();
	}

	async function handleHide(id: string) {
		profiles = profiles.filter((p) => p.id !== id);
		authFetch('/matchup/hide', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ target_user_id: id, reason: 'hide' })
		}).catch(() => {});
	}

	function removeTopCard() {
		profiles = profiles.slice(1);
	}

	function handleChatNow() {
		showMatch = false;
		if (matchedChatId) goto(`/chats/${matchedChatId}`);
	}

	function handleApplyFilters(filters: FilterState) {
		filterStore.apply(filters);
		exhausted = false;
		loadFeed(true);
	}

	async function handleReport(id: string) {
		try {
			await authFetch(`/users/${id}/report`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ reason: 'inappropriate' })
			});
			toast.success($t('feed.toast_report_sent'));
		} catch {
			toast.error($t('feed.toast_error'));
		}
	}
</script>

<!-- Full-screen wrapper -->
<div class="relative h-[100dvh] overflow-hidden bg-black">
	{#if isLoading}
		<div class="absolute inset-0 flex items-center justify-center">
			<div
				class="h-10 w-10 animate-spin rounded-full border-4 border-white/20"
				style="border-top-color: white;"
			></div>
		</div>
	{:else if profiles.length > 0}
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
						onviewprofile={(id) => goto(`/profiles/${id}`)}
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
		<!-- Empty state -->
		<div class="mu-screen absolute inset-0 flex flex-col items-center justify-center px-8">
			<i class="fi fi-rr-heart" style="font-size: 64px; color: #aeb4bc;"></i>
			{#if noProfile}
				<p class="mt-4 text-[20px] font-black text-center" style="color: #7d7d7d;">Налаштуй профіль щоб знаходити партнерів</p>
				<p class="mt-2 text-[14px] font-medium text-center" style="color: #b1b1b1;">Додай дані в профіль і вкажи місто, щоб ми могли знайти людей поруч.</p>
				<a
					href="/settings/profile"
					class="mt-6 flex items-center justify-center rounded-[50px] px-6 py-2.5 text-[14px] font-semibold text-white no-underline"
					style="background: #8984da;"
				>Редагувати профіль</a>
			{:else}
				<p class="mt-4 text-[20px] font-black text-center" style="color: #7d7d7d;">{$t('feed.empty_title')}</p>
				<p class="mt-2 text-[14px] font-medium text-center" style="color: #b1b1b1;">{$t('feed.empty_subtitle')}</p>
				<button
					onclick={() => loadFeed(true)}
					class="mt-6 rounded-[50px] px-6 py-2.5 text-[14px] font-semibold text-white"
					style="background: #8984da;"
				>{$t('feed.refresh')}</button>
			{/if}
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
			<span class="text-[13px] font-semibold text-white">{$t('feed.filter')}</span>
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
</div>

<!-- Match popup -->
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
	onclear={() => { filterStore.clear(); exhausted = false; loadFeed(true); }}
	onapply={handleApplyFilters}
	initialFilters={filterStore.filters}
/>

<!-- 3-dot action menu -->
<CardActionMenu
	open={!!menuProfileId}
	profileId={menuProfileId}
	onclose={() => (menuProfileId = null)}
	onclear={() => { profiles = []; exhausted = false; loadFeed(true); }}
	onhide={() => menuProfileId && handleHide(menuProfileId)}
	onreport={() => menuProfileId && handleReport(menuProfileId)}
/>
