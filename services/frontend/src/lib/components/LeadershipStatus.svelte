<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';
	import CheckIcon from '@lucide/svelte/icons/check';
	import LockIcon from '@lucide/svelte/icons/lock';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';

	interface StatusWithUserInfo {
		id: number;
		level: number;
		turnover_required: number;
		personal_turnover_required: number;
		reward: number;
		metadata: Record<string, unknown>;
		status: 'claimed' | 'can_claim' | 'locked';
		claimed_at?: string;
	}

	interface StatusProgressInfo {
		status_id: number;
		team_turnover_progress: number;
		personal_turnover_progress: number;
	}

	interface UserStatusResponse {
		statuses: StatusWithUserInfo[];
		next_claimable: StatusProgressInfo | null;
	}

	let statuses = $state<StatusWithUserInfo[]>([]);
	let currentStatus = $derived.by(() => {
		const revStats = statuses.toReversed();
		return revStats.find((status) => status.status === 'claimed');
	});
	let nextClaimable = $state<StatusProgressInfo | null>(null);
	let isLoading = $state(true);
	let claimingStatusId = $state<number | null>(null);

	async function fetchStatuses() {
		isLoading = true;
		try {
			const resp = await authFetch('/desim/statuses');
			if (!resp.ok) {
				toast.error('Failed to load statuses');
				return;
			}

			const response: ApiResponse<UserStatusResponse> = await resp.json();
			if (response.data) {
				statuses = response.data.statuses;
				nextClaimable = response.data.next_claimable;
			}
		} catch (error) {
			console.error('Error fetching statuses:', error);
			toast.error('An error occurred while loading statuses');
		} finally {
			isLoading = false;
		}
	}

	async function claimStatus(statusId: number) {
		claimingStatusId = statusId;
		try {
			const resp = await authFetch(`/desim/statuses/${statusId}/claim`, {
				method: 'POST',
			});

			if (!resp.ok) {
				const response: ApiResponse<unknown> = await resp.json();
				toast.error(response.error || 'Failed to claim status');
				return;
			}

			const response: ApiResponse<Transaction> = await resp.json();
			if (response.data) {
				toast.success(`Successfully claimed status! Reward: $${response.data.amount}`);
				// Refresh statuses after successful claim
				await fetchStatuses();
			}
		} catch (error) {
			console.error('Error claiming status:', error);
			toast.error('An error occurred while claiming status');
		} finally {
			claimingStatusId = null;
		}
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function formatNumber(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			minimumFractionDigits: 0,
			maximumFractionDigits: 0,
		}).format(amount);
	}

	function getButtonText(status: StatusWithUserInfo): string {
		if (status.status === 'claimed') {
			return 'Claimed';
		} else if (status.status === 'can_claim') {
			return 'Claim Now';
		} else {
			return 'Locked';
		}
	}

	function getButtonVariant(status: StatusWithUserInfo): 'default' | 'secondary' | 'outline' {
		if (status.status === 'claimed') {
			return 'secondary';
		} else if (status.status === 'can_claim') {
			return 'default';
		} else {
			return 'outline';
		}
	}

	function getProgressForStatus(statusId: number): StatusProgressInfo | null {
		if (nextClaimable && nextClaimable.status_id === statusId) {
			return nextClaimable;
		}
		return null;
	}

	onMount(() => {
		fetchStatuses();
	});
</script>

<div class="mt-12">
	<h2 class="text-foreground mb-6 text-2xl font-bold">Leadership Status</h2>
	<div class="bg-primary mb-6 rounded-lg p-4">
		<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
			<div class="flex flex-col">
				<p class="text-3xl font-bold">
					Status {currentStatus ? currentStatus.level : 0}
				</p>
				<p class=" text-xs">Your current status</p>
			</div>
			{#if nextClaimable}
				{@const nextStatus =
					statuses.find((status) => status.id === nextClaimable?.status_id) || statuses[0]}

				{#if nextStatus.personal_turnover_required > 0}
					<div class="space-y-1">
						<p class="text-xs">Personal Turnover For Next Status</p>
						<div class="h-4 overflow-hidden rounded-full bg-white p-px">
							<div
								class="bg-primary h-full rounded-full transition-all"
								style="width: {Math.min(nextClaimable.personal_turnover_progress * 100, 100)}%"
							></div>
						</div>
						<p class="text-center text-xs">
							{#if nextStatus}
								{formatCurrency(
									nextStatus?.personal_turnover_required * nextClaimable.personal_turnover_progress
								)}
								/ {formatCurrency(nextStatus?.personal_turnover_required)} ({Math.round(
									nextClaimable.personal_turnover_progress * 100
								)}%)
							{/if}
						</p>
					</div>
				{/if}

				{#if nextStatus.turnover_required > 0}
					<div class="space-y-1">
						<p class="text-xs">Team Turnover For Next Status</p>
						<div class="h-4 overflow-hidden rounded-full bg-white p-px">
							<div
								class="bg-primary h-full rounded-full transition-all"
								style="width: {Math.min(nextClaimable.team_turnover_progress * 100, 100)}%"
							></div>
						</div>
						<p class="text-center text-xs">
							{#if nextStatus}
								{formatCurrency(
									nextStatus?.turnover_required * nextClaimable.team_turnover_progress
								)}
								/ {formatCurrency(nextStatus?.turnover_required)} ({Math.round(
									nextClaimable.team_turnover_progress * 100
								)}%)
							{/if}
						</p>
					</div>
				{/if}
			{/if}
		</div>
	</div>

	{#if isLoading}
		<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
			{#each Array(3) as _}
				<Card>
					<CardHeader>
						<div class="bg-muted h-6 animate-pulse rounded"></div>
					</CardHeader>
					<CardContent>
						<div class="bg-muted mb-4 h-20 animate-pulse rounded"></div>
						<div class="bg-muted h-4 animate-pulse rounded"></div>
					</CardContent>
					<CardFooter>
						<div class="bg-muted h-9 w-full animate-pulse rounded"></div>
					</CardFooter>
				</Card>
			{/each}
		</div>
	{:else if statuses.length === 0}
		<Card>
			<CardContent class="py-8">
				<p class="text-muted-foreground text-center">No statuses available</p>
			</CardContent>
		</Card>
	{:else}
		<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
			{#each statuses as status (status.id)}
				{@const progress = getProgressForStatus(status.id)}
				<Card class="flex flex-col">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							Status {status.level}
							{#if status.status === 'locked'}
								<LockIcon class="text-muted-foreground size-4" />
							{/if}
						</CardTitle>
					</CardHeader>

					<CardContent class="flex-1 space-y-4">
						{#if status.reward > 0}
							<div class="py-2 text-center">
								<p class="text-3xl font-bold">{formatCurrency(status.reward)}</p>
								<p class="text-muted-foreground mt-1 text-sm">Reward</p>
							</div>
							<Separator />
						{/if}

						<div class="grid grid-cols-2 gap-4">
							<div class="text-center">
								<p class="text-xl font-semibold">
									{formatCurrency(status.personal_turnover_required)}
								</p>
								<p class="text-muted-foreground mb-1 text-sm">Personal Turnover</p>
								{#if progress}
									<div class="mt-2">
										<div class="bg-muted h-2 overflow-hidden rounded-full">
											<div
												class="bg-primary h-full transition-all"
												style="width: {Math.min(progress.personal_turnover_progress * 100, 100)}%"
											></div>
										</div>
										<p class="text-muted-foreground mt-1 text-xs">
											{Math.round(progress.personal_turnover_progress * 100)}%
										</p>
									</div>
								{/if}
							</div>

							<div class="text-center">
								<p class="text-xl font-semibold">
									{formatCurrency(status.turnover_required)}
								</p>
								<p class="text-muted-foreground mb-1 text-sm">Team Turnover</p>
								{#if progress}
									<div class="mt-2">
										<div class="bg-muted h-2 overflow-hidden rounded-full">
											<div
												class="bg-primary h-full transition-all"
												style="width: {Math.min(progress.team_turnover_progress * 100, 100)}%"
											></div>
										</div>
										<p class="text-muted-foreground mt-1 text-xs">
											{Math.round(progress.team_turnover_progress * 100)}%
										</p>
									</div>
								{/if}
							</div>
						</div>

						{#if status.claimed_at}
							<p class="text-muted-foreground pt-2 text-center text-xs">
								Claimed on {new Date(status.claimed_at).toLocaleDateString()}
							</p>
						{/if}
					</CardContent>

					<CardFooter>
						<Button
							class="w-full"
							variant={getButtonVariant(status)}
							disabled={status.status !== 'can_claim' || claimingStatusId === status.id}
							onclick={() => status.status === 'can_claim' && claimStatus(status.id)}
						>
							{#if claimingStatusId === status.id}
								Claiming...
							{:else}
								{getButtonText(status)}
							{/if}
						</Button>
					</CardFooter>
				</Card>
			{/each}
		</div>
	{/if}
</div>
