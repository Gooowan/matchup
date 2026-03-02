<script lang="ts">
	import BalanceCard from '$lib/components/BalanceCard.svelte';
	import EsimPlanCard from '$lib/components/EsimPlanCard.svelte';
	import EsimPurchaseModal from '$lib/components/modals/EsimPurchaseModal.svelte';
	import EsimDetailsModal from '$lib/components/modals/EsimDetailsModal.svelte';
	import { esimStore } from '$lib/stores/esim_store.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { onMount } from 'svelte';
	import { tokenBalanceStore } from '$stores/token_balance.svelte';
	import { authStore } from '$stores/auth.svelte';

	let selectedPlan = $state<EsimPlan | null>(null);
	let selectedOrder = $state<EsimOrder | null>(null);
	let isPurchaseModalOpen = $state(false);
	let isDetailsModalOpen = $state(false);
	let selectedTab = $state('my-esims');

	onMount(async () => {
		try {
			// Load only EU-30 plans (Europe)
			await esimStore.fetchPlans('EU-30');
			await esimStore.fetchUserOrders();
		} catch (error) {
			console.warn('eSIM service not available:', error);
		}
	});

	// Switch to plans tab if user has no eSIMs
	$effect(() => {
		if (
			!esimStore.isUserOrdersLoading &&
			esimStore.userOrders.length === 0 &&
			selectedTab === 'my-esims'
		) {
			selectedTab = 'plans';
		}
	});

	function openPurchaseModal(plan: EsimPlan) {
		selectedPlan = plan;
		isPurchaseModalOpen = true;
	}

	function openDetailsModal(order: EsimOrder) {
		selectedOrder = order;
		isDetailsModalOpen = true;
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString();
	}

	function getStatusBadgeVariant(status: string): 'default' | 'secondary' | 'destructive' {
		switch (status) {
			case 'completed':
				return 'default';
			case 'pending':
				return 'secondary';
			case 'failed':
				return 'destructive';
			default:
				return 'secondary';
		}
	}
</script>

<EsimPurchaseModal
	bind:isOpen={isPurchaseModalOpen}
	plan={selectedPlan}
	onClose={() => {
		isPurchaseModalOpen = false;
		selectedPlan = null;
	}}
	onSuccess={() => {
		isPurchaseModalOpen = false;
		selectedPlan = null;
		desimBalanceStore.fetchBalance();
		authStore.checkAuth();
		esimStore.fetchUserOrders();
	}}
/>

<EsimDetailsModal
	bind:isOpen={isDetailsModalOpen}
	order={selectedOrder}
	onClose={() => {
		isDetailsModalOpen = false;
		selectedOrder = null;
	}}
/>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-foreground">eSIM Store</h1>
		<p class="mt-2 text-muted-foreground">Purchase and manage your eSIMs</p>
	</div>

	<div class="grid gap-6 grid-cols-1 md:grid-cols-2 mb-6">
			<BalanceCard
				store={desimBalanceStore}
				title="Balance"
				showDetails={false}
				showButtons={true}
				product="desim"
				onSuccess={() => {
					desimBalanceStore.fetchBalance();
				}}
			/>
			<BalanceCard
				store={tokenBalanceStore}
				title="Token Balance"
				showDetails={false}
				showButtons={true}
				product="desim"
				onSuccess={() => {
					tokenBalanceStore.fetchBalance();
				}}
			/>
	</div>

		<Tabs.Root bind:value={selectedTab} class="w-full gap-0">
			<Tabs.List class="grid w-full grid-cols-2 m-0 mb-6">
				<Tabs.Trigger value="my-esims">My eSIMs</Tabs.Trigger>
				<Tabs.Trigger value="plans">Available eSIMs</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="plans" >
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					{#if esimStore.isPlansLoading}
						{#each Array(6) as _}
							<Card>
								<CardHeader>
									<div class="h-5 w-3/4 animate-pulse rounded bg-gray-950/30"></div>
									<div class="mt-1.5 h-3.5 w-1/2 animate-pulse rounded bg-gray-950/30"></div>
								</CardHeader>
								<CardContent>
									<div class="space-y-3">
										<div class="h-4 w-full animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-full animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-2/3 animate-pulse rounded bg-gray-950/30"></div>
										<div class="mt-4 h-10 w-full animate-pulse rounded bg-gray-950/30"></div>
									</div>
								</CardContent>
							</Card>
						{/each}
						{:else if esimStore.error}
							<div class="col-span-full py-8 text-center">
								<p class="text-destructive">{esimStore.error}</p>
								<Button onclick={() => esimStore.fetchPlans('EU-30')} class="mt-4">
									Try Again
								</Button>
							</div>
						{:else if esimStore.plans.length === 0}
							<div class="col-span-full py-8 text-center">
								<p class="text-muted-foreground">No eSIM plans available</p>
							</div>
						{:else}
							{#each esimStore.plans as plan}
								<EsimPlanCard {plan} onPurchase={openPurchaseModal} />
							{/each}
						{/if}
					</div>
			</Tabs.Content>

			<Tabs.Content value="my-esims" >
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
				{#if esimStore.isUserOrdersLoading}
					{#each Array(3) as _}
						<Card>
							<CardHeader>
								<div class="flex items-center justify-between">
									<div class="h-5 w-2/3 animate-pulse rounded bg-gray-950/30"></div>
									<div class="h-5 w-16 animate-pulse rounded bg-gray-950/30"></div>
								</div>
								<div class="mt-1.5 h-3.5 w-1/2 animate-pulse rounded bg-gray-950/30"></div>
							</CardHeader>
							<CardContent class="space-y-4">
								<div class="space-y-2">
									<div class="flex items-center justify-between">
										<div class="h-4 w-20 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="flex items-center justify-between">
										<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-20 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="flex items-center justify-between">
										<div class="h-4 w-28 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-24 animate-pulse rounded bg-gray-950/30"></div>
									</div>
								</div>
								<div class="h-10 w-full animate-pulse rounded bg-gray-950/30"></div>
							</CardContent>
						</Card>
					{/each}
					{:else if esimStore.error}
						<div class="col-span-full py-8 text-center">
							<p class="text-destructive">{esimStore.error}</p>
							<Button onclick={() => esimStore.fetchUserOrders()} class="mt-4">
								Try Again
							</Button>
						</div>
					{:else if esimStore.userOrders.length === 0}
						<div class="col-span-full py-8 text-center">
							<p class="text-muted-foreground">You haven't purchased any eSIMs yet</p>
							<Button onclick={() => (selectedTab = 'plans')} class="mt-4">
								Browse eSIM Plans
							</Button>
						</div>
					{:else}
						{#each esimStore.userOrders as order}
							<Card class="transition-shadow hover:shadow-md">
								<CardHeader>
									<CardTitle class="flex items-center justify-between text-lg">
										<span class="flex-1">{order.packageName || order.packageCode}</span>
										<Badge variant={getStatusBadgeVariant(order.status)} class="ml-2 flex-shrink-0 {order.status === 'processing' ? 'animate-pulse' : ''}">
											{order.status}
										</Badge>
									</CardTitle>
									<CardDescription>
										{order.locationCode || 'eSIM Order'}
									</CardDescription>
								</CardHeader>
								<CardContent class="space-y-4">
									<div class="space-y-2 text-sm">
										<div class="flex items-center justify-between">
											<span class="text-muted-foreground">Order ID:</span>
											<span class="font-medium">#{order.id}</span>
										</div>
										<div class="flex items-center justify-between">
											<span class="text-muted-foreground">Amount:</span>
											<span class="font-medium">{formatCurrency(order.amount)}</span>
										</div>
										<div class="flex items-center justify-between">
											<span class="text-muted-foreground">Purchase Date:</span>
											<span class="font-medium">{formatDate(order.createdAt)}</span>
										</div>
									</div>

									{#if order.status === 'completed'}
										<Button onclick={() => openDetailsModal(order)} class="w-full">
											View Details & QR Code
										</Button>
									{:else if order.status === 'pending' || order.status === 'processing'}
										<Button variant="secondary" class="w-full" disabled>
											{order.status === 'processing' ? 'Processing...' : 'Pending...'}
										</Button>
									{:else}
										<Button variant="destructive" onclick={() => openDetailsModal(order)} class="w-full">
											View Error
										</Button>
									{/if}
								</CardContent>
							</Card>
						{/each}
					{/if}
				</div>
			</Tabs.Content>
		</Tabs.Root>
</div>

