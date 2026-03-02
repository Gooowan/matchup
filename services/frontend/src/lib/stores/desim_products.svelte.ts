import { authFetch } from '$utils/authFetch';

function createDesimProductsStore() {
	let products = $state<DesimProduct[]>([]);
	let userProducts = $state<UserDesimProduct[]>([]);
	let isProductsLoading = $state<boolean>(false);
	let isUserProductsLoading = $state<boolean>(false);
	let error = $state<string | null>(null);

	return {
		get products() {
			return products;
		},
		set products(value) {
			products = value;
		},
		get userProducts() {
			return userProducts;
		},
		set userProducts(value) {
			userProducts = value;
		},
		get isProductsLoading() {
			return isProductsLoading;
		},
		set isProductsLoading(value) {
			isProductsLoading = value;
		},
		get isUserProductsLoading() {
			return isUserProductsLoading;
		},
		set isUserProductsLoading(value) {
			isUserProductsLoading = value;
		},
		get error() {
			return error;
		},
		set error(value) {
			error = value;
		},
		async fetchProducts() {
			this.isProductsLoading = true;
			this.error = null;

			const resp = await authFetch('/desim/products/list');
			if (!resp.ok) {
				this.error = 'Failed to fetch products';
				this.isProductsLoading = false;
				return null;
			}

			const response: ApiResponse<DesimProduct[]> = await resp.json();
			if (resp.status !== 200) {
				this.error = response.error || 'Failed to fetch products';
				this.isProductsLoading = false;
				return null;
			}

			this.products = response.data || [];
			this.isProductsLoading = false;
			return this.products;
		},
		async fetchUserProducts() {
			this.isUserProductsLoading = true;
			this.error = null;

			const resp = await authFetch('/desim/products/purchased');
			if (!resp.ok) {
				this.error = 'Failed to fetch user products';
				this.isUserProductsLoading = false;
				return null;
			}

			const response: ApiPaginatedResponse<UserDesimProduct> = await resp.json();
			if (resp.status !== 200) {
				this.error = response.error || 'Failed to fetch user products';
				this.isUserProductsLoading = false;
				return null;
			}

			this.userProducts = response.data || [];
			this.isUserProductsLoading = false;
			return this.userProducts;
		},
		async purchaseProduct(productId: number) {
			this.isUserProductsLoading = true;
			this.error = null;

			const resp = await authFetch('/desim/products/purchase', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					product_id: productId,
				}),
			});

			if (!resp.ok) {
				this.error = 'Failed to purchase product';
				this.isUserProductsLoading = false;
				return null;
			}

			const response: ApiResponse<PurchaseResult<UserDesimProduct>> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to purchase product';
				this.isUserProductsLoading = false;
				return null;
			}

			await this.fetchUserProducts();
			this.isUserProductsLoading = false;
			return response.data;
		},
		async getCommissionPercent(userProductId: number): Promise<{
			commission_percent: number;
			zero_commission_price: number | null;
		}> {
			const resp = await authFetch(`/desim/products/${userProductId}/commission`);
			if (!resp.ok) {
				this.error = 'Failed to fetch commission';
				throw new Error(this.error);
			}

			const response: ApiResponse<{
				commission_percent: number;
				zero_commission_price: number | null;
			}> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to fetch commission';
				throw new Error(this.error);
			}

			return {
				commission_percent: response.data.commission_percent,
				zero_commission_price: response.data.zero_commission_price,
			};
		},
		async claimProduct(userProductId: number, claimAmount: number) {
			this.isUserProductsLoading = true;
			this.error = null;

			const resp = await authFetch(`/desim/products/${userProductId}/claim`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					claim_amount: claimAmount,
				}),
			});

			if (!resp.ok) {
				this.error = 'Failed to claim product';
				this.isUserProductsLoading = false;
				return null;
			}

			const response: ApiResponse<any> = await resp.json();
			if (resp.status !== 200 || !response.data) {
				this.error = response.error || 'Failed to claim product';
				this.isUserProductsLoading = false;
				return null;
			}

			await this.fetchUserProducts();
			this.isUserProductsLoading = false;
			return response.data;
		},
		async refreshAll() {
			await Promise.all([this.fetchProducts(), this.fetchUserProducts()]);
		},
		clear() {
			this.products = [];
			this.userProducts = [];
			this.error = null;
		},
	};
}

export const desimProductsStore = createDesimProductsStore();
