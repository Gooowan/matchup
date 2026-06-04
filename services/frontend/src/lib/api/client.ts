/**
 * Base API client helpers.
 * All typed wrappers in lib/api/* go through authFetch and this module.
 */
import { authFetch } from '$lib/utils/authFetch';
import { parseApiError } from '$lib/utils/parseApiError';

export interface ApiResponse<T> {
	data: T;
	error?: string;
	error_code?: string;
}

export async function apiGet<T>(path: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> {
	let url = path;
	if (params) {
		const qs = Object.entries(params)
			.filter(([, v]) => v !== undefined && v !== null && v !== '')
			.map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
			.join('&');
		if (qs) url += (url.includes('?') ? '&' : '?') + qs;
	}
	const resp = await authFetch(url);
	if (!resp.ok) {
		const body = await resp.json().catch(() => ({}));
		throw new ApiError(resp.status, parseApiError(body, resp.status));
	}
	const body: ApiResponse<T> = await resp.json();
	return body.data;
}

export async function apiPost<T>(path: string, payload?: unknown): Promise<T> {
	const resp = await authFetch(path, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: payload !== undefined ? JSON.stringify(payload) : undefined
	});
	if (!resp.ok) {
		const body = await resp.json().catch(() => ({}));
		throw new ApiError(resp.status, parseApiError(body, resp.status));
	}
	const body: ApiResponse<T> = await resp.json();
	return body.data;
}

export async function apiPut<T>(path: string, payload?: unknown): Promise<T> {
	const resp = await authFetch(path, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: payload !== undefined ? JSON.stringify(payload) : undefined
	});
	if (!resp.ok) {
		const body = await resp.json().catch(() => ({}));
		throw new ApiError(resp.status, parseApiError(body, resp.status));
	}
	const body: ApiResponse<T> = await resp.json();
	return body.data;
}

export class ApiError extends Error {
	constructor(
		public readonly status: number,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}
