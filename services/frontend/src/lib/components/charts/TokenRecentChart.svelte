<script lang="ts">
	import { LineChart } from 'layerchart';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { tokenSupplyHistoryStore } from '$lib/stores/token_supply_history.svelte';
	import { curveMonotoneX as curve } from 'd3-shape';

	$effect(() => {
		tokenSupplyHistoryStore.fetchRecent();
	});

	const chartData = $derived.by(() => {
		if (!tokenSupplyHistoryStore.recent || tokenSupplyHistoryStore.recent.length === 0) {
			return [];
		}
		return tokenSupplyHistoryStore.recent.map((price, index) => ({
			index: index + 1,
			price: Number(price),
		}));
	});

	function formatPrice(value: number): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: 9,
		}).format(value);
	}

	function formatIndex(value: number): string {
		return '';
	}

	const chartConfig = {
		price: {
			label: 'Price',
			color: '#194BFB',
		},
	} satisfies Chart.ChartConfig;
</script>

<Card class="w-full">
	<CardContent class="">
		<div class="mb-4 flex items-center justify-between">
			<p class="text-lg">
				COMPANY_TOKEN <span class="text-muted-foreground text-sm">/USD</span>
			</p>

			{#if chartData.length > 0}
				<p class="text-lg">{new Intl.NumberFormat(undefined, {
					minimumFractionDigits: 0,
					maximumFractionDigits: 12,
				}).format(Number(chartData[chartData.length - 1].price))}</p>
			{:else}
				<p class="text-muted-foreground">No recent price data available yet</p>
			{/if}
		</div>
		{#if tokenSupplyHistoryStore.isLoadingRecent}
			<div class="flex h-[350px] items-center justify-center">
				<div class="text-muted-foreground">Loading chart data...</div>
			</div>
		{:else if tokenSupplyHistoryStore.recentError}
			<div class="flex h-[350px] items-center justify-center">
				<div class="text-destructive">{tokenSupplyHistoryStore.recentError}</div>
			</div>
		{:else if chartData.length === 0}
			<div class="flex h-[350px] items-center justify-center">
				<div class="text-muted-foreground">No recent price data available yet</div>
			</div>
		{:else}
			<Chart.Container config={chartConfig} class="h-[350px] pl-8">
				<LineChart
					data={chartData}
					x="index"
					y="price"
					yDomain={[0.01, null]}
					props={{
						xAxis: {
							placement: 'bottom',
							format: formatIndex,
						},
						yAxis: {
							placement: 'left',
							format: formatPrice,
						},
						spline: {
							strokeWidth: 2,
							curve: curve,
							stroke: '#194BFB',
						},
					}}
				>
					{#snippet tooltip()}
						<Chart.Tooltip labelKey="price" nameKey="price">
							{#snippet formatter({ value, name })}
								<div class="flex w-full flex-wrap items-center gap-2">
									<div class="flex flex-1 justify-between leading-none gap-2">
										<span class="text-foreground font-mono font-medium tabular-nums">
											{formatPrice(Number(value))}
										</span>
									</div>
								</div>
							{/snippet}
						</Chart.Tooltip>
					{/snippet}
				</LineChart>
			</Chart.Container>
		{/if}
	</CardContent>
</Card>
