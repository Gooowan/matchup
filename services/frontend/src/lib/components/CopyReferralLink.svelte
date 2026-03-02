<script lang="ts">
	import Button from '$lib/components/ui/button/button.svelte';
	import LinkIcon from '@lucide/svelte/icons/link';
	import Check from '@lucide/svelte/icons/check';
	import { getReferralLink } from '$lib/utils/referralLink';
	import { cn } from '$lib/utils';
	import toast from 'svelte-french-toast';

	interface Props {
		referralId: number;
		maxLength?: number;
		class?: string;
	}

	let { referralId, maxLength = 18, class: className }: Props = $props();

	let copied = $state(false);
	let referralLink = $derived(getReferralLink(referralId));

	// Remove protocol and truncate middle, keeping last 4 chars
	let displayLink = $derived.by(() => {
		const linkWithoutProtocol = referralLink.replace(/^https?:\/\//, '');

		if (linkWithoutProtocol.length <= maxLength) {
			return linkWithoutProtocol;
		}

		// Always show last 4 chars (the referral ID digits)
		const endPart = linkWithoutProtocol.slice(-4);
		const startPart = linkWithoutProtocol.slice(0, maxLength - 7); // -7 for "..." and last 4 chars

		return `${startPart}...${endPart}`;
	});

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(referralLink);
			copied = true;
			toast.success('Referral link copied to clipboard');
			setTimeout(() => {
				copied = false;
			}, 2000);
		} catch (err) {
			toast.error('Failed to copy link');
		}
	}
</script>

<Button variant="outline" onclick={copyToClipboard} class={cn('max-w-xs', className)}>
	{#if copied}
		<Check class="h-4 w-4 flex-shrink-0" />
		<span class="font-mono text-xs" style="width: {displayLink.length}ch;">Copied!</span>
	{:else}
		<LinkIcon class="h-4 w-4 flex-shrink-0" />
		<span class="font-mono text-xs" style="width: {displayLink.length}ch;">{displayLink}</span>
	{/if}
</Button>
