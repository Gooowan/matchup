<script lang="ts">
	import BottomSheet from './BottomSheet.svelte';
	import { t } from '$lib/locale';

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
	<!-- Group 2: Hide + Report -->
	<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 16px;">
		<button
			class="flex w-full items-center gap-4"
			onclick={() => { onhide?.(); onclose?.(); }}
		>
			<i class="fi fi-rr-eye-crossed mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
			<span class="mu-text-primary text-[14px] font-semibold">{$t('feed.action_hide')}</span>
		</button>
		<button
			class="flex w-full items-center gap-4"
			onclick={() => { onreport?.(); onclose?.(); }}
		>
			<i class="fi fi-rr-shield-exclamation mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
			<span class="mu-text-primary text-[14px] font-semibold">{$t('feed.action_report')}</span>
		</button>
	</div>
</BottomSheet>
