export interface User {
	id: number;
	email: string;
}
export interface AuthResponse {
	user: User;
	csrfToken: string;
}
export interface DashboardResponse {
	message: string;
	email: string;
}
interface ErrorResponse {
	error: string;
}

export class ApiError extends Error {
	constructor(
		readonly status: number,
		message: string,
	) {
		super(message);
		this.name = 'ApiError';
	}
}

function isErrorResponse(value: unknown): value is ErrorResponse {
	return typeof value === 'object' && value !== null && 'error' in value && typeof value.error === 'string';
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const response = await fetch(`/api${path}`, { credentials: 'include', ...init });
	const body: unknown = await response.json().catch(() => null);
	if (!response.ok)
		throw new ApiError(
			response.status,
			isErrorResponse(body) ? body.error : `Request failed with status ${response.status}`,
		);
	return body as T;
}

function post<T>(path: string, body?: unknown, csrfToken?: string): Promise<T> {
	return request<T>(path, {
		method: 'POST',
		headers: {
			...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
			...(csrfToken === undefined ? {} : { 'X-CSRF-Token': csrfToken }),
		},
		...(body === undefined ? {} : { body: JSON.stringify(body) }),
	});
}

export const api = {
	logout: (csrfToken: string) => post('/auth/logout', undefined, csrfToken),
	me: () => request<AuthResponse>('/auth/me'),
	dashboard: () => request<DashboardResponse>('/dashboard'),
};
