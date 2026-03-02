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
	import * as v from 'valibot';
	import { authStore } from '$stores/auth.svelte';

	onMount(async () => {
		await authStore.logout(false);
	});

	let emailSent = $state(false);
	let sentToEmail = $state('');

	const forgotPasswordSchema = v.object({
		email: v.pipe(v.string(), v.email('auth.error.email')),
	});

	type FormData = v.InferInput<typeof forgotPasswordSchema>;

	const initialData: FormData = {
		email: '',
	};

	const form = superForm(defaults(initialData, valibot(forgotPasswordSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(forgotPasswordSchema),
		validationMethod: 'oninput',
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
		async onUpdate({ form }) {
			if (form.valid) {
				const forgotPasswordData = {
					email: form.data.email,
				};

				const resp = await authFetch('/auth/password/forgot', {
					method: 'POST',
					body: JSON.stringify(forgotPasswordData),
					headers: {
						'Content-Type': 'application/json',
					},
				});
				const response: ApiResponse<{ message: string }> = await resp.json();

				if (resp.status === 200 && response.data) {
					sentToEmail = form.data.email;
					emailSent = true;
					return;
				}

				toast.error(response.error || $t('auth.toast.forgot-password-failed'));
			}
		},
	});

	const { form: formData, enhance } = form;
</script>

<div class="flex h-full flex-col items-center justify-center">
	<Card.Root class="m-auto w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-xl">{$t('auth.forgot-password.title')}</Card.Title>
			<Card.Description>{$t('auth.forgot-password.description')}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if emailSent}
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
									d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
								></path>
							</svg>
						</div>
						<h3 class="text-lg font-semibold">{$t('auth.forgot-password.success.title')}</h3>
						<p class="text-sm">
							{$t('auth.forgot-password.success.message')} <strong>{sentToEmail}</strong>
						</p>
						<p class="text-sm text-neutral-500">
							{$t('auth.forgot-password.success.check-email')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.forgot-password.back-to-login')}
					</Button>
				</div>
			{:else}
				<form use:enhance method="POST" class="space-y-6">
					<Form.Field {form} name="email">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.email.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.email}
									type="email"
									placeholder={$t('auth.email.placeholder')}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Button class="w-full">{$t('auth.forgot-password.submit')}</Form.Button>
					<Button href="/login" class="w-full" variant="ghost">
						{$t('auth.forgot-password.back-to-login')}
					</Button>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
