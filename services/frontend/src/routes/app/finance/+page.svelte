<script lang="ts">
	import BalanceCard from '$lib/components/BalanceCard.svelte';
	import TransactionTable from '$lib/components/TransactionTable.svelte';
	import { desimBalanceStore } from '$lib/stores/desim_balance.svelte';
	import { tokenBalanceStore } from '$stores/token_balance.svelte';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-foreground">Finance Dashboard</h1>
		<p class="mt-2 text-muted-foreground">Manage your funds</p>
	</div>

	<div class="grid gap-6 grid-cols-1 md:grid-cols-2 mb-6">
		<BalanceCard store={desimBalanceStore} title="Balance" showButtons={true} product="desim" />
		<BalanceCard store={tokenBalanceStore} title="Token Balance" showButtons={true} product="desim" />
	</div>

	<Tabs.Root value="transactions">
		<Tabs.List class="w-full">
			<Tabs.Trigger value="transactions">Transactions</Tabs.Trigger>
			<Tabs.Trigger value="token">Token Transactions</Tabs.Trigger>
		</Tabs.List>
		
		<Tabs.Content value="transactions" class="mt-4">
			<TransactionTable store={desimBalanceStore} />
		</Tabs.Content>

		<Tabs.Content value="token" class="mt-4">
			<TransactionTable store={tokenBalanceStore} />
		</Tabs.Content>
	</Tabs.Root>
</div>
