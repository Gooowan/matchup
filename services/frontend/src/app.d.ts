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

	interface LoginData {
		user: UserDTO;
		auth: AuthDTO;
	}

	interface RegistrationRequest {
		email: string;
		password: string;
		profile_data: Record<string, any>;
		metadata?: Record<string, any>;
	}

	interface RegisterData {
		user: UserDTO;
	}

	interface UserDTO {
		id: string;
		role: string;
		email: string;
		profile_data: Record<string, any>;
		created_at: number;
	}

	interface AdminUserDTO {
		id: string;
		email: string;
		inviter_id: string;
		metadata: Record<string, any>;
		profile_data: Record<string, any>;
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
}

export {};
