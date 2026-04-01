<script lang="ts">
	import { authFetch } from '$lib/utils/authFetch';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let errorMsg = $state('');

	async function handleRegister() {
		errorMsg = '';
		if (password !== confirmPassword) {
			errorMsg = 'Passwords do not match';
			return;
		}
		if (password.length < 8) {
			errorMsg = 'Password must be at least 8 characters';
			return;
		}
		isLoading = true;
		try {
			const resp = await authFetch('/auth/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password, profile_data: {} })
			});
			const response = await resp.json();
			if (resp.ok || resp.status === 201) {
				// Send OTP to email
				await authFetch('/auth/otp/send', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ email })
				});
				await goto(`/verify-email?email=${encodeURIComponent(email)}`);
			} else {
				errorMsg = response.error || 'Registration failed. Try a different email.';
			}
		} catch {
			errorMsg = 'Network error. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-[100dvh] flex-col items-center justify-center px-6 pt-safe pb-safe">
	<img src="/match_icon.svg" alt="MatchUp" class="mb-2 h-16 w-16" />
	<h1 class="mb-10 text-[28px] font-black" style="color: #171717;">Create account</h1>

	<div class="flex w-full max-w-sm flex-col gap-4">
		<input
			type="email"
			placeholder="Email"
			bind:value={email}
			autocomplete="email"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Password"
			bind:value={password}
			autocomplete="new-password"
			class="w-full px-5 py-3 text-[14px] font-medium outline-none"
			style="border: 1.5px solid #171717; border-radius: 50px; background: transparent; color: #171717;"
		/>
		<input
			type="password"
			placeholder="Confirm password"
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
			class="mt-2 w-full py-3 text-[14px] font-semibold text-white transition-opacity disabled:opacity-60"
			style="border-radius: 50px; background: #696969;"
		>
			{isLoading ? 'Creating account…' : 'Create account'}
		</button>

		<a href="/login" class="text-center text-[13px] font-medium" style="color: #696969;">
			Already have an account? Sign in
		</a>
	</div>
</div>
