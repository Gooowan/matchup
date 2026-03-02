<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import { getResponsiveDialogContext } from './responsive-dialog.svelte';

	let {
		ref = $bindable(null),
		children,
		...restProps
	}: DialogPrimitive.CloseProps = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

{#if isMobile}
	<DrawerPrimitive.Close bind:ref {...restProps}>
		{@render children?.()}
	</DrawerPrimitive.Close>
{:else}
	<DialogPrimitive.Close bind:ref {...restProps}>
		{@render children?.()}
	</DialogPrimitive.Close>
{/if}

