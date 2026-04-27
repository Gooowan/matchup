<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let email = $state('');
	let password = $state('');
	let isLoading = $state(false);
	let errorMsg = $state('');

	onMount(async () => {
		await authStore.logout(false);
	});

	async function handleLogin() {
		errorMsg = '';
		isLoading = true;
		try {
			const resp = await authFetch('/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password })
			});
			const response = await resp.json();
			if (resp.ok && response.data) {
				authStore.login(response.data.user);
				const firstName = response.data.user?.profile_data?.first_name;
				await goto(firstName ? '/feed' : '/onboarding');
			} else if (resp.status === 403) {
				authFetch('/auth/otp/send', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ email })
				}).catch(() => {});
				await goto('/verify-email?email=' + encodeURIComponent(email));
			} else {
				errorMsg = response.error || 'Невірний email або пароль.';
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isLoading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleLogin();
	}
</script>

<div class="flex min-h-[100dvh] flex-col items-center justify-center px-6 pt-safe pb-safe">
	<!-- Logo -->
	<img src="/match_icon.svg" alt="MatchUp" class="mb-2 h-16 w-16" />
	<h1 class="mb-10 text-[28px] font-black" style="color: #171717;">Увійти</h1>

	<div class="flex w-full max-w-sm flex-col gap-4">
		<input
			type="email"
			placeholder="Ел. пошта"
			bind:value={email}
			onkeydown={handleKeydown}
			autocomplete="email"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Пароль"
			bind:value={password}
			onkeydown={handleKeydown}
			autocomplete="current-password"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>

		{#if errorMsg}
			<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
		{/if}

		<button
			onclick={handleLogin}
			disabled={isLoading || !email || !password}
			class="mt-2 w-full py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-60"
			style="border-radius: 50px; background: #696969;"
		>
			{isLoading ? 'Вхід…' : 'Увійти'}
		</button>

		<a href="/register" class="text-center text-[13px] font-medium" style="color: #696969;">
			Створити акаунт
		</a>
		<a href="/forgotPassword" class="text-center text-[13px] font-medium" style="color: #696969;">
			Забули пароль?
		</a>
	</div>
</div>
