<script lang="ts">
	import { browser } from '$app/environment';
	import BottomSheet from './BottomSheet.svelte';
	import { t } from '$lib/locale';

	export interface FilterState {
		gender?: string;
		ageMin?: number;
		ageMax?: number;
		heightMin?: number;
		heightMax?: number;
		goal?: string;
		program?: string;
		categories?: string[];
		city?: string;
		wantsPartnerToFinance?: string;
		wantsPartnerToRelocate?: boolean;
	}

	interface Props {
		open?: boolean;
		onclose?: () => void;
		onclear?: () => void;
		onapply?: (filters: FilterState) => void;
		initialFilters?: FilterState;
		/** User's own goal ("hobby" = amateur, "professional" = pro). Determines which filters to show. */
		userGoal?: string;
	}

	let { open = false, onclose, onclear, onapply, initialFilters = {}, userGoal = 'professional' }: Props = $props();

	let isAmateur = $derived(userGoal === 'hobby');

	let isDark = $state(browser && document.documentElement.classList.contains('dark'));
	$effect(() => {
		if (!browser) return;
		isDark = document.documentElement.classList.contains('dark');
		const obs = new MutationObserver(() => {
			isDark = document.documentElement.classList.contains('dark');
		});
		obs.observe(document.documentElement, { attributeFilter: ['class'] });
		return () => obs.disconnect();
	});

	// Theme-aware color tokens
	const inputTextColor  = $derived(isDark ? '#ffffff' : '#171717');
	const inputBorderColor = $derived(isDark ? '#3a3a3e' : '#e0e0e0');
	const inputBgColor    = $derived(isDark ? '#2a2a2e' : 'transparent');
	const btnIdleColor    = $derived(isDark ? '#e0e0e0' : '#696969');
	const btnIdleBorder   = $derived(isDark ? '#3a3a3e' : '#d1d5db');

	let gender = $state('');
	let ageMin = $state<number | null>(null);
	let ageMax = $state<number | null>(null);
	let heightMin = $state<number | null>(null);
	let heightMax = $state<number | null>(null);
	let goal = $state('');
	let program = $state('');
	let categories = $state<string[]>([]);
	let city = $state('');
	let wantsPartnerToFinance = $state('');
	let wantsPartnerToRelocate = $state<boolean | null>(null);
	let showAdvanced = $state(false);

	$effect(() => {
		if (!open) return;
		const i = initialFilters;
		gender = i.gender ?? '';
		ageMin = i.ageMin ?? null;
		ageMax = i.ageMax ?? null;
		heightMin = i.heightMin ?? null;
		heightMax = i.heightMax ?? null;
		goal = i.goal ?? '';
		program = i.program ?? '';
		categories = i.categories ? [...i.categories] : [];
		city = i.city ?? '';
		wantsPartnerToFinance = i.wantsPartnerToFinance ?? '';
		wantsPartnerToRelocate = i.wantsPartnerToRelocate ?? null;
	});

	const FINANCE_OPTIONS = $derived<{ value: string; label: string }[]>([
		{ value: 'no', label: $t('filters.finance_no') },
		{ value: 'yes', label: $t('filters.finance_yes') },
		{ value: 'partial', label: $t('filters.finance_partial') }
	]);
	// v1: locked to Київ. FUTURE multi-city: restore UKRAINE_CITIES and the <select> in the template.
	const HARDCODED_CITY = 'Київ';
	/* const UKRAINE_CITIES = ['Київ', 'Харків', 'Одеса', 'Дніпро', 'Запоріжжя', 'Львів', 'Кривий Ріг', 'Миколаїв', 'Вінниця', 'Херсон', 'Полтава', 'Чернігів', 'Черкаси', 'Суми', 'Житомир', 'Хмельницький', 'Рівне', 'Тернопіль', 'Луцьк', 'Ужгород']; */
	const CATEGORIES_UA = $derived([
		{ value: 'kids', label: $t('filters.category_kids') },
		{ value: 'juvenile1', label: $t('filters.category_juvenile1') },
		{ value: 'juvenile2', label: $t('filters.category_juvenile2') },
		{ value: 'junior1', label: $t('filters.category_junior1') },
		{ value: 'junior2', label: $t('filters.category_junior2') },
		{ value: 'youth', label: $t('filters.category_youth') },
		{ value: 'adult', label: $t('filters.category_adult') }
	]);

	function toggleCategory(v: string) {
		categories = categories.includes(v) ? categories.filter((x) => x !== v) : [...categories, v];
	}

	function handleApply() {
		const f: FilterState = {};
		if (gender) f.gender = gender;
		if (ageMin != null) f.ageMin = ageMin;
		if (ageMax != null) f.ageMax = ageMax;
		if (heightMin != null) f.heightMin = heightMin;
		if (heightMax != null) f.heightMax = heightMax;
		if (program) f.program = program;
		f.city = HARDCODED_CITY; // v1 locked; FUTURE: replace with `if (city) f.city = city;`
		if (!isAmateur) {
			if (goal) f.goal = goal;
			if (categories.length && goal !== 'hobby') f.categories = categories;
			if (wantsPartnerToFinance) f.wantsPartnerToFinance = wantsPartnerToFinance;
			if (wantsPartnerToRelocate !== null) f.wantsPartnerToRelocate = wantsPartnerToRelocate;
		}
		onapply?.(f);
		onclose?.();
	}

	function handleClear() {
		gender = '';
		ageMin = null;
		ageMax = null;
		heightMin = null;
		heightMax = null;
		goal = '';
		program = '';
		categories = [];
		city = '';
		wantsPartnerToFinance = '';
		wantsPartnerToRelocate = null;
		onclear?.();
	}
</script>

<BottomSheet {open} {onclose}>
	<!-- Header -->
	<div class="flex items-center justify-between pb-4">
		<button onclick={onclose} class="mu-text-primary text-[14px] font-medium">{$t('filters.cancel')}</button>
		<span class="mu-text-primary text-[16px] font-bold">{$t('filters.title')}</span>
		<button onclick={handleClear} class="text-[14px] font-medium" style="color: #b1b1b1;">{$t('filters.reset')}</button>
	</div>

	<div class="flex flex-col gap-4 pb-6">
		<!-- Gender -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_gender')}</label>
			<div class="flex gap-2">
			{#each [{ value: 'male', label: $t('filters.male') }, { value: 'female', label: $t('filters.female') }] as g}
				<button
					onclick={() => (gender = gender === g.value ? '' : g.value)}
					class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
					style="background: {gender === g.value ? '#8984da' : 'transparent'}; color: {gender === g.value ? 'white' : btnIdleColor}; border: 1.5px solid {gender === g.value ? '#8984da' : btnIdleBorder};"
				>{g.label}</button>
			{/each}
			</div>
		</div>

		<!-- Age -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_age')}</label>
			<div class="flex items-center gap-3">
			<input
				type="number"
				placeholder={$t('filters.from')}
				bind:value={ageMin}
				min="4" max="80"
				class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
				style="color: {inputTextColor}; border-color: {inputBorderColor}; background: {inputBgColor};"
			/>
			<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
			<input
				type="number"
				placeholder={$t('filters.to')}
				bind:value={ageMax}
				min="4" max="80"
				class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
				style="color: {inputTextColor}; border-color: {inputBorderColor}; background: {inputBgColor};"
			/>
			</div>
		</div>

		<!-- Height -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_height')}</label>
			<div class="flex items-center gap-3">
			<input
				type="number"
				placeholder={$t('filters.from')}
				bind:value={heightMin}
				min="100" max="220"
				class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
				style="color: {inputTextColor}; border-color: {inputBorderColor}; background: {inputBgColor};"
			/>
			<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
			<input
				type="number"
				placeholder={$t('filters.to')}
				bind:value={heightMax}
				min="100" max="220"
				class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
				style="color: {inputTextColor}; border-color: {inputBorderColor}; background: {inputBgColor};"
			/>
			</div>
		</div>

		{#if !isAmateur}
			<!-- Goal (pro only) -->
			<div style="display: flex; flex-direction: column; gap: 10px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_goal')}</label>
				<div class="flex gap-2">
					{#each [{ value: 'hobby', label: $t('filters.goal_hobby') }, { value: 'professional', label: $t('filters.goal_professional') }] as g}
						<button
							onclick={() => {
								goal = goal === g.value ? '' : g.value;
								if (goal === 'hobby') categories = [];
							}}
					class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
						style="background: {goal === g.value ? '#8984da' : 'transparent'}; color: {goal === g.value ? 'white' : btnIdleColor}; border: 1.5px solid {goal === g.value ? '#8984da' : btnIdleBorder};"
					>{g.label}</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Program -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_program')}</label>
			<div class="flex gap-2">
				{#each [{ value: 'standard', label: $t('filters.program_standard') }, { value: 'latina', label: $t('filters.program_latina') }, { value: 'both', label: $t('filters.program_both') }] as p}
					<button
						onclick={() => (program = program === p.value ? '' : p.value)}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {program === p.value ? '#8984da' : 'transparent'}; color: {program === p.value ? 'white' : btnIdleColor}; border: 1.5px solid {program === p.value ? '#8984da' : btnIdleBorder};"
					>{p.label}</button>
				{/each}
			</div>
		</div>

	{#if !isAmateur && goal !== 'hobby'}
		<!-- Categories: hidden when own goal is hobby OR when selected goal filter is hobby -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_categories')}</label>
			<div class="flex flex-wrap gap-2">
				{#each CATEGORIES_UA as cat}
					<button
						onclick={() => toggleCategory(cat.value)}
						class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
						style="background: {categories.includes(cat.value) ? '#8984da' : 'transparent'}; color: {categories.includes(cat.value) ? 'white' : btnIdleColor}; border: 1.5px solid {categories.includes(cat.value) ? '#8984da' : btnIdleBorder};"
					>{cat.label}</button>
				{/each}
			</div>
		</div>
	{/if}

	<!-- City — v1 locked to Київ. FUTURE: restore the <select> below and remove the locked display. -->
	<div style="display: flex; flex-direction: column; gap: 10px;">
		<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_city')}</label>
		<div class="flex items-center justify-between">
			<span class="text-[16px] font-semibold" style="color: {inputTextColor};">Київ</span>
			<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
		</div>
		<!-- FUTURE multi-city: restore this select and remove the locked display above
		<select
			bind:value={city}
			class="w-full bg-transparent text-[16px] font-semibold outline-none"
			style="color: {city ? inputTextColor : '#aeb4bc'}; -webkit-appearance: none; appearance: none;"
		>
			<option value="">{$t('filters.any_city')}</option>
			{#each UKRAINE_CITIES as c}
				<option value={c}>{c}</option>
			{/each}
		</select>
		-->
	</div>

		{#if !isAmateur}
			<!-- Partner finance (pro only) -->
			<div style="display: flex; flex-direction: column; gap: 10px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_finance')}</label>
				<div class="flex gap-2">
					{#each FINANCE_OPTIONS as opt}
						<button
							onclick={() => (wantsPartnerToFinance = wantsPartnerToFinance === opt.value ? '' : opt.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
							style="background: {wantsPartnerToFinance === opt.value ? '#8984da' : 'transparent'}; color: {wantsPartnerToFinance === opt.value ? 'white' : btnIdleColor}; border: 1.5px solid {wantsPartnerToFinance === opt.value ? '#8984da' : btnIdleBorder};"
						>{opt.label}</button>
					{/each}
				</div>
			</div>

			<!-- Advanced toggle (pro only) -->
			<button
				onclick={() => (showAdvanced = !showAdvanced)}
				class="flex items-center gap-2 text-[13px] font-semibold"
				style="color: #8984da;"
			>
				<i class="fi fi-rr-settings-sliders" style="font-size: 14px; line-height: 1;"></i>
				{showAdvanced ? $t('filters.hide_advanced') : $t('filters.show_advanced')}
			</button>

			{#if showAdvanced}
				<!-- Relocation preference -->
				<div style="display: flex; flex-direction: column; gap: 10px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('filters.section_relocate')}</label>
					<div class="flex gap-2">
						{#each [{ value: true, label: $t('filters.relocate_yes') }, { value: false, label: $t('filters.relocate_no') }] as opt}
							<button
								onclick={() => (wantsPartnerToRelocate = wantsPartnerToRelocate === opt.value ? null : opt.value)}
								class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
								style="background: {wantsPartnerToRelocate === opt.value ? '#8984da' : 'transparent'}; color: {wantsPartnerToRelocate === opt.value ? 'white' : btnIdleColor}; border: 1.5px solid {wantsPartnerToRelocate === opt.value ? '#8984da' : btnIdleBorder};"
							>{opt.label}</button>
						{/each}
					</div>
				</div>
			{/if}
		{/if}
	</div>

	{#snippet footer()}
		<button
			onclick={handleApply}
			class="w-full py-3 text-[14px] font-semibold text-white"
			style="border-radius: 50px; background: #8984da;"
		>
			{$t('filters.apply')}
		</button>
	{/snippet}
</BottomSheet>
