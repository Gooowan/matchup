<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { captureOnboardingComplete } from '$lib/analytics/posthog';
	import type { AccountType } from '$lib/types/accountType';
	import { isRestrictedAccountType } from '$lib/types/accountType';
import toast from 'svelte-french-toast';
import { t } from '$lib/locale';
import { formatUkrainianPhone } from '$lib/utils/phone';
import { parseApiError } from '$lib/utils/parseApiError';

	const STORAGE_KEY = 'matchup_onboarding';

	const PROGRAM_OPTIONS = $derived([
		{ value: 'latina', label: $t('onboarding.program_latina') },
		{ value: 'standard', label: $t('onboarding.program_standard') },
		{ value: 'both', label: $t('onboarding.program_both') }
	]);

	const GENDER_OPTIONS = $derived([
		{ value: 'male', label: $t('onboarding.gender_male') },
		{ value: 'female', label: $t('onboarding.gender_female') },
		{ value: 'other', label: $t('onboarding.gender_other') }
	]);
	const GOAL_OPTIONS = $derived([
		{ value: 'hobby', label: $t('onboarding.goal_hobby') },
		{ value: 'professional', label: $t('onboarding.goal_professional') }
	]);
	const FINANCE_OPTIONS: { value: string; label: string }[] = $derived([
		{ value: 'no', label: $t('onboarding.finance_no') },
		{ value: 'yes', label: $t('onboarding.finance_yes') },
		{ value: 'partial', label: $t('onboarding.finance_partial') }
	]);
	const CATEGORIES_UA = $derived([
		{ value: 'kids', label: $t('onboarding.category_kids') },
		{ value: 'juvenile1', label: $t('onboarding.category_juvenile1') },
		{ value: 'juvenile2', label: $t('onboarding.category_juvenile2') },
		{ value: 'junior1', label: $t('onboarding.category_junior1') },
		{ value: 'junior2', label: $t('onboarding.category_junior2') },
		{ value: 'youth', label: $t('onboarding.category_youth') },
		{ value: 'adult', label: $t('onboarding.category_adult') }
	]);
	// FUTURE (multi-country): these lists power the country/city pickers. For v1 the
	// country picker is locked to Ukraine (see the `disabled` <select> below and the
	// `country` default 'Україна'). To expand beyond UA: remove the `disabled` attr
	// on the country <select> in step 2 — the rest already keys off `country`.
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

	let step = $state(1);
	const TOTAL_STEPS = 5;

	// Step 1: Basic info
	// account_type may be 'dancer' | 'parent' | 'trainer' | 'club' — the latter
	// two branch into a tailored single-step flow at the bottom of this file.
	let accountType = $state<AccountType>('dancer');
	let isRestricted = $derived(isRestrictedAccountType(accountType));
	let firstName = $state('');
	let lastName = $state('');
	let gender = $state('');
	let birthDate = $state('');
	let heightCm = $state<number | null>(null);

	// Step 2: Dance details
	let danceProgram = $state('');
	let goal = $state('');
	let userCategories = $state<string[]>([]);
	let readyToRelocate = $state<boolean | null>(null);
	let readyToFinance = $state('');
	let bio = $state('');

	// Step 3: Partner preferences
	let prefGender = $state('');
	let ageMin = $state<number | null>(null);
	let ageMax = $state<number | null>(null);
	let heightMin = $state<number | null>(null);
	let heightMax = $state<number | null>(null);
	let prefGoal = $state('');
	let prefProgram = $state('');
	let prefCategories = $state<string[]>([]);
	let prefCountry = $state('Україна'); // locked to Ukraine for v1; expand COUNTRIES list when unlocking
	let prefCity = $state('');
	let wantsFinance = $state('');

	// Step 3: Club
	let selectedClubId = $state<string | null>(null);
	let selectedClubSlug = $state('');
	let selectedClubName = $state('');
	let clubSearchQuery = $state('');
	let clubResults = $state<{ id: string; slug: string; name: string; city: string }[]>([]);
	let showCreateForm = $state(false);
	let isCreatingClub = $state(false);
	let newClubName = $state('');
	let newClubAddress = $state('');
	let wasJustCreated = $state(false);
	let clubSearchTimer: ReturnType<typeof setTimeout> | null = null;

	// Trainer onboarding: optional club affiliation (up to 5).
	let trainerClubSearchQuery = $state('');
	let trainerClubSearchResults = $state<Array<{ id: string; slug: string; name: string; city: string }>>([]);
	let trainerJoinedClubs = $state<Array<{ id: string; slug: string; name: string; city: string }>>([]);
	let trainerJoiningSlug = $state<string | null>(null);
	let trainerClubSearchTimer: ReturnType<typeof setTimeout> | null = null;

	async function searchTrainerClubs() {
		const q = trainerClubSearchQuery.trim();
		if (q.length < 2) { trainerClubSearchResults = []; return; }
		try {
			const params = new URLSearchParams({ q, limit: '10' });
			const resp = await fetch(`${import.meta.env.VITE_API_URL}/clubs?${params}`);
			if (resp.ok) {
				const body = await resp.json();
				trainerClubSearchResults = ((body.data ?? []) as Array<{ id: string; slug: string; name: string; city: string }>)
					.filter((c) => !trainerJoinedClubs.some((j) => j.id === c.id));
			}
		} catch { /* non-fatal */ }
	}

	async function trainerJoinClub(club: { id: string; slug: string; name: string; city: string }) {
		if (trainerJoinedClubs.length >= 5) return;
		trainerJoiningSlug = club.slug;
		try {
			const resp = await authFetch(`/clubs/${club.slug}/join`, { method: 'POST' });
			if (resp.ok) {
				trainerJoinedClubs = [...trainerJoinedClubs, club];
				trainerClubSearchResults = trainerClubSearchResults.filter((c) => c.id !== club.id);
			}
		} catch { /* non-fatal */ } finally {
			trainerJoiningSlug = null;
		}
	}

	// Google Maps import state (inline create form + club restricted flow)
	let showGmapsInputOnboarding = $state(false);
	let gmapsURLOnboarding = $state('');
	let isImportingGmaps = $state(false);
	let newClubPhotos = $state<string[]>([]);
	let newClubLat = $state<number | null>(null);
	let newClubLng = $state<number | null>(null);

	// Club onboarding extra fields (restricted club account flow)
	let newClubWebsite = $state('');
	let newClubPhone = $state('');
	let newClubDescription = $state('');

	const WORKING_DAYS = [
		{ key: 'mon', label: 'Пн' },
		{ key: 'tue', label: 'Вт' },
		{ key: 'wed', label: 'Ср' },
		{ key: 'thu', label: 'Чт' },
		{ key: 'fri', label: 'Пт' },
		{ key: 'sat', label: 'Сб' },
		{ key: 'sun', label: 'Нд' }
	] as const;
	type DayKey = (typeof WORKING_DAYS)[number]['key'];
	interface WorkingHoursDay { key: DayKey; label: string; enabled: boolean; open: string; close: string; }
	let newClubWorkingHoursDays = $state<WorkingHoursDay[]>(
		WORKING_DAYS.map((d) => ({ ...d, enabled: false, open: '09:00', close: '21:00' }))
	);

	// Step 4: Location + photo
	// v1: locked to Kyiv / Ukraine. To open multi-city support: remove the
	// HARDCODED_* constants, restore the dropdown UI below, and re-enable saving
	// the selected city/country.
	const HARDCODED_CITY = 'Київ';
	const HARDCODED_COUNTRY = 'Україна';
	let city = $state(HARDCODED_CITY);
	let country = $state(HARDCODED_COUNTRY);
	let avatarFile = $state<File | null>(null);
	let avatarPreview = $state('');
	let fileInput = $state<HTMLInputElement | null>(null);

	// Extra photos (deferred upload at handleFinish)
	let extraPhotoFiles = $state<File[]>([]);
	let extraPhotoPreviews = $state<string[]>([]);
	let photoFileInput = $state<HTMLInputElement | null>(null);

	let isSaving = $state(false);

	onMount(() => {
		// Seed accountType from the user's profile_data (set on register).
		const initialType = (authStore.user?.profile_data?.account_type as AccountType | undefined) ?? undefined;
		if (initialType) accountType = initialType;

		try {
			const raw = sessionStorage.getItem(STORAGE_KEY);
			if (!raw) return;
			const s = JSON.parse(raw);
			if (s.step) step = s.step;
			if (s.accountType) accountType = s.accountType;
			if (s.firstName) firstName = s.firstName;
			if (s.lastName) lastName = s.lastName;
			if (s.gender) gender = s.gender;
			if (s.birthDate) birthDate = s.birthDate;
			if (s.heightCm != null) heightCm = s.heightCm;
			if (s.danceProgram) danceProgram = s.danceProgram;
			if (s.goal) goal = s.goal;
			if (s.userCategories) userCategories = s.userCategories;
			if (s.readyToRelocate != null) readyToRelocate = s.readyToRelocate;
			if (s.readyToFinance) readyToFinance = s.readyToFinance;
			if (s.bio) bio = s.bio;
			if (s.prefGender) prefGender = s.prefGender;
			if (s.ageMin != null) ageMin = s.ageMin;
			if (s.ageMax != null) ageMax = s.ageMax;
			if (s.heightMin != null) heightMin = s.heightMin;
			if (s.heightMax != null) heightMax = s.heightMax;
			if (s.prefGoal) prefGoal = s.prefGoal;
			if (s.prefProgram) prefProgram = s.prefProgram;
			if (s.prefCategories) prefCategories = s.prefCategories;
			if (s.prefCountry) prefCountry = s.prefCountry;
			if (s.prefCity) prefCity = s.prefCity;
			if (s.wantsFinance) wantsFinance = s.wantsFinance;
			if (s.city) city = s.city;
			if (s.country) country = s.country;
			if (s.selectedClubId) selectedClubId = s.selectedClubId;
			if (s.selectedClubSlug) selectedClubSlug = s.selectedClubSlug;
			if (s.selectedClubName) selectedClubName = s.selectedClubName;
			if (s.wasJustCreated) wasJustCreated = s.wasJustCreated;
			if (s.newClubWebsite) newClubWebsite = s.newClubWebsite;
			if (s.newClubPhone) newClubPhone = s.newClubPhone;
			if (s.newClubDescription) newClubDescription = s.newClubDescription;
			if (s.newClubWorkingHoursDays) newClubWorkingHoursDays = s.newClubWorkingHoursDays;
		} catch {
			// ignore corrupt storage
		}
	});

	$effect(() => {
		sessionStorage.setItem(
			STORAGE_KEY,
			JSON.stringify({
				step, accountType, firstName, lastName, gender, birthDate, heightCm,
				danceProgram, goal, userCategories, readyToRelocate, readyToFinance, bio,
				selectedClubId, selectedClubSlug, selectedClubName, wasJustCreated,
				prefGender, ageMin, ageMax, heightMin, heightMax,
				prefGoal, prefProgram, prefCategories, prefCountry, prefCity,
				wantsFinance,
				city, country,
				newClubWebsite, newClubPhone, newClubDescription, newClubWorkingHoursDays
			})
		);
	});

	// Load clubs immediately when the user reaches the club step
	$effect(() => {
		if (step === 3 && !selectedClubId && !showCreateForm) {
			searchClubs();
		}
	});

	function getAge(dateStr: string): number {
		if (!dateStr) return 0;
		const today = new Date();
		const dob = new Date(dateStr);
		let age = today.getFullYear() - dob.getFullYear();
		const m = today.getMonth() - dob.getMonth();
		if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) age--;
		return age;
	}

	let isParent = $derived(accountType === 'parent');
	let isAmateur = $derived(goal === 'hobby');

	let ageError = $derived(
		birthDate && !isParent && getAge(birthDate) < 18
			? $t('onboarding.age_error')
			: ''
	);

	function toggleUserCategory(v: string) {
		userCategories = userCategories.includes(v)
			? userCategories.filter((x) => x !== v)
			: [...userCategories, v];
	}

	function togglePrefCategory(v: string) {
		prefCategories = prefCategories.includes(v)
			? prefCategories.filter((x) => x !== v)
			: [...prefCategories, v];
	}

	// --- Image crop ---
	const CROP_SIZE = 280;
	let cropMode = $state(false);
	let cropX = $state(0);
	let cropY = $state(0);
	let cropImgW = $state(0);
	let cropImgH = $state(0);
	let cropImgEl = $state<HTMLImageElement | null>(null);
	let dragStart: { x: number; y: number; ox: number; oy: number } | null = null;

	function handleAvatarChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		avatarPreview = URL.createObjectURL(file);
		cropMode = true;
	}

	function initCrop() {
		if (!cropImgEl) return;
		const { naturalWidth: nw, naturalHeight: nh } = cropImgEl;
		const scale = Math.max(CROP_SIZE / nw, CROP_SIZE / nh);
		cropImgW = Math.round(nw * scale);
		cropImgH = Math.round(nh * scale);
		cropX = Math.round((CROP_SIZE - cropImgW) / 2);
		cropY = Math.round((CROP_SIZE - cropImgH) / 2);
	}

	function clampCrop(nx: number, ny: number) {
		cropX = Math.min(0, Math.max(CROP_SIZE - cropImgW, nx));
		cropY = Math.min(0, Math.max(CROP_SIZE - cropImgH, ny));
	}

	function onCropMouseDown(e: MouseEvent) {
		dragStart = { x: e.clientX, y: e.clientY, ox: cropX, oy: cropY };
		window.addEventListener('mousemove', onCropMouseMove);
		window.addEventListener('mouseup', onCropMouseUp);
	}
	function onCropMouseMove(e: MouseEvent) {
		if (!dragStart) return;
		clampCrop(dragStart.ox + e.clientX - dragStart.x, dragStart.oy + e.clientY - dragStart.y);
	}
	function onCropMouseUp() {
		dragStart = null;
		window.removeEventListener('mousemove', onCropMouseMove);
		window.removeEventListener('mouseup', onCropMouseUp);
	}

	function onCropTouchStart(e: TouchEvent) {
		const t = e.touches[0];
		dragStart = { x: t.clientX, y: t.clientY, ox: cropX, oy: cropY };
		window.addEventListener('touchmove', onCropTouchMove, { passive: false });
		window.addEventListener('touchend', onCropTouchEnd);
	}
	function onCropTouchMove(e: TouchEvent) {
		e.preventDefault();
		if (!dragStart) return;
		const t = e.touches[0];
		clampCrop(dragStart.ox + t.clientX - dragStart.x, dragStart.oy + t.clientY - dragStart.y);
	}
	function onCropTouchEnd() {
		dragStart = null;
		window.removeEventListener('touchmove', onCropTouchMove);
		window.removeEventListener('touchend', onCropTouchEnd);
	}

	async function confirmCrop() {
		if (!cropImgEl) return;
		const OUT = 500;
		const ratio = OUT / CROP_SIZE;
		const canvas = document.createElement('canvas');
		canvas.width = OUT;
		canvas.height = OUT;
		const ctx = canvas.getContext('2d')!;
		ctx.drawImage(cropImgEl, cropX * ratio, cropY * ratio, cropImgW * ratio, cropImgH * ratio);
		const blob = await new Promise<Blob>((res) => canvas.toBlob((b) => res(b!), 'image/jpeg', 0.92));
		avatarFile = new File([blob], 'avatar.jpg', { type: 'image/jpeg' });
		avatarPreview = URL.createObjectURL(blob);
		cropMode = false;
	}

	// --- Extra photo crop (step 4) ---
	let extraCropMode = $state(false);
	let extraCropX = $state(0);
	let extraCropY = $state(0);
	let extraCropImgW = $state(0);
	let extraCropImgH = $state(0);
	let extraCropImgEl = $state<HTMLImageElement | null>(null);
	let extraCropPreview = $state('');
	let extraDragStart: { x: number; y: number; ox: number; oy: number } | null = null;

	function handlePhotoChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		extraCropPreview = URL.createObjectURL(file);
		extraCropMode = true;
		(e.target as HTMLInputElement).value = '';
	}

	function initExtraCrop() {
		if (!extraCropImgEl) return;
		const { naturalWidth: nw, naturalHeight: nh } = extraCropImgEl;
		const scale = Math.max(CROP_SIZE / nw, CROP_SIZE / nh);
		extraCropImgW = Math.round(nw * scale);
		extraCropImgH = Math.round(nh * scale);
		extraCropX = Math.round((CROP_SIZE - extraCropImgW) / 2);
		extraCropY = Math.round((CROP_SIZE - extraCropImgH) / 2);
	}

	function clampExtraCrop(nx: number, ny: number) {
		extraCropX = Math.min(0, Math.max(CROP_SIZE - extraCropImgW, nx));
		extraCropY = Math.min(0, Math.max(CROP_SIZE - extraCropImgH, ny));
	}

	function onExtraCropMouseDown(e: MouseEvent) {
		extraDragStart = { x: e.clientX, y: e.clientY, ox: extraCropX, oy: extraCropY };
		window.addEventListener('mousemove', onExtraCropMouseMove);
		window.addEventListener('mouseup', onExtraCropMouseUp);
	}
	function onExtraCropMouseMove(e: MouseEvent) {
		if (!extraDragStart) return;
		clampExtraCrop(extraDragStart.ox + e.clientX - extraDragStart.x, extraDragStart.oy + e.clientY - extraDragStart.y);
	}
	function onExtraCropMouseUp() {
		extraDragStart = null;
		window.removeEventListener('mousemove', onExtraCropMouseMove);
		window.removeEventListener('mouseup', onExtraCropMouseUp);
	}
	function onExtraCropTouchStart(e: TouchEvent) {
		const t = e.touches[0];
		extraDragStart = { x: t.clientX, y: t.clientY, ox: extraCropX, oy: extraCropY };
		window.addEventListener('touchmove', onExtraCropTouchMove, { passive: false });
		window.addEventListener('touchend', onExtraCropTouchEnd);
	}
	function onExtraCropTouchMove(e: TouchEvent) {
		e.preventDefault();
		if (!extraDragStart) return;
		const t = e.touches[0];
		clampExtraCrop(extraDragStart.ox + t.clientX - extraDragStart.x, extraDragStart.oy + t.clientY - extraDragStart.y);
	}
	function onExtraCropTouchEnd() {
		extraDragStart = null;
		window.removeEventListener('touchmove', onExtraCropTouchMove);
		window.removeEventListener('touchend', onExtraCropTouchEnd);
	}
	async function confirmExtraCrop() {
		if (!extraCropImgEl) return;
		const OUT = 500;
		const ratio = OUT / CROP_SIZE;
		const canvas = document.createElement('canvas');
		canvas.width = OUT;
		canvas.height = OUT;
		const ctx2d = canvas.getContext('2d')!;
		ctx2d.drawImage(extraCropImgEl, extraCropX * ratio, extraCropY * ratio, extraCropImgW * ratio, extraCropImgH * ratio);
		const blob = await new Promise<Blob>((res) => canvas.toBlob((b) => res(b!), 'image/jpeg', 0.92));
		const file = new File([blob], 'photo.jpg', { type: 'image/jpeg' });
		extraCropMode = false;
		extraCropPreview = '';
		extraPhotoFiles = [...extraPhotoFiles, file];
		extraPhotoPreviews = [...extraPhotoPreviews, URL.createObjectURL(blob)];
	}

	function removeExtraPhoto(i: number) {
		extraPhotoFiles = extraPhotoFiles.filter((_, idx) => idx !== i);
		extraPhotoPreviews = extraPhotoPreviews.filter((_, idx) => idx !== i);
	}

	function handleClubSearchInput() {
		if (clubSearchTimer) clearTimeout(clubSearchTimer);
		clubSearchTimer = setTimeout(searchClubs, 300);
	}

	async function searchClubs() {
		try {
			const resp = await authFetch(`/clubs?q=${encodeURIComponent(clubSearchQuery)}&limit=20`);
			if (resp.ok) {
				const body = await resp.json();
				clubResults = (body.data ?? []).map((c: { id: string; slug: string; name: string; city: string }) => ({
					id: c.id, slug: c.slug, name: c.name, city: c.city
				}));
			}
		} catch { clubResults = []; }
	}

	function handleClubPhoneInput(e: Event) {
		const target = e.currentTarget as HTMLInputElement;
		newClubPhone = formatUkrainianPhone(target.value);
	}
	function handleClubPhonePaste(e: ClipboardEvent) {
		const text = e.clipboardData?.getData('text') ?? '';
		if (!text) return;
		e.preventDefault();
		newClubPhone = formatUkrainianPhone(text);
	}
	function toggleWorkingDay(key: DayKey) {
		newClubWorkingHoursDays = newClubWorkingHoursDays.map((d) =>
			d.key === key ? { ...d, enabled: !d.enabled } : d
		);
	}
	function setWorkingHour(key: DayKey, field: 'open' | 'close', value: string) {
		newClubWorkingHoursDays = newClubWorkingHoursDays.map((d) =>
			d.key === key ? { ...d, [field]: value } : d
		);
	}

	// Maps English day names from the Places API to the short keys used in
	// newClubWorkingHoursDays, then parses the "HH:MM AM – HH:MM PM" strings
	// into 24h "HH:MM" open/close values and marks the day as enabled.
	const GMAPS_DAY_KEY: Record<string, DayKey> = {
		monday: 'mon', tuesday: 'tue', wednesday: 'wed',
		thursday: 'thu', friday: 'fri', saturday: 'sat', sunday: 'sun'
	};
	function parseHours12(h12: string): string {
		// Parses "9:00 AM", "12:00 PM", "6:30 PM" → "09:00", "12:00", "18:30"
		const m = h12.trim().match(/^(\d{1,2}):(\d{2})\s*(AM|PM)$/i);
		if (!m) return h12.trim().slice(0, 5); // fallback: return as-is up to 5 chars
		let [, hStr, min, period] = m;
		let h = parseInt(hStr, 10);
		if (period.toUpperCase() === 'AM') { if (h === 12) h = 0; }
		else { if (h !== 12) h += 12; }
		return `${String(h).padStart(2, '0')}:${min}`;
	}
	function applyGmapsHours(days: WorkingHoursDay[], hours: Record<string, string>): WorkingHoursDay[] {
		return days.map((d) => {
			const match = Object.entries(hours).find(([k]) => GMAPS_DAY_KEY[k.toLowerCase()] === d.key);
			if (!match) return d;
			const [, rangeStr] = match;
			// "9:00 AM – 9:00 PM"  or "Closed" / "24 hours" etc.
			const parts = rangeStr.split('–').map((s) => s.trim());
			if (parts.length === 2) {
				return { ...d, enabled: true, open: parseHours12(parts[0]), close: parseHours12(parts[1]) };
			}
			// "Closed" → keep disabled; "24 hours" → 00:00–23:59
			if (rangeStr.toLowerCase().includes('24 hour')) {
				return { ...d, enabled: true, open: '00:00', close: '23:59' };
			}
			return d; // "Closed" or unrecognised — leave disabled
		});
	}

	async function importFromGmapsOnboarding() {
		if (!gmapsURLOnboarding.trim()) return;
		isImportingGmaps = true;
		try {
			const resp = await authFetch('/me/clubs/parse-gmaps', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url: gmapsURLOnboarding.trim() })
			});
			if (resp.ok) {
				const { data } = await resp.json();
				if (data.name)      newClubName      = data.name;
				if (data.address)   newClubAddress   = data.address;
				if (data.website)   newClubWebsite   = data.website;
				if (data.phone)     newClubPhone     = formatUkrainianPhone(data.phone);
				if (data.latitude)  newClubLat       = data.latitude;
				if (data.longitude) newClubLng       = data.longitude;
				// Pre-populate working hours toggles from Google's weekday_text.
				if (data.working_hours && Object.keys(data.working_hours).length > 0) {
					newClubWorkingHoursDays = applyGmapsHours(newClubWorkingHoursDays, data.working_hours);
				}
				newClubPhotos = data.photos ?? [];
				showGmapsInputOnboarding = false;
				gmapsURLOnboarding = '';
				toast.success('Заповнено з Google Maps');
			} else {
				toast.error('Не вдалося отримати дані з цього посилання');
			}
		} catch {
			toast.error('Не вдалося отримати дані з цього посилання');
		} finally {
			isImportingGmaps = false;
		}
	}

	async function createAndJoinClub() {
		if (!newClubName.trim()) return;
		isCreatingClub = true;
		try {
		const registerBody: Record<string, unknown> = {
			name: newClubName.trim(),
			country: 'Ukraine',
			city: 'Kyiv',
			address: newClubAddress.trim() || undefined
		};
		// Coord priority: GMaps import > server geocoding (address) > Kyiv centroid.
		// Sending (0,0) triggers Nominatim on the server; on failure it falls back to
		// the Kyiv centroid automatically.
		if (newClubLat !== null && newClubLng !== null) {
			registerBody.latitude = newClubLat;
			registerBody.longitude = newClubLng;
		} else if (newClubAddress.trim()) {
			// Let the server geocode the address.
			registerBody.latitude = 0;
			registerBody.longitude = 0;
		} else {
			// No address — use Kyiv centroid so the club still appears on the map.
			registerBody.latitude = 50.4501;
			registerBody.longitude = 30.5234;
		}
			if (newClubPhotos.length > 0) {
				registerBody.photos = newClubPhotos;
			}
			const resp = await authFetch('/clubs/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(registerBody)
			});
			if (resp.ok) {
				const body = await resp.json();
				selectedClubId = body.data?.id ?? null;
				selectedClubSlug = body.data?.slug ?? '';
				selectedClubName = newClubName.trim();
				wasJustCreated = true;
				// Join as member + claim ownership so the creator can edit the club later.
				if (selectedClubSlug) {
					try {
						await authFetch(`/clubs/${selectedClubSlug}/join`, { method: 'POST' });
					} catch { /* non-fatal */ }
					try {
						await authFetch(`/clubs/${selectedClubSlug}/claim`, { method: 'POST' });
					} catch { /* non-fatal */ }
				}
				showCreateForm = false;
				toast.success('Клуб створено!');
			} else {
				const body = await resp.json().catch(() => ({}));
				toast.error(parseApiError(body, resp.status));
			}
		} catch {
			toast.error('Помилка. Спробуй ще раз.');
		} finally {
			isCreatingClub = false;
		}
	}

	function canAdvance(): boolean {
		if (step === 1) return firstName.length >= 2 && lastName.length >= 2 && !!gender && !!birthDate && !ageError;
		if (step === 2) return !!goal && !!danceProgram;
		return true;
	}

	// --- Tailored finish handlers for restricted account types (trainer / club) ---

	async function handleFinishTrainer() {
		if (firstName.trim().length < 2) {
			toast.error($t('onboarding.first_name_error'));
			return;
		}
		isSaving = true;
		try {
			const nameBody: Record<string, unknown> = { first_name: firstName };
			if (lastName.trim()) nameBody.last_name = lastName.trim();
			const userResp = await authFetch('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(nameBody)
			});
			if (!userResp.ok) throw new Error('Не вдалось зберегти профіль');

			if (avatarFile) await authStore.uploadAvatar(avatarFile);

			const profileBody: Record<string, unknown> = {
				account_type: 'trainer',
				categories: userCategories
			};
			if (gender) profileBody.gender = gender;
			if (bio) profileBody.bio = bio;

			const profileResp = await authFetch('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(profileBody)
			});
			if (!profileResp.ok) throw new Error('Не вдалось зберегти профіль');

			captureOnboardingComplete();
			sessionStorage.removeItem(STORAGE_KEY);
			await authStore.checkAuth();
			goto('/settings');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Щось пішло не так. Спробуй ще раз.');
		} finally {
			isSaving = false;
		}
	}

	async function handleFinishClub() {
		if (newClubName.trim().length < 2) {
			toast.error('Введіть назву клубу');
			return;
		}
		isSaving = true;
		try {
			const displayName = newClubName.trim();

			// 1. Save the club name as first_name so existing app guards (which require
			//    first_name) pass. Clubs don't have a surname, so we omit last_name.
			const userResp = await authFetch('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ first_name: displayName })
			});
			if (!userResp.ok) throw new Error('Не вдалось зберегти профіль');

			if (avatarFile) await authStore.uploadAvatar(avatarFile);

			// 2. Persist account_type on the profile so the nav/route guards apply.
			const profileResp = await authFetch('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ account_type: 'club' })
			});
			if (!profileResp.ok) throw new Error('Не вдалось зберегти профіль');

		// 3. Create the club (UA/Kyiv hardcoded).
		// Coord priority: GMaps import > server geocoding (address) > Kyiv centroid.
		const clubBody: Record<string, unknown> = {
			name: displayName,
			country: 'Ukraine',
			city: 'Kyiv'
		};
		if (newClubAddress.trim()) clubBody.address = newClubAddress.trim();
		if (newClubWebsite.trim()) clubBody.website = newClubWebsite.trim();
		if (newClubPhone.trim()) clubBody.phone = newClubPhone.trim();
		if (newClubDescription.trim()) clubBody.description = newClubDescription.trim();
		if (newClubPhotos.length > 0) clubBody.photos = newClubPhotos;
		const workingHoursObj: Record<string, { open: string; close: string } | null> = {};
		let hasWorkingHours = false;
		for (const day of newClubWorkingHoursDays) {
			if (day.enabled) { workingHoursObj[day.key] = { open: day.open, close: day.close }; hasWorkingHours = true; }
		}
		if (hasWorkingHours) clubBody.working_hours = workingHoursObj;
		if (newClubLat !== null && newClubLng !== null) {
			clubBody.latitude = newClubLat;
			clubBody.longitude = newClubLng;
		} else if (newClubAddress.trim()) {
			// Let the server geocode the address; centroid is the server-side fallback.
			clubBody.latitude = 0;
			clubBody.longitude = 0;
		} else {
			clubBody.latitude = 50.4501;
			clubBody.longitude = 30.5234;
		}

			const clubResp = await authFetch('/clubs/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(clubBody)
			});
			if (clubResp.ok) {
				const data = await clubResp.json();
				const slug = data.data?.slug ?? data.slug ?? '';
				if (slug) {
					// Try to claim ownership; non-fatal if it fails.
					try {
						await authFetch(`/clubs/${slug}/claim`, { method: 'POST' });
					} catch { /* non-fatal */ }
				}
			}

			captureOnboardingComplete();
			sessionStorage.removeItem(STORAGE_KEY);
			await authStore.checkAuth();
			goto('/settings');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Щось пішло не так. Спробуй ще раз.');
		} finally {
			isSaving = false;
		}
	}

	async function handleFinish() {
		isSaving = true;
		async function postOrThrow(path: string, init: RequestInit) {
			const r = await authFetch(path, init);
			if (!r.ok) {
				const b = await r.json().catch(() => ({}));
				throw new Error(parseApiError(b, r.status));
			}
			return r;
		}

		try {
			await postOrThrow('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ first_name: firstName, last_name: lastName })
			});

			if (avatarFile) {
				await authStore.uploadAvatar(avatarFile);
			}

			const body: Record<string, unknown> = {
				gender,
				goal: goal || 'hobby',
				program: danceProgram || 'standard',
				country,
				city,
				account_type: accountType,
				categories: userCategories,
				primary_club_id: selectedClubId ?? null
			};
			if (birthDate) body.birth_date = birthDate;
			if (heightCm) body.height_cm = heightCm;
			if (readyToRelocate !== null) body.ready_to_relocate = readyToRelocate;
			if (readyToFinance) body.ready_to_finance = readyToFinance;
			if (bio) body.bio = bio;

			await postOrThrow('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			if (selectedClubId && selectedClubSlug) {
				try {
					await postOrThrow(`/clubs/${selectedClubSlug}/join`, { method: 'POST' });
				} catch {
					// non-fatal: profile already saved; JoinClub is idempotent (ON CONFLICT DO NOTHING)
				}
			}

			const prefBody: Record<string, unknown> = {
				preferred_categories: prefCategories
			};
			if (prefGender) prefBody.preferred_gender = prefGender;
			if (ageMin !== null) prefBody.age_min = ageMin;
			if (ageMax !== null) prefBody.age_max = ageMax;
			if (heightMin !== null) prefBody.height_min = heightMin;
			if (heightMax !== null) prefBody.height_max = heightMax;
			if (prefGoal) prefBody.preferred_goal = prefGoal;
			if (prefProgram) prefBody.preferred_program = prefProgram;
			if (prefCountry || country) prefBody.preferred_country = prefCountry || country;
			if (prefCity || city) prefBody.preferred_city = prefCity || city;
			if (wantsFinance) prefBody.wants_partner_to_finance = wantsFinance;

			await postOrThrow('/me/preferences', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(prefBody)
			});

			for (const file of extraPhotoFiles) {
				try {
					const formData = new FormData();
					formData.append('photo', file);
					const photoResp = await authFetch('/user/files/photo', { method: 'POST', body: formData });
					if (photoResp.ok) {
						const { data } = await photoResp.json();
						await authFetch('/me/profile/media', {
							method: 'POST',
							headers: { 'Content-Type': 'application/json' },
							body: JSON.stringify({ url: data.url })
						});
					}
				} catch {
					// non-fatal: skip failed photo
				}
			}

			captureOnboardingComplete();
			sessionStorage.removeItem(STORAGE_KEY);
			await authStore.checkAuth();
			goto('/feed');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Щось пішло не так. Спробуй ще раз.');
		} finally {
			isSaving = false;
		}
	}
</script>

{#if isRestricted}
	<!-- Tailored single-step flow for trainer & club accounts -->
	<div class="flex h-[100dvh] flex-col px-6 pt-safe pb-safe" style="background: #dae1eb; overflow: hidden;">
		<div class="flex-shrink-0 pt-14 pb-8">
			<h1 class="text-[28px] font-black" style="color: #171717;">
				{accountType === 'trainer' ? 'Профіль тренера' : 'Профіль клубу'}
			</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">
				{accountType === 'trainer'
					? 'Заповніть основні дані, щоб клієнти могли вас знайти'
					: 'Розкажіть про ваш клуб'}
			</p>
		</div>

		<div
			class="flex-1 min-h-0 overflow-y-auto"
			style="display: flex; flex-direction: column; gap: 16px; padding-bottom: 24px; -webkit-overflow-scrolling: touch;"
		>
			<!-- Avatar -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">
					{accountType === 'trainer' ? 'Фото' : 'Лого клубу'}
				</label>
				<div class="flex items-center gap-4">
					<button
						type="button"
						onclick={() => fileInput?.click()}
						class="relative flex h-[72px] w-[72px] flex-shrink-0 items-center justify-center overflow-hidden rounded-[16px]"
						style="background: #dae1eb;"
						aria-label="Завантажити фото"
					>
						{#if avatarPreview}
							<img src={avatarPreview} alt="Avatar" class="h-full w-full object-cover" />
						{:else}
							<i class="fi fi-rr-camera" style="font-size: 24px; color: #aeb4bc; line-height: 1;"></i>
						{/if}
					</button>
					<input
						bind:this={fileInput}
						type="file"
						accept="image/*"
						class="hidden"
						onchange={handleAvatarChange}
					/>
					<p class="text-[13px] font-medium" style="color: #696969;">
						{avatarPreview ? 'Натисніть, щоб замінити' : 'Натисніть, щоб завантажити'}
					</p>
				</div>
			</div>

		{#if accountType === 'trainer'}
			<!-- Trainer: name + bio + gender + categories -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ІМ'Я ТА ПРІЗВИЩЕ</label>
				<input
					type="text"
					placeholder="Ім'я"
					bind:value={firstName}
					class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
					style="color: #171717; border-color: #e0e0e0;"
				/>
				<input
					type="text"
					placeholder="Прізвище (необов'язково)"
					bind:value={lastName}
					class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
					style="color: #171717; border-color: #e0e0e0;"
				/>
			</div>

			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">СТАТЬ</label>
				<div class="flex gap-2">
					{#each [{ value: 'male', label: 'Чоловік' }, { value: 'female', label: 'Жінка' }, { value: 'other', label: 'Інше' }] as g}
						{@const sel = gender === g.value}
						<button
							type="button"
							onclick={() => (gender = g.value)}
							class="flex-1 rounded-[50px] py-2 text-[13px] font-semibold transition-all"
							style="background: {sel ? '#8984da' : 'transparent'}; color: {sel ? 'white' : '#696969'}; border: 1.5px solid {sel ? '#8984da' : '#d1d5db'};"
						>{g.label}</button>
					{/each}
				</div>
			</div>

			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРО ВАС</label>
				<textarea
					placeholder="Коротко про досвід і спеціалізацію"
					bind:value={bio}
					rows="3"
					class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
					style="color: #171717; border-color: #e0e0e0; resize: none;"
				></textarea>
			</div>

			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КАТЕГОРІЇ</label>
				<div class="flex flex-wrap gap-2">
					{#each CATEGORIES_UA as c}
						{@const selected = userCategories.includes(c.value)}
						<button
							type="button"
							onclick={() => toggleUserCategory(c.value)}
							class="rounded-[50px] px-3 py-1.5 text-[12px] font-semibold transition-all"
							style="background: {selected ? '#8984da' : 'transparent'}; color: {selected ? 'white' : '#696969'}; border: 1.5px solid {selected ? '#8984da' : '#d1d5db'};"
						>{c.label}</button>
					{/each}
				</div>
			</div>
		<!-- Trainer: optional club affiliation -->
		{#if accountType === 'trainer' && trainerJoinedClubs.length < 5}
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<div class="flex items-center justify-between">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">МОЇ КЛУБИ (НЕОБОВ'ЯЗКОВО)</label>
					{#if trainerJoinedClubs.length > 0}
						<span class="text-[11px] font-semibold" style="color: #8984da;">{trainerJoinedClubs.length}/5</span>
					{/if}
				</div>

				{#each trainerJoinedClubs as club}
					<div class="flex items-center gap-2 rounded-[10px] px-3 py-2" style="background: rgba(137,132,218,0.1);">
						<i class="fi fi-rr-bank" style="font-size: 14px; color: #8984da; line-height: 1;"></i>
						<span class="flex-1 truncate text-[13px] font-semibold" style="color: #171717;">{club.name}</span>
						<button
							type="button"
							onclick={() => { trainerJoinedClubs = trainerJoinedClubs.filter((c) => c.id !== club.id); }}
							class="flex h-6 w-6 items-center justify-center"
						>
							<i class="fi fi-rr-cross-small" style="font-size: 16px; color: #aeb4bc; line-height: 1;"></i>
						</button>
					</div>
				{/each}

				<input
					type="text"
					placeholder="Знайти клуб…"
					bind:value={trainerClubSearchQuery}
					oninput={() => {
						if (trainerClubSearchTimer) clearTimeout(trainerClubSearchTimer);
						trainerClubSearchTimer = setTimeout(searchTrainerClubs, 300);
					}}
					class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
					style="color: #171717; border-color: #e0e0e0;"
				/>

				{#if trainerClubSearchResults.length > 0}
					<div class="flex flex-col gap-1">
						{#each trainerClubSearchResults.slice(0, 5) as club}
							<button
								type="button"
								onclick={() => trainerJoinClub(club)}
								disabled={trainerJoiningSlug === club.slug}
								class="flex items-center gap-2 rounded-[10px] px-3 py-2 text-left transition-opacity disabled:opacity-50"
								style="background: rgba(0,0,0,0.04);"
							>
								<i class="fi fi-rr-bank" style="font-size: 14px; color: #aeb4bc; line-height: 1;"></i>
								<span class="flex-1 truncate text-[13px] font-medium" style="color: #171717;">{club.name}</span>
								<span class="text-[12px]" style="color: #aeb4bc;">{club.city}</span>
								<i class="fi fi-rr-plus" style="font-size: 14px; color: #8984da; line-height: 1;"></i>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		{:else}
			<!-- Club: Google Maps import -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ІМПОРТ З GOOGLE MAPS</label>
					{#if !showGmapsInputOnboarding}
						<button
							type="button"
							onclick={() => (showGmapsInputOnboarding = true)}
							class="flex items-center justify-center gap-2 rounded-[50px] py-2 text-[13px] font-semibold"
							style="background: rgba(137,132,218,0.12); color: #8984da;"
						>
							<i class="fi fi-rr-link" style="font-size: 14px; line-height: 1;"></i>
							Заповнити з Google Maps
						</button>
					{:else}
						<div style="display: flex; flex-direction: column; gap: 6px;">
							<input
								type="url"
								bind:value={gmapsURLOnboarding}
								placeholder="Вставте посилання Google Maps…"
								class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
								style="color: #171717; border-color: #e0e0e0;"
							/>
							<div class="flex gap-2">
								<button
									type="button"
									onclick={() => { showGmapsInputOnboarding = false; gmapsURLOnboarding = ''; }}
									class="flex-1 rounded-[50px] py-2 text-[13px] font-semibold"
									style="background: transparent; color: #696969; border: 1.5px solid #d1d5db;"
								>Скасувати</button>
								<button
									type="button"
									onclick={importFromGmapsOnboarding}
									disabled={isImportingGmaps || !gmapsURLOnboarding.trim()}
									class="flex flex-1 items-center justify-center gap-2 rounded-[50px] py-2 text-[13px] font-semibold text-white transition-opacity disabled:opacity-50"
									style="background: #8984da;"
								>
									{#if isImportingGmaps}
										<div class="h-4 w-4 animate-spin rounded-full border-2 border-white/30" style="border-top-color: white;"></div>
									{/if}
									Імпорт
								</button>
							</div>
						</div>
					{/if}
					{#if newClubPhotos.length > 0}
						<div class="flex gap-2 overflow-x-auto pb-1">
							{#each newClubPhotos as url}
								<img src={url} alt="Фото клубу" class="h-[64px] w-[64px] flex-shrink-0 rounded-[10px] object-cover" />
							{/each}
						</div>
					{/if}
				</div>

				<!-- Club: name -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">НАЗВА КЛУБУ <span style="color: #e05252;">*</span></label>
					<input
						type="text"
						placeholder="Salsa Studio Kyiv"
						bind:value={newClubName}
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0;"
					/>
				</div>

				<!-- Club: address + locked city -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">АДРЕСА</label>
					<input
						type="text"
						placeholder="вул. Хрещатик, 1 (необов'язково)"
						bind:value={newClubAddress}
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0;"
					/>
					<div
						class="flex items-center justify-between rounded-[12px] border px-3 py-2.5 text-[14px] font-medium"
						style="color: #171717; border-color: #e0e0e0; background: #f5f5f5;"
					>
						<span>Київ, Україна</span>
						<i class="fi fi-rr-lock" style="font-size: 12px; color: #aeb4bc;"></i>
					</div>
				</div>

				<!-- Club: description -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРО КЛУБ</label>
					<textarea
						placeholder="Короткий опис вашого клубу, спеціалізація, рівні…"
						bind:value={newClubDescription}
						rows="3"
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0; resize: none;"
					></textarea>
				</div>

				<!-- Club: website + phone -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КОНТАКТИ</label>
					<input
						type="url"
						placeholder="https://example.com"
						bind:value={newClubWebsite}
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0;"
					/>
					<input
						type="tel"
						inputmode="tel"
						autocomplete="tel"
						placeholder="+380 50 123 45 67"
						value={newClubPhone}
						oninput={handleClubPhoneInput}
						onpaste={handleClubPhonePaste}
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0;"
					/>
				</div>

				<!-- Club: working hours -->
				<div class="rounded-[20px] ob-card p-4" style="background: white; display: flex; flex-direction: column; gap: 10px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ГРАФІК РОБОТИ</label>
					{#each newClubWorkingHoursDays as day}
						<div class="flex items-center gap-2">
							<button
								type="button"
								onclick={() => toggleWorkingDay(day.key)}
								class="flex h-7 w-9 flex-shrink-0 items-center justify-center rounded-[8px] text-[12px] font-semibold transition-all"
								style="background: {day.enabled ? '#8984da' : 'transparent'}; color: {day.enabled ? 'white' : '#696969'}; border: 1.5px solid {day.enabled ? '#8984da' : '#d1d5db'};"
							>{day.label}</button>
							{#if day.enabled}
								<input
									type="time"
									value={day.open}
									oninput={(e) => setWorkingHour(day.key, 'open', (e.currentTarget as HTMLInputElement).value)}
									class="flex-1 rounded-[8px] border px-2 py-1.5 text-[13px] font-medium outline-none"
									style="color: #171717; border-color: #e0e0e0;"
								/>
								<span class="text-[12px] font-medium" style="color: #aeb4bc;">—</span>
								<input
									type="time"
									value={day.close}
									oninput={(e) => setWorkingHour(day.key, 'close', (e.currentTarget as HTMLInputElement).value)}
									class="flex-1 rounded-[8px] border px-2 py-1.5 text-[13px] font-medium outline-none"
									style="color: #171717; border-color: #e0e0e0;"
								/>
							{:else}
								<span class="text-[13px] font-medium" style="color: #aeb4bc;">Зачинено</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex-shrink-0 pt-3 pb-4">
			<button
				onclick={accountType === 'trainer' ? handleFinishTrainer : handleFinishClub}
				disabled={isSaving || (accountType === 'trainer' ? firstName.trim().length < 2 : newClubName.trim().length < 2)}
				class="w-full rounded-[50px] py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
				style="background: #8984da;"
			>
				{isSaving ? 'Збереження…' : 'Готово'}
			</button>
		</div>

		<!-- Avatar crop modal (shared with main flow) -->
	{#if cropMode}
		<div class="fixed inset-0 z-50 flex items-center justify-center p-6" style="background: rgba(0,0,0,0.6);">
			<div class="flex flex-col gap-4 rounded-[20px] bg-white ob-card p-5">
				<p class="text-[14px] font-semibold" style="color: #171717;">{$t('onboarding.crop_title')}</p>
				<div
					class="relative overflow-hidden rounded-full"
					style="width: {CROP_SIZE}px; height: {CROP_SIZE}px; background: #dae1eb; touch-action: none;"
					onmousedown={onCropMouseDown}
					ontouchstart={onCropTouchStart}
					role="application"
					tabindex="-1"
				>
					<img
						bind:this={cropImgEl}
						src={avatarPreview}
						alt="crop"
						onload={initCrop}
						draggable="false"
						class="absolute select-none"
						style="left: {cropX}px; top: {cropY}px; width: {cropImgW}px; height: {cropImgH}px; user-select: none; pointer-events: none;"
					/>
				</div>
				<div class="flex gap-2">
					<button
						onclick={() => { cropMode = false; avatarPreview = ''; avatarFile = null; }}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold"
						style="background: rgba(174,180,188,0.15); color: #aeb4bc;"
					>{$t('onboarding.crop_cancel')}</button>
					<button
						onclick={confirmCrop}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold text-white"
						style="background: #8984da;"
					>{$t('onboarding.crop_confirm')}</button>
				</div>
			</div>
		</div>
	{/if}
	</div>
{:else}
<div class="flex h-[100dvh] flex-col px-6 pt-safe pb-safe" style="background: #dae1eb; overflow: hidden;">
	<!-- Progress bar -->
	<div class="flex-shrink-0 pb-10 pt-14">
		<div class="mb-6 flex gap-2">
			{#each Array(TOTAL_STEPS) as _, i}
				<button
					onclick={() => {
						if (i < step - 1) step = i + 1;
						else if (i === step && canAdvance()) step = i + 1;
					}}
					class="flex-1 flex items-center py-3"
					style="background: transparent; border: none; padding-left: 0; padding-right: 0; cursor: {i < step - 1 || (i === step && canAdvance()) ? 'pointer' : 'default'};"
				>
					<div class="h-1 w-full rounded-full transition-all" style="background: {i < step ? '#8984da' : '#c5cdd8'};"></div>
				</button>
			{/each}
		</div>

		{#if step === 1}
			<h1 class="text-[28px] font-black" style="color: #171717;">{$t('onboarding.step1_title')}</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">{$t('onboarding.step1_subtitle')}</p>
		{:else if step === 2}
			<h1 class="text-[28px] font-black" style="color: #171717;">{$t('onboarding.step2_title')}</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">{$t('onboarding.step2_subtitle')}</p>
		{:else if step === 3}
			<h1 class="text-[28px] font-black" style="color: #171717;">{$t('onboarding.step3_title')}</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">{$t('onboarding.step3_subtitle')}</p>
		{:else if step === 4}
			<h1 class="text-[28px] font-black" style="color: #171717;">{$t('onboarding.step4_title')}</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">{isParent ? $t('onboarding.step4_subtitle_parent') : $t('onboarding.step4_subtitle_dancer')}</p>
		{:else}
			<h1 class="text-[28px] font-black" style="color: #171717;">{$t('onboarding.step5_title')}</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">{$t('onboarding.step5_subtitle')}</p>
		{/if}
	</div>

	<!-- Step content with fade edges -->
	<div class="relative flex-1 min-h-0">
		<div class="pointer-events-none absolute inset-x-0 top-0 z-10 h-6" style="background: linear-gradient(to bottom, #dae1eb, transparent);"></div>
		<div class="pointer-events-none absolute inset-x-0 bottom-0 z-10 h-10" style="background: linear-gradient(to top, #dae1eb, transparent);"></div>
	<div class="h-full overflow-y-auto" style="gap: 16px; display: flex; flex-direction: column; padding-top: 8px; padding-bottom: 24px; -webkit-overflow-scrolling: touch;">
		{#if step === 1}
			<!-- Account type -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.account_type_label')}</label>
				<div class="flex gap-2">
					<button
						onclick={() => (accountType = 'dancer')}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {accountType === 'dancer' ? '#8984da' : 'transparent'}; color: {accountType === 'dancer' ? 'white' : '#696969'}; border: 1.5px solid {accountType === 'dancer' ? '#8984da' : '#d1d5db'};"
					>{$t('onboarding.account_dancer')}</button>
					<button
						onclick={() => (accountType = 'parent')}
						class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {accountType === 'parent' ? '#8984da' : 'transparent'}; color: {accountType === 'parent' ? 'white' : '#696969'}; border: 1.5px solid {accountType === 'parent' ? '#8984da' : '#d1d5db'};"
					>{$t('onboarding.account_parent')}</button>
				</div>
			</div>

			<!-- Name -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.name_label_parent') : $t('onboarding.name_label_dancer')}</label>
				<input
					type="text"
					placeholder={isParent ? $t('onboarding.name_placeholder_parent') : $t('onboarding.name_placeholder_dancer')}
					bind:value={firstName}
					autofocus
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717; border-bottom: 1px solid #e0e0e0; padding-bottom: 8px;"
				/>
				<input
					type="text"
					placeholder={isParent ? $t('onboarding.surname_placeholder_parent') : $t('onboarding.surname_placeholder_dancer')}
					bind:value={lastName}
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717;"
				/>
			</div>

			<!-- Gender -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.gender_label_parent') : $t('onboarding.gender_label_dancer')}</label>
				<div class="flex gap-2">
					{#each GENDER_OPTIONS as g}
						<button
							onclick={() => (gender = g.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {gender === g.value ? '#8984da' : 'transparent'}; color: {gender === g.value ? 'white' : '#696969'}; border: 1.5px solid {gender === g.value ? '#8984da' : '#d1d5db'};"
						>{g.label}</button>
					{/each}
				</div>
			</div>

			<!-- Birth date + Height -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<div class="flex items-center justify-between">
					<span class="text-[14px] font-semibold" style="color: #171717;">{isParent ? $t('onboarding.birth_date_label_parent') : $t('onboarding.birth_date_label_dancer')}</span>
					<input
						type="date"
						bind:value={birthDate}
						class="bg-transparent text-[14px] font-medium outline-none"
						style="color: #8984da;"
					/>
				</div>
				{#if ageError}
					<p class="text-[12px] font-medium" style="color: #e74c3c;">{ageError}</p>
				{/if}
				<div class="flex items-center justify-between" style="border-top: 1px solid #f0f0f0; padding-top: 12px;">
					<span class="text-[14px] font-semibold" style="color: #171717;">{isParent ? $t('onboarding.height_label_parent') : $t('onboarding.height_label_dancer')}</span>
					<input
						type="number"
						placeholder="—"
						bind:value={heightCm}
						min="100"
						max="250"
						class="w-[80px] bg-transparent text-right text-[14px] font-medium outline-none"
						style="color: #8984da;"
					/>
				</div>
			</div>

		{:else if step === 2}
			<!-- Goal first: it gates the rest of step 2 (amateurs see a short flow). -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.goal_label_parent') : $t('onboarding.goal_label_dancer')}</label>
				<div class="grid grid-cols-2 gap-2">
					{#each GOAL_OPTIONS as g}
						<button
							onclick={() => (goal = g.value)}
							class="rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
							style="background: {goal === g.value ? '#8984da' : 'transparent'}; color: {goal === g.value ? 'white' : '#696969'}; border: 1.5px solid {goal === g.value ? '#8984da' : '#d1d5db'};"
						>{g.label}</button>
					{/each}
				</div>
			</div>

			<!-- Program (everyone) -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.program_label_parent') : $t('onboarding.program_label_dancer')}</label>
				<div class="flex gap-2">
					{#each PROGRAM_OPTIONS as p}
						<button
							onclick={() => (danceProgram = p.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {danceProgram === p.value ? '#8984da' : 'transparent'}; color: {danceProgram === p.value ? 'white' : '#696969'}; border: 1.5px solid {danceProgram === p.value ? '#8984da' : '#d1d5db'};"
						>{p.label}</button>
					{/each}
				</div>
			</div>

			{#if !isAmateur}
				<!-- Categories (pro only — amateurs don't compete) -->
				<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.category_label_parent') : $t('onboarding.category_label_dancer')}</label>
					<div class="flex flex-wrap gap-2">
						{#each CATEGORIES_UA as cat}
							<button
								onclick={() => toggleUserCategory(cat.value)}
								class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
								style="background: {userCategories.includes(cat.value) ? '#8984da' : 'transparent'}; color: {userCategories.includes(cat.value) ? 'white' : '#696969'}; border: 1.5px solid {userCategories.includes(cat.value) ? '#8984da' : '#d1d5db'};"
							>{cat.label}</button>
						{/each}
					</div>
				</div>

				<!-- Ready to relocate (pro only) -->
				<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? $t('onboarding.relocate_label_parent') : $t('onboarding.relocate_label_dancer')}</label>
					<div class="flex gap-2">
						<button
							onclick={() => (readyToRelocate = readyToRelocate === true ? null : true)}
							class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
							style="background: {readyToRelocate === true ? '#8984da' : 'transparent'}; color: {readyToRelocate === true ? 'white' : '#696969'}; border: 1.5px solid {readyToRelocate === true ? '#8984da' : '#d1d5db'};"
						>{$t('onboarding.yes')}</button>
						<button
							onclick={() => (readyToRelocate = readyToRelocate === false ? null : false)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {readyToRelocate === false ? '#8984da' : 'transparent'}; color: {readyToRelocate === false ? 'white' : '#696969'}; border: 1.5px solid {readyToRelocate === false ? '#8984da' : '#d1d5db'};"
						>{$t('onboarding.no')}</button>
					</div>
				</div>

				<!-- Ready to finance (pro only) -->
				<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.finance_label')}</label>
					<div class="flex gap-2">
						{#each FINANCE_OPTIONS as opt}
							<button
								onclick={() => (readyToFinance = readyToFinance === opt.value ? '' : opt.value)}
								class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
								style="background: {readyToFinance === opt.value ? '#8984da' : 'transparent'}; color: {readyToFinance === opt.value ? 'white' : '#696969'}; border: 1.5px solid {readyToFinance === opt.value ? '#8984da' : '#d1d5db'};"
							>{opt.label}</button>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Bio (everyone, optional) -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 8px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.bio_label')}</label>
				<textarea
					placeholder={$t('onboarding.bio_placeholder')}
					bind:value={bio}
					rows="4"
					class="w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
					style="color: #171717;"
				></textarea>
			</div>

		<!-- User's own location — v1 locked to Kyiv/Ukraine -->
		<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.country_label')}</label>
			<div class="flex items-center justify-between">
				<span class="text-[16px] font-semibold" style="color: #171717;">{HARDCODED_COUNTRY}</span>
				<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
			</div>
			<div style="border-top: 1px solid #f0f0f0; padding-top: 12px; display: flex; flex-direction: column; gap: 8px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.city_label')}</label>
				<div class="flex items-center justify-between">
					<span class="text-[16px] font-semibold" style="color: #171717;">{HARDCODED_CITY}</span>
					<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
				</div>
			</div>
		</div>
		<!-- FUTURE multi-city: restore this block and remove the locked display above.
		<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
			<label>{$t('onboarding.country_label')}</label>
			<select bind:value={country} onchange={() => { city = ''; }} disabled ...>
				{#each COUNTRIES as c}<option value={c}>{c}</option>{/each}
			</select>
			<div ...>
				<label>{$t('onboarding.city_label')}</label>
				{#if CITIES_BY_COUNTRY[country]}
					<select bind:value={city} ...>
						<option value="">Оберіть місто</option>
						{#each CITIES_BY_COUNTRY[country] as c}<option value={c}>{c}</option>{/each}
					</select>
				{:else}
					<input type="text" bind:value={city} placeholder="Місто" ... />
				{/if}
			</div>
		</div>
		-->

		{:else if step === 3}
			<!-- Club search / select / create -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ТВІЙ КЛУБ</label>
				<p class="text-[13px] font-medium" style="color: #696969;">Обери клуб, до якого ходиш — або пропусти цей крок</p>

				{#if selectedClubId && !showCreateForm}
					<div class="flex items-center justify-between rounded-[12px] px-3 py-2.5" style="background: #f0effe;">
						<div>
							<p class="text-[14px] font-semibold" style="color: #8984da;">{selectedClubName}</p>
						</div>
						<button onclick={() => { selectedClubId = null; selectedClubSlug = ''; selectedClubName = ''; wasJustCreated = false; }}>
							<i class="fi fi-rr-cross-small" style="font-size: 16px; color: #aeb4bc; line-height: 1;"></i>
						</button>
					</div>
				{:else if !showCreateForm}
					<input
						type="search"
						placeholder="Назва клубу або місто"
						bind:value={clubSearchQuery}
						oninput={handleClubSearchInput}
						class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
						style="color: #171717; border-color: #e0e0e0;"
					/>
					{#if clubResults.length > 0}
						<div style="display: flex; flex-direction: column; gap: 4px;">
							{#each clubResults as club}
								<button
									onclick={() => { selectedClubId = club.id; selectedClubSlug = club.slug; selectedClubName = club.name; wasJustCreated = false; clubSearchQuery = ''; clubResults = []; }}
									class="flex items-center justify-between rounded-[12px] px-3 py-2.5 text-left"
									style="background: #f8f8f8;"
								>
									<div>
										<p class="text-[14px] font-semibold" style="color: #171717;">{club.name}</p>
										<p class="text-[12px] font-medium" style="color: #aeb4bc;">{club.city}</p>
									</div>
									<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #d1d5db; line-height: 1;"></i>
								</button>
							{/each}
						</div>
					{/if}
					<button
						onclick={() => { showCreateForm = true; }}
						class="text-left text-[13px] font-semibold"
						style="color: #8984da;"
					>Не знайшли? Створіть клуб</button>
				{:else}
					<!-- Create club inline form -->
					<div style="display: flex; flex-direction: column; gap: 10px;">
						<!-- Google Maps import -->
						{#if !showGmapsInputOnboarding}
							<button
								type="button"
								onclick={() => (showGmapsInputOnboarding = true)}
								class="flex items-center justify-center gap-2 rounded-[50px] py-2 text-[13px] font-semibold"
								style="background: rgba(137,132,218,0.12); color: #8984da;"
							>
								<i class="fi fi-rr-link" style="font-size: 14px; line-height: 1;"></i>
								Імпорт з Google Maps
							</button>
						{:else}
							<div style="display: flex; flex-direction: column; gap: 6px;">
								<input
									type="url"
									bind:value={gmapsURLOnboarding}
									placeholder="Вставте посилання Google Maps…"
									class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
									style="color: #171717; border-color: #e0e0e0;"
								/>
								<div class="flex gap-2">
									<button
										type="button"
										onclick={() => { showGmapsInputOnboarding = false; gmapsURLOnboarding = ''; }}
										class="flex-1 rounded-[50px] py-2 text-[13px] font-semibold"
										style="background: transparent; color: #696969; border: 1.5px solid #d1d5db;"
									>Скасувати</button>
									<button
										type="button"
										onclick={importFromGmapsOnboarding}
										disabled={isImportingGmaps || !gmapsURLOnboarding.trim()}
										class="flex flex-1 items-center justify-center gap-2 rounded-[50px] py-2 text-[13px] font-semibold text-white transition-opacity disabled:opacity-50"
										style="background: #8984da;"
									>
										{#if isImportingGmaps}
											<div class="h-4 w-4 animate-spin rounded-full border-2 border-white/30" style="border-top-color: white;"></div>
										{/if}
										Імпорт
									</button>
								</div>
							</div>
						{/if}

						<!-- Imported photos preview -->
						{#if newClubPhotos.length > 0}
							<div class="flex gap-2 overflow-x-auto pb-1">
								{#each newClubPhotos as url}
									<img src={url} alt="Фото клубу" class="h-[64px] w-[64px] flex-shrink-0 rounded-[10px] object-cover" />
								{/each}
							</div>
						{/if}

						<input
							type="text"
							placeholder="Назва клубу"
							bind:value={newClubName}
							class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
							style="color: #171717; border-color: #e0e0e0;"
						/>
						<input
							type="text"
							placeholder="Вулиця, номер (необов'язково)"
							bind:value={newClubAddress}
							class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
							style="color: #171717; border-color: #e0e0e0;"
						/>
						<!-- City locked to Київ for v1 -->
						<div
							class="flex items-center justify-between rounded-[12px] border px-3 py-2.5 text-[14px] font-medium"
							style="color: #171717; border-color: #e0e0e0; background: #f5f5f5;"
						>
							<span>Київ, Україна</span>
							<i class="fi fi-rr-lock" style="font-size: 12px; color: #aeb4bc;"></i>
						</div>
						<div class="flex gap-2">
							<button
								onclick={() => {
									showCreateForm = false;
									showGmapsInputOnboarding = false;
									gmapsURLOnboarding = '';
									newClubPhotos = [];
									newClubLat = null;
									newClubLng = null;
								}}
								class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold"
								style="background: transparent; color: #696969; border: 1.5px solid #d1d5db;"
							>{$t('onboarding.step_club_form_cancel')}</button>
							<button
								onclick={createAndJoinClub}
								disabled={!newClubName.trim() || isCreatingClub}
								class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold text-white transition-opacity disabled:opacity-50"
								style="background: #8984da;"
							>{isCreatingClub ? $t('onboarding.step_club_creating') : $t('onboarding.step_club_form_submit')}</button>
						</div>
					</div>
				{/if}
			</div>

			{#if !selectedClubId && !showCreateForm}
				<p class="text-center text-[12px] font-medium" style="color: #aeb4bc;">Можна пропустити і додати клуб пізніше</p>
			{/if}

		{:else if step === 4}
			<!-- Preferred gender -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.pref_gender_label')}</label>
				<div class="flex gap-2">
					{#each [{ value: 'male', label: $t('onboarding.gender_male') }, { value: 'female', label: $t('onboarding.gender_female') }] as g}
						<button
							onclick={() => (prefGender = prefGender === g.value ? '' : g.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {prefGender === g.value ? '#8984da' : 'transparent'}; color: {prefGender === g.value ? 'white' : '#696969'}; border: 1.5px solid {prefGender === g.value ? '#8984da' : '#d1d5db'};"
						>{g.label}</button>
					{/each}
				</div>
			</div>

			<!-- Age range -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.pref_age_label')}</label>
				<div class="flex items-center gap-3">
					<input
						type="number"
						placeholder={$t('onboarding.age_from')}
						bind:value={ageMin}
						min="4"
						max="80"
						class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
						style="color: #171717; border-color: #e0e0e0;"
					/>
					<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
					<input
						type="number"
						placeholder="До"
						bind:value={ageMax}
						min="4"
						max="80"
						class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
						style="color: #171717; border-color: #e0e0e0;"
					/>
				</div>
			</div>

			<!-- Height range -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{isParent ? 'ЗРІСТ ПАРТНЕРА ДИТИНИ (СМ)' : 'ЗРІСТ ПАРТНЕРА (СМ)'}</label>
				<div class="flex items-center gap-3">
					<input
						type="number"
						placeholder="Від"
						bind:value={heightMin}
						min="100"
						max="220"
						class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
						style="color: #171717; border-color: #e0e0e0;"
					/>
					<span class="text-[14px] font-medium" style="color: #aeb4bc;">—</span>
					<input
						type="number"
						placeholder="До"
						bind:value={heightMax}
						min="100"
						max="220"
						class="flex-1 rounded-[12px] border px-3 py-2 text-[14px] font-semibold outline-none text-center"
						style="color: #171717; border-color: #e0e0e0;"
					/>
				</div>
			</div>

			<!-- Program (everyone) -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРОГРАМА</label>
				<div class="flex gap-2">
					{#each [{ value: 'standard', label: 'Стандарт' }, { value: 'latina', label: 'Латина' }, { value: 'both', label: 'Обидві' }] as p}
						<button
							onclick={() => (prefProgram = prefProgram === p.value ? '' : p.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
							style="background: {prefProgram === p.value ? '#8984da' : 'transparent'}; color: {prefProgram === p.value ? 'white' : '#696969'}; border: 1.5px solid {prefProgram === p.value ? '#8984da' : '#d1d5db'};"
						>{p.label}</button>
					{/each}
				</div>
			</div>

			{#if !isAmateur}
				<!-- Preferred goal (pro only) -->
				<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">БАЖАНА ЦІЛЬ</label>
					<div class="flex gap-2">
						{#each [{ value: 'hobby', label: 'Хобі' }, { value: 'professional', label: 'Профі' }] as g}
							<button
								onclick={() => {
									prefGoal = prefGoal === g.value ? '' : g.value;
									if (prefGoal === 'hobby') prefCategories = [];
								}}
								class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
								style="background: {prefGoal === g.value ? '#8984da' : 'transparent'}; color: {prefGoal === g.value ? 'white' : '#696969'}; border: 1.5px solid {prefGoal === g.value ? '#8984da' : '#d1d5db'};"
							>{g.label}</button>
						{/each}
					</div>
				</div>

				{#if prefGoal !== 'hobby'}
					<!-- Categories: hidden when partner goal is hobby (either own or partner) -->
					<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
						<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КАТЕГОРІЯ</label>
						<div class="flex flex-wrap gap-2">
							{#each CATEGORIES_UA as cat}
								<button
									onclick={() => togglePrefCategory(cat.value)}
									class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
									style="background: {prefCategories.includes(cat.value) ? '#8984da' : 'transparent'}; color: {prefCategories.includes(cat.value) ? 'white' : '#696969'}; border: 1.5px solid {prefCategories.includes(cat.value) ? '#8984da' : '#d1d5db'};"
								>{cat.label}</button>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Wants partner to finance (pro only) -->
				<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ФІНАНСУВАННЯ ПАРТНЕРА</label>
					<div class="flex gap-2">
						{#each [{ value: 'no', label: 'Ні' }, { value: 'yes', label: 'Так' }, { value: 'partial', label: 'Частково' }] as f}
							<button
								onclick={() => (wantsFinance = wantsFinance === f.value ? '' : f.value)}
								class="flex-1 rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
								style="background: {wantsFinance === f.value ? '#8984da' : 'transparent'}; color: {wantsFinance === f.value ? 'white' : '#696969'}; border: 1.5px solid {wantsFinance === f.value ? '#8984da' : '#d1d5db'};"
							>{f.label}</button>
						{/each}
					</div>
				</div>
			{/if}

		{:else}
			<!-- Photo -->
			<div class="flex flex-col items-center gap-3 rounded-[20px] bg-white p-6">
				<button
					onclick={() => fileInput?.click()}
					class="relative flex h-[100px] w-[100px] items-center justify-center overflow-hidden rounded-full"
					style="background: #e0e0e0;"
				>
					{#if avatarPreview}
						<img src={avatarPreview} alt="Аватар" class="h-full w-full object-cover" />
					{:else}
						<i class="fi fi-rr-user" style="font-size: 40px; color: #696969;"></i>
					{/if}
					<div class="absolute inset-0 flex items-center justify-center rounded-full" style="background: rgba(0,0,0,0.2);">
						<i class="fi fi-rr-camera text-white" style="font-size: 22px; line-height: 1;"></i>
					</div>
				</button>
				{#if avatarPreview}
					<button
						onclick={() => fileInput?.click()}
						class="text-[13px] font-semibold"
						style="color: #8984da;"
					>Змінити фото</button>
				{:else}
					<span class="text-[14px] font-medium" style="color: #696969;">Додати фото профілю</span>
					<span class="text-center text-[12px] font-medium" style="color: #aeb4bc;">Необов'язково — можна додати пізніше</span>
				{/if}
				<input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleAvatarChange} />
			</div>

			<!-- Extra photos -->
			<div class="rounded-[20px] bg-white ob-card p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<div class="flex items-center justify-between">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ДОДАТКОВІ ФОТО</label>
					<span class="text-[11px] font-medium" style="color: #aeb4bc;">{extraPhotoPreviews.length}/5</span>
				</div>
				<div class="flex gap-3 overflow-x-auto pb-1" style="-webkit-overflow-scrolling: touch;">
					{#each extraPhotoPreviews as preview, i}
						<div class="relative flex-shrink-0" style="width: 80px; height: 80px;">
							<img src={preview} alt="Фото" class="h-full w-full rounded-[12px] object-cover" />
							<button
								onclick={() => removeExtraPhoto(i)}
								class="absolute -top-1.5 -right-1.5 flex h-5 w-5 items-center justify-center rounded-full"
								style="background: #e74c3c;"
							>
								<i class="fi fi-rr-cross-small text-white" style="font-size: 10px; line-height: 1;"></i>
							</button>
						</div>
					{/each}
					{#if extraPhotoPreviews.length < 5}
						<button
							onclick={() => photoFileInput?.click()}
							class="flex flex-shrink-0 items-center justify-center rounded-[12px]"
							style="width: 80px; height: 80px; background: #f0f0f0; border: 1.5px dashed #d1d5db;"
						>
							<i class="fi fi-rr-plus" style="font-size: 20px; color: #aeb4bc;"></i>
						</button>
					{/if}
				</div>
				<p class="text-[11px] font-medium" style="color: #aeb4bc;">Необов'язково — можна додати пізніше</p>
				<input bind:this={photoFileInput} type="file" accept="image/*" class="hidden" onchange={handlePhotoChange} />
			</div>
		{/if}
	</div>
	</div>

	<!-- Crop modal -->
	{#if cropMode}
		<div class="fixed inset-0 z-50 flex flex-col items-center justify-center" style="background: #000;">
			<p class="mb-5 text-[15px] font-semibold text-white">Перетягни, щоб обрізати</p>

			<!-- crop viewport: clip-path forces proper circular clipping on WebKit -->
			<div
				role="slider"
				aria-valuenow={0}
				tabindex="0"
				style="position: relative; width: {CROP_SIZE}px; height: {CROP_SIZE}px; overflow: hidden; border-radius: 50%; cursor: grab; touch-action: none; -webkit-transform: translateZ(0); transform: translateZ(0); will-change: transform;"
				onmousedown={onCropMouseDown}
				ontouchstart={onCropTouchStart}
			>
				<img
					bind:this={cropImgEl}
					src={avatarPreview}
					alt=""
					onload={initCrop}
					style="position: absolute; left: {cropX}px; top: {cropY}px; width: {cropImgW}px; height: {cropImgH}px; user-select: none; pointer-events: none;"
					draggable="false"
				/>
			</div>

			<div class="mt-7 flex gap-4">
				<button
					onclick={() => { cropMode = false; avatarPreview = ''; avatarFile = null; }}
					class="rounded-full px-6 py-3 text-[15px] font-semibold"
					style="background: rgba(255,255,255,0.15); color: white;"
				>Скасувати</button>
				<button
					onclick={confirmCrop}
					class="rounded-full px-6 py-3 text-[15px] font-semibold"
					style="background: #8984da; color: white;"
				>Підтвердити</button>
			</div>
		</div>
	{/if}

	<!-- Extra photo crop modal -->
	{#if extraCropMode}
		<div class="fixed inset-0 z-50 flex flex-col items-center justify-center" style="background: #000;">
			<p class="mb-5 text-[15px] font-semibold text-white">Перетягни, щоб обрізати</p>
			<div
				role="slider"
				aria-valuenow={0}
				tabindex="0"
				style="position: relative; width: {CROP_SIZE}px; height: {CROP_SIZE}px; overflow: hidden; border-radius: 12px; cursor: grab; touch-action: none; -webkit-transform: translateZ(0); transform: translateZ(0); will-change: transform;"
				onmousedown={onExtraCropMouseDown}
				ontouchstart={onExtraCropTouchStart}
			>
				<img
					bind:this={extraCropImgEl}
					src={extraCropPreview}
					alt=""
					onload={initExtraCrop}
					style="position: absolute; left: {extraCropX}px; top: {extraCropY}px; width: {extraCropImgW}px; height: {extraCropImgH}px; user-select: none; pointer-events: none;"
					draggable="false"
				/>
			</div>
			<div class="mt-7 flex gap-4">
				<button
					onclick={() => { extraCropMode = false; extraCropPreview = ''; }}
					class="rounded-full px-6 py-3 text-[15px] font-semibold"
					style="background: rgba(255,255,255,0.15); color: white;"
				>Скасувати</button>
				<button
					onclick={confirmExtraCrop}
					class="rounded-full px-6 py-3 text-[15px] font-semibold"
					style="background: #8984da; color: white;"
				>Підтвердити</button>
			</div>
		</div>
	{/if}

	<!-- Navigation buttons -->
	<div class="flex-shrink-0 flex gap-3 pb-4 pt-6" style="position: relative;">
		<div class="pointer-events-none absolute inset-x-0 -top-8 h-8" style="background: linear-gradient(to top, #dae1eb, transparent);"></div>
		{#if step > 1}
			<button
				onclick={() => (step -= 1)}
				class="flex h-[50px] w-[50px] flex-shrink-0 items-center justify-center rounded-full"
				style="background: white; border: 1.5px solid #d1d5db;"
			>
				<i class="fi fi-rr-angle-left" style="font-size: 18px; color: #696969; line-height: 1;"></i>
			</button>
		{/if}

		{#if step < TOTAL_STEPS}
			<button
				onclick={() => (step += 1)}
				disabled={!canAdvance()}
				class="flex h-[50px] flex-1 items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-colors"
				style="background: {canAdvance() ? '#000000' : '#aeb4bc'};"
			>
				{$t('onboarding.btn_next')}
			</button>
		{:else}
			<button
				onclick={handleFinish}
				disabled={isSaving}
				class="flex h-[50px] flex-1 items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
				style="background: #8984da;"
			>
				{isSaving ? $t('onboarding.btn_saving') : $t('onboarding.btn_finish')}
			</button>
		{/if}
	</div>
</div>
{/if}

<style>
	/* Onboarding is a light-only design. Prevent dark mode from inverting cards. */
	:global(html.dark) .ob-card {
		background: white !important;
	}
</style>
