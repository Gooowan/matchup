import { apiGet } from './client';

export interface ProfilePreview {
	user_id: string;
	first_name?: string;
	last_name?: string;
	avatar?: string;
	gender?: string;
	birth_date?: string;
	height_cm?: number;
	goal?: string;
	program?: string;
	dance_styles?: string[];
	categories?: string[];
	city?: string;
	country?: string;
	bio?: string;
	media_urls?: string[];
	club_name?: string;
	primary_club_id?: string;
	profile_data?: Record<string, unknown>;
	metadata?: Record<string, unknown>;
}

/** Fetch a public profile preview by user ID. */
export function getProfilePreview(userId: string): Promise<ProfilePreview> {
	return apiGet<ProfilePreview>(`/profiles/${userId}/preview`);
}
