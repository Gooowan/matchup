<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { captureOnboardingComplete } from '$lib/analytics/posthog';
	import toast from 'svelte-french-toast';
	import { t } from '$lib/locale';

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
	let accountType = $state<'dancer' | 'parent'>('dancer');
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
	let newClubCity = $state('');
	let wasJustCreated = $state(false);
	let clubSearchTimer: ReturnType<typeof setTimeout> | null = null;

	// Step 4: Location + photo
	let city = $state('');
	let country = $state('Україна'); // locked to Ukraine for v1; remove disabled attr below when unlocking
	let avatarFile = $state<File | null>(null);
	let avatarPreview = $state('');
	let fileInput = $state<HTMLInputElement | null>(null);

	// Extra photos (deferred upload at handleFinish)
	let extraPhotoFiles = $state<File[]>([]);
	let extraPhotoPreviews = $state<string[]>([]);
	let photoFileInput = $state<HTMLInputElement | null>(null);

	let isSaving = $state(false);

	onMount(() => {
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
				city, country
			})
		);
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

	const UA_CITY_CENTROIDS: Record<string, { lat: number; lng: number }> = {
		'Київ': { lat: 50.4501, lng: 30.5234 },
		'Львів': { lat: 49.8397, lng: 24.0297 },
		'Одеса': { lat: 46.4825, lng: 30.7233 },
		'Харків': { lat: 49.9935, lng: 36.2304 },
		'Дніпро': { lat: 48.4647, lng: 35.0462 }
	};

	async function geocodeClub(address: string, cityName: string): Promise<{ lat: number; lng: number }> {
		try {
			const q = encodeURIComponent(`${cityName}, Ukraine${address ? ', ' + address : ''}`);
			const resp = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${q}&limit=1`);
			if (resp.ok) {
				const results = await resp.json();
				if (results.length > 0) {
					return { lat: parseFloat(results[0].lat), lng: parseFloat(results[0].lon) };
				}
			}
		} catch { /* fallback */ }
		return UA_CITY_CENTROIDS[cityName] ?? { lat: 50.4501, lng: 30.5234 };
	}

	function handleClubSearchInput() {
		if (clubSearchTimer) clearTimeout(clubSearchTimer);
		clubSearchTimer = setTimeout(searchClubs, 300);
	}

	async function searchClubs() {
		if (!clubSearchQuery.trim()) { clubResults = []; return; }
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

	async function createAndJoinClub() {
		if (!newClubName.trim()) return;
		isCreatingClub = true;
		try {
			const clubCity = newClubCity.trim() || city;
			const geocoded = await geocodeClub(newClubAddress, clubCity);
			const resp = await authFetch('/clubs/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: newClubName.trim(),
					country: 'Ukraine',
					city: clubCity,
					address: newClubAddress.trim() || undefined,
					latitude: geocoded.lat,
					longitude: geocoded.lng
				})
			});
			if (resp.ok) {
				const body = await resp.json();
				selectedClubId = body.data?.id ?? null;
				selectedClubSlug = body.data?.slug ?? '';
				selectedClubName = newClubName.trim();
				wasJustCreated = true;
				showCreateForm = false;
				toast.success('Клуб створено!');
			} else {
				const body = await resp.json().catch(() => ({}));
				toast.error((body as { error?: string }).error ?? 'Помилка створення клубу');
			}
		} catch {
			toast.error('Помилка. Спробуй ще раз.');
		} finally {
			isCreatingClub = false;
		}
	}

	function canAdvance(): boolean {
		if (step === 1) return firstName.length >= 2 && lastName.length >= 2 && !!gender && !ageError;
		if (step === 2) return !!danceProgram;
		return true;
	}

	async function handleFinish() {
		isSaving = true;
		async function postOrThrow(path: string, init: RequestInit) {
			const r = await authFetch(path, init);
			if (!r.ok) {
				const b = await r.json().catch(() => ({}));
				throw new Error((b as { error?: string }).error ?? `Request to ${path} failed (${r.status})`);
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

			if (selectedClubId && !wasJustCreated && selectedClubSlug) {
				try {
					await postOrThrow(`/clubs/${selectedClubSlug}/join`, { method: 'POST' });
				} catch {
					// non-fatal: profile already saved
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<!-- Program -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Categories -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Ready to relocate -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Ready to finance -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Goal -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Bio -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 8px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.bio_label')}</label>
				<textarea
					placeholder={$t('onboarding.bio_placeholder')}
					bind:value={bio}
					rows="4"
					class="w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
					style="color: #171717;"
				></textarea>
			</div>

			<!-- User's own location -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.country_label')}</label>
				<select
					bind:value={country}
					onchange={() => { city = ''; }}
					disabled
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717; opacity: 1; -webkit-appearance: none; appearance: none;"
				>
					{#each COUNTRIES as c}
						<option value={c}>{c}</option>
					{/each}
				</select>
				<div style="border-top: 1px solid #f0f0f0; padding-top: 12px; display: flex; flex-direction: column; gap: 8px;">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">{$t('onboarding.city_label')}</label>
					{#if CITIES_BY_COUNTRY[country]}
						<select
							bind:value={city}
							class="w-full bg-transparent text-[16px] font-semibold outline-none"
							style="color: {city ? '#171717' : '#aeb4bc'};"
						>
							<option value="">Оберіть місто</option>
							{#each CITIES_BY_COUNTRY[country] as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					{:else}
						<input
							type="text"
							placeholder="Місто"
							bind:value={city}
							class="w-full bg-transparent text-[16px] font-semibold outline-none"
							style="color: #171717;"
						/>
					{/if}
				</div>
			</div>

		{:else if step === 3}
			<!-- Club search / select / create -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
						onclick={() => { showCreateForm = true; newClubCity = city; }}
						class="text-left text-[13px] font-semibold"
						style="color: #8984da;"
					>Не знайшли? Створіть клуб</button>
				{:else}
					<!-- Create club inline form -->
					<div style="display: flex; flex-direction: column; gap: 10px;">
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
						<input
							type="text"
							placeholder="Місто"
							bind:value={newClubCity}
							class="w-full rounded-[12px] border px-3 py-2.5 text-[14px] font-medium outline-none"
							style="color: #171717; border-color: #e0e0e0;"
						/>
						<div class="flex gap-2">
							<button
								onclick={() => (showCreateForm = false)}
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Preferred goal -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">БАЖАНА ЦІЛЬ</label>
				<div class="flex gap-2">
					{#each [{ value: 'hobby', label: 'Хобі' }, { value: 'professional', label: 'Профі' }] as g}
						<button
							onclick={() => (prefGoal = prefGoal === g.value ? '' : g.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {prefGoal === g.value ? '#8984da' : 'transparent'}; color: {prefGoal === g.value ? 'white' : '#696969'}; border: 1.5px solid {prefGoal === g.value ? '#8984da' : '#d1d5db'};"
						>{g.label}</button>
					{/each}
				</div>
			</div>

			<!-- Program -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Categories -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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

			<!-- Wants partner to finance -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
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
				class="flex h-[50px] flex-1 items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
				style="background: #171717;"
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
