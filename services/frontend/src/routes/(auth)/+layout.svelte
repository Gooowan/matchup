<script lang="ts">
	import Footer from '$components/nav/Footer.svelte';
	import BasicHeader from '$components/nav/BasicHeader.svelte';

	let { data, children } = $props();
</script>

<div class="grid min-h-screen grid-rows-[auto_1fr_auto]">
	<BasicHeader logoHref="/login" showActiveStates={true} />

	<main class="my-16 px-4 md:px-8">
		{@render children?.()}
	</main>

	<Footer {data} />
</div>
