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
				await goto('/feed');
			} else {
				errorMsg = response.error || 'Login failed. Check your credentials.';
			}
		} catch {
			errorMsg = 'Network error. Please try again.';
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
	<h1 class="mb-10 text-[28px] font-black" style="color: #171717;">Sign in</h1>

	<div class="flex w-full max-w-sm flex-col gap-4">
		<input
			type="email"
			placeholder="Email"
			bind:value={email}
			onkeydown={handleKeydown}
			autocomplete="email"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Password"
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
			{isLoading ? 'Signing in…' : 'Sign in'}
		</button>

		<a
			href="/register"
			class="text-center text-[13px] font-medium"
			style="color: #696969;"
		>
			Create account
		</a>
		<a
			href="/forgotPassword"
			class="text-center text-[13px] font-medium"
			style="color: #696969;"
		>
			Forgot password?
		</a>
	</div>
</div>
