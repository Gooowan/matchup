<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';

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
					.join(' ') || 'Dancer'
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
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden" style="background: #dae1eb;">
	<div class="pt-safe"></div>

	{#if isLoading}
		<div class="flex flex-1 items-center justify-center">
			<div class="h-10 w-10 animate-spin rounded-full border-4" style="border-color: #d1d5db; border-top-color: #8984da;"></div>
		</div>
	{:else if notFound || !profile}
		<div class="flex flex-1 flex-col items-center justify-center gap-3 px-6">
			<i class="fi fi-rr-user-slash" style="font-size: 48px; color: #aeb4bc;"></i>
			<p class="text-[18px] font-bold" style="color: #171717;">Profile not found</p>
			<button onclick={() => goto(-1 as any)} class="mt-2 text-[14px] font-semibold" style="color: #8984da;">Go back</button>
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

		<!-- Profile details -->
		<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 12px; padding-top: 16px;">
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
						<span class="text-[14px] font-medium" style="color: #171717;">{profile.goal}</span>
					</div>
				{/if}
				{#if profile.program && profile.program !== 'standard'}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-diploma" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{profile.program}</span>
					</div>
				{/if}
				{#if profile.gender}
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-user" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="text-[14px] font-medium" style="color: #171717;">{profile.gender}</span>
					</div>
				{/if}
			</div>

			<!-- Bio -->
			{#if bio}
				<div class="rounded-[20px] bg-white p-4">
					<p class="text-[11px] font-semibold uppercase tracking-wider mb-2" style="color: #aeb4bc;">About</p>
					<p class="text-[14px] font-medium leading-relaxed" style="color: #171717;">{bio}</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
