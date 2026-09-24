export interface User {
	id: number;
	email: string;
}
export interface AuthResponse {
	user: User;
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

function post<T>(path: string, body?: unknown): Promise<T> {
	return request<T>(path, {
		method: 'POST',
		...(body === undefined ? {} : { headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }),
	});
}

export const api = {
	register: (email: string, password: string) => post<AuthResponse>('/auth/register', { email, password }),
	login: (email: string, password: string) => post<AuthResponse>('/auth/login', { email, password }),
	logout: () => post('/auth/logout'),
	me: () => request<AuthResponse>('/auth/me'),
	dashboard: () => request<DashboardResponse>('/dashboard'),
};
