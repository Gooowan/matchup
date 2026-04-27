<script lang="ts">
	import AdminUsersTable from './AdminUsersTable.svelte';
</script>

<div class="container mx-auto px-4 py-8 space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-3xl font-bold ">Управління користувачами</h1>
	</div>

	<AdminUsersTable title="Users" />
</div>
