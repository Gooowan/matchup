<script lang="ts">
	import { authFetch } from '$utils/authFetch';
	import { formatFileSize } from '$utils/format';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import FileIcon from '$svg/file.svg?component';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import SearchIcon from '@lucide/svelte/icons/search';
	import toast from 'svelte-french-toast';
	import { onMount } from 'svelte';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';

	interface Material {
		id: string;
		name: string;
		file_size: number;
		content_type: string;
		created_at: string;
	}

	interface PaginationMeta {
		page: number;
		take: number;
		itemCount: number;
		pageCount: number;
	}

	let materials = $state<Material[]>([]);
	let pagination = $state<PaginationMeta>({
		page: 1,
		take: 6,
		itemCount: 0,
		pageCount: 0
	});
	let searchTerm = $state('');
	let isLoading = $state(false);
	let error = $state<string | null>(null);
	let downloadingId = $state<string | null>(null);

	async function fetchMaterials(page: number = 1, search: string = '') {
		isLoading = true;
		error = null;

		try {
			const params = new URLSearchParams({
				page: page.toString(),
				take: '6',
				...(search && { q: search })
			});

			const resp = await authFetch(`/marketing?${params}`);
			const response = await resp.json();

			if (resp.status === 200) {
				materials = response.data || [];
				pagination = {
					page: response.meta?.page || 1,
					take: response.meta?.take || 6,
					itemCount: response.meta?.itemCount || 0,
					pageCount: response.meta?.pageCount || 0
				};
			} else {
				error = response.error || 'Failed to fetch materials';
			}
		} catch (err) {
			console.error('Materials fetch error:', err);
			error = 'Failed to load marketing materials';
		} finally {
			isLoading = false;
		}
	}

	function handleSearch() {
		searchTerm = searchTerm.trim();
		fetchMaterials(1, searchTerm);
	}

	function handlePageChange(page: number) {
		fetchMaterials(page, searchTerm);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSearch();
		}
	}

	async function handleDownload(material: Material) {
		if (downloadingId === material.id) return;

		downloadingId = material.id;
		try {
			const resp = await authFetch(`/marketing/${material.id}/download`);
			const response = await resp.json();

			if (resp.status === 200 && response.data?.url) {
				window.open(response.data.url, '_blank');
			} else {
				toast.error(response.error || 'Failed to generate download link');
			}
		} catch (err) {
			console.error('Download error:', err);
			toast.error('Failed to download file');
		} finally {
			downloadingId = null;
		}
	}

	onMount(() => {
		fetchMaterials();
	});
</script>

<div class="space-y-6">
	<!-- Section Header -->
	<div>
		<h2 class="text-2xl font-bold text-foreground">Marketing Materials</h2>
		<p class="text-muted-foreground mt-1">Download helpful resources and materials</p>
	</div>

	<!-- Search Bar -->
	<div class="flex gap-2">
		<Input
			bind:value={searchTerm}
			placeholder="Search materials by name..."
			onkeydown={handleKeydown}
			class="flex-1"
		/>
		<Button onclick={handleSearch} disabled={isLoading}>
			<SearchIcon class="h-4 w-4 mr-2" />
			Search
		</Button>
	</div>

	<!-- Materials Grid -->
	{#if isLoading}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			{#each Array(6) as _}
				<Card class="animate-pulse">
					<CardContent class="p-6">
						<div class="flex items-center gap-4">
							<div class="w-10 h-10 bg-muted rounded"></div>
							<div class="flex-1 space-y-2">
								<div class="h-4 bg-muted rounded w-3/4"></div>
								<div class="h-3 bg-muted rounded w-1/2"></div>
							</div>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{:else if error}
		<Card>
			<CardContent class="p-8 text-center">
				<p class="text-muted-foreground">{error}</p>
			</CardContent>
		</Card>
	{:else if materials.length === 0}
		<Card>
			<CardContent class="p-12 text-center">
				<FileIcon class="mx-auto h-12 w-12 text-muted-foreground opacity-50 mb-4" />
				<p class="text-muted-foreground">
					{searchTerm ? 'No materials found matching your search' : 'No marketing materials available'}
				</p>
			</CardContent>
		</Card>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			{#each materials as material (material.id)}
				<Card
					class="cursor-pointer transition-shadow hover:shadow-md group"
					onclick={() => handleDownload(material)}
				>
					<CardContent class="">
						<div class="flex items-center gap-4">
							<!-- File Icon -->
							<div class="flex-shrink-0">
								<FileIcon class="h-12 w-12" />
							</div>

							<!-- Material Info -->
							<div class="flex-1 min-w-0">
								<p class="font-semibold text-foreground truncate">{material.name}</p>
								<p class="text-sm text-muted-foreground">{formatFileSize(material.file_size)}</p>
							</div>

							<!-- Download Button -->
							<Button>
								<DownloadIcon
									class="h-5 w-5 text-white animate-pulse"
								/>
								<p class="">Download</p>
							</Button>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}

	<!-- Pagination -->
	{#if pagination.pageCount > 1 && !isLoading}
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

