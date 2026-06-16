export const ssr = false;
export const prerender = true;

import { building } from '$app/environment';
import type { LayoutServerLoad } from './$types';
import { loadTranslations, translations, defaultLocale } from '$lib/locale';

/** Normalize VITE_HOSTNAME for Set-Cookie domain (hostname only, no protocol). */
function cookieDomain(raw: string | undefined): string | undefined {
	if (!raw) return undefined;
	const host = raw.replace(/^https?:\/\//, '').split('/')[0]?.split(':')[0]?.trim();
	if (!host || host === 'localhost' || !host.includes('.')) return undefined;
	return host.startsWith('.') ? host : `.${host}`;
}

export const load: LayoutServerLoad = async ({ url, cookies }) => {
	const { pathname } = url;

	const initLocale = cookies.get('locale') || defaultLocale;

	// Skip during prerender/build: localhost cannot set cookies for production domains.
	// Locale is restored client-side in +layout.svelte via localStorage.
	if (!building && !cookies.get('locale')) {
		const cookieOptions = {
			maxAge: 60 * 60 * 24 * 365,
			path: '/',
			secure: false,
			httpOnly: false,
		} as const;

		const domain = cookieDomain(import.meta.env.VITE_HOSTNAME);
		if (domain) {
			cookies.set('locale', initLocale, { ...cookieOptions, domain });
		} else {
			cookies.set('locale', initLocale, cookieOptions);
		}
	}

	await loadTranslations(initLocale, pathname);

	return {
		i18n: { locale: initLocale, route: pathname },
		translations: translations.get(),
	};
};
