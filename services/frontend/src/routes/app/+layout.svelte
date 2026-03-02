<script lang="ts">
	import AppLayout from '$lib/components/nav/AppLayout.svelte';
	import { authStore } from '$stores/auth.svelte.js';
	import { onMount } from 'svelte';
	import HomeIcon from '@lucide/svelte/icons/home';
	import SimIcon from '@lucide/svelte/icons/card-sim';
	import WalletIcon from '@lucide/svelte/icons/wallet';
	import UsersIcon from '@lucide/svelte/icons/users';
	import BriefcaseIcon from '@lucide/svelte/icons/briefcase';
	import ShieldIcon from '@lucide/svelte/icons/shield-check';
	import UserPenIcon from '@lucide/svelte/icons/user-pen';
	import FarmIcon from '@lucide/svelte/icons/chart-line';

	let { data, children } = $props();

	interface NavItem {
		href: string;
		label: string;
		icon: typeof HomeIcon;
	}

	const baseNavItems: NavItem[] = [
		{ href: `/app`, label: 'Home', icon: HomeIcon },
		{ href: `/app/farm`, label: 'Farm', icon: FarmIcon },
		{ href: `/app/esim`, label: 'eSIM', icon: SimIcon },
		{ href: `/app/finance`, label: 'Finance', icon: WalletIcon },
		{ href: `/app/referral`, label: 'Referral', icon: UsersIcon },
		{ href: `/app/career`, label: 'Career', icon: BriefcaseIcon },
		{ href: `/app/id`, label: 'Profile', icon: UserPenIcon },
	];

	const navItems = $derived(
		authStore.isAdmin
			? [...baseNavItems, { href: `/admin`, label: 'Admin', icon: ShieldIcon }]
			: baseNavItems
	);

	onMount(async () => {
		if (!authStore.isAuthenticated || !authStore.user) {
			const isAuthenticated = await authStore.checkAuth();
			if (!isAuthenticated) {
				authStore.logout();
			}
		}
	});
</script>

<AppLayout {navItems} basePath="/app" {data}>
	{@render children?.()}
</AppLayout>
