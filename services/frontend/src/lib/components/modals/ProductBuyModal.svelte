<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { desimProductsStore } from '$lib/stores/desim_products.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import { authStore } from '$stores/auth.svelte';
	import toast from 'svelte-french-toast';
	import { goto } from '$app/navigation';
	import ReplenishModal from './ReplenishModal.svelte';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		product: Product | null;
		onSuccess?: () => void;
	}

	let { isOpen = $bindable(), onClose, product, onSuccess }: Props = $props();

	let isLoading = $state(false);
	let error = $state('');
	let isReplenishModalOpen = $state(false);

	function handleClose() {
		if (!isLoading) {
			error = '';
			onClose();
		}
	}

	async function handlePurchase() {
		if (!product) {
			return;
		}

		isLoading = true;
		error = '';

		const result = await desimProductsStore.purchaseProduct(product.id).catch((err) => {
			error = 'Network error occurred';
			toast.error('Failed to purchase product');
			console.error('Product purchase error:', err);
			isLoading = false;
			return null;
		});

		if (!result) {
			if (!error) {
				error = desimProductsStore.error || 'Failed to purchase product';
				toast.error(error);
			}
			isLoading = false;
			return;
		}

		toast.success('Product purchased successfully!');
		onSuccess?.();
		handleClose();
		isLoading = false;
	}

	function handleReplenish() {
		isReplenishModalOpen = true;
		isOpen = false;
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	let hasInsufficientBalance = $derived(
		desimBalanceStore.balance ? desimBalanceStore.balance.available < (product?.amount || 0) : true
	);

	let hasEsim = $derived(authStore.user?.profile_data?.has_esim === true);

	function handleGoToEsim() {
		handleClose();
		goto('/app/esim');
	}
</script>

{#if product}
	<ReplenishModal
		bind:isOpen={isReplenishModalOpen}
		initialAmount={product?.amount - (desimBalanceStore.balance?.available ?? 0)}
		product="desim"
		paymentSystem="bscpay"
		onClose={() => {
			isReplenishModalOpen = false; 
			isOpen = true;
		}}
		onSuccess={() => {
			desimBalanceStore.fetchBalance();
			isReplenishModalOpen = false;
			isOpen = true;
		}}
	/>

	<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
		<ResponsiveDialog.Content class="max-w-md">
			<ResponsiveDialog.Header>
				<ResponsiveDialog.Title>Purchase</ResponsiveDialog.Title>
			</ResponsiveDialog.Header>

			{#if hasEsim}
				<Card class="mt-4">
					<CardHeader>
						<CardTitle class="text-lg">{`${product.metadata?.name ? product.metadata.name : 'Product #' + product.id}`}</CardTitle>
						<CardDescription>
							{product.metadata?.description || 'Description not available'}
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-4">
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium">Price:</span>
							<span class="text-lg font-bold">{formatCurrency(product.amount)}</span>
						</div>

						<div class="flex items-center justify-between">
							<span class="text-sm font-medium">Your Balance:</span>
							<span class="text-lg font-bold">
								{desimBalanceStore.balance
									? formatCurrency(desimBalanceStore.balance.available)
									: 'Loading...'}
							</span>
						</div>

						{#if hasInsufficientBalance}
							<div class="bg-destructive/10 border-destructive/20 rounded-md border p-3">
								<p class="text-destructive text-sm font-medium">Insufficient Balance</p>
								<p class="text-destructive/80 mt-1 text-xs">
									You need {formatCurrency(product.amount - (desimBalanceStore.balance?.available || 0))}
									more to purchase this product.
								</p>
							</div>
						{/if}

						{#if error}
							<div class="bg-destructive/10 border-destructive/20 rounded-md border p-3">
								<p class="text-destructive text-sm">{error}</p>
							</div>
						{/if}
					</CardContent>
				</Card>

				<ResponsiveDialog.Footer class="flex gap-2">
					<Button variant="outline" onclick={handleClose} disabled={isLoading}>Cancel</Button>

					{#if hasInsufficientBalance}
						<Button onclick={handleReplenish} >Deposit</Button>
					{:else}
						<Button onclick={handlePurchase} disabled={isLoading}>
							{isLoading ? 'Processing...' : 'Confirm Purchase'}
						</Button>
					{/if}
				</ResponsiveDialog.Footer>
			{:else}
				<Card class="mt-4">
					<CardHeader>
						<CardTitle class="text-lg">Not Available</CardTitle>
					</CardHeader>
					<CardContent class="pt-6">
						<p class="text-center text-sm text-muted-foreground">
							Farming is not available without an eSIM package
						</p>
					</CardContent>
				</Card>

				<ResponsiveDialog.Footer class="flex gap-2">
					<Button variant="outline" onclick={handleClose}>Cancel</Button>
					<Button onclick={handleGoToEsim}>Go to eSIM</Button>
				</ResponsiveDialog.Footer>
			{/if}
		</ResponsiveDialog.Content>
	</ResponsiveDialog.Root>
{/if}