import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter({ fallback: '200.html' }),
		alias: {
			$components: './src/lib/components',
			$assets: './src/lib/assets',
			$svg: './src/lib/assets/svg',
			$stores: './src/lib/stores',
			$utils: './src/lib/utils',
			$types: './src/lib/types',
		},
		csrf: {
			trustedOrigins: ['*'],
		},
	},
	compilerOptions: {
		warningFilter: (warning) =>
			// @ts-ignore
			!warning.filename?.includes('node_modules') && !warning.code.startsWith('a11y'),
	},
};

export default config;
