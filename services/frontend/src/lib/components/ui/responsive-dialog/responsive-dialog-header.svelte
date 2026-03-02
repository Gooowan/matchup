<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import { getResponsiveDialogContext } from './responsive-dialog.svelte';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

<div
	bind:this={ref}
	data-slot={isMobile ? 'drawer-header' : 'dialog-header'}
	class={cn(
		isMobile ? 'mb-4 flex flex-col gap-1.5' : 'flex flex-col gap-2 text-center sm:text-left',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
