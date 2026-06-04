import { apiGet, apiPost } from './client';

export interface Club {
	id: string;
	name: string;
	slug: string;
	city?: string;
	country?: string;
	latitude?: number;
	longitude?: number;
	description?: string;
	avatar?: string;
	is_active: boolean;
}

export interface ClubMember {
	user_id: string;
	role: string;
	profile_data?: Record<string, unknown>;
	gender?: string;
	goal?: string;
	program?: string;
	categories?: string[];
}

/** Fetch clubs near a coordinate. */
export function getNearbyCLubs(lat: number, lng: number, radiusKm = 50): Promise<Club[]> {
	return apiGet<Club[]>('/clubs', { lat, lng, radius_km: radiusKm });
}

/** List members of a club with optional filter params. */
export function getClubMembers(
	slug: string,
	filters?: {
		gender?: string;
		goal?: string;
		program?: string;
		city?: string;
		age_min?: number;
		age_max?: number;
	}
): Promise<ClubMember[]> {
	return apiGet<ClubMember[]>(`/clubs/${slug}/members`, filters);
}

/** Register a new club (public endpoint). */
export function createClub(data: {
	name: string;
	city: string;
	country: string;
	latitude?: number;
	longitude?: number;
	description?: string;
	website?: string;
	phone?: string;
	working_hours?: Record<string, unknown>;
	photos?: string[];
}): Promise<Club> {
	return apiPost<Club>('/clubs/register', data);
}

/** Get trainers for a club. */
export function getClubTrainers(slug: string): Promise<TrainerCard[]> {
	return apiGet<TrainerCard[]>(`/clubs/${slug}/trainers`);
}

export interface TrainerCard {
	trainer_user_id: string;
	joined_at: string;
	gender: string;
	categories: string[];
	metadata: Record<string, unknown>;
	profile_data: Record<string, unknown>;
}
