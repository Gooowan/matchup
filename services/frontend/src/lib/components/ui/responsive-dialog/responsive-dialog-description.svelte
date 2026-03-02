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
	}: DialogPrimitive.DescriptionProps = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

{#if isMobile}
	<DrawerPrimitive.Description
		bind:ref
		class={cn('text-muted-foreground text-sm', className)}
		{...restProps}
	>
		{@render children?.()}
	</DrawerPrimitive.Description>
{:else}
	<DialogPrimitive.Description
		bind:ref
		class={cn('text-muted-foreground text-sm', className)}
		{...restProps}
	>
		{@render children?.()}
	</DialogPrimitive.Description>
{/if}

