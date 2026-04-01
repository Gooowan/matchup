<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	interface Props {
		open?: boolean;
		onclose?: () => void;
		children?: import('svelte').Snippet;
		overlay?: boolean;
	}

	let { open = false, onclose, children, overlay = true }: Props = $props();

	function handleOverlayClick() {
		onclose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose?.();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	{#if overlay}
		<!-- Semi-transparent overlay -->
		<div
			class="fixed inset-0 z-[55] bg-black/30"
			onclick={handleOverlayClick}
			transition:fade={{ duration: 200 }}
			role="button"
			tabindex="-1"
			aria-label="Close"
		></div>
	{/if}

	<!-- Sheet -->
	<div
		class="fixed right-0 bottom-0 left-0 z-[60] rounded-tl-[20px] rounded-tr-[20px] mu-sheet"
		style="padding: 8px 16px calc(max(env(safe-area-inset-bottom), 8px) + 80px);"
		transition:fly={{ y: 400, duration: 300 }}
	>
		<!-- Drag handle -->
		<div
			class="mx-auto mb-4 h-[5px] w-[80px] rounded-full"
			style="background: var(--mu-handle);"
		></div>

		{@render children?.()}
	</div>
{/if}
