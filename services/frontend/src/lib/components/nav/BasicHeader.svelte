<script lang="ts">
	import { page } from '$app/state';
	import { t } from '$lib/locale';
	import Logotype from '$lib/assets/svg/logotype.svg?component';
	import Button from '$components/ui/button/button.svelte';

	interface Props {
		logoHref?: string;
		showActiveStates?: boolean;
	}

	let { logoHref = '/', showActiveStates = false }: Props = $props();

	let currentRoute = $derived(showActiveStates ? page.url.pathname || '/' : '');
</script>

<header class="flex w-full items-center justify-between gap-4 px-4 py-4 md:px-8">
	<a href={logoHref}>
		<Logotype class="h-8 w-auto fill-black dark:fill-current" />
	</a>

	<nav class="flex items-center gap-1 font-bold md:gap-4">
		<Button
			href="/login"
			variant={showActiveStates && currentRoute == '/login' ? 'secondary' : 'ghost'}
			>{$t('common.login-button')}</Button
		>
		<Button
			href="/register"
			variant={showActiveStates && currentRoute == '/register' ? 'secondary' : 'ghost'}
			>{$t('common.register-button')}</Button
		>
	</nav>
</header>
