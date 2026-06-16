<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import toast from 'svelte-french-toast';

	let userId = $derived(page.params.userId);

	interface ProfilePreview {
		user_id: string;
		dance_styles: string[];
		gender: string;
		birth_date?: string;
		height_cm?: number;
		goal: string;
		program: string;
		categories: string[];
		country?: string;
		city?: string;
		metadata: Record<string, unknown>;
		profile_data: Record<string, unknown>;
	}

	const GOAL_LABELS: Record<string, string> = { competition: 'Змагання', social: 'Соціальні танці', hobby: 'Хобі', teaching: 'Викладання', professional: 'Профі' };
	const PROGRAM_LABELS: Record<string, string> = { latina: 'Латина', standard: 'Стандарт', both: 'Обидва' };
	const GENDER_LABELS: Record<string, string> = { male: 'Чоловік', female: 'Жінка', other: 'Інше' };

	let profile = $state<ProfilePreview | null>(null);
	let isLoading = $state(true);
	let notFound = $state(false);

	// Photo carousel
	let activePhotoIndex = $state(0);
	let photoDragging = false;
	let photoStartX = 0;

	function calcAge(birthDate: string): number {
		const dob = new Date(birthDate);
		return Math.floor((Date.now() - dob.getTime()) / (365.25 * 24 * 3600 * 1000));
	}

	onMount(async () => {
		if (!userId) {
			notFound = true;
			isLoading = false;
			return;
		}

		const isSelf = authStore.user?.id === userId;
		try {
			if (isSelf) {
				const pd = authStore.user?.profile_data ?? {};
				profile = {
					user_id: userId,
					dance_styles: [],
					gender: (pd.gender as string) ?? '',
					goal: '',
					program: '',
					categories: [],
					metadata: {},
					profile_data: pd
				};
				try {
					const resp = await authFetch('/me/profile');
					if (resp.ok) {
						const body = await resp.json();
						const raw = body.data ?? body;
						profile = {
							user_id: raw.user_id ?? userId,
							dance_styles: raw.dance_styles ?? [],
							gender: raw.gender ?? '',
							birth_date: raw.birth_date,
							height_cm: raw.height_cm,
							goal: raw.goal ?? '',
							program: raw.program ?? '',
							categories: raw.categories ?? [],
							country: raw.country,
							city: raw.city,
							metadata: raw.metadata ?? {},
							profile_data: pd
						};
					}
				} catch { /* keep seeded data */ }
			} else {
				const resp = await authFetch(`/profiles/${userId}/preview`);
				if (resp.ok) {
					const body = await resp.json();
					profile = body.data ?? body;
				} else {
					notFound = true;
				}
			}
		} catch {
			if (!isSelf) notFound = true;
		} finally {
			isLoading = false;
		}
	});

	let isSelf = $derived(authStore.user?.id === userId);
	let displayName = $derived(
		profile
			? [(profile.profile_data?.first_name as string) ?? '', (profile.profile_data?.last_name as string) ?? '']
					.filter(Boolean).join(' ') || 'Танцівник'
			: ''
	);
	let avatarUrl = $derived((profile?.profile_data?.avatar as string) ?? '');
	let age = $derived(profile?.birth_date ? calcAge(profile.birth_date) : null);
	let tags = $derived(
		profile
			? [...(profile.categories ?? []), ...(profile.dance_styles ?? [])]
					.filter((v, i, a) => a.indexOf(v) === i).slice(0, 6)
			: []
	);
	let bio = $derived((profile?.metadata?.bio as string) ?? '');
	let location = $derived([profile?.city, profile?.country].filter(Boolean).join(', '));
	let mediaUrls = $derived((profile?.metadata?.media_urls as string[]) ?? []);
	let clubName = $derived((profile as any)?.club_name as string | undefined);

	// All photos: avatar + extra
	let allPhotos = $derived(
		avatarUrl
			? [avatarUrl, ...mediaUrls.filter(u => u !== avatarUrl)]
			: mediaUrls
	);
	let currentPhoto = $derived(allPhotos[activePhotoIndex] ?? '');

	function prevPhoto(e: MouseEvent) {
		e.stopPropagation();
		if (photoDragging) return;
		activePhotoIndex = Math.max(0, activePhotoIndex - 1);
	}
	function nextPhoto(e: MouseEvent) {
		e.stopPropagation();
		if (photoDragging) return;
		activePhotoIndex = Math.min(allPhotos.length - 1, activePhotoIndex + 1);
	}

	let chatId = $state<string | null>(null);
	let actionBusy = $state(false);

	async function handleLike() {
		if (actionBusy) return;
		actionBusy = true;
		try {
			const resp = await authFetch('/matchup/swipe', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ target_user_id: userId, action: 'LIKE', source: 'profile' })
			});
			const body = await resp.json();
			if (resp.ok) {
				chatId = body.data?.chat_id ?? null;
				toast.success(body.data?.is_mutual_match ? 'Це метч! 🎉' : 'Лайк надіслано!');
			}
		} catch {
			toast.error('Щось пішло не так');
		} finally {
			actionBusy = false;
		}
	}

	async function handleMessage() {
		if (chatId) { goto(`/chats/${chatId}`); return; }
		if (actionBusy) return;
		actionBusy = true;
		try {
			const resp = await authFetch('/chats', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ user_id: userId })
			});
			const body = await resp.json();
			if (resp.ok && body.data?.chat_id) {
				goto(`/chats/${body.data.chat_id}`);
			} else if (resp.status === 403) {
				toast.error('Спочатку познайомтесь з цим танцівником');
			} else {
				toast.error(body.error ?? 'Щось пішло не так');
			}
		} catch {
			toast.error('Щось пішло не так');
		} finally {
			actionBusy = false;
		}
	}

	async function handleBlock() {
		if (actionBusy) return;
		actionBusy = true;
		try {
			await authFetch(`/users/${userId}/block`, { method: 'POST' });
			toast.success('Користувача заблоковано');
			history.back();
		} catch {
			toast.error('Не вдалося заблокувати');
		} finally {
			actionBusy = false;
		}
	}
</script>

<div class="mu-screen relative" style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch; background: #0e0e10;">
	{#if isLoading}
		<div class="flex h-[100dvh] items-center justify-center">
			<div class="h-10 w-10 animate-spin rounded-full border-4" style="border-color: rgba(255,255,255,0.15); border-top-color: #8984da;"></div>
		</div>

	{:else if notFound || !profile}
		<div class="flex h-[100dvh] flex-col items-center justify-center gap-4 px-6">
			<div class="flex h-[80px] w-[80px] items-center justify-center rounded-full" style="background: rgba(255,255,255,0.07);">
				<i class="fi fi-rr-user-slash" style="font-size: 36px; color: #aeb4bc;"></i>
			</div>
			<p class="text-[20px] font-bold text-white">Профіль не знайдено</p>
			<button onclick={() => history.back()} class="mt-1 rounded-[50px] px-6 py-2.5 text-[14px] font-semibold text-white" style="background: rgba(137,132,218,0.25); color: #8984da;">
				← Назад
			</button>
		</div>

	{:else}
		<!-- ═══ PHOTO HERO (full-bleed, 65dvh) ═══ -->
		<div class="relative w-full" style="height: 65dvh; flex-shrink: 0;">
			<!-- Photo -->
			{#if currentPhoto}
				<img
					src={currentPhoto}
					alt={displayName}
					class="absolute inset-0 h-full w-full object-cover"
					style="object-position: center top;"
				/>
			{:else}
				<div class="absolute inset-0 flex items-center justify-center" style="background: #1c1c22;">
					<i class="fi fi-rr-user" style="font-size: 80px; color: #3a3a4a;"></i>
				</div>
			{/if}

			<!-- Cinematic vignette -->
			<div class="absolute inset-0 pointer-events-none" style="background: linear-gradient(180deg, rgba(0,0,0,0.45) 0%, transparent 30%, transparent 55%, rgba(0,0,0,0.85) 100%);"></div>

			<!-- Story-style progress bars -->
			{#if allPhotos.length > 1}
				<div class="absolute left-3 right-3 flex gap-1" style="top: max(env(safe-area-inset-top), 12px); margin-top: 4px;">
					{#each allPhotos as _, i}
						<div class="flex-1 overflow-hidden rounded-full" style="height: 3px; background: rgba(255,255,255,0.28);">
							<div class="h-full rounded-full transition-all duration-200" style="width: {i < activePhotoIndex ? '100%' : i === activePhotoIndex ? '100%' : '0%'}; background: {i <= activePhotoIndex ? 'white' : 'transparent'};"></div>
						</div>
					{/each}
				</div>
			{/if}

			<!-- Floating nav row: back ← ……… edit (self only) -->
			<div
				class="absolute left-3 right-3 flex items-center justify-between"
				style="top: max(env(safe-area-inset-top), 12px); margin-top: {allPhotos.length > 1 ? '18px' : '0'};"
			>
				<button
					onclick={() => history.back()}
					class="flex h-[38px] w-[38px] items-center justify-center rounded-full"
					style="background: rgba(0,0,0,0.4); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);"
					aria-label="Назад"
				>
					<i class="fi fi-rr-angle-left text-white" style="font-size: 18px; line-height: 1;"></i>
				</button>

				{#if isSelf}
					<a
						href="/settings/profile"
						class="flex h-[34px] items-center gap-1.5 rounded-full px-4 text-[13px] font-semibold text-white"
						style="background: rgba(137,132,218,0.75); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);"
					>
						<i class="fi fi-rr-user-pen" style="font-size: 13px; line-height: 1;"></i>
						Редагувати
					</a>
				{/if}
			</div>

			<!-- Tap zones: left = prev, right = next (start below top controls) -->
			{#if allPhotos.length > 1}
				<button
					class="absolute left-0"
					style="width: 40%; top: 80px; height: calc(100% - 200px); background: transparent;"
					aria-label="Попереднє фото"
					onclick={prevPhoto}
				></button>
				<button
					class="absolute right-0"
					style="width: 60%; top: 80px; height: calc(100% - 200px); background: transparent;"
					aria-label="Наступне фото"
					onclick={nextPhoto}
				></button>
			{/if}

			<!-- Name / age / location over photo -->
			<div class="absolute left-4 right-4 bottom-5 pointer-events-none">
				<h1 class="text-[30px] font-black leading-tight text-white" style="letter-spacing: -0.5px; text-shadow: 0 2px 12px rgba(0,0,0,0.5);">
					{displayName}{#if age}<span style="font-weight: 400;">, {age}</span>{/if}
				</h1>
				{#if location}
					<div class="mt-1 flex items-center gap-1.5">
						<i class="fi fi-rr-marker" style="font-size: 13px; color: rgba(255,255,255,0.8); line-height: 1;"></i>
						<span class="text-[13px] font-medium" style="color: rgba(255,255,255,0.85);">{location}</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- ═══ DETAILS SECTION ═══ -->
		<div class="flex flex-col px-4" style="gap: 12px; padding-top: 18px; padding-bottom: calc(max(env(safe-area-inset-bottom), 24px) + {isSelf ? '0px' : '80px'});">

			<!-- Tag pills -->
			{#if tags.length > 0 || profile.height_cm}
				<div class="flex flex-wrap gap-2">
					{#each tags as tag}
						<span class="rounded-[50px] px-3 py-1.5 text-[12px] font-semibold" style="background: rgba(137,132,218,0.18); color: #b8b4ee; border: 1px solid rgba(137,132,218,0.3);">{tag}</span>
					{/each}
					{#if profile.height_cm}
						<span class="rounded-[50px] px-3 py-1.5 text-[12px] font-semibold" style="background: rgba(255,255,255,0.07); color: rgba(255,255,255,0.7); border: 1px solid rgba(255,255,255,0.12);">{profile.height_cm} cm</span>
					{/if}
				</div>
			{/if}

			<!-- Info card -->
			{#if profile.goal || profile.program || profile.gender || clubName}
				<div class="rounded-[20px] overflow-hidden" style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.09);">
					{#if profile.goal}
						<div class="flex items-center gap-3 px-4 py-3.5" style="border-bottom: 1px solid rgba(255,255,255,0.07);">
							<div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.2);">
								<i class="fi fi-rr-star" style="font-size: 14px; color: #8984da;"></i>
							</div>
							<div>
								<p class="text-[11px] font-medium" style="color: rgba(255,255,255,0.45);">Ціль</p>
								<p class="text-[14px] font-semibold text-white">{GOAL_LABELS[profile.goal] ?? profile.goal}</p>
							</div>
						</div>
					{/if}
					{#if profile.program}
						<div class="flex items-center gap-3 px-4 py-3.5" style="border-bottom: 1px solid rgba(255,255,255,0.07);">
							<div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.2);">
								<i class="fi fi-rr-diploma" style="font-size: 14px; color: #8984da;"></i>
							</div>
							<div>
								<p class="text-[11px] font-medium" style="color: rgba(255,255,255,0.45);">Програма</p>
								<p class="text-[14px] font-semibold text-white">{PROGRAM_LABELS[profile.program] ?? profile.program}</p>
							</div>
						</div>
					{/if}
					{#if profile.gender}
						<div class="flex items-center gap-3 px-4 py-3.5" style="{clubName ? 'border-bottom: 1px solid rgba(255,255,255,0.07);' : ''}">
							<div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.2);">
								<i class="fi fi-rr-user" style="font-size: 14px; color: #8984da;"></i>
							</div>
							<div>
								<p class="text-[11px] font-medium" style="color: rgba(255,255,255,0.45);">Стать</p>
								<p class="text-[14px] font-semibold text-white">{GENDER_LABELS[profile.gender] ?? profile.gender}</p>
							</div>
						</div>
					{/if}
					{#if clubName}
						<div class="flex items-center gap-3 px-4 py-3.5">
							<div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full" style="background: rgba(137,132,218,0.2);">
								<i class="fi fi-rr-store-alt" style="font-size: 14px; color: #8984da;"></i>
							</div>
							<div>
								<p class="text-[11px] font-medium" style="color: rgba(255,255,255,0.45);">Клуб</p>
								<p class="text-[14px] font-semibold text-white">{clubName}</p>
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Bio -->
			{#if bio}
				<div class="rounded-[20px] px-4 py-4" style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.09);">
					<p class="mb-2 text-[11px] font-semibold uppercase tracking-wider" style="color: rgba(255,255,255,0.35);">Про себе</p>
					<p class="text-[14px] font-medium leading-relaxed" style="color: rgba(255,255,255,0.85);">{bio}</p>
				</div>
			{/if}
		</div>

		<!-- ═══ ACTION BAR ═══ -->
		{#if !isSelf}
			<div
				class="fixed right-0 bottom-0 left-0 flex items-center gap-3 px-4"
				style="padding-bottom: max(env(safe-area-inset-bottom), 20px); padding-top: 12px; background: rgba(14,14,16,0.85); backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px); border-top: 1px solid rgba(255,255,255,0.08);"
			>
				<!-- Pass / Block -->
				<button
					onclick={handleBlock}
					disabled={actionBusy}
					class="flex h-[50px] w-[50px] flex-shrink-0 items-center justify-center rounded-full transition-opacity disabled:opacity-50"
					style="background: rgba(231,76,60,0.15); border: 1.5px solid rgba(231,76,60,0.3);"
					aria-label="Заблокувати"
				>
					<i class="fi fi-rr-ban" style="font-size: 18px; color: #e74c3c; line-height: 1;"></i>
				</button>

				<!-- Like -->
				<button
					onclick={handleLike}
					disabled={actionBusy}
					class="flex flex-1 items-center justify-center gap-2 rounded-[50px] text-[15px] font-bold text-white transition-opacity disabled:opacity-50"
					style="height: 50px; background: linear-gradient(135deg, #8984da, #6c67c4);"
				>
					<i class="fi fi-rr-heart" style="font-size: 16px; line-height: 1;"></i>
					Вподобати
				</button>

				<!-- Message -->
				<button
					onclick={handleMessage}
					disabled={actionBusy}
					class="flex flex-1 items-center justify-center gap-2 rounded-[50px] text-[15px] font-bold transition-opacity disabled:opacity-50"
					style="height: 50px; background: rgba(255,255,255,0.1); border: 1.5px solid rgba(255,255,255,0.18); color: white;"
				>
					<i class="fi fi-rr-comment" style="font-size: 16px; line-height: 1;"></i>
					Написати
				</button>
			</div>
		{/if}
	{/if}
</div>
