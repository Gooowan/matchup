<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import toast from 'svelte-french-toast';

	const DANCE_STYLES = ['Salsa', 'Bachata', 'Ballroom', 'Latin', 'Swing', 'Tango', 'Jazz', 'Contemporary', 'Hip-hop', 'Zouk'];
	const GENDER_OPTIONS = ['male', 'female', 'other'];
	const ROLE_OPTIONS = ['leader', 'follower', 'both'];
	const GOAL_OPTIONS = ['competition', 'social', 'hobby', 'teaching', 'professional'];
	const PROGRAM_OPTIONS = ['standard', 'latin', 'ballroom', 'show', 'bachata', 'salsa'];

	let isSaving = $state(false);
	let isLoading = $state(true);
	let avatarFile = $state<File | null>(null);
	let avatarPreview = $state<string>('');
	let fileInput = $state<HTMLInputElement | null>(null);

	// User (profile_data) fields
	let firstName = $state('');
	let lastName = $state('');

	// Dance profile fields
	let gender = $state('');
	let birthDate = $state('');
	let heightCm = $state<number | null>(null);
	let country = $state('');
	let city = $state('');
	let selectedStyles = $state<string[]>([]);
	let role = $state('');
	let goal = $state('');
	let program = $state('');
	let bio = $state('');
	let readyToRelocate = $state(false);

	onMount(async () => {
		const user = authStore.user;
		if (user?.profile_data) {
			firstName = (user.profile_data as Record<string, string>).first_name ?? '';
			lastName = (user.profile_data as Record<string, string>).last_name ?? '';
			avatarPreview = (user.profile_data as Record<string, string>).avatar ?? '';
		}
		try {
			const resp = await authFetch('/me/profile');
			if (resp.ok) {
				const body = await resp.json();
				const p = body.data ?? body;
				gender = p.gender ?? '';
				birthDate = p.birth_date ?? '';
				heightCm = p.height_cm ?? null;
				country = p.country ?? '';
				city = p.city ?? '';
				selectedStyles = p.categories ?? p.dance_styles ?? [];
				role = p.program ?? '';
				goal = p.goal ?? '';
				program = p.program ?? '';
				bio = (p.metadata?.bio as string) ?? '';
				readyToRelocate = p.ready_to_relocate ?? false;
			}
		} catch {
			// profile may not exist yet
		} finally {
			isLoading = false;
		}
	});

	function toggleStyle(s: string) {
		selectedStyles = selectedStyles.includes(s)
			? selectedStyles.filter((x) => x !== s)
			: [...selectedStyles, s];
	}

	function handleAvatarChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		avatarFile = file;
		avatarPreview = URL.createObjectURL(file);
	}

	async function handleSave() {
		isSaving = true;
		try {
			// 1. Update name
			await authFetch('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ first_name: firstName, last_name: lastName })
			});

			// 2. Upload avatar if changed
			if (avatarFile) {
				await authStore.uploadAvatar(avatarFile);
			}

			// 3. Update dance profile
			const body: Record<string, unknown> = {
				gender,
				birth_date: birthDate,
				categories: selectedStyles,
				dance_styles: selectedStyles,
				country,
				city,
				goal: goal || 'hobby',
				program: program || 'standard',
				ready_to_relocate: readyToRelocate,
				bio
			};
			if (heightCm) body.height_cm = heightCm;

			await authFetch('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			// Refresh auth store
			await authStore.checkAuth();
			toast.success('Profile saved');
			goto('/settings');
		} catch {
			toast.error('Failed to save profile');
		} finally {
			isSaving = false;
		}
	}
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden mu-screen">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} class="flex items-center justify-center" aria-label="Back">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 text-[20px] font-black">Edit Profile</h1>
		<button
			onclick={handleSave}
			disabled={isSaving}
			class="text-[14px] font-semibold transition-opacity disabled:opacity-60"
			style="color: #8984da;"
		>
			{isSaving ? 'Saving…' : 'Save'}
		</button>
	</div>

	{#if isLoading}
		<div class="flex flex-1 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
		</div>
	{:else}
		<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 16px;">

			<!-- Avatar -->
			<div class="mu-card flex flex-col items-center gap-3 rounded-[20px] p-4">
				<button
					onclick={() => fileInput?.click()}
					class="relative flex h-[80px] w-[80px] items-center justify-center overflow-hidden rounded-full"
					style="background: #e0e0e0;"
				>
					{#if avatarPreview}
						<img src={avatarPreview} alt="Avatar" class="h-full w-full object-cover" />
					{:else}
						<i class="fi fi-rr-user" style="font-size: 32px; color: #696969;"></i>
					{/if}
					<div class="absolute inset-0 flex items-center justify-center rounded-full" style="background: rgba(0,0,0,0.25);">
						<i class="fi fi-rr-camera text-white" style="font-size: 18px; line-height: 1;"></i>
					</div>
				</button>
				<span class="text-[13px] font-medium" style="color: #696969;">Tap to change photo</span>
				<input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleAvatarChange} aria-hidden="true" />
			</div>

			<!-- Name -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Name</p>
				<input
					type="text"
					placeholder="First name"
					bind:value={firstName}
					class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
					style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 8px;"
				/>
				<input
					type="text"
					placeholder="Last name"
					bind:value={lastName}
					class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
					style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 8px;"
				/>
			</div>

			<!-- Basic Info -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Basic Info</p>

				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Gender</span>
					<div class="flex gap-2">
						{#each GENDER_OPTIONS as g}
							<button
								onclick={() => (gender = g)}
								class="rounded-[50px] px-3 py-1 text-[12px] font-semibold transition-colors"
								style="background: {gender === g ? '#8984da' : 'transparent'}; color: {gender === g ? 'white' : '#696969'}; border: 1px solid {gender === g ? '#8984da' : '#d1d5db'};"
							>{g}</button>
						{/each}
					</div>
				</div>

				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Birth date</span>
					<input
						type="date"
						bind:value={birthDate}
						class="bg-transparent text-[14px] font-medium outline-none"
						style="color: #8984da;"
					/>
				</div>

				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Height (cm)</span>
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

			<!-- Location -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Location</p>
				<div style="border-bottom: 1px solid var(--mu-divider, #e0e0e0); padding-bottom: 8px;">
					<input
						type="text"
						placeholder="City"
						bind:value={city}
						class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
					/>
				</div>
				<input
					type="text"
					placeholder="Country"
					bind:value={country}
					class="mu-text-primary w-full bg-transparent text-[14px] font-medium outline-none"
				/>
			</div>

			<!-- Dance Details -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Dance Details</p>

				<span class="mu-text-primary text-[14px] font-semibold">Dance styles</span>
				<div class="flex flex-wrap gap-2">
					{#each DANCE_STYLES as s}
						<button
							onclick={() => toggleStyle(s)}
							class="rounded-[50px] px-3 py-1.5 text-[12px] font-semibold transition-colors"
							style="background: {selectedStyles.includes(s) ? '#8984da' : 'transparent'}; color: {selectedStyles.includes(s) ? 'white' : '#696969'}; border: 1px solid {selectedStyles.includes(s) ? '#8984da' : '#d1d5db'};"
						>{s}</button>
					{/each}
				</div>

				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Role</span>
					<div class="flex gap-2">
						{#each ROLE_OPTIONS as r}
							<button
								onclick={() => (role = r)}
								class="rounded-[50px] px-3 py-1 text-[12px] font-semibold transition-colors"
								style="background: {role === r ? '#8984da' : 'transparent'}; color: {role === r ? 'white' : '#696969'}; border: 1px solid {role === r ? '#8984da' : '#d1d5db'};"
							>{r}</button>
						{/each}
					</div>
				</div>

				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Goal</span>
					<select
						bind:value={goal}
						class="bg-transparent text-[14px] font-medium outline-none"
						style="color: #8984da;"
					>
						<option value="">Select…</option>
						{#each GOAL_OPTIONS as o}
							<option value={o}>{o}</option>
						{/each}
					</select>
				</div>
			</div>

			<!-- Bio -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 8px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">About</p>
				<textarea
					placeholder="Tell others about yourself…"
					bind:value={bio}
					rows="3"
					class="mu-text-primary w-full resize-none bg-transparent text-[14px] font-medium leading-relaxed outline-none"
				></textarea>
			</div>

			<!-- Preferences -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Preferences</p>
				<div class="flex items-center justify-between">
					<span class="mu-text-primary text-[14px] font-semibold">Ready to relocate</span>
					<button
						onclick={() => (readyToRelocate = !readyToRelocate)}
						class="relative flex items-center transition-colors"
						style="width: 50px; height: 28px; border-radius: 50px; background: {readyToRelocate ? '#8984da' : '#d1d5db'};"
						role="switch"
						aria-checked={readyToRelocate}
					>
						<div
							class="absolute h-[22px] w-[22px] rounded-full bg-white shadow-sm transition-transform"
							style="transform: translateX({readyToRelocate ? '25px' : '3px'});"
						></div>
					</button>
				</div>
			</div>

		</div>
	{/if}
</div>
