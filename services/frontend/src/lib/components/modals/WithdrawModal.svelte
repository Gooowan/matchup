<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Form from '$lib/components/ui/form';
	import * as Select from '$lib/components/ui/select';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';
	import { valibot } from 'sveltekit-superforms/adapters';
	import { superForm, defaults } from 'sveltekit-superforms';
	import * as v from 'valibot';

	interface Chain {
		chain_type: string;
		tokens: Token[];
	}

	interface Token {
		token_type: string;
		min_amount: number;
		max_amount: number;
	}

	interface PaymentMethods {
		[key: string]: Chain[];
	}

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		product: string;
		onSuccess?: () => void;
		paymentSystem?: string;
		availableBalance?: number;
	}

	let {
		isOpen = $bindable(),
		onClose,
		product,
		onSuccess,
		paymentSystem = 'bscpay',
		availableBalance = 0,
	}: Props = $props();

	let isLoading = $state(false);
	let isLoadingOTP = $state(false);
	let paymentMethods = $state<PaymentMethods>({});
	let selectedPaymentSystem = $state<string>('');
	let selectedChain = $state<string>('');
	let selectedToken = $state<Token | null>(null);
	let isManualTokenSelection = $state(false);
	let stage = $state<'details' | 'otp' | 'success'>('details');
	let otpCode = $state('');
	let withdrawDetails = $state<{
		amount: number;
		chain: string;
		token: string;
		address: string;
	} | null>(null);

	let availablePaymentSystems = $derived(
		Object.keys(paymentMethods).map((key) => ({ value: key, label: key }))
	);

	let availableChains = $derived(() => {
		if (!selectedPaymentSystem || !paymentMethods[selectedPaymentSystem]) return [];
		return paymentMethods[selectedPaymentSystem].map((chain) => ({
			value: chain.chain_type,
			label: chain.chain_type,
		}));
	});

	let availableTokens = $derived(() => {
		if (!selectedPaymentSystem || !selectedChain || !paymentMethods[selectedPaymentSystem])
			return [];
		const chain = paymentMethods[selectedPaymentSystem].find((c) => c.chain_type === selectedChain);
		return chain ? chain.tokens.map((token) => ({ value: token, label: token.token_type })) : [];
	});

	// Derived values to determine if selections should be hidden
	let hidePaymentSystemSelection = $derived(!!paymentSystem);
	let hideChainSelection = $derived(false);
	let hideTokenSelection = $derived(false);

	// Address validation function
	function validateAddress(address: string, chain: string): string | null {
		if (!address) return null;

		address = address.trim();

		// BSC/ETH address validation (0x followed by 40 hex characters)
		if (chain === 'BSC' || chain === 'ETH' || chain === 'ETHEREUM') {
			const ethAddressRegex = /^0x[a-fA-F0-9]{40}$/;
			if (!ethAddressRegex.test(address)) {
				return 'Invalid BSC/ETH address. Must be 0x followed by 40 hexadecimal characters.';
			}
		}
		// Add other chain validators here as needed
		// Solana: base58 encoded, typically 32-44 characters
		// if (chain === 'SOL' || chain === 'SOLANA') { ... }

		return null;
	}

	const withdrawSchema = v.object({
		amount: v.pipe(
			v.number('Amount must be a number'),
			v.minValue(5, 'Minimum withdrawal amount is 10'),
			v.maxValue(25000, 'Maximum withdrawal amount is 25000')
		),
		paymentSystem: v.pipe(v.string(), v.minLength(1, 'Please select a payment system')),
		chain: v.pipe(v.string(), v.minLength(1, 'Please select a chain')),
		token: v.pipe(v.string(), v.minLength(1, 'Please select a token')),
		address: v.pipe(
			v.string(),
			v.minLength(1, 'Please enter a withdrawal address'),
			v.custom((address) => {
				if (!selectedChain) return true;
				return validateAddress(address, selectedChain) === null;
			}, 'Invalid address format for selected chain')
		),
	});

	type FormData = v.InferInput<typeof withdrawSchema>;

	const initialData: FormData = {
		amount: 10,
		paymentSystem: paymentSystem || '',
		chain: '',
		token: '',
		address: '',
	};

	const form = superForm(defaults(initialData, valibot(withdrawSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(withdrawSchema),
		validationMethod: 'oninput',
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
	});

	const { form: formData, enhance } = form;

	let addressError = $state<string | null>(null);

	// Validate address when it changes
	$effect(() => {
		if ($formData.address && selectedChain) {
			addressError = validateAddress($formData.address, selectedChain);
		} else {
			addressError = null;
		}
	});

	// Initialize selected values from props when modal opens
	$effect(() => {
		if (isOpen) {
			if (Object.keys(paymentMethods).length === 0) {
				fetchPaymentMethods();
			}

			// Reset form data with current prop values
			$formData.paymentSystem = paymentSystem || '';
			$formData.chain = '';
			$formData.token = '';
			$formData.amount = 10;
			$formData.address = '';
			stage = 'details';
			otpCode = '';
			withdrawDetails = null;
			addressError = null;

			if (paymentSystem) {
				selectedPaymentSystem = paymentSystem;
			} else {
				selectedPaymentSystem = '';
			}

			selectedChain = '';
			selectedToken = null;
		}
	});

	async function fetchPaymentMethods() {
		const response = await authFetch(`/payments/${product}/methods`);
		if (!response.ok) {
			toast.error('Failed to fetch payment methods');
			return;
		}

		const data: ApiResponse<PaymentMethods> = await response.json();
		if (!data.data) {
			toast.error(data.error || 'Failed to fetch payment methods');
			return;
		}

		paymentMethods = data.data;

		// auto select
		const firstPaymentSystem = Object.keys(paymentMethods)[0];
		if (!firstPaymentSystem) return;

		if (!paymentSystem) {
			selectedPaymentSystem = firstPaymentSystem;
			$formData.paymentSystem = firstPaymentSystem;
		} else {
			selectedPaymentSystem = paymentSystem;
		}

		const chains = paymentMethods[selectedPaymentSystem];
		if (!chains || chains.length === 0) return;

		const firstChain = chains[0];
		selectedChain = firstChain.chain_type;
		$formData.chain = firstChain.chain_type;

		const firstToken = firstChain.tokens[0];
		if (!firstToken) return;

		selectedToken = firstToken;
		$formData.token = firstToken.token_type;
	}

	// Reset selections when payment system changes
	$effect(() => {
		if (selectedPaymentSystem && paymentMethods[selectedPaymentSystem]) {
			const firstChain = paymentMethods[selectedPaymentSystem][0];
			if (firstChain && selectedChain !== firstChain.chain_type) {
				selectedChain = firstChain.chain_type;
				$formData.chain = firstChain.chain_type;

				const firstToken = firstChain.tokens[0];
				if (firstToken) {
					selectedToken = firstToken;
					$formData.token = firstToken.token_type;
					isManualTokenSelection = false;
				}
			}
		}
	});

	// Reset token when chain changes (only if not manually selected)
	$effect(() => {
		if (
			selectedPaymentSystem &&
			selectedChain &&
			paymentMethods[selectedPaymentSystem] &&
			!isManualTokenSelection
		) {
			const chain = paymentMethods[selectedPaymentSystem].find(
				(c) => c.chain_type === selectedChain
			);
			if (chain && chain.tokens.length > 0) {
				const firstToken = chain.tokens[0];
				if (selectedToken?.token_type !== firstToken.token_type) {
					selectedToken = firstToken;
					$formData.token = firstToken.token_type;
				}
			}
		}
	});

	// Validate amount when selected token changes
	$effect(() => {
		if (selectedToken && $formData.amount) {
			if ($formData.amount < selectedToken.min_amount) {
				$formData.amount = selectedToken.min_amount;
			} else if ($formData.amount > selectedToken.max_amount) {
				$formData.amount = selectedToken.max_amount;
			}
		}
	});

	let isValidForm = $derived(() => {
		return (
			$formData.amount >= 10 &&
			$formData.amount <= 25000 &&
			$formData.amount <= availableBalance &&
			$formData.chain &&
			$formData.token &&
			$formData.address &&
			!addressError
		);
	});

	async function handleNext() {
		if (!isValidForm()) {
			toast.error('Please fill in all fields correctly');
			return;
		}

		// Validate amount against available balance
		if ($formData.amount > availableBalance) {
			toast.error('Amount exceeds available balance');
			return;
		}

		isLoadingOTP = true;

		try {
			const response = await authFetch(`/payments/${product}/withdraw/code`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					amount: $formData.amount,
					chain: $formData.chain,
					token: $formData.token,
					address: $formData.address,
				}),
			});

			const result: ApiResponse<string> = await response.json();

			if (response.ok && result.data) {
				withdrawDetails = {
					amount: $formData.amount,
					chain: $formData.chain,
					token: $formData.token,
					address: $formData.address,
				};
				stage = 'otp';
				toast.success('OTP code sent to your email');
			} else {
				toast.error(result.error || 'Failed to send OTP code');
			}
		} catch (error) {
			toast.error('Failed to send OTP code');
			console.error('OTP send error:', error);
		} finally {
			isLoadingOTP = false;
		}
	}

	async function handleConfirm() {
		if (!otpCode || otpCode.length !== 8) {
			toast.error('Please enter a valid 8-digit OTP code');
			return;
		}

		if (!withdrawDetails) {
			toast.error('Withdrawal details missing');
			return;
		}

		isLoading = true;

		try {
			const response = await authFetch(`/payments/${product}/withdraw`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					amount: withdrawDetails.amount,
					chain: withdrawDetails.chain,
					token: withdrawDetails.token,
					address: withdrawDetails.address,
					otp_code: otpCode,
				}),
			});

			const result: ApiResponse<string> = await response.json();

			if (response.ok && result.data) {
				stage = 'success';
				onSuccess?.();
			} else {
				toast.error(result.error || 'Failed to create withdrawal request');
			}
		} catch (error) {
			toast.error('Failed to create withdrawal request');
			console.error('Withdraw error:', error);
		} finally {
			isLoading = false;
		}
	}

	function handleClose() {
		if (!isLoading && !isLoadingOTP) {
			stage = 'details';
			otpCode = '';
			withdrawDetails = null;
			addressError = null;
			onClose();
		}
	}

	function handleSuccessClose() {
		stage = 'details';
		otpCode = '';
		withdrawDetails = null;
		addressError = null;
		onClose();
	}

	function handleBack() {
		stage = 'details';
		otpCode = '';
	}

	function formatNumber(amount: number, maximumFractionDigits: number = 2): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: maximumFractionDigits,
		}).format(amount);
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
	<ResponsiveDialog.Content>
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Withdraw</ResponsiveDialog.Title>
			<ResponsiveDialog.Description>Withdraw funds to your wallet</ResponsiveDialog.Description>
		</ResponsiveDialog.Header>

		{#if stage === 'details'}
			<form use:enhance method="POST" class="space-y-4">
				{#if !hidePaymentSystemSelection}
					<Form.Field {form} name="paymentSystem">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Payment System</Form.Label>
								<Select.Root
									type="single"
									bind:value={selectedPaymentSystem}
									disabled={isLoading || isLoadingOTP}
									onValueChange={(value) => {
										if (value) {
											selectedPaymentSystem = value;
											$formData.paymentSystem = value;
										}
									}}
								>
									<Select.Trigger class="w-full">
										{availablePaymentSystems.find((p) => p.value === selectedPaymentSystem)
											?.label || 'Select payment system'}
									</Select.Trigger>
									<Select.Content>
										{#each availablePaymentSystems as option}
											<Select.Item value={option.value}>{option.label}</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
				{/if}

				<div
					class="grid gap-4"
					class:grid-cols-2={!hideChainSelection && !hideTokenSelection}
					class:grid-cols-1={hideChainSelection || hideTokenSelection}
				>
					{#if !hideChainSelection}
						<Form.Field {form} name="chain">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Chain</Form.Label>
									<Select.Root
										type="single"
										bind:value={selectedChain}
										disabled={isLoading || isLoadingOTP || !selectedPaymentSystem}
										onValueChange={(value) => {
											if (value) {
												selectedChain = value;
												$formData.chain = value;
												isManualTokenSelection = false;
											}
										}}
									>
										<Select.Trigger class="w-full">
											{availableChains().find((c) => c.value === selectedChain)?.label ||
												'Select chain'}
										</Select.Trigger>
										<Select.Content>
											{#each availableChains() as option}
												<Select.Item value={option.value}>{option.label}</Select.Item>
											{/each}
										</Select.Content>
									</Select.Root>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
					{/if}

					{#if !hideTokenSelection}
						<Form.Field {form} name="token">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Token</Form.Label>
									<Select.Root
										type="single"
										value={selectedToken?.token_type || ''}
										disabled={isLoading || isLoadingOTP || !selectedChain}
										onValueChange={(value) => {
											if (value) {
												const token = availableTokens().find(
													(t) => t.value.token_type === value
												)?.value;
												if (token) {
													selectedToken = token;
													$formData.token = token.token_type;
													isManualTokenSelection = true;
												}
											}
										}}
									>
										<Select.Trigger class="w-full">
											{selectedToken?.token_type || 'Select token'}
										</Select.Trigger>
										<Select.Content>
											{#each availableTokens() as option}
												<Select.Item value={option.value.token_type}>{option.label}</Select.Item>
											{/each}
										</Select.Content>
									</Select.Root>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
					{/if}
				</div>

				<Form.Field {form} name="amount">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Amount ({selectedToken?.token_type || ''})</Form.Label>
							<div class="flex gap-1">
								<Input
									{...props}
									bind:value={$formData.amount}
									type="number"
									step="0.01"
									min={10}
									max={25000}
									placeholder="Enter amount (10 - 25000)"
									disabled={isLoading || isLoadingOTP}
									class="flex-1"
								/>
								<Button
									type="button"
									variant="outline"
									onclick={() => ($formData.amount = Math.min(availableBalance, 25000))}
									disabled={isLoading || isLoadingOTP}
								>
									MAX
								</Button>
								<Button
									type="button"
									variant="outline"
									onclick={() => ($formData.amount = Math.min(availableBalance / 2, 25000))}
									disabled={isLoading || isLoadingOTP}
								>
									HALF
								</Button>
								<Button
									type="button"
									variant="outline"
									onclick={() => ($formData.amount = 10)}
									disabled={isLoading || isLoadingOTP}
								>
									MIN
								</Button>
							</div>
						{/snippet}
					</Form.Control>
					<div class="text-muted-foreground mt-1 flex items-center justify-between px-2 text-sm">
						<p>Available: {formatNumber(availableBalance)}</p>
						<p>Min: 10 | Max: 25000</p>
					</div>
					<Form.FieldErrors />
				</Form.Field>

				<Form.Field {form} name="address">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Withdrawal Address</Form.Label>
							<Input
								{...props}
								bind:value={$formData.address}
								type="text"
								placeholder="Enter withdrawal address"
								disabled={isLoading || isLoadingOTP}
								class="w-full font-mono"
							/>
						{/snippet}
					</Form.Control>
					{#if addressError}
						<p class="text-destructive mt-1 text-sm">{addressError}</p>
					{/if}
					<Form.FieldErrors />
				</Form.Field>

				{#if $formData.amount > availableBalance}
					<div class="bg-destructive/10 border-destructive/20 rounded-md border p-3">
						<p class="text-destructive text-sm">
							Amount exceeds available balance ({formatNumber(availableBalance)})
						</p>
					</div>
				{/if}

				<div class="flex flex-col-reverse items-center justify-between gap-2 pt-4 md:flex-row">
					<Button
						class="w-full md:w-auto md:flex-1"
						variant="outline"
						onclick={handleClose}
						disabled={isLoading || isLoadingOTP}
					>
						Cancel
					</Button>
					<Button
						class="w-full md:w-auto md:flex-1"
						onclick={handleNext}
						disabled={isLoading || isLoadingOTP || !isValidForm()}
					>
						{isLoadingOTP ? 'Sending OTP...' : 'Next'}
					</Button>
				</div>
			</form>
		{:else if stage === 'otp'}
			<div class="space-y-4">
				{#if withdrawDetails}
					<div class="space-y-3 rounded-md border p-4">
						<h3 class="font-semibold">Withdrawal Summary</h3>
						<div class="space-y-2 text-sm">
							<div class="flex justify-between">
								<span class="text-muted-foreground">Amount:</span>
								<span class="font-medium">
									{formatNumber(withdrawDetails.amount)}
									{withdrawDetails.token}
								</span>
							</div>
							<div class="flex justify-between">
								<span class="text-muted-foreground">Chain:</span>
								<span class="font-medium">{withdrawDetails.chain}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-muted-foreground">Token:</span>
								<span class="font-medium">{withdrawDetails.token}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-muted-foreground">Address:</span>
								<span class="font-mono text-xs">{withdrawDetails.address}</span>
							</div>
						</div>
					</div>
				{/if}

				<div class="space-y-2">
					<label for="otp-code" class="text-sm font-medium">OTP Code</label>
					<Input
						id="otp-code"
						type="text"
						bind:value={otpCode}
						placeholder="Enter 8-digit code"
						maxlength="8"
						disabled={isLoading}
						class="w-full text-center font-mono text-2xl tracking-widest"
						oninput={(e) => {
							otpCode = e.currentTarget.value.replace(/\D/g, '').slice(0, 8);
						}}
					/>
					<p class="text-muted-foreground text-xs">Enter the 8-digit code sent to your email</p>
				</div>

				<div class="flex flex-col-reverse items-center justify-between gap-2 pt-4 md:flex-row">
					<Button
						class="w-full md:w-auto md:flex-1"
						variant="outline"
						onclick={handleBack}
						disabled={isLoading}
					>
						Back
					</Button>
					<Button
						class="w-full md:w-auto md:flex-1"
						onclick={handleConfirm}
						disabled={isLoading || otpCode.length !== 8}
					>
						{isLoading ? 'Processing...' : 'Confirm Withdraw'}
					</Button>
				</div>
			</div>
		{:else if stage === 'success'}
			<div class="space-y-4">
				<div class="flex flex-col items-center justify-center space-y-4 py-8">
					<!-- Success Icon -->
					<div
						class="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/20"
					>
						<svg
							class="h-8 w-8 text-green-600 dark:text-green-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							xmlns="http://www.w3.org/2000/svg"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M5 13l4 4L19 7"
							></path>
						</svg>
					</div>

					<!-- Success Message -->
					<div class="space-y-2 text-center">
						<h3 class="text-xl font-semibold">Withdrawal Request Created!</h3>
						<p class="text-muted-foreground text-sm">
							Your withdrawal request has been submitted successfully and is now pending.
						</p>
					</div>

					{#if withdrawDetails}
						<div class="w-full space-y-3 rounded-md border p-4">
							<h4 class="text-sm font-semibold">Transaction Details</h4>
							<div class="space-y-2 text-sm">
								<div class="flex justify-between">
									<span class="text-muted-foreground">Amount:</span>
									<span class="font-medium">
										{formatNumber(withdrawDetails.amount)}
										{withdrawDetails.token}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-muted-foreground">Chain:</span>
									<span class="font-medium">{withdrawDetails.chain}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-muted-foreground">Token:</span>
									<span class="font-medium">{withdrawDetails.token}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-muted-foreground">Address:</span>
									<span class="font-mono text-xs">{withdrawDetails.address}</span>
								</div>
							</div>
						</div>
					{/if}

					<div class="bg-muted/50 rounded-md border p-3">
						<p class="text-muted-foreground text-xs">
							Your request will be processed within 1-2 business days.
						</p>
					</div>
				</div>

				<div class="flex justify-center pt-4">
					<Button class="w-full md:w-auto" onclick={handleSuccessClose}>Close</Button>
				</div>
			</div>
		{/if}
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
