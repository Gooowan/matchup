<script lang="ts">
	let { children } = $props();
</script>

<div class="min-h-[100dvh]" style="background: #dae1eb; font-family: 'Inter', sans-serif;">
	{@render children?.()}
</div>
