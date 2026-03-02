<script lang="ts">
	import { authStore } from '$stores/auth.svelte';
	import { userReferralStore } from '$stores/user_referral.svelte';
	import { t } from '$lib/locale';
	import ReferralTable from '$lib/components/ReferralTable.svelte';
	import CopyReferralLink from '$lib/components/CopyReferralLink.svelte';
	import EditableField from '$lib/components/EditableField.svelte';
	import AvatarUpload from '$lib/components/AvatarUpload.svelte';
	import PasswordChangeForm from '$lib/components/PasswordChangeForm.svelte';
	import { Button } from '$lib/components/ui/button';
	import { KeyRound, Shield } from '@lucide/svelte';
	import toast from 'svelte-french-toast';
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';

	let isPasswordModalOpen = $state(false);

	onMount(() => {
		userReferralStore.fetchReferralCount();
	});

	async function handleFirstNameSave(newValue: string) {
		const result = await authStore.updateProfile(
			newValue,
			authStore.user?.profile_data.last_name || ''
		);
		if (result.success) {
			toast.success('First name updated successfully!');
		}
		return result;
	}

	async function handleLastNameSave(newValue: string) {
		const result = await authStore.updateProfile(
			authStore.user?.profile_data.first_name || '',
			newValue
		);
		if (result.success) {
			toast.success('Last name updated successfully!');
		}
		return result;
	}

	async function handleAvatarUpload(file: File) {
		return await authStore.uploadAvatar(file);
	}

	async function handlePasswordChange(currentPassword: string, newPassword: string) {
		return await authStore.changePassword(currentPassword, newPassword);
	}
</script>

<section class="space-y-6 px-4 py-8">
	<div class="">
		<h1 class="text-3xl font-bold text-foreground">Profile</h1>
		<p class="mt-2 text-muted-foreground">Manage your profile settings and team</p>
	</div>

	<Card.Root
		class="grid grid-cols-1 items-start justify-between gap-4 p-4 md:grid-cols-2 lg:grid-cols-3"
	>
		<div class="flex flex-col gap-4 md:col-span-2 lg:col-span-1">
			<p class="text-lg font-bold">Settings</p>
			<div class="flex gap-4">
				<AvatarUpload
					currentAvatar={authStore.user?.profile_data.avatar}
					fallbackText="{authStore.user?.profile_data.first_name.charAt(0) ??
						'D'}{authStore.user?.profile_data.last_name.charAt(0) ?? 'N'}"
					onUpload={handleAvatarUpload}
				/>
				<div class="flex flex-col gap-2">
					<EditableField
						value={authStore.user?.profile_data.first_name ?? 'Desim'}
						label="First Name"
						onSave={handleFirstNameSave}
					/>
					<EditableField
						value={authStore.user?.profile_data.last_name ?? 'Network'}
						label="Last Name"
						onSave={handleLastNameSave}
					/>
					<p class="text-neutral-600">{authStore.user?.email ?? 'email@desim.network'}</p>
				</div>
			</div>
		</div>
		<div class="flex flex-col lg:items-end">
			<div class="space-y-4">
				<div class="">
					<p class="mb-2 text-neutral-600">Referrals</p>
					<div class="flex h-9 items-center gap-2">
						{#if userReferralStore.isLoading}
						<p class="font-bold">Loading...</p>
						{:else if userReferralStore.error}
						<p class="font-bold text-red-500">Error</p>
						{:else if userReferralStore.referralCount}
						<p class="font-bold">
							{userReferralStore.referralCount.direct_count} direct, {userReferralStore
								.referralCount.total_count} total
						</p>
						{:else}
						<p class="font-bold">0</p>
						{/if}
					</div>
				</div>
				<div>
					<p class="mb-2 text-neutral-600">Copy Referral Link</p>
					<CopyReferralLink referralId={authStore.user?.referral_id ?? 0} />
				</div>
			</div>
		</div>
		<div class="flex flex-col lg:items-end">
		<div class="space-y-4">
			<div>
				<p class="mb-2 text-neutral-600">Password</p>
				<Button variant="outline" onclick={() => (isPasswordModalOpen = true)} class="max-w-xs">
					<KeyRound class="h-4 w-4 flex-shrink-0" />
					<span class="text-xs">Change Password</span>
				</Button>
			</div>
			<div>
				<p class="mb-2 text-neutral-600">2FA</p>
				<Button variant="outline" disabled class="max-w-xs">
					<Shield class="h-4 w-4 flex-shrink-0" />
					<span class="text-xs">Coming Soon</span>
				</Button>
			</div>

			</div>
		</div>
	</Card.Root>

	<PasswordChangeForm
		bind:isOpen={isPasswordModalOpen}
		onClose={() => (isPasswordModalOpen = false)}
		onPasswordChange={handlePasswordChange}
	/>

	<ReferralTable
		endpoint="/user/referrals/search"
		title="Global Team Search"
		showSearch={true}
		showLevel={true}
		itemsPerPage={10}
	/>
</section>
