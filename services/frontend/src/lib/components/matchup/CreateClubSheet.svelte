<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	const COUNTRIES = [
		'Україна', 'Польща', 'Германія', 'Чехія', 'Австрія', 'Угорщина', 'Румунія',
		'Словаччина', 'Болгарія', 'Хорватія', 'Франція', 'Іспанія', 'Португалія',
		'Нідерланди', 'Бельгія', 'Швейцарія', 'Велика Британія', 'Ірландія',
		'Швеція', 'Норвегія', 'Данія', 'Фінляндія', 'Естонія', 'Латвія', 'Литва',
		'США', 'Канада', 'Австралія'
	];
	const CITIES_BY_COUNTRY: Record<string, string[]> = {
		'Україна': ['Київ', 'Харків', 'Одеса', 'Дніпро', 'Запоріжжя', 'Львів', 'Кривий Ріг', 'Миколаїв', 'Вінниця', 'Херсон', 'Полтава', 'Чернігів', 'Черкаси', 'Суми', 'Житомир', 'Хмельницький', 'Рівне', 'Тернопіль', 'Луцьк', 'Ужгород'],
		'Польща': ['Варшава', 'Краків', 'Вроцлав', 'Познань', 'Гданськ', 'Лодзь', 'Катовіце', 'Люблін'],
		'Германія': ['Берлін', 'Гамбург', 'Мюнхен', 'Кельн', 'Франкфурт', 'Штутгарт', 'Дюссельдорф', 'Лейпциг'],
		'Чехія': ['Прага', 'Брно', 'Острава', 'Пльзень'],
		'Австрія': ['Відень', 'Грац', 'Лінц', 'Зальцбург'],
		'Угорщина': ['Будапешт', 'Дебрецен', 'Мішкольц', 'Печ'],
		'Румунія': ['Бухарест', 'Клуж-Напока', 'Тімішоара', 'Яси'],
		'Словаччина': ['Братислава', 'Кошіце', 'Прешов', 'Жіліна'],
		'Болгарія': ['Софія', 'Пловдив', 'Варна', 'Бургас'],
		'Хорватія': ['Загреб', 'Спліт', 'Рієка', 'Осієк'],
		'Франція': ['Париж', 'Марсель', 'Ліон', 'Тулуза', 'Ніцца', 'Нант'],
		'Іспанія': ['Мадрид', 'Барселона', 'Валенсія', 'Севілья', 'Більбао'],
		'Португалія': ['Лісабон', 'Порту', 'Брага', 'Коїмбра'],
		'Нідерланди': ['Амстердам', 'Роттердам', 'Гаага', 'Утрехт'],
		'Бельгія': ['Брюссель', 'Антверпен', 'Гент', 'Брюгге'],
		'Швейцарія': ['Цюріх', 'Женева', 'Базель', 'Берн'],
		'Велика Британія': ['Лондон', 'Манчестер', 'Бірмінгем', 'Глазго', 'Лідс'],
		'Ірландія': ['Дублін', 'Корк', 'Голуей', 'Лімерік'],
		'Швеція': ['Стокгольм', 'Гетеборг', 'Мальме', 'Упсала'],
		'Норвегія': ['Осло', 'Берген', 'Трондгейм', 'Ставангер'],
		'Данія': ['Копенгаген', 'Орхус', 'Оденсе', 'Ольборг'],
		'Фінляндія': ['Гельсінкі', 'Тампере', 'Турку', 'Оулу'],
		'Естонія': ['Таллінн', 'Тарту', 'Нарва', 'Пярну'],
		'Латвія': ['Рига', 'Даугавпілс', 'Лієпая', 'Єлгава'],
		'Литва': ['Вільнюс', 'Каунас', 'Клайпеда', 'Шяуляй'],
		'США': ['Нью-Йорк', 'Лос-Анджелес', 'Чикаго', 'Маямі', 'Лас-Вегас', "Х'юстон", 'Даллас', 'Сан-Франциско'],
		'Канада': ['Торонто', 'Ванкувер', 'Монреаль', 'Калгарі', 'Оттава'],
		'Австралія': ['Сідней', 'Мельбурн', 'Брісбен', 'Перт', 'Аделаїда']
	};

	interface Props {
		open?: boolean;
		coords?: { lat: number; lng: number } | null;
		defaultCity?: string;
		defaultCountry?: string;
		onclose?: () => void;
		oncreated?: (slug: string) => void;
	}

	let {
		open = false,
		coords = null,
		defaultCity = '',
		defaultCountry = 'Україна',
		onclose,
		oncreated
	}: Props = $props();

	let name = $state('');
	let address = $state('');
	let city = $state('');
	let country = $state('Україна');
	let website = $state('');
	let phone = $state('');
	let isSubmitting = $state(false);

	$effect(() => {
		if (open) {
			city = defaultCity || '';
			country = defaultCountry || 'Україна';
		}
	});

	function handleCountryChange() {
		city = '';
	}

	function handleClose() {
		if (isSubmitting) return;
		resetForm();
		onclose?.();
	}

	function resetForm() {
		name = '';
		address = '';
		city = defaultCity || '';
		country = defaultCountry || 'Україна';
		website = '';
		phone = '';
		isSubmitting = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!name.trim() || !city.trim()) return;

		isSubmitting = true;
		try {
			const body: Record<string, any> = {
				name: name.trim(),
				country: country.trim() || 'Ukraine',
				city: city.trim(),
				address: address.trim(),
				website: website.trim(),
				phone: phone.trim()
			};
			if (coords) {
				body.latitude = coords.lat;
				body.longitude = coords.lng;
			}

			const resp = await authFetch('/clubs/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			if (resp.ok) {
				const data = await resp.json();
				const slug = data.data?.slug ?? data.slug ?? '';
				// Auto-join so the creator appears as a member immediately
				if (slug) {
					try {
						await authFetch(`/clubs/${slug}/join`, { method: 'POST' });
					} catch {
						// non-fatal — page will still reload the clubs list
					}
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

		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
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
					placeholder="Salsa Studio Kyiv"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

		<!-- Country -->
		<div class="flex flex-col gap-1.5">
			<label class="mu-text-primary text-[13px] font-semibold" for="club-country">
				Країна <span style="color: #e05252;">*</span>
			</label>
			<div class="select-wrapper mu-card mu-border rounded-[12px]" style="border-width: 1px; border-style: solid;">
				<select
					id="club-country"
					bind:value={country}
					onchange={handleCountryChange}
					class="mu-text-primary w-full bg-transparent px-4 py-3 text-[14px] font-medium outline-none"
				>
					{#each COUNTRIES as c}
						<option value={c}>{c}</option>
					{/each}
				</select>
				<i class="fi fi-rr-angle-down select-arrow" style="color: #aeb4bc;"></i>
			</div>
		</div>

		<!-- City -->
		<div class="flex flex-col gap-1.5">
			<label class="mu-text-primary text-[13px] font-semibold" for="club-city">
				{$t('map.city')} <span style="color: #e05252;">*</span>
			</label>
			{#if CITIES_BY_COUNTRY[country]}
				<div class="select-wrapper mu-card mu-border rounded-[12px]" style="border-width: 1px; border-style: solid;">
					<select
						id="club-city"
						bind:value={city}
						required
						class="mu-text-primary w-full bg-transparent px-4 py-3 text-[14px] font-medium outline-none"
					>
						<option value="">Оберіть місто</option>
						{#each CITIES_BY_COUNTRY[country] as c}
							<option value={c}>{c}</option>
						{/each}
					</select>
					<i class="fi fi-rr-angle-down select-arrow" style="color: #aeb4bc;"></i>
				</div>
			{:else}
				<input
					id="club-city"
					bind:value={city}
					type="text"
					required
					placeholder="Місто"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			{/if}
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
					bind:value={phone}
					type="tel"
					placeholder="+380 50 123 4567"
					class="mu-text-primary mu-card mu-border w-full rounded-[12px] px-4 py-3 text-[14px] font-medium outline-none"
					style="border-width: 1px; border-style: solid;"
				/>
			</div>

			<button
				type="submit"
				disabled={isSubmitting || !name.trim() || !city.trim()}
				class="mt-2 flex h-[46px] w-full items-center justify-center gap-2 rounded-[65px] text-[15px] font-bold text-white transition-opacity disabled:opacity-50"
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
		</form>
	</div>
</BottomSheet>

<style>
	.select-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}
	.select-wrapper select {
		-webkit-appearance: none;
		appearance: none;
		flex: 1;
	}
	.select-arrow {
		position: absolute;
		right: 14px;
		font-size: 13px;
		pointer-events: none;
		line-height: 1;
	}
</style>
