<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { fly } from 'svelte/transition';

	export interface FilterState {
		danceStyles: string[];
		role: string;
		distanceKm: number;
		city: string;
	}

	interface Props {
		open?: boolean;
		onclose?: () => void;
		onclear?: () => void;
		onapply?: (filters: FilterState) => void;
		initialFilters?: Partial<FilterState>;
	}

	let { open = false, onclose, onclear, onapply, initialFilters = {} }: Props = $props();

	let activeSubSheet = $state<string | null>(null);

	let danceStyles = $state<string[]>(initialFilters.danceStyles ?? []);
	let role = $state(initialFilters.role ?? '');
	let distanceKm = $state(initialFilters.distanceKm ?? 50);
	let city = $state(initialFilters.city ?? '');

	const DANCE_STYLE_OPTIONS = ['Salsa', 'Bachata', 'Ballroom', 'Latin', 'Swing', 'Tango', 'Jazz', 'Contemporary'];
	const ROLE_OPTIONS = ['Leader', 'Follower', 'Both'];

	function toggleDanceStyle(style: string) {
		if (danceStyles.includes(style)) {
			danceStyles = danceStyles.filter((s) => s !== style);
		} else {
			danceStyles = [...danceStyles, style];
		}
	}

	function handleApply() {
		onapply?.({ danceStyles, role, distanceKm, city });
		onclose?.();
	}

	function handleClear() {
		danceStyles = [];
		role = '';
		distanceKm = 50;
		city = '';
		onclear?.();
	}

	function closeSub() {
		activeSubSheet = null;
	}

	const otherFilters = [
		'Age range',
		'Height range',
		'Skill level',
		'Availability for competitions',
		'Experience in years',
		'Open to training frequency'
	];
</script>

<BottomSheet {open} {onclose}>
	<!-- Header -->
	<div class="flex items-center justify-between pb-2">
		<button onclick={onclose} class="mu-text-primary text-[14px] font-medium">
			Cancel
		</button>
		<span class="mu-text-primary text-[16px] font-bold">Filter</span>
		<button onclick={handleClear} class="text-[14px] font-medium" style="color: #b1b1b1;">
			Clear all
		</button>
	</div>

	<!-- Filter rows -->
	<div class="flex flex-col">
		<!-- Dance style -->
		<button
			class="mu-divider flex w-full items-center justify-between py-4"
			style="border-bottom-width: 1px; border-bottom-style: solid;"
			onclick={() => (activeSubSheet = 'danceStyle')}
		>
			<span class="mu-text-primary text-[16px] font-bold">Dance style</span>
			<div class="flex items-center gap-2">
				{#if danceStyles.length > 0}
					<span class="text-[13px] font-medium" style="color: #8984da;"
						>{danceStyles.join(', ')}</span
					>
				{/if}
				<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
			</div>
		</button>

		<!-- Role -->
		<button
			class="mu-divider flex w-full items-center justify-between py-4"
			style="border-bottom-width: 1px; border-bottom-style: solid;"
			onclick={() => (activeSubSheet = 'role')}
		>
			<span class="mu-text-primary text-[16px] font-bold">Role</span>
			<div class="flex items-center gap-2">
				{#if role}
					<span class="text-[13px] font-medium" style="color: #8984da;">{role}</span>
				{/if}
				<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
			</div>
		</button>

		<!-- Distance radius -->
		<button
			class="mu-divider flex w-full items-center justify-between py-4"
			style="border-bottom-width: 1px; border-bottom-style: solid;"
			onclick={() => (activeSubSheet = 'distance')}
		>
			<span class="mu-text-primary text-[16px] font-bold">Distance radius</span>
			<div class="flex items-center gap-2">
				<span class="text-[13px] font-medium" style="color: #8984da;">{distanceKm} km</span>
				<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
			</div>
		</button>

		<!-- City -->
		<button
			class="mu-divider flex w-full items-center justify-between py-4"
			style="border-bottom-width: 1px; border-bottom-style: solid;"
			onclick={() => (activeSubSheet = 'city')}
		>
			<span class="mu-text-primary text-[16px] font-bold">City / district</span>
			<div class="flex items-center gap-2">
				{#if city}
					<span class="text-[13px] font-medium" style="color: #8984da;">{city}</span>
				{/if}
				<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
			</div>
		</button>

		<!-- Other filters (coming soon) -->
		{#each otherFilters as filter}
			<button
				class="mu-divider flex w-full items-center justify-between py-4"
				style="border-bottom-width: 1px; border-bottom-style: solid;"
				onclick={() => (activeSubSheet = 'soon')}
			>
				<span class="mu-text-primary text-[16px] font-bold text-left">{filter}</span>
				<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
			</button>
		{/each}
	</div>

	<!-- Apply button -->
	<button
		onclick={handleApply}
		class="mt-4 w-full py-3 text-[14px] font-semibold text-white"
		style="border-radius: 50px; background: #696969;"
	>
		Apply filters
	</button>
</BottomSheet>

<!-- Sub-sheets rendered outside BottomSheet to layer on top -->
{#if activeSubSheet === 'danceStyle'}
	<div
		class="fixed inset-0 z-[200]"
		transition:fly={{ y: 400, duration: 300 }}
	>
		<div
			class="absolute inset-0"
			style="background: rgba(0,0,0,0.4);"
			onclick={closeSub}
			role="presentation"
		></div>
		<div
			class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe"
			style="padding-top: 12px;"
		>
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Dance style</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4" style="gap: 0;">
				{#each DANCE_STYLE_OPTIONS as style}
					<button
						class="mu-divider flex items-center justify-between py-3"
						style="border-bottom-width: 1px; border-bottom-style: solid;"
						onclick={() => toggleDanceStyle(style)}
					>
						<span class="mu-text-primary text-[15px] font-semibold">{style}</span>
						{#if danceStyles.includes(style)}
							<i class="fi fi-sr-check-circle text-xl" style="color: #8984da;"></i>
						{:else}
							<div class="h-5 w-5 rounded-full" style="border: 2px solid #d1d5db;"></div>
						{/if}
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'role'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div
			class="absolute inset-0"
			style="background: rgba(0,0,0,0.4);"
			onclick={closeSub}
			role="presentation"
		></div>
		<div
			class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe"
			style="padding-top: 12px;"
		>
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Role</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4" style="gap: 0;">
				{#each ROLE_OPTIONS as r}
					<button
						class="mu-divider flex items-center justify-between py-3"
						style="border-bottom-width: 1px; border-bottom-style: solid;"
						onclick={() => { role = r; closeSub(); }}
					>
						<span class="mu-text-primary text-[15px] font-semibold">{r}</span>
						{#if role === r}
							<i class="fi fi-sr-check-circle text-xl" style="color: #8984da;"></i>
						{:else}
							<div class="h-5 w-5 rounded-full" style="border: 2px solid #d1d5db;"></div>
						{/if}
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'distance'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div
			class="absolute inset-0"
			style="background: rgba(0,0,0,0.4);"
			onclick={closeSub}
			role="presentation"
		></div>
		<div
			class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe"
			style="padding-top: 12px;"
		>
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Distance radius</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col items-center pb-6 gap-4">
				<span class="text-[32px] font-black" style="color: #8984da;">{distanceKm} km</span>
				<input
					type="range"
					min="1"
					max="500"
					bind:value={distanceKm}
					class="w-full"
					style="accent-color: #8984da;"
				/>
				<div class="flex w-full justify-between">
					<span class="text-[12px] font-medium" style="color: #aeb4bc;">1 km</span>
					<span class="text-[12px] font-medium" style="color: #aeb4bc;">500 km</span>
				</div>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'city'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div
			class="absolute inset-0"
			style="background: rgba(0,0,0,0.4);"
			onclick={closeSub}
			role="presentation"
		></div>
		<div
			class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe"
			style="padding-top: 12px;"
		>
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">City / district</span>
				<div class="w-12"></div>
			</div>
			<div class="pb-6">
				<input
					type="text"
					placeholder="Enter city or district"
					bind:value={city}
					class="mu-text-primary w-full px-4 py-3 text-[14px] font-medium outline-none"
					style="border: 1.5px solid var(--mu-text); border-radius: 50px; background: transparent;"
				/>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'soon'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div
			class="absolute inset-0"
			style="background: rgba(0,0,0,0.4);"
			onclick={closeSub}
			role="presentation"
		></div>
		<div
			class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe flex flex-col items-center"
			style="padding-top: 12px; padding-bottom: 40px;"
		>
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<p class="mu-text-primary mt-4 text-[16px] font-bold">Coming soon</p>
			<p class="mu-text-secondary mt-1 text-[13px] font-medium">This filter will be available shortly</p>
		</div>
	</div>
{/if}
