<script lang="ts">
	import { formatCurrency } from '$utils/format';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import ReplenishModal from '$lib/components/modals/ReplenishModal.svelte';
	import ExchangeModal from '$lib/components/modals/ExchangeModal.svelte';
	import WithdrawModal from '$lib/components/modals/WithdrawModal.svelte';
	import { onMount } from 'svelte';

	interface Props {
		store: {
			balance: Balance | null;
			currency: string;
			isBalanceLoading: boolean;
			isTransactionsLoading: boolean;
			error: string | null;
			fetchBalance: () => Promise<Balance | null>;
		};
		title?: string;
		showDetails?: boolean;
		showButtons?: boolean;
		paymentSystem?: string;
		product?: string;
		onSuccess?: () => void;
	}

	let {
		store,
		title = 'Balance',
		showDetails = true,
		showButtons = false,
		paymentSystem = 'bscpay',
		product = 'grow',
		onSuccess,
	}: Props = $props();

	let isReplenishModalOpen = $state(false);
	let isExchangeModalOpen = $state(false);
	let isWithdrawModalOpen = $state(false);

	onMount(() => {
		store.fetchBalance();
	});
</script>

<Card class="w-full">
	<CardHeader>
		<CardTitle class="flex items-center justify-between">
			{title}
			{#if store.isBalanceLoading}
				<Badge class="bg-transparent text-foreground">Loading...</Badge>
			{:else if store.error}
				<Badge variant="destructive">Error</Badge>
			{:else}
				<Badge class="bg-transparent text-foreground">{store.currency}</Badge>
			{/if}
		</CardTitle>
	</CardHeader>
	<CardContent>
		{#if store.error}
			<div class="py-4 text-center">
				<p class="text-sm text-destructive">{store.error}</p>
			</div>
		{:else}
			{@const balance = store.balance || {
				available: 0,
				withdraw_available: 0,
				withdraw_pending: 0,
				on_hold: 0,
				referral: 0,
				total_invested: 0,
			}}
			<div class="space-y-4" class:opacity-50={store.isBalanceLoading}>
				<!-- Main Balance -->
				<div class="text-center">
					<div class="text-3xl font-bold text-foreground">
						{formatCurrency('', balance.available)}
					</div>
					<p class="mt-1 text-sm text-muted-foreground">Available</p>
				</div>

				<!-- Action Buttons -->
				{#if showButtons}
					{#if store.currency === '$COMPANY_TOKEN'}
						<div class="flex justify-center gap-2">
							<Button
								onclick={() => (isExchangeModalOpen = true)}
								class="flex-1"
								disabled={store.isBalanceLoading}
							>
								Exchange
							</Button>
							<Button variant="outline" disabled class="flex-1">Trade <span class="text-[10px] text-muted-foreground mb-2">comming soon</span></Button>
						</div>
					{:else}
						<div class="flex justify-center gap-2">
							<Button
								onclick={() => (isReplenishModalOpen = true)}
								class="flex-1"
								disabled={store.isBalanceLoading}
							>
								Deposit
							</Button>
							<Button
								variant="outline"
								onclick={() => (isWithdrawModalOpen = true)}
								class="flex-1"
								disabled={store.isBalanceLoading}
							>
								Withdraw
							</Button>
						</div>
					{/if}
				{/if}

				{#if showDetails}
					<div class="space-y-3 border-t pt-4">
						<!-- Withdraw Available -->
						<div class="grid grid-cols-2 items-center gap-4">
							<div class="text-lg font-semibold text-foreground">
								{formatCurrency(store.currency, balance.withdraw_available)}
							</div>
							<p class="text-xs text-muted-foreground">Withdraw Available</p>
						</div>

						<!-- Pending Withdraw -->
						<div class="grid grid-cols-2 items-center gap-4">
							<div class="text-lg font-semibold text-foreground">
								{formatCurrency(store.currency, balance.withdraw_pending)}
							</div>
							<p class="text-xs text-muted-foreground">Pending Withdraw</p>
						</div>

						<!-- On Hold -->
						<div class="grid grid-cols-2 items-center gap-4">
							<div class="text-lg font-semibold text-foreground">
								{formatCurrency(store.currency, balance.on_hold)}
							</div>
							<p class="text-xs text-muted-foreground">On Hold</p>
						</div>

						<div class="col-span-2 border-t"></div>
						<!-- Referral -->
						<div class="grid grid-cols-2 items-center gap-4">
							<div class="text-lg font-semibold text-foreground">
								{formatCurrency(store.currency, balance.referral)}
							</div>
							<p class="text-xs text-muted-foreground">Total Referral</p>
						</div>

						<!-- Total Invested -->
						<div class="grid grid-cols-2 items-center gap-4">
							<div class="text-lg font-semibold text-foreground">
								{formatCurrency(store.currency, balance.total_invested)}
							</div>
							<p class="text-xs text-muted-foreground">Total Invested</p>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</CardContent>
</Card>

<!-- Replenish Modal -->
<ReplenishModal
	bind:isOpen={isReplenishModalOpen}
	onClose={() => (isReplenishModalOpen = false)}
	{paymentSystem}
	{product}
	onSuccess={() => {
		// Call the parent's success callback
		onSuccess?.();
	}}
/>

<!-- Exchange Modal -->
{#if store.currency === '$COMPANY_TOKEN'}
	<ExchangeModal
		bind:isOpen={isExchangeModalOpen}
		onClose={() => (isExchangeModalOpen = false)}
		onSuccess={() => {
			// Call the parent's success callback
			onSuccess?.();
		}}
	/>
{/if}

<!-- Withdraw Modal -->
{#if store.currency !== '$COMPANY_TOKEN'}
	{@const balance = store.balance || {
		available: 0,
		withdraw_available: 0,
		withdraw_pending: 0,
		on_hold: 0,
		referral: 0,
		total_invested: 0,
	}}
	<WithdrawModal
		bind:isOpen={isWithdrawModalOpen}
		onClose={() => (isWithdrawModalOpen = false)}
		{product}
		{paymentSystem}
		availableBalance={balance.withdraw_available}
		onSuccess={() => {
			store.fetchBalance();
			onSuccess?.();
		}}
	/>
{/if}
