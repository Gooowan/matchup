<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Form from '$lib/components/ui/form';
	import * as Select from '$lib/components/ui/select';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import * as Tabs from '$lib/components/ui/tabs';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';
	import { valibot } from 'sveltekit-superforms/adapters';
	import { superForm, defaults } from 'sveltekit-superforms';
	import * as v from 'valibot';

	import SolanaPayModal from '$lib/components/modals/SolanaPayModal.svelte';
	import BSCPayModal from '$lib/components/modals/BSCPayModal.svelte';
	import PendingInvoices from '$components/PendingInvoices.svelte';

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
		initialAmount?: number;
		minAmount?: number;
		maxAmount?: number;
		product: string;
		onSuccess?: () => void;
		paymentSystem?: string;
		chain?: string;
		token?: string;
	}

	let {
		isOpen = $bindable(),
		onClose,
		initialAmount = 10,
		minAmount = 10,
		maxAmount = 100_000,
		product,
		onSuccess,
		paymentSystem = 'bscpay',
		chain,
		token,
	}: Props = $props();

	let isSolanaModalOpen = $state(false);
	let solanaPaymentAddress = $state<string>('');
	let solanaPaymentAmount = $state<number>(0);

	let isBSCModalOpen = $state(false);
	let bscPaymentAddress = $state<string>('');
	let bscPaymentAmount = $state<number>(0);
	let bscPaymentToken = $state<string>('BNB');

	function closeSolanaModal() {
		isSolanaModalOpen = false;
		isOpen = true;
	}

	function openSolanaModal() {
		isSolanaModalOpen = true;
		isOpen = false;
	}

	function closeBSCModal() {
		isBSCModalOpen = false;
		isOpen = true;
	}

	function openBSCModal() {
		isBSCModalOpen = true;
		isOpen = false;
	}

	let isLoading = $state(false);
	let paymentMethods = $state<PaymentMethods>({});
	let selectedPaymentSystem = $state<string>('');
	let selectedChain = $state<string>('');
	let selectedToken = $state<Token | null>(null);
	let isManualTokenSelection = $state(false);

	let activeTab = $state('create');
	let hasPendingInvoices = $state(false);
	let hasCheckedPending = $state(false);

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
	let hideChainSelection = $derived(!!chain);
	let hideTokenSelection = $derived(!!token);

	const replenishSchema = v.object({
		amount: v.pipe(
			v.number('Amount must be a number'),
			v.minValue(0.01, 'Amount must be at least 0.01')
		),
		paymentSystem: v.pipe(v.string(), v.minLength(1, 'Please select a payment system')),
		chain: v.pipe(v.string(), v.minLength(1, 'Please select a chain')),
		token: v.pipe(v.string(), v.minLength(1, 'Please select a token')),
	});

	type FormData = v.InferInput<typeof replenishSchema>;

	const initialData: FormData = {
		amount: initialAmount ?? minAmount,
		paymentSystem: paymentSystem || '',
		chain: chain || '',
		token: token || '',
	};

	const form = superForm(defaults(initialData, valibot(replenishSchema)), {
		SPA: true,
		dataType: 'json',
		validators: valibot(replenishSchema),
		validationMethod: 'oninput',
		onError: ({ result }) => {
			toast.error(`${result.error}`);
		},
		async onUpdate({ form }) {
			if (form.valid) {
				isLoading = true;

				const invoiceData = {
					chain: form.data.chain,
					token: form.data.token,
					amount: form.data.amount,
				};

				const resp = await authFetch(`/payments/${product}/invoice/${form.data.paymentSystem}`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify(invoiceData),
				});
				const response: ApiResponse<Invoice> = await resp.json();

				if (resp.status === 200 && response.data) {
					toast.success('Invoice created successfully!');
					isLoading = false;
					switch (form.data.paymentSystem) {
						case 'solanapay':
							if (response.data.metadata?.pool_address) {
								solanaPaymentAddress = response.data.metadata.pool_address;
								solanaPaymentAmount = response.data.amount;
								openSolanaModal();
							}
							break;
						case 'bscpay':
							if (response.data.metadata?.pool_address) {
								bscPaymentAddress = response.data.metadata.pool_address;
								bscPaymentAmount = response.data.amount;
								bscPaymentToken = response.data.token || 'BNB';
								openBSCModal();
							}
							break;
						default:
							onSuccess?.();
							handleClose();
							break;
					}
					return;
				}

				toast.error(response.error || 'Failed to create invoice');
				isLoading = false;
			}
		},
	});

	const { form: formData, enhance } = form;

	// Initialize selected values from props when modal opens
	$effect(() => {
		if (isOpen) {
			if (Object.keys(paymentMethods).length === 0) {
				fetchPaymentMethods();
			}

			// Reset form data with current prop values
			$formData.paymentSystem = paymentSystem || '';
			$formData.chain = chain || '';
			$formData.token = token || '';
			$formData.amount = initialAmount ?? minAmount;

			if (paymentSystem) {
				selectedPaymentSystem = paymentSystem;
			} else {
				selectedPaymentSystem = '';
			}

			if (chain) {
				selectedChain = chain;
			} else {
				selectedChain = '';
			}

			if (token && paymentSystem && chain) {
				// Find the token object from the available tokens
				const foundChain = paymentMethods[paymentSystem]?.find((c) => c.chain_type === chain);
				const foundToken = foundChain?.tokens.find((t) => t.token_type === token);
				if (foundToken) {
					selectedToken = foundToken;
				} else {
					selectedToken = null;
				}
			} else {
				// Clear selected token if props are not provided
				selectedToken = null;
			}
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
		}

		const firstChain = paymentMethods[paymentSystem][0];
		if (!firstChain) return;

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

	// Check for pending invoices when modal opens (only once per session)
	$effect(() => {
		if (isOpen && !hasCheckedPending) {
			checkPendingInvoices();
		}
	});

	async function checkPendingInvoices() {
		hasCheckedPending = true;

		const response = await authFetch(`/payments/${product}/pending?page=1&take=1`);
		if (!response.ok) {
			return;
		}

		const data: ApiPaginatedResponse<Invoice> = await response.json();

		if (data.meta && data.meta.itemCount > 0) {
			hasPendingInvoices = true;
			activeTab = 'pending';
		}
	}

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

	function handleClose() {
		if (!isLoading) {
			hasCheckedPending = false;
			onClose();
		}
	}

	function handleInvoiceClick(invoice: Invoice) {
		const paymentSystem = invoice.metadata?.payment_system;
		if (!paymentSystem) return;

		switch (paymentSystem) {
			case 'solanapay':
				if (invoice.metadata?.pool_address) {
					solanaPaymentAddress = invoice.metadata.pool_address;
					solanaPaymentAmount = invoice.amount;
					openSolanaModal();
				}
				break;
			case 'bscpay':
				if (invoice.metadata?.pool_address) {
					bscPaymentAddress = invoice.metadata.pool_address;
					bscPaymentAmount = invoice.amount;
					bscPaymentToken = invoice.token || 'BNB';
					openBSCModal();
				}
				break;
			default:
				break;
		}
	}
</script>

<SolanaPayModal
	bind:isOpen={isSolanaModalOpen}
	onClose={closeSolanaModal}
	onSuccess={() => {
		closeSolanaModal();
		onSuccess?.();
		handleClose();
	}}
	paymentAddress={solanaPaymentAddress}
	amount={solanaPaymentAmount}
/>

<BSCPayModal
	bind:isOpen={isBSCModalOpen}
	bind:token={bscPaymentToken}
	paymentAddress={bscPaymentAddress}
	amount={bscPaymentAmount}
	onClose={closeBSCModal}
	onSuccess={() => {
		closeBSCModal();
		onSuccess?.();
		handleClose();
	}}
/>

<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
	<ResponsiveDialog.Content>
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Deposit</ResponsiveDialog.Title>
			<ResponsiveDialog.Description>Add funds to your balance</ResponsiveDialog.Description>
		</ResponsiveDialog.Header>

		<Tabs.Root bind:value={activeTab}>
			{#if hasPendingInvoices}
				<Tabs.List class="grid w-full grid-cols-2">
					<Tabs.Trigger value="pending">Pending Invoices</Tabs.Trigger>
					<Tabs.Trigger value="create">Create New</Tabs.Trigger>
				</Tabs.List>
			{/if}

			<Tabs.Content value="pending" class="mt-4">
				<PendingInvoices {product} bind:hasPendingInvoices onInvoiceClick={handleInvoiceClick} />
			</Tabs.Content>

			<Tabs.Content value="create" class={hasPendingInvoices ? 'mt-4' : ''}>
				<form use:enhance method="POST" class="space-y-4">
					{#if !hidePaymentSystemSelection}
						<Form.Field {form} name="paymentSystem">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Payment System</Form.Label>
									<Select.Root
										type="single"
										bind:value={selectedPaymentSystem}
										disabled={isLoading}
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
											disabled={isLoading || !selectedPaymentSystem}
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
											disabled={isLoading || !selectedChain}
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
								<Form.Label>{selectedToken?.token_type || 'Amount'}</Form.Label>
								<Input
									{...props}
									bind:value={$formData.amount}
									type="number"
									step="0.01"
									min={selectedToken?.min_amount ?? minAmount}
									max={selectedToken?.max_amount ?? maxAmount}
									placeholder={selectedToken
										? `Enter amount (${selectedToken.min_amount} - ${selectedToken.max_amount})`
										: 'Enter amount'}
									disabled={isLoading}
									class="w-full"
								/>
							{/snippet}
						</Form.Control>
						{#if selectedToken}
							<div
								class="text-muted-foreground mt-1 flex items-center justify-between px-2 text-sm"
							>
								<p>
									Min: {selectedToken.min_amount}
								</p>
								<p>
									Max: {selectedToken.max_amount}
								</p>
							</div>
						{/if}
						<Form.FieldErrors />
					</Form.Field>

					{#if selectedPaymentSystem === 'demo'}
						<div class="rounded-md p-3">
							<p class="text-destructive text-sm">
								<strong>Note:</strong> This is a demo payment system. The transaction will be processed
								immediately with success status.
							</p>
						</div>
					{/if}

					<div class="flex flex-col-reverse items-center justify-between gap-2 pt-4 md:flex-row">
						<Button
							class="w-full md:w-auto md:flex-1"
							variant="outline"
							onclick={handleClose}
							disabled={isLoading}>Cancel</Button
						>
						<Form.Button class="w-full md:w-auto md:flex-1" disabled={isLoading}>
							{isLoading ? 'Processing...' : 'Create Invoice'}
						</Form.Button>
					</div>
				</form>
			</Tabs.Content>
		</Tabs.Root>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
