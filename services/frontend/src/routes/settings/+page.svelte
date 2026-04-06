<script lang="ts">
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';
	import { authStore } from '$stores/auth.svelte';
	import { browser } from '$app/environment';

	let isDark = $state(false);

	if (browser) {
		isDark = document.documentElement.classList.contains('dark');
	}

	function toggleTheme() {
		isDark = !isDark;
		if (browser) {
			if (isDark) {
				document.documentElement.classList.add('dark');
				localStorage.setItem('mu-theme', 'dark');
			} else {
				document.documentElement.classList.remove('dark');
				localStorage.setItem('mu-theme', 'light');
			}
		}
	}

	let user = $derived(authStore.user);
	let displayName = $derived(
		user?.profile_data?.first_name
			? `${user.profile_data.first_name} ${user.profile_data.last_name ?? ''}`.trim()
			: 'Your Name'
	);
	let avatarUrl = $derived(user?.profile_data?.avatar as string | undefined);
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden" style="background: #dae1eb;">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="px-4 pt-4 pb-2">
		<h1 class="text-[24px] font-black" style="color: #171717;">Settings</h1>
	</div>

	<!-- Scrollable content -->
	<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 16px;">
		<!-- Profile card -->
		<div class="flex items-center gap-4 rounded-[20px] bg-white p-4">
			<div
				class="flex h-[64px] w-[64px] flex-shrink-0 items-center justify-center overflow-hidden rounded-full"
				style="background: #e0e0e0;"
			>
				{#if avatarUrl}
					<img src={avatarUrl} alt={displayName} class="h-full w-full object-cover" />
				{:else}
					<i class="fi fi-rr-user" style="font-size: 28px; color: #696969;"></i>
				{/if}
			</div>
			<div class="flex flex-col gap-0.5">
				<span class="text-[16px] font-bold" style="color: #171717;">{displayName}</span>
				<span class="text-[13px] font-medium" style="color: #696969;">{user?.email ?? ''}</span>
			</div>
		</div>

		<!-- Appearance -->
		<div class="overflow-hidden rounded-[20px] bg-white">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				Appearance
			</p>
			<div class="flex items-center justify-between px-4 pb-4">
				<div class="flex items-center gap-3">
					<i class="fi fi-rr-moon" style="font-size: 18px; color: #171717;"></i>
					<span class="text-[14px] font-semibold" style="color: #171717;">Dark mode</span>
				</div>
				<button
					onclick={toggleTheme}
					class="relative flex items-center transition-colors"
					style="width: 50px; height: 28px; border-radius: 50px; background: {isDark
						? '#8984da'
						: '#d1d5db'};"
					aria-label="Toggle dark mode"
					role="switch"
					aria-checked={isDark}
				>
					<div
						class="absolute h-[22px] w-[22px] rounded-full bg-white shadow-sm transition-transform"
						style="transform: translateX({isDark ? '25px' : '3px'});"
					></div>
				</button>
			</div>
		</div>

		<!-- Account -->
		<div class="overflow-hidden rounded-[20px] bg-white">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				Account
			</p>
			<div class="flex flex-col" style="border-top: 1px solid #f0f0f0;">
				<a href="/settings/profile" class="flex items-center justify-between px-4 py-3">
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-user-pen" style="font-size: 18px; color: #171717;"></i>
						<span class="text-[14px] font-semibold" style="color: #171717;">Edit profile</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</a>
				<a
					href="/forgotPassword"
					class="flex items-center justify-between px-4 py-3"
					style="border-top: 1px solid #f0f0f0;"
				>
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-lock" style="font-size: 18px; color: #171717;"></i>
						<span class="text-[14px] font-semibold" style="color: #171717;">Change password</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</a>
			</div>
		</div>

		<!-- Billing -->
		<div class="overflow-hidden rounded-[20px] bg-white">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				Billing
			</p>
			<div class="flex flex-col" style="border-top: 1px solid #f0f0f0;">
				<div class="flex items-center justify-between px-4 py-3">
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-diamond" style="font-size: 18px; color: #8984da;"></i>
						<span class="text-[14px] font-semibold" style="color: #171717;">Subscription</span>
					</div>
					<span class="text-[13px] font-medium" style="color: #aeb4bc;">Free</span>
				</div>
				<button
					class="flex items-center justify-between px-4 py-3"
					style="border-top: 1px solid #f0f0f0;"
				>
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-credit-card" style="font-size: 18px; color: #171717;"></i>
						<span class="text-[14px] font-semibold" style="color: #171717;">Manage subscription</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</button>
			</div>
		</div>

		<!-- Log out -->
		<div class="overflow-hidden rounded-[20px] bg-white">
			<button
				class="flex w-full items-center gap-3 px-4 py-4"
				onclick={() => authStore.logout()}
			>
				<i class="fi fi-rr-exit" style="font-size: 18px; color: #e74c3c;"></i>
				<span class="text-[14px] font-semibold" style="color: #e74c3c;">Log out</span>
			</button>
		</div>
	</div>

	<BottomNav active="settings" />
</div>
