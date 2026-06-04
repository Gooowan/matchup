<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { authFetch } from '$lib/utils/authFetch';
	import toast from 'svelte-french-toast';
	import { t } from '$lib/locale';

	let chatId = $derived(page.params.id);

	interface Message {
		id: string;
		senderId: string;
		text: string;
		sent: boolean;
		timestamp: string;
		createdAtMs: number;
	}

	interface MessageGroup {
		date: string;
		messages: Message[];
	}

	let groups = $state<MessageGroup[]>([]);
	let isLoading = $state(true);
	let inputText = $state('');
	let messagesEl: HTMLElement;
	let newestMsgMs = $state<number | null>(null);
	let pollInterval: ReturnType<typeof setInterval>;
	let isSending = $state(false);

	let peerName = $state('');
	let peerAvatar = $state('');
	let peerUserId = $state('');
	// Club-chat branding + call affordance.
	let isClub = $state(false);
	let clubPhone = $state('');
	let showCaller = $state(false);

	// Per-message action menu (long-press → Report / Block).
	interface MessageAction {
		messageId: string;
		senderId: string;
		content: string;
	}
	let messageAction = $state<MessageAction | null>(null);

	function scrollToBottom() {
		if (messagesEl) {
			messagesEl.scrollTop = messagesEl.scrollHeight;
		}
	}

	function dateLabel(d: Date): string {
		const today = new Date().toDateString();
		const yesterday = new Date(Date.now() - 86400000).toDateString();
		if (d.toDateString() === today) return $t('chats.today');
		if (d.toDateString() === yesterday) return $t('chats.yesterday');
		return d.toLocaleDateString('uk', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function rawToMsg(m: any): Message {
		const d = new Date(m.created_at);
		return {
			id: m.id,
			senderId: m.sender_id ?? '',
			text: m.content,
			sent: m.is_own ?? false,
			timestamp: d.toLocaleTimeString('uk', { hour: '2-digit', minute: '2-digit' }),
			createdAtMs: m.created_at
		};
	}

	async function handleReportMessage(messageId: string, content: string) {
		messageAction = null;
		try {
			const resp = await authFetch(`/chats/${chatId}/messages/${messageId}/report`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ category: 'inappropriate' })
			});
			if (resp.ok) {
				toast.success($t('chats.report_sent'));
			} else {
				toast.error($t('chats.report_error'));
			}
		} catch {
			toast.error($t('chats.report_error'));
		}
	}

	async function handleBlockFromChat(userId: string) {
		messageAction = null;
		try {
			const resp = await authFetch(`/users/${userId}/block`, { method: 'POST' });
			if (resp.ok) {
				toast.success($t('chats.block_sent'));
			} else {
				toast.error($t('chats.report_error'));
			}
		} catch {
			toast.error($t('chats.report_error'));
		}
	}

	/** Collect server IDs already rendered so the poll can skip duplicates. */
	function existingIds(): Set<string> {
		const s = new Set<string>();
		for (const g of groups) for (const m of g.messages) s.add(m.id);
		return s;
	}

	function appendToGroups(newMsgs: Message[]) {
		const seen = existingIds();
		for (const msg of newMsgs) {
			// Skip any message whose real server ID is already shown.
			if (!msg.id.startsWith('temp-') && seen.has(msg.id)) continue;
			seen.add(msg.id);
			const label = dateLabel(new Date(msg.createdAtMs));
			const last = groups[groups.length - 1];
			if (last?.date === label) {
				last.messages.push(msg);
			} else {
				groups.push({ date: label, messages: [msg] });
			}
		}
		groups = [...groups];
	}

	async function fetchMessages() {
		try {
			const resp = await authFetch(`/chats/${chatId}/messages?limit=50`);
			if (!resp.ok) return;
			const response = await resp.json();
			if (response.data && Array.isArray(response.data)) {
				// API already returns ASC order after our backend fix.
				const msgs: Message[] = response.data.map(rawToMsg);
				// Rebuild groups from scratch.
				const map = new Map<string, Message[]>();
				for (const m of msgs) {
					const label = dateLabel(new Date(m.createdAtMs));
					if (!map.has(label)) map.set(label, []);
					map.get(label)!.push(m);
				}
				groups = Array.from(map.entries()).map(([date, messages]) => ({ date, messages }));
				if (msgs.length > 0) {
					newestMsgMs = msgs[msgs.length - 1].createdAtMs;
				}
			}
		} catch {
			// Keep existing messages on error
		}
	}

	async function fetchNewMessages() {
		if (newestMsgMs === null) {
			await fetchMessages();
			return;
		}
		try {
			const resp = await authFetch(`/chats/${chatId}/messages?limit=50&after=${newestMsgMs}`);
			if (!resp.ok) return;
			const response = await resp.json();
			if (response.data && Array.isArray(response.data) && response.data.length > 0) {
				const newMsgs: Message[] = response.data.map(rawToMsg);
				appendToGroups(newMsgs);
				newestMsgMs = newMsgs[newMsgs.length - 1].createdAtMs;
				setTimeout(scrollToBottom, 50);
			}
		} catch {}
	}

	async function sendMessage() {
		const text = inputText.trim();
		if (!text || isSending) return;

		// Optimistic UI: add a temp message immediately.
		const tempMs = Date.now();
		const tempMsg: Message = {
			id: `temp-${tempMs}`,
			senderId: '',
			text,
			sent: true,
			timestamp: new Date(tempMs).toLocaleTimeString('uk', { hour: '2-digit', minute: '2-digit' }),
			createdAtMs: tempMs
		};
		appendToGroups([tempMsg]);
		inputText = '';
		isSending = true;
		setTimeout(scrollToBottom, 50);

		try {
			const resp = await authFetch(`/chats/${chatId}/messages`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ type: 'TEXT', content: text })
			});
			if (!resp.ok) {
				const body = await resp.json().catch(() => ({}));
				const msg = body.error ?? $t('chats.send_error');
				toast.error(msg);
				// Remove the optimistic temp message.
				for (const g of groups) {
					g.messages = g.messages.filter((m) => m.id !== tempMsg.id);
				}
				groups = groups.filter((g) => g.messages.length > 0);
				inputText = text;
			} else {
				// Replace temp with real message from server response.
				const body = await resp.json();
				const real = body.data ? rawToMsg(body.data) : null;
				if (real) {
					for (const g of groups) {
						const idx = g.messages.findIndex((m) => m.id === tempMsg.id);
						if (idx !== -1) {
							g.messages[idx] = real;
							break;
						}
					}
					groups = [...groups];
					// Advance watermark so the next poll doesn't re-fetch this message.
					newestMsgMs = real.createdAtMs;
				} else {
					// Server acknowledged but returned no message body; advance watermark
					// past the temp timestamp so the poll doesn't see it as a gap.
					newestMsgMs = Math.max(newestMsgMs ?? 0, tempMs);
				}
			}
		} catch {
			toast.error($t('chats.send_error'));
			inputText = text;
		} finally {
			isSending = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
	}

	async function fetchPeer() {
		try {
			const resp = await authFetch(`/chats/${chatId}/meta`);
			if (!resp.ok) return;
			const body = await resp.json();
			const peer = body.data ?? {};
			if (peer.kind === 'club') {
				// Club chat → brand the thread as the club; enable call if phone set.
				isClub = true;
				peerName = peer.club_name || 'Club';
				peerAvatar = peer.club_logo ?? '';
				peerUserId = '';
				clubPhone = peer.club_phone ?? '';
				return;
			}
			const pd = (peer.profile_data ?? {}) as Record<string, string>;
			const fullName = [pd.first_name, pd.last_name].filter(Boolean).join(' ').trim();
			peerName = fullName || 'Match';
			peerAvatar = pd.avatar ?? '';
			peerUserId = peer.id ?? '';
		} catch {
			// keep defaults on error
		}
	}

	onMount(async () => {
		authFetch(`/chats/${chatId}/read`, { method: 'POST' }).catch(() => {});
		await Promise.all([fetchPeer(), fetchMessages()]);
		isLoading = false;
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
		<button
			class="flex flex-1 items-center gap-3 text-left"
			onclick={() => peerUserId && goto(`/profiles/${peerUserId}`)}
			disabled={!peerUserId}
		>
			<div
				class="relative h-[38px] w-[38px] flex-shrink-0 overflow-hidden rounded-full"
				style="border: 1px solid #313131;"
			>
				{#if peerAvatar}
					<img
						src={peerAvatar}
						alt={peerName}
						class="h-full w-full object-cover"
					/>
				{:else}
					<div class="flex h-full w-full items-center justify-center" style="background: #2c2b30;">
						<i class="fi fi-rr-user text-white" style="font-size: 18px;"></i>
					</div>
				{/if}
			</div>
			<span class="text-[18px] font-semibold text-white">{peerName || 'Match'}</span>
		</button>

		<!-- Call button — only for club chats that have a phone number. -->
		{#if isClub && clubPhone}
			<button
				class="glass-pill flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center"
				aria-label={$t('chats.call')}
				onclick={() => (showCaller = true)}
			>
				<i class="fi fi-rr-phone-call" style="font-size: 20px; line-height: 1; color: white;"></i>
			</button>
		{/if}
	</div>

	<!-- Message thread -->
	<div
		bind:this={messagesEl}
		class="flex flex-1 flex-col overflow-y-auto px-4 pt-4"
		style="gap: 32px; padding-bottom: 120px;"
	>
		{#if isLoading}
			<div class="flex flex-1 items-center justify-center py-16">
				<div class="h-8 w-8 animate-spin rounded-full border-4" style="border-color: #313131; border-top-color: #8984da;"></div>
			</div>
		{:else if groups.length === 0}
			<div class="flex flex-col items-center justify-center py-16 gap-3">
				<i class="fi fi-rr-comment" style="font-size: 40px; color: #313131;"></i>
				<p class="text-[14px] font-medium" style="color: #696969;">{$t('chats.empty_thread')}</p>
			</div>
		{/if}

		{#each groups as group}
			<div class="text-center text-[12px] font-normal" style="color: #e1e1e1;">{group.date}</div>

			{#each group.messages as msg}
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="flex"
					class:justify-end={msg.sent}
					class:justify-start={!msg.sent}
					oncontextmenu={(e) => { e.preventDefault(); if (!msg.sent) messageAction = { messageId: msg.id, senderId: msg.senderId, content: msg.text }; }}
				>
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
				placeholder={$t('chats.message_placeholder')}
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

	<!-- Per-message action sheet (long-press / right-click on received messages) -->
	{#if messageAction}
		<div
			class="fixed inset-0 z-50 flex flex-col justify-end"
			style="background: rgba(0,0,0,0.45);"
			role="dialog"
			aria-modal="true"
		>
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div class="absolute inset-0" onclick={() => (messageAction = null)}></div>
			<div class="relative rounded-t-[24px] pb-safe" style="background: #1e1e20;">
				<div class="flex justify-center pt-3 pb-2">
					<div class="h-1 w-10 rounded-full" style="background: #3a3a3a;"></div>
				</div>
				<div class="px-4 pb-2">
					<p class="truncate text-[13px]" style="color: #696969;">{messageAction.content}</p>
				</div>
				<div class="flex flex-col px-4 pb-6 gap-2">
					<button
						class="flex w-full items-center gap-3 rounded-[14px] p-4 text-left"
						style="background: #2c2b30;"
						onclick={() => handleReportMessage(messageAction!.messageId, messageAction!.content)}
					>
						<i class="fi fi-rr-flag" style="font-size: 18px; color: #e74c3c; line-height: 1;"></i>
						<span class="text-[14px] font-semibold text-white">{$t('chats.action_report_message')}</span>
					</button>
					{#if messageAction.senderId && peerUserId}
						<button
							class="flex w-full items-center gap-3 rounded-[14px] p-4 text-left"
							style="background: #2c2b30;"
							onclick={() => handleBlockFromChat(messageAction!.senderId)}
						>
							<i class="fi fi-rr-ban" style="font-size: 18px; color: #e74c3c; line-height: 1;"></i>
							<span class="text-[14px] font-semibold text-white">{$t('chats.action_block_user')}</span>
						</button>
					{/if}
					<button
						class="flex w-full items-center justify-center rounded-[14px] p-4"
						style="background: #2c2b30;"
						onclick={() => (messageAction = null)}
					>
						<span class="text-[14px] font-semibold" style="color: #696969;">{$t('common.cancel')}</span>
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Caller window: shows the club's number and opens the native dialer. -->
	{#if showCaller && isClub && clubPhone}
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 z-50 flex items-center justify-center"
			style="background: rgba(0,0,0,0.6); backdrop-filter: blur(6px);"
			onclick={() => (showCaller = false)}
		>
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div
				class="mx-8 flex w-full max-w-[300px] flex-col items-center rounded-[24px] px-6 py-8"
				style="background: #1f1e22;"
				onclick={(e) => e.stopPropagation()}
			>
				<div
					class="mb-4 flex h-[72px] w-[72px] items-center justify-center overflow-hidden rounded-full"
					style="border: 1px solid #313131; background: #2c2b30;"
				>
					{#if peerAvatar}
						<img src={peerAvatar} alt={peerName} class="h-full w-full object-cover" />
					{:else}
						<i class="fi fi-rr-bank text-white" style="font-size: 28px;"></i>
					{/if}
				</div>
				<p class="text-[18px] font-semibold text-white">{peerName}</p>
				<a
					href="tel:{clubPhone}"
					class="mt-1 text-[15px] font-medium"
					style="color: #8984da;"
				>{clubPhone}</a>

				<a
					href="tel:{clubPhone}"
					class="mt-6 flex h-[46px] w-full items-center justify-center gap-2 rounded-[50px] text-[15px] font-bold text-white"
					style="background: #22c55e;"
				>
					<i class="fi fi-rr-phone-call" style="font-size: 18px; line-height: 1;"></i>
					{$t('chats.call_action')}
				</a>
				<button
					class="mt-3 text-[14px] font-semibold"
					style="color: #898484;"
					onclick={() => (showCaller = false)}
				>{$t('chats.call_cancel')}</button>
			</div>
		</div>
	{/if}
</div>
