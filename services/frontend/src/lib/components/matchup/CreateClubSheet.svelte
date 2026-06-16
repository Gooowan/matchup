<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { t } from '$lib/locale';
	import { formatUkrainianPhone } from '$lib/utils/phone';
	import toast from 'svelte-french-toast';

	// v1: locked to Ukraine / Kyiv. Coords default to Kyiv city centroid so the
	// club always lands on the map even when server-side geocoding is offline.
	const HARDCODED_COUNTRY = 'Ukraine';
	const HARDCODED_COUNTRY_LABEL = 'Україна';
	const HARDCODED_CITY = 'Kyiv';
	const HARDCODED_CITY_LABEL = 'Київ';
	const KYIV_CENTROID = { lat: 50.4501, lng: 30.5234 } as const;

	// FUTURE (multi-country): when expanding beyond Ukraine, swap the locked
	// country/city displays below for the commented-out <select> dropdowns and
	// bind `selectedCountry` / `selectedCity` into the submit body instead of the
	// HARDCODED_* constants.
	// const COUNTRIES = ['Україна', 'Польща', ...];
	// const CITIES_BY_COUNTRY: Record<string, string[]> = { 'Україна': ['Київ', ...], ... };
	// let selectedCountry = $state('Україна');
	// let selectedCity = $state('Київ');

	interface Props {
		open?: boolean;
		coords?: { lat: number; lng: number } | null;
		onclose?: () => void;
		oncreated?: (slug: string) => void;
	}

	let {
		open = false,
		coords = null,
		onclose,
		oncreated
	}: Props = $props();

	let name = $state('');
	let address = $state('');
	let website = $state('');
	let phone = $state('');
	let isSubmitting = $state(false);

	// Google Maps import state
	let showGmapsInput = $state(false);
	let gmapsURL = $state('');
	let isImporting = $state(false);
	let importedPhotos = $state<string[]>([]);
	let importedLat = $state<number | null>(null);
	let importedLng = $state<number | null>(null);
	// City/country resolved from GMaps import; overrides hardcoded constants.
	let importedCity = $state<string | null>(null);
	let importedCountry = $state<string | null>(null);
	// Working hours from GMaps import ({"Monday":"9:00 AM – 9:00 PM",...}).
	let importedWorkingHours = $state<Record<string, string> | null>(null);

	async function importFromGmaps() {
		if (!gmapsURL.trim()) return;
		isImporting = true;
		try {
			const resp = await authFetch('/me/clubs/parse-gmaps', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url: gmapsURL.trim() })
			});
			if (resp.ok) {
				const { data } = await resp.json();
				if (data.name)    name    = data.name;
				if (data.address) address = data.address;
				if (data.website) website = data.website;
				if (data.phone)   phone   = formatUkrainianPhone(data.phone);
				if (data.latitude)  importedLat = data.latitude;
				if (data.longitude) importedLng = data.longitude;
				// Use city/country from Google if the venue is outside Kyiv.
			if (data.city)         importedCity         = data.city;
			if (data.country)      importedCountry      = data.country;
			if (data.working_hours && Object.keys(data.working_hours).length > 0) {
				importedWorkingHours = data.working_hours;
			}
			importedPhotos = data.photos ?? [];
				showGmapsInput = false;
				gmapsURL = '';
				toast.success($t('map.import_gmaps_success'));
			} else {
				const err = await resp.json().catch(() => ({}));
				toast.error((err as { error?: string }).error || $t('map.import_gmaps_error'));
			}
		} catch {
			toast.error($t('map.import_gmaps_error'));
		} finally {
			isImporting = false;
		}
	}

	function handleClose() {
		if (isSubmitting) return;
		resetForm();
		onclose?.();
	}

	function resetForm() {
		name = '';
		address = '';
		website = '';
		phone = '';
		isSubmitting = false;
		showGmapsInput = false;
		gmapsURL = '';
		importedPhotos = [];
		importedLat = null;
		importedLng = null;
		importedCity = null;
		importedCountry = null;
		importedWorkingHours = null;
	}

	function handlePhoneInput(e: Event) {
		const target = e.currentTarget as HTMLInputElement;
		phone = formatUkrainianPhone(target.value);
	}

	function handlePhonePaste(e: ClipboardEvent) {
		const text = e.clipboardData?.getData('text') ?? '';
		if (!text) return;
		e.preventDefault();
		phone = formatUkrainianPhone(text);
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!name.trim()) return;

		isSubmitting = true;
		try {
			// Use Google-imported city/country if available, otherwise fall back to
			// the locked Kyiv/Ukraine defaults.
			const effectiveCountry = importedCountry ?? HARDCODED_COUNTRY;
			const effectiveCity    = importedCity    ?? HARDCODED_CITY;

			const body: Record<string, any> = {
				name: name.trim(),
				country: effectiveCountry,
				city: effectiveCity,
				address: address.trim(),
				website: website.trim() || undefined,
				phone: phone.trim() || undefined,
				working_hours: importedWorkingHours ?? undefined
			};

			// Coord priority: manual map pin > Google Maps import > server Nominatim geocoding.
			// Do NOT pre-fill city centroid — the server will handle fallback and inform
			// the user when only an approximate location was resolved.
			if (coords) {
				body.latitude = coords.lat;
				body.longitude = coords.lng;
			} else if (importedLat !== null && importedLng !== null) {
				body.latitude = importedLat;
				body.longitude = importedLng;
			} else {
				// Let the server geocode (or reject if address is absent/invalid).
				body.latitude = 0;
				body.longitude = 0;
			}

			if (importedPhotos.length > 0) {
				body.photos = importedPhotos;
			}

			const resp = await authFetch('/clubs/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			if (resp.ok) {
				const data = await resp.json();
				const club = data.data?.club ?? data.data ?? {};
				const slug = club.slug ?? '';
				// Warn the user if the server fell back to a city-centroid approximation.
				if (data.data?.geocode_warning) {
					toast(data.data.geocode_warning, { icon: '📍', duration: 6000 });
				}
				// Join as member + claim ownership so the creator can edit the club later.
				if (slug) {
					try {
						await authFetch(`/clubs/${slug}/join`, { method: 'POST' });
					} catch { /* non-fatal */ }
					try {
						await authFetch(`/clubs/${slug}/claim`, { method: 'POST' });
					} catch { /* non-fatal */ }
				}
				toast.success($t('map.create_club_success'));
				resetForm();
				oncreated?.(slug);
			} else {
				const err = await resp.json().catch(() => ({}));
				toast.error(err.error || $t('map.create_club_error'));
			}
		} catch {
			toast.error($t('map.create_club_error'));
		} finally {
			isSubmitting = false;
		}
	}
</script>

<BottomSheet {open} onclose={handleClose}>
	<div class="pb-4">
		<h2 class="mu-text-primary mb-5 text-[18px] font-black">{$t('map.create_club_title')}</h2>

		{#if coords}
			<div class="mb-4 flex items-center gap-2 rounded-[12px] px-3 py-2.5" style="background: rgba(137,132,218,0.12);">
				<i class="fi fi-sr-map-marker" style="font-size: 16px; color: #8984da; line-height: 1;"></i>
				<span class="text-[12px] font-medium" style="color: #8984da;">
					{coords.lat.toFixed(5)}, {coords.lng.toFixed(5)}
				</span>
			</div>
		{/if}

		<form id="create-club-form" onsubmit={handleSubmit} class="flex flex-col gap-4">
			<!-- Google Maps import -->
			{#if !showGmapsInput}
				<button
					type="button"
					onclick={() => (showGmapsInput = true)}
					class="flex items-center justify-center gap-2 rounded-[50px] py-2.5 text-[13px] font-semibold transition-opacity"
					style="background: rgba(137,132,218,0.12); color: #8984da;"
				>
					<i class="fi fi-rr-link" style="font-size: 15px; line-height: 1;"></i>
					{$t('map.import_from_gmaps')}
				</button>
			{:else}
				<div class="flex flex-col gap-2">
					<input
						type="url"
						bind:value={gmapsURL}
						placeholder={$t('map.import_gmaps_placeholder')}
						class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
						style="border-width: 1px; border-style: solid;"
					/>
					<div class="flex gap-2">
						<button
							type="button"
							onclick={() => { showGmapsInput = false; gmapsURL = ''; }}
							class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold"
							style="background: rgba(174,180,188,0.15); color: #aeb4bc;"
						>Скасувати</button>
						<button
							type="button"
							onclick={importFromGmaps}
							disabled={isImporting || !gmapsURL.trim()}
							class="flex flex-1 items-center justify-center gap-2 rounded-[50px] py-2.5 text-[13px] font-semibold text-white transition-opacity disabled:opacity-50"
							style="background: #8984da;"
						>
							{#if isImporting}
								<div class="h-4 w-4 animate-spin rounded-full border-2 border-white/30" style="border-top-color: white;"></div>
								{$t('map.import_gmaps_importing')}
							{:else}
								{$t('map.import_from_gmaps')}
							{/if}
						</button>
					</div>
				</div>
			{/if}

		<!-- Imported photos preview -->
		{#if importedPhotos.length > 0}
			<div class="flex gap-2 overflow-x-auto pb-1" style="-webkit-overflow-scrolling: touch;">
				{#each importedPhotos as photoUrl}
					<img
						src={photoUrl.startsWith('/') ? `${import.meta.env.VITE_API_URL}${photoUrl}` : photoUrl}
						alt="Фото клубу"
						loading="lazy"
						decoding="async"
						class="h-[72px] w-[72px] flex-shrink-0 rounded-[10px] object-cover"
					/>
				{/each}
			</div>
		{/if}

			<!-- Name -->
			<div class="flex flex-col gap-1.5">
				<label class="mu-text-primary text-[13px] font-semibold" for="club-name">
					{$t('map.name')} <span style="color: #e05252;">*</span>
				</label>
				<input
					id="club-name"
					bind:value={name}
					type="text"
					required
					minlength="2"
					maxlength="255"
					placeholder="Salsa Studio Kyiv"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

		<!-- Country (locked to Ukraine for v1) -->
		<div class="flex flex-col gap-1.5">
			<label class="mu-text-primary text-[13px] font-semibold" for="club-country">
				Країна <span style="color: #e05252;">*</span>
			</label>
			<div
				id="club-country"
				class="mu-card mu-border flex items-center justify-between rounded-[12px] px-4 py-3"
				style="border-width: 1px; border-style: solid; opacity: 0.85;"
			>
				<span class="mu-text-primary text-[14px] font-medium">{HARDCODED_COUNTRY_LABEL}</span>
				<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
			</div>
			<!-- FUTURE (multi-country): restore the country dropdown when expanding beyond UA.
			<div class="select-wrapper mu-card mu-border rounded-[12px] px-4 py-3" style="border-width:1px;border-style:solid;">
				<select bind:value={selectedCountry} onchange={() => (selectedCity = '')}>
					{#each COUNTRIES as c}<option value={c}>{c}</option>{/each}
				</select>
				<i class="fi fi-rr-angle-small-down select-arrow"></i>
			</div>
			-->
		</div>

		<!-- City (locked to Kyiv for v1) -->
		<div class="flex flex-col gap-1.5">
			<label class="mu-text-primary text-[13px] font-semibold" for="club-city">
				{$t('map.city')} <span style="color: #e05252;">*</span>
			</label>
			<div
				id="club-city"
				class="mu-card mu-border flex items-center justify-between rounded-[12px] px-4 py-3"
				style="border-width: 1px; border-style: solid; opacity: 0.85;"
			>
				<span class="mu-text-primary text-[14px] font-medium">{HARDCODED_CITY_LABEL}</span>
				<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
			</div>
			<!-- FUTURE (multi-country): restore the city dropdown when expanding beyond UA.
			<div class="select-wrapper mu-card mu-border rounded-[12px] px-4 py-3" style="border-width:1px;border-style:solid;">
				<select bind:value={selectedCity}>
					{#each (CITIES_BY_COUNTRY[selectedCountry] ?? []) as c}<option value={c}>{c}</option>{/each}
				</select>
				<i class="fi fi-rr-angle-small-down select-arrow"></i>
			</div>
			-->
		</div>

		<!-- Address -->
			<div class="flex flex-col gap-1.5">
				<label class="mu-text-primary text-[13px] font-semibold" for="club-address">
					{$t('map.address')}
				</label>
				<input
					id="club-address"
					bind:value={address}
					type="text"
					maxlength="500"
					placeholder="вул. Хрещатик, 1"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

			<!-- Website -->
			<div class="flex flex-col gap-1.5">
				<label class="mu-text-primary text-[13px] font-semibold" for="club-website">
					{$t('map.website')}
				</label>
				<input
					id="club-website"
					bind:value={website}
					type="url"
					maxlength="500"
					placeholder="https://example.com"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

			<!-- Phone -->
			<div class="flex flex-col gap-1.5">
				<label class="mu-text-primary text-[13px] font-semibold" for="club-phone">
					{$t('map.phone')}
				</label>
				<input
					id="club-phone"
					value={phone}
					oninput={handlePhoneInput}
					onpaste={handlePhonePaste}
					type="tel"
					inputmode="tel"
					autocomplete="tel"
					placeholder="+380 50 123 45 67"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

			</form>
	</div>

	{#snippet footer()}
		<button
			type="submit"
			form="create-club-form"
			disabled={isSubmitting || !name.trim()}
			class="flex h-[46px] w-full items-center justify-center gap-2 rounded-[65px] text-[15px] font-bold text-white transition-opacity disabled:opacity-50"
			style="background: #8984da;"
		>
			{#if isSubmitting}
				<div class="h-[18px] w-[18px] animate-spin rounded-full border-2 border-white/30" style="border-top-color: white;"></div>
				{$t('map.create_club_creating')}
			{:else}
				<i class="fi fi-rr-add" style="font-size: 18px; line-height: 1;"></i>
				{$t('map.create_club_submit')}
			{/if}
		</button>
	{/snippet}
</BottomSheet>
