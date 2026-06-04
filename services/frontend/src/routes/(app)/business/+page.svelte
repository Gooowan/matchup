<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import AsyncState from '$lib/components/matchup/AsyncState.svelte';
	import toast from 'svelte-french-toast';
	import { t } from '$lib/locale';

	const DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;
	const DAY_KEY: Record<string, string> = {
		mon: 'business.day_mon', tue: 'business.day_tue', wed: 'business.day_wed',
		thu: 'business.day_thu', fri: 'business.day_fri', sat: 'business.day_sat',
		sun: 'business.day_sun'
	};

	type Day = typeof DAYS[number];

	interface DayHours { open: string; close: string }
	interface Club {
		id: string;
		name: string;
		slug: string;
		description: string;
		address: string;
		phone: string;
		website: string;
		working_hours: Record<string, DayHours | null> | null;
		is_verified: boolean;
		city: string;
		country: string;
		metadata: Record<string, unknown> | null;
	}

	let clubs = $state<Club[]>([]);
	let isLoading = $state(true);
	let saving = $state<string | null>(null);

	// Per-club edit state
	let edits = $state<Record<string, {
		name: string;
		description: string;
		address: string;
		phone: string;
		website: string;
		hours: Record<Day, { enabled: boolean; open: string; close: string }>;
		logoUrl: string;
		logoFile: File | null;
		logoPreview: string;
	}>>({});

	function initEdit(club: Club) {
		const hours = {} as Record<Day, { enabled: boolean; open: string; close: string }>;
		for (const day of DAYS) {
			const h = club.working_hours?.[day] ?? null;
			hours[day] = { enabled: !!h, open: h?.open ?? '09:00', close: h?.close ?? '21:00' };
		}
		const existingLogo = (club.metadata?.logo_url as string) ?? '';
		edits[club.id] = {
			name: club.name ?? '',
			description: club.description ?? '',
			address: club.address ?? '',
			phone: club.phone ?? '',
			website: club.website ?? '',
			hours,
			logoUrl: existingLogo,
			logoFile: null,
			logoPreview: existingLogo
		};
	}

	async function handleLogoChange(clubId: string, e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		edits[clubId].logoFile = file;
		edits[clubId].logoPreview = URL.createObjectURL(file);
	}

	onMount(async () => {
		try {
			const resp = await authFetch('/me/owned-clubs');
			if (resp.ok) {
				const body = await resp.json();
				clubs = body.data ?? [];
				for (const club of clubs) initEdit(club);
			}
		} catch {
			toast.error($t('business.load_error'));
		} finally {
			isLoading = false;
		}
	});

	async function save(club: Club) {
		saving = club.id;
		const edit = edits[club.id];
		const working_hours: Record<string, DayHours | null> = {};
		for (const day of DAYS) {
			working_hours[day] = edit.hours[day].enabled
				? { open: edit.hours[day].open, close: edit.hours[day].close }
				: null;
		}
		try {
			// Upload new logo first, get the URL back.
			let logoUrl = edit.logoUrl;
			if (edit.logoFile) {
				const form = new FormData();
				form.append('photo', edit.logoFile);
				const uploadResp = await authFetch('/user/files/photo', { method: 'POST', body: form });
				if (uploadResp.ok) {
					const uploadBody = await uploadResp.json();
					logoUrl = uploadBody.data?.url ?? logoUrl;
					edits[club.id].logoUrl = logoUrl;
					edits[club.id].logoFile = null;
				} else {
					toast.error($t('business.save_error'));
					return;
				}
			}

			const resp = await authFetch(`/clubs/${club.slug}/manage`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: edit.name.trim(),
					description: edit.description,
					address: edit.address,
					phone: edit.phone,
					website: edit.website,
					working_hours,
					logo_url: logoUrl
				})
			});
			if (resp.ok) {
				toast.success($t('business.save_success'));
				club.name = edit.name.trim() || club.name;
			} else {
				const err = await resp.json().catch(() => ({}));
				toast.error((err as any).error || $t('business.save_error'));
			}
		} catch {
			toast.error($t('business.save_error_generic'));
		} finally {
			saving = null;
		}
	}
</script>

<div class="mu-screen" style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch;">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} aria-label="Back">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 text-[20px] font-black">{$t('business.title')}</h1>
	</div>

	<AsyncState
		loading={isLoading}
		empty={!isLoading && clubs.length === 0}
		emptyIcon="fi-rr-store-alt"
		emptyText={$t('business.empty')}
	>
	{#if !isLoading && clubs.length > 0}
		<div class="flex flex-col px-4 pb-[100px]" style="gap: 20px; padding-top: 12px;">
			{#each clubs as club}
				{@const edit = edits[club.id]}
				<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 14px;">
			<!-- Club header: logo + name + badge -->
				<div class="flex items-center gap-3">
					<!-- Logo picker -->
					<label class="relative flex-shrink-0 cursor-pointer">
						<div
							class="flex h-[64px] w-[64px] items-center justify-center overflow-hidden rounded-[16px]"
							style="background: #e5e7eb;"
						>
							{#if edit.logoPreview}
								<img src={edit.logoPreview} alt={club.name} class="h-full w-full object-cover" />
							{:else}
								<i class="fi fi-rr-bank" style="font-size: 28px; color: #aeb4bc;"></i>
							{/if}
						</div>
						<!-- Edit overlay -->
						<div
							class="absolute inset-0 flex items-center justify-center rounded-[16px]"
							style="background: rgba(0,0,0,0.35);"
						>
							<i class="fi fi-rr-camera" style="font-size: 14px; color: white;"></i>
						</div>
						<input
							type="file"
							accept="image/jpeg,image/png,image/webp"
							class="hidden"
							onchange={(e) => handleLogoChange(club.id, e)}
						/>
					</label>
					<div class="flex-1 min-w-0">
						<p class="mu-text-primary text-[16px] font-black truncate">{club.name}</p>
						<p class="text-[12px] font-medium" style="color: #aeb4bc;">{club.city}, {club.country}</p>
					</div>
					{#if club.is_verified}
						<span class="rounded-[65px] px-2 py-0.5 text-[11px] font-semibold text-white" style="background: #22c55e;">{$t('business.verified')}</span>
					{:else}
						<span class="rounded-[65px] px-2 py-0.5 text-[11px] font-semibold flex-shrink-0" style="background: #fef3c7; color: #92400e;">{$t('business.pending')}</span>
					{/if}
				</div>

					<!-- Club name -->
					<div style="display: flex; flex-direction: column; gap: 4px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('business.name_label')}</label>
						<input
							type="text"
							bind:value={edit.name}
							maxlength="255"
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
							style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 6px;"
						/>
						{#if edit.name.trim().length > 0 && edit.name.trim().length < 2}
							<span class="text-[11px] font-medium" style="color: #e05252;">{$t('business.name_label')} — min 2</span>
						{/if}
					</div>

					<!-- Description -->
					<div style="display: flex; flex-direction: column; gap: 4px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('business.description_label')}</label>
						<textarea
							bind:value={edit.description}
							rows="2"
							maxlength="2000"
							placeholder={$t('business.description_placeholder')}
							class="mu-text-primary w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
							style="border: 1px solid var(--mu-divider, #e0e0e0); border-radius: 12px; padding: 8px;"
						></textarea>
					</div>

					<!-- Address / Phone / Website -->
					<div style="display: flex; flex-direction: column; gap: 8px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('business.contact_label')}</label>
						<input
							type="text"
							placeholder={$t('business.address_placeholder')}
							bind:value={edit.address}
							maxlength="500"
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
							style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 6px;"
						/>
						<input
							type="tel"
							placeholder={$t('business.phone_placeholder')}
							bind:value={edit.phone}
							maxlength="50"
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
							style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 6px;"
						/>
						<input
							type="url"
							placeholder={$t('business.website_placeholder')}
							bind:value={edit.website}
							maxlength="500"
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
						/>
					</div>

					<!-- Working hours -->
					<div style="display: flex; flex-direction: column; gap: 8px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('business.hours_label')}</label>
						{#each DAYS as day}
							<div class="flex items-center gap-3">
								<!-- Toggle -->
								<button
									onclick={() => (edit.hours[day].enabled = !edit.hours[day].enabled)}
									class="relative flex-shrink-0 transition-colors"
									style="width: 40px; height: 22px; border-radius: 50px; background: {edit.hours[day].enabled ? '#8984da' : '#d1d5db'};"
									role="switch"
									aria-checked={edit.hours[day].enabled}
								>
									<div
										class="absolute h-[18px] w-[18px] rounded-full bg-white shadow-sm transition-transform"
										style="top: 2px; transform: translateX({edit.hours[day].enabled ? '20px' : '2px'});"
									></div>
								</button>
								<span class="w-[80px] text-[13px] font-medium mu-text-primary">{$t(DAY_KEY[day])}</span>
								{#if edit.hours[day].enabled}
									<input
										type="time"
										bind:value={edit.hours[day].open}
										class="bg-transparent text-[13px] font-medium outline-none"
										style="color: #8984da;"
									/>
									<span class="text-[13px]" style="color: #aeb4bc;">–</span>
									<input
										type="time"
										bind:value={edit.hours[day].close}
										class="bg-transparent text-[13px] font-medium outline-none"
										style="color: #8984da;"
									/>
								{:else}
									<span class="text-[13px] font-medium" style="color: #aeb4bc;">{$t('business.closed')}</span>
								{/if}
							</div>
						{/each}
					</div>

					<!-- Save button -->
					<button
						onclick={() => save(club)}
						disabled={saving === club.id || edit.name.trim().length < 2}
						class="flex h-[44px] w-full items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
						style="background: #8984da;"
					>
						{saving === club.id ? $t('business.saving') : $t('business.save')}
					</button>
				</div>
			{/each}
		</div>
	{/if}
	</AsyncState>
</div>
