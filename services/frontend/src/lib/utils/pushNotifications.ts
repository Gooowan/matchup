import { Capacitor } from '@capacitor/core';
import { authFetch } from './authFetch';

export async function registerPushNotifications(): Promise<void> {
	if (!Capacitor.isNativePlatform()) return;

	try {
		const { PushNotifications } = await import('@capacitor/push-notifications');

		const permResult = await PushNotifications.requestPermissions();
		if (permResult.receive !== 'granted') return;

		await PushNotifications.register();

		PushNotifications.addListener('registration', async (token) => {
			await authFetch('/me/push-token', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token: token.value, platform: 'ios' })
			});
		});
	} catch {
		// Graceful no-op — push is non-critical
	}
}
