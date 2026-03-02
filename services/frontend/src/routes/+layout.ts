import type { LayoutLoad } from './$types';
import { addTranslations, setLocale, setRoute } from '$lib/locale';

export const load: LayoutLoad = async ({ data }) => {
	const { i18n, translations } = data;
	const { locale, route } = i18n;

	addTranslations(translations);

	await setRoute(route);
	await setLocale(locale);

	return { i18n };
};
