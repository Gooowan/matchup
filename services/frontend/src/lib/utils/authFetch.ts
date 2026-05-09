import { goto } from '$app/navigation';
import { authStore } from '$stores/auth.svelte';
import toast from 'svelte-french-toast';

export async function authFetch(endpoint: string, init: RequestInit = {}): Promise<Response> {
	init.credentials = 'include';

	const response = await fetch(`${import.meta.env.VITE_API_URL}${endpoint}`, init);

	if (response.status === 401) {
		toast.error('Session expired, please login again');
		authStore.user = null;
		authStore.isAuthenticated = false;
		goto('/login');
	}

	return response;
}
