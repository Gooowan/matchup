<script lang="ts">
	import * as Table from '$lib/components/ui/table/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import XIcon from '@lucide/svelte/icons/x';
	import HotwalletBalanceCard from '$lib/components/admin/HotwalletBalanceCard.svelte';
	import PendingWithdrawalsCard from '$lib/components/admin/PendingWithdrawalsCard.svelte';
	import { adminWithdrawsStore } from '$lib/stores/admin_withdraws.svelte';
	import { onMount } from 'svelte';
	import toast from 'svelte-french-toast';
	import { authFetch } from '$utils/authFetch';

	const itemsPerPage = 10;

	let pagination = $state<PaginationMeta>({
		page: 1,
		take: itemsPerPage,
		itemCount: 0,
		pageCount: 0,
	});

	let selectedTransaction = $state<Transaction | null>(null);
	let actionType = $state<'approve' | 'reject' | null>(null);
	let isActionLoading = $state(false);
	let isDialogOpen = $state(false);
	let pendingWithdrawals = $state<Record<string, number>>({ USDC: 0, USDT: 0 });
	let isPendingWithdrawalsLoading = $state(true);

	async function fetchData(page: number = 1) {
		const [transactionsResult, balanceResult] = await Promise.all([
			adminWithdrawsStore.fetchTransactions(page, itemsPerPage),
			adminWithdrawsStore.fetchHotwalletBalance(),
		]);

		if (transactionsResult?.meta) {
			pagination = {
				page: transactionsResult.meta.page || 1,
				take: transactionsResult.meta.take || itemsPerPage,
				itemCount: transactionsResult.meta.itemCount || 0,
				pageCount: transactionsResult.meta.pageCount || 0,
			};
		}
	}

	function handlePageChange(page: number) {
		fetchData(page);
	}

	async function fetchPendingWithdrawals() {
		isPendingWithdrawalsLoading = true;
		const desimResp = await authFetch('/admin/desim/stats');
		if (desimResp.status === 200) {
			const desimData = await desimResp.json();
			if (desimData.data?.pendingWithdrawals) {
				pendingWithdrawals = {
					USDC: desimData.data.pendingWithdrawals.USDC || 0,
					USDT: desimData.data.pendingWithdrawals.USDT || 0,
				};
			}
		}
		isPendingWithdrawalsLoading = false;
	}

	onMount(() => {
		fetchData();
		fetchPendingWithdrawals();
	});

	function formatCurrency(amount: number, decimals: number = 2): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: decimals,
		}).format(Math.abs(amount));
	}

	function formatDateTime(timestamp: number): string {
		const date = new Date(timestamp);
		const hours = date.getHours().toString().padStart(2, '0');
		const minutes = date.getMinutes().toString().padStart(2, '0');
		const day = date.getDate().toString().padStart(2, '0');
		const month = (date.getMonth() + 1).toString().padStart(2, '0');
		const year = date.getFullYear().toString().slice(-2);
		return `${hours}:${minutes} ${day}.${month}.${year}`;
	}

	function capitalize(text: string): string {
		if (!text) return text;
		return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
	}

	function getTokenFromTransaction(transaction: Transaction): string {
		return (transaction.metadata?.token as string) || 'UNKNOWN';
	}

	function getAddressFromTransaction(transaction: Transaction): string {
		return (transaction.metadata?.address as string) || 'N/A';
	}

	function canApprove(transaction: Transaction): boolean {
		const token = getTokenFromTransaction(transaction);
		return adminWithdrawsStore.hasSufficientBalance(transaction.amount, token);
	}

	function openApproveDialog(transaction: Transaction) {
		if (!canApprove(transaction)) {
			const token = getTokenFromTransaction(transaction);
			toast.error(`Insufficient ${token.toUpperCase()} balance in hotwallet`);
			return;
		}
		selectedTransaction = transaction;
		actionType = 'approve';
		isDialogOpen = true;
	}

	function openRejectDialog(transaction: Transaction) {
		selectedTransaction = transaction;
		actionType = 'reject';
		isDialogOpen = true;
	}

	function closeDialog() {
		if (!isActionLoading) {
			selectedTransaction = null;
			actionType = null;
			isDialogOpen = false;
		}
	}

	async function handleApprove() {
		if (!selectedTransaction) return;

		isActionLoading = true;
		try {
			await adminWithdrawsStore.approveWithdrawal(selectedTransaction.id);
			toast.success('Withdrawal approved successfully');

			// Refresh data
			await Promise.all([fetchData(pagination.page), fetchPendingWithdrawals()]);

			// Close dialog and reset state
			selectedTransaction = null;
			actionType = null;
			isDialogOpen = false;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Failed to approve withdrawal');
		} finally {
			isActionLoading = false;
		}
	}

	async function handleReject() {
		if (!selectedTransaction) return;

		isActionLoading = true;
		try {
			await adminWithdrawsStore.rejectWithdrawal(selectedTransaction.id);
			toast.success('Withdrawal rejected successfully');

			// Refresh transactions and pending withdrawals
			await Promise.all([
				adminWithdrawsStore.fetchTransactions(pagination.page, itemsPerPage),
				fetchPendingWithdrawals(),
			]);

			// Close dialog and reset state
			selectedTransaction = null;
			actionType = null;
			isDialogOpen = false;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Failed to reject withdrawal');
		} finally {
			isActionLoading = false;
		}
	}

	function getRowData(index: number): 'loading' | 'error' | 'empty' | 'data' | 'filler' {
		if (adminWithdrawsStore.isTransactionsLoading) return 'loading';
		if (adminWithdrawsStore.error) {
			return index === Math.floor(itemsPerPage / 2) ? 'error' : 'filler';
		}
		if (adminWithdrawsStore.transactions.length === 0) {
			return index === Math.floor(itemsPerPage / 2) ? 'empty' : 'filler';
		}
		return index < adminWithdrawsStore.transactions.length ? 'data' : 'filler';
	}

	function truncateAddress(address: string): string {
		if (address.length <= 12) return address;
		return `${address.slice(0, 6)}...${address.slice(-6)}`;
	}
</script>

<svelte:head>
	<title>Admin - Withdrawals</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<h1 class="text-foreground text-3xl font-bold">Withdrawals Management</h1>
		<p class="text-muted-foreground mt-2">Review and process pending withdrawal requests</p>
	</div>

	<!-- Hotwallet Balance and Pending Withdrawals Cards -->
	<div class="mb-6 grid gap-4 md:grid-cols-2">
		<HotwalletBalanceCard />
		<PendingWithdrawalsCard isLoading={isPendingWithdrawalsLoading} {pendingWithdrawals} />
	</div>

	<!-- Mobile Pagination (top) -->
	{#if adminWithdrawsStore.transactions.length > 0 && pagination.pageCount > 1}
		<div class="mb-4 flex justify-center md:hidden">
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

	<!-- Transactions Table -->
	<Card>
		<CardContent>
			<!-- Mobile View -->
			<div class="sm:hidden">
				{#if adminWithdrawsStore.isTransactionsLoading}
					<div class="divide-border space-y-7 divide-y">
						{#each Array(itemsPerPage) as _}
							<div class="space-y-6 pb-7 last:pb-0">
								<div class="flex items-center justify-between">
									<div class="h-4 w-20 animate-pulse rounded bg-gray-950/30"></div>
									<div class="h-6 w-16 animate-pulse rounded bg-gray-950/30"></div>
								</div>
								<div class="flex flex-col items-center justify-between">
									<div class="h-8 w-40 animate-pulse rounded bg-gray-950/30"></div>
									<div class="mt-2 h-3 w-24 animate-pulse rounded bg-gray-950/30"></div>
								</div>
								<div class="flex gap-2">
									<div class="h-9 flex-1 animate-pulse rounded bg-gray-950/30"></div>
									<div class="h-9 flex-1 animate-pulse rounded bg-gray-950/30"></div>
								</div>
							</div>
						{/each}
					</div>
				{:else if adminWithdrawsStore.error}
					<div class="py-8 text-center">
						<p class="text-destructive">{adminWithdrawsStore.error}</p>
						<Button onclick={() => fetchData(pagination.page)} class="mt-2">Retry</Button>
					</div>
				{:else if adminWithdrawsStore.transactions.length === 0}
					<div class="py-8 text-center">
						<p class="text-muted-foreground">No pending withdrawals</p>
					</div>
				{:else}
					<div class="divide-border space-y-7 divide-y">
						{#each adminWithdrawsStore.transactions as transaction}
							<div class="space-y-6 pb-7 last:pb-0">
								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium">{capitalize(transaction.type)}</p>
										<p class="text-muted-foreground text-xs">
											{getTokenFromTransaction(transaction)}
										</p>
									</div>
									<Badge variant="secondary">{capitalize(transaction.status)}</Badge>
								</div>
								<div class="flex flex-col items-center justify-between">
									<div class="text-2xl font-bold">
										{formatCurrency(transaction.amount, 2)}
										{getTokenFromTransaction(transaction)}
									</div>
									<div class="text-muted-foreground/60 text-sm">
										{formatDateTime(transaction.created_at)}
									</div>
									<div class="text-muted-foreground mt-2 text-xs">
										To: {truncateAddress(getAddressFromTransaction(transaction))}
									</div>
								</div>
								<div class="flex gap-2">
									<Button
										variant="default"
										size="sm"
										class="flex-1"
										onclick={() => openApproveDialog(transaction)}
										disabled={!canApprove(transaction)}
									>
										<CheckIcon class="mr-1 size-4" />
										Approve
									</Button>
									<Button
										variant="destructive"
										size="sm"
										class="flex-1"
										onclick={() => openRejectDialog(transaction)}
									>
										<XIcon class="mr-1 size-4" />
										Reject
									</Button>
								</div>
								{#if !canApprove(transaction)}
									<p class="text-destructive text-center text-xs">
										Insufficient {getTokenFromTransaction(transaction)} balance
									</p>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Desktop Table View -->
			<Table.Root class="hidden table-fixed sm:table">
				<Table.Header>
					<Table.Row>
						<!-- <Table.Head class="w-[12%]">Type</Table.Head> -->
						<Table.Head class="w-[12%]">Amount</Table.Head>
						<Table.Head class="w-[10%]">Token</Table.Head>
						<Table.Head class="w-[20%]">Address</Table.Head>
						<Table.Head class="w-[12%]">Status</Table.Head>
						<Table.Head class="w-[14%]">Date</Table.Head>
						<Table.Head class="w-[20%]">Actions</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each Array(itemsPerPage) as _, index}
						{@const rowType = getRowData(index)}
						{@const transaction =
							rowType === 'data' ? adminWithdrawsStore.transactions[index] : null}

						{#if rowType === 'loading'}
							<Table.Row>
								<!-- <Table.Cell>
									<div class="h-[22px] w-3/4 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell> -->
								<Table.Cell>
									<div class="h-[22px] w-2/3 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-[22px] w-16 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-[22px] w-full animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-[22px] w-20 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-[22px] w-24 animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
								<Table.Cell>
									<div class="h-[22px] w-full animate-pulse rounded bg-gray-950/30"></div>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'error'}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center">
									<div class="py-4">
										<p class="text-destructive">{adminWithdrawsStore.error}</p>
										<Button onclick={() => fetchData(pagination.page)} class="mt-2">Retry</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'empty'}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center">
									<p class="text-muted-foreground">No pending withdrawals</p>
								</Table.Cell>
							</Table.Row>
						{:else if rowType === 'data' && transaction}
							<Table.Row>
								<!-- <Table.Cell class="font-medium">
									{capitalize(transaction.type)}
								</Table.Cell> -->
								<Table.Cell>
									{formatCurrency(transaction.amount, 2)}
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline">
										{getTokenFromTransaction(transaction)}
									</Badge>
								</Table.Cell>
								<Table.Cell class="font-mono text-xs">
									<span title={getAddressFromTransaction(transaction)}>
										{truncateAddress(getAddressFromTransaction(transaction))}
									</span>
								</Table.Cell>
								<Table.Cell>
									<Badge variant="secondary">
										{capitalize(transaction.status)}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{formatDateTime(transaction.created_at)}
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-2">
										<Button
											variant="default"
											size="sm"
											onclick={() => openApproveDialog(transaction)}
											disabled={!canApprove(transaction)}
											title={!canApprove(transaction)
												? `Insufficient ${getTokenFromTransaction(transaction)} balance`
												: 'Approve withdrawal'}
										>
											<CheckIcon class="mr-1 size-4" />
											Approve
										</Button>
										<Button
											variant="destructive"
											size="sm"
											onclick={() => openRejectDialog(transaction)}
										>
											<XIcon class="mr-1 size-4" />
											Reject
										</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<!-- Filler row -->
							<Table.Row class="h-[53px]">
								<!-- <Table.Cell class="text-transparent">-</Table.Cell> -->
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
								<Table.Cell class="text-transparent">-</Table.Cell>
							</Table.Row>
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</CardContent>
	</Card>

	<!-- Pagination (bottom) -->
	{#if adminWithdrawsStore.transactions.length > 0 && pagination.pageCount > 1}
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
</div>

<!-- Confirmation Dialog -->
<ResponsiveDialog.Root bind:open={isDialogOpen} onOpenChange={closeDialog}>
	<ResponsiveDialog.Content>
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>
				{actionType === 'approve' ? 'Approve Withdrawal' : 'Reject Withdrawal'}
			</ResponsiveDialog.Title>
			<ResponsiveDialog.Description>
				{#if selectedTransaction}
					{#if actionType === 'approve'}
						Are you sure you want to approve this withdrawal of
						<strong
							>{formatCurrency(selectedTransaction.amount, 2)}
							USD
						</strong>
						in <strong>{getTokenFromTransaction(selectedTransaction)}</strong>
						to address
						<strong class="break-all">{getAddressFromTransaction(selectedTransaction)}</strong>?
						<br /><br />
						This will execute the blockchain transaction and cannot be undone.
					{:else}
						Are you sure you want to reject this withdrawal of
						<strong
							>{formatCurrency(selectedTransaction.amount, 2)}
							{getTokenFromTransaction(selectedTransaction)}</strong
						>?
						<br /><br />
						The funds will be returned to the user's balance.
					{/if}
				{/if}
			</ResponsiveDialog.Description>
		</ResponsiveDialog.Header>
		<ResponsiveDialog.Footer>
			<Button variant="outline" onclick={closeDialog} disabled={isActionLoading} class="flex-1">
				Cancel
			</Button>
			<Button
				onclick={actionType === 'approve' ? handleApprove : handleReject}
				disabled={isActionLoading}
				class={actionType === 'approve' ? 'bg-primary flex-1' : 'bg-destructive flex-1'}
			>
				{#if isActionLoading}
					Processing...
				{:else}
					{actionType === 'approve' ? 'Approve' : 'Reject'}
				{/if}
			</Button>
		</ResponsiveDialog.Footer>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>
