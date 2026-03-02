<script lang="ts">
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import * as Avatar from '$lib/components/ui/avatar/index.js';

	interface Props {
		referral: DirectReferralStat;
		onClick: (userId: string, displayName: string) => void;
	}

	let { referral, onClick }: Props = $props();

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 2,
		}).format(amount);
	}

	function getDisplayName(): string {
		const firstName = referral.profile_data?.first_name || '';
		const lastName = referral.profile_data?.last_name || '';
		const name = `${firstName} ${lastName}`.trim();
		return name || referral.email;
	}

	function handleClick() {
		onClick(referral.user_id, getDisplayName());
	}
</script>

<Card class="cursor-pointer transition-shadow hover:shadow-md" onclick={handleClick}>
	<CardContent class="px-4">
		<div class="flex items-center justify-between">
			<div class="flex flex-col lg:flex-row justify-between gap-2 w-full">
				<div class="flex gap-2 items-center">
					<Avatar.Root class="h-9 w-9 rounded-lg">
						{#if referral.profile_data?.avatar}
							<Avatar.Image
								src={referral.profile_data.avatar}
								alt={getDisplayName()}
							/>
						{/if}
						<Avatar.Fallback>
							{getDisplayName().charAt(0).toUpperCase()}
						</Avatar.Fallback>
					</Avatar.Root>

					<div class="flex flex-col gap-0">
						<Badge variant="default" class="text-[10px] py-px px-1">Level {referral.current_rank_level}</Badge>
						<div class="flex items-center gap-2">
							<p class="font-medium text-nowrap">{getDisplayName()}</p>
						</div>
						<div class="text-xs text-muted-foreground">
							{referral.email}
						</div>
					</div>
				</div>
				<div class="grid grid-cols-2 md:flex flex-wrap md:justify-center gap-6 w-full lg:justify-end px-4 lg:px-8 lg:gap-12">
					<div class="flex flex-col items-center justify-center col-span-2">
						<p class="text-xl font-bold">{formatCurrency(referral.personal_turnover)}</p>
						<p class="text-sm font-light text-muted-foreground">Investment</p>
					</div>
					<div class="flex flex-col items-center justify-center ">
						<p class="text-xl font-bold">{referral.direct_count}</p>
						<p class="text-sm font-light text-muted-foreground">Direct Team</p>
					</div>
					<div class="flex flex-col items-center justify-center ">
						<p class="text-xl font-bold">{formatCurrency(referral.direct_turnover)}</p>
						<p class="text-sm font-light text-muted-foreground">Direct Sales</p>
					</div>
					<div class="flex flex-col items-center justify-center ">
						<p class="text-xl font-bold">{referral.total_team_count}</p>
						<p class="text-sm font-light text-muted-foreground">Team</p>
					</div>
					<div class="flex flex-col items-center justify-center ">
						<p class="text-xl font-bold">{formatCurrency(referral.total_team_turnover)}</p>
						<p class="text-sm font-light text-muted-foreground">Team Sales</p>
					</div>
				</div>
			</div>
			<ChevronRightIcon class="h-5 w-5 text-muted-foreground" />
		</div>
	</CardContent>
</Card>

