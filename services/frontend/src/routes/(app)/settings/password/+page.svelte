<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { goto } from '$app/navigation';

	type Step = 'request' | 'verify' | 'new-password' | 'done';

	let step = $state<Step>('request');
	let emailShown = $state('');
	let digits = $state<string[]>(Array(5).fill(''));
	let inputs: HTMLInputElement[] = [];
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let errorMsg = $state('');

	function handleDigitInput(i: number, e: Event) {
		const val = (e.target as HTMLInputElement).value.replace(/\D/g, '');
		digits[i] = val.slice(-1);
		digits = [...digits];
		if (val && i < 4) inputs[i + 1]?.focus();
	}

	function handleDigitKeydown(i: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !digits[i] && i > 0) {
			inputs[i - 1]?.focus();
		}
	}

	function handleDigitPaste(e: ClipboardEvent) {
		e.preventDefault();
		const pasted = e.clipboardData?.getData('text').replace(/\D/g, '') ?? '';
		if (pasted.length >= 5) {
			digits = pasted.slice(0, 5).split('');
		}
	}

	async function handleSendCode() {
		errorMsg = '';
		isLoading = true;
		try {
			const resp = await authFetch('/user/password/change-otp/request', { method: 'POST' });
			const data = await resp.json();
			if (resp.ok) {
				emailShown = data.data?.email ?? authStore.user?.email ?? '';
				step = 'verify';
			} else {
				errorMsg = data.error || 'Не вдалося надіслати код. Спробуй ще раз.';
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isLoading = false;
		}
	}

	// Verified code is stored so it can be used in the final password-change call
	let verifiedCode = $state('');

	async function handleVerifyCodeAndProceed() {
		errorMsg = '';
		const code = digits.join('');
		if (code.length < 5) {
			errorMsg = 'Введи повний 5-значний код';
			return;
		}
		verifiedCode = code;
		step = 'new-password';
	}

	async function handleChangePassword() {
		errorMsg = '';
		if (newPassword !== confirmPassword) {
			errorMsg = 'Паролі не збігаються';
			return;
		}
		if (newPassword.length < 8) {
			errorMsg = 'Пароль має містити щонайменше 8 символів';
			return;
		}
		isLoading = true;
		try {
			const resp = await authFetch('/user/password/change-otp/confirm', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ code: verifiedCode, new_password: newPassword })
			});
			const data = await resp.json();
			if (resp.ok) {
				step = 'done';
			} else {
				errorMsg = data.error || 'Не вдалося змінити пароль. Спробуй ще раз.';
				if (data.error?.includes('код') || data.error?.includes('Код')) {
					step = 'verify';
					digits = Array(5).fill('');
				}
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-[100dvh] flex-col px-6 pt-safe pb-safe">
	<!-- Header -->
	<div class="flex items-center gap-3 py-4">
		<button
			onclick={() => (step === 'request' ? goto('/settings') : step === 'verify' ? (step = 'request') : step === 'new-password' ? (step = 'verify') : goto('/settings'))}
			class="flex h-9 w-9 items-center justify-center"
			style="border-radius: 50%; background: #f4f4f5;"
		>
			<i class="fi fi-rr-angle-left" style="font-size: 16px; line-height: 1; color: #171717;"></i>
		</button>
		<span class="text-[17px] font-bold" style="color: #171717;">Зміна пароля</span>
	</div>

	<div class="flex flex-1 flex-col items-center justify-center gap-6 pb-10">
		{#if step === 'request'}
			<img src="/match_icon.svg" alt="MatchUp" class="h-14 w-14" />
			<div class="text-center">
				<h2 class="mb-2 text-[22px] font-black" style="color: #171717;">Зміна пароля</h2>
				<p class="text-[14px] font-medium" style="color: #696969;">
					Ми надішлемо 5-значний код на твою електронну пошту для підтвердження
				</p>
			</div>

			{#if errorMsg}
				<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
			{/if}

			<button
				onclick={handleSendCode}
				disabled={isLoading}
				class="w-full max-w-sm py-3 text-[14px] font-semibold text-white transition-colors disabled:opacity-60 bg-black active:bg-[#696969]"
				style="border-radius: 50px;"
			>
				{isLoading ? 'Надсилання…' : 'Надіслати код'}
			</button>

		{:else if step === 'verify'}
			<img src="/match_icon.svg" alt="MatchUp" class="h-14 w-14" />
			<div class="text-center">
				<h2 class="mb-2 text-[22px] font-black" style="color: #171717;">Перевір пошту</h2>
				<p class="text-[14px] font-medium" style="color: #696969;">
					Ми надіслали 5-значний код на<br /><span style="color: #171717;">{emailShown}</span>
				</p>
			</div>

			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="flex gap-3" onpaste={handleDigitPaste}>
				{#each digits as digit, i}
					<input
						bind:this={inputs[i]}
						type="text"
						inputmode="numeric"
						maxlength="1"
						value={digit}
						oninput={(e) => handleDigitInput(i, e)}
						onkeydown={(e) => handleDigitKeydown(i, e)}
						class="h-14 w-12 text-center text-[22px] font-bold outline-none transition-colors"
						style="border: 1.5px solid {digit ? '#8984da' : '#d1d5db'}; border-radius: 14px; background: transparent; color: #171717;"
					/>
				{/each}
			</div>

			{#if errorMsg}
				<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
			{/if}

			<button
				onclick={handleVerifyCodeAndProceed}
				disabled={isLoading || digits.join('').length < 5}
				class="w-full max-w-sm py-3 text-[14px] font-semibold text-white transition-colors disabled:opacity-60 bg-black active:bg-[#696969]"
				style="border-radius: 50px;"
			>
				{isLoading ? 'Перевірка…' : 'Підтвердити'}
			</button>

			<button
				onclick={() => { digits = Array(5).fill(''); handleSendCode(); }}
				disabled={isLoading}
				class="text-[13px] font-medium transition-opacity disabled:opacity-60"
				style="color: #696969;"
			>
				Надіслати код ще раз
			</button>

		{:else if step === 'new-password'}
			<img src="/match_icon.svg" alt="MatchUp" class="h-14 w-14" />
			<div class="text-center">
				<h2 class="mb-2 text-[22px] font-black" style="color: #171717;">Новий пароль</h2>
				<p class="text-[14px] font-medium" style="color: #696969;">
					Введи новий пароль для свого акаунту
				</p>
			</div>

			<div class="flex w-full max-w-sm flex-col gap-3">
				<input
					type="password"
					placeholder="Новий пароль"
					bind:value={newPassword}
					autocomplete="new-password"
					class="w-full px-5 py-3 text-[14px] font-medium outline-none"
					style="border: 1.5px solid #d1d5db; border-radius: 50px; background: transparent; color: #171717;"
				/>
				<input
					type="password"
					placeholder="Повторити пароль"
					bind:value={confirmPassword}
					autocomplete="new-password"
					class="w-full px-5 py-3 text-[14px] font-medium outline-none"
					style="border: 1.5px solid #d1d5db; border-radius: 50px; background: transparent; color: #171717;"
				/>

				{#if errorMsg}
					<p class="text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
				{/if}

				<button
					onclick={handleChangePassword}
					disabled={isLoading || !newPassword || !confirmPassword}
					class="mt-1 w-full py-3 text-[14px] font-semibold text-white transition-colors disabled:opacity-60 bg-black active:bg-[#696969]"
					style="border-radius: 50px;"
				>
					{isLoading ? 'Збереження…' : 'Змінити пароль'}
				</button>
			</div>

		{:else if step === 'done'}
			<img src="/match_icon.svg" alt="MatchUp" class="h-14 w-14" />
			<div class="text-center">
				<h2 class="mb-2 text-[22px] font-black" style="color: #171717;">Пароль змінено!</h2>
				<p class="text-[14px] font-medium" style="color: #696969;">
					Твій пароль успішно змінено
				</p>
			</div>
			<button
				onclick={() => goto('/settings')}
				class="w-full max-w-sm py-3 text-[14px] font-semibold text-white bg-black active:bg-[#696969]"
				style="border-radius: 50px;"
			>
				Повернутись до налаштувань
			</button>
		{/if}
	</div>
</div>
