/**
 * Ukrainian phone-number formatter.
 *
 * Accepts many paste/typed variants and outputs a canonical `+380 XX XXX XX XX`
 * display string. Returns the raw input unchanged when it cannot be confidently
 * parsed as a Ukrainian number so users can still type freely.
 *
 * Examples:
 *   "0501234567"            -> "+380 50 123 45 67"
 *   "380501234567"          -> "+380 50 123 45 67"
 *   "+380501234567"         -> "+380 50 123 45 67"
 *   "+38 (050) 123-45-67"   -> "+380 50 123 45 67"
 *   "(050) 123 45 67"       -> "+380 50 123 45 67"
 *   "+44 20 7946 0958"      -> "+44 20 7946 0958" (unchanged, non-UA)
 */
export function formatUkrainianPhone(raw: string): string {
	if (!raw) return '';
	const trimmed = raw.trim();
	const hasPlus = trimmed.startsWith('+');

	const digits = trimmed.replace(/\D+/g, '');

	let body: string | null = null;

	if (digits.length === 12 && digits.startsWith('380')) {
		body = digits.slice(3);
	} else if (digits.length === 11 && digits.startsWith('80')) {
		body = digits.slice(2);
	} else if (digits.length === 10 && digits.startsWith('0')) {
		body = digits.slice(1);
	} else if (digits.length === 9 && !hasPlus) {
		body = digits;
	}

	if (body === null || body.length !== 9) {
		return raw;
	}

	const operator = body.slice(0, 2);
	const part1 = body.slice(2, 5);
	const part2 = body.slice(5, 7);
	const part3 = body.slice(7, 9);

	return `+380 ${operator} ${part1} ${part2} ${part3}`.replace(/\s+$/, '');
}

/**
 * Returns the digits-only E.164 representation if the input is a Ukrainian
 * phone number, otherwise returns the raw input. Useful for storing/comparing.
 */
export function toUkrainianE164(raw: string): string {
	const formatted = formatUkrainianPhone(raw);
	if (!formatted.startsWith('+380 ')) return raw;
	return '+' + formatted.replace(/\D+/g, '');
}
