<script lang="ts">
	import * as Table from '$lib/components/ui/table/index.js';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import UserIcon from '@lucide/svelte/icons/user';
	import ReferralDetailsModal from '$lib/components/modals/ReferralDetailsModal.svelte';
	import { onMount } from 'svelte';

	interface Props {
		store: {
			transactions: Transaction[];
			currency: string;
			isTransactionsLoading: boolean;
			error: string | null;
			fetchTransactions: (page?: number, take?: number) => Promise<any>;
		};
		itemsPerPage?: number;
	}

	let { store, itemsPerPage = 10 }: Props = $props();

	let pagination = $state<PaginationMeta>({
		page: 1,
		take: itemsPerPage,
		itemCount: 0,
		pageCount: 0,
	});

	let dialogOpen = $state(false);
	let selectedUserId = $state<string | null>(null);
	let selectedDisplayName = $state<string>('');
	let isMissedProfit = $state(false);

	async function fetchTransactions(page: number = 1) {
		const result = await store.fetchTransactions(page, itemsPerPage);
		if (result?.meta) {
			pagination = {
				page: result.meta.page || 1,
				take: result.meta.take || itemsPerPage,
				itemCount: result.meta.itemCount || 0,
				pageCount: result.meta.pageCount || 0,
			};
		}
	}

	function handlePageChange(page: number) {
		fetchTransactions(page);
	}

	onMount(() => {
		fetchTransactions();
	});

	function formatCurrency(currency: string, amount: number): string {
		return (
			new Intl.NumberFormat('en-US', {
				minimumFractionDigits: 0,
				maximumFractionDigits: 9,
			}).format(Math.abs(amount)) +
			' ' +
			currency
		);
	}

	function getTransactionColor(transaction: Transaction): string {
		if (transaction.type === 'MISSED_PROFIT') return 'text-yellow-600';
		if (transaction.amount > 0 && transaction.status === 'SUCCESS') return 'text-green-600';
		return '';
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
		if (text === 'MISSED_PROFIT') return 'Missed Profit';
		return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
	}

	function getRowData(index: number): 'loading' | 'error' | 'empty' | 'data' | 'filler' {
		if (store.isTransactionsLoading) return 'loading';
		if (store.error) {
			return index === Math.floor(itemsPerPage / 2) ? 'error' : 'filler';
		}
		if (store.transactions.length === 0) {
			return index === Math.floor(itemsPerPage / 2) ? 'empty' : 'filler';
		}
		return index < store.transactions.length ? 'data' : 'filler';
	}

	function isTransactionClickable(transaction: Transaction): boolean {
		return transaction.initiator_id !== transaction.owner_id || transaction.type === 'MISSED_PROFIT';
	}

	function handleTransactionClick(transaction: Transaction) {
		if (isTransactionClickable(transaction)) {
			selectedUserId = transaction.initiator_id;
			selectedDisplayName = ''; // We don't have the name yet, will be fetched by modal
			isMissedProfit = transaction.type === 'MISSED_PROFIT';
			dialogOpen = true;
		}
	}

	// Reset isMissedProfit when dialog closes
	$effect(() => {
		if (!dialogOpen) {
			isMissedProfit = false;
		}
	});
</script>

{#if store.transactions.length > 0 && pagination.pageCount > 1}
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

<Card>
	<CardContent>
		<!-- Mobile View - Keep separate for different layout -->
		<div class="sm:hidden">
			{#if store.isTransactionsLoading}
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
						</div>
					{/each}
				</div>
			{:else if store.error}
				<div class="py-8 text-center">
					<p class="text-destructive">{store.error}</p>
					<button
						onclick={() => fetchTransactions(pagination.page)}
						class="bg-primary text-primary-foreground hover:bg-primary/90 mt-2 rounded px-4 py-2"
					>
						Retry
					</button>
				</div>
			{:else if store.transactions.length === 0}
				<div class="py-8 text-center">
					<p class="text-muted-foreground">No transactions found</p>
				</div>
			{:else}
				<div class="divide-border space-y-7 divide-y">
					{#each store.transactions as transaction}
						<div 
							class={`space-y-6 pb-7 last:pb-0 ${isTransactionClickable(transaction) ? 'cursor-pointer hover:bg-accent/50 rounded-lg p-2 -m-2 transition-colors' : ''}`}
							onclick={() => handleTransactionClick(transaction)}
							role={isTransactionClickable(transaction) ? 'button' : undefined}
							tabindex={isTransactionClickable(transaction) ? 0 : undefined}
							onkeydown={(e) => isTransactionClickable(transaction) && e.key === 'Enter' && handleTransactionClick(transaction)}
						>
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									{capitalize(transaction.type)}
									{#if isTransactionClickable(transaction)}
										<UserIcon class="size-4 text-muted-foreground" />
									{/if}
								</div>
                                <Badge variant={transaction.status === 'SUCCESS' ? 'default' : 'secondary'}>
                                    {capitalize(transaction.status)}
                                </Badge>
							</div>
							<div class="flex flex-col items-center justify-between">
								<div 
                                    class={getTransactionColor(transaction)}
                                >
									{formatCurrency(store.currency, transaction.amount)}
								</div>
								<div class="text-muted-foreground/60 text-sm">
									{formatDateTime(transaction.created_at)}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<Table.Root class="hidden table-fixed sm:table">
			<Table.Header>
				<Table.Row>
					<Table.Head class="w-[25%]">Type</Table.Head>
					<Table.Head class="w-[35%]">Amount</Table.Head>
					<Table.Head class="w-[20%]">Status</Table.Head>
					<Table.Head class="w-[20%]">Date</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each Array(itemsPerPage) as _, index}
					{@const rowType = getRowData(index)}
					{@const transaction = rowType === 'data' ? store.transactions[index] : null}

					{#if rowType === 'loading'}
						<Table.Row>
							<Table.Cell>
								<div class="h-[22px] w-3/4 animate-pulse rounded bg-gray-950/30"></div>
							</Table.Cell>
							<Table.Cell>
								<div class="h-[22px] w-2/3 animate-pulse rounded bg-gray-950/30"></div>
							</Table.Cell>
							<Table.Cell>
								<div class="h-[22px] w-20 animate-pulse rounded bg-gray-950/30"></div>
							</Table.Cell>
							<Table.Cell>
								<div class="h-[22px] w-24 animate-pulse rounded bg-gray-950/30"></div>
							</Table.Cell>
						</Table.Row>
					{:else if rowType === 'error'}
						<Table.Row>
							<Table.Cell colspan={4} class="text-center">
								<div class="py-4">
									<p class="text-destructive">{store.error}</p>
									<button
										onclick={() => fetchTransactions(pagination.page)}
										class="bg-primary text-primary-foreground hover:bg-primary/90 mt-2 rounded px-4 py-2"
									>
										Retry
									</button>
								</div>
							</Table.Cell>
						</Table.Row>
					{:else if rowType === 'empty'}
						<Table.Row>
							<Table.Cell colspan={4} class="text-center">
								<p class="text-muted-foreground">No transactions found</p>
							</Table.Cell>
						</Table.Row>
					{:else if rowType === 'data' && transaction}
						<Table.Row 
							class={isTransactionClickable(transaction) ? 'cursor-pointer hover:bg-accent/50 transition-colors' : ''}
							onclick={() => handleTransactionClick(transaction)}
							role={isTransactionClickable(transaction) ? 'button' : undefined}
							tabindex={isTransactionClickable(transaction) ? 0 : undefined}
							onkeydown={(e) => isTransactionClickable(transaction) && e.key === 'Enter' && handleTransactionClick(transaction)}
						>
							<Table.Cell class="font-medium">
								<div class="flex items-center gap-2">
									{capitalize(transaction.type)}
									{#if isTransactionClickable(transaction)}
										<UserIcon class="size-4 text-muted-foreground" />
									{/if}
								</div>
							</Table.Cell>
							<Table.Cell 
								class={getTransactionColor(transaction)}
							>
								{formatCurrency(store.currency, transaction.amount)}
							</Table.Cell>
							<Table.Cell>
								<Badge variant={transaction.status === 'SUCCESS' ? 'default' : 'secondary'}>
									{capitalize(transaction.status)}
								</Badge>
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{formatDateTime(transaction.created_at)}
							</Table.Cell>
						</Table.Row>
					{:else}
						<!-- Filler row -->
						<Table.Row class="h-[39px]">
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

{#if store.transactions.length > 0 && pagination.pageCount > 1}
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

<ReferralDetailsModal 
	bind:isOpen={dialogOpen} 
	userId={selectedUserId} 
	displayName={selectedDisplayName}
	isMissedProfit={isMissedProfit}
/>
