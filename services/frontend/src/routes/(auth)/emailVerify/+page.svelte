<script lang="ts">
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	import * as Card from '$lib/components/ui/card/index.js';
	import Button from '$components/ui/button/button.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { authStore } from '$stores/auth.svelte';

	let isLoading = $state(false);
	let verificationSuccess = $state(false);
	let hasError = $state(false);
	let errorMessage = $state('');

	onMount(async () => {
		await authStore.logout(false);

		const token = page.url.searchParams.get('token');
		if (!token) {
			hasError = true;
			errorMessage = $t('auth.error.missing-token');
			toast.error($t('auth.error.missing-token'));
			return;
		}

		// Auto-submit verification
		isLoading = true;
		await verifyEmail(token);
	});

	async function verifyEmail(token: string) {
		try {
			const resp = await authFetch('/auth/verify/email', {
				method: 'POST',
				body: JSON.stringify({ token }),
				headers: {
					'Content-Type': 'application/json',
				},
			});
			const response: ApiResponse<{ user: any }> = await resp.json();

			if (resp.status === 200 && response.data) {
				verificationSuccess = true;
				isLoading = false;
				return;
			}

			hasError = true;
			errorMessage = response.error || $t('auth.toast.email-verify-failed');
			toast.error(errorMessage);
			isLoading = false;
		} catch (error) {
			hasError = true;
			errorMessage = $t('auth.toast.email-verify-error');
			toast.error(errorMessage);
			isLoading = false;
		}
	}
</script>

<div class="flex h-full flex-col items-center justify-center">
	<Card.Root class="m-auto w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-xl">{$t('auth.email-verify.title')}</Card.Title>
			<Card.Description>{$t('auth.email-verify.description')}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if isLoading}
				<div class="space-y-6 text-center">
					<div class="space-y-2">
						<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full">
							<svg class="h-12 w-12 animate-spin text-blue-600" fill="none" viewBox="0 0 24 24">
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
								></path>
							</svg>
						</div>
						<p class="text-sm text-neutral-500">{$t('auth.email-verify.verifying')}</p>
					</div>
				</div>
			{:else if verificationSuccess}
				<div class="space-y-6 text-center">
					<div class="space-y-2">
						<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full">
							<svg
								class="h-12 w-12 text-green-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M5 13l4 4L19 7"
								></path>
							</svg>
						</div>
						<h3 class="text-lg font-semibold">{$t('auth.email-verify.success.title')}</h3>
						<p class="text-sm text-neutral-500">
							{$t('auth.email-verify.success.message')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.email-verify.back-to-login')}
					</Button>
				</div>
			{:else if hasError}
				<div class="space-y-6 text-center">
					<div class="space-y-2">
						<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full">
							<svg
								class="h-12 w-12 text-red-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M6 18L18 6M6 6l12 12"
								></path>
							</svg>
						</div>
						<h3 class="text-lg font-semibold">{$t('auth.email-verify.error.title')}</h3>
						<p class="text-sm text-red-600">{errorMessage}</p>
						<p class="text-sm text-neutral-500">
							{$t('auth.email-verify.error.message')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.email-verify.back-to-login')}
					</Button>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
