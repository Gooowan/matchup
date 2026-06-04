<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
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

	function calcAge(birthDate: string): number {
		const dob = new Date(birthDate);
		const diff = Date.now() - dob.getTime();
		return Math.floor(diff / (365.25 * 24 * 3600 * 1000));
	}

	onMount(async () => {
		try {
			const resp = await authFetch(`/profiles/${userId}/preview`);
			if (resp.ok) {
				const body = await resp.json();
				profile = body.data ?? body;
			} else {
				notFound = true;
			}
		} catch {
			notFound = true;
		} finally {
			isLoading = false;
		}
	});

	let displayName = $derived(
		profile
			? [
					(profile.profile_data?.first_name as string) ?? '',
					(profile.profile_data?.last_name as string) ?? ''
				]
					.filter(Boolean)
					.join(' ') || 'Танцівник'
			: ''
	);
	let avatarUrl = $derived((profile?.profile_data?.avatar as string) ?? '');
	let age = $derived(profile?.birth_date ? calcAge(profile.birth_date) : null);
	let tags = $derived(
		profile
			? [
					...(profile.categories ?? []),
					...(profile.dance_styles ?? [])
				]
					.filter((v, i, a) => a.indexOf(v) === i)
					.slice(0, 5)
			: []
	);
	let bio = $derived((profile?.metadata?.bio as string) ?? '');
	let location = $derived([profile?.city, profile?.country].filter(Boolean).join(', '));
	let mediaUrls = $derived((profile?.metadata?.media_urls as string[]) ?? []);
	let clubName = $derived((profile as any)?.club_name as string | undefined);

	// Lightbox state
	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);

	function openLightbox(index: number) {
		lightboxIndex = index;
		lightboxOpen = true;
	}
	function closeLightbox() { lightboxOpen = false; }
	function lightboxNext() { lightboxIndex = (lightboxIndex + 1) % mediaUrls.length; }
	function lightboxPrev() { lightboxIndex = (lightboxIndex - 1 + mediaUrls.length) % mediaUrls.length; }

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
				if (body.data?.is_mutual_match) {
					chatId = body.data?.chat_id ?? null;
					toast.success('Це метч!');
				} else {
					toast.success('Лайк надіслано!');
				}
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

<div class="mu-screen" style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch; background: #dae1eb;">
	<div class="pt-safe"></div>

	{#if isLoading}
		<div class="flex min-h-[60vh] items-center justify-center">
			<div class="h-10 w-10 animate-spin rounded-full border-4" style="border-color: #d1d5db; border-top-color: #8984da;"></div>
		</div>
	{:else if notFound || !profile}
		<div class="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-6">
			<i class="fi fi-rr-user-slash" style="font-size: 48px; color: #aeb4bc;"></i>
			<p class="text-[18px] font-bold" style="color: #171717;">Профіль не знайдено</p>
			<button onclick={() => goto(-1 as any)} class="mt-2 text-[14px] font-semibold" style="color: #8984da;">Назад</button>
		</div>
	{:else}
		<!-- Photo header -->
		<div class="relative" style="height: 60vh; flex-shrink: 0;">
			{#if avatarUrl}
				<img src={avatarUrl} alt={displayName} class="absolute inset-0 h-full w-full object-cover" />
			{:else}
				<div class="absolute inset-0 flex items-center justify-center" style="background: #c5cdd8;">
					<i class="fi fi-rr-user" style="font-size: 80px; color: #8984da;"></i>
				</div>
			{/if}
			<!-- Gradient overlay -->
			<div class="absolute inset-0" style="background: linear-gradient(180deg, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0) 40%, rgba(0,0,0,0.7) 100%);"></div>

			<!-- Back button -->
			<button
				onclick={() => history.back()}
				class="glass-pill absolute flex h-[38px] w-[38px] items-center justify-center"
				style="top: 16px; left: 16px;"
				aria-label="Back"
			>
				<i class="fi fi-rr-angle-left" style="font-size: 20px; line-height: 1; color: white;"></i>
			</button>

			<!-- Name over photo -->
			<div class="absolute left-4 right-4" style="bottom: 16px;">
				<h1 class="text-[28px] font-black leading-tight text-white" style="letter-spacing: -1px;">
					{displayName}{#if age}, {age}{/if}
				</h1>
				{#if location}
					<div class="mt-1 flex items-center gap-1.5">
						<i class="fi fi-rr-marker text-white" style="font-size: 14px; line-height: 1;"></i>
						<span class="text-[13px] font-medium text-white">{location}</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- Extra photos -->
		{#if mediaUrls.length > 0}
			<div class="flex gap-3 overflow-x-auto px-4 py-3" style="-webkit-overflow-scrolling: touch; flex-shrink: 0;">
				{#each mediaUrls as url, i}
					<button onclick={() => openLightbox(i)} class="flex-shrink-0 rounded-[16px] overflow-hidden focus:outline-none" style="width: 150px; height: 200px;">
						<img src={url} alt="" class="h-full w-full object-cover" />
					</button>
				{/each}
			</div>
		{/if}

		<!-- Profile details -->
		<div class="flex flex-col px-4 pb-[140px]" style="gap: 12px; padding-top: 16px;">
			<!-- Tags -->
			{#if tags.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each tags as tag}
						<span
							class="rounded-[65px] px-3 py-1.5 text-[12px] font-semibold"
							style="background: white; color: #171717; border: 1px solid #d1d5db;"
						>{tag}</span>
					{/each}
					{#if profile.height_cm}
						<span class="rounded-[65px] px-3 py-1.5 text-[12px] font-semibold" style="background: white; color: #171717; border: 1px solid #d1d5db;">
							{profile.height_cm} cm
						</span>
					{/if}
				</div>
			{/if}

			<!-- Info rows -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 10px;">
				{#if profile.goal}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-star" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{GOAL_LABELS[profile.goal] ?? profile.goal}</span>
					</div>
				{/if}
				{#if profile.program}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-diploma" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{PROGRAM_LABELS[profile.program] ?? profile.program}</span>
					</div>
				{/if}
				{#if profile.gender}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-user" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{GENDER_LABELS[profile.gender] ?? profile.gender}</span>
					</div>
				{/if}
				{#if clubName}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-store-alt" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{clubName}</span>
					</div>
				{/if}
			</div>

			<!-- Bio -->
			{#if bio}
				<div class="rounded-[20px] bg-white p-4">
					<p class="text-[11px] font-semibold uppercase tracking-wider mb-2" style="color: #aeb4bc;">ПРО СЕБЕ</p>
					<p class="text-[14px] font-medium leading-relaxed" style="color: #171717;">{bio}</p>
				</div>
			{/if}
		</div>

		<!-- Action bar -->
		{#if profile}
			<div
				class="absolute right-0 bottom-0 left-0 flex items-center gap-3 px-4"
				style="padding-bottom: max(env(safe-area-inset-bottom), 24px); padding-top: 12px; background: #dae1eb; border-top: 1px solid #c5cdd8;"
			>
				<!-- Like -->
				<button
					onclick={handleLike}
					disabled={actionBusy}
					class="flex flex-1 items-center justify-center gap-2 rounded-[50px] py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
					style="background: #8984da;"
				>
					<i class="fi fi-rr-heart" style="font-size: 16px; line-height: 1;"></i>
					Вподобати
				</button>

				<!-- Message -->
				<button
					onclick={handleMessage}
					disabled={actionBusy}
					class="flex flex-1 items-center justify-center gap-2 rounded-[50px] py-3 text-[14px] font-semibold transition-opacity disabled:opacity-50"
					style="background: white; color: #171717; border: 1.5px solid #d1d5db;"
				>
					<i class="fi fi-rr-comment" style="font-size: 16px; line-height: 1;"></i>
					Написати
				</button>

				<!-- Block -->
				<button
					onclick={handleBlock}
					disabled={actionBusy}
					class="flex h-[46px] w-[46px] flex-shrink-0 items-center justify-center rounded-full transition-opacity disabled:opacity-50"
					style="background: white; border: 1.5px solid #d1d5db;"
					aria-label="Block user"
				>
					<i class="fi fi-rr-ban" style="font-size: 18px; color: #e74c3c; line-height: 1;"></i>
				</button>
			</div>
		{/if}
	{/if}
</div>

<!-- Photo lightbox -->
{#if lightboxOpen && mediaUrls.length > 0}
	<div
		class="fixed inset-0 z-[200] flex items-center justify-center"
		style="background: rgba(0,0,0,0.92);"
		onclick={closeLightbox}
		role="dialog"
		aria-label="Photo viewer"
	>
		<img
			src={mediaUrls[lightboxIndex]}
			alt=""
			class="max-h-[85dvh] max-w-[92vw] rounded-[16px] object-contain"
			onclick={(e) => e.stopPropagation()}
		/>
		{#if mediaUrls.length > 1}
			<button
				onclick={(e) => { e.stopPropagation(); lightboxPrev(); }}
				class="absolute left-3 flex h-10 w-10 items-center justify-center rounded-full"
				style="background: rgba(255,255,255,0.15);"
				aria-label="Previous photo"
			>
				<i class="fi fi-rr-angle-left text-white" style="font-size: 18px;"></i>
			</button>
			<button
				onclick={(e) => { e.stopPropagation(); lightboxNext(); }}
				class="absolute right-3 flex h-10 w-10 items-center justify-center rounded-full"
				style="background: rgba(255,255,255,0.15);"
				aria-label="Next photo"
			>
				<i class="fi fi-rr-angle-right text-white" style="font-size: 18px;"></i>
			</button>
			<div class="absolute bottom-6 flex gap-2">
				{#each mediaUrls as _, i}
					<div
						class="h-1.5 rounded-full transition-all"
						style="width: {i === lightboxIndex ? '24px' : '6px'}; background: {i === lightboxIndex ? 'white' : 'rgba(255,255,255,0.4)'};"
					></div>
				{/each}
			</div>
		{/if}
		<button
			onclick={closeLightbox}
			class="absolute top-4 right-4 flex h-10 w-10 items-center justify-center rounded-full"
			style="background: rgba(255,255,255,0.15);"
			aria-label="Close"
		>
			<i class="fi fi-rr-cross text-white" style="font-size: 16px;"></i>
		</button>
	</div>
{/if}
