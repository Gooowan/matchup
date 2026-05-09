<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { tick } from 'svelte';
	import { t } from '$lib/locale';

	interface Props {
		open?: boolean;
		onclose?: () => void;
		children?: import('svelte').Snippet;
		overlay?: boolean;
	}

	let { open = false, onclose, children, overlay = true }: Props = $props();

	let dragY = $state(0);
	let dragging = $state(false);
	let startY = 0;

	function handleOverlayClick() {
		onclose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose?.();
	}

	function onDragStart(e: PointerEvent) {
		dragging = true;
		startY = e.clientY;
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
	}

	function onDragMove(e: PointerEvent) {
		if (!dragging) return;
		const dy = e.clientY - startY;
		dragY = Math.max(0, dy);
	}

	async function onDragEnd() {
		if (!dragging) return;
		dragging = false;
		if (dragY > 80) {
			onclose?.();
			await tick();
		}
		dragY = 0;
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
			aria-label={$t('common.close')}
		></div>
	{/if}

	<!-- Sheet -->
	<div
		class="fixed right-0 bottom-0 left-0 z-[60] rounded-tl-[20px] rounded-tr-[20px] mu-sheet"
		style="
			padding: 8px 16px calc(max(env(safe-area-inset-bottom), 8px) + 80px);
			transform: translateY({dragY}px);
			transition: {dragging ? 'none' : 'transform 200ms ease'};
		"
		transition:fly={{ y: 400, duration: 300 }}
	>
		<!-- Drag handle — touch target -->
		<div
			class="relative -mx-4 flex cursor-grab items-center justify-center py-3"
			style="touch-action: none;"
			onpointerdown={onDragStart}
			onpointermove={onDragMove}
			onpointerup={onDragEnd}
			onpointercancel={onDragEnd}
			role="button"
			tabindex="-1"
			aria-label={$t('common.drag_to_close')}
		>
			<div
				class="h-[5px] w-[80px] rounded-full transition-colors"
				style="background: {dragging ? '#8984da' : 'var(--mu-handle)'};"
			></div>
		</div>

		{@render children?.()}
	</div>
{/if}
