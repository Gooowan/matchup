import i18n from 'sveltekit-i18n';
import lang from './lang.json';

export const defaultLocale = 'uk';

const config = {
	translations: {
		en: { lang },
		uk: { lang },
	},
	loaders: [
		// --- English ---
		{ locale: 'en', key: 'common', loader: async () => (await import('./en/common.json')).default },
		{ locale: 'en', key: 'landing', routes: ['/'], loader: async () => (await import('./en/landing.json')).default },
		{ locale: 'en', key: 'auth', routes: ['/login', '/register', '/forgotPassword', '/resetPassword'], loader: async () => (await import('./en/auth.json')).default },
		{ locale: 'en', key: 'nav', loader: async () => (await import('./en/nav.json')).default },
		{ locale: 'en', key: 'map', routes: ['/map', '/settings/clubs'], loader: async () => (await import('./en/map.json')).default },
		{ locale: 'en', key: 'feed', routes: ['/feed'], loader: async () => (await import('./en/feed.json')).default },
		{ locale: 'en', key: 'swipe', loader: async () => (await import('./en/swipe.json')).default },
		{ locale: 'en', key: 'chats', routes: ['/chats', '/chats/[id]'], loader: async () => (await import('./en/chats.json')).default },
		{ locale: 'en', key: 'marketplace', routes: ['/marketplace'], loader: async () => (await import('./en/marketplace.json')).default },
		{ locale: 'en', key: 'settings', loader: async () => (await import('./en/settings.json')).default },
		{ locale: 'en', key: 'profile', routes: ['/profiles/[userId]'], loader: async () => (await import('./en/profile.json')).default },
		{ locale: 'en', key: 'filters', loader: async () => (await import('./en/filters.json')).default },
		{ locale: 'en', key: 'onboarding', routes: ['/onboarding'], loader: async () => (await import('./en/onboarding.json')).default },
		{ locale: 'en', key: 'business', routes: ['/business'], loader: async () => (await import('./en/business.json')).default },
		{ locale: 'en', key: 'trainers', loader: async () => (await import('./en/trainers.json')).default },

		// --- Ukrainian ---
		{ locale: 'uk', key: 'common', loader: async () => (await import('./uk/common.json')).default },
		{ locale: 'uk', key: 'landing', routes: ['/'], loader: async () => (await import('./uk/landing.json')).default },
		{ locale: 'uk', key: 'auth', routes: ['/login', '/register', '/forgotPassword', '/resetPassword'], loader: async () => (await import('./uk/auth.json')).default },
		{ locale: 'uk', key: 'nav', loader: async () => (await import('./uk/nav.json')).default },
		{ locale: 'uk', key: 'map', routes: ['/map', '/settings/clubs'], loader: async () => (await import('./uk/map.json')).default },
		{ locale: 'uk', key: 'feed', routes: ['/feed'], loader: async () => (await import('./uk/feed.json')).default },
		{ locale: 'uk', key: 'swipe', loader: async () => (await import('./uk/swipe.json')).default },
		{ locale: 'uk', key: 'chats', routes: ['/chats', '/chats/[id]'], loader: async () => (await import('./uk/chats.json')).default },
		{ locale: 'uk', key: 'marketplace', routes: ['/marketplace'], loader: async () => (await import('./uk/marketplace.json')).default },
		{ locale: 'uk', key: 'settings', loader: async () => (await import('./uk/settings.json')).default },
		{ locale: 'uk', key: 'profile', routes: ['/profiles/[userId]'], loader: async () => (await import('./uk/profile.json')).default },
		{ locale: 'uk', key: 'filters', loader: async () => (await import('./uk/filters.json')).default },
		{ locale: 'uk', key: 'onboarding', routes: ['/onboarding'], loader: async () => (await import('./uk/onboarding.json')).default },
		{ locale: 'uk', key: 'business', routes: ['/business'], loader: async () => (await import('./uk/business.json')).default },
		{ locale: 'uk', key: 'trainers', loader: async () => (await import('./uk/trainers.json')).default },
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
