<script lang="ts">
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import QRCodeStyling from 'qr-code-styling';
	import Button from '$lib/components/ui/button/button.svelte';
	import toast from 'svelte-french-toast';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		onSuccess: () => void;
		paymentAddress: string;
		amount: number;
		token?: string;
	}

	let {
		isOpen = $bindable(),
		onClose,
		paymentAddress,
		amount,
		onSuccess,
		token = $bindable('BNB'),
	}: Props = $props();

	let qrCodeData: HTMLElement;

	interface TokenConfig {
		decimals: number;
		symbol: string;
		contractAddress: string | null;
	}

	// Token configuration
	function getTokenConfig(): TokenConfig {
		switch (token) {
			case 'USDT':
				return {
					decimals: 6,
					symbol: 'USDT',
					contractAddress: '0x55d398326f99059fF775485246999027B3197955',
				};
			case 'USDC':
				return {
					decimals: 18,
					symbol: 'USDC',
					contractAddress: '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d',
				};
			default:
				return { decimals: 18, symbol: 'BNB', contractAddress: null };
		}
	}

	// Construct BSC Pay URL
	function getPaymentUrl(): string {
		if (token === 'BNB') {
			return `ethereum:${paymentAddress}@56?value=${(amount * Math.pow(10, 18)).toString()}&label=BSC%20Payment`;
		} else {
			const config = getTokenConfig();
			return `ethereum:${config.contractAddress}@56/transfer?address=${paymentAddress}&uint256=${(amount * Math.pow(10, config.decimals)).toString()}`;
		}
	}

	function handleClose() {
		onClose();
	}

	function copyAddress() {
		navigator.clipboard.writeText(paymentAddress);
		toast.success('Address copied to clipboard');
	}

	$effect(() => {
		if (isOpen && paymentAddress && amount && qrCodeData) {
			generateQR();
		}
	});

	function generateQR() {
		qrCodeData.innerHTML = '';

		const qr = new QRCodeStyling({
			type: 'svg',
			width: 256,
			height: 256,
			data: getPaymentUrl(),
			margin: 8,
			qrOptions: {
				typeNumber: 0,
				mode: 'Byte',
				errorCorrectionLevel: 'Q',
			},
			backgroundOptions: { color: 'white', round: 0.1 },
			dotsOptions: { type: 'extra-rounded', color: 'black' },
			cornersSquareOptions: {
				type: 'extra-rounded',
				color: 'black',
			},
			cornersDotOptions: { type: 'square', color: 'black' },
			// imageOptions: { hideBackgroundDots: true, imageSize: 0.15, margin: 8 },
			// image: `data:image/svg+xml;utf8,<svg fill="${encodeURIComponent('black')}" height="16" viewBox="0 0 16 14" width="16" xmlns="http://www.w3.org/2000/svg"><path d="M8 0C3.6 0 0 3.6 0 8s3.6 8 8 8 8-3.6 8-8-3.6-8-8-8zm0 14c-3.3 0-6-2.7-6-6s2.7-6 6-6 6 2.7 6 6-2.7 6-6 6z"/><circle cx="8" cy="8" r="2"/></svg>`,
		});

		qr.append(qrCodeData);
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
	<ResponsiveDialog.Content class="max-w-md">
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Transfer {getTokenConfig().symbol} (BSC)</ResponsiveDialog.Title>
		</ResponsiveDialog.Header>
		<div class="space-y-4">
			<!-- QR Code -->
			<div class="flex justify-center">
				<div bind:this={qrCodeData}></div>
			</div>

			<!-- Amount to Send -->
			<div class="bg-muted/50 rounded-lg border p-4">
				<p class="text-muted-foreground mb-1 text-sm font-medium">Amount to Send</p>
				<p class="text-2xl font-bold">{amount} {getTokenConfig().symbol}</p>
			</div>

			<!-- Payment Address -->
			<div class="bg-muted/50 rounded-lg border p-4">
				<p class="text-muted-foreground mb-2 text-sm font-medium">Payment Address</p>
				<div class="flex items-center gap-2">
					<code class="bg-background flex-1 break-all rounded px-2 py-1 text-xs">
						{paymentAddress}
					</code>
					<Button size="sm" variant="outline" onclick={copyAddress} class="shrink-0">
						<svg
							class="h-4 w-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							xmlns="http://www.w3.org/2000/svg"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
							></path>
						</svg>
					</Button>
				</div>
			</div>

			<!-- Network Info -->
			<div class="rounded-lg border bg-yellow-50 p-4 dark:bg-yellow-900/20">
				<p class="mb-1 text-sm font-medium text-yellow-800 dark:text-yellow-200">Network</p>
				<p class="text-sm text-yellow-700 dark:text-yellow-300">BSC (Binance Smart Chain)</p>
				<p class="mt-1 text-xs text-yellow-600 dark:text-yellow-400">Chain ID: 56</p>
			</div>

			<!-- Instructions -->
			<div class="space-y-2">
				{#if token === 'BNB'}
					<p class="text-muted-foreground text-center text-sm">
						Scan QR code with your BSC wallet, or
					</p>
					<Button href={getPaymentUrl()} class="w-full">Open in wallet</Button>
				{:else}
					<p class="text-muted-foreground text-center text-sm">
						Send exactly {amount}
						{getTokenConfig().symbol} to the address above using your BSC wallet
					</p>
					<div class="rounded-lg border bg-blue-50 p-3 dark:bg-blue-900/20">
						<p class="text-xs text-blue-800 dark:text-blue-200">
							Make sure to send {getTokenConfig().symbol} tokens, not BNB. The transaction will fail
							if you send the wrong token type.
						</p>
					</div>
				{/if}
			</div>

			<!-- Warning -->
			<div
				class="rounded-lg border border-orange-200 bg-orange-50 p-3 dark:border-orange-800 dark:bg-orange-900/20"
			>
				<p class="text-xs text-orange-800 dark:text-orange-200">
					Make sure you're connected to BSC network (Chain ID: 56) in your wallet
				</p>
			</div>
		</div>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
