<script lang="ts">
	import { Card, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import CopyReferralLink from '$lib/components/CopyReferralLink.svelte';
	import * as Avatar from '$lib/components/ui/avatar/index.js';

	interface Props {
		stats: UserReferralStats | null;
		isRoot?: boolean;
		displayName?: string;
		compact?: boolean;
	}

	let { stats, isRoot = false, displayName, compact = false }: Props = $props();

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	const getDisplayName = $derived.by(() => {
		if (displayName) {
			return displayName;
		}
		if (!stats) return '';
		const firstName = stats.profile_data?.first_name || '';
		const lastName = stats.profile_data?.last_name || '';
		const name = `${firstName} ${lastName}`.trim();
		return name || stats.email;
	});

	// Placeholder values for loading state - use 0 and empty strings
	const placeholderStats: UserReferralStats = {
		email: 'loading@example.com',
		referral_id: 0,
		current_rank_level: 0,
		personal_turnover: 0,
		direct_count: 0,
		direct_turnover: 0,
		total_team_count: 0,
		total_team_turnover: 0,
		profile_data: {},
		current_rank_metadata: {},
		user_id: '',
		inviter_id: null,
	};

	// Use $derived to make it reactive to stats changes
	const displayStats = $derived(stats || placeholderStats);
</script>

<Card
	class={`mb-4 flex flex-col gap-4 px-4 ${compact ? '' : 'md:flex-row md:items-center md:justify-between'}`}
>
	<div class="flex flex-col gap-2">
		<div class="flex items-center gap-2">
			<Avatar.Root class="h-14 w-14 rounded-lg">
					<Avatar.Image src={displayStats.profile_data.avatar} alt={getDisplayName} />
				<Avatar.Fallback>
					{getDisplayName.charAt(0).toUpperCase()}
				</Avatar.Fallback>
			</Avatar.Root>
			<div class="flex flex-col gap-2">
				<CardTitle class="flex items-center gap-2">
					<p>
						{#if isRoot}
							Your Referral Stats
						{:else}
							{getDisplayName}'s Stats
						{/if}
					</p>
					<Badge variant="default">Level {displayStats.current_rank_level}</Badge>
					<Badge variant="outline">#{displayStats.referral_id}</Badge>
				</CardTitle>
				<span class="text-muted-foreground text-sm">{displayStats.email}</span>
			</div>
		</div>
	</div>
	<div class={compact ? '' : 'md:w-1/3'}>
		<div class="flex w-min flex-col gap-2">
			<span class="text-muted-foreground text-sm">Copy Referral Link</span>
			<CopyReferralLink referralId={displayStats.referral_id} />
		</div>
	</div>
</Card>

<div class={`grid grid-cols-2 gap-4 ${compact ? '' : 'lg:grid-cols-5'}`}>
	<div
		class={`col-span-2 flex flex-col items-center justify-center rounded-lg bg-[#194BFB] p-4 ${compact ? '' : 'lg:col-span-1'}`}
	>
		<p class="text-2xl font-bold">{formatCurrency(displayStats.personal_turnover)}</p>
		<p class="text-sm font-light text-white">Personal Turnover</p>
	</div>
	<div class="flex flex-col items-center justify-center rounded-lg bg-[#194BFB] p-4">
		<p class="text-2xl font-bold">{displayStats.direct_count}</p>
		<p class="text-sm font-light text-white">Direct Referrals</p>
	</div>
	<div class="flex flex-col items-center justify-center rounded-lg bg-[#194BFB] p-4">
		<p class="text-2xl font-bold">{displayStats.total_team_count}</p>
		<p class="text-sm font-light text-white">Total Team</p>
	</div>
	<div class="flex flex-col items-center justify-center rounded-lg bg-[#194BFB] p-4">
		<p class="text-2xl font-bold">{formatCurrency(displayStats.direct_turnover)}</p>
		<p class="text-sm font-light text-white">Direct Turnover</p>
	</div>
	<div class="flex flex-col items-center justify-center rounded-lg bg-[#194BFB] p-4">
		<p class="text-2xl font-bold">{formatCurrency(displayStats.total_team_turnover)}</p>
		<p class="text-sm font-light text-white">Total Team Turnover</p>
	</div>
</div>
