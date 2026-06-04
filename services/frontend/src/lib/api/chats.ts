import { apiGet, apiPost } from './client';

export interface ChatPeer {
	id: string;
	profile_data: {
		first_name?: string;
		avatar?: string;
		[key: string]: unknown;
	};
	/** Present (== "club") when this chat is with a club rather than a user. */
	kind?: string;
	club_name?: string;
	club_logo?: string;
	club_phone?: string;
	club_slug?: string;
}

export interface ChatClub {
	id: string;
	name: string;
	logo?: string;
	slug: string;
	phone?: string;
}

export interface ChatDTO {
	id: string;
	other_user_id: string;
	other_user?: ChatPeer;
	/** Set for club chats — brand the inbox row as the club. */
	club?: ChatClub;
	unread_count: number;
	/** True for club chats (drives the Business tab). */
	is_club_owner: boolean;
	last_message?: {
		content: string;
		created_at: string;
	};
	last_activity?: string;
}

export interface MessageDTO {
	id: string;
	chat_id: string;
	sender_id: string;
	content: string;
	is_own: boolean;
	created_at: string;
}

export interface MessagesPage {
	messages: MessageDTO[];
}

/** Fetch chat inbox. */
export function listChats(): Promise<ChatDTO[]> {
	return apiGet<ChatDTO[]>('/chats');
}

/** Get or create a DM with another user (requires mutual match for dancers; open for trainers). */
export function getOrCreateChat(otherUserId: string): Promise<{ chat_id: string }> {
	return apiPost<{ chat_id: string }>('/chats', { user_id: otherUserId });
}

/** Fetch messages in a chat thread. */
export function getMessages(chatId: string, opts?: { after?: number; cursor?: string }): Promise<MessagesPage> {
	return apiGet<MessagesPage>(`/chats/${chatId}`, opts);
}

/** Send a message. */
export function sendMessage(chatId: string, content: string): Promise<MessageDTO> {
	return apiPost<MessageDTO>(`/chats/${chatId}`, { content });
}

/** Mark chat as read. */
export function markRead(chatId: string): Promise<void> {
	return apiPost<void>(`/chats/${chatId}/read`);
}
