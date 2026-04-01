import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import svg from '@poppanator/sveltekit-svg';

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
		})
	]
});
