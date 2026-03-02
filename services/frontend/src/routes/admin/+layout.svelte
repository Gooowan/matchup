<script lang="ts">
	import AppLayout from '$lib/components/nav/AppLayout.svelte';
	import { authStore } from '$stores/auth.svelte.js';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import HomeIcon from '@lucide/svelte/icons/home';
	import UsersIcon from '@lucide/svelte/icons/users';
	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
	import BackIcon from '@lucide/svelte/icons/arrow-left';
	import WalletIcon from '@lucide/svelte/icons/wallet';

	let { data, children } = $props();
	let isAuthorized = $state(false);

	const navItems = [
		{ href: '/app', label: 'App', icon: BackIcon },
		{ href: '/admin', label: 'Dashboard', icon: HomeIcon },
		{ href: '/admin/users', label: 'Users', icon: UsersIcon },
		{ href: '/admin/withdraws', label: 'Withdrawals', icon: WalletIcon },
		{ href: '/admin/marketing', label: 'Marketing', icon: FolderOpenIcon },
	];

	async function checkAuth() {
		if (!authStore.isAuthenticated || !authStore.user) {
			await authStore.checkAuth();
		}

		if (!authStore.isAdmin) {
			goto('/app');
			return;
		}

		isAuthorized = true;
	}

	onMount(async () => {
		await checkAuth();
	});
</script>

<svelte:head>
	<title>Admin Panel | Company Name</title>
</svelte:head>

{#if isAuthorized}
	<AppLayout
		{navItems}
		basePath="/admin"
		showReferralLink={false}
		showFooter={false}
		showUserProfile={true}
		{data}
	>
		{@render children?.()}
	</AppLayout>
{:else}
	<div class="flex min-h-screen items-center justify-center">
		<div class="text-center">
			<h1 class="text-2xl font-bold">Access Denied</h1>
			<p class="text-muted-foreground">You do not have permission to access this area.</p>
		</div>
	</div>
{/if}
