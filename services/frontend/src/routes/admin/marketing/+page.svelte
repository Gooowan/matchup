<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$utils/authFetch';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import TrashIcon from '@lucide/svelte/icons/trash';
	import EditIcon from '@lucide/svelte/icons/edit';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import FileIcon from '@lucide/svelte/icons/file';
	import toast from 'svelte-french-toast';

	interface MarketingMaterial {
		id: string;
		name: string;
		file_size: number;
		content_type: string;
		visible: boolean;
		created_at: string;
		updated_at?: string;
	}

	interface PaginationMeta {
		page: number;
		take: number;
		itemCount: number;
		pageCount: number;
	}

	let materials = $state<MarketingMaterial[]>([]);
	let pagination = $state<PaginationMeta>({
		page: 1,
		take: 10,
		itemCount: 0,
		pageCount: 0
	});
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// Upload dialog
	let uploadDialogOpen = $state(false);
	let uploadFile: File | null = $state(null);
	let uploadName = $state('');
	let isUploading = $state(false);

	// Edit dialog
	let editDialogOpen = $state(false);
	let editMaterial = $state<MarketingMaterial | null>(null);
	let editName = $state('');
	let isEditing = $state(false);

	async function fetchMaterials(page: number = 1) {
		isLoading = true;
		error = null;

		const params = new URLSearchParams({
			page: page.toString(),
			take: pagination.take.toString()
		});

		const resp = await authFetch(`/admin/marketing?${params}`);
		if (!resp.ok) {
			error = 'Failed to fetch materials';
			isLoading = false;
			return;
		}

		const response = await resp.json();
		if (response.error) {
			error = response.error;
			isLoading = false;
			return;
		}

		materials = response.data || [];
		pagination = {
			page: response.meta?.page || 1,
			take: response.meta?.take || 10,
			itemCount: response.meta?.itemCount || 0,
			pageCount: response.meta?.pageCount || 0
		};
		isLoading = false;
	}

	function handlePageChange(page: number) {
		fetchMaterials(page);
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function openUploadDialog() {
		uploadFile = null;
		uploadName = '';
		uploadDialogOpen = true;
	}

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			uploadFile = target.files[0];
			if (!uploadName) {
				uploadName = target.files[0].name;
			}
		}
	}

	async function handleUpload() {
		if (!uploadFile || !uploadName.trim()) {
			toast.error('Please select a file and enter a name');
			return;
		}

		isUploading = true;
		const formData = new FormData();
		formData.append('file', uploadFile);
		formData.append('name', uploadName.trim());

		const resp = await authFetch('/admin/marketing/upload', {
			method: 'POST',
			body: formData
		});

		if (resp.ok) {
			const response = await resp.json();
			if (response.error) {
				toast.error(response.error);
			} else {
				toast.success('Material uploaded successfully');
				uploadDialogOpen = false;
				fetchMaterials(pagination.page);
			}
		} else {
			toast.error('Failed to upload material');
		}

		isUploading = false;
	}

	function openEditDialog(material: MarketingMaterial) {
		editMaterial = material;
		editName = material.name;
		editDialogOpen = true;
	}

	async function handleSaveName() {
		if (!editMaterial || !editName.trim()) {
			toast.error('Please enter a name');
			return;
		}

		isEditing = true;
		const resp = await authFetch(`/admin/marketing/${editMaterial.id}/name`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name: editName.trim() })
		});

		if (resp.ok) {
			toast.success('Material name updated');
			editDialogOpen = false;
			fetchMaterials(pagination.page);
		} else {
			toast.error('Failed to update material name');
		}

		isEditing = false;
	}

	async function toggleVisibility(material: MarketingMaterial) {
		const newVisibility = !material.visible;

		const resp = await authFetch(`/admin/marketing/${material.id}/visibility`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ visible: newVisibility })
		});

		if (resp.ok) {
			toast.success(
				newVisibility ? 'Material is now visible to users' : 'Material is now hidden from users'
			);
			fetchMaterials(pagination.page);
		} else {
			toast.error('Failed to update visibility');
		}
	}

	async function deleteMaterial(material: MarketingMaterial) {
		if (
			!confirm(
				`Are you sure you want to delete "${material.name}"? This action cannot be undone.`
			)
		) {
			return;
		}

		const resp = await authFetch(`/admin/marketing/${material.id}`, {
			method: 'DELETE'
		});

		if (resp.ok) {
			toast.success('Material deleted successfully');
			fetchMaterials(pagination.page);
		} else {
			toast.error('Failed to delete material');
		}
	}

	onMount(() => {
		fetchMaterials();
	});
</script>

<div class="container mx-auto px-4 py-8 space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold">Marketing Materials</h1>
			<p class="text-muted-foreground">Manage downloadable materials for users</p>
		</div>
		<Button onclick={openUploadDialog}>
			<UploadIcon class="mr-2 h-4 w-4" />
			Upload Material
		</Button>
	</div>

	<Card>
		<CardHeader>
			<CardTitle class="flex items-center justify-between">
				Materials
				{#if isLoading}
					<Badge variant="secondary">Loading...</Badge>
				{:else if error}
					<Badge variant="destructive">Error</Badge>
				{:else}
					<Badge variant="outline">{pagination.itemCount} total</Badge>
				{/if}
			</CardTitle>
		</CardHeader>
		<CardContent>
			{#if error}
				<div class="py-8 text-center text-muted-foreground">
					<p>Error: {error}</p>
				</div>
			{:else if materials.length === 0 && !isLoading}
				<div class="py-12 text-center text-muted-foreground">
					<FileIcon class="mx-auto mb-4 h-12 w-12 opacity-50" />
					<p>No marketing materials yet</p>
					<p class="text-sm">Upload your first material to get started</p>
				</div>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Name</Table.Head>
							<Table.Head>File Size</Table.Head>
							<Table.Head>Type</Table.Head>
							<Table.Head>Visibility</Table.Head>
							<Table.Head>Created</Table.Head>
							<Table.Head class="text-right">Actions</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each materials as material (material.id)}
							<Table.Row>
								<Table.Cell class="font-medium">
									{material.name}
								</Table.Cell>
								<Table.Cell>
									{formatFileSize(material.file_size)}
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline">{material.content_type}</Badge>
								</Table.Cell>
								<Table.Cell>
									<Button
										variant="ghost"
										size="sm"
										onclick={() => toggleVisibility(material)}
										class="gap-2"
									>
										{#if material.visible}
											<EyeIcon class="h-4 w-4 text-green-600" />
											<span class="text-green-600">Visible</span>
										{:else}
											<EyeOffIcon class="h-4 w-4 text-gray-400" />
											<span class="text-gray-400">Hidden</span>
										{/if}
									</Button>
								</Table.Cell>
								<Table.Cell>
									{formatDate(material.created_at)}
								</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex justify-end gap-2">
										<Button
											variant="outline"
											size="sm"
											onclick={() => openEditDialog(material)}
										>
											<EditIcon class="h-4 w-4" />
										</Button>
										<Button
											variant="destructive"
											size="sm"
											onclick={() => deleteMaterial(material)}
										>
											<TrashIcon class="h-4 w-4" />
										</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>

				{#if pagination.pageCount > 1}
					<div class="mt-4 flex items-center justify-center">
						<Pagination.Root
							count={pagination.itemCount}
							perPage={pagination.take}
							page={pagination.page}
							onPageChange={(page) => handlePageChange(page)}
							let:pages
						>
							<Pagination.Content>
								<Pagination.Item>
									<Pagination.PrevButton>
										<ChevronLeftIcon class="h-4 w-4" />
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
											<Pagination.Link {page} isActive={pagination.page === page.value}>
												{page.value}
											</Pagination.Link>
										</Pagination.Item>
									{/if}
								{/each}
								<Pagination.Item>
									<Pagination.NextButton>
										<span class="hidden sm:block">Next</span>
										<ChevronRightIcon class="h-4 w-4" />
									</Pagination.NextButton>
								</Pagination.Item>
							</Pagination.Content>
						</Pagination.Root>
					</div>
				{/if}
			{/if}
		</CardContent>
	</Card>
</div>

<!-- Upload Dialog -->
<Dialog.Root bind:open={uploadDialogOpen}>
	<Dialog.Content class="sm:max-w-[425px]">
		<Dialog.Header>
			<Dialog.Title>Upload Marketing Material</Dialog.Title>
			<Dialog.Description>
				Upload a file that users can download. Maximum file size is 50 MB.
			</Dialog.Description>
		</Dialog.Header>
		<div class="grid gap-4 py-4">
			<div class="grid gap-2">
				<Label for="file-upload">File</Label>
				<Input
					id="file-upload"
					type="file"
					onchange={handleFileSelect}
					disabled={isUploading}
					accept="*/*"
				/>
				{#if uploadFile}
					<p class="text-sm text-muted-foreground">
						Selected: {uploadFile.name} ({formatFileSize(uploadFile.size)})
					</p>
				{/if}
			</div>
			<div class="grid gap-2">
				<Label for="material-name">Display Name</Label>
				<Input
					id="material-name"
					bind:value={uploadName}
					placeholder="Enter display name for the material"
					disabled={isUploading}
				/>
			</div>
		</div>
		<Dialog.Footer>
			<Button
				variant="outline"
				onclick={() => (uploadDialogOpen = false)}
				disabled={isUploading}
			>
				Cancel
			</Button>
			<Button onclick={handleUpload} disabled={isUploading || !uploadFile || !uploadName.trim()}>
				{isUploading ? 'Uploading...' : 'Upload'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Edit Name Dialog -->
<Dialog.Root bind:open={editDialogOpen}>
	<Dialog.Content class="sm:max-w-[425px]">
		<Dialog.Header>
			<Dialog.Title>Edit Material Name</Dialog.Title>
			<Dialog.Description>
				Change the display name for this marketing material.
			</Dialog.Description>
		</Dialog.Header>
		<div class="grid gap-4 py-4">
			<div class="grid gap-2">
				<Label for="edit-name">Display Name</Label>
				<Input
					id="edit-name"
					bind:value={editName}
					placeholder="Enter new display name"
					disabled={isEditing}
				/>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (editDialogOpen = false)} disabled={isEditing}>
				Cancel
			</Button>
			<Button onclick={handleSaveName} disabled={isEditing || !editName.trim()}>
				{isEditing ? 'Saving...' : 'Save'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

