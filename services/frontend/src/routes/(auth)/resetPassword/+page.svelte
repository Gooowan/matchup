<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let errorMsg = $state('');
	let resetSuccess = $state(false);
	let token = $state('');
	let tokenMissing = $state(false);

	onMount(async () => {
		await authStore.logout(false);
		const urlToken = page.url.searchParams.get('token');
		if (!urlToken) {
			tokenMissing = true;
		} else {
			token = urlToken;
		}
	});

	async function handleReset() {
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
			const resp = await authFetch('/auth/password/reset', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token, password })
			});
			const response = await resp.json();
			if (resp.ok) {
				resetSuccess = true;
			} else {
				errorMsg = response.error || 'Не вдалося скинути пароль. Спробуй ще раз.';
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

	{#if tokenMissing}
		<h1 class="mb-3 text-[24px] font-black" style="color: #171717;">Недійсне посилання</h1>
		<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
			Посилання для скидання пароля недійсне або застаріло.
		</p>
		<a
			href="/forgotPassword"
			class="w-full max-w-sm py-3 text-center text-[14px] font-semibold text-white"
			style="border-radius: 50px; background: #696969; display: block;"
		>Запросити нове посилання</a>
	{:else if resetSuccess}
		<h1 class="mb-3 text-[24px] font-black" style="color: #171717;">Пароль змінено</h1>
		<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
			Твій пароль успішно змінено. Тепер можеш увійти.
		</p>
		<a
			href="/login"
			class="w-full max-w-sm py-3 text-center text-[14px] font-semibold text-white"
			style="border-radius: 50px; background: #696969; display: block;"
		>Увійти</a>
	{:else}
		<h1 class="mb-3 text-[24px] font-black" style="color: #171717;">Новий пароль</h1>
		<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
			Введи новий пароль для свого акаунту
		</p>

		<div class="flex w-full max-w-sm flex-col gap-4">
			<input
				type="password"
				placeholder="Новий пароль"
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
				onclick={handleReset}
				disabled={isLoading || !password || !confirmPassword}
				class="mt-2 w-full py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-60"
				style="border-radius: 50px; background: #696969;"
			>
				{isLoading ? 'Збереження…' : 'Змінити пароль'}
			</button>

			<a href="/login" class="text-center text-[13px] font-medium" style="color: #696969;">
				Повернутися до входу
			</a>
		</div>
	{/if}
</div>
