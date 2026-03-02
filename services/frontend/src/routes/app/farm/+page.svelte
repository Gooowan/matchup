<script lang="ts">
	import BalanceCard from '$lib/components/BalanceCard.svelte';
	import ProductBuyModal from '$lib/components/modals/ProductBuyModal.svelte';
	import ClaimProductModal from '$lib/components/modals/ClaimProductModal.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import { tokenBalanceStore } from '$stores/token_balance.svelte';
	import { desimProductsStore } from '$lib/stores/desim_products.svelte';
	import { tokenSupplyHistoryStore } from '$lib/stores/token_supply_history.svelte';
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
	import TokenRecentChart from '$components/charts/TokenRecentChart.svelte';

	let selectedProduct = $state(null);
	let isProductModalOpen = $state(false);
	let selectedUserProduct = $state(null);
	let isClaimModalOpen = $state(false);
	let selectedTab = $state('my-products');

	const TOKEN_FLOOR_PRICE = 0.01;

	onMount(() => {
		desimProductsStore.refreshAll();
	});

	// Switch to products tab if user has no products
	$effect(() => {
		if (
			!desimProductsStore.isUserProductsLoading &&
			desimProductsStore.userProducts.length === 0 &&
			selectedTab === 'my-products'
		) {
			selectedTab = 'products';
		}
	});

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function formatToken(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function openProductModal(product: any) {
		selectedProduct = product;
		isProductModalOpen = true;
	}

	function openClaimModal(userProduct: any) {
		selectedUserProduct = userProduct;
		isClaimModalOpen = true;
	}
</script>

<ProductBuyModal
	bind:isOpen={isProductModalOpen}
	product={selectedProduct}
	onClose={() => {
		isProductModalOpen = false;
		selectedProduct = null;
	}}
	onSuccess={() => {
		isProductModalOpen = false;
		selectedProduct = null;
		desimBalanceStore.fetchBalance();
		tokenBalanceStore.fetchBalance();
		desimProductsStore.fetchUserProducts();
	}}
/>

<ClaimProductModal
	bind:isOpen={isClaimModalOpen}
	userProduct={selectedUserProduct}
	onClose={() => {
		isClaimModalOpen = false;
		selectedUserProduct = null;
	}}
	onSuccess={() => {
		isClaimModalOpen = false;
		selectedUserProduct = null;
		desimBalanceStore.fetchBalance();
		tokenBalanceStore.fetchBalance();
		desimProductsStore.fetchUserProducts();
	}}
/>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<h1 class="text-foreground text-3xl font-bold">Farm $COMPANY_TOKEN</h1>
		<p class="text-muted-foreground mt-2">Get your share of the Global Power</p>
	</div>

	<div class="mb-6 grid grid-cols-1 gap-6 md:grid-cols-2">
		<BalanceCard
			store={desimBalanceStore}
			title="Balance"
			showDetails={false}
			showButtons={true}
			paymentSystem={import.meta.env.ENABLE_DEMO_PAYMENTS ? 'demo' : 'bscpay'}
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

	<div class="mb-6">
		<TokenRecentChart />
	</div>

	<!-- Products and User Products Tabs -->
	<div class="mt-6">
		<Tabs.Root bind:value={selectedTab} class="w-full gap-0">
			<Tabs.List class="m-0 grid w-full grid-cols-2">
				<Tabs.Trigger value="my-products">My Products</Tabs.Trigger>
				<Tabs.Trigger value="products">Available Products</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="products" class="mt-6">
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					{#if desimProductsStore.isProductsLoading}
						{#each Array(3) as _}
							<Card>
								<CardHeader>
									<div class="bg-muted h-4 w-3/4 animate-pulse rounded"></div>
									<div class="bg-muted h-3 w-1/2 animate-pulse rounded"></div>
								</CardHeader>
								<CardContent>
									<div class="space-y-2">
										<div class="bg-muted h-4 w-full animate-pulse rounded"></div>
										<div class="bg-muted h-4 w-2/3 animate-pulse rounded"></div>
									</div>
								</CardContent>
							</Card>
						{/each}
					{:else if desimProductsStore.error}
						<div class="col-span-full py-8 text-center">
							<p class="text-destructive">{desimProductsStore.error}</p>
							<Button onclick={() => desimProductsStore.fetchProducts()} class="mt-4">
								Try Again
							</Button>
						</div>
					{:else if desimProductsStore.products.length === 0}
						<div class="col-span-full py-8 text-center">
							<p class="text-muted-foreground">No products available</p>
						</div>
					{:else}
						{#each desimProductsStore.products as product}
							{@const productBG = product.metadata.bg}
							<Card
								style="background: linear-gradient(120deg, {productBG}66, {productBG}1a); border-color: {productBG}1a;"
								class="transition-shadow hover:shadow-md"
							>
								<CardHeader class="gap-1">
									<CardTitle class="flex items-center justify-between">
										{product.metadata?.name || `Product #${product.id}`}
										<Badge variant="default">Available</Badge>
									</CardTitle>
									<CardDescription>
										{product.metadata?.description || 'Description not available'}
									</CardDescription>
								</CardHeader>
								<img src={product.metadata.img} />
								<CardContent class="space-y-4">
									<div class="text-center text-2xl font-bold">
										{formatCurrency(product.amount)}
									</div>
									<div class="space-y-1">

										<div class="flex w-full justify-between">
											<span class="text-muted-foreground text-sm">Tokens per Day:</span>
											<span class="text-sm">
												{(Number(product.amount / TOKEN_FLOOR_PRICE * product.metadata.return_rate) / 30).toFixed(0) || '0'}
											</span>
										</div>
										<div class="flex w-full justify-between">
											<span class="text-muted-foreground text-sm">Return Rate:</span>
											<span class="text-sm">
												{Number(product.metadata.return_rate) * 100 || '0'}% per cycle
											</span>
										</div>
										<p class="text-muted-foreground text-sm">
											This package generates you {(product.amount * 2 / TOKEN_FLOOR_PRICE).toFixed(0)} tokens in total in 2 months. You can hold them or sell anytime for {formatCurrency(product.amount * 2)}
										</p>
									</div>
									<Button onclick={() => openProductModal(product)} class="w-full">Purchase</Button>
								</CardContent>
							</Card>
						{/each}
					{/if}
				</div>
			</Tabs.Content>

			<Tabs.Content value="my-products" class="mt-6">
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					{#if desimProductsStore.isUserProductsLoading}
						{#each Array(3) as _}
							<Card>
								<CardHeader>
									<div class="bg-muted h-4 w-3/4 animate-pulse rounded"></div>
									<div class="bg-muted h-3 w-1/2 animate-pulse rounded"></div>
								</CardHeader>
								<CardContent>
									<div class="space-y-2">
										<div class="bg-muted h-4 w-full animate-pulse rounded"></div>
										<div class="bg-muted h-4 w-2/3 animate-pulse rounded"></div>
									</div>
								</CardContent>
							</Card>
						{/each}
					{:else if desimProductsStore.error}
						<div class="col-span-full py-8 text-center">
							<p class="text-destructive">{desimProductsStore.error}</p>
							<Button onclick={() => desimProductsStore.fetchUserProducts()} class="mt-4">
								Try Again
							</Button>
						</div>
					{:else if desimProductsStore.userProducts.length === 0}
						<div class="col-span-full py-8 text-center">
							<p class="text-muted-foreground">You haven't purchased any products yet</p>
							<Button onclick={() => (selectedTab = 'products')} class="mt-4">
								Browse Products
							</Button>
						</div>
					{:else}
						{#each desimProductsStore.userProducts as userProduct}
							{@const productBG = userProduct.product_metadata.bg}
							<Card
								style="background: linear-gradient(120deg, {productBG}66, {productBG}1a); border-color: {productBG}1a;"
							>
								<CardHeader class="gap-0">
									<CardTitle class="flex items-center justify-between">
										{userProduct.product_metadata?.name || `Product #${userProduct.product_id}`}
										<Badge variant={userProduct.is_active ? 'default' : 'secondary'}>
											{userProduct.is_active ? 'Active' : 'Inactive'}
										</Badge>
									</CardTitle>
									<CardDescription>
										{userProduct.product_metadata?.description || 'Grow token investment product'}
									</CardDescription>
								</CardHeader>
								<img src={userProduct.product_metadata.img} />
								<CardContent class="space-y-4">
									<div class="flex flex-col items-center justify-between">
										<p class="text-2xl font-bold">
											{formatToken(userProduct.accrued_balance) || '0'} 
										</p>
										<p class="text-sm">
											~ {formatCurrency(userProduct.accrued_balance * (tokenSupplyHistoryStore.currentPrice || 0))}
										</p>
										<p class="text-muted-foreground text-sm">Available to claim</p>
									</div>
									{#if userProduct.accrued_balance > 0}
										<Button
											onclick={() => openClaimModal(userProduct)}
											class="w-full"
											variant="default"
										>
											Claim
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
</div>
