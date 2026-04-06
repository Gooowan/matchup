<script lang="ts">
	import BottomNav from '$lib/components/matchup/BottomNav.svelte';
</script>

<div class="flex h-[100dvh] flex-col items-center justify-center gap-4" style="background: #dae1eb;">
	<i class="fi fi-rr-shopping-bag" style="font-size: 48px; color: #696969;"></i>
	<p class="text-[18px] font-bold" style="color: #171717;">Marketplace</p>
	<p class="text-[14px] font-medium" style="color: #7d7d7d;">Coming soon</p>
</div>

<BottomNav active="marketplace" />
