<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';

	let chatId = $derived(page.params.id);

	interface Message {
		id: string;
		text: string;
		sent: boolean;
		timestamp: string;
	}

	interface MessageGroup {
		date: string;
		messages: Message[];
	}

	let groups = $state<MessageGroup[]>([
		{
			date: 'Today',
			messages: [
				{ id: '1', text: 'Hey! When are you free to practice?', sent: false, timestamp: '18:20' },
				{ id: '2', text: "I'm free Saturday afternoon!", sent: true, timestamp: '18:22' },
				{ id: '3', text: "Perfect, let's meet at the studio at 3pm", sent: false, timestamp: '18:24' },
				{ id: '4', text: 'Sounds great, see you there 🎵', sent: true, timestamp: '18:25' }
			]
		}
	]);

	let inputText = $state('');
	let messagesEl: HTMLElement;
	let lastMessageId = $state<string | null>(null);
	let pollInterval: ReturnType<typeof setInterval>;

	function scrollToBottom() {
		if (messagesEl) {
			messagesEl.scrollTop = messagesEl.scrollHeight;
		}
	}

	function mapMessages(raw: any[]): MessageGroup[] {
		const today = new Date().toDateString();
		const yesterday = new Date(Date.now() - 86400000).toDateString();
		const grouped = new Map<string, Message[]>();

		for (const m of raw) {
			const d = new Date(m.created_at);
			const label =
				d.toDateString() === today
					? 'Today'
					: d.toDateString() === yesterday
						? 'Yesterday'
						: d.toLocaleDateString('en', { day: 'numeric', month: 'short', year: 'numeric' });
			if (!grouped.has(label)) grouped.set(label, []);
			grouped.get(label)!.push({
				id: m.id,
				text: m.content,
				sent: m.is_own ?? false,
				timestamp: d.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })
			});
		}

		return Array.from(grouped.entries()).map(([date, messages]) => ({ date, messages }));
	}

	async function fetchMessages() {
		try {
			const resp = await authFetch(`/chats/${chatId}/messages?limit=50`);
			if (resp.ok) {
				const response = await resp.json();
				if (response.data && Array.isArray(response.data)) {
					groups = mapMessages(response.data);
					const last = response.data[response.data.length - 1];
					lastMessageId = last?.id ?? null;
				}
			}
		} catch {
			// Keep existing messages on error
		}
	}

	async function fetchNewMessages() {
		if (!lastMessageId) return fetchMessages();
		try {
			const resp = await authFetch(
				`/chats/${chatId}/messages?limit=20&after=${lastMessageId}`
			);
			if (resp.ok) {
				const response = await resp.json();
				if (response.data && Array.isArray(response.data) && response.data.length > 0) {
					const newMsgs = response.data;
					const last = newMsgs[newMsgs.length - 1];
					lastMessageId = last.id;

					// Append to last group or create new group
					const today = 'Today';
					const lastGroup = groups[groups.length - 1];
					if (lastGroup?.date === today) {
						lastGroup.messages = [
							...lastGroup.messages,
							...newMsgs.map((m: any) => ({
								id: m.id,
								text: m.content,
								sent: m.is_own ?? false,
								timestamp: new Date(m.created_at).toLocaleTimeString('en', {
									hour: '2-digit',
									minute: '2-digit'
								})
							}))
						];
						groups = [...groups];
					} else {
						groups = mapMessages([
							...groups.flatMap((g) =>
								g.messages.map((m) => ({ id: m.id, content: m.text, is_own: m.sent }))
							),
							...newMsgs
						]);
					}
					setTimeout(scrollToBottom, 50);
				}
			}
		} catch {}
	}

	async function sendMessage() {
		const text = inputText.trim();
		if (!text) return;

		// Optimistic update
		const tempId = `temp-${Date.now()}`;
		const lastGroup = groups[groups.length - 1];
		if (lastGroup) {
			lastGroup.messages.push({
				id: tempId,
				text,
				sent: true,
				timestamp: new Date().toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })
			});
			groups = [...groups];
		}
		inputText = '';
		setTimeout(scrollToBottom, 50);

		try {
			await authFetch(`/chats/${chatId}/messages`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ type: 'TEXT', content: text })
			});
		} catch {}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
	}

	onMount(async () => {
		await fetchMessages();
		setTimeout(scrollToBottom, 100);
		pollInterval = setInterval(fetchNewMessages, 3000);
	});

	onDestroy(() => {
		clearInterval(pollInterval);
	});
</script>

<div class="flex h-[100dvh] flex-col overflow-hidden" style="background: #151517;">
	<!-- Top gradient -->
	<div
		class="pointer-events-none absolute top-0 right-0 left-0 z-10 h-[122px]"
		style="background: linear-gradient(180deg, rgba(0,0,0,0.8) 0%, rgba(0,0,0,0) 100%);"
	></div>

	<!-- Header -->
	<div
		class="relative z-20 flex items-center gap-4 px-4"
		style="margin-top: max(env(safe-area-inset-top), 8px); height: 54px;"
	>
		<!-- Back button -->
		<button
			onclick={() => goto('/chats')}
			class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
			aria-label="Back"
		>
			<i class="fi fi-rr-angle-left" style="font-size: 20px; line-height: 1; color: white;"></i>
		</button>

		<!-- Avatar + name -->
		<div class="flex flex-1 items-center gap-3">
			<div
				class="relative h-[38px] w-[38px] flex-shrink-0 overflow-hidden rounded-full"
				style="border: 1px solid #313131;"
			>
				<img
					src="https://images.unsplash.com/photo-1518611012118-696072aa579a?w=80"
					alt="Match"
					class="h-full w-full object-cover"
				/>
				<div
					class="absolute right-0 top-0 h-2.5 w-2.5 rounded-full border-2"
					style="background: #22c55e; border-color: #151517;"
				></div>
			</div>
			<span class="text-[18px] font-semibold text-white">Match</span>
		</div>

		<!-- Options button -->
		<button
			class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
			aria-label="Options"
		>
			<i
				class="fi fi-rr-menu-dots-vertical rotate-90"
				style="font-size: 20px; line-height: 1; color: white;"
			></i>
		</button>
	</div>

	<!-- Message thread -->
	<div
		bind:this={messagesEl}
		class="flex flex-1 flex-col overflow-y-auto px-4 pt-4"
		style="gap: 32px; padding-bottom: 120px;"
	>
		{#each groups as group}
			<div class="text-center text-[12px] font-normal" style="color: #e1e1e1;">{group.date}</div>

			{#each group.messages as msg}
				<div class="flex" class:justify-end={msg.sent} class:justify-start={!msg.sent}>
					<div
						class="max-w-[75%]"
						style="
							background: {msg.sent ? '#2c2b30' : '#706f77'};
							border-radius: {msg.text.length < 30 ? '50px' : '14px'};
							padding: 8px 16px;
						"
					>
						<p class="text-[16px] font-normal leading-snug" style="color: {msg.sent ? '#e1e1e1' : 'white'};">
							{msg.text}
						</p>
						<div class="mt-1 flex items-center gap-1" class:justify-end={msg.sent}>
							<span class="text-[12px] font-medium" style="color: {msg.sent ? '#7c7c7c' : '#b4b0b0'};">
								{msg.timestamp}
							</span>
							{#if msg.sent}
								<i class="fi fi-rr-check" style="font-size: 11px; color: #7c7c7c;"></i>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		{/each}
	</div>

	<!-- Input bar -->
	<div
		class="absolute right-0 bottom-0 left-0 flex items-center gap-3 px-4"
		style="padding-bottom: max(env(safe-area-inset-bottom), 34px); padding-top: 8px; background: #151517;"
	>
		<button
			class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
			aria-label="Attach"
		>
			<i class="fi fi-rr-clip" style="font-size: 20px; line-height: 1; color: white;"></i>
		</button>
		<div class="glass-pill flex flex-1 items-center px-4" style="height: 38px;">
			<input
				type="text"
				placeholder="Message"
				bind:value={inputText}
				onkeydown={handleKeydown}
				class="w-full bg-transparent text-[14px] font-semibold outline-none"
				style="color: white; caret-color: white;"
			/>
		</div>
		{#if inputText.trim()}
			<button
				class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
				onclick={sendMessage}
				aria-label="Send"
			>
				<i class="fi fi-rr-paper-plane" style="font-size: 20px; line-height: 1; color: white;"></i>
			</button>
		{/if}
	</div>
</div>
