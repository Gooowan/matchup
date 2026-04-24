<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';

	const DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;
	const DAY_LABELS: Record<string, string> = {
		mon: 'Monday', tue: 'Tuesday', wed: 'Wednesday', thu: 'Thursday',
		fri: 'Friday', sat: 'Saturday', sun: 'Sunday'
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
	}

	let clubs = $state<Club[]>([]);
	let isLoading = $state(true);
	let saving = $state<string | null>(null);

	// Per-club edit state
	let edits = $state<Record<string, {
		description: string;
		address: string;
		phone: string;
		website: string;
		hours: Record<Day, { enabled: boolean; open: string; close: string }>;
	}>>({});

	function initEdit(club: Club) {
		const hours = {} as Record<Day, { enabled: boolean; open: string; close: string }>;
		for (const day of DAYS) {
			const h = club.working_hours?.[day] ?? null;
			hours[day] = { enabled: !!h, open: h?.open ?? '09:00', close: h?.close ?? '21:00' };
		}
		edits[club.id] = {
			description: club.description ?? '',
			address: club.address ?? '',
			phone: club.phone ?? '',
			website: club.website ?? '',
			hours
		};
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
			toast.error('Failed to load your clubs');
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
			const resp = await authFetch(`/clubs/${club.id}/manage`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					description: edit.description,
					address: edit.address,
					phone: edit.phone,
					website: edit.website,
					working_hours
				})
			});
			if (resp.ok) {
				toast.success('Club updated');
			} else {
				toast.error('Failed to save — are you the club owner?');
			}
		} catch {
			toast.error('Something went wrong');
		} finally {
			saving = null;
		}
	}
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden mu-screen">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} aria-label="Back">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 text-[20px] font-black">Business Panel</h1>
	</div>

	{#if isLoading}
		<div class="flex flex-1 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
		</div>
	{:else if clubs.length === 0}
		<div class="flex flex-1 flex-col items-center justify-center gap-3 px-8 text-center">
			<i class="fi fi-rr-store-alt" style="font-size: 48px; color: #aeb4bc;"></i>
			<p class="text-[18px] font-bold mu-text-primary">No clubs yet</p>
			<p class="text-[14px] font-medium" style="color: #696969;">Find your club on the map and tap Claim to manage it here.</p>
		</div>
	{:else}
		<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 20px; padding-top: 12px;">
			{#each clubs as club}
				{@const edit = edits[club.id]}
				<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 14px;">
					<!-- Club name + badge -->
					<div class="flex items-center gap-2">
						<div class="flex-1">
							<p class="mu-text-primary text-[16px] font-black">{club.name}</p>
							<p class="text-[12px] font-medium" style="color: #aeb4bc;">{club.city}, {club.country}</p>
						</div>
						{#if club.is_verified}
							<span class="rounded-[65px] px-2 py-0.5 text-[11px] font-semibold text-white" style="background: #22c55e;">Verified</span>
						{:else}
							<span class="rounded-[65px] px-2 py-0.5 text-[11px] font-semibold" style="background: #fef3c7; color: #92400e;">Pending</span>
						{/if}
					</div>

					<!-- Description -->
					<div style="display: flex; flex-direction: column; gap: 4px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Description</label>
						<textarea
							bind:value={edit.description}
							rows="2"
							placeholder="Tell dancers about your club…"
							class="mu-text-primary w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
							style="border: 1px solid var(--mu-divider, #e0e0e0); border-radius: 12px; padding: 8px;"
						></textarea>
					</div>

					<!-- Address / Phone / Website -->
					<div style="display: flex; flex-direction: column; gap: 8px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Contact</label>
						<input
							type="text"
							placeholder="Address"
							bind:value={edit.address}
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
							style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 6px;"
						/>
						<input
							type="tel"
							placeholder="Phone"
							bind:value={edit.phone}
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
							style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 6px;"
						/>
						<input
							type="url"
							placeholder="Website"
							bind:value={edit.website}
							class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
						/>
					</div>

					<!-- Working hours -->
					<div style="display: flex; flex-direction: column; gap: 8px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Working Hours</label>
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
								<span class="w-[80px] text-[13px] font-medium mu-text-primary">{DAY_LABELS[day]}</span>
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
									<span class="text-[13px] font-medium" style="color: #aeb4bc;">Closed</span>
								{/if}
							</div>
						{/each}
					</div>

					<!-- Save button -->
					<button
						onclick={() => save(club)}
						disabled={saving === club.id}
						class="flex h-[44px] w-full items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
						style="background: #8984da;"
					>
						{saving === club.id ? 'Saving…' : 'Save changes'}
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>
