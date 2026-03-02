import { authFetch } from '$lib/utils/authFetch';

interface HotwalletBalance {
	bnb: number;
	usdt: number;
	usdc: number;
}

function createAdminWithdrawsStore() {
	let transactions = $state<Transaction[]>([]);
	let hotwalletBalance = $state<HotwalletBalance | null>(null);
	let isTransactionsLoading = $state(false);
	let isBalanceLoading = $state(false);
	let error = $state<string | null>(null);

	return {
		get transactions() {
			return transactions;
		},
		get hotwalletBalance() {
			return hotwalletBalance;
		},
		get isTransactionsLoading() {
			return isTransactionsLoading;
		},
		get isBalanceLoading() {
			return isBalanceLoading;
		},
		get error() {
			return error;
		},
		get currency() {
			return 'USD';
		},

		async fetchTransactions(page: number = 1, take: number = 10) {
			isTransactionsLoading = true;
			error = null;

			const resp = await authFetch(`/admin/desim/withdraw/pending?page=${page}&take=${take}`);
			if (!resp.ok) {
				error = 'Failed to fetch pending withdrawals';
				isTransactionsLoading = false;
				return null;
			}

			const response: ApiPaginatedResponse<Transaction> = await resp.json();
			if (resp.status !== 200) {
				error = response.error || 'Failed to fetch pending withdrawals';
				isTransactionsLoading = false;
				return null;
			}

			transactions = response.data || [];
			isTransactionsLoading = false;
			return {
				transactions: response.data || [],
				meta: response.meta,
			};
		},

		async fetchHotwalletBalance() {
			isBalanceLoading = true;
			error = null;

			const resp = await authFetch('/admin/desim/hotwallet/balance');
			if (!resp.ok) {
				error = 'Failed to fetch hotwallet balance';
				isBalanceLoading = false;
				return null;
			}

			const response: ApiResponse<HotwalletBalance> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				error = response.error || 'Failed to fetch hotwallet balance';
				isBalanceLoading = false;
				return null;
			}

			hotwalletBalance = response.data;
			isBalanceLoading = false;
			return response.data;
		},

		async approveWithdrawal(transactionId: string) {
			const resp = await authFetch(`/admin/desim/withdraw/${transactionId}/approve`, {
				method: 'POST',
			});

			if (!resp.ok) {
				const response: ApiResponse<unknown> = await resp.json();
				throw new Error(response.error || 'Failed to approve withdrawal');
			}

			const response: ApiResponse<Transaction> = await resp.json();
			return response.data;
		},

		async rejectWithdrawal(transactionId: string) {
			const resp = await authFetch(`/admin/desim/withdraw/${transactionId}/reject`, {
				method: 'POST',
			});

			if (!resp.ok) {
				const response: ApiResponse<unknown> = await resp.json();
				throw new Error(response.error || 'Failed to reject withdrawal');
			}

			const response: ApiResponse<Transaction> = await resp.json();
			return response.data;
		},

		hasSufficientBalance(amount: number, token: string): boolean {
			if (!hotwalletBalance) return false;

			const absAmount = Math.abs(amount);
			const tokenLower = token.toLowerCase();

			switch (tokenLower) {
				case 'bnb':
					return hotwalletBalance.bnb >= absAmount;
				case 'usdt':
					return hotwalletBalance.usdt >= absAmount;
				case 'usdc':
					return hotwalletBalance.usdc >= absAmount;
				default:
					return false;
			}
		},
	};
}

export const adminWithdrawsStore = createAdminWithdrawsStore();

