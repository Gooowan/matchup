<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	let { children } = $props();

	onMount(() => {
		if (!browser) return;
		if (localStorage.getItem('mu-theme') === 'dark') {
			document.documentElement.classList.add('dark');
		}
	});
</script>

<svelte:head>
	<title>MatchUp</title>
</svelte:head>

{@render children?.()}
