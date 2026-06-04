/**
 * Converts an API error response body into a user-friendly Ukrainian message.
 *
 * Priority:
 *  1. Stable `error_code` from the backend → localized message via the map below.
 *  2. Raw validator leak detection (Key: '...', failed on the '...' tag) → generic.
 *  3. HTTP status code → generic message.
 *  4. Fallback generic.
 */

const ERROR_CODE_MAP: Record<string, string> = {
	// Validation — field-level
	REQUIRED_EMAIL: 'Введіть адресу електронної пошти',
	REQUIRED_PASSWORD: 'Введіть пароль',
	REQUIRED_FIELD: 'Заповніть усі обов\'язкові поля',
	INVALID_EMAIL: 'Введіть коректну адресу електронної пошти',
	PASSWORD_TOO_SHORT: 'Пароль має містити щонайменше 8 символів',
	VALUE_TOO_SHORT: 'Значення занадто коротке',
	VALUE_TOO_LONG: 'Значення занадто довге',
	VALUE_TOO_SMALL: 'Значення занадто мале',
	VALUE_TOO_LARGE: 'Значення занадто велике',
	INVALID_OPTION: 'Обрано невірний варіант',
	INVALID_URL: 'Введіть коректне посилання (наприклад, https://example.com)',
	INVALID_PHONE: 'Введіть коректний номер телефону',
	INVALID_ID: 'Невірний ідентифікатор',
	INVALID_FIELD: 'Одне з полів містить некоректне значення',
	INVALID_REQUEST: 'Некоректний запит. Перевір введені дані.',

	// Auth
	UNAUTHORIZED: 'Сесія закінчилась. Увійдіть знову.',
	RESET_FAILED: 'Не вдалося надіслати листа. Спробуй пізніше.',
	INVALID_RESET_TOKEN: 'Посилання для скидання пароля недійсне або застаріле.',
};

/** Raw validator string patterns that should never be shown to users. */
const RAW_VALIDATOR_PATTERNS = [
	"Key: '",
	"failed on the '",
	"Field validation for",
	"failed to parse token",
];

function isRawSystemError(msg: string): boolean {
	return RAW_VALIDATOR_PATTERNS.some((p) => msg.includes(p));
}

function messageForStatus(status: number): string {
	if (status === 401 || status === 403) return 'Сесія закінчилась. Увійдіть знову.';
	if (status === 404) return 'Не знайдено.';
	if (status === 409) return 'Цей запис вже існує.';
	if (status === 429) return 'Забагато запитів. Зачекайте трохи і спробуйте ще раз.';
	if (status >= 500) return 'Помилка сервера. Спробуй пізніше.';
	return 'Щось пішло не так. Спробуй ще раз.';
}

export interface ApiErrorBody {
	error?: string;
	error_code?: string;
}

/**
 * Returns a user-friendly error message from an API error response body.
 * Pass `status` for HTTP-status-based fallbacks.
 */
export function parseApiError(body: ApiErrorBody, status?: number): string {
	// 1. Stable error code mapping
	if (body.error_code && ERROR_CODE_MAP[body.error_code]) {
		return ERROR_CODE_MAP[body.error_code];
	}

	// 2. If the raw error string looks like an internal leak, replace it
	const rawError = typeof body.error === 'string' ? body.error : '';
	if (rawError && !isRawSystemError(rawError)) {
		return rawError;
	}

	// 3. Status-based fallback
	if (status !== undefined) {
		return messageForStatus(status);
	}

	return 'Щось пішло не так. Спробуй ще раз.';
}

/**
 * Reads and parses an error from a failed Response object.
 */
export async function parseResponseError(resp: Response): Promise<string> {
	try {
		const body: ApiErrorBody = await resp.json().catch(() => ({}));
		return parseApiError(body, resp.status);
	} catch {
		return messageForStatus(resp.status);
	}
}
