<script lang="ts">
	import { page } from '$app/state';

	import { Toaster } from 'svelte-french-toast';

	import '../app.css';
	import { t } from '$lib/locale';

	let { data, children } = $props();

	let currentRoute = $derived(page.url.pathname || '/');

	let currentPage = $derived(
		currentRoute.split('/').pop() == '' ? 'landing' : currentRoute.split('/').pop()
	);
</script>

<svelte:head>
	<title>{$t(`common.title-${currentPage}`)} | Desim</title>
</svelte:head>

<div
	class="absolute top-0 right-0 left-0 -z-1 h-dvh bg-cover bg-no-repeat md:bg-contain"
	style="
		background-image: url('/bg.png');
		transform: translateZ(0); 
		will-change: transform
	"
></div>

<Toaster
	toastOptions={{
		style:
			'border-radius: 0.75rem; backdrop-filter: blur(12px); border: 1px solid rgba(59, 76, 107, 0.3); background: linear-gradient(to bottom right, rgba(0, 0, 0, 0.4), rgba(0, 0, 0, 0.7)); box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05); color: oklch(0.985 0 0); font-size: 0.875rem; padding: 1rem;'
	}}
/>
{@render children?.()}
