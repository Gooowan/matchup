<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import FilterSheet from '$lib/components/matchup/FilterSheet.svelte';
	import BottomSheet from '$lib/components/matchup/BottomSheet.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { filterStore } from '$stores/filters.svelte';
	import toast from 'svelte-french-toast';
	import { t } from '$lib/locale';

	type Category = 'clubs' | 'events';

	let activeCategory = $state<Category>('clubs');
	let showFilter = $state(false);
	let selectedEntity = $state<MapEntity | null>(null);
	let isDark = $state(browser && document.documentElement.classList.contains('dark'));
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	// Events popup state
	let showEventsPopup = $state(false);
	let eventsTimer: ReturnType<typeof setTimeout> | null = null;

	// Leaflet map refs
	let mapContainer: HTMLElement;
	let map: any;
	let L: any;

	interface MapEntity {
		type: Category;
		id: string;
		name: string;
		location: string;
		lat: number;
		lng: number;
		logoUrl?: string;
		photos?: string[];
		phone?: string;
		website?: string;
		address?: string;
		slug?: string;
	}

	interface ClubMember {
		user_id: string;
		profile_data: Record<string, string> | null;
		birth_date: string | null;
		gender: string;
		goal: string;
		program: string;
		categories: string[];
		country: string | null;
		city: string | null;
	}

	interface ClubTrainer {
		trainer_user_id: string;
		first_name: string;
		last_name?: string;
		avatar?: string;
		gender: string;
		city?: string;
		categories: string[];
	}

	type DayHours = { open?: string; close?: string };
	interface ClubDetail {
		description?: string;
		working_hours?: Record<string, DayHours | string | null>;
		website?: string;
	}

	const HOURS_DAY_ORDER = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;

	function orderedWorkingHours(
		hours: Record<string, DayHours | string | null> | undefined
	): Array<[string, DayHours | string | null]> {
		if (!hours) return [];
		const entries = Object.entries(hours);
		const indexOf = (day: string) => {
			const i = HOURS_DAY_ORDER.indexOf(day.toLowerCase().slice(0, 3) as (typeof HOURS_DAY_ORDER)[number]);
			return i === -1 ? HOURS_DAY_ORDER.length : i;
		};
		return entries.sort(([a], [b]) => indexOf(a) - indexOf(b));
	}

	function formatHours(value: DayHours | string | null): string {
		if (value == null) return $t('map.closed');
		if (typeof value === 'string') return value;
		if (value.open && value.close) return `${value.open} – ${value.close}`;
		return $t('map.closed');
	}

	function dayLabel(day: string): string {
		const key = day.toLowerCase().slice(0, 3);
		const label = $t(`map.day_${key}`);
		return label === `map.day_${key}` ? day : label;
	}

	let entities = $state<MapEntity[]>([]);
	let userLat = $state(0);
	let userLng = $state(0);
	let isLocating = $state(false);
	let clubMembers = $state<ClubMember[]>([]);
	let clubTrainers = $state<ClubTrainer[]>([]);
	let clubDetail = $state<ClubDetail | null>(null);
	let isLoadingMembers = $state(false);
	let showCaller = $state(false);
	let activePhotoIndex = $state(0);
	let photoScrollEl = $state<HTMLElement | null>(null);

	// --- Events popup ---
	function openEventsPopup() {
		showEventsPopup = true;
		if (eventsTimer) clearTimeout(eventsTimer);
		eventsTimer = setTimeout(closeEventsPopup, 3000);
	}

	function closeEventsPopup() {
		if (eventsTimer) clearTimeout(eventsTimer);
		eventsTimer = null;
		showEventsPopup = false;
		activeCategory = 'clubs';
	}

	function calcAge(birthDateStr: string | null): number | null {
		if (!birthDateStr) return null;
		const dob = new Date(birthDateStr);
		const diff = Date.now() - dob.getTime();
		const age = Math.floor(diff / (365.25 * 24 * 3600 * 1000));
		return age >= 0 ? age : null;
	}

	// Member filters are now applied server-side; retain client-side categories
	// filter since the backend query doesn't have a categories column.
	function applyMemberFilters(members: ClubMember[]): ClubMember[] {
		const f = filterStore.filters;
		if (!f.categories?.length) return members;
		return members.filter((m) =>
			m.categories?.some((c) => f.categories!.includes(c))
		);
	}

	let filteredMembers = $derived(applyMemberFilters(clubMembers));

	function getPosition(): Promise<GeolocationPosition> {
		return new Promise((resolve, reject) => {
			navigator.geolocation.getCurrentPosition(resolve, (err) => {
				// kCLErrorLocationUnknown (iOS) surfaces as POSITION_UNAVAILABLE (code 2).
				// Retry without high accuracy — works when GPS is unavailable but
				// network/cell location is still possible.
				if (err.code === 2 /* POSITION_UNAVAILABLE */) {
					navigator.geolocation.getCurrentPosition(resolve, reject, {
						enableHighAccuracy: false,
						timeout: 10000,
						maximumAge: 60000
					});
				} else {
					reject(err);
				}
			}, {
				enableHighAccuracy: true,
				timeout: 10000,
				maximumAge: 30000
			});
		});
	}

	async function loadClubs(q?: string) {
		try {
			const params = new URLSearchParams();
			if (q) params.set('q', q);
			// Fetch up to the server-side max (100) so newly-created unverified clubs
			// are included — they sort after verified ones but would otherwise fall off
			// the default page of 20.
			params.set('limit', '100');
			// Do NOT filter clubs by the partner city preference — that is a dancer
			// filter and would hide venues in other cities, including newly-created ones.
			const url = `${import.meta.env.VITE_API_URL}/clubs?${params.toString()}`;
			const resp = await fetch(url);
			if (resp.ok) {
				const body = await resp.json();
				const clubs = body.data ?? [];
				const clubEntities: MapEntity[] = clubs
					.filter(
						(c: any) =>
							Number.isFinite(c.latitude) &&
							Number.isFinite(c.longitude) &&
							!(c.latitude === 0 && c.longitude === 0)
					)
			.map((c: any) => {
				const meta = c.metadata ?? {};
				const toAbsolute = (u: string) =>
					u?.startsWith('/') ? `${import.meta.env.VITE_API_URL}${u}` : u;
				const rawPhotos: string[] = Array.isArray(meta.photos) ? meta.photos : [];
				const photos = rawPhotos.map(toAbsolute).filter(Boolean);
				const rawLogoUrl: string | undefined =
					meta.logo_url ?? meta.logo ?? rawPhotos[0];
				const logoUrl: string | undefined = rawLogoUrl ? toAbsolute(rawLogoUrl) : undefined;
				return {
					type: 'clubs' as Category,
					id: c.id ?? c.slug,
					slug: c.slug,
					name: c.name,
					location: [c.city, c.country].filter(Boolean).join(', '),
					lat: c.latitude,
					lng: c.longitude,
					address: c.address?.String ?? c.address ?? '',
					phone: c.phone?.String ?? c.phone ?? '',
					website: c.website?.String ?? c.website ?? '',
					logoUrl,
					photos
				};
			});
				entities = clubEntities;
			}
		} catch {}
	}

	async function loadClubMembers(slug: string) {
		if (!slug) return;
		isLoadingMembers = true;
		clubMembers = [];
		clubTrainers = [];
		clubDetail = null;
		try {
			const f = filterStore.filters;
			const params = new URLSearchParams();
			if (f.gender) params.set('gender', f.gender);
			if (f.goal) params.set('goal', f.goal);
			if (f.program) params.set('program', f.program);
			if (f.city) params.set('city', f.city);
			if (f.ageMin != null) params.set('age_min', String(f.ageMin));
			if (f.ageMax != null) params.set('age_max', String(f.ageMax));
			const qs = params.toString();

			// Parallel fetch: members + trainers + club detail.
			const [membersResp, trainersResp, detailResp] = await Promise.all([
				authFetch(`/clubs/${slug}/members${qs ? '?' + qs : ''}`),
				fetch(`${import.meta.env.VITE_API_URL}/clubs/${slug}/trainers`),
				fetch(`${import.meta.env.VITE_API_URL}/clubs/${slug}`)
			]);

			if (membersResp.ok) {
				const body = await membersResp.json();
				clubMembers = (body.data ?? []) as ClubMember[];
			}
			if (trainersResp.ok) {
				const body = await trainersResp.json();
				clubTrainers = (body.data ?? []) as ClubTrainer[];
			}
			if (detailResp.ok) {
				const body = await detailResp.json();
				const club = body.data?.club ?? body.data ?? {};
				clubDetail = {
					description: club.description?.String ?? club.description ?? undefined,
					working_hours: club.working_hours ?? undefined,
					website: club.website?.String ?? club.website ?? undefined
				};
			}
		} catch {
			clubMembers = [];
		} finally {
			isLoadingMembers = false;
		}
	}

	function zoomToResults(results: MapEntity[]) {
		if (!map || !L || results.length === 0) return;
		if (results.length === 1) {
			map.flyTo([results[0].lat, results[0].lng], 16, { animate: true, duration: 1.2, easeLinearity: 0.1 });
		} else {
			const bounds = L.latLngBounds(results.map((e) => [e.lat, e.lng]));
			map.flyToBounds(bounds, { padding: [60, 60], maxZoom: 15, animate: true, duration: 1.2, easeLinearity: 0.1 });
		}
	}

	function handleSearchInput() {
		if (searchTimeout) clearTimeout(searchTimeout);
		searchTimeout = setTimeout(async () => {
			await loadClubs(searchQuery || undefined);
			addMarkers();
			// Only zoom when user typed something; clear resets to default view
			if (searchQuery) {
				const results = entities.filter((e) => e.type === 'clubs');
				zoomToResults(results);
			}
		}, 300);
	}

	let filtered = $derived(entities.filter((e) => e.type === activeCategory));

	onMount(async () => {
		await filterStore.load();
		const leaflet = await import('leaflet');
		L = leaflet.default || leaflet;

		delete (L.Icon.Default.prototype as any)._getIconUrl;
		L.Icon.Default.mergeOptions({
			iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
			iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
			shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png'
		});

		map = L.map(mapContainer, {
			center: [50.4501, 30.5234],
			zoom: 12,
			zoomControl: false,
			attributionControl: false
		});

		isDark = document.documentElement.classList.contains('dark');
		// Dark theme keeps the dark_all tiles but a CSS filter lightens them to gray
		// instead of near-black. Light theme uses the light_all gray tiles.
		const tileStyle = isDark ? 'dark_all' : 'light_all';
		L.tileLayer(`https://{s}.basemaps.cartocdn.com/${tileStyle}/{z}/{x}/{y}{r}.png`, {
			maxZoom: 19
		}).addTo(map);

		if (browser && navigator.geolocation) {
			getPosition().then((pos) => {
				userLat = pos.coords.latitude;
				userLng = pos.coords.longitude;
				map.setView([userLat, userLng], 14);
			}).catch(() => {});
		}

		await loadClubs();
		addMarkers();
	});

	onDestroy(() => {
		if (searchTimeout) clearTimeout(searchTimeout);
		if (eventsTimer) clearTimeout(eventsTimer);
		map?.remove();
	});

	function addMarkers() {
		if (!map || !L) return;
		map.eachLayer((layer: any) => {
			if (layer instanceof L.Marker) map.removeLayer(layer);
		});

		if (activeCategory !== 'clubs') return;

		filtered.forEach((entity) => {
			const iconHtml = `<i class="fi fi-sr-map-marker-home" style="font-size:24px;color:#8984da;"></i>`;
			const icon = L.divIcon({
				html: iconHtml,
				className: '',
				iconSize: [24, 24],
				iconAnchor: [12, 24]
			});

			L.marker([entity.lat, entity.lng], { icon })
				.addTo(map)
				.on('click', () => {
					selectedEntity = entity;
					if (entity.slug) loadClubMembers(entity.slug);
				});
		});
	}

	$effect(() => {
		if (map && L) {
			activeCategory;
			addMarkers();
		}
	});

	async function nearMe() {
		if (!browser || !map || isLocating) return;
		isLocating = true;
		try {
			const pos = await getPosition();
			userLat = pos.coords.latitude;
			userLng = pos.coords.longitude;
			map.setView([userLat, userLng], 15);
			addMarkers();
		} catch {
			toast.error($t('map.geolocation_failed'));
		} finally {
			isLocating = false;
		}
	}

	async function messageUser(userId: string) {
		try {
			const resp = await authFetch('/chats', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ user_id: userId })
			});
			if (resp.ok) {
				const body = await resp.json();
				const chatId = body.data?.chat_id;
				if (chatId) { goto(`/chats/${chatId}`); return; }
			}
			toast.error($t('map.chat_error'));
		} catch {
			toast.error($t('map.chat_error'));
		}
	}

	async function openChat(entity: MapEntity) {
		if (!entity.slug) {
			// No club slug — nothing to chat with; open the inbox.
			goto('/chats');
			return;
		}
		// Close the sheet immediately so the user sees feedback.
		selectedEntity = null;
		try {
			const resp = await authFetch(`/clubs/${entity.slug}/chat`, { method: 'POST' });
			if (resp.ok) {
				const body = await resp.json();
				const chatId = body.data?.chat_id ?? body.chat_id;
				if (chatId) {
					goto(`/chats/${chatId}`);
					return;
				}
				// API returned OK but no chat_id — unexpected; show an error.
				toast.error($t('map.chat_error'));
			} else {
				toast.error($t('map.chat_error'));
			}
		} catch {
			toast.error($t('map.chat_error'));
		}
		// On any error we stay on the map page (selectedEntity already cleared).
	}

	function handleApplyFilters() {
		loadClubs(searchQuery || undefined).then(() => addMarkers());
	}
</script>

<svelte:head>
	<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
</svelte:head>

<!-- Screen wrapper -->
<div class="mu-screen relative h-[100dvh] overflow-hidden">
	<!-- Map fills everything. Neutral gray background avoids a white/black flash
	     before tiles paint (matches the gray basemap in both themes). -->
	<div
		bind:this={mapContainer}
		class="absolute inset-0"
		style="z-index: 0; background: {isDark ? '#3a3a3e' : '#e5e7eb'};"
	></div>

	<!-- Overlay UI layer -->
	<div class="pointer-events-none absolute inset-0" style="z-index: 10;">
		<!-- Search input -->
		<div
			class="pointer-events-auto absolute left-4 right-4 mu-card mu-border flex items-center gap-3"
			style="
				top: max(calc(env(safe-area-inset-top) + 8px), 54px);
				height: 38px;
				border-radius: 20px;
				border-width: 1px;
				border-style: solid;
				padding: 0 16px;
			"
		>
			<i class="fi fi-rr-search mu-text-primary" style="font-size: 20px; line-height: 1; flex-shrink: 0;"></i>
		<input
			type="text"
			inputmode="search"
			enterkeyhint="search"
			bind:value={searchQuery}
			oninput={handleSearchInput}
			placeholder={$t('map.search_placeholder')}
			class="mu-text-primary w-full bg-transparent text-[14px] font-semibold outline-none placeholder:font-normal"
			style="color: inherit;"
		/>
		</div>

		<!-- Two-tab bar: Clubs + Events -->
		<div
			class="pointer-events-auto absolute left-4 right-4 mu-card mu-border flex items-center gap-1"
			style="
				top: max(calc(env(safe-area-inset-top) + 54px), 106px);
				height: 36px;
				border-radius: 50px;
				border-width: 1px;
				border-style: solid;
				padding: 4px;
			"
		>
			<button
				class="flex flex-1 items-center justify-center rounded-[20px] text-[14px] font-semibold transition-colors"
				style="
					height: 28px;
					background: {activeCategory === 'clubs' ? (isDark ? '#8984da' : '#696969') : 'transparent'};
					color: {activeCategory === 'clubs' ? 'white' : (isDark ? '#e1e1e1' : '#171717')};
				"
				onclick={() => (activeCategory = 'clubs')}
			>
				{$t('map.tab_clubs')}
			</button>

			<button
				class="relative flex flex-1 items-center justify-center gap-1 rounded-[20px] text-[14px] font-semibold opacity-50"
				style="height: 28px; color: {isDark ? '#e1e1e1' : '#171717'};"
				onclick={() => { activeCategory = 'events'; openEventsPopup(); }}
			>
				{$t('map.tab_events')}
				<span
					class="rounded-full px-1 text-[10px] font-bold text-white"
					style="background: #8984da; padding: 1px 5px; line-height: 1.4;"
				>{$t('map.tab_events_soon')}</span>
			</button>
		</div>

		<!-- Filter FAB (bottom left) -->
		<button
			class="pointer-events-auto absolute flex items-center gap-4"
			style="
				bottom: calc(var(--bottom-nav-clearance) + max(env(safe-area-inset-bottom), 8px) + 8px);
				left: 16px;
				height: 38px;
				background: #696969;
				border-radius: 65px;
				padding: 0 24px 0 16px;
				color: white;
			"
			onclick={() => (showFilter = true)}
		>
			<i class="fi fi-rr-settings-sliders" style="font-size: 20px; line-height: 1;"></i>
			<span class="text-[14px] font-semibold">{$t('map.filter')}</span>
		</button>

		<!-- Near me FAB (bottom right) -->
		<button
			class="pointer-events-auto absolute flex items-center gap-4 transition-opacity disabled:opacity-60"
			style="
				bottom: calc(var(--bottom-nav-clearance) + max(env(safe-area-inset-bottom), 8px) + 8px);
				right: 16px;
				height: 38px;
				background: #696969;
				border-radius: 65px;
				padding: 0 24px 0 16px;
				color: white;
			"
			onclick={nearMe}
			disabled={isLocating}
		>
			{#if isLocating}
				<div class="h-[18px] w-[18px] animate-spin rounded-full border-2 border-white/30" style="border-top-color: white;"></div>
			{:else}
				<i class="fi fi-rr-map-marker" style="font-size: 20px; line-height: 1;"></i>
			{/if}
			<span class="text-[14px] font-semibold">{$t('map.near_me')}</span>
		</button>
	</div>
</div>

<!-- Events popup with blurred backdrop -->
{#if showEventsPopup}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 flex items-center justify-center"
		style="z-index: 200; backdrop-filter: blur(8px); background: rgba(0,0,0,0.15);"
		onclick={closeEventsPopup}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
		<div
			class="mx-6 rounded-[24px] px-8 py-10 text-center shadow-2xl"
			style="background: rgba(255,255,255,0.95); max-width: 320px; width: 100%;"
			onclick={closeEventsPopup}
		>
			<div class="mb-4 flex justify-center">
				<div class="flex h-[64px] w-[64px] items-center justify-center rounded-full" style="background: linear-gradient(135deg, #8984da, #b4b0e8);">
					<i class="fi fi-rr-calendar text-white" style="font-size: 28px; line-height: 1;"></i>
				</div>
			</div>
			<h2 class="text-[20px] font-black" style="color: #171717;">{$t('map.events_brief_title')}</h2>
			<p class="mt-1 text-[13px] font-medium" style="color: #696969;">{$t('map.events_brief_subtitle')}</p>
			<!-- Auto-dismiss progress bar -->
			<div class="mt-5 h-[3px] w-full overflow-hidden rounded-full" style="background: #e5e7eb;">
				<div class="event-countdown h-full rounded-full" style="background: #8984da;"></div>
			</div>
		</div>
	</div>
{/if}

<!-- Club bottom sheet -->
<BottomSheet open={!!selectedEntity} onclose={() => { selectedEntity = null; clubMembers = []; clubTrainers = []; clubDetail = null; showCaller = false; activePhotoIndex = 0; }}>
	{#if selectedEntity}
		<!-- Club header -->
		<div class="mb-4 flex items-start gap-4">
			{#if selectedEntity.logoUrl}
				<img src={selectedEntity.logoUrl} alt={selectedEntity.name} loading="lazy" decoding="async" class="h-[72px] w-[72px] flex-shrink-0 rounded-[16px] object-cover" />
			{:else}
				<div class="flex h-[72px] w-[72px] flex-shrink-0 items-center justify-center rounded-[16px]" style="background: rgba(137,132,218,0.15);">
					<i class="fi fi-rr-bank" style="font-size: 28px; color: #8984da;"></i>
				</div>
			{/if}
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p class="mu-text-primary text-[18px] font-black leading-tight">{selectedEntity.name}</p>
				<div class="flex items-center gap-1.5">
					<i class="fi fi-rr-marker mu-text-secondary" style="font-size: 13px; line-height: 1;"></i>
					<span class="mu-text-secondary truncate text-[12px] font-medium">{selectedEntity.location}</span>
				</div>
				{#if selectedEntity.address}
					<span class="mu-text-secondary truncate text-[11px]">{selectedEntity.address}</span>
				{/if}
			</div>
		</div>

		<!-- Photo gallery -->
		{#if selectedEntity.photos && selectedEntity.photos.length > 0}
			{@const photos = selectedEntity.photos}
			<div class="mb-4 -mx-4">
				<!-- Scrollable photo strip -->
				<div
					bind:this={photoScrollEl}
					class="flex overflow-x-auto snap-x snap-mandatory"
					style="scrollbar-width: none; -webkit-overflow-scrolling: touch; scroll-behavior: smooth;"
					onscroll={(e) => {
						const el = e.currentTarget as HTMLElement;
						const idx = Math.round(el.scrollLeft / el.clientWidth);
						activePhotoIndex = idx;
					}}
				>
					{#each photos as photoUrl, i}
						<div class="flex-shrink-0 snap-center" style="width: 100vw; max-width: 100%; aspect-ratio: 1;">
							<img
								src={photoUrl}
								alt="{selectedEntity.name} фото {i + 1}"
								loading={i === 0 ? 'eager' : 'lazy'}
								decoding="async"
								class="h-full w-full object-cover"
							/>
						</div>
					{/each}
				</div>
				<!-- Dot indicators (only if >1 photo) -->
				{#if photos.length > 1}
					<div class="mt-2 flex justify-center gap-1.5">
						{#each photos as _, i}
							<button
								aria-label="Фото {i + 1}"
								onclick={() => {
									activePhotoIndex = i;
									if (photoScrollEl) {
										photoScrollEl.scrollTo({ left: i * photoScrollEl.clientWidth, behavior: 'smooth' });
									}
								}}
								class="rounded-full transition-all"
								style="width: {activePhotoIndex === i ? '18px' : '6px'}; height: 6px; background: {activePhotoIndex === i ? '#8984da' : 'rgba(137,132,218,0.3)'};"
							></button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Description -->
		{#if clubDetail?.description}
			<p class="mu-text-primary mb-4 text-[13px] leading-relaxed" style="color: rgba(var(--mu-text-rgb, 23,23,23),0.75);">
				{clubDetail.description}
			</p>
		{/if}

		<!-- Opening hours -->
		{#if clubDetail?.working_hours && Object.keys(clubDetail.working_hours).length > 0}
			<div class="mb-4 flex flex-col gap-1.5">
				<p class="text-[11px] font-bold uppercase tracking-wider" style="color: #aeb4bc;">{$t('map.hours')}</p>
				{#each orderedWorkingHours(clubDetail.working_hours) as [day, hours]}
					<div class="flex justify-between text-[12px]">
						<span class="mu-text-secondary font-medium">{dayLabel(day)}</span>
						<span class="mu-text-primary font-semibold">{formatHours(hours)}</span>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Phone & Website row -->
		{#if selectedEntity.phone || clubDetail?.website}
			<div class="mb-4 flex flex-wrap gap-2">
				{#if selectedEntity.phone}
					<button
						class="flex items-center gap-1.5 rounded-[50px] px-3 py-1.5 text-[12px] font-semibold"
						style="background: rgba(34,197,94,0.12); color: #16a34a;"
						onclick={() => (showCaller = true)}
					>
						<i class="fi fi-rr-phone-call" style="font-size: 13px; line-height: 1;"></i>
						{selectedEntity.phone}
					</button>
				{/if}
				{#if clubDetail?.website}
					<a
						href={clubDetail.website}
						target="_blank"
						rel="noopener noreferrer"
						class="flex items-center gap-1.5 rounded-[50px] px-3 py-1.5 text-[12px] font-semibold"
						style="background: rgba(137,132,218,0.12); color: #8984da; text-decoration: none;"
					>
						<i class="fi fi-rr-globe" style="font-size: 13px; line-height: 1;"></i>
						{$t('map.website')}
					</a>
				{/if}
			</div>
		{/if}

		<!-- Message / Contact button -->
		<button
			class="mb-5 flex h-[44px] w-full items-center justify-center gap-2 rounded-[65px]"
			style="background: #8984da;"
			onclick={() => selectedEntity && openChat(selectedEntity)}
		>
			<i class="fi fi-rr-comment-heart" style="font-size: 18px; color: white;"></i>
			<span class="text-[14px] font-semibold text-white">{$t('map.write')}</span>
		</button>

		<!-- Trainers horizontal row -->
		{#if clubTrainers.length > 0}
			<div class="mb-5">
				<p class="mb-3 text-[11px] font-bold uppercase tracking-wider" style="color: #aeb4bc;">{$t('map.trainers')}</p>
				<div class="flex gap-3 overflow-x-auto pb-1" style="-webkit-overflow-scrolling: touch; scrollbar-width: none;">
				{#each clubTrainers as trainer (trainer.trainer_user_id)}
					<div class="flex flex-shrink-0 flex-col items-center gap-1" style="width: 72px;">
						<div class="relative">
							<img
								src={trainer.avatar || 'https://images.unsplash.com/photo-1545959570-a94084071b5d?w=200'}
								alt={trainer.first_name ?? 'Trainer'}
								loading="lazy"
								decoding="async"
								class="h-[60px] w-[60px] rounded-full object-cover"
							/>
						</div>
						<p class="w-full truncate text-center text-[11px] font-bold mu-text-primary">{trainer.first_name ?? ''}</p>
						{#if trainer.categories?.length}
							<p class="w-full truncate text-center text-[10px]" style="color: #8984da;">{trainer.categories[0]}</p>
						{/if}
							<button
								class="mt-0.5 rounded-[50px] px-2 py-0.5 text-[10px] font-semibold text-white"
								style="background: #8984da;"
								onclick={() => messageUser(trainer.trainer_user_id)}
							>
								{$t('trainers.message')}
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Dancers horizontal scroll (~2 visible at a time) -->
		<div>
			<p class="mb-3 text-[11px] font-bold uppercase tracking-wider" style="color: #aeb4bc;">{$t('map.dancers')}</p>

			{#if isLoadingMembers}
				<div class="flex justify-center py-4">
					<div class="h-6 w-6 animate-spin rounded-full border-2 border-[#8984da]/30" style="border-top-color: #8984da;"></div>
				</div>
			{:else if filteredMembers.length === 0}
				<p class="py-3 text-center text-[13px] font-medium" style="color: #aeb4bc;">{$t('map.no_dancers')}</p>
			{:else}
				<div class="flex gap-3 overflow-x-auto pb-2" style="-webkit-overflow-scrolling: touch; scrollbar-width: none;">
					{#each filteredMembers as member}
						{@const avatar = member.profile_data?.avatar ?? ''}
						{@const firstName = member.profile_data?.first_name ?? ''}
						{@const age = calcAge(member.birth_date)}
						{@const tags = [member.goal, member.program].filter(Boolean) as string[]}
						<button
							class="mu-card mu-border flex flex-shrink-0 flex-col gap-2 overflow-hidden rounded-[16px] p-2 text-left"
							style="border-width: 1px; border-style: solid; width: calc(50vw - 32px); min-width: 140px; max-width: 180px;"
							onclick={() => goto(`/profiles/${member.user_id}`)}
						>
							<div class="aspect-square w-full overflow-hidden rounded-[12px]" style="background: #dae1eb;">
							{#if avatar}
								<img src={avatar} alt={firstName} loading="lazy" decoding="async" class="h-full w-full object-cover" />
								{:else}
									<div class="flex h-full w-full items-center justify-center">
										<i class="fi fi-rr-user" style="font-size: 28px; color: #aeb4bc;"></i>
									</div>
								{/if}
							</div>
							<div class="flex flex-col gap-0.5 px-1 pb-1">
								<p class="mu-text-primary truncate text-[13px] font-bold">
									{firstName || '—'}{age !== null ? `, ${age}` : ''}
								</p>
								{#if tags.length > 0}
									<p class="truncate text-[10px] font-semibold" style="color: #8984da;">{tags.join(' · ')}</p>
								{/if}
								{#if member.city}
									<p class="mu-text-secondary truncate text-[10px] font-medium">{member.city}</p>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</BottomSheet>

<!-- Caller window -->
{#if showCaller && selectedEntity?.phone}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 flex items-center justify-center"
		style="z-index: 400; background: rgba(0,0,0,0.55); backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);"
		onclick={() => (showCaller = false)}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
		<div
			class="mu-card mx-8 flex w-full max-w-[300px] flex-col items-center rounded-[24px] px-6 py-8"
			style="box-shadow: 0 8px 40px rgba(0,0,0,0.35);"
			onclick={(e) => e.stopPropagation()}
		>
			<div
				class="mb-4 flex h-[72px] w-[72px] items-center justify-center overflow-hidden rounded-full"
				style="background: #8984da20; border: 1.5px solid #8984da44;"
			>
				<i class="fi fi-rr-bank" style="font-size: 28px; color: #8984da;"></i>
			</div>
			<p class="mu-text-primary text-[18px] font-semibold">{selectedEntity.name}</p>
			<a href="tel:{selectedEntity.phone}" class="mt-1 text-[15px] font-medium" style="color: #8984da;">{selectedEntity.phone}</a>
			<a
				href="tel:{selectedEntity.phone}"
				class="mt-6 flex h-[46px] w-full items-center justify-center gap-2 rounded-[50px] text-[15px] font-bold text-white"
				style="background: #22c55e;"
			>
				<i class="fi fi-rr-phone-call" style="font-size: 18px; line-height: 1;"></i>
				{$t('chats.call_action')}
			</a>
			<button
				class="mu-text-secondary mt-3 text-[14px] font-semibold"
				onclick={() => (showCaller = false)}
			>{$t('chats.call_cancel')}</button>
		</div>
	</div>
{/if}

<!-- Filter sheet -->
<FilterSheet
	open={showFilter}
	onclose={() => (showFilter = false)}
	onclear={() => { filterStore.clear(); handleApplyFilters(); }}
	onapply={(f) => { filterStore.apply(f); handleApplyFilters(); }}
	initialFilters={filterStore.filters}
/>

<style>
.event-countdown {
	width: 100%;
	animation: event-countdown-shrink 3s linear forwards;
}
@keyframes event-countdown-shrink {
	from { width: 100%; }
	to   { width: 0%; }
}
/* In dark mode lighten the near-black CartoDB dark tiles to a neutral gray. */
:global(.dark .leaflet-tile-pane) {
	filter: brightness(2.2) grayscale(0.35);
}
</style>
