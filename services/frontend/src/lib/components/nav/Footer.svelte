<script lang="ts">
	import { t } from '$lib/locale';

	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';

	import Locale from '$components/stuff/Locale.svelte';

	let { data } = $props();
</script>

<footer
class="flex w-full flex-col items-center justify-between gap-8 px-4 py-4 md:flex-row md:px-8"
>
<p class="order-last text-neutral-400 md:order-none">Copyright © 2025 Company Name</p>
<nav class="flex flex-col items-center gap-6 font-bold md:flex-row">
	<a href={`/Privacy_Policy.pdf`} target="_blank" class="inline-flex gap-2 hover:underline"
		>{$t('common.footer-privacy')} <ArrowUpRightIcon /></a
	>

	<a href={`/Terms_of_Use.pdf`} target="_blank" class="inline-flex gap-2 hover:underline"
		>{$t('common.footer-terms')} <ArrowUpRightIcon /></a
	>

	<a href={`/Cookie_Policy.pdf`} target="_blank" class="inline-flex gap-2 hover:underline"
		>{$t('common.footer-cookies')} <ArrowUpRightIcon /></a
	>
</nav>
<div class="flex gap-2">
	<Locale />
</div>
</footer>
