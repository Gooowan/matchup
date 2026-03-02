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
	}

	let { isOpen = $bindable(), onClose, paymentAddress, amount, onSuccess }: Props = $props();

	let qrCodeData: HTMLElement;

	// Construct Solana Pay URL
	const paymentUrl = $derived(`solana:${paymentAddress}?amount=${amount}&label=Payment`);

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
			data: paymentUrl,
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
			imageOptions: { hideBackgroundDots: true, imageSize: 0.15, margin: 8 },
			image: `data:image/svg+xml;utf8,<svg fill="${encodeURIComponent('black')}" height="16" viewBox="0 0 16 14" width="16" xmlns="http://www.w3.org/2000/svg"><path d="m15.9176 11.038-2.6413 2.7713c-.0574.0602-.1269.1082-.2041.141s-.1604.0497-.2446.0497h-12.520966c-.059744 0-.118187-.0171-.168147-.0491-.0499596-.0321-.0892609-.0777-.1130745-.1313-.02381372-.0536-.03110249-.1129-.02097081-.1705.01013171-.0576.03724251-.111.07800141-.1538l2.6432769-2.7713c.05726-.06.12651-.1079.20346-.1407s.15996-.0498.2439-.05h12.52032c.0597 0 .1182.0171.1681.0492.05.032.0893.0776.1131.1313.0238.0536.0311.1128.021.1704-.0102.0576-.0373.1111-.078.1538zm-2.6413-5.58067c-.0574-.0602-.1269-.1082-.2041-.141s-.1604-.04971-.2446-.04966h-12.520966c-.059744 0-.118187.01708-.168147.04913-.0499596.03205-.0892609.07768-.1130745.13129-.02381372.0536-.03110249.11285-.02097081.17045.01013171.05761.03724251.11106.07800141.15379l2.6432769 2.77134c.05726.06004.12651.10794.20346.14073.07695.0328.15996.04979.2439.04993h12.52032c.0597 0 .1182-.01707.1681-.04913.05-.03205.0893-.07768.1131-.13129.0238-.0536.0311-.11285.021-.17045-.0102-.05761-.0373-.11106-.078-.15379zm-12.969666-1.99066h12.520966c.0842.00004.1674-.01687.2446-.04967s.1467-.0808.2041-.141l2.6413-2.771333c.0407-.042736.0678-.096189.078-.153792.0101-.057603.0028-.116847-.021-.170453s-.0631-.0992385-.1131-.1312911c-.0499-.0320526-.1084-.04912893-.1681-.0491309h-12.52032c-.08394.00013975-.16695.0171339-.2439.0499304s-.1462.0806976-.20346.1407366l-2.6425955 2.771333c-.0407196.04269-.0678184.09609-.07797306.15363-.01015467.05754-.00292373.11673.02080606.17031.0237297.05358.0629266.09922.1127835.13132.049857.03211.108207.04928.167893.04941z"/></svg>`,
		});

		qr.append(qrCodeData);
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen} onOpenChange={handleClose}>
	<ResponsiveDialog.Content class="max-w-md">
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Transfer Solana</ResponsiveDialog.Title>
		</ResponsiveDialog.Header>
		<div class="space-y-4">
			<!-- QR Code -->
			<div class="flex justify-center">
				<div bind:this={qrCodeData}></div>
			</div>

			<!-- Amount to Send -->
			<div class="bg-muted/50 rounded-lg border p-4">
				<p class="text-muted-foreground mb-1 text-sm font-medium">Amount to Send</p>
				<p class="text-2xl font-bold">{amount} SOL</p>
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

			<!-- Instructions -->
			<div class="space-y-2">
				<p class="text-muted-foreground text-center text-sm">
					Scan QR code with your Solana wallet, or
				</p>
				<Button href={paymentUrl} class="w-full">Open in wallet</Button>
			</div>
		</div>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
