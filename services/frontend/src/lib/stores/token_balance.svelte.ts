import { authFetch } from '$utils/authFetch';

function createTokenBalanceStore() {
	let balance = $state<Balance | null>(null);
	let transactions = $state<Transaction[]>([]);
	let isBalanceLoading = $state<boolean>(false);
	let isTransactionsLoading = $state<boolean>(false);
	let error = $state<string | null>(null);

	return {
		get currency() {
			return "$DESIM";
		},
		get balance() {
			return balance;
		},
		set balance(value) {
			balance = value;
		},
		get transactions() {
			return transactions;
		},
		set transactions(value) {
			transactions = value;
		},
		get isBalanceLoading() {
			return isBalanceLoading;
		},
		set isBalanceLoading(value) {
			isBalanceLoading = value;
		},
		get isTransactionsLoading() {
			return isTransactionsLoading;
		},
		set isTransactionsLoading(value) {
			isTransactionsLoading = value;
		},
		get error() {
			return error;
		},
		set error(value) {
			error = value;
		},
		async fetchBalance() {
			this.isBalanceLoading = true;
			this.error = null;

			const resp = await authFetch('/desim/token/balance');
			if (!resp.ok) {
				this.error = 'Failed to fetch balance';
				this.isBalanceLoading = false;
				return null;
			}

			const response: ApiResponse<Balance> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to fetch balance';
				this.isBalanceLoading = false;
				return null;
			}

			this.balance = response.data;
			this.isBalanceLoading = false;
			return response.data;
		},
		async fetchTransactions(page: number = 1, take: number = 10) {
			this.isTransactionsLoading = true;
			this.error = null;

			const resp = await authFetch(`/desim/token/transactions?page=${page}&take=${take}`);
			if (!resp.ok) {
				this.error = 'Failed to fetch transactions';
				this.isTransactionsLoading = false;
				return null;
			}

			const response: ApiPaginatedResponse<Transaction> = await resp.json();
			if (resp.status !== 200) {
				this.error = response.error || 'Failed to fetch transactions';
				this.isTransactionsLoading = false;
				return null;
			}

			this.transactions = response.data || [];
			this.isTransactionsLoading = false;
			return {
				transactions: response.data || [],
				meta: response.meta,
			};
		},
		async refreshAll() {
			await Promise.all([
				this.fetchBalance(),
				this.fetchTransactions(),
			]);
		},
		clear() {
			this.balance = null;
			this.transactions = [];
			this.error = null;
		},
	};
}

export const tokenBalanceStore = createTokenBalanceStore();
