<script lang="ts" context="module">
	import { getContext, setContext } from 'svelte';

	const RESPONSIVE_DIALOG_CTX = Symbol('responsive-dialog');

	export function getResponsiveDialogContext() {
		return getContext<{ isMobile: boolean }>(RESPONSIVE_DIALOG_CTX);
	}

	export function setResponsiveDialogContext(isMobile: boolean) {
		setContext(RESPONSIVE_DIALOG_CTX, { isMobile });
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import { IsMobile } from '$lib/hooks/is-mobile.svelte.js';

	let {
		open = $bindable(false),
		onOpenChange,
		...restProps
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		[key: string]: any;
	} = $props();

	const isMobile = new IsMobile();

	// Set context for child components to know if we're on mobile
	setResponsiveDialogContext(isMobile.current);

	// Update context reactively when screen size changes
	$effect(() => {
		setResponsiveDialogContext(isMobile.current);
	});
</script>

{#if isMobile.current}
	<DrawerPrimitive.Root bind:open {onOpenChange} {...restProps} />
{:else}
	<DialogPrimitive.Root bind:open {onOpenChange} {...restProps} />
{/if}

