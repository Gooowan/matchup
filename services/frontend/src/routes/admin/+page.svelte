<script lang="ts">
	import { onMount } from 'svelte';
	import { authFetch } from '$utils/authFetch';

	let totalUsers = $state(0);
	let isLoading = $state(true);

	onMount(async () => {
		try {
			const resp = await authFetch('/admin/stats');
			if (resp.ok) {
				const body = await resp.json();
				totalUsers = body.data?.totalUsers ?? 0;
			}
		} catch {
			// ignore
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="container mx-auto px-4 py-8" style="max-width: 800px;">
	<h1 class="mb-2 text-[28px] font-black" style="color: #171717;">Адмін-панель</h1>
	<p class="mb-8 text-[14px] font-medium" style="color: #696969;">Огляд платформи</p>

	<!-- Stats -->
	<div class="mb-8 grid grid-cols-2 gap-4">
		<div class="rounded-[20px] bg-white p-5 shadow-sm">
			<p class="mb-1 text-[11px] font-semibold uppercase tracking-wider" style="color: #aeb4bc;">КОРИСТУВАЧІ</p>
			{#if isLoading}
				<div class="h-8 w-16 animate-pulse rounded-lg" style="background: #e0e0e0;"></div>
			{:else}
				<p class="text-[32px] font-black" style="color: #171717;">{totalUsers.toLocaleString()}</p>
			{/if}
			<p class="mt-1 text-[12px] font-medium" style="color: #aeb4bc;">Зареєстровано</p>
		</div>
	</div>

	<!-- Navigation -->
	<div class="flex flex-col gap-3">
		<a href="/admin/users" class="flex items-center justify-between rounded-[20px] bg-white p-4 shadow-sm">
			<div class="flex items-center gap-3">
				<i class="fi fi-rr-users" style="font-size: 20px; color: #8984da;"></i>
				<div>
					<p class="text-[15px] font-semibold" style="color: #171717;">Управління користувачами</p>
					<p class="text-[12px] font-medium" style="color: #696969;">Пошук, перегляд, редагування</p>
				</div>
			</div>
			<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
		</a>

		<a href="/admin/moderation" class="flex items-center justify-between rounded-[20px] bg-white p-4 shadow-sm">
			<div class="flex items-center gap-3">
				<i class="fi fi-rr-shield-check" style="font-size: 20px; color: #8984da;"></i>
				<div>
					<p class="text-[15px] font-semibold" style="color: #171717;">Модерація</p>
					<p class="text-[12px] font-medium" style="color: #696969;">Скарги, блокування</p>
				</div>
			</div>
			<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
		</a>

		<a href="/admin/marketing" class="flex items-center justify-between rounded-[20px] bg-white p-4 shadow-sm">
			<div class="flex items-center gap-3">
				<i class="fi fi-rr-megaphone" style="font-size: 20px; color: #8984da;"></i>
				<div>
					<p class="text-[15px] font-semibold" style="color: #171717;">Маркетингові матеріали</p>
					<p class="text-[12px] font-medium" style="color: #696969;">Завантаження файлів</p>
				</div>
			</div>
			<i class="fi fi-rr-angle-right" style="font-size: 14px; color: #aeb4bc;"></i>
		</a>
	</div>
</div>
