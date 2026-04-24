<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';

	interface Props {
		open?: boolean;
		profileId?: string | null;
		onclose?: () => void;
		onclear?: () => void;
		onhide?: () => void;
		onreport?: () => void;
	}

	let { open = false, profileId = null, onclose, onclear, onhide, onreport }: Props = $props();
</script>

<BottomSheet {open} {onclose}>
	<!-- Group 1: Clear the screen -->
	<div class="mu-card mb-2 rounded-[20px] p-4">
		<button
			class="flex w-full items-center gap-4"
			onclick={() => { onclear?.(); onclose?.(); }}
		>
			<i class="fi fi-rr-expand mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
			<span class="mu-text-primary text-[14px] font-semibold">Clear the screen</span>
		</button>
	</div>

	<!-- Group 2: Hide + Report -->
	<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 16px;">
		<button
			class="flex w-full items-center gap-4"
			onclick={() => { onhide?.(); onclose?.(); }}
		>
			<i class="fi fi-rr-eye-crossed mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
			<span class="mu-text-primary text-[14px] font-semibold">Hide</span>
		</button>
		<button
			class="flex w-full items-center gap-4"
			onclick={() => { onreport?.(); onclose?.(); }}
		>
			<i class="fi fi-rr-shield-exclamation mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
			<span class="mu-text-primary text-[14px] font-semibold">Report</span>
		</button>
	</div>
</BottomSheet>
