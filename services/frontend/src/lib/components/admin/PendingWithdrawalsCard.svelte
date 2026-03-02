<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import FileTextIcon from '@lucide/svelte/icons/file-text';

	interface Props {
		isLoading?: boolean;
		pendingWithdrawals: Record<string, number>;
	}

	let { isLoading = false, pendingWithdrawals }: Props = $props();
</script>

<Card>
	<CardHeader>
		<CardTitle class="flex items-center gap-2">
			<FileTextIcon class="size-5" />
			Pending Withdrawals
		</CardTitle>
	</CardHeader>
	<CardContent>
		{#if isLoading}
			<div class="grid gap-4 md:grid-cols-2">
				{#each Array(2) as _}
					<div class="space-y-2">
						<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
						<div class="h-8 w-32 animate-pulse rounded bg-gray-950/30"></div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="grid gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<p class="text-muted-foreground text-sm">USDC</p>
					<p class="text-2xl font-bold">
						${pendingWithdrawals.USDC?.toLocaleString(undefined, {
							minimumFractionDigits: 0,
							maximumFractionDigits: 2,
						}) || '0.00'}
					</p>
				</div>
				<div class="space-y-2">
					<p class="text-muted-foreground text-sm">USDT</p>
					<p class="text-2xl font-bold">
						${pendingWithdrawals.USDT?.toLocaleString(undefined, {
							minimumFractionDigits: 0,
							maximumFractionDigits: 2,
						}) || '0.00'}
					</p>
				</div>
			</div>
		{/if}
	</CardContent>
</Card>
