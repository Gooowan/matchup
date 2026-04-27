<script lang="ts">
	interface Props {
		loading: boolean;
		error?: string;
		empty?: boolean;
		emptyIcon?: string;
		emptyText?: string;
		onRetry?: () => void;
		children?: import('svelte').Snippet;
	}

	let { loading, error, empty, emptyIcon = 'fi-rr-inbox', emptyText = 'Nothing here yet', onRetry, children }: Props = $props();
</script>

{#if loading}
	<div class="flex flex-1 items-center justify-center py-16">
		<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
	</div>
{:else if error}
	<div class="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
		<i class="fi fi-rr-exclamation" style="font-size: 40px; color: #aeb4bc;"></i>
		<p class="text-[15px] font-semibold mu-text-primary">Something went wrong</p>
		<p class="text-[13px] font-medium" style="color: #696969;">{error}</p>
		{#if onRetry}
			<button
				onclick={onRetry}
				class="mt-2 rounded-[50px] px-5 py-2 text-[13px] font-semibold text-white"
				style="background: #8984da;"
			>Try again</button>
		{/if}
	</div>
{:else if empty}
	<div class="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
		<i class="fi {emptyIcon}" style="font-size: 40px; color: #aeb4bc;"></i>
		<p class="text-[15px] font-semibold mu-text-primary">{emptyText}</p>
	</div>
{:else}
	{@render children?.()}
{/if}
