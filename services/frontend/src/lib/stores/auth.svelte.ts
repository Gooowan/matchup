import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authFetch } from '$utils/authFetch';

function createAuthStore() {
	let user = $state<UserDTO | null>(null);
	let isAuthenticated = $state<boolean>(false);
	let loggingOut = false;

	return {
		get user() {
			return user;
		},
		set user(value) {
			user = value;
		},
		get isAuthenticated() {
			return isAuthenticated;
		},
		set isAuthenticated(value) {
			isAuthenticated = value;
		},
		get isAdmin() {
			return user?.role === 'ADMIN';
		},
		hasRole(role: string) {
			return user?.role === role;
		},
		async login(userData: UserDTO) {
			this.user = userData;
			this.isAuthenticated = true;
		},
		async checkAuth() {
			const resp = await authFetch('/user/profile');

			const response: ApiResponse<{ user: UserDTO }> = await resp.json();

			if (resp.status === 200 && response.data) {
				this.user = response.data.user;
				this.isAuthenticated = true;
				return true;
			}

			this.logout();
			return false;
		},
		async logout(redirect: boolean = true, redirectUrl: string = '/login') {
			if (loggingOut) return;
			loggingOut = true;

			this.user = null;
			this.isAuthenticated = false;

			try {
				const apiUrl = import.meta.env.VITE_API_URL;
				await fetch(`${apiUrl}/auth/logout`, { method: 'POST', credentials: 'include' });
			} catch {
				// Best-effort; ignore logout API errors
			}

			if (browser && redirect) {
				await goto(redirectUrl);
			}

			loggingOut = false;
		},
		async updateProfile(firstName: string, lastName: string) {
			const resp = await authFetch('/user/profile/update', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					first_name: firstName,
					last_name: lastName,
				}),
			});

			const response: ApiResponse<{ user: UserDTO }> = await resp.json();

			if (resp.status === 200 && response.data) {
				this.user = response.data.user;
				return { success: true, user: response.data.user };
			}

			return { success: false, error: response.error || 'Failed to update profile' };
		},
		async uploadAvatar(file: File) {
			const formData = new FormData();
			formData.append('avatar', file);

			const resp = await authFetch('/user/files/avatar', {
				method: 'POST',
				body: formData,
			});

			const response: ApiResponse<{ avatar: string }> = await resp.json();

			if (resp.status === 200 && response.data) {
				// Update user avatar in store
				if (this.user) {
					this.user = {
						...this.user,
						profile_data: {
							...this.user.profile_data,
							avatar: response.data.avatar,
						},
					};
				}
				return { success: true, avatar: response.data.avatar };
			}

			return { success: false, error: response.error || 'Failed to upload avatar' };
		},
		async changePassword(currentPassword: string, newPassword: string) {
			const resp = await authFetch('/user/password/change', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					password: currentPassword,
					new_password: newPassword,
				}),
			});

			const response: ApiResponse<string> = await resp.json();

			if (resp.status === 200) {
				return { success: true };
			}

			return { success: false, error: response.error || 'Failed to change password' };
		},
	};
}

export const authStore = createAuthStore();
