<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import { cn } from '$lib/utils.js';
	import { getResponsiveDialogContext } from './responsive-dialog.svelte';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: DialogPrimitive.TitleProps = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

{#if isMobile}
	<DrawerPrimitive.Title
		bind:ref
		class={cn('text-lg font-semibold leading-none tracking-tight', className)}
		{...restProps}
	>
		{@render children?.()}
	</DrawerPrimitive.Title>
{:else}
	<DialogPrimitive.Title
		bind:ref
		class={cn('text-lg font-semibold leading-none tracking-tight', className)}
		{...restProps}
	>
		{@render children?.()}
	</DialogPrimitive.Title>
{/if}
