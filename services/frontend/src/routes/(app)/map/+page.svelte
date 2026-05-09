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

	let entities = $state<MapEntity[]>([]);
	let userLat = $state(0);
	let userLng = $state(0);
	let isLocating = $state(false);
	let clubMembers = $state<ClubMember[]>([]);
	let isLoadingMembers = $state(false);

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

	function applyMemberFilters(members: ClubMember[]): ClubMember[] {
		const f = filterStore.filters;
		return members.filter((m) => {
			if (f.gender && m.gender !== f.gender) return false;
			if (f.goal && m.goal !== f.goal) return false;
			if (f.program && m.program !== f.program) return false;
			if (f.categories && f.categories.length > 0) {
				const overlap = m.categories.some((c) => f.categories!.includes(c));
				if (!overlap) return false;
			}
			const age = calcAge(m.birth_date);
			if (f.ageMin != null && age != null && age < f.ageMin) return false;
			if (f.ageMax != null && age != null && age > f.ageMax) return false;
			return true;
		});
	}

	let filteredMembers = $derived(applyMemberFilters(clubMembers));

	function getPosition(): Promise<GeolocationPosition> {
		return new Promise((resolve, reject) => {
			navigator.geolocation.getCurrentPosition(resolve, reject, {
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
			const city = filterStore.filters.city;
			if (city) params.set('city', city);
			const url = `${import.meta.env.VITE_API_URL}/clubs${params.toString() ? '?' + params.toString() : ''}`;
			const resp = await fetch(url);
			if (resp.ok) {
				const body = await resp.json();
				const clubs = body.data ?? [];
				const clubEntities: MapEntity[] = clubs
					.filter((c: any) => c.latitude && c.longitude)
					.map((c: any) => ({
						type: 'clubs' as Category,
						id: c.id ?? c.slug,
						slug: c.slug,
						name: c.name,
						location: [c.city, c.country].filter(Boolean).join(', '),
						lat: c.latitude,
						lng: c.longitude,
						address: c.address?.String ?? c.address ?? '',
						phone: c.phone?.String ?? c.phone ?? '',
						website: c.website?.String ?? c.website ?? ''
					}));
				entities = clubEntities;
			}
		} catch {}
	}

	async function loadClubMembers(slug: string) {
		if (!slug) return;
		isLoadingMembers = true;
		clubMembers = [];
		try {
			const resp = await authFetch(`/clubs/${slug}/members?limit=50`);
			if (resp.ok) {
				const body = await resp.json();
				clubMembers = (body.data ?? []) as ClubMember[];
			}
		} catch {
			clubMembers = [];
		} finally {
			isLoadingMembers = false;
		}
	}

	function handleSearchInput() {
		if (searchTimeout) clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			loadClubs(searchQuery || undefined).then(() => addMarkers());
		}, 300);
	}

	let filtered = $derived(entities.filter((e) => e.type === activeCategory));

	onMount(async () => {
		filterStore.load();
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

	async function openChat(entity: MapEntity) {
		selectedEntity = null;
		if (entity.slug) {
			try {
				const resp = await authFetch(`/clubs/${entity.slug}/chat`, { method: 'POST' });
				if (resp.ok) {
					const body = await resp.json();
					const chatId = body.data?.chat_id ?? body.chat_id;
					if (chatId) {
						goto(`/chats/${chatId}`);
						return;
					}
				}
			} catch {}
		}
		goto('/chats');
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
	<!-- Map fills everything -->
	<div bind:this={mapContainer} class="absolute inset-0" style="z-index: 0;"></div>

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
				type="search"
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
				bottom: calc(101px + max(env(safe-area-inset-bottom), 8px) + 8px);
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
				bottom: calc(101px + max(env(safe-area-inset-bottom), 8px) + 8px);
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

<style>
.event-countdown {
	width: 100%;
	animation: event-countdown-shrink 3s linear forwards;
}
@keyframes event-countdown-shrink {
	from { width: 100%; }
	to   { width: 0%; }
}
</style>

<!-- Club bottom sheet -->
<BottomSheet open={!!selectedEntity} onclose={() => { selectedEntity = null; clubMembers = []; }}>
	{#if selectedEntity}
		<!-- Club header -->
		<div class="mb-4 flex items-start gap-4">
			{#if selectedEntity.logoUrl}
				<img src={selectedEntity.logoUrl} alt={selectedEntity.name} class="h-[80px] w-[80px] flex-shrink-0 rounded-full object-cover" />
			{:else}
				<div class="flex h-[80px] w-[80px] flex-shrink-0 items-center justify-center rounded-full" style="background: #696969;">
					<i class="fi fi-rr-bank text-white" style="font-size: 32px;"></i>
				</div>
			{/if}
			<div class="flex flex-col gap-1">
				<p class="mu-text-primary text-[18px] font-black">{selectedEntity.name}</p>
				<div class="flex items-center gap-1.5">
					<i class="fi fi-rr-marker mu-text-primary" style="font-size: 15px;"></i>
					<span class="mu-text-primary text-[12px] font-medium">{selectedEntity.location}</span>
				</div>
			</div>
		</div>

		{#if selectedEntity.address}
			<p class="mu-text-primary mb-3 text-[12px] font-medium">{selectedEntity.address}</p>
		{/if}

		<!-- Chat button -->
		<button
			class="flex h-[38px] w-full items-center justify-center gap-2 rounded-[65px] mb-5"
			style="background: #696969;"
			onclick={() => selectedEntity && openChat(selectedEntity)}
		>
			<i class="fi fi-rr-comment-heart" style="font-size: 20px; color: white;"></i>
			<span class="text-[14px] font-semibold text-white">{$t('map.write')}</span>
		</button>

		<!-- Dancers section -->
		<div>
			<p class="mu-text-primary mb-3 text-[13px] font-bold uppercase tracking-wider" style="color: #aeb4bc;">{$t('map.dancers')}</p>

			{#if isLoadingMembers}
				<div class="flex justify-center py-4">
					<div class="h-6 w-6 animate-spin rounded-full border-2 border-[#8984da]/30" style="border-top-color: #8984da;"></div>
				</div>
			{:else if filteredMembers.length === 0}
				<p class="py-3 text-center text-[13px] font-medium" style="color: #aeb4bc;">{$t('map.no_dancers')}</p>
			{:else}
				<div class="flex gap-3 overflow-x-auto pb-2" style="-webkit-overflow-scrolling: touch;">
					{#each filteredMembers as member}
						{@const avatar = member.profile_data?.avatar ?? ''}
						{@const firstName = member.profile_data?.first_name ?? ''}
						{@const age = calcAge(member.birth_date)}
						<button
							class="flex flex-shrink-0 flex-col items-center gap-1.5"
							style="width: 72px;"
							onclick={() => goto(`/profiles/${member.user_id}`)}
						>
							{#if avatar}
								<img src={avatar} alt={firstName} class="h-[72px] w-[72px] rounded-full object-cover" />
							{:else}
								<div class="flex h-[72px] w-[72px] items-center justify-center rounded-full" style="background: #dae1eb;">
									<i class="fi fi-rr-user" style="font-size: 28px; color: #aeb4bc;"></i>
								</div>
							{/if}
							<p class="mu-text-primary w-full truncate text-center text-[12px] font-semibold">{firstName || '—'}</p>
							{#if age !== null}
								<p class="text-[11px] font-medium" style="color: #aeb4bc;">{age} р.</p>
							{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</BottomSheet>

<!-- Filter sheet -->
<FilterSheet
	open={showFilter}
	onclose={() => (showFilter = false)}
	onclear={() => { filterStore.clear(); handleApplyFilters(); }}
	onapply={(f) => { filterStore.apply(f); handleApplyFilters(); }}
	initialFilters={filterStore.filters}
/>
