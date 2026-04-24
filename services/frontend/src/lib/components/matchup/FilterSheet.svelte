<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { fly } from 'svelte/transition';

	export interface FilterState {
		danceStyles?: string[];
		role?: string;
		distanceKm?: number;
		city?: string;
		ageMin?: number;
		ageMax?: number;
		heightMin?: number;
		heightMax?: number;
		skillLevel?: string;
		availableForCompetitions?: boolean;
		experienceYears?: string;
		trainingFrequency?: string;
	}

	interface Props {
		open?: boolean;
		onclose?: () => void;
		onclear?: () => void;
		onapply?: (filters: FilterState) => void;
		initialFilters?: FilterState;
	}

	let { open = false, onclose, onclear, onapply, initialFilters = {} }: Props = $props();

	let activeSubSheet = $state<string | null>(null);

	let danceStyles = $state<string[]>(initialFilters.danceStyles ?? []);
	let role = $state(initialFilters.role ?? '');
	let distanceKm = $state(initialFilters.distanceKm ?? 50);
	let city = $state(initialFilters.city ?? '');
	let ageMin = $state(initialFilters.ageMin ?? 18);
	let ageMax = $state(initialFilters.ageMax ?? 60);
	let heightMin = $state(initialFilters.heightMin ?? 140);
	let heightMax = $state(initialFilters.heightMax ?? 210);
	let skillLevel = $state(initialFilters.skillLevel ?? '');
	let availableForCompetitions = $state(initialFilters.availableForCompetitions ?? false);
	let experienceYears = $state(initialFilters.experienceYears ?? '');
	let trainingFrequency = $state(initialFilters.trainingFrequency ?? '');

	const DANCE_STYLE_OPTIONS = ['Salsa', 'Bachata', 'Ballroom', 'Latin', 'Swing', 'Tango', 'Jazz', 'Contemporary', 'Hip-hop', 'Zouk'];
	const ROLE_OPTIONS = ['Leader', 'Follower', 'Both'];
	const SKILL_LEVELS = ['Beginner', 'Intermediate', 'Advanced', 'Professional'];
	const EXPERIENCE_OPTIONS = ['< 1 year', '1–3 years', '3–5 years', '5–10 years', '10+ years'];
	const FREQUENCY_OPTIONS = ['Casual', 'Regular', 'Intensive'];

	function toggleDanceStyle(style: string) {
		danceStyles = danceStyles.includes(style)
			? danceStyles.filter((s) => s !== style)
			: [...danceStyles, style];
	}

	function handleApply() {
		onapply?.({ danceStyles, role, distanceKm, city, ageMin, ageMax, heightMin, heightMax, skillLevel, availableForCompetitions, experienceYears, trainingFrequency });
		onclose?.();
	}

	function handleClear() {
		danceStyles = [];
		role = '';
		distanceKm = 50;
		city = '';
		ageMin = 18;
		ageMax = 60;
		heightMin = 140;
		heightMax = 210;
		skillLevel = '';
		availableForCompetitions = false;
		experienceYears = '';
		trainingFrequency = '';
		onclear?.();
	}

	function closeSub() {
		activeSubSheet = null;
	}

	const filterRows: { key: string; label: string; getValue: () => string }[] = [
		{ key: 'danceStyle', label: 'Dance style', getValue: () => danceStyles.length ? danceStyles.join(', ') : '' },
		{ key: 'role', label: 'Role', getValue: () => role },
		{ key: 'age', label: 'Age range', getValue: () => ageMin !== 18 || ageMax !== 60 ? `${ageMin}–${ageMax}` : '' },
		{ key: 'height', label: 'Height range', getValue: () => heightMin !== 140 || heightMax !== 210 ? `${heightMin}–${heightMax} cm` : '' },
		{ key: 'skill', label: 'Skill level', getValue: () => skillLevel },
		{ key: 'distance', label: 'Distance radius', getValue: () => `${distanceKm} km` },
		{ key: 'city', label: 'City / district', getValue: () => city },
		{ key: 'competitions', label: 'Availability for competitions', getValue: () => availableForCompetitions ? 'Yes' : '' },
		{ key: 'experience', label: 'Experience in years', getValue: () => experienceYears },
		{ key: 'frequency', label: 'Open to training frequency', getValue: () => trainingFrequency }
	];
</script>

<BottomSheet {open} {onclose}>
	<!-- Header -->
	<div class="flex items-center justify-between pb-2">
		<button onclick={onclose} class="mu-text-primary text-[14px] font-medium">Cancel</button>
		<span class="mu-text-primary text-[16px] font-bold">Filter</span>
		<button onclick={handleClear} class="text-[14px] font-medium" style="color: #b1b1b1;">Clear all</button>
	</div>

	<!-- Filter rows -->
	<div class="flex flex-col">
		{#each filterRows as row}
			<button
				class="mu-divider flex w-full items-center justify-between py-4"
				style="border-bottom-width: 1px; border-bottom-style: solid;"
				onclick={() => (activeSubSheet = row.key)}
			>
				<span class="mu-text-primary text-[16px] font-bold text-left">{row.label}</span>
				<div class="flex items-center gap-2">
					{#if row.getValue()}
						<span class="text-[13px] font-medium" style="color: #8984da;">{row.getValue()}</span>
					{/if}
					<i class="fi fi-rr-angle-small-right mu-text-primary text-xl leading-none"></i>
				</div>
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

<!-- ── Sub-sheets ── -->

{#if activeSubSheet === 'danceStyle'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Dance style</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4">
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
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Role</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4">
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

{#if activeSubSheet === 'age'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Age range</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col items-center gap-4 pb-6">
				<span class="text-[28px] font-black" style="color: #8984da;">{ageMin} – {ageMax}</span>
				<div class="w-full">
					<label class="mu-text-secondary mb-1 block text-[12px] font-medium">Minimum age: {ageMin}</label>
					<input type="range" min="18" max="60" bind:value={ageMin} class="w-full" style="accent-color: #8984da;" />
				</div>
				<div class="w-full">
					<label class="mu-text-secondary mb-1 block text-[12px] font-medium">Maximum age: {ageMax}</label>
					<input type="range" min="18" max="80" bind:value={ageMax} class="w-full" style="accent-color: #8984da;" />
				</div>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'height'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Height range</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col items-center gap-4 pb-6">
				<span class="text-[28px] font-black" style="color: #8984da;">{heightMin} – {heightMax} cm</span>
				<div class="w-full">
					<label class="mu-text-secondary mb-1 block text-[12px] font-medium">Minimum: {heightMin} cm</label>
					<input type="range" min="140" max="210" bind:value={heightMin} class="w-full" style="accent-color: #8984da;" />
				</div>
				<div class="w-full">
					<label class="mu-text-secondary mb-1 block text-[12px] font-medium">Maximum: {heightMax} cm</label>
					<input type="range" min="140" max="220" bind:value={heightMax} class="w-full" style="accent-color: #8984da;" />
				</div>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'skill'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Skill level</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4">
				{#each SKILL_LEVELS as lvl}
					<button
						class="mu-divider flex items-center justify-between py-3"
						style="border-bottom-width: 1px; border-bottom-style: solid;"
						onclick={() => { skillLevel = lvl; closeSub(); }}
					>
						<span class="mu-text-primary text-[15px] font-semibold">{lvl}</span>
						{#if skillLevel === lvl}
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
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Distance radius</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col items-center gap-4 pb-6">
				<span class="text-[32px] font-black" style="color: #8984da;">{distanceKm} km</span>
				<input type="range" min="1" max="500" bind:value={distanceKm} class="w-full" style="accent-color: #8984da;" />
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
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
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
					autofocus
					class="mu-text-primary w-full px-4 py-3 text-[14px] font-medium outline-none"
					style="border: 1.5px solid var(--mu-text); border-radius: 50px; background: transparent;"
				/>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'competitions'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Competitions</span>
				<div class="w-12"></div>
			</div>
			<div class="flex items-center justify-between pb-6 py-3">
				<span class="mu-text-primary text-[15px] font-semibold">Available for competitions</span>
				<button
					onclick={() => (availableForCompetitions = !availableForCompetitions)}
					class="relative flex items-center transition-colors"
					style="width: 50px; height: 28px; border-radius: 50px; background: {availableForCompetitions ? '#8984da' : '#d1d5db'};"
					role="switch"
					aria-checked={availableForCompetitions}
				>
					<div class="absolute h-[22px] w-[22px] rounded-full bg-white shadow-sm transition-transform"
						style="transform: translateX({availableForCompetitions ? '25px' : '3px'});"></div>
				</button>
			</div>
		</div>
	</div>
{/if}

{#if activeSubSheet === 'experience'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Experience</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4">
				{#each EXPERIENCE_OPTIONS as exp}
					<button
						class="mu-divider flex items-center justify-between py-3"
						style="border-bottom-width: 1px; border-bottom-style: solid;"
						onclick={() => { experienceYears = exp; closeSub(); }}
					>
						<span class="mu-text-primary text-[15px] font-semibold">{exp}</span>
						{#if experienceYears === exp}
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

{#if activeSubSheet === 'frequency'}
	<div class="fixed inset-0 z-[200]" transition:fly={{ y: 400, duration: 300 }}>
		<div class="absolute inset-0" style="background: rgba(0,0,0,0.4);" onclick={closeSub} role="presentation"></div>
		<div class="mu-sheet absolute right-0 bottom-0 left-0 rounded-t-[20px] px-4 pb-safe" style="padding-top: 12px;">
			<div class="mx-auto mb-4 h-[4px] w-[40px] rounded-full" style="background: var(--mu-handle);"></div>
			<div class="flex items-center justify-between pb-4">
				<button onclick={closeSub} class="mu-text-primary text-[14px] font-medium">Done</button>
				<span class="mu-text-primary text-[16px] font-bold">Training frequency</span>
				<div class="w-12"></div>
			</div>
			<div class="flex flex-col pb-4">
				{#each FREQUENCY_OPTIONS as freq}
					<button
						class="mu-divider flex items-center justify-between py-3"
						style="border-bottom-width: 1px; border-bottom-style: solid;"
						onclick={() => { trainingFrequency = freq; closeSub(); }}
					>
						<span class="mu-text-primary text-[15px] font-semibold">{freq}</span>
						{#if trainingFrequency === freq}
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
