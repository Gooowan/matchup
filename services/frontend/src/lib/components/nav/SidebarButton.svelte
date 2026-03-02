<script lang="ts">
	import { Button, type ButtonVariant } from '$components/ui/button/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/context.svelte.js';
	import { IsMobile } from '$lib/hooks/is-mobile.svelte.js';
	import { cn } from '$lib/utils.js';

	let { 
		href = undefined, 
		variant = 'ghost' as ButtonVariant, 
		isFullWidth = true,
		children, 
		onclick = undefined,
		...restProps 
	}: {
		href?: string;
		variant?: ButtonVariant;
		isFullWidth?: boolean;
		children?: any;
		onclick?: (e: MouseEvent) => void;
		[key: string]: any;
	} = $props();

	const sidebar = useSidebar();
	const isMobile = new IsMobile();

	function handleClick(e: MouseEvent) {
		if (onclick) {
			onclick(e);
		}
		if (isMobile.current) {
			sidebar.toggle();
		}
	}
</script>

<Button
	{...restProps}
	{variant}
	class={cn(isFullWidth ? "w-full justify-start p-4 group-data-[collapsible=icon]:justify-center" : restProps.class, 'h-12 text-lg gap-4 !px-4',restProps.class)}
	{href}
	onclick={handleClick}
>
	{@render children?.()}
</Button>
