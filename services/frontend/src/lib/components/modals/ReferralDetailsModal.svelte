<script lang="ts">
	import { authFetch } from '$utils/authFetch';
	import * as ResponsiveDialog from '$lib/components/ui/responsive-dialog';
	import ReferralStatsInfo from '$lib/components/ReferralStatsInfo.svelte';

	interface Props {
		isOpen: boolean;
		userId: string | null;
		displayName?: string;
		isMissedProfit?: boolean;
	}

	let { isOpen = $bindable(), userId, displayName, isMissedProfit = false }: Props = $props();

	let stats = $state<UserReferralStats | null>(null);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// Fetch stats when userId changes and dialog is open
	$effect(() => {
		if (isOpen && userId) {
			fetchStats(userId);
		} else if (!isOpen) {
			// Reset when dialog closes
			stats = null;
			error = null;
		}
	});

	async function fetchStats(id: string) {
		isLoading = true;
		error = null;

		try {
			const resp = await authFetch(`/desim/stats/${id}`);
			if (resp.ok) {
				const response: ApiResponse<UserReferralStats> = await resp.json();
				if (response.data) {
					stats = response.data;
				} else {
					error = response.error || 'Failed to fetch stats';
				}
			} else {
				error = 'Failed to fetch stats';
			}
		} catch (err) {
			console.error('Failed to fetch user stats:', err);
			error = 'Failed to fetch stats';
		} finally {
			isLoading = false;
		}
	}
</script>

<ResponsiveDialog.Root bind:open={isOpen}>
	<ResponsiveDialog.Content class="max-w-4xl max-h-[90vh] overflow-y-auto">
		<ResponsiveDialog.Header>
			<ResponsiveDialog.Title>Referral Details</ResponsiveDialog.Title>
		</ResponsiveDialog.Header>

		<div class="mt-4">
			{#if isMissedProfit}
				<div class="mb-4 rounded-lg border border-yellow-500/50 bg-yellow-500/10 p-4">
					<p class="text-sm text-yellow-700 dark:text-yellow-400">
						You have missed this referral transaction because you have yet to start farming.
					</p>
				</div>
			{/if}
			{#if isLoading}
				<div class="flex items-center justify-center py-8">
					<div class="text-muted-foreground">Loading...</div>
				</div>
			{:else if error}
				<div class="flex items-center justify-center py-8">
					<div class="text-destructive">{error}</div>
				</div>
			{:else}
				<ReferralStatsInfo stats={stats} isRoot={false} {displayName} compact={true} />
			{/if}
		</div>
	</ResponsiveDialog.Content>
</ResponsiveDialog.Root>

