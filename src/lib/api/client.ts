import type {
	AuthResponse,
	DashboardResponse,
	ErrorResponse,
	HealthResponse,
	LoginRequest,
	MessageResponse,
	RegisterRequest,
} from './types';

const BASE_PATH = '/api';

/** Thrown for any non-2xx API response. */
export class ApiError extends Error {
	readonly status: number;

	constructor(status: number, message: string) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
	}
}

function isErrorResponse(body: unknown): body is ErrorResponse {
	return typeof body === 'object' && body !== null && typeof (body as ErrorResponse).error === 'string';
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const res = await fetch(`${BASE_PATH}${path}`, {
		credentials: 'include',
		...init,
	});

	const body: unknown = await res.json().catch(() => null);

	if (!res.ok) {
		throw new ApiError(res.status, isErrorResponse(body) ? body.error : `Request failed with status ${res.status}`);
	}

	return body as T;
}

function post<T>(path: string, body?: unknown): Promise<T> {
	return request<T>(path, {
		method: 'POST',
		...(body === undefined ? {} : { headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }),
	});
}

/**
 * Typed wrapper over the Go API. Request and response types come from
 * src/lib/api/types.ts, which is generated from internal/domain.
 */
export const api = {
	health: () => request<HealthResponse>('/health'),
	register: (body: RegisterRequest) => post<AuthResponse>('/auth/register', body),
	login: (body: LoginRequest) => post<AuthResponse>('/auth/login', body),
	logout: () => post<MessageResponse>('/auth/logout'),
	me: () => request<AuthResponse>('/auth/me'),
	dashboard: () => request<DashboardResponse>('/dashboard'),
};
