import { page } from '$app/state';

export const getReferralLink = (referralId: number) => {
	return `https://${page.url.hostname}/register?rel=${referralId}`;
};
