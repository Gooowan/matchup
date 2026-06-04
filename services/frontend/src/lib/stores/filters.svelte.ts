import { authFetch } from '$lib/utils/authFetch';
import type { FilterState } from '$lib/components/matchup/FilterSheet.svelte';

let filters = $state<FilterState>({});
let loaded = $state(false);

/** Count of active filter dimensions (for the badge). */
function countActiveFilters(f: FilterState): number {
	let n = 0;
	if (f.gender) n++;
	if (f.ageMin != null || f.ageMax != null) n++;
	if (f.heightMin != null || f.heightMax != null) n++;
	if (f.goal) n++;
	if (f.program) n++;
	if (f.categories?.length) n++;
	if (f.city) n++;
	if (f.wantsPartnerToFinance) n++;
	if (f.wantsPartnerToRelocate != null) n++;
	return n;
}

/** Shared load implementation so methods can call it without relying on `this`. */
async function loadPreferences() {
	if (loaded) return;
	try {
		const resp = await authFetch('/me/preferences');
		if (resp.ok) {
			const body = await resp.json();
			const p = body.data ?? body;
			filters = {
				gender: p.preferred_gender || undefined,
				ageMin: p.age_min ?? undefined,
				ageMax: p.age_max ?? undefined,
				heightMin: p.height_min ?? undefined,
				heightMax: p.height_max ?? undefined,
				goal: p.preferred_goal || undefined,
				program: p.preferred_program || undefined,
				categories: p.preferred_categories?.length ? p.preferred_categories : undefined,
				city: p.preferred_city || undefined,
				wantsPartnerToFinance: p.wants_partner_to_finance || undefined,
				wantsPartnerToRelocate: p.wants_partner_to_relocate ?? undefined
			};
		}
	} catch {
		// keep empty filters on error
	} finally {
		loaded = true;
	}
}

export const filterStore = {
	get filters() {
		return filters;
	},
	get loaded() {
		return loaded;
	},
	get activeCount() {
		return countActiveFilters(filters);
	},

	load: loadPreferences,

	apply(f: FilterState): Promise<void> {
		filters = f;
		const prefBody: Record<string, unknown> = {};
		if (f.gender) prefBody.preferred_gender = f.gender;
		if (f.ageMin != null) prefBody.age_min = f.ageMin;
		if (f.ageMax != null) prefBody.age_max = f.ageMax;
		if (f.heightMin != null) prefBody.height_min = f.heightMin;
		if (f.heightMax != null) prefBody.height_max = f.heightMax;
		if (f.goal) prefBody.preferred_goal = f.goal;
		if (f.program) prefBody.preferred_program = f.program;
		if (f.categories?.length) prefBody.preferred_categories = f.categories;
		if (f.city) prefBody.preferred_city = f.city;
		if (f.wantsPartnerToFinance) prefBody.wants_partner_to_finance = f.wantsPartnerToFinance;
		if (f.wantsPartnerToRelocate != null) prefBody.wants_partner_to_relocate = f.wantsPartnerToRelocate;
		return authFetch('/me/preferences', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(prefBody)
		}).then(() => {}).catch(() => {});
	},

	async clear(): Promise<void> {
		filters = {};
		loaded = false; // force re-read from server after reset
		try {
			await authFetch('/me/preferences', { method: 'DELETE' });
			// Re-load so local state reflects what the server seeded (city lock, etc.)
			await loadPreferences();
		} catch {
			// best-effort; local state is already cleared
		}
	}
};
