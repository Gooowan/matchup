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
	import { Badge } from '$lib/components/ui/badge';
	import toast from 'svelte-french-toast';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		order: EsimOrder | null;
	}

	let { isOpen = $bindable(), onClose, order }: Props = $props();

	function copyToClipboard(text: string, label: string) {
		navigator.clipboard.writeText(text).then(() => {
			toast.success(`${label} copied to clipboard!`);
		});
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
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

{#if order}
	<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={onClose}>
		<ResponsiveDialog.Content class="max-w-2xl">
			<ResponsiveDialog.Header>
				<ResponsiveDialog.Title>eSIM Details</ResponsiveDialog.Title>
				<ResponsiveDialog.Description>
					Order #{order.id} - {order.packageName || order.packageCode}
				</ResponsiveDialog.Description>
			</ResponsiveDialog.Header>

			<div class="space-y-4">
				<!-- Status Card -->
				<Card>
					<CardHeader>
						<div class="flex items-center justify-between">
							<CardTitle class="text-base">Order Status</CardTitle>
							<Badge variant={getStatusBadgeVariant(order.status)}>
								{order.status.toUpperCase()}
							</Badge>
						</div>
					</CardHeader>
					<CardContent class="space-y-2 text-sm">
						<div class="flex justify-between">
							<span class="text-muted-foreground">Order ID:</span>
							<span class="font-medium">#{order.id}</span>
						</div>
						{#if order.orderNo}
							<div class="flex justify-between">
								<span class="text-muted-foreground">Order Number:</span>
								<span class="font-medium">{order.orderNo}</span>
							</div>
						{/if}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Amount Paid:</span>
							<span class="font-medium">{formatCurrency(order.amount)}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-muted-foreground">Purchase Date:</span>
							<span class="font-medium">{formatDate(order.createdAt)}</span>
						</div>
					</CardContent>
				</Card>

				{#if order.status === 'pending'}
					<!-- Pending State -->
					<Card>
						<CardContent class="py-6 text-center">
							<div class="mb-4 flex justify-center">
								<div
									class="border-primary h-12 w-12 animate-spin rounded-full border-4 border-t-transparent"
								></div>
							</div>
							<p class="font-medium">Processing your eSIM order...</p>
							<p class="text-muted-foreground mt-2 text-sm">
								You will receive activation details shortly. This usually takes a few minutes.
							</p>
						</CardContent>
					</Card>
				{:else if order.status === 'completed'}
					<!-- eSIM Details -->
					{#if order.qrCode}
						<Card>
							<CardHeader>
								<CardTitle class="text-base">QR Code</CardTitle>
								<CardDescription>Scan this QR code to activate your eSIM</CardDescription>
							</CardHeader>
							<CardContent class="flex flex-col items-center">
								<div class="rounded-lg bg-white p-4">
									<img
										src={order.qrCode}
										alt="eSIM QR Code"
										class="h-64 w-64 object-contain"
										onerror={(e) => {
											e.currentTarget.style.display = 'none';
											e.currentTarget.nextElementSibling.style.display = 'flex';
										}}
									/>
									<div class="hidden h-64 w-64 items-center justify-center border-2 border-dashed">
										<p class="text-muted-foreground text-center text-sm">Failed to load QR code</p>
									</div>
								</div>
								<Button
									variant="outline"
									size="sm"
									class="mt-4"
									onclick={() => copyToClipboard(order.qrCode!, 'QR Code URL')}
								>
									Copy QR Code Data
								</Button>
							</CardContent>
						</Card>
					{/if}

					{#if order.iccid || order.activationCode}
						<Card>
							<CardHeader>
								<CardTitle class="text-base">Activation Details</CardTitle>
								<CardDescription>Use these details to activate your eSIM manually</CardDescription>
							</CardHeader>
							<CardContent class="space-y-3">
								{#if order.iccid}
									<div>
										<label class="text-muted-foreground text-sm font-medium">ICCID:</label>
										<div class="mt-1 flex items-center gap-2">
											<code class="bg-muted flex-1 rounded px-3 py-2 text-sm">
												{order.iccid}
											</code>
											<Button
												variant="outline"
												size="sm"
												onclick={() => copyToClipboard(order.iccid!, 'ICCID')}
											>
												Copy
											</Button>
										</div>
									</div>
								{/if}

								{#if order.activationCode}
									<div>
										<label class="text-muted-foreground text-sm font-medium">
											Activation Code:
										</label>
										<div class="mt-1 flex items-center gap-2">
											<code class="bg-muted flex-1 break-all rounded px-3 py-2 text-sm">
												{order.activationCode}
											</code>
											<Button
												variant="outline"
												size="sm"
												onclick={() => copyToClipboard(order.activationCode!, 'Activation Code')}
											>
												Copy
											</Button>
										</div>
									</div>
								{/if}
							</CardContent>
						</Card>
					{/if}
				{:else if order.status === 'failed'}
					<!-- Failed State -->
					<Card>
						<CardContent class="py-6 text-center">
							<div class="mb-4 flex justify-center">
								<div
									class="bg-destructive/10 flex h-12 w-12 items-center justify-center rounded-full"
								>
									<span class="text-destructive text-2xl">✕</span>
								</div>
							</div>
							<p class="text-destructive font-medium">Order Failed</p>
							<p class="text-muted-foreground mt-2 text-sm">
								There was an issue processing your eSIM order. Please contact support.
							</p>
						</CardContent>
					</Card>
				{/if}
			</div>

			<ResponsiveDialog.Footer>
				<Button onclick={onClose}>Close</Button>
			</ResponsiveDialog.Footer>
		</ResponsiveDialog.Content>
	</ResponsiveDialog.Root>
{/if}
