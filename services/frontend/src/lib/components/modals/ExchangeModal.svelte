<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import { Card } from '$lib/components/ui/card';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';
	import { formatCurrency } from '$utils/format';
	import { tokenBalanceStore } from '$lib/stores/token_balance.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import { tokenSupplyHistoryStore } from '$lib/stores/token_supply_history.svelte';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import WalletIcon from '@lucide/svelte/icons/wallet';
	import Badge from '$lib/components/ui/badge/badge.svelte';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		onSuccess?: () => void;
	}

	const TOKEN_FLOOR_PRICE = 0.01; // 1 token = $0.01 USD

	let { isOpen = $bindable(), onClose, onSuccess }: Props = $props();

	let isLoading = $state(false);
	let tokenAmountInput = $state('');
	let usdAmountInput = $state('');
	let inputMode = $state<'token' | 'usd'>('token');
	let currentTokenPrice = $state<number | null>(null);
	let showPriceWarning = $derived(currentTokenPrice !== null && currentTokenPrice < 1.0);

	// Derived values
	let tokenAmount = $derived(parseFloat(tokenAmountInput) || 0);
	let usdAmount = $derived(parseFloat(usdAmountInput) || 0);
	let maxTokenAmount = $derived(
		tokenBalanceStore.balance ? tokenBalanceStore.balance.available : 0
	);
	let isInsufficientBalance = $derived(tokenAmount > maxTokenAmount && tokenAmount > 0);
	let isValidAmount = $derived(tokenAmount > 0);
	let canExchange = $derived(isValidAmount && !isInsufficientBalance && !isLoading);

	// Fetch current token price when modal opens
	$effect(() => {
		if (isOpen) {
			currentTokenPrice = tokenSupplyHistoryStore.currentPrice;
			// Reset form
			tokenAmountInput = '';
			usdAmountInput = '';
			inputMode = 'token';
		}
	});

	// Handle token amount input
	function handleTokenInput(event: Event) {
		const target = event.target as HTMLInputElement;
		const value = target.value;

		if (value === '' || /^\d*\.?\d*$/.test(value)) {
			tokenAmountInput = value;
			inputMode = 'token';
			// Calculate USD amount
			const tokenValue = parseFloat(value) || 0;
			usdAmountInput = tokenValue > 0 ? (tokenValue * TOKEN_FLOOR_PRICE).toFixed(2) : '';
		}
	}

	// Handle USD amount input
	function handleUSDInput(event: Event) {
		const target = event.target as HTMLInputElement;
		const value = target.value;

		if (value === '' || /^\d*\.?\d*$/.test(value)) {
			usdAmountInput = value;
			inputMode = 'usd';
			// Calculate token amount
			const usdValue = parseFloat(value) || 0;
			tokenAmountInput = usdValue > 0 ? (usdValue / TOKEN_FLOOR_PRICE).toFixed(2) : '';
		}
	}

	// Helper functions for quick amount buttons
	function setTokenAmount(amount: number) {
		tokenAmountInput = amount.toFixed(2);
		inputMode = 'token';
		usdAmountInput = (amount * TOKEN_FLOOR_PRICE).toFixed(2);
	}

	function setUSDAmount(amount: number) {
		usdAmountInput = amount.toFixed(2);
		inputMode = 'usd';
		tokenAmountInput = (amount / TOKEN_FLOOR_PRICE).toFixed(2);
	}

	// Derived max USD amount
	let maxUSDAmount = $derived(maxTokenAmount * TOKEN_FLOOR_PRICE);

	function handleClose() {
		if (!isLoading) {
			onClose();
		}
	}

	async function handleExchange() {
		if (!isValidAmount) {
			toast.error('Please enter a valid token amount');
			return;
		}

		if (isInsufficientBalance) {
			toast.error('Insufficient token balance');
			return;
		}

		isLoading = true;

		try {
			const requestBody = { token_amount: tokenAmount };

			const resp = await authFetch('/desim/token/exchange', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(requestBody),
			});

			if (!resp.ok) {
				const errorData: ApiResponse<unknown> = await resp.json();
				toast.error(errorData.error || 'Failed to exchange tokens');
				return;
			}

			const response: ApiResponse<{
				token_transaction: Transaction;
				usd_transaction: Transaction;
			}> = await resp.json();

			toast.success(
				`Successfully exchanged ${formatCurrency('$COMPANY_TOKEN', tokenAmount)} for ${formatCurrency('USD', usdAmount)}`
			);

			// Refresh both balances
			await Promise.all([tokenBalanceStore.fetchBalance(), desimBalanceStore.fetchBalance()]);

			// Reset form
			tokenAmountInput = '';
			usdAmountInput = '';

			onSuccess?.();
			handleClose();
		} catch (error) {
			console.error('Exchange error:', error);
			toast.error('An error occurred during exchange');
		} finally {
			isLoading = false;
		}
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
	<ResponsiveDialog.Content class="sm:max-w-md">
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Exchange</ResponsiveDialog.Title>
			<ResponsiveDialog.Description>Convert your $COMPANY_TOKEN to USD</ResponsiveDialog.Description>
		</ResponsiveDialog.Header>

		<div class="space-y-6">
			{#if showPriceWarning}
				<Alert>
					<InfoIcon class="h-4 w-4" />
					<AlertDescription>
						Until the token reaches closer to $1, exchanges will use the fixed rate of $0.01 per
						token.
					</AlertDescription>
				</Alert>
			{/if}

			<div class="relative space-y-2">
				<div class="rounded-lg bg-[#04050D] p-4">
					<div class="space-y-4">
						<div class="flex items-center justify-between">
							<Label for="token-amount">Selling $COMPANY_TOKEN</Label>
							<div class="text-muted-foreground flex items-center gap-2">
								<WalletIcon class="h-3 w-3" />
								{formatCurrency('$DESIM', maxTokenAmount)}
							</div>
						</div>
						<Input
							id="token-amount"
							type="text"
							placeholder="0.00"
							bind:value={tokenAmountInput}
							oninput={handleTokenInput}
							disabled={isLoading}
						/>
						<div class="flex justify-between">
							<p
								class="min-h-5 text-sm transition-opacity"
								class:text-destructive={isInsufficientBalance && isValidAmount}
								class:text-muted-foreground={!isInsufficientBalance || !isValidAmount}
							>
								{#if isInsufficientBalance && isValidAmount}
									Insufficient balance
								{:else}
									Token amount to exchange
								{/if}
							</p>
							<div class="flex gap-1">
								<Badge class="cursor-pointer" onclick={() => setTokenAmount(maxTokenAmount / 2)}>
									HALF
								</Badge>
								<Badge class="cursor-pointer" onclick={() => setTokenAmount(maxTokenAmount)}>
									MAX
								</Badge>
							</div>
						</div>
					</div>
				</div>

				<div
					class="stroke-primary absolute left-1/2 top-1/2 z-[500] -translate-x-1/2 -translate-y-1/2 rounded-full bg-[#04050D] p-2"
				>
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 14l-7 7m0 0l-7-7m7 7V3"
						></path>
					</svg>
				</div>

				<div class="rounded-lg bg-[#04050D] p-4">
					<div class="space-y-4">
						<div class="flex items-center justify-between">
							<Label for="usd-amount">Receiving USD</Label>
							<div class="text-muted-foreground flex items-center gap-2">
								<WalletIcon class="h-3 w-3" />
								{formatCurrency('USD', desimBalanceStore.balance?.available || 0)}
							</div>
						</div>
						<Input
							id="usd-amount"
							type="text"
							value={usdAmountInput}
							placeholder="0.00"
							oninput={handleUSDInput}
							disabled={isLoading}
						/>
						<div class="flex justify-between">
							<p class="text-muted-foreground text-sm">
								Rate: 1 token = ${TOKEN_FLOOR_PRICE.toFixed(2)} USD
							</p>
							<div class="flex gap-1">
								<Badge class="cursor-pointer" onclick={() => setUSDAmount(maxUSDAmount / 2)}>
									HALF
								</Badge>
								<Badge class="cursor-pointer" onclick={() => setUSDAmount(maxUSDAmount)}>MAX</Badge>
							</div>
						</div>
					</div>
				</div>
			</div>

			<p class="text-muted-foreground text-sm">
				Balance after exchange: <br />
				{formatCurrency('USD', (desimBalanceStore.balance?.available || 0) + usdAmount)} USD
			</p>

			{#if showPriceWarning && usdAmount > 0}
				<Alert variant="destructive">
					<InfoIcon class="h-4 w-4" />
					<AlertDescription>
						Warning: You are selling at a minimal price <br />
						with a {formatCurrency('USD', usdAmount)} value.
						<br />
						Token potential value now is ${(tokenAmount * (currentTokenPrice || 0)).toFixed(2)}
					</AlertDescription>
				</Alert>
			{/if}

			<div class="flex flex-col-reverse gap-2 md:flex-row">
				<Button variant="outline" onclick={handleClose} disabled={isLoading} class="flex-1">
					Cancel
				</Button>
				<Button onclick={handleExchange} disabled={!canExchange} class="flex-1">
					{#if isLoading}
						Exchanging...
					{:else if isInsufficientBalance && isValidAmount}
						Insufficient Balance
					{:else if !isValidAmount}
						Enter Amount
					{:else}
						Exchange
					{/if}
				</Button>
			</div>
		</div>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
