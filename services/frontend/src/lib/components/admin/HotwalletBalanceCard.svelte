<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import WalletIcon from '@lucide/svelte/icons/wallet';
	import { adminWithdrawsStore } from '$lib/stores/admin_withdraws.svelte';

	function formatCurrency(amount: number, decimals: number = 2): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: decimals,
		}).format(Math.abs(amount));
	}
</script>

<Card>
	<CardHeader>
		<CardTitle class="flex items-center gap-2">
			<WalletIcon class="size-5" />
			Hotwallet Balance
		</CardTitle>
	</CardHeader>
	<CardContent>
		{#if adminWithdrawsStore.isBalanceLoading}
			<div class="grid gap-4 md:grid-cols-3">
				{#each Array(3) as _}
					<div class="space-y-2">
						<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
						<div class="h-8 w-32 animate-pulse rounded bg-gray-950/30"></div>
					</div>
				{/each}
			</div>
		{:else if adminWithdrawsStore.hotwalletBalance}
			<div class="grid gap-4 md:grid-cols-3">
				<div class="space-y-2">
					<p class="text-muted-foreground text-sm">BNB</p>
					<p class="text-2xl font-bold">
						{formatCurrency(adminWithdrawsStore.hotwalletBalance.bnb, 6)}
					</p>
				</div>
				<div class="space-y-2">
					<p class="text-muted-foreground text-sm">USDT</p>
					<p class="text-2xl font-bold">
						{formatCurrency(adminWithdrawsStore.hotwalletBalance.usdt, 2)}
					</p>
				</div>
				<div class="space-y-2">
					<p class="text-muted-foreground text-sm">USDC</p>
					<p class="text-2xl font-bold">
						{formatCurrency(adminWithdrawsStore.hotwalletBalance.usdc, 2)}
					</p>
				</div>
			</div>
		{:else}
			<p class="text-muted-foreground">Failed to load hotwallet balance</p>
		{/if}
	</CardContent>
</Card>

