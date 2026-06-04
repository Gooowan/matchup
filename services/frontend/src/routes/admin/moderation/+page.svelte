<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';

	// --- User reports (existing) ---
	interface UserReport {
		id: string;
		reporter_id: string;
		reporter_email: string;
		reported_id: string;
		reported_email: string;
		category: string;
		comment: string;
		created_at: string;
	}

	// --- Message reports (new) ---
	interface MessageReport {
		id: string;
		message_id: string;
		chat_id: string;
		reporter_id: string;
		reported_user_id: string;
		category: string;
		comment: string;
		content_snapshot: string;
		current_content: string;
		status: string;
		created_at: number;
	}

	type ActiveTab = 'users' | 'messages';
	let activeTab = $state<ActiveTab>('users');

	let userReports = $state<UserReport[]>([]);
	let messageReports = $state<MessageReport[]>([]);

	let isLoadingUsers = $state(true);
	let isLoadingMessages = $state(false);
	let messagesLoaded = $state(false);

	let banning = $state<string | null>(null);
	let resolving = $state<string | null>(null);
	let hiding = $state<string | null>(null);

	onMount(async () => {
		await loadUserReports();
	});

	async function loadUserReports() {
		isLoadingUsers = true;
		try {
			const resp = await authFetch('/admin/reports');
			if (resp.ok) {
				const body = await resp.json();
				userReports = body.data ?? [];
			} else {
				toast.error('Не вдалося завантажити скарги');
			}
		} catch {
			toast.error('Не вдалося завантажити скарги');
		} finally {
			isLoadingUsers = false;
		}
	}

	async function loadMessageReports() {
		if (messagesLoaded) return;
		isLoadingMessages = true;
		try {
			const resp = await authFetch('/admin/chats/message-reports?status=open&limit=100');
			if (resp.ok) {
				const body = await resp.json();
				messageReports = body.data ?? [];
				messagesLoaded = true;
			} else {
				toast.error('Не вдалося завантажити скарги на повідомлення');
			}
		} catch {
			toast.error('Не вдалося завантажити скарги на повідомлення');
		} finally {
			isLoadingMessages = false;
		}
	}

	function switchTab(tab: ActiveTab) {
		activeTab = tab;
		if (tab === 'messages' && !messagesLoaded) loadMessageReports();
	}

	// --- User report actions ---
	async function banUser(userId: string, email: string) {
		if (banning) return;
		banning = userId;
		try {
			const resp = await authFetch(`/admin/users/${userId}/ban`, { method: 'POST' });
			if (resp.ok) {
				toast.success(`Заблоковано ${email}`);
				userReports = userReports.filter((r) => r.reported_id !== userId);
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
		userReports = userReports.filter((r) => r.id !== reportId);
	}

	// --- Message report actions ---
	async function hideAndResolve(report: MessageReport) {
		if (hiding) return;
		hiding = report.id;
		try {
			// Hide the message first, then resolve the report.
			await authFetch(`/admin/chats/messages/${report.message_id}`, { method: 'DELETE' });
			await authFetch(`/admin/chats/message-reports/${report.id}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: 'resolved' })
			});
			toast.success('Повідомлення приховано, скаргу вирішено');
			messageReports = messageReports.filter((r) => r.id !== report.id);
		} catch {
			toast.error('Не вдалося обробити скаргу');
		} finally {
			hiding = null;
		}
	}

	async function resolveReport(reportId: string, status: 'resolved' | 'dismissed') {
		if (resolving) return;
		resolving = reportId;
		try {
			const resp = await authFetch(`/admin/chats/message-reports/${reportId}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status })
			});
			if (resp.ok) {
				messageReports = messageReports.filter((r) => r.id !== reportId);
				toast.success(status === 'resolved' ? 'Скаргу вирішено' : 'Скаргу відхилено');
			} else {
				toast.error('Не вдалося обробити скаргу');
			}
		} catch {
			toast.error('Не вдалося обробити скаргу');
		} finally {
			resolving = null;
		}
	}

	function formatDate(val: string | number): string {
		const d = typeof val === 'number' ? new Date(val) : new Date(val);
		return d.toLocaleString('uk-UA', {
			month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Модерація | Адмін</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto">
	<h1 class="text-2xl font-bold mb-6">Черга модерації</h1>

	<!-- Tab switcher -->
	<div class="flex gap-2 mb-6">
		<button
			onclick={() => switchTab('users')}
			class="rounded-lg px-4 py-2 text-sm font-semibold transition-colors"
			style="background: {activeTab === 'users' ? '#8984da' : '#f3f4f6'}; color: {activeTab === 'users' ? 'white' : '#374151'};"
		>
			Скарги на користувачів
			{#if userReports.length > 0}
				<span class="ml-1.5 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 text-white text-[11px] font-bold px-1">
					{userReports.length}
				</span>
			{/if}
		</button>
		<button
			onclick={() => switchTab('messages')}
			class="rounded-lg px-4 py-2 text-sm font-semibold transition-colors"
			style="background: {activeTab === 'messages' ? '#8984da' : '#f3f4f6'}; color: {activeTab === 'messages' ? 'white' : '#374151'};"
		>
			Скарги на повідомлення
			{#if messageReports.length > 0}
				<span class="ml-1.5 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 text-white text-[11px] font-bold px-1">
					{messageReports.length}
				</span>
			{/if}
		</button>
	</div>

	<!-- User Reports Tab -->
	{#if activeTab === 'users'}
		{#if isLoadingUsers}
			<div class="flex items-center justify-center py-20">
				<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
			</div>
		{:else if userReports.length === 0}
			<div class="flex flex-col items-center justify-center py-20 gap-3 text-center">
				<p class="text-lg font-semibold text-gray-500">Скарг немає — все чисто!</p>
			</div>
		{:else}
			<div class="flex flex-col gap-4">
				{#each userReports as report (report.id)}
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

	<!-- Message Reports Tab -->
	{:else if activeTab === 'messages'}
		{#if isLoadingMessages}
			<div class="flex items-center justify-center py-20">
				<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #e0e0e0; border-top-color: #8984da;"></div>
			</div>
		{:else if messageReports.length === 0}
			<div class="flex flex-col items-center justify-center py-20 gap-3 text-center">
				<p class="text-lg font-semibold text-gray-500">Скарг на повідомлення немає!</p>
			</div>
		{:else}
			<div class="flex flex-col gap-4">
				{#each messageReports as report (report.id)}
					<div class="rounded-xl border bg-white p-4 shadow-sm" style="border-color: #e0e0e0;">
						<div class="flex items-start justify-between gap-4">
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2 mb-2">
									<span class="rounded-full px-2 py-0.5 text-xs font-semibold text-white" style="background: #e74c3c;">
										{report.category}
									</span>
									<span class="text-xs text-gray-400">{formatDate(report.created_at)}</span>
									<!-- Deep link to the chat thread -->
									<a
										href="/chats/{report.chat_id}"
										target="_blank"
										rel="noreferrer"
										class="text-xs underline"
										style="color: #8984da;"
									>Відкрити чат ↗</a>
								</div>

								<!-- Message snapshot (what was reported) -->
								<div class="rounded-lg px-3 py-2 mb-2" style="background: #fff3cd; border: 1px solid #ffc107;">
									<p class="text-[11px] font-semibold text-gray-500 mb-0.5">Повідомлення на момент скарги:</p>
									<p class="text-sm text-gray-800">{report.content_snapshot}</p>
								</div>

								<!-- Current content (may differ if edited) -->
								{#if report.current_content !== report.content_snapshot}
									<div class="rounded-lg px-3 py-2 mb-2" style="background: #f3f4f6; border: 1px solid #e0e0e0;">
										<p class="text-[11px] font-semibold text-gray-500 mb-0.5">Поточний текст:</p>
										<p class="text-sm text-gray-600">{report.current_content}</p>
									</div>
								{/if}

								{#if report.comment}
									<p class="text-xs text-gray-500 mt-1">Коментар репортера: <em>{report.comment}</em></p>
								{/if}

								<div class="flex gap-2 mt-2 text-xs text-gray-400">
									<span>Репортер: {report.reporter_id.slice(0, 8)}…</span>
									<span>Порушник: {report.reported_user_id.slice(0, 8)}…</span>
								</div>
							</div>

							<div class="flex flex-col gap-2 flex-shrink-0">
								<button
									onclick={() => hideAndResolve(report)}
									disabled={!!hiding || !!resolving}
									class="rounded-lg px-3 py-2 text-xs font-semibold text-white transition-opacity disabled:opacity-50"
									style="background: #e74c3c;"
								>
									{hiding === report.id ? 'Приховую…' : 'Приховати + вирішити'}
								</button>
								<button
									onclick={() => resolveReport(report.id, 'resolved')}
									disabled={!!hiding || !!resolving}
									class="rounded-lg px-3 py-2 text-xs font-semibold text-white transition-opacity disabled:opacity-50"
									style="background: #22c55e;"
								>
									{resolving === report.id ? 'Обробка…' : 'Вирішити'}
								</button>
								<button
									onclick={() => resolveReport(report.id, 'dismissed')}
									disabled={!!hiding || !!resolving}
									class="rounded-lg px-3 py-2 text-xs font-semibold text-gray-600 border transition-opacity disabled:opacity-50"
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
	{/if}
</div>
