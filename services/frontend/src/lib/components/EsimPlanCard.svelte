<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		plan: EsimPlan;
		onPurchase: (plan: EsimPlan) => void;
	}

	let { plan, onPurchase }: Props = $props();

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function formatDuration(duration: number, unit: string): string {
		if (unit === 'DAY') {
			return `${duration} ${duration === 1 ? 'day' : 'days'}`;
		}
		return `${duration} ${unit}`;
	}
</script>

<Card class="transition-shadow hover:shadow-md">
	<CardHeader>
		<CardTitle class="flex items-center justify-between text-lg gap-2">
			<span>{plan.name}</span>
			<Badge variant="default">Available</Badge>
		</CardTitle>
		<CardDescription class="text-sm">
			{plan.description || plan.locationCode}
		</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
		<div class="space-y-2">
			<div class="flex items-center justify-between">
				<span class="text-sm font-medium text-muted-foreground">Price:</span>
				<span class="text-xl font-bold text-foreground">{formatCurrency(plan.price)}</span>
			</div>

			<div class="flex items-center justify-between">
				<span class="text-sm font-medium text-muted-foreground">Data:</span>
				<Badge variant="outline" class="font-semibold">
					{plan.data}
				</Badge>
			</div>

			<div class="flex items-center justify-between">
				<span class="text-sm font-medium text-muted-foreground">Duration:</span>
				<span class="text-sm font-medium">{formatDuration(plan.duration, plan.durationUnit)}</span>
			</div>

			<div class="flex items-center justify-between">
				<span class="text-sm font-medium text-muted-foreground">Speed:</span>
				<span class="text-sm font-medium">{plan.speed}</span>
			</div>
		</div>

		<Button onclick={() => onPurchase(plan)} class="w-full">Purchase</Button>
	</CardContent>
</Card>
