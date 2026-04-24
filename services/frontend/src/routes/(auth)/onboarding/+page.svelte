<script lang="ts">
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import toast from 'svelte-french-toast';

	const DANCE_STYLES = ['Salsa', 'Bachata', 'Ballroom', 'Latin', 'Swing', 'Tango', 'Jazz', 'Contemporary', 'Hip-hop', 'Zouk'];
	const GENDER_OPTIONS = [
		{ value: 'male', label: 'Male' },
		{ value: 'female', label: 'Female' },
		{ value: 'other', label: 'Other' }
	];
	const ROLE_OPTIONS = [
		{ value: 'leader', label: 'Leader' },
		{ value: 'follower', label: 'Follower' },
		{ value: 'both', label: 'Both' }
	];
	const GOAL_OPTIONS = [
		{ value: 'competition', label: 'Competitions' },
		{ value: 'social', label: 'Social dancing' },
		{ value: 'hobby', label: 'Hobby' },
		{ value: 'teaching', label: 'Teaching' },
		{ value: 'professional', label: 'Professional' }
	];
	const LEVEL_OPTIONS = [
		{ value: 'beginner', label: 'Beginner' },
		{ value: 'intermediate', label: 'Intermediate' },
		{ value: 'advanced', label: 'Advanced' },
		{ value: 'professional', label: 'Professional' }
	];

	let step = $state(1);
	const TOTAL_STEPS = 3;

	// Step 1: Basic info
	let firstName = $state('');
	let lastName = $state('');
	let gender = $state('');
	let birthDate = $state('');
	let heightCm = $state<number | null>(null);

	// Step 2: Dance details
	let selectedStyles = $state<string[]>([]);
	let role = $state('');
	let goal = $state('');
	let level = $state('');

	// Step 3: Location + photo
	let city = $state('');
	let country = $state('');
	let avatarFile = $state<File | null>(null);
	let avatarPreview = $state('');
	let fileInput = $state<HTMLInputElement | null>(null);

	let isSaving = $state(false);

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

	function canAdvance(): boolean {
		if (step === 1) return firstName.length >= 2 && lastName.length >= 2 && !!gender;
		if (step === 2) return selectedStyles.length > 0 && !!role;
		return true;
	}

	async function handleFinish() {
		isSaving = true;
		try {
			// Update user name
			await authFetch('/user/profile/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ first_name: firstName, last_name: lastName })
			});

			// Upload avatar if provided
			if (avatarFile) {
				await authStore.uploadAvatar(avatarFile);
			}

			// Create dance profile
			const body: Record<string, unknown> = {
				gender,
				categories: selectedStyles,
				dance_styles: selectedStyles,
				goal: goal || 'hobby',
				program: level || 'standard',
				country,
				city
			};
			if (birthDate) body.birth_date = birthDate;
			if (heightCm) body.height_cm = heightCm;

			await authFetch('/me/profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			await authStore.checkAuth();
			goto('/feed');
		} catch {
			toast.error('Something went wrong. Please try again.');
		} finally {
			isSaving = false;
		}
	}
</script>

<div class="flex min-h-[100dvh] flex-col px-6 pt-safe pb-safe" style="background: #dae1eb;">
	<!-- Progress bar -->
	<div class="pb-8 pt-12">
		<div class="mb-6 flex gap-2">
			{#each Array(TOTAL_STEPS) as _, i}
				<div
					class="h-1 flex-1 rounded-full transition-all"
					style="background: {i < step ? '#8984da' : '#c5cdd8'};"
				></div>
			{/each}
		</div>

		{#if step === 1}
			<h1 class="text-[28px] font-black" style="color: #171717;">Tell us about you</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">Step 1 of 3 — Basic info</p>
		{:else if step === 2}
			<h1 class="text-[28px] font-black" style="color: #171717;">Your dance life</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">Step 2 of 3 — Dance details</p>
		{:else}
			<h1 class="text-[28px] font-black" style="color: #171717;">Almost done!</h1>
			<p class="mt-1 text-[14px] font-medium" style="color: #696969;">Step 3 of 3 — Location & photo</p>
		{/if}
	</div>

	<!-- Step content -->
	<div class="flex-1 overflow-y-auto" style="gap: 16px; display: flex; flex-direction: column;">
		{#if step === 1}
			<!-- First name -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Name</label>
				<input
					type="text"
					placeholder="First name"
					bind:value={firstName}
					autofocus
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717; border-bottom: 1px solid #e0e0e0; padding-bottom: 8px;"
				/>
				<input
					type="text"
					placeholder="Last name"
					bind:value={lastName}
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717;"
				/>
			</div>

			<!-- Gender -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Gender</label>
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
					<span class="text-[14px] font-semibold" style="color: #171717;">Birth date</span>
					<input
						type="date"
						bind:value={birthDate}
						class="bg-transparent text-[14px] font-medium outline-none"
						style="color: #8984da;"
					/>
				</div>
				<div class="flex items-center justify-between" style="border-top: 1px solid #f0f0f0; padding-top: 12px;">
					<span class="text-[14px] font-semibold" style="color: #171717;">Height (cm)</span>
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
			<!-- Dance styles -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Dance styles</label>
				<div class="flex flex-wrap gap-2">
					{#each DANCE_STYLES as s}
						<button
							onclick={() => toggleStyle(s)}
							class="rounded-[50px] px-3 py-1.5 text-[13px] font-semibold transition-all"
							style="background: {selectedStyles.includes(s) ? '#8984da' : 'transparent'}; color: {selectedStyles.includes(s) ? 'white' : '#696969'}; border: 1.5px solid {selectedStyles.includes(s) ? '#8984da' : '#d1d5db'};"
						>{s}</button>
					{/each}
				</div>
			</div>

			<!-- Role -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Role</label>
				<div class="flex gap-2">
					{#each ROLE_OPTIONS as r}
						<button
							onclick={() => (role = r.value)}
							class="flex-1 rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {role === r.value ? '#8984da' : 'transparent'}; color: {role === r.value ? 'white' : '#696969'}; border: 1.5px solid {role === r.value ? '#8984da' : '#d1d5db'};"
						>{r.label}</button>
					{/each}
				</div>
			</div>

			<!-- Level -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Skill level</label>
				<div class="grid grid-cols-2 gap-2">
					{#each LEVEL_OPTIONS as l}
						<button
							onclick={() => (level = l.value)}
							class="rounded-[50px] py-2.5 text-[14px] font-semibold transition-all"
							style="background: {level === l.value ? '#8984da' : 'transparent'}; color: {level === l.value ? 'white' : '#696969'}; border: 1.5px solid {level === l.value ? '#8984da' : '#d1d5db'};"
						>{l.label}</button>
					{/each}
				</div>
			</div>

			<!-- Goal -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Goal</label>
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

		{:else}
			<!-- Photo -->
			<div class="flex flex-col items-center gap-3 rounded-[20px] bg-white p-6">
				<button
					onclick={() => fileInput?.click()}
					class="relative flex h-[100px] w-[100px] items-center justify-center overflow-hidden rounded-full"
					style="background: #e0e0e0;"
				>
					{#if avatarPreview}
						<img src={avatarPreview} alt="Avatar" class="h-full w-full object-cover" />
					{:else}
						<i class="fi fi-rr-user" style="font-size: 40px; color: #696969;"></i>
					{/if}
					<div class="absolute inset-0 flex items-center justify-center rounded-full" style="background: rgba(0,0,0,0.2);">
						<i class="fi fi-rr-camera text-white" style="font-size: 22px; line-height: 1;"></i>
					</div>
				</button>
				<span class="text-[14px] font-medium" style="color: #696969;">Add a profile photo</span>
				<span class="text-[12px] font-medium text-center" style="color: #aeb4bc;">Optional — you can add one later</span>
				<input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleAvatarChange} />
			</div>

			<!-- Location -->
			<div class="rounded-[20px] bg-white p-4" style="display: flex; flex-direction: column; gap: 12px;">
				<label class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Location</label>
				<input
					type="text"
					placeholder="City"
					bind:value={city}
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717; border-bottom: 1px solid #e0e0e0; padding-bottom: 8px;"
				/>
				<input
					type="text"
					placeholder="Country"
					bind:value={country}
					class="w-full bg-transparent text-[16px] font-semibold outline-none"
					style="color: #171717;"
				/>
			</div>
		{/if}
	</div>

	<!-- Navigation buttons -->
	<div class="flex gap-3 pb-6 pt-4">
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
				style="background: #696969;"
			>
				Continue
			</button>
		{:else}
			<button
				onclick={handleFinish}
				disabled={isSaving}
				class="flex h-[50px] flex-1 items-center justify-center rounded-[50px] text-[14px] font-semibold text-white transition-opacity disabled:opacity-50"
				style="background: #8984da;"
			>
				{isSaving ? 'Creating profile…' : 'Start matching 🎉'}
			</button>
		{/if}
	</div>
</div>
