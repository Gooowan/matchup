import { apiGet, apiPost } from './client';

export interface FeedCandidate {
	user_id: string;
	dance_styles: string[];
	gender: string;
	birth_date?: string;
	height_cm?: number;
	goal: string;
	program: string;
	categories: string[];
	country?: string;
	city?: string;
	metadata: Record<string, unknown>;
	profile_data: Record<string, unknown>;
	distance_km: number;
	source?: string;
}

export interface FeedPage {
	candidates: FeedCandidate[];
}

export type SwipeAction = 'LIKE' | 'PASS';

export interface SwipeResult {
	match: boolean;
	match_id?: string;
}

/** Get the next page of feed candidates. */
export function getFeed(limit = 20): Promise<FeedPage> {
	return apiGet<FeedPage>('/matchup/feed', { limit });
}

/** Get trainer profiles for the trainer tab. */
export function getTrainers(limit = 20, offset = 0): Promise<FeedCandidate[]> {
	return apiGet<FeedCandidate[]>('/matchup/trainers', { limit, offset });
}

/** Record a swipe action (LIKE or PASS). */
export function swipe(targetUserId: string, action: SwipeAction, source?: string): Promise<SwipeResult> {
	return apiPost<SwipeResult>('/matchup/swipe', {
		target_user_id: targetUserId,
		action,
		source
	});
}
