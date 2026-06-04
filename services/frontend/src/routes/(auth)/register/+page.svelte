<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Capacitor } from '@capacitor/core';
	import type { AccountType } from '$lib/types/accountType';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { isRestrictedAccountType } from '$lib/types/accountType';

	let accountType = $state<AccountType>('dancer');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let isGoogleLoading = $state(false);
	let errorMsg = $state('');

	const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID ?? '';

	onMount(() => {
		if (googleClientId && !Capacitor.isNativePlatform()) {
			const script = document.createElement('script');
			script.src = 'https://accounts.google.com/gsi/client';
			script.async = true;
			document.head.appendChild(script);
		}
	});

	if (typeof window !== 'undefined') {
		(window as any).__matchupGoogleRegCallback = (credentialResponse: any) => {
			if (credentialResponse?.credential) {
				handleGoogleCredential(credentialResponse.credential);
			}
		};
	}

	function navigateAfterLogin(user: any) {
		const pd = user?.profile_data;
		const dest = !pd?.account_type ? '/onboarding' : isRestrictedAccountType(pd?.account_type) ? '/settings' : '/feed';
		goto(dest);
	}

	async function handleGoogleCredential(idToken: string) {
		isGoogleLoading = true;
		errorMsg = '';
		try {
			const resp = await authFetch('/auth/google', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ id_token: idToken })
			});
			const response = await resp.json();
			if (resp.ok && response.data?.user) {
				authStore.login(response.data.user);
				navigateAfterLogin(response.data.user);
			} else {
				errorMsg = parseApiError(response, resp.status);
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isGoogleLoading = false;
		}
	}

	async function handleGoogleNative() {
		if (isGoogleLoading) return;
		isGoogleLoading = true;
		errorMsg = '';
		try {
			const { SocialLogin } = await import('@capgo/capacitor-social-login');
			await SocialLogin.initialize({ google: { webClientId: googleClientId } });
			const result = await SocialLogin.login({ provider: 'google', options: {} });
			const idToken = (result?.result as any)?.idToken;
			if (idToken) {
				await handleGoogleCredential(idToken);
			} else {
				errorMsg = 'Google sign-in failed. Please try again.';
			}
		} catch (e: any) {
			if (e?.message !== 'The user canceled the sign-in flow.') {
				errorMsg = 'Помилка входу через Google. Спробуй ще раз.';
			}
		} finally {
			isGoogleLoading = false;
		}
	}

	const ROLE_TILES: { id: AccountType; label: string; icon: string }[] = [
		{ id: 'dancer', label: 'Танцюрист', icon: 'fi-rr-user' },
		{ id: 'parent', label: 'Батько', icon: 'fi-rr-users-alt' },
		{ id: 'trainer', label: 'Тренер', icon: 'fi-rr-graduation-cap' },
		{ id: 'club', label: 'Клуб', icon: 'fi-rr-bank' }
	];

	async function handleRegister() {
		errorMsg = '';
		if (password !== confirmPassword) {
			errorMsg = 'Паролі не збігаються';
			return;
		}
		if (password.length < 8) {
			errorMsg = 'Пароль має містити щонайменше 8 символів';
			return;
		}
		isLoading = true;
		try {
			const resp = await authFetch('/auth/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					email,
					password,
					profile_data: { account_type: accountType }
				})
			});
			const response = await resp.json();
			if (resp.ok || resp.status === 201) {
				await goto(`/verify-email?email=${encodeURIComponent(email)}`);
			} else {
				errorMsg = parseApiError(response, resp.status);
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-[100dvh] flex-col items-center justify-center px-6 pt-safe pb-safe">
	<img src="/match_icon.svg" alt="MatchUp" class="mb-2 h-16 w-16" />
	<h1 class="mb-6 text-[28px] font-black" style="color: #171717;">Створити акаунт</h1>

	<div class="flex w-full max-w-sm flex-col gap-4">
		<!-- Role picker -->
		<div class="grid grid-cols-2 gap-2.5">
			{#each ROLE_TILES as role}
				{@const selected = accountType === role.id}
				<button
					type="button"
					onclick={() => (accountType = role.id)}
					class="flex flex-col items-center justify-center gap-1.5 py-3 transition-colors"
					style="border-radius: 20px; {selected
						? 'background: #171717; color: #ffffff; border: 1.5px solid #171717;'
						: 'background: transparent; color: #171717; border: 1.5px solid #d1d5db;'}"
					aria-pressed={selected}
				>
					<i class="fi {role.icon}" style="font-size: 22px; line-height: 1;"></i>
					<span class="text-[13px] font-semibold">{role.label}</span>
				</button>
			{/each}
		</div>

		<input
			type="email"
			placeholder="Ел. пошта"
			bind:value={email}
			autocomplete="email"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Пароль"
			bind:value={password}
			autocomplete="new-password"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Підтвердити пароль"
			bind:value={confirmPassword}
			autocomplete="new-password"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>

		{#if errorMsg}
			<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
		{/if}

		<button
			onclick={handleRegister}
			disabled={isLoading || !email || !password || !confirmPassword}
			class="mt-2 w-full py-3 text-[14px] font-semibold text-white transition-colors disabled:opacity-60 bg-black active:bg-[#696969]"
			style="border-radius: 50px;"
		>
			{isLoading ? 'Створення акаунту…' : 'Створити акаунт'}
		</button>

		{#if googleClientId}
			<!-- Divider -->
			<div class="flex items-center gap-3">
				<div class="h-px flex-1" style="background: #e0e0e0;"></div>
				<span class="text-[12px] font-medium" style="color: #969696;">або</span>
				<div class="h-px flex-1" style="background: #e0e0e0;"></div>
			</div>

			{#if Capacitor.isNativePlatform()}
				<button
					onclick={handleGoogleNative}
					disabled={isGoogleLoading}
					class="flex w-full items-center justify-center gap-3 py-3 text-[14px] font-semibold transition-colors disabled:opacity-60"
					style="border: 1.5px solid #e0e0e0; border-radius: 50px; background: white; color: #171717;"
				>
					{#if isGoogleLoading}
						<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300" style="border-top-color: #4285F4;"></div>
					{:else}
						<svg width="18" height="18" viewBox="0 0 18 18" xmlns="http://www.w3.org/2000/svg">
							<path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615Z" fill="#4285F4"/>
							<path d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18Z" fill="#34A853"/>
							<path d="M3.964 10.706A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.706V4.962H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.038l3.007-2.332Z" fill="#FBBC05"/>
							<path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.962L3.964 7.294C4.672 5.163 6.656 3.58 9 3.58Z" fill="#EA4335"/>
						</svg>
					{/if}
					{isGoogleLoading ? 'Вхід…' : 'Продовжити з Google'}
				</button>
			{:else}
				<div
					id="g_id_onload"
					data-client_id={googleClientId}
					data-callback="__matchupGoogleRegCallback"
					data-auto_prompt="false"
				></div>
				<div
					class="g_id_signin"
					data-type="standard"
					data-shape="pill"
					data-theme="outline"
					data-text="signup_with"
					data-size="large"
					data-width="360"
					data-locale="uk"
					style="width: 100%; display: flex; justify-content: center; margin-bottom: -4px;"
				></div>
			{/if}
		{/if}

		<a href="/login" class="text-center text-[13px] font-medium" style="color: #696969;">
			Вже є акаунт? Увійти
		</a>
	</div>
</div>
