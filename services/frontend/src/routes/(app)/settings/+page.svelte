<script lang="ts">
	import { authStore } from '$stores/auth.svelte';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';

	let isDark = $state(false);
	let ownsClubs = $state(false);
	let subscriptionName = $state<string | null>(null);

	if (browser) {
		isDark = document.documentElement.classList.contains('dark');
	}

	onMount(async () => {
		try {
			const resp = await authFetch('/me/owned-clubs');
			if (resp.ok) {
				const body = await resp.json();
				ownsClubs = Array.isArray(body.data) && body.data.length > 0;
			}
		} catch {
			// non-fatal
		}

		// Load real subscription tier; show "Безкоштовна" only when confirmed absent.
		try {
			const subResp = await authFetch('/subscriptions/my/active');
			if (subResp.ok) {
				const subBody = await subResp.json();
				subscriptionName = subBody.data?.subscription_name ?? null;
			}
		} catch {
			// non-fatal — stays null → shows "Безкоштовна"
		}
	});

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
			: 'Твій профіль'
	);
	let avatarUrl = $derived(user?.profile_data?.avatar as string | undefined);
	let accountType = $derived(user?.profile_data?.account_type as string | undefined);
	// Trainers can join/browse clubs like dancers; only club-type accounts
	// manage entirely via the Business panel and skip this page.
	let showMyClubs = $derived(accountType !== 'club');
	// Show business panel for club/trainer accounts and non-dancer club owners.
	// Dancers never see the business panel even if they happen to own a club.
	let showBusinessPanel = $derived(
		accountType !== 'dancer' && (accountType === 'club' || accountType === 'trainer' || ownsClubs)
	);
	// Club accounts manage identity entirely in the Business panel; no dancer-style profile page.
	let showEditProfile = $derived(accountType !== 'club');
</script>

<div
	class="mu-screen"
	style="height: 100dvh; overflow-y: auto; -webkit-overflow-scrolling: touch;"
>
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="px-4 pt-4 pb-2">
		<h1 class="mu-text-primary text-[24px] font-black">Налаштування</h1>
	</div>

	<!-- Content — scrolls with the whole page wrapper above -->
	<div
		class="flex flex-col px-4"
		style="gap: 16px; padding-bottom: calc(env(safe-area-inset-bottom) + 100px);"
	>
		<!-- Profile card -->
		<div class="mu-card flex items-center gap-4 rounded-[20px] p-4">
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
			<div class="flex min-w-0 flex-col gap-0.5">
				<span class="mu-text-primary truncate text-[16px] font-bold">{displayName}</span>
				<span class="mu-text-secondary truncate text-[13px] font-medium">{user?.email ?? ''}</span>
			</div>
		</div>

		<!-- Appearance -->
		<div class="mu-card overflow-hidden rounded-[20px]">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				ВИГЛЯД
			</p>
			<div class="flex items-center justify-between px-4 pb-4">
				<div class="flex items-center gap-3">
					<i class="fi fi-rr-moon mu-text-primary" style="font-size: 18px;"></i>
					<span class="mu-text-primary text-[14px] font-semibold">Темний режим</span>
				</div>
				<button
					onclick={toggleTheme}
					class="relative flex flex-shrink-0 items-center transition-colors"
					style="width: 51px; height: 31px; border-radius: 50px; background: {isDark
						? '#8984da'
						: '#d1d5db'};"
					aria-label="Увімкнути темний режим"
					role="switch"
					aria-checked={isDark}
				>
					<div
						class="absolute h-[27px] w-[27px] rounded-full bg-white shadow-sm transition-transform"
						style="transform: translateX({isDark ? '22px' : '2px'});"
					></div>
				</button>
			</div>
		</div>

		<!-- Account -->
		<div class="mu-card overflow-hidden rounded-[20px]">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				АКАУНТ
			</p>
			<div class="mu-divider flex flex-col" style="border-top-width: 1px; border-top-style: solid;">
		{#if showEditProfile}
			<a href="/settings/profile" class="flex items-center justify-between px-4 py-3">
				<div class="flex items-center gap-3">
					<i class="fi fi-rr-user-pen mu-text-primary" style="font-size: 18px;"></i>
					<span class="mu-text-primary text-[14px] font-semibold">Редагувати профіль</span>
				</div>
				<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
			</a>
			{#if user?.id}
				<a
					href="/profiles/{user.id}"
					class="mu-divider flex items-center justify-between px-4 py-3"
					style="border-top-width: 1px; border-top-style: solid;"
				>
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-eye mu-text-primary" style="font-size: 18px;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">Переглянути мій профіль</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</a>
			{/if}
		{/if}
				{#if showMyClubs}
					<a
						href="/settings/clubs"
						class="mu-divider flex items-center justify-between px-4 py-3"
						style="border-top-width: 1px; border-top-style: solid;"
					>
						<div class="flex items-center gap-3">
							<i class="fi fi-rr-bank mu-text-primary" style="font-size: 18px;"></i>
							<span class="mu-text-primary text-[14px] font-semibold">Мої клуби</span>
						</div>
						<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
					</a>
				{/if}
				{#if showBusinessPanel}
					<a
						href="/business"
						class="mu-divider flex items-center justify-between px-4 py-3"
						style="border-top-width: 1px; border-top-style: solid;"
					>
						<div class="flex items-center gap-3">
							<i class="fi fi-rr-store-alt mu-text-primary" style="font-size: 18px;"></i>
							<span class="mu-text-primary text-[14px] font-semibold">Бізнес-панель</span>
						</div>
						<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
					</a>
				{/if}
				<a
					href="/settings/password"
					class="mu-divider flex items-center justify-between px-4 py-3"
					style="border-top-width: 1px; border-top-style: solid;"
				>
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-lock mu-text-primary" style="font-size: 18px;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">Змінити пароль</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</a>
			</div>
		</div>

		<!-- Billing -->
		<div class="mu-card overflow-hidden rounded-[20px]">
			<p
				class="px-4 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider"
				style="color: #aeb4bc;"
			>
				ПІДПИСКА
			</p>
			<div class="mu-divider flex flex-col" style="border-top-width: 1px; border-top-style: solid;">
				<div class="flex items-center justify-between px-4 py-3">
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-diamond" style="font-size: 18px; color: #8984da;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">Підписка</span>
					</div>
					<span class="text-[13px] font-medium" style="color: {subscriptionName ? '#8984da' : '#aeb4bc'};">
					{subscriptionName ?? 'Безкоштовна'}
				</span>
				</div>
				<a
					href="/settings/subscription"
					class="mu-divider flex items-center justify-between px-4 py-3"
					style="border-top-width: 1px; border-top-style: solid;"
				>
					<div class="flex items-center gap-3">
						<i class="fi fi-rr-credit-card mu-text-primary" style="font-size: 18px;"></i>
						<span class="mu-text-primary text-[14px] font-semibold">Керувати підпискою</span>
					</div>
					<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
				</a>
			</div>
		</div>

		<!-- Log out -->
		<div class="mu-card overflow-hidden rounded-[20px]">
			<button
				class="flex w-full items-center gap-3 px-4 py-4"
				onclick={() => authStore.logout()}
			>
				<i class="fi fi-rr-exit" style="font-size: 18px; color: #e74c3c;"></i>
				<span class="text-[14px] font-semibold" style="color: #e74c3c;">Вийти</span>
			</button>
		</div>
	</div>
</div>
