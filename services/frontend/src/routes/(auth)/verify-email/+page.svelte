<script lang="ts">
	import { page } from '$app/state';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { goto } from '$app/navigation';

	let email = $derived(page.url.searchParams.get('email') ?? '');
	let digits = $state<string[]>(Array(5).fill(''));
	let inputs: HTMLInputElement[] = [];
	let isLoading = $state(false);
	let isResending = $state(false);
	let errorMsg = $state('');
	let successMsg = $state('');

	function handleInput(i: number, e: Event) {
		const val = (e.target as HTMLInputElement).value.replace(/\D/g, '');
		digits[i] = val.slice(-1);
		digits = [...digits];
		if (val && i < 4) inputs[i + 1]?.focus();
	}

	function handleKeydown(i: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !digits[i] && i > 0) {
			inputs[i - 1]?.focus();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		e.preventDefault();
		const pasted = e.clipboardData?.getData('text').replace(/\D/g, '') ?? '';
		if (pasted.length >= 5) {
			digits = pasted.slice(0, 5).split('');
		}
	}

	async function handleVerify() {
		errorMsg = '';
		const code = digits.join('');
		if (code.length < 5) {
			errorMsg = 'Введи повний 5-значний код';
			return;
		}
		isLoading = true;
		try {
			const resp = await fetch(`${import.meta.env.VITE_API_URL}/auth/otp/verify`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, code })
			});
			const response = await resp.json();
			if (resp.ok) {
				// Cookie is now set by the server; update local auth state from returned user
				const user = response.data?.user;
				if (user) {
					authStore.user = user;
					authStore.isAuthenticated = true;
				}
				const firstName = user?.profile_data?.first_name ?? authStore.user?.profile_data?.first_name;
				await goto(firstName ? '/feed' : '/onboarding');
			} else {
				errorMsg = response.error || 'Невірний код підтвердження';
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		} finally {
			isLoading = false;
		}
	}

	async function handleResend() {
		isResending = true;
		errorMsg = '';
		successMsg = '';
		try {
			const resp = await fetch(`${import.meta.env.VITE_API_URL}/auth/otp/send`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email })
			});
			if (resp.ok) {
				successMsg = 'Код надіслано!';
				setTimeout(() => { successMsg = ''; }, 3000);
			} else {
				errorMsg = 'Не вдалося надіслати код. Спробуй ще раз.';
			}
		} catch {
			errorMsg = 'Помилка мережі. Спробуй ще раз.';
		}
		isResending = false;
	}
</script>

<div class="flex min-h-[100dvh] flex-col items-center justify-center px-6 pt-safe pb-safe">
	<img src="/match_icon.svg" alt="MatchUp" class="mb-4 h-16 w-16" />
	<h1 class="mb-2 text-[28px] font-black" style="color: #171717;">Перевір пошту</h1>
	<p class="mb-10 text-center text-[14px] font-medium" style="color: #696969;">
		Ми надіслали код на<br /><span style="color: #171717;">{email}</span>
	</p>

	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="mb-8 flex gap-2" onpaste={handlePaste}>
		{#each digits as digit, i}
			<input
				bind:this={inputs[i]}
				type="text"
				inputmode="numeric"
				maxlength="1"
				value={digit}
				oninput={(e) => handleInput(i, e)}
				onkeydown={(e) => handleKeydown(i, e)}
				class="h-12 w-10 text-center text-[20px] font-bold outline-none transition-colors"
				style="border: 1.5px solid {digit ? '#8984da' : '#171717'}; border-radius: 12px; background: transparent; color: #171717;"
			/>
		{/each}
	</div>

	{#if errorMsg}
		<p class="mb-3 text-center text-[13px] font-medium text-red-500">{errorMsg}</p>
	{/if}
	{#if successMsg}
		<p class="mb-3 text-center text-[13px] font-medium" style="color: #22c55e;">{successMsg}</p>
	{/if}

	<button
		onclick={handleVerify}
		disabled={isLoading || digits.join('').length < 5}
		class="mb-4 w-full max-w-sm py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-60"
		style="border-radius: 50px; background: #696969;"
	>
		{isLoading ? 'Перевірка…' : 'Підтвердити'}
	</button>

	<button
		onclick={handleResend}
		disabled={isResending}
		class="text-[13px] font-medium transition-opacity disabled:opacity-60"
		style="color: #696969;"
	>
		{isResending ? 'Надсилання…' : 'Надіслати код ще раз'}
	</button>
</div>
