import '@poppanator/sveltekit-svg/dist/svg.d.ts';

declare global {
	namespace App {
		interface Platform {
			env: Env;
			cf: CfProperties;
			ctx: ExecutionContext;
		}
	}

	interface ApiResponse<T> {
		error?: string;
		data?: T;
	}

	interface PaginationMeta {
		page: number;
		take: number;
		itemCount: number;
		pageCount: number;
	}

	interface ApiPaginatedResponse<T> {
		error?: string;
		data?: T[];
		meta?: PaginationMeta;
	}

	interface Invoice {
		id: string;
		owner_id: string;
		amount: number;
		chain: string;
		token: string;
		status: string;
		metadata?: Record<string, any>;
		created_at: string;
		updated_at: string;
	}

	interface LoginData {
		user: UserDTO;
		auth: AuthDTO;
	}

	interface RegistrationRequest {
		email: string;
		password: string;
		referral_id?: number;
		profile_data: Record<string, any>;
		metadata?: Record<string, any>;
		extensions?: Record<string, any>;
	}

	interface RegisterData {
		user: UserDTO;
	}

	interface UserDTO {
		id: string;
		role: string;
		telegram_id: number;
		email: string;
		referral_id: number;
		inviter_id: string;
		telegram_data: Record<string, any>;
		profile_data: Record<string, any>;
		created_at: number;
	}

	interface AdminUserDTO {
		id: string;
		telegram_id?: number;
		email: string;
		referral_id: number;
		inviter_id: string;
		metadata: Record<string, any>;
		profile_data: Record<string, any>;
		telegram_data: Record<string, any>;
		created_at: number;
		role: string;
		auth_nonce: number;
		forgot_password_token?: string;
		email_verification_token?: string;
	}

	interface AuthDTO {
		token: string;
		expires: number;
	}

	// ============= BALANCE TYPES =============
	interface Balance {
		available: number;
		withdraw_available: number;
		withdraw_pending: number;
		on_hold: number;
		referral: number;
		total_invested: number;
	}

	// ============= TRANSACTION TYPES =============
	interface Transaction {
		id: string;
		is_fake: boolean;
		created_at: number;
		initiator_id: string;
		owner_id: string;
		type: TransactionType;
		amount: number;
		status: TransactionStatus;
		metadata: Record<string, any>;
	}

	type TransactionStatus = 'SUCCESS' | 'PENDING' | 'HOLD' | 'CHECK' | 'REJECTED' | 'FAILED';

	type TransactionType = 'REPLENISH' | 'WITHDRAW' | 'PURCHASE' | 'PASSIVE' | 'REFERRAL' | 'MISSED_PROFIT';

	// ============= MLM TYPES =============
	interface UserFirstLineRow {
		id: string;
		email: string;
		referral_id: number;
		inviter_id: string;
		created_at: number;
		profile_data: Record<string, any>;
	}

	interface SearchUserReferralsDeepRow {
		id: string;
		email: string;
		referral_id: number;
		inviter_id: string;
		created_at: number;
		profile_data: Record<string, any>;
		level: number;
	}

	interface UserReferralStats {
		user_id: string;
		inviter_id: string | null;
		email: string;
		referral_id: number;
		profile_data: Record<string, any>;
		personal_turnover: number;
		direct_count: number;
		direct_turnover: number;
		total_team_count: number;
		total_team_turnover: number;
		current_rank_level: number;
		current_rank_metadata: Record<string, any>;
	}

	interface DirectReferralStat extends UserReferralStats {
		// Same structure as UserReferralStats, returned from firstLine endpoint
	}

	// ============= PRODUCT TYPES =============
	interface PurchaseResult<T> {
		user_product: T;
		transaction: Transaction;
	}

	// ============= PRODUCT TYPES =============
	interface Product {
		id: number;
		amount: number;
		metadata: Record<string, any>;
	}

	interface UserProduct {
		id: number;
		is_fake: boolean;
		is_active: boolean;
		owner_id: string;
		product_id: number;
		created_at: number;
		metadata: Record<string, any>;
	}

	interface DesimProduct extends Product {
		metadata: Record<string, any>;
	}

	interface UserDesimProduct extends UserProduct {
		accrued_balance: number;
		product_metadata: Record<string, any>;
		product_amount: number;
	}

	// ============= REQUEST/RESPONSE TYPES =============
	interface PurchaseProductRequest {
		product_id: number;
	}

	interface GetUserReferralParams {
		id: string;
	}

	interface GetUserReferralFirstLineParams {
		id: string;
	}

	interface GetUserReferralFirstLineCountParams {
		id: string;
	}

	interface GetUserReferralsParams {
		q?: string;
	}

	// ============= ESIM TYPES =============
	interface EsimPlan {
		packageCode: string;
		name: string;
		description: string;
		price: number;
		priceRaw: number;
		data: string;
		dataBytes: number;
		dataGB: number;
		duration: number;
		durationUnit: string;
		locationCode: string;
		location: string;
		speed: string;
		countries: string;
		networks?: any[];
	}

	interface EsimOrder {
		id: number;
		transactionId: string;
		orderNo?: string;
		packageCode: string;
		packageName?: string;
		locationCode?: string;
		status: 'pending' | 'completed' | 'failed';
		amount: number;
		iccid?: string;
		qrCode?: string;
		activationCode?: string;
		apiResponse?: Record<string, any>;
		createdAt: string;
		updatedAt: string;
		message?: string;
	}
}

export {};
