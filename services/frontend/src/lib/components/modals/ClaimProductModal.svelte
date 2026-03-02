<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { desimProductsStore } from '$lib/stores/desim_products.svelte';
	import toast from 'svelte-french-toast';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		userProduct: UserDesimProduct | null;
		onSuccess?: () => void;
	}

	let { isOpen = $bindable(), onClose, userProduct, onSuccess }: Props = $props();

	let isLoading = $state(false);
	let isLoadingCommission = $state(false);
	let error = $state('');
	let claimAmount = $state('');
	let commissionPercent = $state<number | null>(null);
	let zeroCommissionPrice = $state<number | null>(null);
	let commissionError = $state('');

	function handleClose() {
		if (!isLoading) {
			error = '';
			commissionError = '';
			claimAmount = '';
			commissionPercent = null;
			zeroCommissionPrice = null;
			onClose();
		}
	}

	async function fetchCommission() {
		if (!userProduct || !claimAmount) {
			commissionPercent = null;
			zeroCommissionPrice = null;
			return;
		}

		const amount = parseFloat(claimAmount);
		if (isNaN(amount) || amount <= 0) {
			commissionPercent = null;
			zeroCommissionPrice = null;
			return;
		}

		if (amount > (userProduct.accrued_balance || 0)) {
			commissionError = 'Claim amount exceeds available balance';
			commissionPercent = null;
			zeroCommissionPrice = null;
			return;
		}

		isLoadingCommission = true;
		commissionError = '';

		try {
			const result = await desimProductsStore.getCommissionPercent(userProduct.id);
			commissionPercent = result.commission_percent;
			zeroCommissionPrice = result.zero_commission_price;
		} catch (err) {
			commissionError = 'Failed to fetch commission';
			commissionPercent = null;
			zeroCommissionPrice = null;
			console.error('Commission fetch error:', err);
		} finally {
			isLoadingCommission = false;
		}
	}

	async function handleClaim() {
		if (!userProduct || !claimAmount) {
			return;
		}

		const amount = parseFloat(claimAmount);
		if (isNaN(amount) || amount <= 0) {
			error = 'Please enter a valid claim amount';
			return;
		}

		if (amount > (userProduct.accrued_balance || 0)) {
			error = 'Claim amount exceeds available balance';
			return;
		}

		isLoading = true;
		error = '';

		const result = await desimProductsStore.claimProduct(userProduct.id, amount).catch((err) => {
			error = 'Network error occurred';
			toast.error('Failed to claim product');
			console.error('Product claim error:', err);
			isLoading = false;
			return null;
		});

		if (!result) {
			if (!error) {
				error = desimProductsStore.error || 'Failed to claim product';
				toast.error(error);
			}
			isLoading = false;
			return;
		}

		toast.success('Product claimed successfully!');
		onSuccess?.();
		handleClose();
		isLoading = false;
	}

	function formatNumber(amount: number, maximumFractionDigits: number = 9): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: maximumFractionDigits,
		}).format(amount);
	}

	let availableToClaim = $derived(userProduct?.accrued_balance || 0);
	let claimAmountNum = $derived(parseFloat(claimAmount) || 0);
	let commissionAmount = $derived(
		commissionPercent !== null && claimAmountNum > 0
			? (claimAmountNum * commissionPercent) / 100
			: 0
	);
	let finalClaimableAmount = $derived(claimAmountNum - commissionAmount);

	let isValidClaimAmount = $derived(
		claimAmountNum > 0 && claimAmountNum <= availableToClaim && !isNaN(claimAmountNum)
	);

	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		if (claimAmount && isValidClaimAmount) {
			debounceTimer = setTimeout(() => {
				fetchCommission();
			}, 500);
		} else {
			commissionPercent = null;
			zeroCommissionPrice = null;
			commissionError = '';
		}

		return () => {
			if (debounceTimer) {
				clearTimeout(debounceTimer);
			}
		};
	});
</script>

{#if userProduct}
	<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
		<ResponsiveDialog.Content class="max-w-md">
			<ResponsiveDialog.Header>
				<ResponsiveDialog.Title>Claim $COMPANY_TOKEN from {`${userProduct.product_metadata.name}`}</ResponsiveDialog.Title>
			</ResponsiveDialog.Header>
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">Available to Claim:</span>
						<span class="text-lg font-bold">{formatNumber(availableToClaim, 2)} $COMPANY_TOKEN</span>
					</div>

					<div class="space-y-2">
						<Label for="claim-amount">Claim Amount ($COMPANY_TOKEN)</Label>
						<div class="flex gap-1">
							<Input
								id="claim-amount"
								type="number"
								step="0.01"
								min="0"
								max={availableToClaim}
								bind:value={claimAmount}
								placeholder="Enter amount to claim"
								disabled={isLoading}
							/>
							<Button variant="outline" onclick={() => claimAmount = availableToClaim.toString()}>MAX</Button>
							<Button variant="outline" onclick={() => claimAmount = (availableToClaim / 2).toString()}>HALF</Button>
						</div>
					</div>

					<!-- Commission section - always rendered to prevent layout shift -->
					<div class="space-y-2 rounded-md border p-3 min-h-[120px]">
						{#if isLoadingCommission}
							<div class="flex items-center justify-center h-full min-h-[80px]">
								<span class="text-muted-foreground text-sm">Loading commission...</span>
							</div>
						{:else if commissionError}
							<div class="bg-destructive/10 border-destructive/20 rounded-md border p-3">
								<p class="text-destructive text-sm">{commissionError}</p>
							</div>
						{:else if commissionPercent !== null && claimAmountNum > 0}
							<div class="space-y-2">
								{#if zeroCommissionPrice === null}
									<p class="text-sm text-green-600 font-medium">
										You can claim all the amount without commission.
									</p>
								{:else}
									<p class="text-sm">
										Commission will be zero when token price reaches ${formatNumber(zeroCommissionPrice)}.
									</p>
								{/if}
								<div class="flex items-center justify-between">
									<span class="text-sm font-medium">Commission:</span>
									<span class="text-sm font-medium">
										{formatNumber(commissionPercent)}%
									</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="text-sm font-medium">Commission Amount:</span>
									<span class="text-muted-foreground text-sm font-medium">
										-{formatNumber(commissionAmount)} $COMPANY_TOKEN
									</span>
								</div>
								<div class="flex items-center justify-between border-t pt-2">
									<span class="text-sm font-bold">Claimable Amount:</span>
									<span class="text-primary text-lg font-bold">
										{formatNumber(finalClaimableAmount)} $COMPANY_TOKEN
									</span>
								</div>
							</div>
						{:else}
							<div class="flex items-center justify-center h-full min-h-[80px]">
								<span class="text-muted-foreground text-sm">Enter amount to see commission details</span>
							</div>
						{/if}
					</div>

					{#if error}
						<div class="bg-destructive/10 border-destructive/20 rounded-md border p-3">
							<p class="text-destructive text-sm">{error}</p>
						</div>
					{/if}

			<ResponsiveDialog.Footer class="flex gap-2">
				<Button variant="outline" onclick={handleClose} disabled={isLoading}>Cancel</Button>
				<Button
					onclick={handleClaim}
					disabled={isLoading || !isValidClaimAmount || isLoadingCommission}
				>
					{isLoading ? 'Processing...' : 'Confirm Claim'}
				</Button>
			</ResponsiveDialog.Footer>
		</ResponsiveDialog.Content>
	</ResponsiveDialog.Root>
{/if}
