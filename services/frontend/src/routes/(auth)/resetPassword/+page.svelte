<script lang="ts">
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { valibot } from 'sveltekit-superforms/adapters';
	import { superForm, defaults } from 'sveltekit-superforms';

	import Button from '$components/ui/button/button.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as v from 'valibot';
	import { authStore } from '$stores/auth.svelte';

	let resetSuccess = $state(false);
	let hasTokenError = $state(false);
	let token = $state('');

	const resetPasswordSchema = v.pipeAsync(
		v.objectAsync({
			password: v.pipe(v.string(), v.minLength(8, 'auth.error.password-length')),
			confirm_password: v.pipe(v.string(), v.minLength(8, 'auth.error.password-length')),
		}),
		v.forwardAsync(
			v.partialCheckAsync(
				[['password'], ['confirm_password']],
				(input) => input.password === input.confirm_password,
				'auth.error.password-mismatch'
			),
			['confirm_password']
		)
	);

	type FormData = v.InferInput<typeof resetPasswordSchema>;

	const initialData: FormData = {
		password: '',
		confirm_password: '',
	};

	const form = superForm(defaults(initialData, valibot(resetPasswordSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(resetPasswordSchema),
		validationMethod: 'oninput',
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
		async onUpdate({ form }) {
			if (form.valid && token) {
				const resetData = {
					token,
					password: form.data.password,
				};

				const resp = await authFetch('/auth/password/reset', {
					method: 'POST',
					body: JSON.stringify(resetData),
					headers: {
						'Content-Type': 'application/json',
					},
				});
				const response: ApiResponse<{ message: string }> = await resp.json();

				if (resp.status === 200 && response.data) {
					resetSuccess = true;
					return;
				}

				toast.error(response.error || $t('auth.toast.reset-password-failed'));
			}
		},
	});

	const { form: formData, enhance } = form;

	onMount(async () => {
		await authStore.logout(false);

		const urlToken = page.url.searchParams.get('token');
		if (!urlToken) {
			hasTokenError = true;
			toast.error($t('auth.error.missing-token'));
			return;
		}
		token = urlToken;
	});
</script>

<div class="flex h-full flex-col items-center justify-center">
	<Card.Root class="m-auto w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-xl">{$t('auth.reset-password.title')}</Card.Title>
			<Card.Description>{$t('auth.reset-password.description')}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if resetSuccess}
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
						<h3 class="text-lg font-semibold">{$t('auth.reset-password.success.title')}</h3>
						<p class="text-sm text-neutral-500">
							{$t('auth.reset-password.success.message')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.reset-password.back-to-login')}
					</Button>
				</div>
			{:else if hasTokenError}
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
						<h3 class="text-lg font-semibold">{$t('auth.reset-password.error.title')}</h3>
						<p class="text-sm text-red-600">{$t('auth.error.missing-token')}</p>
						<p class="text-sm text-neutral-500">
							{$t('auth.reset-password.error.message')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.reset-password.back-to-login')}
					</Button>
				</div>
			{:else}
				<form use:enhance method="POST" class="space-y-6">
					<Form.Field {form} name="password">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.password.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.password}
									type="password"
									placeholder={$t('auth.password.placeholder')}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field {form} name="confirm_password">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.confirm-password.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.confirm_password}
									type="password"
									placeholder={$t('auth.confirm-password.placeholder')}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Button class="w-full">{$t('auth.reset-password.submit')}</Form.Button>
					<Button href="/login" class="w-full" variant="ghost">
						{$t('auth.reset-password.back-to-login')}
					</Button>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

