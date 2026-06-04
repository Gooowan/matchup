import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'com.matchup.app',
	appName: 'MatchUp',
	webDir: 'build',
	server: {
		iosScheme: 'https'
	},
	plugins: {
		PushNotifications: {
			presentationOptions: ['badge', 'sound', 'alert']
		},
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
		},
		// Google Sign-In is now handled via @capgo/capacitor-social-login.
		// webClientId is passed at runtime from VITE_GOOGLE_CLIENT_ID; no static
		// config key is required by the new plugin.
	}
};

export default config;
