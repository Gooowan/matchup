import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'com.matchup.app',
	appName: 'MatchUp',
	webDir: 'build',
	server: {
		iosScheme: 'https'
	},
	plugins: {
		StatusBar: {
			style: 'Dark',
			backgroundColor: '#dae1eb'
		},
		SplashScreen: {
			launchAutoHide: false,
			backgroundColor: '#dae1eb',
			showSpinner: false
		},
		Keyboard: {
			resize: 'body',
			resizeOnFullScreen: true
		}
	}
};

export default config;
