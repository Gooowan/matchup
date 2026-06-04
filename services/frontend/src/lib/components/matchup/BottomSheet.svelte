<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { tick } from 'svelte';
	import { t } from '$lib/locale';

	interface Props {
		open?: boolean;
		onclose?: () => void;
		children?: import('svelte').Snippet;
		footer?: import('svelte').Snippet;
		overlay?: boolean;
	}

	let { open = false, onclose, children, footer, overlay = true }: Props = $props();

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

	<!-- Sheet — flex column, capped at 90dvh so long forms stay scrollable -->
	<div
		class="fixed right-0 bottom-0 left-0 z-[60] flex flex-col rounded-tl-[20px] rounded-tr-[20px] mu-sheet"
		style="
			max-height: 90dvh;
			transform: translateY({dragY}px);
			transition: {dragging ? 'none' : 'transform 200ms ease'};
		"
		transition:fly={{ y: 400, duration: 300 }}
	>
		<!-- Drag handle — listeners attached here only to avoid conflicting with scroll -->
		<div
			class="flex-shrink-0 px-4 pt-2"
			style="touch-action: none;"
			onpointerdown={onDragStart}
			onpointermove={onDragMove}
			onpointerup={onDragEnd}
			onpointercancel={onDragEnd}
		>
			<div
				class="relative flex cursor-grab items-center justify-center py-2"
				role="button"
				tabindex="-1"
				aria-label={$t('common.drag_to_close')}
			>
				<div
					class="h-[5px] w-[80px] rounded-full transition-colors"
					style="background: {dragging ? '#8984da' : 'var(--mu-handle)'};"
				></div>
			</div>
		</div>

		<!-- Scrollable body — flex-1 min-h-0 so overflow-y-auto is bounded on short screens -->
		<div
			class="flex-1 min-h-0 overflow-y-auto px-4"
			style="
				padding-bottom: {footer ? '8px' : 'calc(max(env(safe-area-inset-bottom), 8px) + 80px)'};
				-webkit-overflow-scrolling: touch;
			"
		>
			{@render children?.()}
		</div>

		{#if footer}
			<!-- Sticky footer (outside scroll region — always visible) -->
			<div
				class="flex-shrink-0 px-4 pt-2"
				style="padding-bottom: calc(max(env(safe-area-inset-bottom), 8px) + 8px);"
			>
				{@render footer()}
			</div>
		{/if}
	</div>
{/if}
