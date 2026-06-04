export type AccountType = 'dancer' | 'parent' | 'trainer' | 'club';

export const ACCOUNT_TYPES: readonly AccountType[] = ['dancer', 'parent', 'trainer', 'club'] as const;

/**
 * Account types that have restricted access: only Settings, Marketplace, Chats.
 * Map and Feed are hidden, and routing is guarded in (app)/+layout.svelte.
 */
export function isRestrictedAccountType(t: string | null | undefined): boolean {
	return t === 'trainer' || t === 'club';
}
