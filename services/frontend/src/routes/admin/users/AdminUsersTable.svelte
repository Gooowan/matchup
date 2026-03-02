<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$utils/authFetch';
	import AdminEditDialog from './AdminUserEditDialog.svelte';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import EditIcon from '@lucide/svelte/icons/edit';

	interface Props {
		title?: string;
		showSearch?: boolean;
		itemsPerPage?: number;
	}

	let { title = 'Users', showSearch = true, itemsPerPage = 10 }: Props = $props();

	let users = $state<AdminUserDTO[]>([]);
	let pagination = $state<PaginationMeta>({
		page: 1,
		take: itemsPerPage,
		itemCount: 0,
		pageCount: 0,
	});
	let searchTerm = $state('');
	let isLoading = $state(false);
	let error = $state<string | null>(null);
	let editDialogOpen = $state(false);
	let selectedUser = $state<AdminUserDTO | null>(null);

	async function fetchUsers(page: number = 1, search: string = '') {
		isLoading = true;
		error = null;

		const params = new URLSearchParams({
			page: page.toString(),
			take: itemsPerPage.toString(),
			...(search && { q: search }),
		});

		const resp = await authFetch(`/admin/users/search?${params}`);
		if (!resp.ok) {
			error = 'Failed to fetch users';
			console.error('Users fetch error:', resp.status, resp.statusText);
			isLoading = false;
			return;
		}

		const response: ApiPaginatedResponse<AdminUserDTO> = await resp.json();
		if (response.error) {
			error = response.error;
			console.error('Users API error:', response.error);
			isLoading = false;
			return;
		}

		users = response.data || [];
		pagination = {
			page: response.meta?.page || 1,
			take: response.meta?.take || itemsPerPage,
			itemCount: response.meta?.itemCount || 0,
			pageCount: response.meta?.pageCount || 0,
		};
		isLoading = false;
	}

	function handleSearch() {
		searchTerm = searchTerm.trim();
		fetchUsers(1, searchTerm);
	}

	function handlePageChange(page: number) {
		fetchUsers(page, searchTerm);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSearch();
		}
	}

	function formatDate(timestamp: number): string {
		return new Date(timestamp).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}

	function getRoleBadgeVariant(role: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (role) {
			case 'ADMIN':
				return 'destructive';
			case 'USER':
				return 'default';
			default:
				return 'outline';
		}
	}

	function getDisplayName(profileData: Record<string, any>): string {
		const firstName = profileData?.first_name || '';
		const lastName = profileData?.last_name || '';
		return `${firstName} ${lastName}`.trim() || 'Unknown User';
	}

	function handleEditUser(user: AdminUserDTO) {
		selectedUser = user;
		editDialogOpen = true;
	}

	function handleDialogClose() {
		editDialogOpen = false;
		selectedUser = null;
	}

	function handleUserSaved() {
		fetchUsers(pagination.page, searchTerm);
	}

	onMount(() => {
		fetchUsers();
	});
</script>

<Card>
	<CardHeader>
		<CardTitle class="mb-2 flex items-center justify-between">
			{title}
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
					placeholder="Search by name, email, or referral ID..."
					onkeydown={handleKeydown}
					class="flex-1"
				/>
				<Button onclick={handleSearch} disabled={isLoading}>Search</Button>
			</div>
		{/if}
	</CardHeader>
	<CardContent>
		{#if isLoading}
			<div class="space-y-3 p-6">
				{#each Array(5) as _}
					<div class="bg-muted h-16 animate-pulse rounded"></div>
				{/each}
			</div>
		{:else if error}
			<div class="py-8 text-center">
				<p class="text-destructive mb-4">{error}</p>
				<Button onclick={() => fetchUsers(pagination.page, searchTerm)}>Retry</Button>
			</div>
		{:else if users.length > 0}
			<div class="overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="min-w-[200px]">User</Table.Head>
							<Table.Head class="hidden min-w-[200px] md:table-cell">Email</Table.Head>
							<Table.Head class="hidden min-w-[120px] sm:table-cell">Referral ID</Table.Head>
							<Table.Head class="hidden min-w-[80px] lg:table-cell">Role</Table.Head>
							<Table.Head class="hidden min-w-[100px] lg:table-cell">Joined</Table.Head>
							<Table.Head
								class=" sticky right-0 min-w-[60px] shadow-[-4px_0_8px_-4px_rgba(0,0,0,0.1)]"
								>Actions</Table.Head
							>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each users as user}
							<Table.Row>
								<Table.Cell class="min-w-[200px] font-medium">
									<div class="flex items-center gap-3">
										<Avatar.Root class="h-8 w-8 shrink-0">
											{#if user.profile_data?.avatar}
												<Avatar.Image
													src={user.profile_data.avatar}
													alt={getDisplayName(user.profile_data)}
												/>
											{/if}
											<Avatar.Fallback>
												{getDisplayName(user.profile_data).charAt(0).toUpperCase()}
											</Avatar.Fallback>
										</Avatar.Root>
										<div class="min-w-0">
											<div class="truncate font-medium">{getDisplayName(user.profile_data)}</div>
											<div class="text-muted-foreground truncate text-sm md:hidden">
												{user.email}
											</div>
											<div class="mt-1 flex gap-2 sm:hidden">
												<Badge variant="outline" class="text-xs">
													#{user.referral_id}
												</Badge>
												<Badge variant={getRoleBadgeVariant(user.role)} class="text-xs lg:hidden">
													{user.role}
												</Badge>
											</div>
										</div>
									</div>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground hidden min-w-[200px] md:table-cell">
									<div class="truncate">{user.email}</div>
								</Table.Cell>
								<Table.Cell class="hidden min-w-[120px] sm:table-cell">
									<Badge variant="outline">
										#{user.referral_id}
									</Badge>
								</Table.Cell>
								<Table.Cell class="hidden min-w-[80px] lg:table-cell">
									<Badge variant={getRoleBadgeVariant(user.role)}>
										{user.role}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground hidden min-w-[100px] lg:table-cell">
									{formatDate(user.created_at)}
								</Table.Cell>
								<Table.Cell
									class=" sticky right-0 min-w-[60px] shadow-[-4px_0_8px_-4px_rgba(0,0,0,0.1)]"
								>
									<Button
										variant="ghost"
										size="sm"
										onclick={() => handleEditUser(user)}
										class="h-8 w-8 p-0"
									>
										<EditIcon class="h-4 w-4" />
									</Button>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{:else}
			<div class="py-8 text-center">
				<p class="text-muted-foreground">
					{searchTerm ? 'No users found matching your search.' : 'No users found.'}
				</p>
			</div>
		{/if}
	</CardContent>
</Card>

{#if users.length > 0 && pagination.pageCount > 1}
	<div class="mt-4 flex justify-center">
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

<AdminEditDialog
	user={selectedUser}
	open={editDialogOpen}
	onClose={handleDialogClose}
	onSaved={handleUserSaved}
/>
