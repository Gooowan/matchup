<script lang="ts">
	import { Pencil, Check, X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';

	interface Props {
		value: string;
		label: string;
		onSave: (newValue: string) => Promise<{ success: boolean; error?: string }>;
		minLength?: number;
		maxLength?: number;
	}

	let { value, label, onSave, minLength = 2, maxLength = 50 }: Props = $props();

	let isEditing = $state(false);
	let editValue = $state(value);
	let isLoading = $state(false);
	let error = $state('');

	function startEditing() {
		editValue = value;
		isEditing = true;
		error = '';
	}

	function cancelEditing() {
		isEditing = false;
		editValue = value;
		error = '';
	}

	function validateValue(val: string): string | null {
		if (val.length < minLength) {
			return `Must be at least ${minLength} characters`;
		}
		if (val.length > maxLength) {
			return `Must be no more than ${maxLength} characters`;
		}
		return null;
	}

	async function saveValue() {
		const validationError = validateValue(editValue);
		if (validationError) {
			error = validationError;
			return;
		}

		isLoading = true;
		error = '';

		const result = await onSave(editValue);

		isLoading = false;

		if (result.success) {
			isEditing = false;
			error = '';
		} else {
			error = result.error || 'Failed to save';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			saveValue();
		} else if (e.key === 'Escape') {
			cancelEditing();
		}
	}
</script>

{#if isEditing}
	<div class="flex flex-col gap-2">
		<div class="flex items-center gap-2">
			<Input
				type="text"
				bind:value={editValue}
				onkeydown={handleKeydown}
				disabled={isLoading}
				class="flex-1 bg-transparent"
				placeholder={label}
			/>
			<Button
				size="icon"
				variant="ghost"
				onclick={saveValue}
				disabled={isLoading}
				class="h-8 w-8 shrink-0"
			>
				<Check class="h-4 w-4" />
			</Button>
			<Button
				size="icon"
				variant="ghost"
				onclick={cancelEditing}
				disabled={isLoading}
				class="h-8 w-8 shrink-0"
			>
				<X class="h-4 w-4" />
			</Button>
		</div>
		{#if error}
			<p class="text-sm text-red-500">{error}</p>
		{/if}
	</div>
{:else}
	<div class="flex h-8 items-center gap-2">
		<span class="font-bold">{value}</span>
		<Button
			size="icon"
			variant="ghost"
			onclick={startEditing}
			class="h-6 w-6 shrink-0 opacity-60 hover:opacity-100"
		>
			<Pencil class="h-3 w-3" />
		</Button>
	</div>
{/if}
