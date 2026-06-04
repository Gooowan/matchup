<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import XIcon from '@lucide/svelte/icons/x';
	import type { Snippet } from 'svelte';
	import { getResponsiveDialogContext } from './responsive-dialog.svelte';
	import DrawerOverlay from '$lib/components/ui/drawer/drawer-overlay.svelte';
	import DialogOverlay from '$lib/components/ui/dialog/dialog-overlay.svelte';
	import { cn, type WithoutChildrenOrChild } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		portalProps,
		children,
		showCloseButton = true,
		...restProps
	}: WithoutChildrenOrChild<DialogPrimitive.ContentProps> & {
		portalProps?: DialogPrimitive.PortalProps | DrawerPrimitive.PortalProps;
		children: Snippet;
		showCloseButton?: boolean;
	} = $props();

	const context = getResponsiveDialogContext();
	const isMobile = $derived(context?.isMobile ?? false);
</script>

{#if isMobile}
	<DrawerPrimitive.Portal {...portalProps}>
		<DrawerOverlay />
		<DrawerPrimitive.Content
			bind:ref
			data-slot="drawer-content"
			class={cn(
				'group/drawer-content bg-linear-to-b from-[#1A082B] fixed z-50 flex h-auto flex-col to-black px-4',
				'data-[vaul-drawer-direction=top]:inset-x-0 data-[vaul-drawer-direction=top]:top-0 data-[vaul-drawer-direction=top]:mb-24 data-[vaul-drawer-direction=top]:max-h-[80dvh] data-[vaul-drawer-direction=top]:rounded-b-lg data-[vaul-drawer-direction=top]:border-b',
				'data-[vaul-drawer-direction=bottom]:inset-x-0 data-[vaul-drawer-direction=bottom]:bottom-0 data-[vaul-drawer-direction=bottom]:mt-24 data-[vaul-drawer-direction=bottom]:max-h-[80dvh] data-[vaul-drawer-direction=bottom]:rounded-t-lg data-[vaul-drawer-direction=bottom]:border-t',
				'data-[vaul-drawer-direction=right]:inset-y-0 data-[vaul-drawer-direction=right]:right-0 data-[vaul-drawer-direction=right]:h-svh data-[vaul-drawer-direction=right]:w-3/4 data-[vaul-drawer-direction=right]:border-l data-[vaul-drawer-direction=right]:sm:max-w-sm',
				'data-[vaul-drawer-direction=left]:inset-y-0 data-[vaul-drawer-direction=left]:left-0 data-[vaul-drawer-direction=left]:h-svh data-[vaul-drawer-direction=left]:w-3/4 data-[vaul-drawer-direction=left]:border-r data-[vaul-drawer-direction=left]:sm:max-w-sm',
				className
			)}
			{...restProps}
		>
			<div
				class="mx-auto my-2 hidden h-1 w-[100px] shrink-0 rounded-full bg-white group-data-[vaul-drawer-direction=bottom]/drawer-content:block"
			></div>
			<div data-vaul-no-drag class="overflow-y-auto flex-1 pb-6 pt-4">
				{@render children?.()}
			</div>
		</DrawerPrimitive.Content>
	</DrawerPrimitive.Portal>
{:else}
	<DialogPrimitive.Portal {...portalProps}>
		<DialogOverlay />
		<DialogPrimitive.Content
			bind:ref
			data-slot="dialog-content"
			class={cn(
				'bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed left-[50%] top-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg duration-200 sm:max-w-lg',
				className
			)}
			{...restProps}
		>
			{@render children?.()}
			{#if showCloseButton}
				<DialogPrimitive.Close
					class="ring-offset-background focus:ring-ring rounded-xs focus:outline-hidden absolute end-4 top-4 opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 disabled:pointer-events-none [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:shrink-0"
				>
					<XIcon />
					<span class="sr-only">Close</span>
				</DialogPrimitive.Close>
			{/if}
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
{/if}
