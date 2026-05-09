<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { authFetch } from '$lib/utils/authFetch';
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

	interface Props {
		open?: boolean;
		joinedClubIds?: string[];
		onclose?: () => void;
		onjoined?: (club: Club) => void;
	}

	let { open = false, joinedClubIds = [], onclose, onjoined }: Props = $props();

	let query = $state('');
	let results = $state<Club[]>([]);
	let isSearching = $state(false);
	let joiningSlug = $state<string | null>(null);
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		if (open) {
			query = '';
			results = [];
			doSearch();
		} else {
			query = '';
			results = [];
			isSearching = false;
		}
	});

	function handleInput() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(doSearch, 300);
	}

	async function doSearch() {
		isSearching = true;
		try {
			const params = new URLSearchParams({ q: query.trim(), limit: '20' });
			const resp = await fetch(`${import.meta.env.VITE_API_URL}/clubs?${params}`);
			if (resp.ok) {
				const body = await resp.json();
				const all = (body.data ?? []) as Club[];
				results = all.filter((c) => !joinedClubIds.includes(c.id));
			}
		} catch {
			results = [];
		} finally {
			isSearching = false;
		}
	}

	async function joinClub(club: Club) {
		joiningSlug = club.slug;
		try {
			const resp = await authFetch(`/clubs/${club.slug}/join`, { method: 'POST' });
			if (resp.ok) {
				onjoined?.(club);
				onclose?.();
			} else {
				const err = await resp.json().catch(() => ({}));
				toast.error((err as { error?: string }).error || $t('settings.clubs_join_error'));
			}
		} catch {
			toast.error($t('settings.clubs_join_error'));
		} finally {
			joiningSlug = null;
		}
	}
</script>

<BottomSheet {open} onclose={onclose}>
	<div class="pb-2">
		<h2 class="mu-text-primary mb-4 text-[18px] font-black">{$t('settings.clubs_find_existing')}</h2>

		<!-- Search input -->
		<div
			class="mu-card mu-border mb-4 flex items-center gap-3 rounded-[14px]"
			style="padding: 10px 14px; border-width: 1px; border-style: solid;"
		>
			<i class="fi fi-rr-search" style="font-size: 18px; color: #aeb4bc; line-height: 1; flex-shrink: 0;"></i>
			<input
				type="search"
				bind:value={query}
				oninput={handleInput}
				placeholder={$t('settings.clubs_search_placeholder')}
				class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
			/>
			{#if isSearching}
				<div class="h-4 w-4 flex-shrink-0 animate-spin rounded-full border-2" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
			{/if}
		</div>

		<!-- Results -->
		{#if results.length > 0}
			<div class="flex flex-col" style="gap: 2px;">
				{#each results as club}
					<button
						onclick={() => joinClub(club)}
						disabled={joiningSlug === club.slug}
						class="mu-card flex items-center gap-3 rounded-[14px] px-4 py-3 text-left transition-opacity disabled:opacity-60"
					>
						<div class="flex h-[40px] w-[40px] flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.15);">
							<i class="fi fi-rr-bank" style="font-size: 18px; color: #8984da;"></i>
						</div>
						<div class="flex flex-1 flex-col gap-0.5 overflow-hidden">
							<span class="mu-text-primary truncate text-[14px] font-semibold">{club.name}</span>
							<span class="truncate text-[12px] font-medium" style="color: #aeb4bc;">{[club.city, club.country].filter(Boolean).join(', ')}</span>
						</div>
						{#if joiningSlug === club.slug}
							<div class="h-4 w-4 animate-spin rounded-full border-2 flex-shrink-0" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
						{:else}
							<i class="fi fi-rr-add flex-shrink-0" style="font-size: 18px; color: #8984da;"></i>
						{/if}
					</button>
				{/each}
			</div>
		{:else if !isSearching}
			<p class="py-6 text-center text-[13px] font-medium" style="color: #aeb4bc;">{$t('settings.clubs_no_results')}</p>
		{/if}
	</div>
</BottomSheet>
