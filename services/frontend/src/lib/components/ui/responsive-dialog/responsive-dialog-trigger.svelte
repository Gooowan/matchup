<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import { getResponsiveDialogContext } from './responsive-dialog.svelte';

	let {
		ref = $bindable(null),
		children,
		...restProps
	}: DialogPrimitive.TriggerProps = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

{#if isMobile}
	<DrawerPrimitive.Trigger bind:ref {...restProps}>
		{@render children?.()}
	</DrawerPrimitive.Trigger>
{:else}
	<DialogPrimitive.Trigger bind:ref {...restProps}>
		{@render children?.()}
	</DialogPrimitive.Trigger>
{/if}

