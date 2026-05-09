<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import ClubSearchSheet from '$lib/components/matchup/ClubSearchSheet.svelte';
	import CreateClubSheet from '$lib/components/matchup/CreateClubSheet.svelte';
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	interface Club {
		id: string;
		slug: string;
		name: string;
		city: string;
		country: string;
		is_verified?: boolean;
	}

	let joinedClubs = $state<Club[]>([]);
	let isLoading = $state(true);
	let primaryClubId = $state<string | null>(
		(authStore.user?.profile_data?.primary_club_id as string) ?? null
	);
	let leavingSlug = $state<string | null>(null);
	let settingPrimaryId = $state<string | null>(null);

	let showSearchSheet = $state(false);
	let showCreateSheet = $state(false);

	let userCity = $derived((authStore.user?.profile_data?.city as string) ?? '');
	let userCountry = $derived((authStore.user?.profile_data?.country as string) ?? 'Ukraine');

	onMount(async () => {
		await loadMyClubs();
	});

	async function loadMyClubs() {
		isLoading = true;
		try {
			const resp = await authFetch('/me/clubs');
			if (resp.ok) {
				const body = await resp.json();
				joinedClubs = (body.data ?? []) as Club[];
			}
		} catch {
			joinedClubs = [];
		} finally {
			isLoading = false;
		}
	}

	async function setPrimary(club: Club) {
		if (settingPrimaryId) return;
		settingPrimaryId = club.id;
		const prevId = primaryClubId;
		primaryClubId = club.id;
		try {
			const resp = await authFetch('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ primary_club_id: club.id })
			});
			if (resp.ok) {
				if (authStore.user) {
					authStore.user = {
						...authStore.user,
						profile_data: { ...authStore.user.profile_data, primary_club_id: club.id }
					};
				}
			} else {
				primaryClubId = prevId;
				toast.error($t('settings.clubs_join_error'));
			}
		} catch {
			primaryClubId = prevId;
		} finally {
			settingPrimaryId = null;
		}
	}

	async function leaveClub(club: Club) {
		if (!confirm($t('settings.clubs_leave_confirm'))) return;
		leavingSlug = club.slug;
		try {
			const resp = await authFetch(`/clubs/${club.slug}/join`, { method: 'DELETE' });
			if (resp.ok) {
				joinedClubs = joinedClubs.filter((c) => c.id !== club.id);
				if (primaryClubId === club.id) {
					primaryClubId = null;
					await authFetch('/me/profile', {
						method: 'PUT',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ primary_club_id: null })
					});
					if (authStore.user) {
						authStore.user = {
							...authStore.user,
							profile_data: { ...authStore.user.profile_data, primary_club_id: null }
						};
					}
				}
			} else {
				toast.error($t('settings.clubs_leave_error'));
			}
		} catch {
			toast.error($t('settings.clubs_leave_error'));
		} finally {
			leavingSlug = null;
		}
	}

	async function handleJoined(club: Club) {
		await loadMyClubs();
		// Make the newly joined club primary if user had none
		if (!primaryClubId) {
			await setPrimary(club);
		}
	}

	async function handleCreated(slug: string) {
		showCreateSheet = false;
		await loadMyClubs();
		// Find the newly created club and make it primary
		const created = joinedClubs.find((c) => c.slug === slug);
		if (created && !primaryClubId) {
			await setPrimary(created);
		}
	}
</script>

<div class="mu-screen flex h-[100dvh] flex-col overflow-hidden">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex flex-shrink-0 items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} class="flex items-center justify-center" aria-label="Назад">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 text-[20px] font-black">{$t('settings.clubs_title')}</h1>
	</div>

	{#if isLoading}
		<div class="flex flex-1 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
		</div>
	{:else}
		<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 16px; padding-top: 8px; -webkit-overflow-scrolling: touch;">

			<!-- Joined clubs list -->
			<div class="mu-card overflow-hidden rounded-[20px]">
				<p class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">
					{$t('settings.clubs_section_joined')}
				</p>

				{#if joinedClubs.length === 0}
					<div class="mu-divider px-4 pb-5 pt-2" style="border-top-width: 1px; border-top-style: solid;">
						<p class="py-4 text-center text-[14px] font-medium" style="color: #aeb4bc;">{$t('settings.clubs_empty')}</p>
					</div>
				{:else}
					<div class="mu-divider flex flex-col" style="border-top-width: 1px; border-top-style: solid;">
						{#each joinedClubs as club, i}
							<div
								class="flex items-center gap-3 px-4 py-3 {i > 0 ? 'mu-divider' : ''}"
								style="{i > 0 ? 'border-top-width: 1px; border-top-style: solid;' : ''}"
							>
								<!-- Club icon -->
								<div class="flex h-[44px] w-[44px] flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.15);">
									<i class="fi fi-rr-bank" style="font-size: 20px; color: #8984da;"></i>
								</div>

								<!-- Club info -->
								<div class="flex min-w-0 flex-1 flex-col gap-0.5 overflow-hidden">
									<div class="flex items-center gap-2">
										<span class="mu-text-primary truncate text-[14px] font-semibold">{club.name}</span>
										{#if primaryClubId === club.id}
											<span class="flex-shrink-0 rounded-full px-2 py-0.5 text-[10px] font-bold text-white" style="background: #8984da;">
												{$t('settings.clubs_primary_label')}
											</span>
										{/if}
									</div>
									<span class="truncate text-[12px] font-medium" style="color: #aeb4bc;">{[club.city, club.country].filter(Boolean).join(', ')}</span>
								</div>

								<!-- Actions -->
								<div class="flex flex-shrink-0 items-center gap-1">
									<!-- Star: set primary -->
									<button
										onclick={() => setPrimary(club)}
										disabled={primaryClubId === club.id || settingPrimaryId === club.id}
										class="flex h-[34px] w-[34px] items-center justify-center rounded-full transition-opacity disabled:opacity-40"
										aria-label={$t('settings.clubs_make_primary')}
									>
										{#if settingPrimaryId === club.id}
											<div class="h-4 w-4 animate-spin rounded-full border-2" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
										{:else}
											<i
												class="fi {primaryClubId === club.id ? 'fi-sr-star' : 'fi-rr-star'}"
												style="font-size: 18px; color: {primaryClubId === club.id ? '#8984da' : '#aeb4bc'}; line-height: 1;"
											></i>
										{/if}
									</button>

									<!-- Leave -->
									<button
										onclick={() => leaveClub(club)}
										disabled={leavingSlug === club.slug}
										class="flex h-[34px] w-[34px] items-center justify-center rounded-full transition-opacity disabled:opacity-40"
										aria-label={$t('settings.clubs_leave')}
									>
										{#if leavingSlug === club.slug}
											<div class="h-4 w-4 animate-spin rounded-full border-2" style="border-color: rgba(174,180,188,0.3); border-top-color: #e74c3c;"></div>
										{:else}
											<i class="fi fi-rr-cross-small" style="font-size: 20px; color: #aeb4bc; line-height: 1;"></i>
										{/if}
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Actions card -->
			<div class="mu-card overflow-hidden rounded-[20px]">
				<p class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">
					{$t('settings.clubs_section_actions')}
				</p>
				<div class="mu-divider flex flex-col" style="border-top-width: 1px; border-top-style: solid;">
					<button
						class="flex items-center gap-3 px-4 py-3 text-left"
						onclick={() => (showSearchSheet = true)}
					>
						<i class="fi fi-rr-search mu-text-primary" style="font-size: 18px;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">{$t('settings.clubs_find_existing')}</span>
					</button>
					<button
						class="mu-divider flex items-center gap-3 px-4 py-3 text-left"
						style="border-top-width: 1px; border-top-style: solid;"
						onclick={() => (showCreateSheet = true)}
					>
						<i class="fi fi-rr-add mu-text-primary" style="font-size: 18px;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">{$t('settings.clubs_create_new')}</span>
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Club search sheet -->
<ClubSearchSheet
	open={showSearchSheet}
	joinedClubIds={joinedClubs.map((c) => c.id)}
	onclose={() => (showSearchSheet = false)}
	onjoined={handleJoined}
/>

<!-- Create club sheet -->
<CreateClubSheet
	open={showCreateSheet}
	coords={null}
	defaultCity={userCity}
	defaultCountry={userCountry}
	onclose={() => (showCreateSheet = false)}
	oncreated={handleCreated}
/>
