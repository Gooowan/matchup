<!-- Example usage of BalanceCard component -->
<script lang="ts">
	import BalanceCard from './BalanceCard.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import { onMount } from 'svelte';

	onMount(() => {
		// Fetch balance data when component mounts
		desimBalanceStore.fetchBalance();
	});
</script>

<div class="mx-auto max-w-md space-y-4">
	<!-- Basic usage -->
	<BalanceCard store={desimBalanceStore} title="Grow Balance" />

	<!-- With details hidden -->
	<BalanceCard store={desimBalanceStore} title="Quick Balance" showDetails={false} />

	<!-- With action buttons -->
	<BalanceCard
		store={desimBalanceStore}
		title="Grow Balance with Actions"
		showButtons={true}
		product="desim"
	/>
</div>
