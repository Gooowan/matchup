let count = $state(0);

export const unreadStore = {
	get count() {
		return count;
	},
	/**
	 * Sync total unread count from the server-returned chat list.
	 * Called after /chats is fetched; replaces the manual increment pattern.
	 */
	syncFromChats(chats: Array<{ unread_count?: number }>) {
		count = chats.reduce((sum, c) => sum + (c.unread_count ?? 0), 0);
	},
	/** Increment locally on a confirmed match (before next /chats sync). */
	increment() {
		count++;
	},
	reset() {
		count = 0;
	}
};
