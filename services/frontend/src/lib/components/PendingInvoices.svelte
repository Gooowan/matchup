<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Pagination from '$lib/components/ui/pagination';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import { authFetch } from '$lib/utils/authFetch';
	import Button from './ui/button/button.svelte';

	interface Props {
		product: string;
		onInvoiceClick?: (invoice: Invoice) => void;
		hasPendingInvoices: boolean;
	}

	let { product, onInvoiceClick, hasPendingInvoices = $bindable() }: Props = $props();

	let pendingInvoices = $state<Invoice[]>([]);
	let invoicePagination = $state<PaginationMeta>({
		page: 1,
		take: 2,
		itemCount: 0,
		pageCount: 0,
	});

	async function fetchPendingInvoices(page: number = 1) {
		const params = new URLSearchParams({
			page: page.toString(),
			take: invoicePagination.take.toString(),
		});

		const response = await authFetch(`/payments/${product}/pending?${params}`);
		if (!response.ok) {
			return;
		}

		const data: ApiPaginatedResponse<Invoice> = await response.json();
		if (!data.data) {
			return;
		}

		pendingInvoices = data.data;
		invoicePagination = {
			page: data.meta?.page || 1,
			take: data.meta?.take || 2,
			itemCount: data.meta?.itemCount || 0,
			pageCount: data.meta?.pageCount || 0,
		};
	}

	function handleInvoiceClick(invoice: Invoice) {
		onInvoiceClick?.(invoice);
	}

	function handleInvoicePageChange(page: number) {
		fetchPendingInvoices(page);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	$effect(() => {
		if (hasPendingInvoices && pendingInvoices.length == 0) {
			fetchPendingInvoices();
		}
	});
</script>

{#if hasPendingInvoices && pendingInvoices.length > 0}
	<div class="space-y-4">
		<div class="flex flex-col gap-4 ">
			{#each pendingInvoices as invoice}
				<Card.Root class="cursor-pointer p-4" onclick={() => handleInvoiceClick(invoice)}>
						<div class="flex items-center justify-between">
							<Card.Title class="text-xl font-bold">
								{invoice.amount}
								{invoice.token}
							</Card.Title>
							<Button onclick={() => handleInvoiceClick(invoice)}>Pay</Button>
						</div>
						<div class="grid grid-cols-2 space-y-4 border-t pt-4">
							<div class="space-y-1">
								<p class="text-muted-foreground text-sm">Chain:</p>
								<p class="font-medium">{invoice.chain}</p>
							</div>
							<div class="space-y-1 text-sm">
								<p class="text-muted-foreground">Created:</p>
								<p>{formatDate(invoice.created_at)}</p>
							</div>
						</div>
				</Card.Root>
			{/each}
		</div>

		{#if invoicePagination.pageCount > 1}
			<div class="flex justify-center">
				<Pagination.Root
					count={invoicePagination.itemCount}
					perPage={invoicePagination.take}
					bind:page={invoicePagination.page}
					onPageChange={(page) => handleInvoicePageChange(page)}
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
{:else}
	<div class="py-8 text-center">
		<p class="text-muted-foreground">No pending invoices found.</p>
	</div>
{/if}
