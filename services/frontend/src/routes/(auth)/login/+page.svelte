<script lang="ts">
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { valibot } from 'sveltekit-superforms/adapters';
	import { superForm, defaults } from 'sveltekit-superforms';

	import Button from '$components/ui/button/button.svelte';
	import { authStore } from '$stores/auth.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import * as v from 'valibot';

	let isLoading = $state(false);

	onMount(async () => {
		await authStore.logout(false);
	});

	const loginFormSchema = v.object({
		email: v.pipe(v.string(), v.email('auth.error.email')),
		password: v.pipe(v.string(), v.minLength(8, 'auth.error.password-length')),
	});

	type FormData = v.InferInput<typeof loginFormSchema>;

	const initialData: FormData = {
		email: '',
		password: '',
	};

	const form = superForm(defaults(initialData, valibot(loginFormSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(loginFormSchema),
		validationMethod: 'oninput',
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
		async onUpdate({ form }) {
			if (form.valid) {
				isLoading = true;

				const loginData = {
					email: form.data.email,
					password: form.data.password,
				};

				const resp = await authFetch('/auth/login', {
					method: 'POST',
					body: JSON.stringify(loginData),
					headers: {
						'Content-Type': 'application/json',
					},
				});
				const response: ApiResponse<LoginData> = await resp.json();

				if (resp.status === 200 && response.data) {
					authStore.login(response.data.user);
					await goto('/app');
					isLoading = false;
					return;
				}

				toast.error(response.error || $t('auth.toast.login-failed'));
				isLoading = false;
			}
		},
	});

	const { form: formData, enhance } = form;
</script>

<div class="flex h-full flex-col items-center justify-center">
	<Card.Root class="m-auto w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-xl">{$t('auth.login.title')}</Card.Title>
			<Card.Description>{$t('auth.login.description')}</Card.Description>
		</Card.Header>
		<Card.Content>
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

				<Button class="pl-0" href="/forgotPassword" variant="link">
					{$t('auth.login.forgot')}
				</Button>

				<Form.Button class="w-full" disabled={isLoading}
					>{isLoading ? $t('common.loading') : $t('auth.login.submit')}</Form.Button
				>
				<Button href="/register" class="w-full" variant="ghost">
					{$t('auth.login.register')}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
