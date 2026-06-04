<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import toast from 'svelte-french-toast';
	import { parseApiError } from '$lib/utils/parseApiError';

	// v1: locked to Kyiv. To open multi-city: restore UKRAINE_CITIES dropdown below.
	const HARDCODED_CITY_PROFILE = 'Київ';
	const HARDCODED_COUNTRY_PROFILE = 'Україна';
	// const UKRAINE_CITIES = ['Київ', 'Харків', 'Одеса', 'Дніпро', 'Запоріжжя', 'Львів', 'Кривий Ріг', 'Миколаїв', 'Вінниця', 'Херсон', 'Полтава', 'Чернігів', 'Черкаси', 'Суми', 'Житомир', 'Хмельницький', 'Рівне', 'Тернопіль', 'Луцьк', 'Ужгород'];
	const CATEGORIES_UA = [
		{ value: 'kids', label: 'Діти' },
		{ value: 'juvenile1', label: 'Ювенали 1' },
		{ value: 'juvenile2', label: 'Ювенали 2' },
		{ value: 'junior1', label: 'Юніори 1' },
		{ value: 'junior2', label: 'Юніори 2' },
		{ value: 'youth', label: 'Молодь' },
		{ value: 'adult', label: 'Дорослі' }
	];
	const PROGRAM_OPTIONS = [
		{ value: 'latina', label: 'Латина' },
		{ value: 'standard', label: 'Стандарт' },
		{ value: 'both', label: 'Обидва' }
	];
	const GOAL_OPTIONS = [
		{ value: 'hobby', label: 'Хобі' },
		{ value: 'professional', label: 'Профі' }
	];
	const FINANCE_OPTIONS: { value: string; label: string }[] = [
		{ value: 'no', label: 'Ні' },
		{ value: 'yes', label: 'Так' },
		{ value: 'partial', label: 'Частково' }
	];
	const GENDER_OPTIONS = [
		{ value: 'male', label: 'Чоловік' },
		{ value: 'female', label: 'Жінка' },
		{ value: 'other', label: 'Інше' }
	];

	let accountType = $derived(authStore.user?.profile_data?.account_type as string | undefined);
	let isDancerLike = $derived(accountType !== 'trainer' && accountType !== 'club');

	let isSaving = $state(false);
	let isLoading = $state(true);
	let avatarFile = $state<File | null>(null);
	let avatarPreview = $state('');
	let origAvatarPreview = '';
	let fileInput = $state<HTMLInputElement | null>(null);
	let mediaUrls = $state<string[]>([]);
	let photoFileInput = $state<HTMLInputElement | null>(null);
	let isUploadingPhoto = $state(false);

	let firstName = $state('');
	let lastName = $state('');
	let gender = $state('');
	let birthDate = $state('');
	let heightCm = $state<number | null>(null);
	let city = $state('');
	let program = $state('');
	let goal = $state('');
	let bio = $state('');
	let readyToRelocate = $state(false);
	let readyToFinance = $state('');
	let categories = $state<string[]>([]);
	let profileVisible = $state(true);

	onMount(async () => {
		// Club accounts have no dancer-style profile; redirect to Business panel.
		if (authStore.user?.profile_data?.account_type === 'club') {
			goto('/business');
			return;
		}
		const user = authStore.user;
		if (user?.profile_data) {
			const pd = user.profile_data as Record<string, unknown>;
			firstName = (pd.first_name as string) ?? '';
			lastName = (pd.last_name as string) ?? '';
			avatarPreview = (pd.avatar as string) ?? '';
			origAvatarPreview = avatarPreview;
		}
		try {
			const resp = await authFetch('/me/profile');
			if (resp.ok) {
				const body = await resp.json();
				const p = body.data ?? body;
				gender = p.gender ?? '';
				birthDate = p.birth_date?.slice(0, 10) ?? '';
				heightCm = p.height_cm ?? null;
				city = p.city ?? '';
				program = p.program ?? '';
				goal = p.goal ?? '';
				bio = (p.metadata?.bio as string) ?? '';
				readyToRelocate = p.ready_to_relocate ?? false;
				readyToFinance = p.ready_to_finance ?? '';
				categories = p.categories ?? [];
				profileVisible = p.visible ?? true;
				if (Array.isArray(p.metadata?.media_urls)) {
					mediaUrls = p.metadata.media_urls as string[];
				}
			}
		} catch {
			// profile may not exist yet
		} finally {
			isLoading = false;
		}
	});

	// --- Avatar crop ---
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
		origAvatarPreview = avatarPreview;
		avatarPreview = URL.createObjectURL(file);
		cropMode = true;
		(e.target as HTMLInputElement).value = '';
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
		const ctx2d = canvas.getContext('2d')!;
		ctx2d.drawImage(cropImgEl, cropX * ratio, cropY * ratio, cropImgW * ratio, cropImgH * ratio);
		const blob = await new Promise<Blob>((res) => canvas.toBlob((b) => res(b!), 'image/jpeg', 0.92));
		avatarFile = new File([blob], 'avatar.jpg', { type: 'image/jpeg' });
		avatarPreview = URL.createObjectURL(blob);
		origAvatarPreview = avatarPreview;
		cropMode = false;
	}

	// --- Extra photo crop ---
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
		isUploadingPhoto = true;
		try {
			const formData = new FormData();
			formData.append('photo', file);
			const resp = await authFetch('/user/files/photo', { method: 'POST', body: formData });
			if (!resp.ok) throw new Error('upload failed');
			const { data } = await resp.json();
			await authFetch('/me/profile/media', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url: data.url })
			});
			mediaUrls = [...mediaUrls, data.url];
		} catch {
			toast.error('Не вдалося завантажити фото');
		} finally {
			isUploadingPhoto = false;
		}
	}

	async function deletePhoto(url: string) {
		try {
			await authFetch('/me/profile/media', {
				method: 'DELETE',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url })
			});
			mediaUrls = mediaUrls.filter((u) => u !== url);
		} catch {
			toast.error('Не вдалося видалити фото');
		}
	}

	function toggleCategory(v: string) {
		categories = categories.includes(v) ? categories.filter((x) => x !== v) : [...categories, v];
	}

	async function postOrThrow(path: string, init: RequestInit) {
		const r = await authFetch(path, init);
		if (!r.ok) {
			const b = await r.json().catch(() => ({}));
			throw new Error(parseApiError(b, r.status));
		}
		return r;
	}

	async function handleSave() {
		// Client-side validation — instant feedback before hitting the server.
		if (firstName.trim().length < 2) {
			toast.error('Введіть імʼя (мінімум 2 символи)');
			return;
		}
		// City is required only for dancer/parent; trainers use the hardcoded value anyway.
		if (isDancerLike && !city) {
			toast.error('Оберіть місто');
			return;
		}
		if (isDancerLike && heightCm && (heightCm < 100 || heightCm > 250)) {
			toast.error('Зріст має бути від 100 до 250 см');
			return;
		}
		isSaving = true;
		try {
			await postOrThrow('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ first_name: firstName, last_name: lastName })
			});

			if (avatarFile) {
				await authStore.uploadAvatar(avatarFile);
			}

			// Build the profile body per account type: trainers skip partner-matching fields.
			const body: Record<string, unknown> = {
				gender,
				country: HARDCODED_COUNTRY_PROFILE,
				city: HARDCODED_CITY_PROFILE,
				program: program || 'standard',
				bio,
				categories,
				visible: profileVisible,
			};
			if (isDancerLike) {
				body.goal = goal || 'hobby';
				body.ready_to_relocate = readyToRelocate;
				if (readyToFinance) body.ready_to_finance = readyToFinance;
				if (birthDate) body.birth_date = birthDate;
				if (heightCm) body.height_cm = heightCm;
			}

			await postOrThrow('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			await authStore.checkAuth();
			toast.success('Профіль збережено');
			goto('/settings');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Помилка збереження');
		} finally {
			isSaving = false;
		}
	}
</script>

<div class="mu-screen" style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch;">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex flex-shrink-0 items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} class="flex items-center justify-center" aria-label="Назад">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 truncate text-[20px] font-black">Редагувати профіль</h1>
		<button
			onclick={handleSave}
			disabled={isSaving}
			class="text-[14px] font-semibold transition-opacity disabled:opacity-60"
			style="color: #8984da;"
		>
			{isSaving ? 'Збереження…' : 'Зберегти'}
		</button>
	</div>

	{#if isLoading}
		<div class="flex min-h-[60vh] items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
		</div>
	{:else}
		<div class="flex flex-col px-4 pb-[100px]" style="gap: 16px; padding-top: 8px;">

			<!-- Avatar -->
			<div class="mu-card flex flex-col items-center gap-3 rounded-[20px] p-6">
				<button
					onclick={() => fileInput?.click()}
					class="relative flex h-[100px] w-[100px] items-center justify-center overflow-hidden rounded-full"
					style="background: rgba(174,180,188,0.2);"
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
					<button onclick={() => fileInput?.click()} class="text-[13px] font-semibold" style="color: #8984da;">Змінити фото</button>
				{:else}
					<span class="text-[14px] font-medium" style="color: #696969;">Торкнись, щоб змінити фото</span>
				{/if}
				<input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleAvatarChange} />
			</div>

			<!-- Extra photos -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<div class="flex items-center justify-between">
					<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ФОТО</label>
					<span class="text-[11px] font-medium" style="color: #aeb4bc;">{mediaUrls.length}/5</span>
				</div>
				<div class="flex gap-3 overflow-x-auto pb-1" style="-webkit-overflow-scrolling: touch;">
					{#each mediaUrls as url}
						<div class="relative flex-shrink-0" style="width: 80px; height: 80px;">
							<img src={url} alt="Фото" loading="lazy" decoding="async" class="h-full w-full rounded-[12px] object-cover" />
							<button
								onclick={() => deletePhoto(url)}
								class="absolute -top-1.5 -right-1.5 flex h-5 w-5 items-center justify-center rounded-full"
								style="background: #e74c3c;"
							>
								<i class="fi fi-rr-cross-small text-white" style="font-size: 10px; line-height: 1;"></i>
							</button>
						</div>
					{/each}
					{#if mediaUrls.length < 5}
						<button
							onclick={() => photoFileInput?.click()}
							disabled={isUploadingPhoto}
							class="mu-divider flex flex-shrink-0 items-center justify-center rounded-[12px] transition-opacity disabled:opacity-60"
							style="width: 80px; height: 80px; background: rgba(174,180,188,0.12); border: 1.5px dashed rgba(174,180,188,0.5); border-top-style: dashed; border-right-style: dashed; border-bottom-style: dashed; border-left-style: dashed;"
						>
							{#if isUploadingPhoto}
								<div class="h-5 w-5 animate-spin rounded-full border-2" style="border-color: rgba(174,180,188,0.3); border-top-color: #8984da;"></div>
							{:else}
								<i class="fi fi-rr-plus" style="font-size: 20px; color: #aeb4bc;"></i>
							{/if}
						</button>
					{/if}
				</div>
				<input bind:this={photoFileInput} type="file" accept="image/*" class="hidden" onchange={handlePhotoChange} />
			</div>

			<!-- Visibility toggle -->
			<div class="mu-card rounded-[20px] p-4">
				<div class="flex items-center justify-between gap-3">
					<div class="flex flex-col gap-0.5">
						<span class="mu-text-primary text-[14px] font-semibold">Видимість профілю</span>
						<span class="text-[12px] font-medium" style="color: #aeb4bc;">
							{profileVisible ? 'Профіль видно у стрічці та на карті' : 'Профіль прихований від нових знайомств'}
						</span>
					</div>
					<button
						role="switch"
						aria-checked={profileVisible}
						onclick={() => (profileVisible = !profileVisible)}
						class="relative flex-shrink-0 rounded-full transition-colors"
						style="width: 44px; height: 26px; background: {profileVisible ? '#8984da' : '#d1d5db'};"
					>
						<span
							class="absolute rounded-full bg-white shadow transition-transform"
							style="width: 20px; height: 20px; top: 3px; left: 3px; transform: translateX({profileVisible ? '18px' : '0'});"
						></span>
					</button>
				</div>
			</div>

			<!-- Name -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ІМ'Я</label>
				<input
					type="text"
					placeholder="Ім'я"
					bind:value={firstName}
					class="mu-text-primary mu-divider w-full bg-transparent text-[16px] font-semibold outline-none"
					style="border-bottom-width: 1px; border-bottom-style: solid; padding-bottom: 8px;"
				/>
				<input
					type="text"
					placeholder="Прізвище"
					bind:value={lastName}
					class="mu-text-primary w-full bg-transparent text-[16px] font-semibold outline-none"
				/>
			</div>

			<!-- Basic info -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ОСНОВНА ІНФОРМАЦІЯ</label>
				<div class="flex gap-2">
					{#each GENDER_OPTIONS as g}
						<button
							onclick={() => (gender = g.value)}
							class="flex-1 rounded-[50px] py-2 text-[13px] font-semibold transition-all"
							style="background: {gender === g.value ? '#8984da' : 'transparent'}; color: {gender === g.value ? 'white' : '#696969'}; border: 1.5px solid {gender === g.value ? '#8984da' : 'rgba(174,180,188,0.4)'};"
						>{g.label}</button>
					{/each}
				</div>
			{#if isDancerLike}
				<div class="mu-divider flex items-center justify-between" style="border-top-width: 1px; border-top-style: solid; padding-top: 12px;">
					<span class="mu-text-primary shrink-0 text-[14px] font-semibold">Дата народження</span>
					<input
						type="date"
						bind:value={birthDate}
						class="shrink-0 bg-transparent text-[14px] font-medium outline-none"
						style="color: #8984da;"
					/>
				</div>
				<div class="mu-divider flex items-center justify-between" style="border-top-width: 1px; border-top-style: solid; padding-top: 12px;">
					<span class="mu-text-primary text-[14px] font-semibold">Зріст (см)</span>
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
			{/if}
			</div>

		<!-- v1: locked to Kyiv / Ukraine.
		     FUTURE multi-city: restore the UKRAINE_CITIES city <select> and country
		     <select>/<input> below; also uncomment UKRAINE_CITIES in the script. -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">МІСЦЕЗНАХОДЖЕННЯ</label>
				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[16px] font-semibold">{HARDCODED_CITY_PROFILE}</span>
					<i class="fi fi-rr-lock" style="font-size: 13px; color: #aeb4bc;"></i>
				</div>
				<div class="mu-divider" style="border-top-width: 1px; border-top-style: solid; padding-top: 12px;">
					<span class="text-[16px] font-semibold" style="color: #aeb4bc;">{HARDCODED_COUNTRY_PROFILE}</span>
				</div>
			</div>
			<!-- FUTURE multi-city:
			<div class="mu-card rounded-[20px] p-4" ...>
				<label>МІСЦЕЗНАХОДЖЕННЯ</label>
				<select bind:value={city} ...>
					<option value="">Оберіть місто</option>
					{#each UKRAINE_CITIES as c}<option value={c}>{c}</option>{/each}
				</select>
				<div ...><input type="text" value="Україна" readonly ... /></div>
			</div>
			-->

			<!-- Program -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРОГРАМА ТАНЦІВ</label>
				<div class="flex gap-2">
					{#each PROGRAM_OPTIONS as p}
						<button
							onclick={() => (program = p.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {program === p.value ? '#8984da' : 'transparent'}; color: {program === p.value ? 'white' : '#696969'}; border: 1.5px solid {program === p.value ? '#8984da' : 'rgba(174,180,188,0.4)'};"
						>{p.label}</button>
					{/each}
				</div>
			</div>

		<!-- Goal (dancer/parent only) -->
		{#if isDancerLike}
		<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ЦІЛЬ</label>
			<div class="grid grid-cols-2 gap-2">
				{#each GOAL_OPTIONS as g}
					<button
						onclick={() => (goal = g.value)}
						class="rounded-[50px] py-2.5 text-[13px] font-semibold transition-all"
						style="background: {goal === g.value ? '#8984da' : 'transparent'}; color: {goal === g.value ? 'white' : '#696969'}; border: 1.5px solid {goal === g.value ? '#8984da' : 'rgba(174,180,188,0.4)'};"
					>{g.label}</button>
				{/each}
			</div>
		</div>
		{/if}

			<!-- Categories -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КАТЕГОРІЯ</label>
				<div class="flex flex-wrap gap-2">
					{#each CATEGORIES_UA as cat}
						<button
							onclick={() => toggleCategory(cat.value)}
							class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
							style="background: {categories.includes(cat.value) ? '#8984da' : 'transparent'}; color: {categories.includes(cat.value) ? 'white' : '#696969'}; border: 1.5px solid {categories.includes(cat.value) ? '#8984da' : 'rgba(174,180,188,0.4)'};"
						>{cat.label}</button>
					{/each}
				</div>
			</div>

			<!-- Bio -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 8px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">ПРО СЕБЕ</label>
				<textarea
					placeholder="Розкажи про себе…"
					bind:value={bio}
					rows="4"
					class="mu-text-primary w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
				></textarea>
			</div>

		<!-- Preferences (dancer/parent only: relocate + finance) -->
		{#if isDancerLike}
		<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
			<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">УПОДОБАННЯ</label>
			<div class="flex items-center justify-between">
				<span class="mu-text-primary text-[14px] font-semibold">Готовий/а до переїзду</span>
				<button
					onclick={() => (readyToRelocate = !readyToRelocate)}
					class="relative flex items-center transition-colors"
					style="width: 50px; height: 28px; border-radius: 50px; background: {readyToRelocate ? '#8984da' : 'rgba(174,180,188,0.4)'};"
					role="switch"
					aria-checked={readyToRelocate}
				>
					<div
						class="absolute h-[22px] w-[22px] rounded-full bg-white shadow-sm transition-transform"
						style="transform: translateX({readyToRelocate ? '25px' : '3px'});"
					></div>
				</button>
			</div>
			<div style="display: flex; flex-direction: column; gap: 8px;">
				<span class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">МОЖЛИВІСТЬ ФІНАНСУВАТИ ПАРТНЕРА</span>
				<div class="flex gap-2">
					{#each FINANCE_OPTIONS as opt}
						<button
							onclick={() => (readyToFinance = readyToFinance === opt.value ? '' : opt.value)}
							class="flex-1 rounded-[50px] py-2 text-[13px] font-semibold transition-all"
							style="background: {readyToFinance === opt.value ? '#8984da' : 'transparent'}; color: {readyToFinance === opt.value ? 'white' : '#696969'}; border: 1.5px solid {readyToFinance === opt.value ? '#8984da' : 'rgba(174,180,188,0.4)'};"
						>{opt.label}</button>
					{/each}
				</div>
			</div>
		</div>
		{/if}
		</div>
	{/if}
</div>

<!-- Avatar crop modal -->
{#if cropMode}
	<div class="fixed inset-0 z-50 flex flex-col items-center justify-center" style="background: #000;">
		<p class="mb-5 text-[15px] font-semibold text-white">Перетягни, щоб обрізати</p>
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
				onclick={() => { cropMode = false; avatarPreview = origAvatarPreview; }}
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
