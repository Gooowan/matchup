<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { referralStatsStore } from '$lib/stores/referral_stats.svelte';
	import { authStore } from '$stores/auth.svelte';
	import ReferralStatsInfo from '$lib/components/ReferralStatsInfo.svelte';
	import ReferralListItem from '$lib/components/ReferralListItem.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardContent } from '$lib/components/ui/card';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';

	let currentPage = $state(1);
	const itemsPerPage = 10;
	let searchTerm = $state('');

	function getDisplayName(profileData: Record<string, any>, email: string): string {
		const firstName = profileData?.first_name || '';
		const lastName = profileData?.last_name || '';
		const name = `${firstName} ${lastName}`.trim();
		return name || email;
	}

	function handleReferralClick(userId: string, displayName: string, stats?: DirectReferralStat) {
		currentPage = 1;
		searchTerm = '';
		referralStatsStore.navigateTo(userId, displayName, stats);
	}

	function handleBreadcrumbClick(index: number) {
		currentPage = 1;
		searchTerm = '';
		const userId = authStore.user?.id;
		if (userId) {
			referralStatsStore.navigateToPath(index, userId);
		}
	}

	function handlePageChange(page: number) {
		currentPage = page;
		const userId =
			referralStatsStore.navigationPath.length > 0
				? referralStatsStore.navigationPath[referralStatsStore.navigationPath.length - 1].id
				: authStore.user?.id;
		if (userId) {
			referralStatsStore.refreshCurrentView(page, itemsPerPage, searchTerm, userId);
		}
	}

	function handleSearch() {
		currentPage = 1;
		const userId =
			referralStatsStore.navigationPath.length > 0
				? referralStatsStore.navigationPath[referralStatsStore.navigationPath.length - 1].id
				: authStore.user?.id;
		if (userId) {
			referralStatsStore.refreshCurrentView(1, itemsPerPage, searchTerm.trim(), userId);
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSearch();
		}
	}

	let hasLoaded = $state(false);

	// Wait for auth and load root view reactively
	$effect(() => {
		// Wait for auth to be checked
		if (!authStore.isAuthenticated || !authStore.user) {
			authStore.checkAuth().then((isAuthenticated) => {
				if (!isAuthenticated) {
					return;
				}
				const userId = authStore.user?.id;
				if (userId && !hasLoaded && referralStatsStore.navigationPath.length === 0) {
					hasLoaded = true;
					referralStatsStore.loadRootView(userId);
				}
			});
			return;
		}

		const userId = authStore.user?.id;
		if (userId && !hasLoaded && referralStatsStore.navigationPath.length === 0) {
			hasLoaded = true;
			referralStatsStore.loadRootView(userId);
		}
	});

	onDestroy(() => {
		referralStatsStore.clear();
	});
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-4">
		<h1 class="text-foreground text-3xl font-bold">Referral Stats</h1>
		<p class="text-muted-foreground mt-2">View and navigate through your referral network</p>
	</div>

	<!-- Breadcrumb Navigation - Always visible to prevent layout shift -->
	<div class="mb-4">
		<div class="text-muted-foreground flex items-center gap-2 text-sm">
			<button
				onclick={() => handleBreadcrumbClick(-1)}
				class="hover:text-foreground transition-colors"
			>
				My Stats
			</button>
			{#each referralStatsStore.navigationPath as item, index}
				<span>/</span>
				<button
					onclick={() => handleBreadcrumbClick(index)}
					class="hover:text-foreground transition-colors"
				>
					{item.displayName}
				</button>
			{/each}
		</div>
	</div>

	<!-- Error State -->
	{#if referralStatsStore.error}
		<div class="py-8 text-center">
			<p class="text-destructive mb-4">{referralStatsStore.error}</p>
			<Button
				onclick={async () => {
					const userId =
						referralStatsStore.navigationPath.length > 0
							? referralStatsStore.navigationPath[referralStatsStore.navigationPath.length - 1].id
							: authStore.user?.id;
					if (userId) {
						await referralStatsStore.loadRootView(userId);
					}
				}}
			>
				Retry
			</Button>
		</div>
	{/if}

	<!-- Stats Display - Always render with placeholder values during loading -->
	<div class="mb-6">
		<ReferralStatsInfo
			stats={referralStatsStore.currentStats}
			isRoot={referralStatsStore.navigationPath.length === 0}
			displayName={referralStatsStore.navigationPath.length > 0
				? referralStatsStore.navigationPath[referralStatsStore.navigationPath.length - 1]
						.displayName
				: undefined}
		/>
	</div>

	<!-- Direct Referrals List - Always render structure -->
	<div class="mb-6">
		<h2 class="mb-2 text-xl font-semibold">
			{#if referralStatsStore.navigationPath.length === 0}
				Your Direct Referrals
			{:else}
				{referralStatsStore.navigationPath[referralStatsStore.navigationPath.length - 1]
					.displayName}'s Direct Referrals
			{/if}
		</h2>
		<div class="mb-6 flex gap-2">
			<Input
				bind:value={searchTerm}
				placeholder="Search by name or email..."
				onkeydown={handleKeydown}
				class="flex-1"
				disabled={referralStatsStore.isReferralsLoading}
			/>
			<Button onclick={handleSearch} disabled={referralStatsStore.isReferralsLoading}>
				Search
			</Button>
		</div>
	</div>

	<!-- Referrals List - Show loading inline, preserve layout -->
	{#if referralStatsStore.isReferralsLoading && referralStatsStore.currentReferrals.length === 0}
		{@const skeletonCount = referralStatsStore.currentPagination?.take || 3}
		<div class="mb-6 space-y-3">
			{#each Array(skeletonCount) as _, i (i)}
				<Card>
					<CardContent class="px-4">
						<div class="flex items-center justify-between">
							<div class="flex w-full flex-col justify-between gap-2 lg:flex-row">
								<div class="flex flex-col gap-2">
									<div class="flex items-center gap-2">
										<div class="h-5 w-32 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-5 w-16 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-5 w-12 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="h-3 w-40 animate-pulse rounded bg-gray-950/30"></div>
								</div>
								<div class="grid w-full grid-cols-2 gap-6 px-4 md:flex md:flex-wrap md:justify-center lg:justify-end lg:gap-12 lg:px-8">
									<div class="flex flex-col items-center justify-center gap-1">
										<div class="h-6 w-8 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-24 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="flex flex-col items-center justify-center gap-1">
										<div class="h-6 w-8 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="flex flex-col items-center justify-center gap-1">
										<div class="h-6 w-16 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-20 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="flex flex-col items-center justify-center gap-1">
										<div class="h-6 w-20 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-28 animate-pulse rounded bg-gray-950/30"></div>
									</div>
								</div>
							</div>
							<div class="h-5 w-5 animate-pulse rounded bg-gray-950/30"></div>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{:else if referralStatsStore.currentReferrals.length > 0}
		<div class="mb-6 space-y-3 relative">
			{#if referralStatsStore.isReferralsLoading}
				<div class="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-lg pointer-events-none">
					<div class="text-sm text-muted-foreground">Loading...</div>
				</div>
			{/if}
			{#each referralStatsStore.currentReferrals as referral (referral.user_id)}
				<ReferralListItem
					{referral}
					onClick={(userId, displayName) => handleReferralClick(userId, displayName, referral)}
				/>
			{/each}
		</div>

		<!-- Pagination -->
		{#if referralStatsStore.currentPagination && referralStatsStore.currentPagination.pageCount > 1}
			<div class="flex justify-center">
				<Pagination.Root
					count={referralStatsStore.currentPagination.itemCount}
					perPage={referralStatsStore.currentPagination.take}
					bind:page={currentPage}
					onPageChange={(page) => handlePageChange(page)}
				>
					{#snippet children({ pages, currentPage: currentPageValue })}
						<Pagination.Content>
							<Pagination.Item>
								<Pagination.PrevButton>
									<ChevronLeftIcon class="size-4" />
									<span class="hidden sm:block">Previous</span>
								</Pagination.PrevButton>
							</Pagination.Item>
							{#each pages as pageItem (pageItem.key)}
								{#if pageItem.type === 'ellipsis'}
									<Pagination.Item>
										<Pagination.Ellipsis />
									</Pagination.Item>
								{:else}
									<Pagination.Item>
										<Pagination.Link page={pageItem} isActive={currentPageValue === pageItem.value}>
											{pageItem.value}
										</Pagination.Link>
									</Pagination.Item>
								{/if}
							{/each}
							<Pagination.Item>
								<Pagination.NextButton>
									<span class="hidden sm:block">Next</span>
									<ChevronRightIcon class="size-4" />
								</Pagination.NextButton>
							</Pagination.Item>
						</Pagination.Content>
					{/snippet}
				</Pagination.Root>
			</div>
		{/if}
	{:else if !referralStatsStore.isReferralsLoading && referralStatsStore.currentStats}
		<div class="py-8 text-center">
			<p class="text-muted-foreground">No direct referrals found.</p>
		</div>
	{/if}
</div>
