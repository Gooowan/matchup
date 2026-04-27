<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authFetch } from '$lib/utils/authFetch';
	import AsyncState from '$lib/components/matchup/AsyncState.svelte';

	interface Plan {
		id: string;
		name: string;
		description: { String: string; Valid: boolean } | string;
		duration_days: number;
		price_cents: number;
		is_active: boolean;
	}

	interface ActiveSub {
		id: string;
		status: string;
		started_at: { Time: string };
		expired_at: { Time: string };
		subscription_name: string;
	}

	let plans = $state<Plan[]>([]);
	let activeSub = $state<ActiveSub | null>(null);
	let isLoading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const [plansResp, subResp] = await Promise.all([
				authFetch('/subscriptions/plans'),
				authFetch('/subscriptions/my/active')
			]);
			if (plansResp.ok) {
				const body = await plansResp.json();
				plans = body.data ?? [];
			}
			if (subResp.ok) {
				const body = await subResp.json();
				activeSub = body.data ?? null;
			}
		} catch {
			error = 'Failed to load subscription info';
		} finally {
			isLoading = false;
		}
	});

	function planDescription(p: Plan): string {
		if (typeof p.description === 'string') return p.description;
		return p.description?.Valid ? p.description.String : '';
	}

	function formatPrice(cents: number): string {
		if (cents === 0) return 'Free';
		return `$${(cents / 100).toFixed(2)}`;
	}

	function formatDate(isoish: string | undefined): string {
		if (!isoish) return '';
		const d = new Date(isoish);
		return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	let expiryStr = $derived(activeSub ? formatDate(activeSub.expired_at?.Time) : '');
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden mu-screen">
	<div class="pt-safe"></div>

	<!-- Header -->
	<div class="flex items-center gap-3 px-4 pt-4 pb-2">
		<button onclick={() => goto('/settings')} aria-label="Back">
			<i class="fi fi-rr-angle-left mu-text-primary" style="font-size: 20px; line-height: 1;"></i>
		</button>
		<h1 class="mu-text-primary flex-1 text-[20px] font-black">Subscription</h1>
	</div>

	<AsyncState loading={isLoading} {error}>
		<div class="flex flex-1 flex-col overflow-y-auto px-4 pb-[100px]" style="gap: 16px; padding-top: 12px;">

			<!-- Current plan banner -->
			<div
				class="rounded-[20px] p-4"
				style="background: linear-gradient(135deg, #8984da 0%, #a89de8 100%);"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-[12px] font-semibold uppercase tracking-wider text-white/70">Current plan</p>
						<p class="mt-1 text-[22px] font-black text-white">
							{activeSub ? activeSub.subscription_name : 'Free'}
						</p>
						{#if activeSub}
							<p class="mt-0.5 text-[13px] font-medium text-white/80">Renews {expiryStr}</p>
						{:else}
							<p class="mt-0.5 text-[13px] font-medium text-white/80">Upgrade to unlock all features</p>
						{/if}
					</div>
					<i class="fi fi-rr-diamond text-white" style="font-size: 36px; line-height: 1; opacity: 0.9;"></i>
				</div>
			</div>

			<!-- Available plans -->
			{#if plans.length > 0}
				<p class="text-[11px] font-semibold uppercase tracking-wider px-1" style="color: #aeb4bc;">Available plans</p>
				{#each plans as plan}
					<div
						class="mu-card rounded-[20px] p-4"
						style="border: {activeSub?.subscription_name === plan.name ? '2px solid #8984da' : '1.5px solid transparent'};"
					>
						<div class="flex items-start justify-between gap-3">
							<div class="flex-1">
								<p class="mu-text-primary text-[16px] font-black">{plan.name}</p>
								{#if planDescription(plan)}
									<p class="mt-0.5 text-[13px] font-medium" style="color: #696969;">{planDescription(plan)}</p>
								{/if}
								<p class="mt-1 text-[12px] font-medium" style="color: #aeb4bc;">
									{plan.duration_days} days
								</p>
							</div>
							<div class="flex-shrink-0 text-right">
								<p class="text-[20px] font-black" style="color: #8984da;">{formatPrice(plan.price_cents)}</p>
								{#if plan.price_cents > 0}
									<p class="text-[11px] font-medium" style="color: #aeb4bc;">/ period</p>
								{/if}
							</div>
						</div>

						{#if activeSub?.subscription_name === plan.name}
							<div class="mt-3 flex items-center gap-1.5">
								<div class="h-[8px] w-[8px] rounded-full" style="background: #22c55e;"></div>
								<span class="text-[12px] font-semibold" style="color: #22c55e;">Active</span>
							</div>
						{/if}
					</div>
				{/each}
			{/if}

			<!-- Purchase via app notice -->
			<div class="rounded-[20px] bg-white p-4" style="border: 1px dashed #d1d5db;">
				<div class="flex items-start gap-3">
					<i class="fi fi-brands-apple" style="font-size: 22px; color: #171717; flex-shrink: 0; line-height: 1; margin-top: 2px;"></i>
					<div>
						<p class="text-[14px] font-bold mu-text-primary">Purchase on the App</p>
						<p class="mt-1 text-[13px] font-medium leading-relaxed" style="color: #696969;">
							Subscriptions are purchased through the iOS app via Apple In-App Purchases. Open MatchUp on your iPhone to upgrade.
						</p>
					</div>
				</div>
			</div>

			<!-- What's included in Premium -->
			<div class="mu-card rounded-[20px] p-4" style="display: flex; flex-direction: column; gap: 10px;">
				<p class="text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">Premium features</p>
				{#each [
					{ icon: 'fi-rr-heart', text: 'Unlimited swipes per day' },
					{ icon: 'fi-rr-eye', text: 'See who liked you' },
					{ icon: 'fi-rr-settings-sliders', text: 'Advanced filters' },
					{ icon: 'fi-rr-badge', text: 'Premium badge on profile' }
				] as feature}
					<div class="flex items-center gap-3">
						<i class="fi {feature.icon}" style="font-size: 16px; color: #8984da; flex-shrink: 0;"></i>
						<span class="mu-text-primary text-[14px] font-medium">{feature.text}</span>
					</div>
				{/each}
			</div>
		</div>
	</AsyncState>
</div>
