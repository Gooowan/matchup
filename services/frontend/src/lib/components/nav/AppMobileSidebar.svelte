<script lang="ts">
	import { Button } from '$components/ui/button/index.js';
	import CopyReferralLink from '$lib/components/CopyReferralLink.svelte';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import HomeIcon from '@lucide/svelte/icons/home';
	import XIcon from '@lucide/svelte/icons/x';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import UserPenIcon from '@lucide/svelte/icons/user-pen';
	import Logotype from '$lib/assets/svg/logotype.svg?component';

	interface NavItem {
		href: string;
		label: string;
		icon: typeof HomeIcon;
	}

	interface User {
		email: string;
		referral_id: number;
		profile_data: Record<string, any>;
	}

	let {
		isOpen = $bindable(false),
		navItems,
		currentPath,
		user,
		onLogout,
	}: {
		isOpen?: boolean;
		navItems: NavItem[];
		currentPath: string;
		user: User | null;
		onLogout: () => void;
	} = $props();

	$effect(() => {
		if (isOpen) {
			document.body.style.overflow = 'hidden';
			document.body.style.position = 'fixed';
			document.body.style.width = '100%';
		} else {
			document.body.style.overflow = '';
			document.body.style.position = '';
			document.body.style.width = '';
		}
	});

	function handleNavClick() {
		isOpen = false;
	}

	function handleClose() {
		isOpen = false;
	}
</script>

{#if isOpen}
	<div
		class="bg-linear-to-b absolute inset-0 z-[9999] flex h-screen w-screen flex-col from-[#000] via-[#000]/50 to-[#000] backdrop-blur-xl"
		style="height: 100vh; height: 100dvh;"
	>
		<!-- Header -->
		<div class="flex items-center justify-between border-b p-4">
			<a href="/app" onclick={handleNavClick}>
				<Logotype class="h-6 w-auto fill-black dark:fill-current" />
			</a>
			<button
				onclick={handleClose}
				class="hover:bg-accent flex size-12 items-center justify-center rounded-md"
			>
				<XIcon class="h-6 w-6" />
			</button>
		</div>

		<!-- Content - Scrollable -->
		<div class="flex-1 overflow-y-auto">
			<div class="p-4">
				<!-- Navigation Items -->
				<div class="space-y-2">
					{#each navItems as item}
						{@const Icon = item.icon}
						<Button
							variant={currentPath === item.href ? 'outline_primary' : 'ghost'}
							class="h-12 w-full justify-start gap-4 px-4 text-lg"
							href={item.href}
							onclick={handleNavClick}
						>
							<Icon class="h-5 w-5" />
							<span>{item.label}</span>
						</Button>
					{/each}
				</div>

				<div class="py-2">
					<Separator />
				</div>

				<!-- Referral Link -->
				{#if user}
					<div class="space-y-2">
						<p class="text-muted-foreground text-sm">Copy my referral link</p>
						<CopyReferralLink referralId={user.referral_id} />
					</div>

					<div class="py-2">
						<Separator />
					</div>
				{/if}
			</div>
		</div>

		<!-- Footer -->
		<div class="border-t p-4">
			<div class="space-y-2">

				<!-- User Info -->
				{#if user}
					<div class="flex items-center gap-2">
						<Avatar.Root class="size-7">
							<Avatar.Image src={user.profile_data.avatar} alt={user.email} />
							<Avatar.Fallback class="text-xs">
								{user.profile_data.first_name.charAt(0) ?? 'D'}{user.profile_data.last_name.charAt(
									0
								) ?? 'S'}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-medium">
								{user.profile_data.first_name ?? 'Company'}
								{user.profile_data.last_name ?? 'Network'}
							</p>
							<p class="text-muted-foreground truncate text-xs">
								{user.email ?? 'email@company.network'}
							</p>
						</div>
					</div>
				{/if}

				<div class="py-2">
					<Separator />
				</div>
				<!-- Logout Button -->
				<Button
					variant="ghost"
					class="h-12 w-full justify-start gap-4 px-4 text-lg"
					onclick={() => {
						handleNavClick();
						onLogout();
					}}
				>
					<LogOutIcon class="h-5 w-5" />
					<span>Logout</span>
				</Button>
			</div>
		</div>
	</div>
{/if}
