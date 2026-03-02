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
	import { esimStore } from '$lib/stores/esim_store.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import toast from 'svelte-french-toast';
	import ReplenishModal from './ReplenishModal.svelte';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		plan: EsimPlan | null;
		onSuccess?: () => void;
	}

	let { isOpen = $bindable(), onClose, plan, onSuccess }: Props = $props();

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
		if (!plan) {
			return;
		}

		isLoading = true;
		error = '';

		const result = await esimStore.purchaseEsim(plan.packageCode).catch((err) => {
			error = 'Network error occurred';
			toast.error('Failed to purchase eSIM');
			console.error('eSIM purchase error:', err);
			isLoading = false;
			return null;
		});

		if (!result) {
			if (!error) {
				error = esimStore.error || 'Failed to purchase eSIM';
				toast.error(error);
			}
			isLoading = false;
			return;
		}

		toast.success('eSIM ordered successfully! You will receive details shortly.');
		desimBalanceStore.fetchBalance();
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
		plan && desimBalanceStore.balance
			? desimBalanceStore.balance.available < plan.price
			: false
	);
</script>

{#if plan}
	<ReplenishModal
		bind:isOpen={isReplenishModalOpen}
		initialAmount={plan?.price - (desimBalanceStore.balance?.available ?? 0)}
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
				<ResponsiveDialog.Title>Purchase eSIM</ResponsiveDialog.Title>
				<ResponsiveDialog.Description>
					Confirm your eSIM purchase. You will receive activation details shortly.
				</ResponsiveDialog.Description>
			</ResponsiveDialog.Header>

			<Card class="mt-4">
				<CardHeader>
					<CardTitle class="text-lg">{plan.name}</CardTitle>
					<CardDescription>
						{plan.description || plan.locationCode}
					</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium">Data:</span>
							<span class="font-semibold">{plan.data}</span>
						</div>

						<div class="flex items-center justify-between">
							<span class="text-sm font-medium">Duration:</span>
							<span class="font-semibold">{plan.duration} days</span>
						</div>

						<div class="flex items-center justify-between border-t pt-2">
							<span class="text-sm font-medium">Price:</span>
							<span class="text-lg font-bold">{formatCurrency(plan.price)}</span>
						</div>

						<div class="flex items-center justify-between">
							<span class="text-sm font-medium">Your Balance:</span>
							<span class="text-lg font-bold">
								{desimBalanceStore.balance
									? formatCurrency(desimBalanceStore.balance.available)
									: 'Loading...'}
							</span>
						</div>
					</div>

					{#if hasInsufficientBalance}
						<div class="rounded-md border border-destructive/20 bg-destructive/10 p-3">
							<p class="text-sm font-medium text-destructive">Insufficient Balance</p>
							<p class="mt-1 text-xs text-destructive/80">
								You need {formatCurrency(
									plan.price - (desimBalanceStore.balance?.available || 0)
								)} more to purchase this eSIM.
							</p>
						</div>
					{/if}

					{#if error}
						<div class="rounded-md border border-destructive/20 bg-destructive/10 p-3">
							<p class="text-sm text-destructive">{error}</p>
						</div>
					{/if}
				</CardContent>
			</Card>

			<ResponsiveDialog.Footer class="flex gap-2">
				<Button variant="outline" onclick={handleClose} disabled={isLoading}>Cancel</Button>

				{#if hasInsufficientBalance}
					<Button onclick={handleReplenish}>Deposit</Button>
				{:else}
					<Button onclick={handlePurchase} disabled={isLoading}>
						{isLoading ? 'Processing...' : 'Confirm Purchase'}
					</Button>
				{/if}
			</ResponsiveDialog.Footer>
		</ResponsiveDialog.Content>
	</ResponsiveDialog.Root>
{/if}

