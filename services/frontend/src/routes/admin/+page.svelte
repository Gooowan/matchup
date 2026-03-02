<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$utils/authFetch';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { adminWithdrawsStore } from '$lib/stores/admin_withdraws.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import UsersIcon from '@lucide/svelte/icons/users';
	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import DollarSignIcon from '@lucide/svelte/icons/dollar-sign';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import TokenPriceChart from '$components/charts/TokenPriceChart.svelte';
	import HotwalletBalanceCard from '$lib/components/admin/HotwalletBalanceCard.svelte';
	import PendingWithdrawalsCard from '$lib/components/admin/PendingWithdrawalsCard.svelte';

	let stats = $state({
		totalUsers: 0,
		totalTransactions: 0,
		pendingInvoices: 0,
		totalRevenue: 0,
		esimOrders: 0,
		esimSumWithMarkup: 0,
		esimSumWithoutMarkup: 0,
		pendingWithdrawals: { USDC: 0, USDT: 0 } as Record<string, number>,
	});
	
	let isLoading = $state(true);

	onMount(async () => {
		adminWithdrawsStore.fetchHotwalletBalance()
		// Fetch core stats
		const coreResp = await authFetch('/admin/stats');
		if (coreResp.status === 200) {
			const coreData = await coreResp.json();
			if (coreData.data) {
				stats.totalUsers = coreData.data.totalUsers || 0;
			}
		}

		// Fetch desim stats
		const desimResp = await authFetch('/admin/desim/stats');
		if (desimResp.status === 200) {
			const desimData = await desimResp.json();
			if (desimData.data) {
				stats.totalTransactions = desimData.data.totalUserDesimProducts || 0;
				stats.totalRevenue = desimData.data.totalInvested || 0;
				if (desimData.data.pendingWithdrawals) {
					stats.pendingWithdrawals = {
						USDC: desimData.data.pendingWithdrawals.USDC || 0,
						USDT: desimData.data.pendingWithdrawals.USDT || 0,
					};
				}
			}
		}

		// Fetch payments stats
		const paymentsResp = await authFetch('/admin/payments/stats');
		if (paymentsResp.status === 200) {
			const paymentsData = await paymentsResp.json();
			if (paymentsData.data) {
				stats.pendingInvoices = paymentsData.data.totalSuccessInvoicesSum || 0;
			}
		}

		// Fetch esim stats
		const esimResp = await authFetch('/admin/esim/stats');
		if (esimResp.status === 200) {
			const esimData = await esimResp.json();
			if (esimData.data) {
				stats.esimOrders = esimData.data.totalEsimOrders || 0;
				stats.esimSumWithMarkup = esimData.data.totalSumWithMarkup || 0;
				stats.esimSumWithoutMarkup = esimData.data.totalSumWithMarkup - esimData.data.totalSumWithoutMarkup || 0;
			}
		}

		isLoading = false;
	});
</script>

<div class="container mx-auto px-4 py-8 space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
			<p class="text-muted-foreground">Monitor your platform's key metrics and performance</p>
		</div>
		<Badge variant="outline" class="flex items-center gap-2">
			<TrendingUpIcon class="h-3 w-3" />
			Live Data
		</Badge>
	</div>

	<!-- Stats Cards -->
	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Total Users</CardTitle>
				<UsersIcon class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				{#if isLoading}
					<div class="h-7 w-16 bg-muted animate-pulse rounded"></div>
				{:else}
					<div class="text-2xl font-bold">{stats.totalUsers.toLocaleString()}</div>
				{/if}
				<p class="text-xs text-muted-foreground">Registered accounts</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Total Packeges Sold</CardTitle>
				<CreditCardIcon class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				{#if isLoading}
					<div class="h-7 w-16 bg-muted animate-pulse rounded"></div>
				{:else}
					<div class="text-2xl font-bold">{stats.totalTransactions.toLocaleString()}</div>
				{/if}
				<p class="text-xs text-muted-foreground">Total of company products users bought</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Total Invested</CardTitle>
				<DollarSignIcon class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				{#if isLoading}
					<div class="h-7 w-20 bg-muted animate-pulse rounded"></div>
				{:else}
					<div class="text-2xl font-bold">${stats.totalRevenue.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })}</div>
				{/if}
				<p class="text-xs text-muted-foreground">Total sum invested in company products</p>
			</CardContent>
		</Card>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Total Success Deposits</CardTitle>
				<FileTextIcon class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				{#if isLoading}
					<div class="h-7 w-16 bg-muted animate-pulse rounded"></div>
				{:else}
					<div class="text-2xl font-bold">${stats.pendingInvoices.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })}</div>
				{/if}
				<p class="text-xs text-muted-foreground">Total sum of success deposits</p>
			</CardContent>
		</Card>

	</div>

	<!-- Hotwallet and Pending Withdrawals Cards -->
	<div class="grid gap-4 md:grid-cols-2">
		<HotwalletBalanceCard />
		<PendingWithdrawalsCard {isLoading} pendingWithdrawals={stats.pendingWithdrawals} />
	</div>

	<!-- eSIM Stats Card -->
	<Card>
		<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
			<CardTitle class="text-sm font-medium">eSIM Orders</CardTitle>
			<CreditCardIcon class="h-4 w-4 text-muted-foreground" />
		</CardHeader>
		<CardContent>
			{#if isLoading}
				<div class="space-y-2">
					<div class="h-7 w-16 bg-muted animate-pulse rounded"></div>
					<div class="h-5 w-24 bg-muted animate-pulse rounded"></div>
					<div class="h-5 w-24 bg-muted animate-pulse rounded"></div>
				</div>
			{:else}
				<div class="space-y-2">
					<div class="text-2xl font-bold">{stats.esimOrders}</div>
					<p class="text-xs text-muted-foreground mt-2">Total of esim orders</p>
					<div class="text-sm text-muted-foreground">
						<span class="font-medium">Total amount with markup:</span> ${stats.esimSumWithMarkup.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })}
					</div>
					<div class="text-sm text-muted-foreground">
						<span class="font-medium">Total markup:</span> ${stats.esimSumWithoutMarkup.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })}
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	<div class="mb-6">
		<TokenPriceChart />
	</div>
</div>
