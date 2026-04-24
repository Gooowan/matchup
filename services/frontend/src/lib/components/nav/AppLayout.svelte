<script lang="ts">
	import SidebarButton from './SidebarButton.svelte';
	import MobileSidebar from './AppMobileSidebar.svelte';
	import Footer from './Footer.svelte';

	import { authStore } from '$stores/auth.svelte.js';
	import { page } from '$app/stores';
	import Logotype from '$lib/assets/svg/logotype.svg?component';
	import Logo from '$lib/assets/svg/logo.svg?component';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import HomeIcon from '@lucide/svelte/icons/home';

	interface NavItem {
		href: string;
		label: string;
		icon: typeof HomeIcon;
	}

	let {
		children,
		navItems,
		showFooter = true,
		showUserProfile = true,
		basePath = '/app',
		collapsible = 'icon',
		variant = 'floating',
		data,
	}: {
		data?: any;
		children?: any;
		navItems: NavItem[];
		showFooter?: boolean;
		showUserProfile?: boolean;
		basePath?: string;
		collapsible?: 'icon' | 'none';
		variant?: 'floating' | 'sidebar' | 'inset';
	} = $props();

	let mobileMenuOpen = $state(false);

	function handleLogout() {
		authStore.logout();
	}
</script>

<Sidebar.Provider>
	<Sidebar.Root
		{variant}
		collapsible={collapsible === 'none' ? undefined : collapsible}
		class="hidden md:flex"
	>
		<Sidebar.Header
			class="flex-row items-center justify-between p-4 group-data-[collapsible=icon]:flex-col group-data-[collapsible=icon]:items-center md:p-6"
		>
			<a href={basePath} class="flex items-center justify-center">
				<Logo
					class="mx-auto hidden h-6 w-6 fill-black group-data-[collapsible=icon]:block dark:fill-current"
				/>
				<Logotype
					class="h-6 w-auto fill-black group-data-[collapsible=icon]:hidden dark:fill-current"
				/>
			</a>
			{#if collapsible === 'icon'}
				<Sidebar.Trigger class="size-12" />
			{/if}
		</Sidebar.Header>

		<div class="px-4">
			<Separator />
		</div>

		<Sidebar.Content>
			<Sidebar.Group class="p-4 md:p-6">
				<Sidebar.Menu>
					{#each navItems as item}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton tooltipContent={item.label}>
								{#snippet child({ props })}
									{@const Icon = item.icon}
									<SidebarButton
										href={item.href}
										variant={$page.url.pathname === item.href ? 'outline_primary' : 'ghost'}
									>
										<Icon class="h-5 w-5" />
										<span class="group-data-[collapsible=icon]:hidden">{item.label}</span>
									</SidebarButton>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.Group>
		</Sidebar.Content>

		{#if showUserProfile}
			<Sidebar.Footer class="px-6 pb-6">
				<Sidebar.Content>
					<Sidebar.Group class="p-0">
						<Sidebar.Menu>
							<div class="py-2">
								<Separator />
							</div>
							<div
								class="flex items-center justify-between group-data-[collapsible=icon]:!justify-center"
							>
								<div class="flex items-center gap-2">
									<Avatar.Root class="size-7">
										<Avatar.Image
											src={authStore.user?.profile_data.avatar}
											alt={authStore.user?.email}
										/>
										<Avatar.Fallback class="text-xs">
											{authStore.user?.profile_data.first_name.charAt(0) ??
												'D'}{authStore.user?.profile_data.last_name.charAt(0) ?? 'S'}
										</Avatar.Fallback>
									</Avatar.Root>
									<div class="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
										<p class="truncate text-sm font-medium">
											{authStore.user?.profile_data.first_name ?? 'Company'}
											{authStore.user?.profile_data.last_name ?? 'Network'}
										</p>
										<p class="text-muted-foreground truncate text-xs">
											{authStore.user?.email ?? 'email@company.network'}
										</p>
									</div>
								</div>
							</div>

							<div class="py-2">
								<Separator />
							</div>

							<Sidebar.MenuItem class="">
								<Sidebar.MenuButton tooltipContent="Logout">
									{#snippet child({ props })}
										<SidebarButton variant="ghost" onclick={handleLogout}>
											<LogOutIcon class="h-5 w-5" />
											<span class="group-data-[collapsible=icon]:hidden">Logout</span>
										</SidebarButton>
									{/snippet}
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						</Sidebar.Menu>
					</Sidebar.Group>
				</Sidebar.Content>
			</Sidebar.Footer>
		{:else}
			<Sidebar.Footer>
				<div class="p-4">
					<Sidebar.MenuItem>
						<Sidebar.MenuButton tooltipContent="Logout">
							{#snippet child({ props })}
								<SidebarButton variant="ghost" onclick={handleLogout}>
									<LogOutIcon class="h-5 w-5" />
									<span class="group-data-[collapsible=icon]:hidden">Logout</span>
								</SidebarButton>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</div>
			</Sidebar.Footer>
		{/if}
	</Sidebar.Root>

	<div class="mx-auto w-full max-w-screen-lg flex-1">
		<header class="flex items-center justify-between border-b px-4 py-4 md:hidden">
			<a href={basePath}>
				<Logotype class="h-6 w-auto fill-black dark:fill-current" />
			</a>
			<button
				onclick={() => (mobileMenuOpen = true)}
				class="hover:bg-accent flex size-12 items-center justify-center rounded-md"
			>
				<MenuIcon class="h-6 w-6" />
			</button>
		</header>

		{@render children?.()}

		{#if showFooter}
			<Footer {data} />
		{/if}
	</div>

	<!-- Mobile Sidebar -->
	<MobileSidebar
		bind:isOpen={mobileMenuOpen}
		{navItems}
		currentPath={$page.url.pathname}
		user={authStore.user}
		onLogout={handleLogout}
	/>
</Sidebar.Provider>
