export const ssr = false;
export const prerender = true;

import type { LayoutServerLoad } from './$types';
import { loadTranslations, translations, defaultLocale } from '$lib/locale';

export const load: LayoutServerLoad = async ({ url, cookies }) => {
	const { pathname } = url;

	const initLocale = cookies.get('locale') || defaultLocale;

	if (!cookies.get('locale')) {
		cookies.set('locale', initLocale, {
			maxAge: 60 * 60 * 24 * 365,
			path: '/',
			domain: `.${import.meta.env.VITE_HOSTNAME}`,
			secure: false,
			httpOnly: false,
		});
	}

	await loadTranslations(initLocale, pathname);

	return {
		i18n: { locale: initLocale, route: pathname },
		translations: translations.get(),
	};
};
