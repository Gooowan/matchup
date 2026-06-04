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
	import { authStore } from '$stores/auth.svelte';
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
	let loadError = $state(false);
	// Client-side set of all profile IDs that have been shown or swiped in this session.
	// Prevents the same person reappearing while a swipe request is still in-flight.
	const seenIds = new Set<string>();

	let showMatch = $state(false);
	let matchedProfile = $state<DancerProfile | null>(null);
	let matchedChatId = $state<string | null>(null);
	let showFilter = $state(false);
	let menuProfileId = $state<string | null>(null);
	let hasNotification = $state(false);
	let noProfile = $state(false);
	let userGoal = $state('professional');

	type FeedTab = 'partners' | 'trainers';
	let activeTab = $state<FeedTab>('partners');

	interface TrainerCard {
		user_id: string;
		first_name: string;
		last_name?: string;
		avatar?: string;
		gender: string;
		city?: string;
		bio?: string;
		categories: string[];
	}
	let trainers = $state<TrainerCard[]>([]);
	let trainersLoading = $state(false);
	let trainersLoaded = $state(false);

	async function loadTrainers() {
		if (trainersLoaded || trainersLoading) return;
		trainersLoading = true;
		try {
			const resp = await authFetch('/matchup/trainers?limit=50');
			if (resp.ok) {
				const body = await resp.json();
				trainers = body.data ?? [];
			}
		} catch { /* keep empty */ }
		finally {
			trainersLoading = false;
			trainersLoaded = true;
		}
	}

	async function messageTrainer(userId: string) {
		try {
			const resp = await authFetch('/chats', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ user_id: userId })
			});
			if (resp.ok) {
				const body = await resp.json();
				const chatId = body.data?.chat_id;
				if (chatId) goto(`/chats/${chatId}`);
			} else {
				toast.error($t('feed.toast_error'));
			}
		} catch {
			toast.error($t('feed.toast_error'));
		}
	}

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
		school: (pd.school as string) || (pd.club_name as string) || undefined,
		goals: c.goal
		};
	}

	async function loadFeed(replace = false) {
		if (exhausted && !replace) return;
		if (isLoadingMore && !replace) return;
		isLoadingMore = !replace;
		isLoading = replace;
		loadError = false;
		if (replace) {
			exhausted = false;
			seenIds.clear();
		}

		try {
			const resp = await authFetch('/matchup/feed?limit=20');
			if (resp.ok) {
				const body = await resp.json();
				// Backend signals "no profile" so we show the complete-your-profile CTA
				// rather than a generic empty deck.
				if (body.data?.no_profile) {
					noProfile = true;
					profiles = [];
					return;
				}
			noProfile = false;
			const candidates: FeedCandidate[] = body.data?.candidates ?? [];
			const mapped = candidates.map(mapCandidate);
			// The server already excludes previously swiped profiles, so a truly
			// empty response means the pool is exhausted. Use seenIds only to
			// de-duplicate within the same browsing session (e.g. rapid top-ups
			// that race each other), not as the exhaustion signal.
			if (mapped.length === 0) {
				exhausted = true;
			}
			if (replace) {
				seenIds.clear();
				mapped.forEach((p) => seenIds.add(p.id));
				profiles = mapped;
			} else {
				const fresh = mapped.filter((p) => !seenIds.has(p.id));
				fresh.forEach((p) => seenIds.add(p.id));
				profiles = [...profiles, ...fresh];
			}
			} else {
				loadError = replace; // show error state only on full-page loads
			}
		} catch {
			loadError = replace;
		} finally {
			isLoading = false;
			isLoadingMore = false;
		}
	}

	onMount(async () => {
		await filterStore.load();
		try {
			const resp = await authFetch('/me/profile');
			if (resp.ok) {
				const body = await resp.json();
				const profile = body.data ?? body;
				if (profile?.goal) userGoal = profile.goal;
			}
			// 404 and other non-OK cases are handled by the feed's no_profile flag
		} catch {
			// ignore; feed will surface the no_profile state if needed
		}
		loadFeed(true);
	});

	async function handleLike(id: string) {
		captureSwipe('LIKE', 'feed');
		seenIds.add(id);
		const liked = profiles.find((p) => p.id === id) ?? null;
		// Optimistically remove the top card; we'll restore it on unrecoverable errors.
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
			} else if (resp.status === 429) {
				// Rate-limited — put the card back and let the user know.
				if (liked) profiles = [liked, ...profiles];
				seenIds.delete(id);
				toast.error($t('feed.toast_rate_limit'));
			} else if (resp.status === 401) {
				// Session expired — authFetch will redirect; no toast needed.
			} else {
				// Any other server error — restore the card so the user can retry.
				if (liked) profiles = [liked, ...profiles];
				seenIds.delete(id);
				toast.error($t('feed.toast_error'));
			}
		} catch {
			// Network failure — put the card back so the user can try again.
			if (liked) profiles = [liked, ...profiles];
			seenIds.delete(id);
			toast.error($t('feed.toast_error'));
		}
		// Top-up after swipe is persisted so the server excludes the swiped user.
		if (profiles.length < 5 && !exhausted) loadFeed();
	}

	async function handlePass(id: string) {
		captureSwipe('PASS', 'feed');
		seenIds.add(id);
		const passed = profiles.find((p) => p.id === id) ?? null;
		removeTopCard();
		try {
			const resp = await authFetch('/matchup/swipe', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ target_user_id: id, action: 'PASS', source: 'feed' })
			});
			if (resp.status === 429) {
				if (passed) profiles = [passed, ...profiles];
				seenIds.delete(id);
				toast.error($t('feed.toast_rate_limit'));
			} else if (!resp.ok && resp.status !== 401) {
				// Any server error — restore the card.
				if (passed) profiles = [passed, ...profiles];
				seenIds.delete(id);
				toast.error($t('feed.toast_error'));
			}
		} catch {
			if (passed) profiles = [passed, ...profiles];
			seenIds.delete(id);
			toast.error($t('feed.toast_error'));
		}
		// Top-up after swipe is persisted.
		if (profiles.length < 5 && !exhausted) loadFeed();
	}

	async function handleHide(id: string) {
		seenIds.add(id);
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

	async function handleChatNow() {
		showMatch = false;
		if (matchedChatId) {
			goto(`/chats/${matchedChatId}`);
			return;
		}
		// Swipe returned no chat_id (CreateChat failed transiently) — create it now.
		if (!matchedProfile) return;
		try {
			const resp = await authFetch('/chats', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ user_id: matchedProfile.id })
			});
			if (resp.ok) {
				const body = await resp.json();
				const chatId = body.data?.chat_id;
				if (chatId) { goto(`/chats/${chatId}`); return; }
			}
			toast.error($t('feed.toast_error'));
		} catch {
			toast.error($t('feed.toast_error'));
		}
	}

	async function handleApplyFilters(filters: FilterState) {
		// Await the PUT so the server has the new preferences before the feed reloads.
		await filterStore.apply(filters);
		exhausted = false;
		loadFeed(true);
	}

	async function handleReport(id: string) {
		try {
			const resp = await authFetch(`/users/${id}/report`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ category: 'inappropriate' })
			});
			if (resp.ok) {
				toast.success($t('feed.toast_report_sent'));
			} else {
				toast.error($t('feed.toast_error'));
			}
		} catch {
			toast.error($t('feed.toast_error'));
		}
	}
</script>

<!-- Full-screen wrapper -->
<div class="relative h-[100dvh] overflow-hidden" style="background: linear-gradient(135deg, #8984da 0%, #b4b0e8 50%, #c8c8e8 100%)">
	{#if loadError && profiles.length === 0}
		<!-- Inline error state with retry -->
		<div class="absolute inset-0 flex flex-col items-center justify-center px-8">
			<i class="fi fi-rr-wifi-slash" style="font-size: 48px; color: rgba(255,255,255,0.35);"></i>
			<p class="mt-4 text-center text-[18px] font-bold text-white">{$t('feed.error_title')}</p>
			<p class="mt-2 text-center text-[13px] font-medium" style="color: rgba(255,255,255,0.55);">{$t('feed.error_subtitle')}</p>
			<button
				onclick={() => loadFeed(true)}
				class="mt-6 rounded-[50px] px-6 py-2.5 text-[14px] font-semibold text-white"
				style="background: #8984da;"
			>{$t('feed.retry')}</button>
		</div>
	{:else if activeTab === 'trainers'}
		<!-- Trainers feed tab -->
		{#if trainersLoading}
			<div class="absolute inset-0 flex items-center justify-center">
				<div class="h-10 w-10 animate-spin rounded-full border-4 border-white/20" style="border-top-color: white;"></div>
			</div>
		{:else if trainers.length === 0}
			<div class="absolute inset-0 flex flex-col items-center justify-center px-8">
				<i class="fi fi-rr-graduation-cap" style="font-size: 64px; color: rgba(255,255,255,0.35);"></i>
				<p class="mt-4 text-center text-[20px] font-black text-white">{$t('feed.tab_trainers')}</p>
				<p class="mt-2 text-center text-[14px] font-medium" style="color: rgba(255,255,255,0.6);">
					{$t('feed.trainers_coming_soon')}
				</p>
			</div>
		{:else}
			<!-- Scrollable list of trainer cards -->
			<div class="absolute inset-0 overflow-y-auto" style="padding: calc(max(env(safe-area-inset-top),8px) + 104px) 16px calc(var(--bottom-nav-clearance,101px) + 16px);">
				<div class="flex flex-col gap-3">
				{#each trainers as trainer (trainer.user_id)}
					<div class="flex items-center gap-3 rounded-[16px] p-3" style="background: rgba(255,255,255,0.07);">
						<img
							src={trainer.avatar || 'https://images.unsplash.com/photo-1545959570-a94084071b5d?w=200'}
							alt={trainer.first_name ?? 'Trainer'}
							loading="lazy"
							decoding="async"
							class="h-[56px] w-[56px] flex-shrink-0 rounded-full object-cover"
						/>
						<div class="min-w-0 flex-1">
							<p class="truncate text-[15px] font-bold text-white">{trainer.first_name ?? ''} {trainer.last_name ?? ''}</p>
							{#if trainer.categories?.length}
								<p class="truncate text-[12px] font-medium text-white/60">{trainer.categories.join(' · ')}</p>
							{/if}
							{#if trainer.bio}
								<p class="mt-0.5 truncate text-[11px]" style="color: rgba(255,255,255,0.45);">{trainer.bio}</p>
								{/if}
							</div>
							<button
								class="flex flex-shrink-0 items-center gap-1.5 rounded-[50px] px-3 py-2 text-[12px] font-semibold text-white"
								style="background: rgba(137,132,218,0.85);"
								onclick={() => messageTrainer(trainer.user_id)}
							>
								<i class="fi fi-rr-comment" style="font-size: 13px; line-height: 1;"></i>
								{$t('trainers.message')}
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{:else if isLoading}
		<div class="absolute inset-0 flex items-center justify-center">
			<div
				class="h-10 w-10 animate-spin rounded-full border-4 border-white/20"
				style="border-top-color: white;"
			></div>
		</div>
	{:else if profiles.length > 0}
		<!-- Card stack: background cards + top interactive card -->
		{#each profiles.slice(0, 3).reverse() as profile, i (profile.id)}
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
					<!-- {#key} forces a fresh component (and fresh spring) when a new card becomes top -->
					{#key profile.id}
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
					{/key}
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
			style="background: linear-gradient(180deg, rgba(137,132,218,0.55) 0%, rgba(137,132,218,0) 100%);"
		></div>

		<!-- Bottom gradient overlay (scaled to at most 40dvh) -->
		<div
			class="pointer-events-none absolute right-0 bottom-0 left-0 z-30"
			style="height: min(295px, 40dvh); background: linear-gradient(180deg, rgba(137,132,218,0) 0%, rgba(90,85,180,0.92) 100%);"
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

	<!-- Glassmorphic header — pointer-events: none on the wrapper so swipes that
	     start near the top edge still land on the card; only direct children are clickable. -->
	<div
		class="absolute right-0 left-0 z-40 flex items-center justify-between px-4 pointer-events-none"
		style="top: max(env(safe-area-inset-top), 8px); height: 54px;"
	>
		<!-- Filter pill button with active-count badge -->
		<button
			class="glass-pill relative flex items-center gap-2 px-3 pointer-events-auto"
			style="height: 38px;"
			onclick={() => (showFilter = true)}
			aria-label="Filter"
		>
			<i class="fi fi-rr-settings-sliders" style="font-size: 16px; line-height: 1; color: white;"></i>
			<span class="text-[13px] font-semibold text-white">{$t('feed.filter')}</span>
			{#if filterStore.activeCount > 0}
				<span
					class="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white"
					style="background: #e74c3c;"
				>{filterStore.activeCount}</span>
			{/if}
		</button>

		<div class="flex items-center gap-2 pointer-events-auto">
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
			{#if activeTab === 'partners' && profiles.length > 0}
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

	<!-- Partners / Trainers tab switcher, second row below the header -->
	<div
		class="pointer-events-none absolute right-0 left-0 z-40 flex justify-center px-4"
		style="top: calc(max(env(safe-area-inset-top), 8px) + 54px);"
	>
	<div class="glass-pill pointer-events-auto flex items-center" style="height: 36px; padding: 4px;">
		<button
			onclick={() => (activeTab = 'partners')}
			class="flex items-center justify-center rounded-[20px] px-4 text-[12px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'partners' ? 'white' : 'transparent'}; color: {activeTab === 'partners' ? '#3a3a3a' : 'white'};"
		>
			{$t('feed.tab_partners')}
		</button>
		<button
			onclick={() => { activeTab = 'trainers'; loadTrainers(); }}
			class="flex items-center justify-center rounded-[20px] px-4 text-[12px] font-semibold transition-colors"
			style="height: 28px; background: {activeTab === 'trainers' ? 'white' : 'transparent'}; color: {activeTab === 'trainers' ? '#3a3a3a' : 'white'};"
		>
			{$t('feed.tab_trainers')}
		</button>
	</div>
	</div>
</div>

<!-- Match popup -->
<MatchPopup
	open={showMatch}
	myPhoto={authStore.user?.profile_data?.avatar ?? ''}
	theirProfile={matchedProfile}
	onchat={handleChatNow}
	onclose={() => (showMatch = false)}
/>

<!-- Filter sheet -->
<FilterSheet
	open={showFilter}
	onclose={() => (showFilter = false)}
	onclear={async () => { await filterStore.clear(); exhausted = false; loadFeed(true); }}
	onapply={handleApplyFilters}
	initialFilters={filterStore.filters}
	{userGoal}
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
