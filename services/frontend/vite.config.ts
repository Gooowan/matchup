import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import svg from '@poppanator/sveltekit-svg';
import { sentryVitePlugin } from '@sentry/vite-plugin';

import { defineConfig } from 'vite';

export default defineConfig({
	envDir: '../../',
	server: {
		allowedHosts: [
			'0.0.0.0',
			'.ptzhn.in.ua',
			'.ptzhn.com.ua',
			'.tunnel.in.ua',
			'.potuzhno.in.ua',
			'.front.testinger.cx.ua',
			'.desim.network'
		]
	},
	build: {
		// Source maps are required for Sentry to display readable stack traces.
		// They are uploaded to Sentry in CI and are NOT shipped to end users.
		sourcemap: true,
		rollupOptions: {
			// Capacitor native plugins are not available in the browser build;
			// they are resolved at runtime by the Capacitor iOS/Android bridge.
			external: ['@capacitor/push-notifications']
		}
	},
	plugins: [
		tailwindcss(),
		sveltekit(),
		svg({
			type: 'component',
			includePaths: ['./src/lib/assets/svg'],
			svgoOptions: {
				multipass: true,
				plugins: [
					{
						name: 'preset-default',
						params: { overrides: { removeViewBox: false } }
					}
				]
			}
		}),
		// Upload source maps to Sentry during CI builds only.
		// Requires SENTRY_AUTH_TOKEN, SENTRY_ORG, SENTRY_PROJECT env vars.
		process.env.SENTRY_AUTH_TOKEN &&
			sentryVitePlugin({
				authToken: process.env.SENTRY_AUTH_TOKEN,
				org: process.env.SENTRY_ORG,
				project: process.env.SENTRY_PROJECT ?? 'matchup-frontend'
			})
	].filter(Boolean)
});
