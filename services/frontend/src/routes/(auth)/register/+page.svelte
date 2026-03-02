<script lang="ts">
	import { t } from '$lib/locale';
	import toast from 'svelte-french-toast';

	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { valibot } from 'sveltekit-superforms/adapters';
	import { superForm, defaults, setMessage } from 'sveltekit-superforms';

	import Button from '$components/ui/button/button.svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import { authStore } from '$stores/auth.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as v from 'valibot';

	let { data } = $props();

	let inviterInfo: any = $state(null);
	let inviterValidating = $state(false);
	let registrationSuccess = $state(false);
	let registeredEmail = $state('');

	const emailCache = new Map<string, boolean>();
	const inviterCache = new Map<number, { valid: boolean; inviter?: any }>();

	const checkEmailAvailability = async (email: string): Promise<boolean> => {
		const emailValidation = v.safeParse(v.pipe(v.string(), v.email()), email);
		if (!emailValidation.success) {
			return false;
		}

		if (emailCache.has(email)) {
			return emailCache.get(email)!;
		}
		
		const resp = await authFetch('/auth/check/email', {
			method: 'POST',
			body: JSON.stringify({ email }),
			headers: { 'Content-Type': 'application/json' },
		});
		const response: ApiResponse<{ available: boolean }> = await resp.json();

		if (resp.status === 200 && response.data) {
			const isAvailable = response.data.available === true;
			emailCache.set(email, isAvailable);
			return isAvailable;
		}

		emailCache.set(email, false);
		return false;
	};

	const checkInviterValid = async (referralId: number): Promise<boolean> => {
		if (inviterCache.has(referralId)) {
			const cached = inviterCache.get(referralId)!;
			inviterInfo = cached.inviter;
			return cached.valid;
		}
		
		inviterValidating = true;
		
		const resp = await authFetch('/auth/check/inviter', {
			method: 'POST',
			body: JSON.stringify({ referral_id: referralId }),
			headers: { 'Content-Type': 'application/json' },
		});
		const response: ApiResponse<{ inviter: any }> = await resp.json();

		if (resp.status === 200 && response.data) {
			inviterInfo = response.data.inviter;
			const validationResult = { valid: true, inviter: response.data.inviter };
			inviterCache.set(referralId, validationResult);
			inviterValidating = false;
			return true;
		}

		inviterInfo = null;
		const validationResult = { valid: false };
		inviterCache.set(referralId, validationResult);
		inviterValidating = false;
		return false;
	};

	const registerFormSchema = v.pipeAsync(
		v.objectAsync({
			email: v.pipeAsync(
				v.string(), 
				v.email('auth.error.email'),
				v.checkAsync(checkEmailAvailability, 'auth.error.email-taken')
			),
			first_name: v.pipe(v.string(), v.minLength(1, 'auth.error.first-name-required')),
			last_name: v.pipe(v.string(), v.minLength(1, 'auth.error.last-name-required')),
			password: v.pipe(v.string(), v.minLength(8, 'auth.error.password-length')),
			confirm_password: v.pipe(v.string(), v.minLength(8, 'auth.error.password-length')),
			referral_id: v.nullableAsync(
				v.pipeAsync(
					v.number(),
					v.checkAsync(checkInviterValid, 'auth.error.invalid-referral-id')
				)
			)
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

	type FormData = v.InferInput<typeof registerFormSchema>;

	const initialData: FormData = {
		email: '',
		first_name: '',
		last_name: '',
		referral_id: null,
		password: '',
		confirm_password: ''
	};

	const form = superForm(defaults(initialData, valibot(registerFormSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(registerFormSchema),
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
		validationMethod: 'oninput',
		async onUpdate({ form }) {
			if (form.valid) {
				const registrationData: RegistrationRequest = {
					email: form.data.email,
					password: form.data.password,
					referral_id: form.data.referral_id === null ? undefined : form.data.referral_id,
					profile_data: {
						first_name: form.data.first_name,
						last_name: form.data.last_name,
					},
				};

				const resp = await authFetch('/auth/register', {
					method: 'POST',
					body: JSON.stringify(registrationData),
					headers: {
						'Content-Type': 'application/json',
					},
				});
				const response: ApiResponse<RegisterData> = await resp.json();

				if (resp.status === 201 && response.data?.user) {
					registeredEmail = form.data.email;
					registrationSuccess = true;
					emailCache.clear();
					inviterCache.clear();
					setMessage(form, $t('auth.success.registration.title'));
					return;
				}

				toast.error(response.error || $t('auth.toast.registration-failed'));
			}
		},
	});

	const { form: formData, enhance } = form;

	onMount(async () => {
		await authStore.logout(false);

		const referralId = page.url.searchParams.get('rel');
		if (referralId) {
			$formData.referral_id = parseInt(referralId);
		}
	});

</script>

<div class="flex h-full flex-col items-center justify-center">
	<Card.Root class="m-auto w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-xl">{$t('auth.register.title')}</Card.Title>
			<Card.Description>{$t('auth.register.description')}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if registrationSuccess}
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
						<h3 class="text-lg font-semibold">{$t('auth.success.registration.title')}</h3>
						<p class="text-sm">
							{$t('auth.success.registration.email-sent')} <strong>{registeredEmail}</strong>
						</p>
						<p class="text-sm text-neutral-500">
							{$t('auth.success.registration.check-email')}
						</p>
					</div>

					<Button href="/login" class="w-full">
						{$t('auth.register.login')}
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

					<Form.Field {form} name="first_name">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.first-name.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.first_name}
									type="text"
									placeholder={$t('auth.first-name.placeholder')}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field {form} name="last_name">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.last-name.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.last_name}
									type="text"
									placeholder={$t('auth.last-name.placeholder')}
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

					<Form.Field {form} name="referral_id">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>{$t('auth.referral-id.label')}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.referral_id}
									type="number"
									placeholder={$t('auth.referral-id.placeholder')}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />

						{#if inviterValidating}
							<p class="mt-1 text-sm text-gray-500">{$t('auth.validation.validating-inviter')}</p>
						{:else if inviterInfo && inviterInfo.profile_data}
							<p class="mt-1 text-sm text-green-600">
								✓ {$t('auth.validation.inviter-found')}
								{inviterInfo.profile_data.first_name}
								{inviterInfo.profile_data.last_name}
							</p>
						{/if}
					</Form.Field>

					<Form.Button class="w-full">{$t('auth.register.submit')}</Form.Button>
					<Button href="/login" class="w-full" variant="ghost">
						{$t('auth.register.login')}
					</Button>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
