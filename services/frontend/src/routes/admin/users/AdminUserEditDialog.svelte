<script lang="ts">
	import { authFetch } from '$utils/authFetch';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	interface Props {
		user: AdminUserDTO | null;
		open: boolean;
		onClose: () => void;
		onSaved: () => void;
	}

	let { user, open, onClose, onSaved }: Props = $props();

	let formData = $state({
		email: '',
		role: 'USER',
		profile_data: {
			first_name: '',
			last_name: '',
			avatar: '',
		},
		metadata: {},
	});

	let isLoading = $state(false);
	let error = $state<string | null>(null);

	function resetForm() {
		if (user) {
			formData = {
				email: user.email,
				role: user.role,
				profile_data: {
					first_name: user.profile_data?.first_name || '',
					last_name: user.profile_data?.last_name || '',
					avatar: user.profile_data?.avatar || '',
				},
				metadata: user.metadata || {},
			};
		}
		error = null;
	}

	async function handleSubmit() {
		if (!user) return;

		isLoading = true;
		error = null;

		const updateData: Record<string, unknown> = {
			email: formData.email !== user.email ? formData.email : undefined,
			role: formData.role !== user.role ? formData.role : undefined,
			profile_data:
				JSON.stringify(formData.profile_data) !== JSON.stringify(user.profile_data)
					? formData.profile_data
					: undefined,
		};

		// Remove undefined values
		Object.keys(updateData).forEach((key) => {
			if (updateData[key] === undefined) {
				delete updateData[key];
			}
		});

		if (Object.keys(updateData).length === 0) {
			onClose();
			isLoading = false;
			return;
		}

		const resp = await authFetch(`/admin/users/${user.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(updateData),
		});

		if (!resp.ok) {
			error = 'Failed to update user';
			console.error('Update user error:', resp.status, resp.statusText);
			isLoading = false;
			return;
		}

		const response = await resp.json();
		if (response.error) {
			error = response.error;
			console.error('Update user API error:', response.error);
			isLoading = false;
			return;
		}

		isLoading = false;
		onSaved();
		onClose();
	}

	$effect(() => {
		if (open && user) {
			resetForm();
		}
	});

	$effect(() => {
		if (!open) {
			error = null;
		}
	});
</script>

<ResponsiveDialog.Root
	{open}
	onOpenChange={(newOpen) => {
		if (!newOpen) onClose();
	}}
>
	<ResponsiveDialog.Content>
		<div class="mx-auto w-full max-w-sm overflow-y-auto">
			<ResponsiveDialog.Header class="text-left">
				<ResponsiveDialog.Title>Edit User</ResponsiveDialog.Title>
				<ResponsiveDialog.Description>
					Make changes to the user account. Click save when you're done.
				</ResponsiveDialog.Description>
			</ResponsiveDialog.Header>

			{#if error}
				<div class="bg-destructive/15 mx-4 rounded-md p-3">
					<p class="text-destructive text-sm">{error}</p>
				</div>
			{/if}

			<form
				class="grid gap-4 px-4"
				onsubmit={(e) => {
					e.preventDefault();
					handleSubmit();
				}}
			>
				<div class="grid gap-4">
					<div class="grid gap-2">
						<Label for="first_name">First Name</Label>
						<Input
							id="first_name"
							bind:value={formData.profile_data.first_name}
							placeholder="First name"
							disabled={isLoading}
						/>
					</div>
					<div class="grid gap-2">
						<Label for="last_name">Last Name</Label>
						<Input
							id="last_name"
							bind:value={formData.profile_data.last_name}
							placeholder="Last name"
							disabled={isLoading}
						/>
					</div>
				</div>

				<div class="grid gap-2">
					<Label for="email">Email</Label>
					<Input
						id="email"
						type="email"
						bind:value={formData.email}
						placeholder="user@example.com"
						disabled={isLoading}
					/>
				</div>

				<div class="grid gap-2">
					<Label for="role">Role</Label>
					<Select.Root type="single" bind:value={formData.role}>
						<Select.Trigger>
							{formData.role ?? 'Select role'}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="USER" label="USER">USER</Select.Item>
							<Select.Item value="ADMIN" label="ADMIN">ADMIN</Select.Item>
						</Select.Content>
					</Select.Root>
				</div>

				<div class="grid gap-2">
					<Label for="avatar">Avatar URL</Label>
					<Input
						id="avatar"
						bind:value={formData.profile_data.avatar}
						placeholder="https://example.com/avatar.jpg"
						disabled={isLoading}
					/>
				</div>

				<Button type="submit" disabled={isLoading} class="mt-4">
					{isLoading ? 'Saving...' : 'Save Changes'}
				</Button>
			</form>

			<ResponsiveDialog.Footer class="pt-2">
				<ResponsiveDialog.Close>
					{#snippet child({ props })}
						<Button {...props} variant="outline" disabled={isLoading}>Cancel</Button>
					{/snippet}
				</ResponsiveDialog.Close>
			</ResponsiveDialog.Footer>
		</div>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
