<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import toast from 'svelte-french-toast';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		onPasswordChange: (
			currentPassword: string,
			newPassword: string
		) => Promise<{ success: boolean; error?: string }>;
	}

	let { isOpen = $bindable(), onClose, onPasswordChange }: Props = $props();

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let errors = $state({
		currentPassword: '',
		newPassword: '',
		confirmPassword: '',
	});

	function validateForm(): boolean {
		const newErrors = {
			currentPassword: '',
			newPassword: '',
			confirmPassword: '',
		};

		let isValid = true;

		if (!currentPassword) {
			newErrors.currentPassword = 'Current password is required';
			isValid = false;
		}

		if (!newPassword) {
			newErrors.newPassword = 'New password is required';
			isValid = false;
		} else if (newPassword.length < 8) {
			newErrors.newPassword = 'Password must be at least 8 characters';
			isValid = false;
		}

		if (!confirmPassword) {
			newErrors.confirmPassword = 'Please confirm your new password';
			isValid = false;
		} else if (newPassword !== confirmPassword) {
			newErrors.confirmPassword = 'Passwords do not match';
			isValid = false;
		}

		errors = newErrors;
		return isValid;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		isLoading = true;

		try {
			const result = await onPasswordChange(currentPassword, newPassword);

			if (result.success) {
				toast.success('Password changed successfully!');
				// Reset form
				currentPassword = '';
				newPassword = '';
				confirmPassword = '';
				errors = {
					currentPassword: '',
					newPassword: '',
					confirmPassword: '',
				};
				// Close modal on success
				onClose();
			} else {
				const errorMessage = result.error || 'Failed to change password';
				toast.error(errorMessage);
				errors.currentPassword = errorMessage;
			}
		} catch (err) {
			toast.error('Network error occurred');
		} finally {
			isLoading = false;
		}
	}

	function handleClose() {
		if (!isLoading) {
			onClose();
		}
	}

	function clearError(field: 'currentPassword' | 'newPassword' | 'confirmPassword') {
		errors[field] = '';
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen}>
	<ResponsiveDialog.Content>
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Change Password</ResponsiveDialog.Title>
			<ResponsiveDialog.Description>
				Update your password. Make sure it's at least 8 characters long.
			</ResponsiveDialog.Description>
		</ResponsiveDialog.Header>

		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<Label for="current-password">Current Password</Label>
				<Input
					id="current-password"
					type="password"
					bind:value={currentPassword}
					oninput={() => clearError('currentPassword')}
					disabled={isLoading}
					placeholder="Enter current password"
					class="bg-transparent"
				/>
				{#if errors.currentPassword}
					<p class="text-sm text-red-500">{errors.currentPassword}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="new-password">New Password</Label>
				<Input
					id="new-password"
					type="password"
					bind:value={newPassword}
					oninput={() => clearError('newPassword')}
					disabled={isLoading}
					placeholder="Enter new password (min 8 characters)"
					class="bg-transparent"
				/>
				{#if errors.newPassword}
					<p class="text-sm text-red-500">{errors.newPassword}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="confirm-password">Confirm New Password</Label>
				<Input
					id="confirm-password"
					type="password"
					bind:value={confirmPassword}
					oninput={() => clearError('confirmPassword')}
					disabled={isLoading}
					placeholder="Confirm new password"
					class="bg-transparent"
				/>
				{#if errors.confirmPassword}
					<p class="text-sm text-red-500">{errors.confirmPassword}</p>
				{/if}
			</div>

			<ResponsiveDialog.Footer>
				<Button type="button" variant="outline" onclick={handleClose} disabled={isLoading}>
					Cancel
				</Button>
				<Button type="submit" disabled={isLoading}>
					{isLoading ? 'Changing Password...' : 'Change Password'}
				</Button>
			</ResponsiveDialog.Footer>
		</form>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>

