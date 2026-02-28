export const customFetch = async <T>({
	url,
	method,
	params,
	data,
	headers,
}: {
	url: string;
	method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
	params?: Record<string, string>;
	data?: unknown;
	headers?: Record<string, string>;
	signal?: AbortSignal;
}): Promise<T> => {
	const searchParams = params ? `?${new URLSearchParams(params)}` : '';
	const fullUrl = `/api${url}${searchParams}`;

	const res = await fetch(fullUrl, {
		method,
		credentials: 'include',
		...(data ? { body: JSON.stringify(data), headers: { 'Content-Type': 'application/json', ...headers } } : { headers }),
	});

	const body = await res.json().catch(() => null);

	if (!res.ok) {
		const msg = body && typeof body === 'object' && 'error' in body ? (body as Record<string, string>).error : `Request failed with status ${res.status}`;
		throw new Error(msg);
	}

	return body as T;
};

export default customFetch;
