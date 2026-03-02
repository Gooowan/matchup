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
	data-slot={isMobile ? 'drawer-footer' : 'dialog-footer'}
	class={cn(
		isMobile
			? 'mt-auto flex flex-col gap-2 p-4'
			: 'flex flex-col-reverse gap-2 sm:flex-row sm:justify-end',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>

