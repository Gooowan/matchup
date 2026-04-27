<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { onMount } from 'svelte';

	let email = $state('');
	let isLoading = $state(false);
	let errorMsg = $state('');
	let emailSent = $state(false);

	onMount(async () => {
		await authStore.logout(false);
	});

	async function handleSubmit() {
		errorMsg = '';
		isLoading = true;
		try {
			const resp = await authFetch('/auth/password/forgot', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email })
			});
			const response = await resp.json();
			if (resp.ok) {
				emailSent = true;
			} else {
				errorMsg = response.error || 'Не вдалося надіслати листа. Спробуй ще раз.';
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

	{#if emailSent}
		<h1 class="mb-3 text-[24px] font-black" style="color: #171717;">Перевір пошту</h1>
		<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
			Ми надіслали посилання для скидання пароля на<br /><span style="color: #171717;">{email}</span>
		</p>
		<a
			href="/login"
			class="w-full max-w-sm py-3 text-center text-[14px] font-semibold text-white transition-opacity"
			style="border-radius: 50px; background: #696969; display: block;"
		>Повернутися до входу</a>
	{:else}
		<h1 class="mb-3 text-[24px] font-black" style="color: #171717;">Забули пароль?</h1>
		<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
			Введи свій email і ми надішлемо посилання для скидання пароля
		</p>

		<div class="flex w-full max-w-sm flex-col gap-4">
			<input
				type="email"
				placeholder="Ел. пошта"
				bind:value={email}
				autocomplete="email"
				class="w-full px-5 py-3 text-[14px] font-medium outline-none"
				style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
			/>

			{#if errorMsg}
				<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
			{/if}

			<button
				onclick={handleSubmit}
				disabled={isLoading || !email}
				class="mt-2 w-full py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-60"
				style="border-radius: 50px; background: #696969;"
			>
				{isLoading ? 'Надсилання…' : 'Надіслати посилання'}
			</button>

			<a href="/login" class="text-center text-[13px] font-medium" style="color: #696969;">
				Повернутися до входу
			</a>
		</div>
	{/if}
</div>
