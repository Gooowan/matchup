<script lang="ts">
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';

	import GlobeIcon from '@lucide/svelte/icons/globe';

	import { locale, setLocale } from '$lib/locale';
	import lang from '$lib/locale/lang.json';
	import { authFetch } from '$utils/authFetch';
	import { authStore } from '$stores/auth.svelte';

	const langToCountry: Record<string, string> = {
		en: 'gb',
		uk: 'ua',
		es: 'es',
	};

	function getFlagFromLang(langCode: string) {
		const countryCode = langToCountry[langCode.toLowerCase()];
		if (!countryCode) return null;

		return countryCode
			.toUpperCase()
			.split('')
			.map((char) => String.fromCodePoint(0x1f1e6 + char.charCodeAt(0) - 65))
			.join('');
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger disabled class={buttonVariants({ variant: 'outline' })}>
		<GlobeIcon class="h-[1.2rem] w-[1.2rem]" />
		{/* @ts-ignore */ null}
		<span class="text-sm font-normal">{lang[$locale] ?? $locale}</span>
		<span class="sr-only">Change language</span>
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end">
		{#each Object.entries(lang) as [key, value]}
			<DropdownMenu.Item>
				<button
					type="button"
					class="w-full text-left"
					onclick={async () => {
						setLocale(key);
						if (authStore.isAuthenticated) {
							const response = await authFetch(`/user/locale`, {
								method: 'POST',
								body: JSON.stringify({ locale: key }),
							});
							if (!response.ok) {
								console.error('Failed to set locale:', response.status);
							}
						} else {
							document.cookie = `locale=${key}; max-age=${60 * 60 * 24 * 365}; path=/; domain=.${window.location.hostname}; secure=false; httpOnly=false; SameSite=Lax`;
						}
					}}
				>
					{getFlagFromLang(key)}
					{value}
				</button>
			</DropdownMenu.Item>
		{/each}
	</DropdownMenu.Content>
</DropdownMenu.Root>
