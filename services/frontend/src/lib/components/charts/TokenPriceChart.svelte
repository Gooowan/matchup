<script lang="ts">
	import { LineChart } from 'layerchart';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { Card, CardContent, CardDescription, CardTitle } from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import { tokenSupplyHistoryStore, type TimeRange } from '$lib/stores/token_supply_history.svelte';
	import { curveMonotoneX as curve } from 'd3-shape';

	let selectedRange = $state<TimeRange>('last_24h');

	$effect(() => {
		tokenSupplyHistoryStore.fetchHistory(selectedRange);
	});

	const chartData = $derived.by(() => {
		if (!tokenSupplyHistoryStore.history || tokenSupplyHistoryStore.history.length === 0) {
			return [];
		}
		return tokenSupplyHistoryStore.history.map((point) => ({
			date: new Date(point.date),
			// Keep null values to create gaps in the chart
			price: point.price != null ? Number(point.price) : null,
		}));
	});

	function formatAxisDate(value: Date): string {
		if (!(value instanceof Date)) return String(value);
		
		switch (selectedRange) {
			case 'last_24h':
				return value.toLocaleString('en-US', {
					hour: '2-digit',
					minute: '2-digit',
					hour12: false,
				});
			case 'weekly':
				return value.toLocaleDateString('en-US', {
					weekday: 'short',
				});
			case 'all_time':
				return value.toLocaleDateString('en-US', {
					month: 'short',
					day: 'numeric',
				});
			default:
				return value.toLocaleDateString('en-US');
		}
	}

	function formatPrice(value: number): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 9,
		}).format(value);
	}

	const chartConfig = {
		price: {
			label: "Price",
			color: "#194BFB",
		},
	} satisfies Chart.ChartConfig;

	function tooltipLabelFormatter(value: any) {
		if (!value) return '';
		
		// Value should be the date from our data
		const date = value instanceof Date ? value : new Date(value);
		
		switch (selectedRange) {
			case 'last_24h':
				return date.toLocaleString('en-US', {
					month: 'short',
					day: 'numeric',
					hour: '2-digit',
					minute: '2-digit',
					hour12: false,
				});
			case 'weekly':
				return date.toLocaleDateString('en-US', {
					weekday: 'short',
					month: 'short',
					day: 'numeric',
				});
			case 'all_time':
				return date.toLocaleDateString('en-US', {
					month: 'short',
					day: 'numeric',
					year: 'numeric',
				});
			default:
				return date.toLocaleDateString('en-US');
		}
	}
</script>

<Card class="w-full">
	<CardContent class="">
		<Tabs.Root bind:value={selectedRange} class="w-full">
			<div class="flex flex-col md:flex-row justify-between md:items-center mb-4 gap-2">
				<p class="text-lg">
					COMPANY_TOKEN <span class="text-sm text-muted-foreground">/USD</span>	
				</p>
				<Tabs.List class="w-full md:w-auto">
					<Tabs.Trigger value="last_24h">24H</Tabs.Trigger>
					<Tabs.Trigger value="weekly">Week</Tabs.Trigger>
					<Tabs.Trigger value="all_time">All Time</Tabs.Trigger>
				</Tabs.List>
			</div>
			{#if tokenSupplyHistoryStore.isLoading}
				<div class="flex h-[350px] items-center justify-center">
					<div class="text-muted-foreground">Loading chart data...</div>
				</div>
			{:else if tokenSupplyHistoryStore.error}
				<div class="flex h-[350px] items-center justify-center">
					<div class="text-destructive">{tokenSupplyHistoryStore.error}</div>
				</div>
			{:else if chartData.length === 0}
				<div class="flex h-[350px] items-center justify-center">
					<div class="text-muted-foreground">No price history available yet</div>
				</div>
			{:else}
				<Chart.Container config={chartConfig} class="h-[350px] pl-8">
					<LineChart 
						data={chartData} 
						x="date" 
						y="price"
						yDomain={[0.01, null]}
						props={{
							xAxis: {
								placement: 'bottom',
								format: formatAxisDate,
							},
							yAxis: {
								placement: 'left',
								format: formatPrice,
							},
						spline: {
							strokeWidth: 2,
							curve: curve,
							defined: (d: any) => d.price != null, // Create gaps for null values
							stroke: '#194BFB',
						},
						}}
					>
						{#snippet tooltip()}
							<Chart.Tooltip 
								labelKey="date" 
								nameKey="price"
								labelFormatter={tooltipLabelFormatter}
							>
								{#snippet formatter({ value, name })}
									<div class="flex w-full flex-wrap items-center gap-2">
										<div class="flex flex-1 justify-between leading-none">
											<span class="font-mono font-medium tabular-nums text-foreground">
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
		</Tabs.Root>
	</CardContent>
</Card>


