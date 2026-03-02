import { authFetch } from '$utils/authFetch';

export type TimeRange = 'all_time' | 'weekly' | 'last_24h';

export interface TokenSupplyHistoryPoint {
	date: string;
	price: number | null;
}

const REFRESH_INTERVAL = 60 * 1000; // 1 minute in milliseconds

function createTokenSupplyHistoryStore() {
	let history = $state<TokenSupplyHistoryPoint[]>([]);
	let isLoading = $state<boolean>(false);
	let error = $state<string | null>(null);
	let selectedRange = $state<TimeRange>('weekly');
	
	let currentPrice = $state<number | null>(null);
	let isLoadingPrice = $state<boolean>(false);
	let priceError = $state<string | null>(null);
	let refreshIntervalId: ReturnType<typeof setInterval> | null = null;
	
	let recent = $state<number[]>([]);
	let isLoadingRecent = $state<boolean>(false);
	let recentError = $state<string | null>(null);

	const store = {
		get history() {
			return history;
		},
		set history(value) {
			history = value;
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
		get selectedRange() {
			return selectedRange;
		},
		set selectedRange(value) {
			selectedRange = value;
		},
		get currentPrice() {
			return currentPrice;
		},
		set currentPrice(value) {
			currentPrice = value;
		},
		get isLoadingPrice() {
			return isLoadingPrice;
		},
		set isLoadingPrice(value) {
			isLoadingPrice = value;
		},
		get priceError() {
			return priceError;
		},
		set priceError(value) {
			priceError = value;
		},
		get recent() {
			return recent;
		},
		set recent(value) {
			recent = value;
		},
		get isLoadingRecent() {
			return isLoadingRecent;
		},
		set isLoadingRecent(value) {
			isLoadingRecent = value;
		},
		get recentError() {
			return recentError;
		},
		set recentError(value) {
			recentError = value;
		},
		async fetchCurrentPrice() {
			store.isLoadingPrice = true;
			store.priceError = null;

			const resp = await authFetch('/desim/token/price');
			if (!resp.ok) {
				store.priceError = 'Failed to fetch current token price';
				store.isLoadingPrice = false;
				return null;
			}

			const response: ApiResponse<{ price: number }> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				store.priceError = response.error || 'Failed to fetch current token price';
				store.isLoadingPrice = false;
				return null;
			}

			store.currentPrice = response.data.price;
			store.isLoadingPrice = false;
			return response.data.price;
		},
		startPriceRefresh() {
			// Fetch immediately on start
			this.fetchCurrentPrice();
			
			// Set up interval for automatic refresh
			if (refreshIntervalId) {
				clearInterval(refreshIntervalId);
			}
			refreshIntervalId = setInterval(() => {
				this.fetchCurrentPrice();
			}, REFRESH_INTERVAL);
		},
		stopPriceRefresh() {
			if (refreshIntervalId) {
				clearInterval(refreshIntervalId);
				refreshIntervalId = null;
			}
		},
		async fetchHistory(range: TimeRange) {
			store.isLoading = true;
			store.error = null;
			store.selectedRange = range;

			const resp = await authFetch(`/desim/token/supply/history?range=${range}`);
			if (!resp.ok) {
				store.error = 'Failed to fetch token supply history';
				store.isLoading = false;
				return null;
			}

			const response: ApiResponse<TokenSupplyHistoryPoint[]> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				store.error = response.error || 'Failed to fetch token supply history';
				store.isLoading = false;
				return null;
			}

			store.history = response.data;
			store.isLoading = false;
			return response.data;
		},
		async fetchRecent() {
			store.isLoadingRecent = true;
			store.recentError = null;

			const resp = await authFetch('/desim/token/recent');
			if (!resp.ok) {
				store.recentError = 'Failed to fetch recent token prices';
				store.isLoadingRecent = false;
				return null;
			}

			const response: ApiResponse<number[]> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				store.recentError = response.error || 'Failed to fetch recent token prices';
				store.isLoadingRecent = false;
				return null;
			}

			store.recent = response.data;
			store.isLoadingRecent = false;
			return response.data;
		},
		clear() {
			store.history = [];
			store.error = null;
			store.currentPrice = null;
			store.priceError = null;
			store.recent = [];
			store.recentError = null;
		},
	};
	
	// Start price refresh on store creation
	store.startPriceRefresh();
	
	return store;
}

export const tokenSupplyHistoryStore = createTokenSupplyHistoryStore();

