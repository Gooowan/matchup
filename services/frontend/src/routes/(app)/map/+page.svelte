<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import FilterSheet, { type FilterState } from '$lib/components/matchup/FilterSheet.svelte';
	import BottomSheet from '$lib/components/matchup/BottomSheet.svelte';
	import { authFetch } from '$lib/utils/authFetch';

	type Category = 'schools' | 'dancers' | 'tailoring' | 'events';

	let activeCategory = $state<Category>('schools');
	let showFilter = $state(false);
	let selectedEntity = $state<MapEntity | null>(null);
	let activeFilters = $state<FilterState>({});
	let isDark = $state(browser && document.documentElement.classList.contains('dark'));

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
		rating?: number;
		ratingCount?: number;
		phone?: string;
		website?: string;
		hours?: string;
		address?: string;
		slug?: string;
		// Event-specific
		date?: string;
		description?: string;
		ticketsUrl?: string;
		bannerUrl?: string;
	}

	let entities = $state<MapEntity[]>([]);
	let userLat = $state(0);
	let userLng = $state(0);

	async function loadClubs() {
		try {
			const resp = await fetch(`${import.meta.env.VITE_API_URL}/clubs`);
			if (resp.ok) {
				const body = await resp.json();
				const clubs = body.data ?? [];
				const schoolEntities: MapEntity[] = clubs
					.filter((c: any) => c.latitude && c.longitude)
					.map((c: any) => ({
						type: 'schools' as Category,
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
				entities = [...entities.filter((e) => e.type !== 'schools'), ...schoolEntities];
			}
		} catch {}
	}

	async function loadNearbyDancers(lat: number, lng: number) {
		try {
			const resp = await authFetch('/map/nearby/radius', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ latitude: lat, longitude: lng, radius_km: activeFilters.distanceKm ?? 20 })
			});
			if (resp.ok) {
				const body = await resp.json();
				const nearby = body.data?.users ?? body.data ?? [];
				const dancerEntities: MapEntity[] = nearby
					.filter((u: any) => u.latitude && u.longitude)
					.map((u: any) => {
						const pd = u.profile_data ?? {};
						return {
							type: 'dancers' as Category,
							id: u.user_id ?? u.id,
							name: [pd.first_name, pd.last_name].filter(Boolean).join(' ') || 'Dancer',
							location: [u.city, u.country].filter(Boolean).join(', '),
							lat: u.latitude,
							lng: u.longitude,
							logoUrl: pd.avatar ?? ''
						};
					});
				entities = [...entities.filter((e) => e.type !== 'dancers'), ...dancerEntities];
			}
		} catch {}
	}

	let filtered = $derived(
		entities.filter((e) => {
			if (e.type !== activeCategory) return false;
			if (activeFilters.city && e.location && !e.location.toLowerCase().includes((activeFilters.city ?? '').toLowerCase())) return false;
			return true;
		})
	);

	onMount(async () => {
		const leaflet = await import('leaflet');
		L = leaflet.default || leaflet;

		delete (L.Icon.Default.prototype as any)._getIconUrl;
		L.Icon.Default.mergeOptions({
			iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
			iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
			shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png'
		});

		map = L.map(mapContainer, {
			center: [40.7128, -74.006],
			zoom: 14,
			zoomControl: false,
			attributionControl: false
		});

		isDark = document.documentElement.classList.contains('dark');
		const tileStyle = isDark ? 'dark_all' : 'light_all';
		L.tileLayer(`https://{s}.basemaps.cartocdn.com/${tileStyle}/{z}/{x}/{y}{r}.png`, {
			maxZoom: 19
		}).addTo(map);

		// Try to get user location, fall back to default
		navigator.geolocation.getCurrentPosition(
			async (pos) => {
				userLat = pos.coords.latitude;
				userLng = pos.coords.longitude;
				map.setView([userLat, userLng], 14);
				await loadNearbyDancers(userLat, userLng);
				addMarkers();
			},
			() => {}
		);

		await loadClubs();
		addMarkers();
	});

	onDestroy(() => {
		map?.remove();
	});

	function addMarkers() {
		if (!map || !L) return;
		map.eachLayer((layer: any) => {
			if (layer instanceof L.Marker) map.removeLayer(layer);
		});

		filtered.forEach((entity) => {
			const iconHtml =
				entity.type === 'events'
					? `<i class="fi fi-sr-calendar" style="font-size:24px;color:#696969;"></i>`
					: entity.type === 'schools'
						? `<i class="fi fi-sr-map-marker-home" style="font-size:24px;color:#8984da;"></i>`
						: `<i class="fi fi-sr-marker" style="font-size:24px;color:#696969;"></i>`;

			const icon = L.divIcon({
				html: iconHtml,
				className: '',
				iconSize: [24, 24],
				iconAnchor: [12, 24]
			});

			L.marker([entity.lat, entity.lng], { icon })
				.addTo(map)
				.on('click', () => (selectedEntity = entity));
		});
	}

	$effect(() => {
		if (map && L) {
			activeCategory;
			addMarkers();
		}
	});

	async function nearMe() {
		if (!browser || !map) return;
		navigator.geolocation.getCurrentPosition(
			async (pos) => {
				userLat = pos.coords.latitude;
				userLng = pos.coords.longitude;
				map.setView([userLat, userLng], 15);
				await loadNearbyDancers(userLat, userLng);
				addMarkers();
			},
			() => {}
		);
	}

	async function openChat(entity: MapEntity) {
		selectedEntity = null;
		if (entity.type === 'schools' && entity.id) {
			try {
				const resp = await authFetch(`/clubs/${entity.id}/chat`, { method: 'POST' });
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
		<!-- Search bar -->
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
			<span class="mu-text-primary text-[14px] font-semibold">Search</span>
		</div>

		<!-- Category tab bar -->
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
			{#each (['schools', 'dancers', 'tailoring', 'events'] as const) as cat}
				<button
					class="flex flex-1 items-center justify-center rounded-[20px] text-[14px] font-semibold transition-colors"
					style="
						height: 28px;
						background: {activeCategory === cat ? (isDark ? '#8984da' : '#696969') : 'transparent'};
						color: {activeCategory === cat ? 'white' : (isDark ? '#e1e1e1' : '#171717')};
					"
					onclick={() => (activeCategory = cat)}
				>
					{cat.charAt(0).toUpperCase() + cat.slice(1)}
				</button>
			{/each}
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
			<span class="text-[14px] font-semibold">Filter</span>
		</button>

		<!-- Near me FAB (bottom right) -->
		<button
			class="pointer-events-auto absolute flex items-center gap-4"
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
		>
			<i class="fi fi-rr-map-marker" style="font-size: 20px; line-height: 1;"></i>
			<span class="text-[14px] font-semibold">Near me</span>
		</button>
	</div>
</div>

<!-- Entity bottom sheet -->
<BottomSheet open={!!selectedEntity} onclose={() => (selectedEntity = null)}>
	{#if selectedEntity}
		{#if selectedEntity.type === 'events'}
			<!-- Event card -->
			{#if selectedEntity.bannerUrl}
				<div class="relative mb-3 overflow-hidden rounded-[20px]" style="height: 179px;">
					<img src={selectedEntity.bannerUrl} alt={selectedEntity.name} class="h-full w-full object-cover" />
					<div class="absolute inset-0" style="background: linear-gradient(180deg, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0) 100%);"></div>
				</div>
			{/if}
			{#if selectedEntity.date}
				<p class="mu-text-primary mb-1 text-[14px] font-extrabold">{selectedEntity.date}</p>
			{/if}
			<p class="mu-text-primary mb-2 text-[18px] font-black">{selectedEntity.name}</p>
			<div class="mb-1 flex items-center gap-1.5">
				<i class="fi fi-rr-marker mu-text-primary" style="font-size: 15px;"></i>
				<span class="mu-text-primary text-[12px] font-medium">{selectedEntity.location}</span>
			</div>
			{#if selectedEntity.ticketsUrl}
				<div class="mb-1 flex items-center gap-1.5">
					<i class="fi fi-rr-ticket mu-text-primary" style="font-size: 15px;"></i>
					<span class="mu-text-primary text-[12px] font-medium">{selectedEntity.ticketsUrl}</span>
				</div>
			{/if}
			{#if selectedEntity.description}
				<p class="mu-text-primary mb-4 text-[12px] font-medium leading-snug">{selectedEntity.description}</p>
			{/if}
			<div class="flex items-center gap-2">
				<button class="flex h-[38px] items-center justify-center rounded-[65px] px-4" style="background: #696969;">
					<i class="fi fi-rr-bookmark" style="font-size: 20px; color: white;"></i>
				</button>
				<button class="flex h-[38px] flex-1 items-center justify-center gap-4 rounded-[65px] px-4" style="background: #696969;">
					<i class="fi fi-rr-calendar-pen" style="font-size: 20px; color: white;"></i>
					<span class="text-[14px] font-semibold text-white">Add to calendar</span>
				</button>
			</div>
		{:else}
			<!-- School / Dancer / Tailor card -->
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
					{#if selectedEntity.ratingCount}
						<div class="flex items-center gap-1.5">
							<span class="mu-text-primary text-[12px] font-medium">★★★★★ ({selectedEntity.ratingCount})</span>
						</div>
					{/if}
				</div>
			</div>

			{#if selectedEntity.hours}
				<p class="mu-text-primary mb-1 text-[12px] font-medium">Open hours: {selectedEntity.hours}</p>
			{/if}
			{#if selectedEntity.address}
				<p class="mu-text-primary mb-3 text-[12px] font-medium">Address: {selectedEntity.address}</p>
			{/if}

			{#if selectedEntity.photos && selectedEntity.photos.length}
				<div class="mb-4 flex gap-2">
					{#each selectedEntity.photos as photo}
						<div class="overflow-hidden rounded-[20px]" style="width: 217px; height: 107px; flex-shrink: 0;">
							<img src={photo} alt="" class="h-full w-full object-cover" />
						</div>
					{/each}
				</div>
			{/if}

			<div class="flex items-center gap-2">
				<button class="flex h-[38px] items-center justify-center rounded-[65px] px-4" style="background: #696969;">
					<i class="fi {selectedEntity.type === 'dancers' ? 'fi-rr-heart' : 'fi-rr-bookmark'}" style="font-size: 20px; color: white;"></i>
				</button>
				<button class="flex h-[38px] flex-1 items-center justify-center rounded-[65px] px-6" style="background: #696969;">
					<span class="text-[14px] font-semibold text-white">View Profile</span>
				</button>
				<button
					class="flex h-[38px] items-center gap-4 rounded-[65px]"
					style="background: #696969; padding: 0 24px 0 16px;"
					onclick={() => selectedEntity && openChat(selectedEntity)}
				>
					<i class="fi fi-rr-comment-heart" style="font-size: 20px; color: white;"></i>
					<span class="text-[14px] font-semibold text-white">Chat</span>
				</button>
			</div>
		{/if}
	{/if}
</BottomSheet>

<!-- Filter sheet -->
<FilterSheet
	open={showFilter}
	onclose={() => (showFilter = false)}
	onclear={() => (activeFilters = { danceStyles: [], role: '', distanceKm: 50, city: '' })}
	onapply={(f) => { activeFilters = f; }}
	initialFilters={activeFilters}
/>
