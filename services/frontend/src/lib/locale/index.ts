import i18n from 'sveltekit-i18n';
import lang from './lang.json';

export const defaultLocale = 'en';

const config = {
	translations: {
		en: { lang },
		uk: { lang },
	},
	loaders: [
		{
			locale: 'en',
			key: 'common',
			loader: async () => (await import('./en/common.json')).default,
		},
		{
			locale: 'en',
			key: 'landing',
			routes: ['/'],
			loader: async () => (await import('./en/landing.json')).default,
		},
		{
			locale: 'en',
			key: 'auth',
			routes: ['/login', '/register', '/forgotPassword', '/emailVerify', '/resetPassword'],
			loader: async () => (await import('./en/auth.json')).default,
		},
		{
			locale: 'uk',
			key: 'common',
			loader: async () => (await import('./uk/common.json')).default,
		},
		{
			locale: 'uk',
			key: 'landing',
			routes: ['/'],
			loader: async () => (await import('./uk/landing.json')).default,
		},
		{
			locale: 'uk',
			key: 'auth',
			routes: ['/login', '/register', '/forgotPassword', '/emailVerify', '/resetPassword'],
			loader: async () => (await import('./uk/auth.json')).default,
		},
	],
};

export const {
	t,
	locale,
	locales,
	loading,
	translations,
	loadTranslations,
	addTranslations,
	setLocale,
	setRoute,
} = new i18n(config);
