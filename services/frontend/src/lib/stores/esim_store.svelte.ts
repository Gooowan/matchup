import { authFetch } from '$utils/authFetch';

function createEsimStore() {
	let plans = $state<EsimPlan[]>([]);
	let userOrders = $state<EsimOrder[]>([]);
	let isPlansLoading = $state<boolean>(false);
	let isUserOrdersLoading = $state<boolean>(false);
	let error = $state<string | null>(null);
	let selectedLocationCode = $state<string>('');

	return {
		get plans() {
			return plans;
		},
		set plans(value) {
			plans = value;
		},
		get userOrders() {
			return userOrders;
		},
		set userOrders(value) {
			userOrders = value;
		},
		get isPlansLoading() {
			return isPlansLoading;
		},
		set isPlansLoading(value) {
			isPlansLoading = value;
		},
		get isUserOrdersLoading() {
			return isUserOrdersLoading;
		},
		set isUserOrdersLoading(value) {
			isUserOrdersLoading = value;
		},
		get error() {
			return error;
		},
		set error(value) {
			error = value;
		},
		get selectedLocationCode() {
			return selectedLocationCode;
		},
		set selectedLocationCode(value) {
			selectedLocationCode = value;
		},

		async fetchPlans(locationCode?: string) {
			this.isPlansLoading = true;
			this.error = null;

			const url = locationCode 
				? `/desim/esim/plans?locationCode=${locationCode}`
				: '/desim/esim/plans';

			try {
				const resp = await authFetch(url);
				if (!resp.ok) {
					if (resp.status === 404) {
						this.error = 'eSIM service is not available. Please contact support.';
					} else {
						this.error = 'Failed to fetch eSIM plans';
					}
					this.isPlansLoading = false;
					return null;
				}

				const response: ApiResponse<{ plans: EsimPlan[] }> = await resp.json();
				if (resp.status !== 200 || !response.data) {
					this.error = response.error || 'Failed to fetch eSIM plans';
					this.isPlansLoading = false;
					return null;
				}

				this.plans = response.data.plans;
				this.isPlansLoading = false;
				return response.data.plans;
			} catch (error) {
				this.error = 'eSIM service unavailable';
				this.isPlansLoading = false;
				return null;
			}
		},

		async purchaseEsim(packageCode: string) {
			this.isUserOrdersLoading = true;
			this.error = null;

			const resp = await authFetch('/desim/esim/purchase', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					packageCode: packageCode,
				}),
			});

			if (!resp.ok) {
				this.error = 'Failed to purchase eSIM';
				this.isUserOrdersLoading = false;
				return null;
			}

			const response: ApiResponse<{ order: EsimOrder }> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to purchase eSIM';
				this.isUserOrdersLoading = false;
				return null;
			}

			await this.fetchUserOrders();

			this.startOrderPolling(response.data.order.id);

			this.isUserOrdersLoading = false;
			return response.data.order;
		},

		async fetchUserOrders(page: number = 1, limit: number = 20) {
			this.isUserOrdersLoading = true;
			this.error = null;

			const resp = await authFetch(`/desim/esim/orders?page=${page}&limit=${limit}`);
			if (!resp.ok) {
				this.error = 'Failed to fetch eSIM orders';
				this.isUserOrdersLoading = false;
				return null;
			}

			const response: ApiResponse<{ orders: EsimOrder[], pagination: any }> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to fetch eSIM orders';
				this.isUserOrdersLoading = false;
				return null;
			}

			this.userOrders = response.data.orders;
			this.isUserOrdersLoading = false;
			return {
				orders: response.data.orders,
				pagination: response.data.pagination,
			};
		},

		async getOrderDetails(orderId: number) {
			this.error = null;

			const resp = await authFetch(`/desim/esim/order/${orderId}`);
			if (!resp.ok) {
				this.error = 'Failed to fetch order details';
				return null;
			}

			const response: ApiResponse<{ order: EsimOrder }> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to fetch order details';
				return null;
			}

			return response.data.order;
		},

		async refreshAll() {
			await Promise.all([
				this.fetchPlans(this.selectedLocationCode),
				this.fetchUserOrders(),
			]);
		},

		startOrderPolling(orderId: number) {
			// Poll for order updates every 3 seconds for up to 2 minutes
			let attempts = 0;
			const maxAttempts = 40; // 40 * 3 seconds = 2 minutes
			
			const pollInterval = setInterval(async () => {
				attempts++;
				
				try {
					const orderDetails = await this.getOrderDetails(orderId);
					
					// If order is completed, stop polling and refresh the list
					if (orderDetails && orderDetails.status === 'completed') {
						clearInterval(pollInterval);
						await this.fetchUserOrders(); // Refresh the list
						return;
					}
					
					if (attempts >= maxAttempts) {
						clearInterval(pollInterval);
						await this.fetchUserOrders();
					}
				} catch (error) {
					console.error('Error polling order status:', error);
				}
			}, 3000); 
		},

		clear() {
			this.plans = [];
			this.userOrders = [];
			this.error = null;
			this.selectedLocationCode = '';
		},
	};
}

export const esimStore = createEsimStore();

