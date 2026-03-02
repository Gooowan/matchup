<!--
	ReferralTable Component
	
	A reusable component for displaying referral data with search and pagination.
	Uses existing types from app.d.ts and shadcn UI components.
	
	Props:
	- endpoint: string - API endpoint for referrals search (e.g., '/desim/referrals/search')
	- title?: string - Table title (default: 'Referrals')
	- showSearch?: boolean - Show search input (default: true)
	- showLevel?: boolean - Show referral level column (default: true)
	- itemsPerPage?: number - Items per page (default: 10)
	
	Features:
	- Search by name or email
	- Pagination with proper shadcn pagination component
	- Loading states and error handling
	- Avatar display with fallbacks
	- Level badges with different variants
	- Responsive design
	
	Usage:
	<ReferralTable 
		endpoint="/desim/referrals/search" 
		title="My Referrals" 
		showSearch={true}
		showLevel={true}
		itemsPerPage={10}
	/>
-->
<script lang="ts">
	import { authFetch } from '$utils/authFetch';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import ReferralDetailsModal from '$components/modals/ReferralDetailsModal.svelte';

	interface Props {
		endpoint: string; // e.g., '/desim/referrals/search'
		title?: string;
		showSearch?: boolean;
		showLevel?: boolean;
		itemsPerPage?: number;
	}

	let {
		endpoint,
		title = 'Referrals',
		showSearch = true,
		showLevel = true,
		itemsPerPage = 10,
	}: Props = $props();

	let referrals = $state<SearchUserReferralsDeepRow[]>([]);
	let pagination = $state<PaginationMeta>({
		page: 1,
		take: itemsPerPage,
		itemCount: 0,
		pageCount: 0,
	});
	let searchTerm = $state('');
	let isLoading = $state(false);
	let error = $state<string | null>(null);
	let dialogOpen = $state(false);
	let selectedUserId = $state<string | null>(null);
	let selectedDisplayName = $state<string>('');

	async function fetchReferrals(page: number = 1, search: string = '') {
		isLoading = true;
		error = null;

		try {
			const params = new URLSearchParams({
				page: page.toString(),
				take: itemsPerPage.toString(),
				...(search && { q: search }),
			});

			const resp = await authFetch(`${endpoint}?${params}`);
			const response: ApiPaginatedResponse<SearchUserReferralsDeepRow> = await resp.json();

			if (resp.status === 200) {
				referrals = response.data || [];
				pagination = {
					page: response.meta?.page || 1,
					take: response.meta?.take || itemsPerPage,
					itemCount: response.meta?.itemCount || 0,
					pageCount: response.meta?.pageCount || 0,
				};
			} else {
				error = response.error || 'Failed to fetch referrals';
			}
		} catch (err) {
			console.error('Referrals fetch error:', err);
			error = 'Failed to fetch referrals';
		} finally {
			isLoading = false;
		}
	}

	function handleSearch() {
		searchTerm = searchTerm.trim();
		fetchReferrals(1, searchTerm);
	}

	function handlePageChange(page: number) {
		fetchReferrals(page, searchTerm);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSearch();
		}
	}

	function handleRowClick(referral: SearchUserReferralsDeepRow) {
		selectedUserId = referral.id;
		selectedDisplayName = getDisplayName(referral.profile_data);
		dialogOpen = true;
	}

	function formatDate(timestamp: number): string {
		return new Date(timestamp).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}

	function getLevelBadgeVariant(
		level: number
	): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (level) {
			case 1:
				return 'default';
			case 2:
				return 'secondary';
			case 3:
				return 'outline';
			default:
				return 'outline';
		}
	}

	function getDisplayName(profileData: Record<string, any>): string {
		const firstName = profileData?.first_name || '';
		const lastName = profileData?.last_name || '';
		return `${firstName} ${lastName}`.trim() || 'Unknown User';
	}

	// Determine what to show for each row index
	function getRowData(index: number): 'loading' | 'error' | 'empty' | 'data' | 'filler' {
		if (isLoading) return 'loading';
		if (error) {
			return index === Math.floor(itemsPerPage / 2) ? 'error' : 'filler';
		}
		if (referrals.length === 0) {
			return index === Math.floor(itemsPerPage / 2) ? 'empty' : 'filler';
		}
		return index < referrals.length ? 'data' : 'filler';
	}

	// Initial load
	$effect(() => {
		fetchReferrals();
	});
</script>

<div class="space-y-6">
	<Card>
		<CardHeader>
			<CardTitle class="mb-2 flex h-6 items-center justify-between">
				<p>
					{title}
				</p>
				{#if isLoading}
					<Badge variant="secondary">Loading...</Badge>
				{:else if error}
					<Badge variant="destructive">Error</Badge>
				{/if}
			</CardTitle>
			{#if showSearch}
				<div class="flex gap-2">
					<Input
						bind:value={searchTerm}
						placeholder="Search by name or email..."
						onkeydown={handleKeydown}
						class="flex-1"
					/>
					<Button onclick={handleSearch} disabled={isLoading}>Search</Button>
				</div>
			{/if}
		</CardHeader>
		<CardContent>
			<!-- Mobile View -->
			<div class="sm:hidden">
			{#if isLoading}
				<div class="divide-border space-y-7 divide-y">
					{#each Array(itemsPerPage) as _}
						<div class="-m-2 mb-6 p-2 pb-6 last:mb-0 last:pb-0">
							<div class="w-full space-y-2">
								<div class="flex justify-between gap-2">
									<div class="flex items-center gap-2">
										<div class="h-9 w-9 animate-pulse rounded-full bg-gray-950/30"></div>
										<div class="space-y-1">
											<div class="h-4 w-32 animate-pulse rounded bg-gray-950/30"></div>
											<div class="h-3 w-48 animate-pulse rounded bg-gray-950/30"></div>
										</div>
									</div>
									{#if showLevel}
										<div class="h-5 w-16 animate-pulse rounded bg-gray-950/30"></div>
									{/if}
								</div>
								<div class="grid grid-cols-2 gap-4">
									<div class="space-y-1">
										<div class="h-3 w-16 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-24 animate-pulse rounded bg-gray-950/30"></div>
									</div>
									<div class="space-y-1">
										<div class="h-3 w-20 animate-pulse rounded bg-gray-950/30"></div>
										<div class="h-4 w-16 animate-pulse rounded bg-gray-950/30"></div>
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
				{:else if error}
					<div class="py-8 text-center">
						<p class="text-destructive mb-4">{error}</p>
						<Button onclick={() => fetchReferrals(pagination.page, searchTerm)}>Retry</Button>
					</div>
				{:else if referrals.length === 0}
					<div class="py-8 text-center">
						<p class="text-muted-foreground">
							{searchTerm ? 'No referrals found matching your search.' : 'No referrals found.'}
						</p>
					</div>
				{:else}
					<div class="divide-border space-y-7 divide-y">
						{#each referrals as referral}
							<div
								class="hover:bg-accent/50 -m-2 mb-6 flex cursor-pointer gap-6 space-y-6 p-2 pb-6 transition-colors last:mb-0 last:pb-0"
								onclick={() => handleRowClick(referral)}
								role="button"
								tabindex={0}
								onkeydown={(e) => e.key === 'Enter' && handleRowClick(referral)}
							>
								<div class="w-full space-y-2">
									<div class="flex justify-between gap-2">
										<div class="flex items-center gap-2">
											<Avatar.Root class="h-9 w-9">
												{#if referral.profile_data?.avatar}
													<Avatar.Image
														src={referral.profile_data.avatar}
														alt={getDisplayName(referral.profile_data)}
													/>
												{/if}
												<Avatar.Fallback>
													{getDisplayName(referral.profile_data).charAt(0).toUpperCase()}
												</Avatar.Fallback>
											</Avatar.Root>

											<div class="">
												<div class="flex gap-2">
													<div class="font-medium">{getDisplayName(referral.profile_data)}</div>
												</div>
												<div class="text-muted-foreground">{referral.email}</div>
											</div>
										</div>

										{#if showLevel}
											<div class="h-min">
												<Badge variant={getLevelBadgeVariant(referral.level)}>
													Line {referral.level}
												</Badge>
											</div>
										{/if}
									</div>

									<div class="grid grid-cols-2 gap-4">
										<div class="space-y-1">
											<div class="text-muted-foreground">Joined</div>
											<div class="font-medium">{formatDate(referral.created_at)}</div>
										</div>
										<div class="space-y-1">
											<div class="text-muted-foreground">Referral ID</div>
											<div class="font-medium">#{referral.referral_id}</div>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Desktop Table - Single table with conditional row content -->
			<Table.Root class="hidden table-fixed sm:table">
				<Table.Header>
					<Table.Row>
						<Table.Head class="w-[30%]">User</Table.Head>
						<Table.Head class="w-[25%]">Email</Table.Head>
						<Table.Head class="w-[15%]">Referral ID</Table.Head>
						{#if showLevel}
							<Table.Head class="w-[15%]">Line</Table.Head>
						{/if}
						<Table.Head class={showLevel ? 'w-[15%]' : 'w-[30%]'}>Joined</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each Array(itemsPerPage) as _, index}
						{@const rowType = getRowData(index)}
						{@const referral = rowType === 'data' ? referrals[index] : null}
						{@const colSpan = showLevel ? 5 : 4}

						{#if rowType === 'loading'}
							<Table.Row>
								<Table.Cell>
									<div class="flex items-center gap-3">
										<div class="h-8 w-8 animate-pulse rounded-full bg-gray-950/30"></div>
										<div class="h-4 w-24 animate-pulse rounded bg-gray-950/30"></div>
									</div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-4 w-32 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-5 w-16 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								{#if showLevel}
									<Table.Cell>
										<div class="h-5 w-12 animate-pulse rounded bg-gray-950/30"></div>
									</Table.Cell>
								{/if}
								<Table.Cell>
									<div class="h-4 w-20 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'error'}
							<Table.Row>
								<Table.Cell colspan={colSpan} class="text-center">
									<div class="py-4">
										<p class="text-destructive mb-4">{error}</p>
										<Button onclick={() => fetchReferrals(pagination.page, searchTerm)}>
											Retry
										</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'empty'}
							<Table.Row>
								<Table.Cell colspan={colSpan} class="text-center">
									<p class="text-muted-foreground">
										{searchTerm
											? 'No referrals found matching your search.'
											: 'No referrals found.'}
									</p>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'data' && referral}
							<Table.Row
								class="hover:bg-accent/50 cursor-pointer transition-colors"
								onclick={() => handleRowClick(referral)}
								role="button"
								tabindex={0}
								onkeydown={(e) => e.key === 'Enter' && handleRowClick(referral)}
							>
								<Table.Cell class="font-medium">
									<div class="flex items-center gap-3">
										<Avatar.Root class="h-8 w-8">
											{#if referral.profile_data?.avatar}
												<Avatar.Image
													src={referral.profile_data.avatar}
													alt={getDisplayName(referral.profile_data)}
												/>
											{/if}
											<Avatar.Fallback>
												{getDisplayName(referral.profile_data).charAt(0).toUpperCase()}
											</Avatar.Fallback>
										</Avatar.Root>
										<div>
											<div class="font-medium">{getDisplayName(referral.profile_data)}</div>
										</div>
									</div>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{referral.email}
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline">
										#{referral.referral_id}
									</Badge>
								</Table.Cell>
								{#if showLevel}
									<Table.Cell>
										<Badge variant={getLevelBadgeVariant(referral.level)}>
											Line {referral.level}
										</Badge>
									</Table.Cell>
								{/if}
								<Table.Cell class="text-muted-foreground">
									{formatDate(referral.created_at)}
								</Table.Cell>
							</Table.Row>
						{:else}
							<!-- Filler row -->
							<Table.Row class="h-[49px]">
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								{#if showLevel}
									<Table.Cell class="text-transparent">-</Table.Cell>
								{/if}
								<Table.Cell class="text-transparent">-</Table.Cell>
							</Table.Row>
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</CardContent>
	</Card>

	{#if referrals.length > 0 && pagination.pageCount > 1}
		<div class="flex justify-center">
			<Pagination.Root
				count={pagination.itemCount}
				perPage={pagination.take}
				bind:page={pagination.page}
				onPageChange={(page) => handlePageChange(page)}
			>
				{#snippet children({ pages, currentPage })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton>
								<ChevronLeftIcon class="size-4" />
								<span class="hidden sm:block">Previous</span>
							</Pagination.PrevButton>
						</Pagination.Item>
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link {page} isActive={currentPage === page.value}>
										{page.value}
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
</div>
<ReferralDetailsModal
	bind:isOpen={dialogOpen}
	userId={selectedUserId}
	displayName={selectedDisplayName}
/>
