<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { Upload } from '@lucide/svelte';
	import toast from 'svelte-french-toast';

	interface Props {
		currentAvatar?: string;
		fallbackText: string;
		onUpload: (file: File) => Promise<{ success: boolean; avatar?: string; error?: string }>;
	}

	let { currentAvatar, fallbackText, onUpload }: Props = $props();

	let fileInput: HTMLInputElement;
	let isUploading = $state(false);
	let uploadProgress = $state(0);
	let error = $state('');

	function handleClick() {
		if (!isUploading) {
			fileInput.click();
		}
	}

	function validateFile(file: File): string | null {
		// Check file type
		const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];
		if (!validTypes.includes(file.type)) {
			return 'Invalid file type. Only JPG, PNG, and WEBP files are allowed.';
		}

		// Check file size (2MB = 2 * 1024 * 1024 bytes)
		const maxSize = 2 * 1024 * 1024;
		if (file.size > maxSize) {
			return 'File size must be less than 2MB.';
		}

		return null;
	}

	async function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];

		if (!file) return;

		// Validate file
		const validationError = validateFile(file);
		if (validationError) {
			error = validationError;
			toast.error(validationError);
			target.value = '';
			return;
		}

		error = '';
		isUploading = true;
		uploadProgress = 0;

		// Simulate progress (since FormData upload doesn't provide real progress easily)
		const progressInterval = setInterval(() => {
			uploadProgress += 10;
			if (uploadProgress >= 90) {
				clearInterval(progressInterval);
			}
		}, 100);

		try {
			const result = await onUpload(file);

			clearInterval(progressInterval);
			uploadProgress = 100;

			if (result.success) {
				toast.success('Avatar updated successfully!');
				setTimeout(() => {
					isUploading = false;
					uploadProgress = 0;
				}, 500);
			} else {
				error = result.error || 'Failed to upload avatar';
				toast.error(error);
				isUploading = false;
				uploadProgress = 0;
			}
		} catch (err) {
			clearInterval(progressInterval);
			error = 'Network error occurred';
			toast.error(error);
			isUploading = false;
			uploadProgress = 0;
		}

		target.value = '';
	}
</script>

<div class="relative inline-block">
	<button
		type="button"
		onclick={handleClick}
		disabled={isUploading}
		class="group relative cursor-pointer disabled:cursor-not-allowed"
	>
		<Avatar.Root class="h-28 w-28 rounded-lg">
			<Avatar.Image src={currentAvatar} alt="User avatar" />
			<Avatar.Fallback>{fallbackText}</Avatar.Fallback>
		</Avatar.Root>

		<!-- Overlay on hover -->
		<div
			class="absolute inset-0 flex items-center justify-center rounded-lg bg-black/50 opacity-0 transition-opacity group-hover:opacity-100 {isUploading
				? 'opacity-100'
				: ''}"
		>
			{#if isUploading}
				<div class="flex flex-col items-center gap-1">
					<div class="text-xs text-white">{uploadProgress}%</div>
				</div>
			{:else}
				<Upload class="h-6 w-6 text-white" />
			{/if}
		</div>
	</button>

	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/jpg,image/png,image/webp"
		onchange={handleFileChange}
		class="hidden"
	/>

	{#if error && !isUploading}
		<p class="mt-1 text-xs text-red-500">{error}</p>
	{/if}
</div>
