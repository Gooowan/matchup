<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';

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
	}

	interface Props {
		open?: boolean;
		onclose?: () => void;
		onclear?: () => void;
		onapply?: (filters: FilterState) => void;
		initialFilters?: FilterState;
	}

	let { open = false, onclose, onclear, onapply, initialFilters = {} }: Props = $props();

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
	});

	const FINANCE_OPTIONS: { value: string; label: string }[] = [
		{ value: 'no', label: 'Ні' },
		{ value: 'yes', label: 'Так' },
		{ value: 'partial', label: 'Частково' }
	];
	const UKRAINE_CITIES = ['Київ', 'Харків', 'Одеса', 'Дніпро', 'Запоріжжя', 'Львів', 'Кривий Ріг', 'Миколаїв', 'Вінниця', 'Херсон', 'Полтава', 'Чернігів', 'Черкаси', 'Суми', 'Житомир', 'Хмельницький', 'Рівне', 'Тернопіль', 'Луцьк', 'Ужгород'];
	const CATEGORIES_UA = [
		{ value: 'kids', label: 'Діти' },
		{ value: 'juvenile1', label: 'Ювенали 1' },
		{ value: 'juvenile2', label: 'Ювенали 2' },
		{ value: 'junior1', label: 'Юніори 1' },
		{ value: 'junior2', label: 'Юніори 2' },
		{ value: 'youth', label: 'Молодь' },
		{ value: 'adult', label: 'Дорослі' }
	];

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
		if (goal) f.goal = goal;
		if (program) f.program = program;
		if (categories.length) f.categories = categories;
		if (city) f.city = city;
		if (wantsPartnerToFinance) f.wantsPartnerToFinance = wantsPartnerToFinance;
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
		onclear?.();
	}
</script>

<BottomSheet {open} {onclose}>
	<!-- Header -->
	<div class="flex items-center justify-between pb-4">
		<button onclick={onclose} class="mu-text-primary text-[14px] font-medium">Скасувати</button>
		<span class="mu-text-primary text-[16px] font-bold">Фільтри</span>
		<button onclick={handleClear} class="text-[14px] font-medium" style="color: #b1b1b1;">Скинути</button>
	</div>

	<div class="flex flex-col gap-4 pb-6">
		<!-- Gender -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ШУКАЮ</label>
			<div class="flex gap-2">
				{#each [{ value: 'male', label: 'Чоловік' }, { value: 'female', label: 'Жінка' }] as g}
					<button
						onclick={() => (gender = gender === g.value ? '' : g.value)}
						class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
						style="background: {gender === g.value ? '#8984da' : 'transparent'}; color: {gender === g.value ? 'white' : '#696969'}; border: 1.5px solid {gender === g.value ? '#8984da' : '#d1d5db'};"
					>{g.label}</button>
				{/each}
			</div>
		</div>

		<!-- Age -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ВІК ПАРТНЕРА</label>
			<div class="flex items-center gap-3">
				<input
					type="number"
					placeholder="Від"
					bind:value={ageMin}
					min="4" max="80"
					class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
					style="color: #171717; border-color: #e0e0e0;"
				/>
				<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
				<input
					type="number"
					placeholder="До"
					bind:value={ageMax}
					min="4" max="80"
					class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
					style="color: #171717; border-color: #e0e0e0;"
				/>
			</div>
		</div>

		<!-- Height -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ЗРІСТ ПАРТНЕРА (СМ)</label>
			<div class="flex items-center gap-3">
				<input
					type="number"
					placeholder="Від"
					bind:value={heightMin}
					min="100" max="220"
					class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
					style="color: #171717; border-color: #e0e0e0;"
				/>
				<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
				<input
					type="number"
					placeholder="До"
					bind:value={heightMax}
					min="100" max="220"
					class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
					style="color: #171717; border-color: #e0e0e0;"
				/>
			</div>
		</div>

		<!-- Goal -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">БАЖАНА ЦІЛЬ</label>
			<div class="flex gap-2">
				{#each [{ value: 'hobby', label: 'Хобі' }, { value: 'professional', label: 'Профі' }] as g}
					<button
						onclick={() => (goal = goal === g.value ? '' : g.value)}
						class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
						style="background: {goal === g.value ? '#8984da' : 'transparent'}; color: {goal === g.value ? 'white' : '#696969'}; border: 1.5px solid {goal === g.value ? '#8984da' : '#d1d5db'};"
					>{g.label}</button>
				{/each}
			</div>
		</div>

		<!-- Program -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРОГРАМА</label>
			<div class="flex gap-2">
				{#each [{ value: 'standard', label: 'Стандарт' }, { value: 'latina', label: 'Латина' }, { value: 'both', label: 'Обидва' }] as p}
					<button
						onclick={() => (program = program === p.value ? '' : p.value)}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {program === p.value ? '#8984da' : 'transparent'}; color: {program === p.value ? 'white' : '#696969'}; border: 1.5px solid {program === p.value ? '#8984da' : '#d1d5db'};"
					>{p.label}</button>
				{/each}
			</div>
		</div>

		<!-- Categories -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КАТЕГОРІЯ</label>
			<div class="flex flex-wrap gap-2">
				{#each CATEGORIES_UA as cat}
					<button
						onclick={() => toggleCategory(cat.value)}
						class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
						style="background: {categories.includes(cat.value) ? '#8984da' : 'transparent'}; color: {categories.includes(cat.value) ? 'white' : '#696969'}; border: 1.5px solid {categories.includes(cat.value) ? '#8984da' : '#d1d5db'};"
					>{cat.label}</button>
				{/each}
			</div>
		</div>

		<!-- City -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">МІСТО ПОШУКУ</label>
			<select
				bind:value={city}
				class="w-full bg-transparent text-[16px] font-semibold outline-none"
				style="color: {city ? '#171717' : '#aeb4bc'}; -webkit-appearance: none; appearance: none;"
			>
				<option value="">Будь-яке місто</option>
				{#each UKRAINE_CITIES as c}
					<option value={c}>{c}</option>
				{/each}
			</select>
		</div>

		<!-- Partner finance -->
		<div style="display: flex; flex-direction: column; gap: 10px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПАРТНЕР ГОТОВИЙ ФІНАНСУВАТИ</label>
			<div class="flex gap-2">
				{#each FINANCE_OPTIONS as opt}
					<button
						onclick={() => (wantsPartnerToFinance = wantsPartnerToFinance === opt.value ? '' : opt.value)}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {wantsPartnerToFinance === opt.value ? '#8984da' : 'transparent'}; color: {wantsPartnerToFinance === opt.value ? 'white' : '#696969'}; border: 1.5px solid {wantsPartnerToFinance === opt.value ? '#8984da' : '#d1d5db'};"
					>{opt.label}</button>
				{/each}
			</div>
		</div>
	</div>

	<!-- Apply button -->
	<button
		onclick={handleApply}
		class="w-full py-3 text-[14px] font-semibold text-white"
		style="border-radius: 50px; background: #8984da;"
	>
		Застосувати
	</button>
</BottomSheet>
