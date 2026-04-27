<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';

	interface Report {
		id: string;
		reporter_id: string;
		reporter_email: string;
		reported_id: string;
		reported_email: string;
		category: string;
		comment: string;
		created_at: string;
	}

	let reports = $state<Report[]>([]);
	let isLoading = $state(true);
	let banning = $state<string | null>(null);

	onMount(async () => {
		try {
			const resp = await authFetch('/admin/reports');
			if (resp.ok) {
				const body = await resp.json();
				reports = body.data ?? [];
			} else {
				toast.error('Не вдалося завантажити скарги');
			}
		} catch {
			toast.error('Не вдалося завантажити скарги');
		} finally {
			isLoading = false;
		}
	});

	async function banUser(userId: string, email: string) {
		if (banning) return;
		banning = userId;
		try {
			const resp = await authFetch(`/admin/users/${userId}/ban`, { method: 'POST' });
			if (resp.ok) {
				toast.success(`Заблоковано ${email}`);
				reports = reports.filter((r) => r.reported_id !== userId);
			} else {
				toast.error('Не вдалося заблокувати користувача');
			}
		} catch {
			toast.error('Не вдалося заблокувати користувача');
		} finally {
			banning = null;
		}
	}

	async function dismissReport(reportId: string) {
		reports = reports.filter((r) => r.id !== reportId);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString('uk-UA', {
			month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Модерація | Адмін</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto">
	<h1 class="text-2xl font-bold mb-6">Черга модерації</h1>

	{#if isLoading}
		<div class="flex items-center justify-center py-20">
			<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
		</div>
	{:else if reports.length === 0}
		<div class="flex flex-col items-center justify-center py-20 gap-3 text-center">
			<p class="text-lg font-semibold text-gray-500">Скарг немає — все чисто!</p>
		</div>
	{:else}
		<div class="flex flex-col gap-4">
			{#each reports as report (report.id)}
				<div class="rounded-xl border bg-white p-4 shadow-sm" style="border-color: #e0e0e0;">
					<div class="flex items-start justify-between gap-4">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-2">
								<span class="rounded-full px-2 py-0.5 text-xs font-semibold text-white" style="background: #e74c3c;">
									{report.category}
								</span>
								<span class="text-xs text-gray-400">{formatDate(report.created_at)}</span>
							</div>

							<div class="flex flex-col gap-1 text-sm">
								<div>
									<span class="font-medium text-gray-500">Від:</span>
									<span class="ml-1 text-gray-700">{report.reporter_email || report.reporter_id}</span>
								</div>
								<div>
									<span class="font-medium text-gray-500">На:</span>
									<span class="ml-1 font-semibold text-gray-900">{report.reported_email || report.reported_id}</span>
								</div>
								{#if report.comment}
									<div class="mt-1 rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-600 border" style="border-color: #e0e0e0;">
										{report.comment}
									</div>
								{/if}
							</div>
						</div>

						<div class="flex flex-col gap-2 flex-shrink-0">
							<button
								onclick={() => banUser(report.reported_id, report.reported_email)}
								disabled={!!banning}
								class="rounded-lg px-4 py-2 text-sm font-semibold text-white transition-opacity disabled:opacity-50"
								style="background: #e74c3c;"
							>
								{banning === report.reported_id ? 'Блокування…' : 'Заблокувати'}
							</button>
							<button
								onclick={() => dismissReport(report.id)}
								class="rounded-lg px-4 py-2 text-sm font-semibold text-gray-600 border transition-opacity"
								style="border-color: #d1d5db;"
							>
								Відхилити
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
