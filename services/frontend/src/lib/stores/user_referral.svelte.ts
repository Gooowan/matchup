import { authFetch } from '$utils/authFetch';

interface ReferralCount {
    direct_count: number;
    total_count: number;
}

function createUserReferralStore() {
    let referralCount = $state<ReferralCount | null>(null);
    let isLoading = $state<boolean>(false);
    let error = $state<string | null>(null);

    return {
        get referralCount() {
            return referralCount;
        },
        set referralCount(value) {
            referralCount = value;
        },
        get isLoading() {
            return isLoading;
        },
        set isLoading(value) {
            isLoading = value;
        },
        get error() {
            return error;
        },
        set error(value) {
            error = value;
        },
        async fetchReferralCount() {
            this.isLoading = true;
            this.error = null;

            try {
                const resp = await authFetch('/user/referrals/count');
                const response: ApiResponse<ReferralCount> = await resp.json();

                if (resp.status === 200 && response.data) {
                    this.referralCount = response.data;
                    return response.data;
                } else {
                    this.error = response.error || 'Failed to fetch referral count';
                    return null;
                }
            } catch (err) {
                console.error('Referral count fetch error:', err);
                this.error = 'Failed to fetch referral count';
                return null;
            } finally {
                this.isLoading = false;
            }
        },
        clear() {
            this.referralCount = null;
            this.error = null;
        }
    };
}

export const userReferralStore = createUserReferralStore();
